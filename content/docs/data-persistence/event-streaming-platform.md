---
title: Event Streaming Platform
weight: 50
---

{{< callout type="warning" >}}
Many concepts below inherit from the [Distributed Database topic](Distributed-Database.md),
please have a look at it before diving into this topic!
{{< /callout >}}

## Message Queue

We've used {{< term mq >}} for [decoupling a system](Microservice.md#service-decoupling) in the first topic.
When referring a {{< term mq >}},
we want to mention a shared message container between many processes.

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

However, we've left out its detailed implementation which includes two aspects:

### 1. Message Delivery

First, there are two common ways messages can be delivered to consumers.

#### Streaming

**Streaming** means messages should be consumed one-by-one immediately after being produced.
This workflow let the system quickly adapt and handle messages.

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
s: Service 
    class: server
}
e -> m1 -> s
e -> m2 -> s
```

#### Batching

Another delivery style is **Batching**,
meaning messages can be stacked up to be processed in large batches.
This results in far less resources consumptions,
as we don't need a *24/7* system to handle messages,
consumers can be lightweight short-lived processes.

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
s: Service 
    class: server
}
e -> m1 -> b
e -> m2 -> b
b -> s
```

### 2. Message Durability

Next, a {{< term mq >}} should show how message are persisted overtime.

For some solutions,
they choose to temporarily store messages and delete them after being consumed.

```d2
s1: Service 1 {
    class: server
}
m: Message Queue {
    class: mq
}
s2: Service 2 {
    class: server
}
s1 -> m: Produce a message
s2 <- m: Consume the message
m -> m: Delete the message 
```

This behavior reduces resources consumptions.
However, it's insensible for sensitive systems;
Messages are valuable,
they help keep track of what happened in a system.
Thus,
some solutions try to save messages durably in physical disks.

## Event Streaming Platform

{{< term esp >}} is a {{< term mq >}} implementation
in a distributed manner with high availability and fault-tolerance.

First, let's explain the term **Event**.

### Event

**Message** is a broad term referring to every piece of information transmitted within a system.
In common, messages can be categorized into:

1. **Command** is a request sent to the system.
2. **Event** refers to a fact happened in the past.

For example,
in a payment transaction:

- **Command**: the client starts the transaction with a command `InitiateTransaction`.
- **Event**: within the execution, the transaction can leave some events in the system,
such as `AccountBalanceChanged`, `TransactionCompleted`, `TransactionFailed`...

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

**Event** is used instead of **Message**,
based on a real usage pattern,
in fact, many systems opt for just storing events durably and ignoring commands.
We'll explain why in the [Event-driven Architecture topic](event-driven-architecture).

We've also explained **Streaming** above,
that means events should be delivered and handled immediately once they're created.
Theory enough!
Let's see how to build an {{< term esp >}}!
In the next section,
we'll use [**Apache Kafka**](https://kafka.apache.org/) as demonstration purposes,
this is the most common {{< esp >}} currently.

## Event Streaming Cluster

{{< term esp >}} works as a decentralized cluster revolving around many servers, usually called **brokers**.

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

{{< term kk >}} favors consistency over availability.
One of the brokers will be elected as the **Controller** (**Leader**) node by the [Raft algorithm](Peer-to-peer-Architecture.md#).

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

Topic is the concept grouping all events of the same type for simple management, e.g. `AccountCreated` topic will have following events

```yaml
AccountCreated:
- event1:
    accountId: acc1
    at: 00:01
- event2:
    accountId: acc2
    at: 00:02
```

Simply, you can imagine a topic as a {{< term sql >}} table containing events (rows) of the same type.

### Partition

Maintaining a topic on a single broker is inefficient.
Because the storage and access volume of each topic can vary, possibly creating imbalances between brokers

Hence, a topic is divided into smaller **partitions** distributed across brokers.
That reminds us of [sharding](Peer-to-peer-Architecture.md#shard) in a {{< term p2p >}} cluster.

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

Of course,
We need to replicate a partition to different brokers to guarantee data loss.

For example,
`Partition 1`, `Partition 2`, `Partition 3` are replicated to `Broker 3`, `Broker 1`, `Broker 2` respectively.

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

For simplicity, {{< term esp >}} basically supports two operations: **Produce** (Write) and **Consume** (Read).
there is no update or delete operations.
They typically organize events as **only-appended files** in brokers,
changing something in between will cause entries to move around.

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

Producing is nothing but appending data to the primary partition,
further synchronizing to replicas.

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

### In-sync Replica (ISR)

Periodically, replicas will fetch and compare with data from the primary partition.
This ensures a new (or just corrupted) replica can keep up with the latest data.

**In-sync Replicas (ISR)** refers to replicas that are in sync with the primary partition.
This is configured with a duration value, a replica is considered **out-of-sync**
if it falls behind (the last fetch) by more than the threshold.
For example, the ISR threshold is set as `2s`,  `S2` is supposed as **out-of-sync**.

```d2
l: Primary (Time = 00:05, ISP = 2s) {
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
a producing request contains an acknowledgement (**ACK**)
value defining the number of brokers (including the primary partition) an event will be saved before responding back to producers.

There are 3 types of **ACK** settings for producing:

- **ACK=0**: No save is requested. That means data loss is not a problem.

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

- **ACK=1**: Only the primary partition is requested. Data loss may occur if it goes down before replicating to any replica.

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
p -> r: 4. Replicate (data loss if fails here) {
    class: error-conn
}
```

- **ACK=ALL**: All **ISRs** are requested to save data. Data loss is evaded in case the primary partition is corrupted.

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

Keep in mind these patterns as they're extremely important in the [Delivery Semantics](#delivery-semantics) section.

Why does it use **ISRs** instead of normal replicas?
There is a high chance that an **out-of-sync** replica crashes or is partitioned,
waiting for it will slow down the partition, or even unavailable if the replica is actually down.
When the replica is in-sync again, it can fetch missed events from the primary partition.

## Consumer

**Event Streaming** usually implements [Long Polling](Communication-Protocols.md#long-polling) for event delivery,
as it helps decouple the service from consumers, increasing availability.

### Commit Offset

Callback to the data structure, **Offset** refers to the position of events in only-appended files.

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
    }
}
```

To prevent consumers from pulling duplicated events,
a partition needs to member the **offset** of the last consumed event of each consumer.

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

Consumers periodically pull new events from the primary partition,
handle them and **commit offset** (increase the last offset) to ignore consumed events.

```d2

shape: sequence_diagram
c: Consumer {
    class: client
}
q: Event {
    class: mq
}
c -> q: 1. Consumes
q -> c: 2. Return an event
c -> c: 3. Handle the event
c -> q: 4. Commit offset
```

### Consumer Group

A topic can be large with a lot of partitions,
letting a single consumer handle it is not a efficient solution.
Instead, we can build a consumer as a group,
different workers can read different partitions concurrently.

Consumers within a group use the same name to commit offset
to make sure the group can't read an event twice.

For example, `Group 1` consumes `Topic 1`.
Each consumer member tries to read a partition

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
}
q.t1.p1 -> c.cg1.c1
q.t1.p2 -> c.cg1.c2
```

`Group 2` consumes `Topic 2`.
The group has only 1 member, so, it must consume from all partitions of the topic.

```d2
grid-rows: 2
q: Topics {
    t2: Topic 2 {
        p1: Partition 1    
        p2: Partition 2    
    }
}
c: Consumers {
    cg2: Consumer Group 2 {
        c1: Consumer 1
    }
}
q.t2.p1 -> c.cg2.c1
q.t2.p2 -> c.cg2.c1
```

As a side node,
replicas are only used for data backup and recovery, consumers must work with the primary server.
Unlike other data stores, read operation generates no side effect;
In {{< term esp >}}, event consumption is [non-idempotent](API-Design.md#idempotency),
it will change consumer offset, and a replica is expected to not modify data.

## Delivery Semantics

There are many problems arisen when both committing in **Event Streaming** and another datasource.
These are two separate steps, errors can happen in between to make them mismatched.

For example, a consumer processes an event successfully,
makes changes in another store but crashing before committing offset.
When the consumer lives again, it will unexpectedly read the event again.

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

**Delivery Semantics** guarantee the delivery of events based on the collaboration producing and consumption.
There are 3 types of them offering tradeoffs between latency and durability.

### At-most-once Delivery

As the name suggests, this approach ensures events will be delivered **0** or 1 time

#### Producer {id="prod_amo"}

**Producer** sends events with **ACK=0** to get lowest latency.
Even the requests fail, they will be not retried to prevent data duplication.

For example, a producer can’t receive the response from the broker due to a network error.
If the producer tries to re-produce the event,
we'll encounter duplicated events.
Thus, we'll prohibit retrying, despite failures.

```d2
shape: sequence_diagram
p: Producer {
    class: server
}
q: Partition {
    class: mq
}
p -> q: Produce an event
q -> p: Respond but the producer can not receive {
    class: error-conn
} 
p --> q: Continue without retry {
    class: error-conn
}
```

#### Consumer {id="con_amo"}

In the consumer side, it commits events **before** handling it.
This helps prevent from handling an event twice in case the consumer fails to commit it.

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

For example,
a consumer succeeds handling an event but crashes before committing it.
If the consumer commits offset late,
it will unpextedly pulls the event again after recovering.

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

```

Thus, events should be committed before handling.

This approach ensures that events will be only delivered at most 1 time.
Moreover, it brings the lowest latency.
It's perfectly fit for cases which **data loss is acceptable**, e.g., metrics monitoring.

### At-least-once Delivery

This approach ensures events will be delivered once or **more times**.

#### Producer {id="prod_alo"}

**Producer** sends events with **ACK=1 or ALL** and **retry on failure** to ensure them to be persisted.
Although, as we've discussed, the retry behavior may cause data duplication.

```d2
shape: sequence_diagram
p: Producer {
    class: server
}
q: Broker {
    class: mq
}
p -> q: Produce an event
q -> p: Respond but the producer can not receive {
    class: error-conn
} 
p --> q: Retry to produce the event (duplicated) {
    class: error-conn
}
```

#### Consumer {id="con_alo"}

**Consumer** commits events **after** handling it.
That means an event can be consumed more than once if the consumer fails to commit it.

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

This approach ensures data durability but possibly encountering data duplication,
the performance is also lower than **At-most-once Delivery**.
Consider this model if data duplication is acceptable or **resolvable**, e.g., user activity tracking,...

### Exactly-once Delivery

Events are delivered exactly once, this is the most expected and hardest setup.
Unfortunately, this model is not supported natively by {{< term esp >}} solutions,
we need to equip the system with more techniques to achieve it.

#### Exactly-once Producer

**Producer** works similarly to **At-least-once Delivery** with **ACK=ALL** for reliable durability.

To prevent duplication, it will assign idempotency keys to producers:

- A producer is assigned a **PID** (producer id) and a **seq** (sequential number).
The producer increases the number **locally** after receiving an producing acknowledgement.
- Partitions maintain pairs of **(PID, seq)** to ignore duplicates.

For example, a producer produces an event but
fails to receive the respective acknowledgement due to a network error.
It then retries to produce the event again,
but the event is ignored as its **seq** is outdated.

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
p -> sp: Publish an event (seq = 1)
sp -> p: Ignore (seq = 2 > event seq = 1)
p {
   "seq = 2" 
}
```

#### Exactly-once Consumer

{{< term esp >}} is unable to resolve the duplicated consumption problem itself when a consumer fails to commit,
because it has no idea whether the consumer has failed or succeeded.
Resolving the issue needs additional workloads,
typically, we have 3 solutions.

##### 1. Consume-process-produce Pipeline

For a specific scenario when an event is transformed to another event and doesn’t make changes to external datasets,
{{< term esp >}} can help natively.

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

{{< term esp >}} can implement **Transactional Commit**.
An event has 2 states: **committed** and **uncommitted**, by default, consumers can only see committed events.

Any failure between will abort the entire transaction and make uncommitted changes invalidated.

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

Obviously, it's not exactly-once delivery.
This approach ensures that an event only generates a result (the associated event),
if the consumer fails in between, it will not produce duplicated events.
However, it's only possible when the even processing generates no side effect.

##### Two-phase Commit

In [the distributed transaction](Low-level-Protocols.md#two-phase-commit-2pc) topic,
we will talk about commiting changes in multiple stores concurrently.

In short,
we leverage `Two-phase commit` to make sure that the offset and changes (in another store) are both committed

```d2

shape: sequence_diagram
c: Consumer {
    class: client
}
q: Event Streaming  {
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

Two-phase commit comes with the problem of infinite blocking (we will explain in depth in [the corresponding topic](Low-level-Protocols.md#two-phase-commit-2pc)).
Sometimes, this feature is even not supported.  

##### Event Idempotency

Similar to [Request Idempotency](API-Design.md#idempotency), this method **requires** events to have a unique id.
In other words, this supports `exactly-once delivery` by filtering duplication on the consumer side

For example, we store completed events persistently,
then we check the existence of events (by `id`) to consider continuing processing them

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
q -> c: Pull and handle "event-1"
c -> d: Save "event-1" to the database {
    style.bold: true
}
c -- c: The consumer is corrupted here, the event won't be committed {
    class: error-conn
}
c -> c: Recovers {
    style.bold: true
}
q -> c: Pull "event-1" because it wasn't committed
c -> d: Verify and see that "event-1" was consumed
c -> c: Ignores the event {
    style.bold: true
}
```

This approach is preferred than `Two-phase commit` because of its simplicity,
frequently combined with the [Saga](Compensating-Protocols.md#saga) distributed pattern  

We see that there is no real exactly-once delivery when working with external systems,
you may equip additional handlers to achieve it.
`Exactly-once` delivery is perfectly fit for critical systems which both data loss and duplication are unacceptable,
e.g., banking systems
