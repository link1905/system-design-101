---
title: Compensation Protocols
weight: 20
---

In this section, we're diving into simpler protocols.
They work abstractly without the help of low-level concepts, but more challenging to prevent inconsistencies.

In short, **Compensation Protocols** allow transactions to commit data in a **reversible** way.
Afterwards, if anything goes wrong, the transactions rollback the system by compensating for the changes.

They are challenging to design because they do not meet the **Isolation** requirement,
**Commit** and **Compensate** phases are like from **separate transactions**.
If the design is malfunctioned, many things can come before the **Compensate** phase to update data in a
way making it use skewed data and generate inconsistencies.
Moreover, the separation makes it impossible to provide **Strong Consistency**, which
is strictly required in some critical systems.

## Try-Confirm/Cancel (TCC)

**Try-Confirm/Cancel (TCC)** is an approach similar to **2PC**, but without locking.
We also have two phases with a coordinator:

1. **Try**: The coordinator asks participants to perform **tentative** actions, such as reserving resources.
Please note that data is **actually committed**, not only dirty data with locking.

2. **Confirm or Cancel**:

- **Confirm**: If all participants tried successfully, the coordinator will request them to confirm (completing the rest).
- **Cancel**: If any participant failed to prepare, other participants will revert their changes.

For example, we have a transaction transferring money between accounts in different servers:

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
    c -> s2: Update balance = balance + amount 
}
"2. Cancel (If any no)" {
    s1 -> c: Yes
    s2 -> c: No (this account is being blocked by the bank)
    c -> s1: Compensate balance = balance + amount 
}
```

Why do we update the `A`'s balance in the `Try` phase, but not `B`?
The `Try` must not give harmful effects to the system.
It's a bad design if we increase the `B` balance in the first place,
the temporary increment shouldn't be used until it becomes valid.

Like **2PC**, the most advantage of **TCC** is parallel processing,
which steps can be performed simultaneously.
However, **TCC** creates high couplings between services as they need to
expose the low-level interfaces (**Try-Confirm-Cancel**),
requiring deep understanding of participating services.

## Saga

**Saga** can be treated as the naivest solution for distributed transactions.
In **Saga**, transactions are divided into phases executed **step-by-step**.
When any step fails, we move backward to compensate the previous ones.

For example, an e-commerce system has three services.
When a user makes an order to `Order Service`:

1. `Stock Service` reduces the number of ordered items.
2. `Payment Service` completes the payment.

We have two ways to complete this request in **Saga**:

### Orchestration Saga

First, we build a central coordinator like what we did in **TCC** or **2PC**.
The coordinator will perform the actions in order.
If any action fails, it will go backward to compensate for the previous ones.

When everything works perfectly, it may look like this:

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
c -> o: Order
o -> o: Create the order
o -> s1: Reduce the number of items
o -> s2: Process the payment
o -> o: Complete the order
```

For example, the payment step fails.
The `Stock Service` will be called to compensate for the first step

```d2
o: Order Service (Coordinator) {
    class: server
}
s1: Stock Service {
    class: server
}
s2: Payment Service {
    class: server
}
o -> o: Create the order
o -> s1: Reduce the number of items
o -> s2: Fail to process the payment {
    class: error-conn
}
o -> s1: Increase the number of items (compensate) {
    style.bold: true
}
o -> o: Invalidate the order (compensate) {
    style.bold: true
}
```

As easy as ABC, right? We actually implement what we think in real life.
No partial operation or locking!

Once again,
why do we let the payment happen last?
The payment feature is typically implemented through a third-party solution.
If we execute it earlier, it's hard to perform compensation jobs.
Thus, we'll attempt to complete internal works first.

However, it's so slow,
the **Saga** pattern cannot support **parallelism**.
A transaction's operations must be performed in order,
although they can be parallelized productively.

### Choreography Saga

**Choreography** implements the Saga pattern from the asynchronous manner.
A transaction is completed by the collaboration of relevant events along the way.

Continue with the previous example.
Services communicate implicitly through an event stream.

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
client -> o: Request order
o -> o: Create the order
o -> e: OrderCreatedEvent {
    style.bold: true
}
s1 <- e: OrderCreatedEvent
s1 -> s1: Reduce the number of items
s1 -> e: OrderReservedEvent {
    style.bold: true
}
s2 <- e: OrderReservedEvent
s2 -> s2: Process the payment
```

In case of failure, compensation events will be generated.
For example, the payment step fails,
it causes chaining compensation for the previous steps.

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
s2 -> s2: Fail to process the payment {
    class: error-conn
}
s2 -> e: PaymentFailedEvent {
    style.bold: true
}
s1 <- e: PaymentFailedEvent
s1 -> s1: Increase the number of items (compensate) {
    style.bold: true
}
s1 -> e: OrderReleasedEvent {
    style.bold: true
}
o <- e: OrderReleasedEvent
o -> o: Invalidate the order (compensate) {
    style.bold: true
}
```

The **Orchestration** requires a central coordinator and direct communication,
degrading availability and causing the {{< spof >}} problem.
However, it conveniently wraps everything inside the coordinator,
**Choreography** scatters transactions' logic in different services,
making sourcecode become extremely tough to understand and maintain.

On the other hand,
**Choreography** helps to boost the availability and decouple the system.
However, its complexity can be overwhelming and exceed the benefits.
Moreover, **Saga** is slow by nature,
and the **Choreography** approach makes it worse with the indirect paradigm.

## Transaction Recovery

### Transaction State

The coordinator and participating services can crash anytime.
They must keep track transactions' state to retry them if necessary.

As we said, a transaction can be divided into different steps performed at different moments.
Making transactions **idempotent** (a transaction is marked with **unique identifier**) is an effective way
prevent duplications and help recover transactions on failure.

### Coordinator Recovery

The coordinator service may periodically scan the state store to fulfil incomplete transactions

Continue with the previous example,
the `Order Service` keeps track of the current transaction step:

```yaml
Order Service (coordinator): 
  transaction-1234:
    Step: Payment Service
    State: Processing
```

We also duplicate transaction state in participating services for safety.

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

For example,
the `Order Service` (coordinator) fails before receiving the response from the `Payment Service`,
causing mismatched states (`Processing` and `Complete`) in different services.
After reviving, the `Order Service` observes that the transaction is pending,
it will try to call the `Payment Service` again to complete it.
Fortunately, duplicating the state ensures
that the `Payment Service` will not reprocess the transaction again.

```d2
shape: sequence_diagram
o: Order Service
p: Payment Service
o.s1: "transaction-1234: Processing"
o -> p: Process the order
p."transaction-1234: Processing"
p -> p: Process the payment successfully
p."transaction-1234: Complete"
p -> o: Respond but the Order Service crashed {
    class: error-conn
}
o -> o: Recover
o.s2: "transaction-1234: Processing"
o -> p: Continue processing the order
p -> o: The payment has been completed
o."transaction-1234: Complete"
```

### Choreographer Recovery

#### Consume-process-produce Pipeline

We've discussed this pattern in the [Event Streaming]({{< ref "event-streaming-platform#consume-process-produce-pipeline >}}).
The **Choreography Saga** helps to evade duplication **without** idempotency.
**Consume-process-produce Pipeline** appears when a service doesn’t make changes to other datasets rather than the event stream.

```d2
shape: sequence_diagram
sp: Event Stream {
    class: mq
}
c: Coordinator {
    class: server
}
c -> sp: Begin transaction
c <- sp: Process an event
c -> sp: Produce a new event 
c -> sp: Commit
```

#### Transactional Outbox Pattern

Another problem of the **Choreography Saga** ensures that database changes are atomically published.

For example, we execute a transaction in an internal store,
we want to publish the associated event **atomically** indicating the completion.

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

**Transactional Outbox Pattern** requires storing external calls as
part of internal transitions.

Now, we need to create a utility **Outbox** table inside the business database to hold intended events.
The transaction of creating an order also makes up an associated record in the **Outbox** table,
instead of firing the event immediately.

```d2
s: Order Service {
    class: server
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
s -> store: Create order (transaction)
```

A subsidiary process (called **Relay Process**) periodically scans and produces incomplete events from the **Outbox** table.

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
s -> store.e: 3. Set State = Produced
```

If the third step, if the processor crashes before updating the outbox record,
it may produce the event twice.
Luckily, based on the idempotency key (e.g. `OrderId`), consumers can check and ignore duplicated events.

## Saga Modelling

As we've said, **Saga** is a compensation protocol and doesn't offer the **Isolation** property.
Before compensated, dirty data can become a break point of the system.
Therefore, the most critical part of **Saga** is defining a safe and reliable workflow.

1. First and foremost, we need to have a mindset that any step is going to **fail** now and then.
And when a step fails, we need to make sure that it's not intimidating to the system.

2. Second, in a workflow, not every step can be compensated, especially when external factors are related.
Let's imagine we have a step transferring money to a user's bank;
To revert the step, we need to retrieve back the amount from the bank.
That's nearly impossible, a bank can’t permit withdrawals without user consent.

Hence, in Saga, we should categorize actions into three groups:

1. **Compensable Action** can be rolled back using a compensation action if needed.
This type is usually internal workloads willing to invalidate actions.
2. **Pivot Action** in a workflow that marks a **point of no return**.
Once it succeeds, all following actions must be completed, and **no compensation** can be applied.
It's typically the last **Compensable** action.
3. **Retryable Action** can be safely retried without causing inconsistencies.
This action is irreversible and supposed to **retry** until successful.

Basically, we will design Saga workflows as below.
After the pivot step is completed, compensation is no longer applicable.

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
      style.fill: ${colors.b}
      width: 500
   }
   p: Pivot Actions {
      style.fill: ${colors.c}
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
      r: "2. ReserveItem()" {
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
If the delivery task is a part of the system,
we may place it before the payment step.
In essence, we need to design a workflow that is safer and more manageable.

## Saga Serialization Anomalies

Normally, multiple sagas can be executed concurrently and modify the same data.
That easily leads to serialization anomalies,
like what we've discussed in the [Concurrency Control]({{< ref "concurrency-control" >}}) topic.
This is even more challenging to control,
as a step may comprise different transactions (transaction and compensation),
making low-level supports (locking, predicate locking...) from the database layer useless.

We'll briefly consider some common anomalies and how to resolve them.

### Dirty Read

The first anomaly is **Dirty Read**.
Similar to rollback, the compensation phase can make a piece of data become dirty for previous transactions.

Let's take an example.
When a user makes a big order, a discount voucher is given to the user.

```d2
shape: sequence_diagram
s1: Order Service
v: Voucher Service
p: Payment Service
s1 -> s: "1. CreateOrder()"
s1 -> v: "2. CreateVoucher()"
s1 -> p: "3. ProcessPayment()"
```

Unfortunately, the final step fails.
We need to revert the second step by deleting the voucher.
However, before the deletion, another saga comes in between and uses the voucher.

```d2
shape: sequence_diagram
s1: Order Service
v: Voucher Service 
p: Payment Service
vs: Voucher usage saga
s1 -> s: "1. CreateOrder()"
s1 -> v: "2. CreateVoucher()"
s1 -> s2: "3. ProcessPayment()" {
   class: error-conn
}
vs -> v: Use the voucher {
   style.bold: true
}
s1 -> o: "4. DeleteVoucher()"
```

The result is unexpected, since we let an invalid voucher be used.
There are several ways to resolve this situation

1. First, we may rearrange the flow by moving the voucher to the **Retryable** phase.
Although it's an internal and compensable workload, it's dangerous to live in the **Compensation** phase.
This solution is called **Pessimistic View**,
considering harmful actions are likely to be compensated, and should be moved to **Retryable** phase.

```d2
shape: sequence_diagram
s1: Order Service
v: Voucher Service
p: Payment Service
s1 -> s: "1. CreateOrder()"
s1 -> p: "2. ProcessPayment()"
s1 -> v: "3. CreateVoucher()"
```

2. Another approach is using a **flag** field working as a lock.
For example, when a new voucher is created, we assign its state as `Pending` marking it unavailable to use.
When the creation saga completes, it changes the state to `Approved` to allow usage.
This solution is called **Semantic Locking**, we build an application-level locking to prevent anomalies.

```d2
shape: sequence_diagram
s1: Order Service
v: Voucher Service 
p: Payment Service
vs: Voucher usage saga
s1 -> s: "1. CreateOrder()"
v {
   "state = Pending"
}
s1 -> v: "2. CreateVoucher()"
s1 -> p: "3. ProcessPayment()" {
   class: error-conn
}
vs -> v: Fail to use the pending voucher {
   class: error-conn
}
s1 -> v: "4. ApproveVoucher()" 
v {
   "state = Approved"
}
```

The second approach may be a bit overhead for this example.
Let's move to the next anomaly to see its real power.

### Lost Update

The next anomaly is **Lost Update**, this is a certain case of [Write Skew]({{< ref "concurrency-control#write-skew" >}}),
when updates are unexpectedly overwritten by others.

For example, the `OrderService` creates an order,
which is marked `Approved` at the end of the saga.
However, in between, a `CancelOrder` saga executes to set its state to `Cancelled`.
As a consequence, the `CreateOrder` saga will overwrite the change from the `CancelOrder` saga,
the cancelled order still continues being processed within the system.

```d2
shape: sequence_diagram
o: Order Service
p: Payment Service
c: CancelOrder Saga
o -> o: "1. CreateOrder()"
o -> p: "2. ProcessPayment()" {
   class: error-conn
}
c -> o: Set state {
   style.bold: true
}
o {
   "state = Cancelled"
}
o -> o: Set state {
   class: error-conn
}
o {
   "state = Approved"
}
```

**Semantic Locking** is an effective approach for `Lost Update`.
Like before, we lock orders with the `Pending` state until they're successfully created.

```d2
shape: sequence_diagram
o: Order Service
p: Payment Service
c: CancelOrder Saga
o -> o: "1. CreateOrder()" {
   "state = Pending"
}
o -> p: "2. ProcessPayment()" {
   class: error-conn
}
c -> o: Fails to set state of a pending order {
   style.bold: true
}
o -> o: Set state {
   class: error-conn
}
o {
   "state = Approved"
}
```

**Semantic Locking** is nothing but a distributed lock,
reducing parallelism and negatively affecting the performance.
In the first place, we should design a reliable flow and limit locking instead.
