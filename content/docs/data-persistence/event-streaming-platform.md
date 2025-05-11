---
title: Event Streaming Platform
weight: 50
---

{{< callout type="warning" >}}
Many of the concepts discussed below build upon the [Distributed Database topic](Distributed-Database.md).
Please review it before diving into this section.
{{< /callout >}}

## Message Queue

We've previously used a {{< term mq >}} to illustrate [system decoupling](Microservice.md#service-decoupling) in microservices.
In general, a {{< term mq >}} refers to a shared message container that facilitates communication between multiple processes.

```d2
s1: Server 1 {
    class: server
}
s2: Server 2 {
    class: server
}
m: Message Queue {
    class: mq
}
s1 <-> m
s2 <-> m
```

However, we haven’t yet explored the detailed mechanics of a message queue. It typically involves two key aspects:

### 1. Message Delivery

There are two common strategies for delivering messages to consumers:

#### Streaming

**Streaming** delivery means messages are consumed one by one, immediately after being produced.
This approach allows systems to react and process events as soon as possible.

```d2
direction: right
e: Event Stream {
    class: mq
}
m1: Message 1 {
    class: msg
}
m2: Message 2 {
    class: msg
}
s: Service {
    class: server
}
e -> m1 -> s
e -> m2 -> s
```

#### Batching

In **Batching**, messages are accumulated and processed together in groups.
This approach significantly reduces resource consumption,
as it removes the need for continuously running services.
Consumers can instead be lightweight,
short-lived processes triggered on demand or at scheduled intervals.

```d2
direction: right
e: Event Stream {
    class: mq
}
m1: Message 1 {
    class: msg
}
m2: Message 2 {
    class: msg
}
b: Batch {
    class: batch
}
s: Service {
    class: server
}
e -> m1 -> b
e -> m2 -> b
b -> s
```

### 2. Message Durability

Another critical aspect of {{< term mq >}} is **message durability**—how messages are stored over time.

Some implementations store messages temporarily and delete them once they are consumed:

```d2
direction: right
s1: Service 1 {
    class: server
}
m: Message Queue {
    class: mq
}
s2: Service 2 {
    class: server
}
s1 -> m: 1. Produce a message
s2 <- m: 2. Consume the message
m -> m: 3. Delete the message
```

While this behavior is resource-efficient,
it's not suitable for systems requiring high reliability or audit trails.
In such cases, messages are considered valuable—they record what occurred within the system.
To address this, more robust solutions persist messages durably on physical storage,
ensuring they are retained even after being consumed.

## Event Streaming Platform

An {{< term esp >}} is a distributed implementation of a {{< term mq >}},
designed to offer high availability and fault tolerance.

Before diving deeper, let’s first clarify the concept of an **Event**.

### Event

The term **Message** broadly refers to any piece of information exchanged within a system.
Messages generally fall into two main categories:

1. **Command** – A directive sent to the system, requesting it to perform a specific action.
2. **Event** – A record of something that has already occurred.

Let’s consider a payment transaction as an example:

* **Command**: The client begins the transaction by issuing a command such as `InitiateTransaction`.
* **Event**: As the system processes the transaction, it generates events like `AccountBalanceChanged`, `TransactionCompleted`, or `TransactionFailed`.

```d2
c: Client {
    class: client
}
s: Payment Service {
    class: server
}
m1: AccountBalanceChanged {
    class: msg
}
m2: TransactionCompleted {
    class: msg
}
m3: TransactionFailed {
    class: msg
}
c -> s: "InitiateTransaction"
s -> m1
s -> m2
s -> m3
```

In modern systems, the term **Event** is often favored over **Message**,
as it better reflects real-world usage.
Many architectures focus on durably storing events and may even bypass command persistence entirely.
We'll explore the reasons behind this in the [Event-driven Architecture](event-driven-architecture) section.

We’ve also touched on the concept of **Streaming**,
which emphasizes the immediate delivery and processing of events as soon as they are produced.

Let’s explore how to build an {{< term esp >}}.
In the following section, we’ll draw heavily on concepts from **Apache Kafka**—
the most widely used {{< term esp >}} in the industry today.

## Event Streaming Cluster

An {{< term esp >}} operates as a decentralized cluster composed of multiple servers,
commonly referred to as **brokers**.

```d2
b1: Broker 1 {
    class: server
}
b2: Broker 2 {
    class: server
}
b3: Broker 3 {
    class: server
}
b1 <-> b2 <-> b3 <-> b1
```

For example, {{< term kk >}} prioritizes **consistency** over **availability**.
One broker is elected as the **Controller** (or **Leader**) node using the [Raft algorithm](Peer-to-peer-Architecture.md#).

```d2
b1: Broker 1 (Controller) {
    class: server
}
b2: Broker 2 (Follower) {
    class: server
}
b3: Broker 3 (Follower) {
    class: server
}
b1 -> b2: Replicate logs {
    style.animated: true
}
b1 -> b3: Replicate logs {
    style.animated: true
}
```

### Topic

A **Topic** is a logical grouping of events of the same type, simplifying event organization and management.
For example, a topic named `AccountCreated` might contain events like the following:

```yaml
AccountCreated:
  - event1:
      accountId: acc1
      email: mylovelyemail@mail.com
      at: 00:01
  - event2:
      accountId: acc2
      email: nottoday@mail.com
      at: 00:02
```

> You can think of a topic as a {{< term sql >}} table, where each row represents a single event of the same type.

### Partition

Storing an entire topic on a single broker is inefficient.
The storage and access load for each topic may vary significantly,
potentially causing uneven load distribution across brokers.

To address this, topics are split into smaller units called **partitions**, which are distributed across brokers.
This concept is similar to [sharding](Peer-to-peer-Architecture.md#shard) in a {{< term p2p >}} cluster.

```d2
classes: {
  part: {
    width: 200
  }
}
grid-columns: 1
t: Topic {
  direction: right
  width: 600
  grid-gap: 0
  grid-columns: 3
  p1: Partition 1 {
    class: part
  }
  p2: Partition 2 {
    class: part
  }
  p3: Partition 3 {
    class: part
  }
}
sv: "" {
  direction: right
  grid-columns: 3
  s1: Broker 1 {
    class: db
  }
  s2: Broker 2 {
    class: db
  }
  s3: Broker 3 {
    class: db
  }
}

t.p1 -> sv.s1
t.p2 -> sv.s2
t.p3 -> sv.s3
```

### Partition Replication

To ensure fault tolerance and prevent data loss, partitions must be replicated across multiple brokers.

For instance, `Partition 1`, `Partition 2`, and `Partition 3` are each assigned to a primary broker and replicated to another broker:

```d2
classes: {
  part: {
      width: 335
  }
}
grid-columns: 1
db: Topic {
  direction: right
  grid-gap: 0
  grid-columns: 3
  p1: Partition 1 {
    class: part
  }
  p2: Partition 2 {
    class: part
  }
  p3: Partition 3 {
    class: part
  }
}

peer: Cluster {
  s1: "Broker 1" {
    grid-gap: 50
    grid-columns: 1
    p1: Partition 1 primary
    p2: Partition 2 replica
  }
  s2: "Broker 2" {
    grid-gap: 50
    grid-columns: 1
    p1: Partition 2 primary
    p2: Partition 3 replica
  }
  s3: "Broker 3" {
    grid-gap: 50
    grid-columns: 1
    p1: Partition 3 primary
    p2: Partition 1 replica
  }
  s1.p1 -> s3.p2: Replicate {
    style.animated: true
  }
  s2.p1 -> s1.p2: Replicate {
    style.animated: true
  }
  s3.p1 -> s2.p2: Replicate {
    style.animated: true
  }
}
db.p1 -> peer.s1.p1
db.p2 -> peer.s2.p1
db.p3 -> peer.s3.p1
```

## Data Structure

At a fundamental level, an {{< term esp >}} supports two operations: **Produce** (write) and **Consume** (read).
There are no update or delete operations.

Events are stored as **append-only files** on brokers.
Modifying an event in the middle of the log would require shifting all subsequent entries,
which is inefficient and generally avoided.

```d2
b: Broker {
    grid-rows: 1
    grid-gap: 0
    f1: "AccountCreated.events" {
        grid-columns: 1
        grid-gap: 0
        e1: |||yaml
        accountId: acc01
        at: 00:01
        |||
        e2: |||yaml
        accountId: acc02
        at: 00:02
        |||
        e3: "...(space for new events)..."
    }
    f2: "BalanceChanged.events" {
        grid-columns: 1
        grid-gap: 0
        e1: |||yaml
        accountId: acc01
        balance: 100
        |||
        e2: |||yaml
        accountId: acc01
        balance: 50
        |||
        e3: |||yaml
        accountId: acc02
        balance: 30
        |||
        e4: "..."
    }
}
```

## Producing

Producing simply means appending data to the primary partition and subsequently synchronizing it to the replicas.

```d2
direction: right
mq: MQ cluster {
    b1: Broker 1 (Primary)
    b2: Broker 2 (Replica)
}
c: Client {
    class: client
}
c -> mq.b1: 1. Produce an event
mq.b1 -> mq.b2: 2. Replicate {
    style.animated: true
}
```

### In-Sync Replica (ISR)

Replicas periodically fetch and compare data from the primary partition.
This ensures that any newly added or previously corrupted replicas can catch up with the latest data.

**In-Sync Replicas (ISR)** are those replicas currently in sync with the primary partition.
This is governed by a configurable time threshold. If a replica's last fetch exceeds the threshold, it is considered **out-of-sync**.
For example, with an ISR threshold of 2 seconds, a replica that fetched data last at 00:02 while the primary is at 00:05 is **out-of-sync**.

```d2
l: Primary (Time = 00:05, ISR Threshold = 2s) {
    shape: server
}
b: Replica 1 (LastFetch = 00:04) {
    shape: server
}
c: Replica 2 (LastFetch = 00:02) {
    shape: generic-error
}
```

Similar to [Quorum-based consistency](Distributed-Database.md#quorum-based-consistency),
a produce request includes an acknowledgement (**ACK**) setting.
This determines how many brokers (including the primary) must successfully save the event before the producer receives a response.
Please keep **ACK** in mind, this mechanism is crucial for understanding [Delivery Semantics](#delivery-semantics).

There are three **ACK** levels:

- **ACK=0**: No replication is required. This allows for the lowest latency but risks data loss.

```d2
shape: sequence_diagram
s: Producer {
    class: server
}
p: Primary {
    class: server
}
s -> p: 1. Produce (ACK = 0)
p -> s: 2. Respond
p -> p: 3. Save
```

- **ACK=1**: Only the primary must save the data. If the primary fails before replication, data may be lost.

```d2
shape: sequence_diagram
s: Producer {
    class: server
}
p: Primary {
    class: server
}
r: Replica {
    class: server
}
s -> p: 1. Produce (ACK = 1)
p -> p: 2. Save
p -> s: 3. Respond
p -> r: 4. Replicate (data loss if failure occurs here) {
    class: error-conn
}
```

- **ACK=ALL**: All **ISRs** must save the data. This ensures no data loss even if the primary fails.

```d2
shape: sequence_diagram
s: Producer {
    class: server
}
p: Primary {
    class: server
}
r1: In-sync Replica {
    class: server
}
r2: Out-of-sync Replica {
    class: server
}
s -> p: 1. Produce (ACK = ALL)
p -> p: 2. Save
p -> r1: 3. Replicate
p -> s: 4. Respond
```

Why use **ISRs** instead of all replicas?
**Out-of-sync replicas** might be slow or unavailable due to crashes or partitioning.
Waiting for them can degrade performance or block the partition entirely. Once a replica becomes in-sync again, it will fetch the missed events from the primary.

## Consumer

**Event Streaming** typically uses [Long Polling](Communication-Protocols.md#long-polling) for event delivery.
This approach decouples producers and consumers, improving system availability.

### Commit Offset

In streaming systems, **Offset** refers to the position of an event in an append-only log.

```d2
b: Partition {
    grid-rows: 1
    grid-gap: 0
    f1: "AccountCreated.events" {
        grid-columns: 1
        grid-gap: 0
        e1: |||yaml
        offset: 0
        accountId: acc01
        |||
        e2: |||yaml
        offset: 1
        accountId: acc02
        |||
        e3: |||yaml
        offset: 2
        accountId: acc10
        |||
    }
}
```

To avoid processing the same event multiple times,
each partition tracks the **last consumed offset** for every consumer.

```d2
b: Partition {
    c: "Offsets" {
        c: |||yaml
        consumer1:
            lastOffset: 0
        consumer2:
            lastOffset: 1
        |||
    }
    f1: "AccountCreated.events" {
        grid-columns: 1
        grid-gap: 0
        e1: |||yaml
        offset: 0
        accountId: acc01
        |||
        e2: |||yaml
        offset: 1
        accountId: acc02
        |||
    }
}
```

Consumers periodically fetch new events from the primary, handle them, and then **commit/increase the offset** to avoid reprocessing.

```d2
shape: sequence_diagram
c: Consumer {
    class: client
}
q: Event {
    class: mq
}
c -> q: 1. Consume
q -> c: 2. Return an event
c -> c: 3. Handle the event
c -> q: 4. Commit offset
```

### Consumer Group

A topic may have multiple partitions, making it inefficient for a single consumer to handle alone.
Instead, a **consumer group** allows multiple consumers to read different partitions in parallel.

All consumers in a group share the same name and commit offset collectively, ensuring each event is processed only once by the group.

For example, in `Group 1` each consumer handles a separate partition from `Topic 1`.
In `Group 2` there is only one consumer, so it handles all partitions of the topic.

```d2
grid-rows: 2
q: Topics {
    t1: Topic 1 {
        p1: Partition 1
        p2: Partition 2
    }
}
c: Consumers {
    cg1: Consumer Group 1 {
        c1: Consumer 1
        c2: Consumer 2
    }
    cg2: Consumer Group 2 {
        c1: Consumer 1
    }
}
q.t1.p1 -> c.cg1.c1
q.t1.p2 -> c.cg1.c2
q.t1.p1 -> c.cg2.c1
q.t1.p2 -> c.cg2.c1
```

**Note:** Replicas are only used for backup and recovery. Consumers must always read from the primary.
Unlike traditional databases, read operations leave the system unchanged.
In {{< term esp >}}, consumption is [non-idempotent](API-Design.md#idempotency) because it changes the consumer’s offset.

## Delivery Semantics

Numerous challenges arise when committing changes to both **Event Streaming** platforms
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
e: Event Streaming Platform {
    class: mq
}
c <- e: Pull an event
c -> c: Process the event
c -> d: Make changes
c -> c: Crash and cannot commit offset {
    class: error-conn
}
c -- e: "..."
c -> c: Recover
c <- e: Pull and process the event again {
    style.bold: true
}
```

**Delivery semantics** define the guarantees around event delivery during production and consumption. There are three main types, each offering trade-offs between latency, durability, and reliability.

### At-most-once Delivery

### At-most-once Delivery

This delivery model ensures that an event is delivered **zero or one time**.

#### Producer {id="prod_amo"}

The **producer** sends events with **ACK=0** to achieve the lowest latency.
Even if a request fails, it won't be retried to avoid duplication.

For example,
if the producer doesn’t receive a response from the broker due to a network error and retries the operation,
duplicated events may occur.
To prevent this, retries are disabled—even in failure scenarios.

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
p --> q: Continue without retry {
    class: error-conn
}
```

#### Consumer {id="con_amo"}

The **consumer** commits the event **before** handling it. This approach ensures the event won't be processed more than once if the consumer crashes before committing.

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
c -> q: Commit the last offset {
    style.bold: true
}
c -> c: Handle the event
```

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
    class: generic-error
}
c -- c: Recover
c -> q: Consume the event again {
    class: generic-error
}
```

Thus, events must be committed **before** they are processed.

This model guarantees delivery **at most once** and offers the lowest latency.
It is suitable for scenarios where **data loss is acceptable**, such as metrics collection.

### At-least-once Delivery

This model guarantees that every event is delivered **at least once**, possibly more.

#### Producer {id="prod_alo"}

The **producer** uses **ACK=1 or ALL** and enables retries on failure to ensure event persistence.
However, retries may result in duplicate events.

```d2
shape: sequence_diagram
p: Producer {
    class: server
}
q: Broker {
    class: mq
}
p -> q: Produce an event
q -> p: Respond but the producer cannot receive {
    class: error-conn
}
p --> q: Retry to produce the event (duplicated) {
    class: error-conn
}
```

#### Consumer {id="con_alo"}

The **consumer** commits the offset **after** processing. If it crashes before committing, the event may be reprocessed.

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
c -> c: Process the event
c -> d: Make changes
c -> c: Crash and cannot commit offset {
    class: error-conn
}
c -- e: "..."
c -> c: Recover
c <- e: Pull the event again {
    style.bold: true
}
```

This method offers stronger durability guarantees but may result in duplicate data and reduced performance compared to **at-most-once** delivery.
It's a good fit when **data duplication is acceptable or resolvable**, such as in user activity tracking.

### Exactly-once Delivery

This is the most reliable but also the most complex delivery model, ensuring each event is delivered **exactly once**.
Native {{< term esp >}} solutions don't fully support this out of the box, so additional techniques are required.

#### Exactly-once Producer

The producer works similarly to the **at-least-once** model, using **ACK=ALL**. To prevent duplication, it uses **idempotency keys**:

* Each producer is assigned a **PID** (producer ID) and a **seq** (sequence number), which it increments locally after receiving an acknowledgment.
* The broker tracks the `(PID, seq)` pair to detect and ignore duplicates.

Example:

```d2
shape: sequence_diagram
p: Producer {
    class: server
}
sp: Partition {
    class: mq
}
p {
    "seq = 0"
}
p -- sp: Register
sp {
    "PID = P1, seq = 0"
}
p -> sp: Produce an event (seq = 0)
sp {
    "PID = P1, seq = 1"
}
sp -> p: Respond
p {
    "seq = 1"
}
p -> sp: Produce an event (seq = 1)
sp {
    "PID = P1, seq = 2"
}
sp -> p: Fail to respond {
    class: error-conn
}
p -> sp: Retry (seq = 1)
sp -> p: Ignore (seq = 2 > event seq = 1) {
  style.bold: true
}
p {
   "seq = 2"
}
```

#### Exactly-once Consumer

{{< term esp >}} cannot distinguish whether a consumer has processed an event if it crashes before committing. To achieve exactly-once semantics, we must introduce one of the following approaches:

##### 1. Consume–Process–Produce Pipeline

If the event is simply transformed into another event (with no external side effects), {{< term esp >}} can manage this flow.

```d2
direction: right
sp: Event Streaming Platform {
    class: mq
}
o: Original event {
    class: event
}
c: Consumer {
    class: server
}
t: New event {
    class: event
}
sp -- o
o -> c: 1. Pull the event
c -- t: 2. Transform to a new event
t -> sp: 3. Produce new event
```

###### Transactional Commit

{{< term esp >}} supports **Transactional Commit**, where consumers can only see **committed** events.
If a failure occurs mid-process, the transaction is aborted, and all uncommitted changes are discarded.

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

This guarantees the **transformation process** won't produce duplicates.
However, it only applies to stateless processing with no external side effects.

##### Two-phase Commit

To synchronize multiple systems, we can use [{{< term 2pc >}}](Low-level-Protocols.md#two-phase-commit-2pc):

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
c -> d: PREPARE {
    style.bold: true
}
c -> q: PREPARE {
    style.bold: true
}
q -> c: Pull an event
c -> d: Handle the event and make changes
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
This is discussed further in [this topic](Low-level-Protocols.md#two-phase-commit-2pc).

##### Event Idempotency

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

This approach is often preferred over {{< term 2pc >}} due to its relative simplicity and better fault tolerance.
It is commonly used in combination with the [Saga](Compensation-Protocol.md) pattern to manage long-running, distributed operations.

However, this method does not provide full atomicity across the system—
there can be consistency drift between the streaming platform and the underlying data store.
We’ll explore this limitation in more detail in the [Distributed Transaction](Distributed-Database.md) section.

Ultimately, **exactly-once** delivery is difficult to guarantee when external systems are involved.
It requires additional mechanisms and is best suited for **mission-critical systems** where both data loss and duplication are unacceptable (e.g., banking platforms).
