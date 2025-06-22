---
title: Compensation Protocols
weight: 20
next: caching-patterns
---

This section delves into simpler protocols that operate abstractly,
without relying on low-level concepts.

In essence, **Compensation Protocols** enable transactions to commit data in a **reversible** manner.
If an issue arises later, these transactions can roll back the system by compensating for the previously committed changes.

Designing these protocols is challenging because they don't inherently meet the **Isolation** requirement.
The **Commit** and **Compensate** phases often behave as if they are parts of **separate transactions**.
If the design is flawed, other operations might update data before the **Compensate** phase.
This can lead to the use of skewed data and the generation of inconsistencies.
Furthermore, this separation makes it impossible to provide **Strong Consistency**,
a strict requirement in some critical systems.

## Try-Confirm/Cancel (TCC)

**Try-Confirm/Cancel (TCC)** is an approach similar to **Two-Phase Commit (2PC)** but operates without locking mechanisms.
It also involves two phases coordinated by a central entity:

1. **Try**: The coordinator instructs participants to perform **tentative** actions, such as reserving resources.
It's important to note that data is **actually committed** during this phase, not merely marked as dirty data.

2. **Confirm or Cancel**:

    - **Confirm**: If all participants successfully complete their **Try** actions,
    the coordinator requests them to confirm, thereby finalizing the rest of the operations.
    - **Cancel**: If any participant fails to prepare, other participants will revert their changes.

Consider a transaction transferring money between accounts located on different servers:

{{% steps %}}

### Try Phase

During the **Try** phase, the coordinator initiates the transaction by sending `Try` requests to all participating services (in this case, both banks).

- `Account A`'s balance is decreased by the transaction amount.
- The system verifies if `Account B` is valid and can receive funds.

```d2
shape: sequence_diagram

c: Coordinator {
    class: server
}
s1: Server A (Account A) {
    class: db
}
s2: Server B (Account B) {
    class: db
}

"1. Try" {
    c -> s1: Update balance = balance - amount
    c -> s2: Verify a valid account
}
```

### Confirm Phase

If all participants successfully complete their `Try` operations.
It sends `Confirm` requests to all participants to finalize the transaction.

In this example, `Account B` will now definitively increase its balance by the amount debited from Account A.
The transaction is considered successful and complete.

```d2
shape: sequence_diagram

c: Coordinator {
    class: server
}
s1: Server A (Account A) {
    class: db
}
s2: Server B (Account B) {
    class: db
}

"1. Try" {
    c -> s1: Update balance = balance - amount
    c -> s2: Verify a valid account
}

"2. Confirm (If all yes)" {
    s1 -> c: Yes
    s2 -> c: Yes
    c -> s2: Update balance = balance + amount {
        style.bold: true
    }
}
```

### Cancel Phase

For instance, if `Account B` is blocked by the bank, `Server B` would respond with `No`.
The coordinator then instructs `Server A` to compensate for the change.

```d2
shape: sequence_diagram

c: Coordinator {
    class: server
}
s1: Server A (Account A) {
    class: db
}
s2: Server B (Account B) {
    class: db
}

"1. Try" {
    c -> s1: Update balance = balance - amount
    c -> s2: Verify a valid account
}

"2. Cancel (If any no)" {
    s1 -> c: Yes
    s2 -> c: No (B is being blocked by the bank){
        class: error-conn
    }
    c -> s1: Compensate balance = balance + amount {
        style.bold: true
    }
}
```

{{% /steps %}}

You might wonder why Account A's balance is updated in the **Try** phase, but Account B's is not.
The **Try** phase must not introduce harmful effects to the system.
It would be a poor design to increase Account B's balance initially,
as this temporary increment shouldn't be usable until it's validated.

Similar to **2PC**,
the primary advantage of **TCC** is its support for parallel processing,
allowing steps to be performed simultaneously.
However, **TCC** creates tight couplings between services because they need to expose low-level **Try-Confirm-Cancel** interfaces.
This necessitates a deep understanding of the participating services.

## Saga Pattern

**Saga** can be considered the most straightforward solution for distributed transactions.
In a saga, transactions are broken down into a series of steps executed sequentially.
If any step fails, the process moves backward to compensate for the preceding successful steps.

For instance, imagine an e-commerce system with three services. When a user places an order with the `Order Service`:

1. The `Order Service` creates the order record.
2. The `Stock Service` reduces the quantity of the ordered items.
3. The `Payment Service` completes the payment.

There are two primary ways to implement this request using **Saga**:

### Orchestration Saga

In this approach, a central coordinator is established, much like in **TCC** or **2PC**.
The coordinator executes the actions in sequence.
If any action fails, it orchestrates the backward compensation of previous actions.

When everything proceeds smoothly, the process might look like this:

```d2
shape: sequence_diagram

c: Client {
    class: client
}
o: Order Service (Coordinator) {
    class: server
}
s1: Stock Service {
    class: server
}
s2: Payment Service {
    class: server
}

c -> o: Order()
o -> o: "1. CreateOrder()"
o -> s1: "2. ReduceItems()"
o -> s2: "3. ProcessPayment()"
o -> o: "CompleteOrder()"
```

If, for example, the payment step fails, other services will be instructed to compensate for the previous steps:

```d2
shape: sequence_diagram
c: Client {
    class: client
}
o: Order Service (Coordinator) {
    class: server
}
s1: Stock Service {
    class: server
}
s2: Payment Service {
    class: server
}
c -> o: Order()
o -> o: "1. CreateOrder()"
o -> s1: "2. ReduceItems()"
o -> s2: "3. ProcessPayment() => failed" {
    class: error-conn
}
o -> s1: "IncreaseItems()" {
    style.bold: true
}
o -> o: "InvalidateOrder()" {
    style.bold: true
}
```

This seems straightforward, as it mirrors real-life sequential processes without partial operations or locking.

Again, why is the payment processed last?
Payment features are often implemented via third-party solutions.
Executing the payment earlier would make compensation more difficult.
Therefore, internal tasks are typically completed first.

However, this sequential nature means the **Saga** pattern, does not inherently support **parallelism**.
A transaction's operations must be performed in order, even if some could be productively parallelized.

### Choreography Saga

**Choreography Saga** implements the pattern asynchronously.
A transaction is completed through the collaboration of relevant events exchanged along the way.

Continuing with the e-commerce example, services communicate implicitly through an event stream:

```d2
shape: sequence_diagram
client: Client {
    class: client
}
o: Order Service (Coordinator) {
    class: server
}
s1: Stock Service {
    class: server
}
s2: Payment Service {
    class: server
}
e: Event Stream {
    class: mq
}
client -> o: "Order()"
o -> o: "CreateOrder()"
o -> e: OrderCreatedEvent {
    style.bold: true
}
s1 <- e: OrderCreatedEvent
s1 -> s1: "ReduceItems()"
s1 -> e: OrderReservedEvent {
    style.bold: true
}
s2 <- e: OrderReservedEvent
s2 -> s2: "ProcessPayment()"
```

In case of a failure, compensation events are generated.
For example, if the payment step fails, it triggers a chain of compensation events for the preceding steps:

```d2
shape: sequence_diagram
o: Order Service (Coordinator) {
    class: server
}
s1: Stock Service {
    class: server
}
s2: Payment Service {
    class: server
}
e: Event Stream {
    class: mq
}

s2 -> s2: Fail to process the payment {
    class: error-conn
}
s2 -> e: PaymentFailedEvent {
    style.bold: true
}
s1 <- e: PaymentFailedEvent
s1 -> s1: "IncreaseItems()" {
    style.bold: true
}
s1 -> e: OrderReleasedEvent {
    style.bold: true
}
o <- e: OrderReleasedEvent
o -> o: "InvalidateOrder()" {
    style.bold: true
}
```

**Orchestration** requires a central coordinator and direct communication,
which can degrade availability and introduce a {{< term spof >}}.
However, it conveniently centralizes the transaction logic within the coordinator.
In contrast, **Choreography** distributes the transaction logic across different services,
potentially making the source code difficult to understand and maintain.

Although **Choreography** can enhance availability and decouple the system,
its complexity can be overwhelming and may outweigh the benefits.
Moreover, **Saga** is inherently slower due to its sequential nature,
and the **Choreography** approach can exacerbate this with its indirect communication paradigm.

## Saga Modelling

As stated earlier, **Saga** is a compensation protocol and does not offer the **Isolation** property.
Before compensation occurs, dirty data can create vulnerabilities in the system.
Therefore, the most critical aspect of **Saga** is defining a safe and reliable workflow.

1. First and foremost, it's crucial to adopt the mindset that any step can **fail** occasionally.
When a step fails, we must ensure it doesn't jeopardize the system.

2. Second, not every step in a workflow can be compensated, especially when external factors are involved.
Imagine a step that transfers money to a user's bank account.
Reverting this step would require retrieving the amount from the bank,
which is nearly impossible as a bank typically won’t permit withdrawals without user consent.

Hence, in a saga, actions should be categorized into three groups:

1. **Compensable Action**: Can be rolled back using a corresponding compensation action if necessary.
These are usually internal workloads that can readily invalidate previous actions.
2. **Pivot Action**: Represents a **point of no return** in a workflow. Once it succeeds,
all subsequent actions must be completed, and **no compensation** can be applied.
It's typically the last **Compensable** action in a sequence.
3. **Retryable Action**: Can be safely retried multiple times without causing inconsistencies.
This action is irreversible and is expected to be **retried** until successful.

Essentially, Saga workflows are designed as follows.
After the pivot step is completed, compensation is no longer an option.

```d2
grid-rows: 2
Arrow {
   class: none
   grid-rows: 1
   grid-gap: 0
   m1 {
      class: none
      width: 1
   }
   m2 {
      class: none
      width: 500
   }
   m3 {
      class: none
      width: 1
   }
   m4 {
      class: none
      width: 200
   }
   m5 {
      class: none
      width: 1
   }
   m6 {
      class: none
      width: 390
   }
   m7 {
      class: none
      width: 1
   }
   m3 -> m1: Compensate
   m5 -> m7: Retry
}

s: "" {
   grid-rows: 1
   grid-gap: 0
   c: Compensable Actions {
      style.fill: ${colors.i1}
      width: 500
   }
   p: Pivot Actions {
      style.fill: ${colors.i2}
      width: 200
   }
   r: Retryable Actions {
      style.fill: ${colors.e}
      width: 400
   }
}
```

Back to the e-commerce example, we'll sort the transactions as follows

1. The `Order Service` makes an order `CreateOrder()`.
2. The `Stock Service` reserves the number of ordered items `ReserveItem()`.
3. The `Payment Service` processes the payment `ProcessPayment()`.
4. Finally, the `Delivery Service` creates a request `CreateDeliveryRequest()`.

```d2
grid-rows: 2
Arrow {
   class: none
   grid-rows: 1
   grid-gap: 0
   m1 {
      class: none
      width: 1
   }
   m2 {
      class: none
      width: 500
   }
   m3 {
      class: none
      width: 1
   }
   m4 {
      class: none
      width: 200
   }
   m5 {
      class: none
      width: 1
   }
   m6 {
      class: none
      width: 300
   }
   m7 {
      class: none
      width: 1
   }
   m3 -> m1: Compensate
   m5 -> m7: Retry
}

s: Saga {
   grid-rows: 1
   grid-gap: 0
   c: Compensable {
      grid-rows: 1
      c: "1. CreateOrder()" {
         width: 200
      }
      r: "2. ReserveItems()" {
         width: 200
      }
      c -> r
   }
   p: Pivot {
      label.near: top-center
      p: "3. ProcessPayment()" {
         width: 200
      }
   }
   r: Retryable {
      label.near: top-center
      c: "4. CreateDeliveryRequest()" {
         width: 200
      }
   }
   c.r -> p.p
   p.p -> r.c
}
```

The first two steps are internal processes, they're safe to compensate.
Payment and delivery requirements usually depend on third-party solutions,
it may be not feasible to revert them.
After the payment step is finished, we have to guarantee the completion of the final step.

However, this workflow varies based on the system.
If the delivery task is a part of the system, we may place it before the payment step.

### Saga Serialization Anomalies

Normally, multiple sagas can be executed concurrently and modify the same data.
That easily leads to serialization anomalies,
like what we've discussed in the [Concurrency Control]({{< ref "concurrency-control" >}}) topic.

We'll briefly consider some common anomalies and how to resolve them.

#### Dirty Read

The first anomaly is **Dirty Read**.
Similar to a rollback,
the compensation phase can cause data to become **dirty** for transactions that read it before compensation.

Let's consider an example: When a user places a large order, a discount voucher is issued to them.

```d2
shape: sequence_diagram

o: Order Service
v: Voucher Service
p: Payment Service

o -> o: "1. CreateOrder()"
o -> v: "2. CreateVoucher()"
o -> p: "3. ProcessPayment()"
```

Unfortunately, the final step (payment) fails.
The system needs to revert the second step by deleting the voucher.
However, before the deletion occurs, another saga might use that voucher.

```d2
shape: sequence_diagram

o: Order Service
v: Voucher Service
p: Payment Service
vs: Voucher usage saga
o -> o: "1. CreateOrder()"
o -> v: "2. CreateVoucher()"
o -> p: "3. ProcessPayment()" {
   class: error-conn
}
vs -> v: "UseVoucher()" {
   style.bold: true
}
o -> v: "DeleteVoucher()" {
    style.bold: true
}
```

The result is unexpected because an invalid voucher was used. There are several ways to resolve this:

1. **Rearrange the flow**: Move the voucher creation to the **Retryable** phase.
Although it's an internal and compensable workload,
its potential for causing issues makes it safer in the **Retryable** phase.

    This solution is known as the **Pessimistic View**,
    which assumes that actions prone to causing harm are likely to be compensated and should thus be deferred to a later point.

    ```d2
    shape: sequence_diagram

    o: Order Service
    v: Voucher Service
    p: Payment Service

    o -> o: "1. CreateOrder()"
    o -> p: "2. ProcessPayment()"
    o -> v: "3. CreateVoucher()"
    ```

2. **Lock data**: Employ a flag field that acts as a lock.
For example, when a new voucher is created, its state is set to `Pending`, marking it as unavailable for use.
When the saga that created the voucher completes successfully, it changes the state to `Approved`, allowing usage.

    This solution is called **Semantic Locking**, where application-level locking is implemented to prevent anomalies.

    ```d2
    shape: sequence_diagram

    s: Order Service
    v: Voucher Service
    p: Payment Service
    vs: Voucher Usage Saga

    s -> s: "1. CreateOrder()"
    s -> v: "2. CreateVoucher()"
    v {
        "state = Pending"
    }
    s -> p: "3. ProcessPayment()"
    vs -> v: Fail to use the pending voucher {
        class: error-conn
    }
    s -> v: "4. ApproveVoucher()" {
        style.bold: true
    }
    v {
        "state = Approved"
    }
    ```

The second approach may be a bit overhead for this example.
Let's move to the next anomaly to see its real power.

#### Lost Update

The next anomaly is **Lost Update**.
This is a specific case of [Write Skew]({{< ref "concurrency-control#write-skew" >}}),
where updates are unexpectedly overwritten by other concurrent operations.

For example, the `OrderService` creates an order, which is marked `Approved` at the end of its saga.
However, in the interim, a `CancelOrder Saga` executes and sets the order's state to `Cancelled`.
Consequently, the `CreateOrder Saga` might overwrite the change from the `CancelOrder` saga,
resulting in a cancelled order still being processed by the system.

```d2
shape: sequence_diagram
o: Order Service
p: Payment Service
c: CancelOrder Saga

o -> o: "1. CreateOrder()"
o -> p: "2. ProcessPayment()"
c -> o: Set state = Cancelled {
   style.bold: true
}
o {
   "state = Cancelled"
}
o -> o: "3. CompleteOrder()" {
   class: error-conn
}
o {
   "state = Completed"
}
```

**Semantic Locking** is an effective approach for preventing **Lost Update**.
As before, orders are locked with a `Pending` state until they are successfully created.

```d2
shape: sequence_diagram
o: Order Service
p: Payment Service
c: CancelOrder Saga
o -> o: "1. CreateOrder()"
o {
   "state = Pending"
}
o -> p: "2. ProcessPayment()"
c -> o: "Fails to set state = Cancelled of a pending order" {
    class: error-conn
}
o -> o: "3. CompleteOrder()" {
    style.bold: true
}
o {
   "state = Completed"
}
```

**Semantic Locking** is nothing but a distributed lock,
reducing parallelism and negatively affecting the performance.
In the first place, we should design a reliable flow and limit locking instead.

## Saga Transaction Recovery

### Transaction State

Both the coordinator and participating services can crash at any time.
Therefore, they must persist the state of transactions to enable retries if necessary.

Making transactions **idempotent** (where each transaction is marked with a **unique identifier**)
is an effective way to prevent duplications and aid in transaction recovery upon failure.

### Coordinator Recovery

The coordinator service may periodically scan its state store to complete any unfinished transactions.
Continuing with the previous example, the `Order Service` keeps track of the current transaction step:

```yaml
Order Service (coordinator):
  transaction-1234:
    Step: Payment Service
    State: Processing
```

For added safety, the transaction state is also **duplicated** in the participating services:

```yaml
Order Service (coordinator):
  transaction-1234:
    Step: Payment Service
    State: Processing
Stock Service:
  transaction-1234: Complete
Payment Service:
  transaction-1234: Complete
```

For example, if the `Order Service` (coordinator) fails before receiving a response from the `Payment Service`,
this can lead to mismatched states (`Processing` and `Complete`) across different services.

```d2
shape: sequence_diagram

o: Order Service
p: Payment Service

o.s1: "transaction-1234: Processing"
o -> p: "ProcessPayment()"
p."transaction-1234: Complete"
o -> o: Crash {
    class: error-conn
}
p -> o: Respond but the Order Service crashed {
    class: error-conn
}
```

After restarting, the `Order Service` will observe that the transaction is pending and will attempt to call the `Payment Service` again to complete it.
Fortunately, duplicating the state ensures that the `Payment Service` will not reprocess the transaction.

```d2
shape: sequence_diagram

o: Order Service
p: Payment Service

o.s1: "transaction-1234: Processing"
p.s2: "transaction-1234: Complete"
o -> o: Recover
o -> p: "ProcessPayment()" {
    style.bold: true
}
p -> o: The payment has been completed
o."transaction-1234: Complete"
```

### Choreographer Recovery

#### Consume-process-produce Pipeline

This pattern was discussed in the [Event Streaming]({{< ref "event-streaming-platform#consume-process-produce-pipeline" >}}) section.
The **Choreography Saga** helps to avoid duplication without relying on idempotency in this specific pipeline.

The **Consume-Process-Produce Pipeline** is applicable when a service's changes are confined to the event stream and do not affect other datasets.

```d2
shape: sequence_diagram

c: Coordinator {
    class: server
}
sp: Event Stream {
    class: mq
}

c -> sp: Begin transaction
c <- sp: Consume an event
c -> sp: Produce a new event
c -> sp: Commit
```

#### Transactional Outbox Pattern

Another challenge with **Choreography Saga** is ensuring that database changes are effectively published as events.
For example, when a transaction is executed in an internal data store,
we want to publish an associated event indicating its completion.

```d2
shape: sequence_diagram
s: Order Service {
    class: server
}
store: Order Store {
    class: db
}
e: Event Stream {
    class: mq
}

s -> store: Create order
s -> e: Publish event
```

The **Transactional Outbox Pattern** addresses this by storing external calls (event publications) as part of the internal database transaction.

For example, the transaction that creates an order also creates an associated record in the **Outbox** table,
instead of immediately firing the event.

```d2
s: Order Service {
    class: server
}
store: Order Store {
    o: "Order Table" {
        shape: sql_table
        OrderId: 1234
        TotalPrice: 2000
    }
    e: "Outbox Table" {
        shape: sql_table
        EventType: OrderCreated
        State: New
        EventData: "OrderId: 1234, TotalPrice: 2000"
    }
}
s -> store: CreateOrderTransaction()
```

A separate process (called a **Relay Process**) periodically scans the **Outbox** table and publishes any unprocessed events.

```d2
s: Relay Process {
    class: process
}
store: Order Store {
    o: "Order Table" {
        shape: sql_table
        OrderId: 1234
        UserId: 1905
        TotalPrice: 2000
    }
    e: "Outbox Table" {
        shape: sql_table
        EventType: OrderCreated
        State: New
        EventData: "OrderId: 1234, UserId: 1905, TotalPrice: 2000"
    }
}
e: Event Stream {
    class: mq
}

s <- store.e: 1. Scan
s -> e: 2. Publish events
s -> store.e: 3. Set State = Complete {
    style.bold: true
}
```

If the relay process crashes before updating the outbox record in the third step, it might publish the event **twice**.
Fortunately, consumers can use an idempotency key (e.g., `OrderId`) to check for and ignore duplicated events.
