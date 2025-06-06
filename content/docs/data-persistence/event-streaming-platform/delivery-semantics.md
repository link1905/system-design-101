---
title: Delivery Semantics
weight: 10
---

Numerous challenges arise when committing changes to both {{< term esp >}}
and other data sources simultaneously.
These two steps are independent, and failures between them can lead to inconsistencies.

For example, a consumer might successfully process an event and apply changes to another data store but crash before committing the event offset.
Upon recovery, the consumer may reprocess the same event unexpectedly.

```d2
shape: sequence_diagram
c: Consumer {
  class: server
}
d: Store {
  class: db
}
e: Partition {
  class: mq
}
c <- e: Pull an event
c -> d: Make changes
c -> c: Crash and cannot commit offset {
  class: error-conn
}
c -- e {
  style.stroke-dash: 3
}
c -> c: Recover
c <- e: Pull and process the event again {
  class: error-conn
}
```

**Delivery semantics** define the guarantees provided for event delivery during production and consumption.
There are three main types, each offering different trade-offs between latency, durability, and reliability.

## At-most-once Delivery

This delivery model ensures that an event is delivered **zero or one time**.

### Producer {id="prod_amo"}

The **producer** sends events with **ACK=0** to achieve the lowest latency.
Even if a request fails, it won't be retried to avoid duplication.

For example,
if the producer doesn’t receive a response from the broker due to a network error and retries the operation,
duplicated events may occur.

```d2
shape: sequence_diagram
p: Producer {
  class: server
}
q: Partition {
  class: mq
}
p -> q: Produce an event
q -> p: Respond but the producer cannot receive {
  class: error-conn
}
p -> p: Timeout
p -> q: Retry to produce the event again {
  class: error-conn
}
```

To prevent this, retries are disabled, even in failure scenarios.

```d2
shape: sequence_diagram
p: Producer {
  class: server
}
q: Partition {
  class: mq
}
p -> q: Produce an event
q -> p: Respond but the producer cannot receive {
  class: error-conn
}
p -> q: Continue without retry {
  style.bold: true
}
```

### Consumer {id="con_amo"}

The **consumer** commits the event **before** handling it. This approach ensures the event won't be processed more than once if the consumer crashes before committing.

For instance, if a consumer processes an event successfully but crashes before committing, and the commit is delayed, the event will be reprocessed after recovery.

```d2
shape: sequence_diagram
c: Consumer {
    class: client
}
q: Partition {
    class: mq
}
c -> q: Consume
q -> c: Return an event
c -> c: Handle the event
c -> c: Crash and fail to commit offset {
  class: error-conn
}
c -- q {
  style.stroke-dash: 3
}
c -> c: Recover
c -> q: Consume the event again {
  class: error-conn
}
```

Thus, events must be committed **before** they are processed.

```d2
shape: sequence_diagram
c: Consumer {
    class: client
}
q: Partition {
    class: mq
}
c -> q: Consume
q -> c: Return an event
c -> q: Commit offset immediately {
    style.bold: true
}
c -> c: Handle the event
```

This model guarantees delivery at most once and offers the lowest latency.
It is suitable for scenarios where **data loss is acceptable**, such as metrics collection.

## At-least-once Delivery

This model guarantees that every event is delivered **at least once**, possibly more.

### Producer {id="prod_alo"}

The **producer** uses **ACK=1 or ALL** and enables retries on failure to ensure event persistence.
However, enabling retries can lead to duplicate events.

```d2
shape: sequence_diagram
p: Producer {
    class: server
}
q: Partition {
    class: mq
}
p -> q: Produce an event
q -> p: Respond but the producer cannot receive {
    class: error-conn
}
p --> q: Retry to produce the event (duplicated) {
  style.bold: true
}
```

### Consumer {id="con_alo"}

The **consumer** commits the offset **after** processing the event.

```d2
shape: sequence_diagram
c: Consumer {
    class: server
}
e: Partition {
    class: mq
}
c <- e: Pull an event
c -> c: Process the event
c -> e: Commit offset
```

If it crashes before committing, the event may be reprocessed.

```d2
shape: sequence_diagram
c: Consumer {
    class: server
}
e: Partition {
    class: mq
}
c <- e: Pull an event
c -> c: Process the event
c -> c: Crash and cannot commit offset {
    class: error-conn
}
c -- e {
  style.stroke-dash: 3
}
c -> c: Recover
c <- e: Pull the event again {
    style.bold: true
}
```

This method provides stronger durability guarantees but may lead to duplicates and higher latency compared to **at-most-once** delivery.
It is well-suited for scenarios where duplicate data is acceptable or can be handled, such as in user activity tracking.

## Exactly-once Delivery

This is the most reliable but also the most complex delivery model, ensuring each event is delivered **exactly once**.
{{< term esp >}} solutions don't fully support this out of the box, so additional techniques are required.

### Exactly-once Producer

The producer functions similarly to the **at-least-once** model, allowing retries on failures and using **ACK=ALL** for durability.

To prevent duplication, it uses idempotency keys.
Each producer is assigned a **PID (producer ID)** and a **seq (sequence number)**, which it increments locally after receiving an acknowledgment.

Example:

```d2
shape: sequence_diagram
p: Producer (P1)
sp: Partition
p {
  "seq = 0"
}
p -> sp: Register
sp {
  "PID = P1, seq = 0"
}
p -> sp: Produce an event (seq = 0)
sp {
  "PID = P1, seq = 0 -> 1"
}
sp -> p: Respond
p {
  "seq = 0 -> 1"
}
```

The partition ignores any events with outdated sequence numbers, effectively preventing duplicates.

For example, if `P1` sends an event but fails to receive an acknowledgment, it will resend the event.
Because the event’s sequence number is outdated, the partition ignores it, avoiding duplication.

```d2
shape: sequence_diagram
p: Producer (P1)
sp: Partition
p {
  "seq = 1"
}
sp {
  "PID = P1, seq = 1"
}
p -> sp: Produce an event (seq = 1)
sp {
  "PID = P1, seq = 1 -> 2"
}
sp -> p: Fail to receive the acknowledgement {
  class: error-conn
}
p -> sp: Retry to produce the event (seq = 1)
sp -> p: Ignore (producer seq = 2 > event seq = 1) {
  style.bold: true
}
p {
  "seq = 1 -> 2"
}
```

### Exactly-once Consumer

{{< term esp >}} cannot tell whether a consumer has processed an event or not.
To achieve exactly-once semantics, we must introduce one of the following approaches:

#### 1. Consume–Process–Produce Pipeline

If the event is simply transformed into another event (with no external side effects), {{< term esp >}} can manage this flow.

```d2
grid-rows: 1
horizontal-gap: 100
p: Partition {
  class: mq
}
e {
  class: none
  grid-rows: 2
  o: Event {
    class: msg
  }
  t: New event {
    class: msg
  }
}
c: Consumer {
  class: server
}
p -- e.o
e.o -> c: Consume
c -> e.t: Transform to a new event
e.t -> p
```

##### Transactional Commit

To support this, **Transactional Commit** is implemented, ensuring that consumers can only see committed events.
If a failure occurs during processing, the transaction is aborted and all uncommitted changes are discarded.

```d2
shape: sequence_diagram
c: Consumer {
    class: client
}
q: Partition {
    class: mq
}
c -> q: 1. Begin a transaction {
    style.bold: true
}
q -> c: 2. Consume an event (Uncommitted)
c -> c: 3. Handle the event
c -> q: 4. Produce a new event (Uncommitted)
c -> q: 5. Commit the transaction {
    style.bold: true
}
```

This guarantees the transformation process won't produce duplicates.
However, it only applies to stateless processing with no external side effects.

#### 2. Two-phase Commit

To synchronize multiple systems, we can use [{{< term 2pc >}}]({{< ref "low-level-protocols#two-phase-commit-2pc" >}}):

```d2
shape: sequence_diagram
c: Consumer {
    class: client
}
q: Partition {
    class: mq
}
d: External Store {
    class: db
}
c -> d: PREPARE {
    style.bold: true
}
c -> q: PREPARE {
    style.bold: true
}
c -> q: Consume an event
c -> d: Make changes
d -> c: VOTE YES
q -> c: VOTE YES
c -> d: COMMIT {
    style.bold: true
}
c -> q: COMMIT {
    style.bold: true
}
```

However, {{< term 2pc >}} introduces the risk of infinite blocking and is not always supported.
This is discussed further in [this topic]({{< ref "low-level-protocols#two-phase-commit-2pc" >}}).

#### 3. Event Idempotency

This approach requires each event to carry a **unique identifier**.
Consumers then check for this ID before processing.

For example, a consumer verifies whether an event has already been processed before handling it.

```d2
shape: sequence_diagram
c: Consumer {
    class: client
}
q: Event Streaming {
    class: mq
}
d: External Store {
    class: db
}
q -> c: Pull "event-1"
c -> d: Handle and save "event-1" {
    style.bold: true
}
c -- c: The consumer crashes, offset not committed {
    class: error-conn
}
c -> c: Recover {
    style.bold: true
}
q -> c: Pull "event-1" again
c -> d: Check and find "event-1" already stored
c -> q: Commit the event {
    style.bold: true
}
```

This approach is often preferred over {{< term 2pc >}} due to its simplicity and better fault tolerance.
It is commonly used in combination with the [Saga]({{< ref "compensating-protocols" >}})
pattern to manage long-running, distributed operations.

However, this method does not provide **atomicity** across the system,
there can be consistency drift between the streaming platform and the external data store.
We’ll explore this limitation in more detail in the [Distributed Transaction]({{< ref "distributed-transaction" >}}) topic.

Ultimately, **exactly-once** delivery is difficult to guarantee when external systems are involved.
It requires additional mechanisms and is best suited for **mission-critical systems** where both data loss and duplication are unacceptable (e.g., banking platforms).
