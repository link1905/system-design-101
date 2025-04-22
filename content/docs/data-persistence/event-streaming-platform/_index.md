---
title: Event Streaming Platform
weight: 50
---

## Messaging

We've previously used {{< term msg >}} to illustrate [system decoupling]({{< ref "microservice#service-decoupling" >}}) in the first topic.

Typically, **{{< term msg >}}** refers to a shared message channel that facilitates communication between multiple processes or services.

```d2
grid-rows: 1
grid-gap: 100
s1: Service 1 {
    class: server
}
m: Message Channel {
    class: mq
}
s2: Service 2 {
    class: server
}
s1 <-> m
s2 <-> m
```

However, we haven’t yet explored how to actually implement {{< term msg >}}.

In this section, let's look at two fundamental aspects of **{{< term msg >}}**:

### 1. Message Delivery

There are two common strategies for delivering messages to consumers:

#### Streaming

**Streaming** means messages are consumed one by one, immediately after being produced.
This approach enables systems to react and process events as soon as they occur.

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
e -- m1 -> s
e -- m2 -> s
```

#### Batching

In **batching**, messages are accumulated and processed together in groups.
This can significantly reduce resource usage, since it removes the need for services to run continuously.
Consumers can instead be lightweight, short-lived processes triggered on demand or at specific intervals.

```d2
direction: right
e: Event Stream {
  class: mq
}
b: Batch {
  m1: Message 1 {
    class: msg
  }
  m2: Message 2 {
    class: msg
  }
}
s: Service {
  class: server
}
e -- b -> s
```

### 2. Message Durability

Another critical aspect of {{< term msg >}} is **message durability**, which describes how messages are stored and retained over time.

#### Message Queuing

The first approach uses a message **queue**, usually based on the classic first-in, first-out ([Queue Data Structure](https://www.geeksforgeeks.org/queue-data-structure/)).
Messages are temporarily stored and removed once they are consumed. For instance, messages can be sequentially consumed by services as shown below:

```d2
grid-columns: 1
m1: {
  class: none
  grid-rows: 1
  horizontal-gap: 100
  s: Service 1 {
    class: server
  }
  m: Message Channel {
    grid-rows: 1
    grid-gap: 0
    m1: Message 1 {
      class: msg
    }
    m2: Message 2 {
      class: msg
    }
    m3: Message 3 {
      class: msg
    }
  }
  s <- m.m1: Consume
}
m2 {
  class: none
  grid-rows: 1
  horizontal-gap: 100
  vertical-gap: 0
  s: Service 2 {
    class: server
  }
  m: Message Channel {
    grid-rows: 1
    grid-gap: 0
    m2: Message 2 {
      class: msg
    }
    m3: Message 3 {
      class: msg
    }
  }
  s <- m.m2: Consume
}
m3 {
  class: none
  grid-rows: 1
  horizontal-gap: 100
  vertical-gap: 0
  s: Service 1 {
    class: server
  }
  m: Message Channel {
    grid-rows: 1
    grid-gap: 0
    m3: Message 3 {
      class: msg
    }
  }
  s <- m.m3: Consume
}
```

While this is efficient in terms of resource usage, it isn’t suitable for systems that require high reliability or audit trails.
In those scenarios, messages are often considered valuable records of what occurred within the system.

#### Message Persistence

To address this, more robust solutions persist messages durably on storage—ensuring they are retained even after being consumed.
A key feature is that a single message can be consumed by multiple consumers.

```d2
grid-columns: 1
m1: {
  class: none
  grid-rows: 2
  s: {
    class: none
    grid-rows: 1
    s1: Service 1 {
      class: server
    }
    s2: Service 2 {
      class: server
    }
  }
  m: Message Channel {
    grid-rows: 1
    grid-gap: 0
    m1: Message 1 {
      class: msg
    }
    m2: Message 2 {
      class: msg
    }
  }
  s.s1 <- m.m1: Consume
  s.s2 <- m.m1: Consume
}
m2 {
  class: none
  grid-rows: 2
  s: {
    class: none
    grid-rows: 1
    s1: Service 1 {
      class: server
    }
    s2: Service 2 {
      class: server
    }
  }
  m: Message Bus {
    grid-rows: 1
    grid-gap: 0
    m1: Message 1 {
      class: msg
    }
    m2: Message 2 {
      class: msg
    }
  }
  s.s1 <- m.m2: Consume
  s.s2 <- m.m2: Consume
}
```

## Event Streaming Platform

An {{< term esp >}} is a distributed implementation of a {{< term msg >}},
designed to offer high availability and fault tolerance.

Before diving deeper, let’s first clarify the concept of an **Event**.

### Event

The term **Message** broadly refers to any piece of information exchanged within a system.
Messages generally fall into two main categories:

1. **Command** – A directive sent to the system, requesting it to perform a specific action.
2. **Event** – A record of something that has already occurred.

Let’s consider a payment transaction as an example:

- **Command**: The client begins the transaction by issuing a command such as `InitiateTransaction`.
- **Event**: As the system processes the transaction, it generates events like `AccountBalanceChanged`, `TransactionCompleted`, or `TransactionFailed`.

```d2
direction: right
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
c -> s: "InitiateTransaction" {
  style.bold: true
}
s -> m1
s -> m2
s -> m3
```

Many architectures prioritize **durable storage of events** and may even bypass persistent storage of commands entirely.
This explains why the term **Event** is often preferred over **Message** or **Command**.
We’ll delve deeper into this in the [Event-driven Architecture]({{< ref "event-driven-architecture" >}}) topic.

### Streaming Platform

Briefly,
{{< term esp >}} is a messaging system that combines two key features:

- **Streaming:** Messages are delivered and consumed immediately after they’re produced, enabling real-time processing.
- **Persistence:** Messages are durably stored in the underlying storage layer, allowing for reliable delivery and replay.

Let’s explore how to build an {{< term esp >}}.

{{< callout type="info" >}}
In the following section, we’ll focus on core concepts popularized by **Apache Kafka**—
the industry’s most widely adopted solution today.
{{< /callout >}}

## Event Streaming Cluster

An {{< term esp >}} operates as a decentralized cluster composed of multiple servers,
commonly referred to as **brokers**.

```d2
direction: right
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
One broker is elected as the **Controller** (or **Leader**) node using the [Raft algorithm]({{< ref "consensus-protocol" >}}).

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

{{< callout type="info" >}}
You can think of a topic as a table, where each row represents a single event of the same type.
{{< /callout >}}

### Data Structure

At a fundamental level, an {{< term esp >}} supports two operations: **Produce** (write) and **Consume** (read).
There are no update or delete operations.

Events are stored as **append-only files** on brokers.
Modifying an event in the middle of the log would require shifting all subsequent entries,
which is inefficient and generally avoided.

```d2
b: Broker {
    grid-rows: 1
    f1: "AccountCreated.Events file" {
      e1: |||yaml
      event1:
        accountId: acc1
        email: mylovelyemail@mail.com
        at: 00:01
      event2:
        accountId: acc2
        email: nottoday@mail.com
        at: 00:02
      ...(space for new events)...
      |||
    }
    f2: "BalanceChanged.Events file" {
      e1: |||yaml
      event1:
        accountId: acc01
        balance: 100
      event2:
        accountId: acc01
        balance: 50
      event3:
        accountId: acc02
        balance: 30
      ...(space for new events)...
      |||
    }
}
```

### Partition

Storing an entire topic on a single broker is inefficient.
The storage and access load for each topic may vary significantly,
potentially causing uneven load distribution across brokers.

To address this, topics are split into smaller units called **partitions**, which are distributed across brokers.
This concept is similar to [sharding]({{< ref "peer-to-peer-architecture#shard" >}}) in a {{< term p2p >}} cluster.

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

For example, with an ISR threshold of `2 seconds`,
a replica that fetched data last at `00:02` while the primary is at `00:05` is **out-of-sync**.

```d2
grid-rows: 1
horizontal-gap: 150
l: "Primary (Time = 00:05, ISR Threshold = 2s)" {
  class: server
}
b: "Replica 1 (LastFetch = 00:04)" {
  class: server
}
c: "Replica 2 (LastFetch = 00:02)" {
  class: generic-error
}
```

Similar to [Quorum-based Consistency]({{< ref "distributed-database#quorum-based-consistency" >}}),
a produce request includes an acknowledgement (**ACK**) setting.
This determines how many brokers (including the primary) must successfully save the event before the producer receives a response.

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
p -> r: 4. Crash before replicating (data loss) {
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
p -> r1: 3. Replicate {
  style.bold: true
}
p -> s: 4. Respond
p -> p: 5. Crash here but no data loss {
  class: error-conn
}
```

Using **ISRs (In-Sync Replicas)** instead of all replicas is important
because out-of-sync replicas might be slow or unavailable due to crashes or partitioning.
Waiting for all replicas can degrade performance or block the partition entirely.
Once a replica becomes in-sync again, it will fetch any missed events from the primary.

Please keep **ACK** in mind—this mechanism is crucial for understanding [Delivery Semantics](#delivery-semantics).

## Consuming

**Event Streaming** typically uses [Long Polling]({{< ref "communication-protocols#long-polling" >}}) for event delivery.
This approach decouples producers and consumers, improving system availability.

### Commit Offset

In streaming systems, **Offset** refers to the sequential position of an event in an append-only log.
Please note that event offsets are managed at the partition level, not globally across the entire topic.

```d2
p1: Partition 1 {
    f1: "AccountCreated.Events file" {
      c: |||yaml
      event1:
        offset: 1
        accountId: acc1
      event3:
        offset: 2
        accountId: acc3
      |||
    }
}
p2: Partition 2 {
    f1: "AccountCreated.Events file" {
      c: |||yaml
      event2:
        offset: 1
        accountId: acc2
      event4:
        offset: 2
        accountId: acc4
      |||
    }
}
```

To prevent duplicate processing, each partition keeps track of the last consumed offset for each consumer.

```d2
p1: Partition 1 {
    c: "Offsets" {
        c: |||yaml
        consumer1:
            lastOffset: 1
        consumer2:
            lastOffset: 2
        |||
    }
    f1: "AccountCreated.Events file" {
      c: |||yaml
      event1:
        offset: 1
        accountId: acc1
      event3:
        offset: 2
        accountId: acc3
      |||
    }
}
p2: Partition 2 {
    c: "Offsets" {
        c: |||yaml
        consumer1:
            lastOffset: 1
        consumer2:
            lastOffset: 1
        |||
    }
    f1: "AccountCreated.Events file" {
      c: |||yaml
      event2:
        offset: 1
        accountId: acc2
      event4:
        offset: 2
        accountId: acc4
      |||
    }
}
```

Consumers periodically fetch new events from partitions, handle them, and then **commit/increase the offset** to avoid reprocessing.

```d2
shape: sequence_diagram
c: Consumer {
    class: client
}
q: Partition {
    class: mq
}
c -> q: 1. Consume
q -> c: 2. Return an event
c -> c: 3. Handle the event
c -> q: 4. Commit offset
```

### Consumer Group

A topic can have multiple partitions, making it inefficient for a single consumer to handle all alone.
Instead, a **consumer group** allows multiple consumers to read different partitions in parallel.

All consumers in a group share the same name and commit offset collectively,
ensuring each event is processed only **once by the group**.

For example:

- The `AccountCreated` topic is divided into two partitions, and each partition keeps track of its own consumer offsets.
- In `Group A`, each consumer is assigned to a different partition.
- In `Group B`, there is only one consumer, so it processes all partitions.
- In `Group C`, there are three consumers, which is more than the number of partitions, so one consumer remains idle.

```d2
grid-rows: 2
t1: AccountCreated {
    p1: Partition 1 {
      c: "Offsets" {
          c: |||yaml
          consumerGroupA:
              lastOffset: 1
          consumerGroupB:
              lastOffset: 2
          consumerGroupC:
              lastOffset: 2
          |||
      }
      f1: "AccountCreated.Events file" {
        c: |||yaml
        event1:
          offset: 1
          accountId: acc1
        event3:
          offset: 2
          accountId: acc3
        |||
      }
    }
    p2: Partition 2 {
      c: "Offsets" {
          c: |||yaml
          consumerGroupA:
              lastOffset: 2
          consumerGroupB:
              lastOffset: 1
          consumerGroupC:
              lastOffset: 2
          |||
      }
      f1: "AccountCreated.Events file" {
        c: |||yaml
        event2:
          offset: 1
          accountId: acc2
        event4:
          offset: 2
          accountId: acc4
        |||
      }
    }
}
c: Consumers {
    cg1: Consumer Group A {
        c1: Consumer 1
        c2: Consumer 2
    }
    cg2: Consumer Group B {
        c1: Consumer 1
    }
    cg3: Consumer Group C {
        c1: Consumer 1
        c2: Consumer 2
        c3: Consumer 3
    }
}
t1.p1 -> c.cg1.c1
t1.p2 -> c.cg1.c2
t1.p1 -> c.cg2.c1
t1.p2 -> c.cg2.c1
t1.p1 -> c.cg3.c1
t1.p2 -> c.cg3.c2
```

Partition replicas serve solely for backup and recovery purposes. Consumers must always read from the primary broker.
Unlike traditional databases, where read operations do not alter the state of the system,
in {{< term esp >}}, consumption is [non-idempotent]({{< ref "api-design#request-idempotency" >}}) because each read updates the consumer’s offset.
