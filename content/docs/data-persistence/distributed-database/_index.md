---
title: Distributed Database
weight: 20
prev: sql-database
next: master-slave-architecture
---

As discussed in the [{{< term hs >}}]({{< ref "microservice#horizontal-scaling">}}) section,
hosting everything on a single server becomes impractical as systems grow.
This is particularly true for databases, which prioritize data durability.
Overloading a single server risks turning it into a {{< term spof >}};
any corruption could result in the loss of the entire system’s data.

To address this, databases should operate across multiple servers,
enhancing durability and distributing workloads efficiently.

However, managing and scaling a database cluster is significantly more complex than handling a typical service.
It involves challenges related to reliable data persistence, consistency, and availability.
In this section, we'll explore how to design a {{< term ddb >}}

## Data Replication

In a database cluster, {{< term dr >}} involves copying data from one server to others.
These destination servers are known as **replicas**.

For example, when a change occurs,
the `Primary` server propagates it to `Replica 1` and `Replica 2`.
As a result, these replicas maintain the same data as the `Primary`.

```d2

direction: right
dc: Database cluster {
    w: Primary {
      class: db
    }
    r1: Replica 1 {
      class: db
    }
    r2: Replica 2 {
      class: db
    }
    w -> r1: Replicate {
      style.animated: true
    }
    w -> r2: Replicate {
      style.animated: true
    }
}
```

Why do we need to replicate data?

- **Prevent Data Loss**: If data exists only on a single server, a failure could result in total data loss.
Distributing data ensures it can be recovered in the event of failure.
- **Improve Performance**: Multiple servers can handle read requests concurrently, reducing the load on any single server.
- **Increase Availability**: If the primary server fails, a replica can be promoted to replace it quickly, minimizing downtime.

## Data Consistency

Once data is distributed across multiple servers,
maintaining consistency becomes the primary challenge of {{< term dr >}}.

This consistency largely depends on how data replication is handled.
Specifically, how a database server synchronizes with its replicas after a client (web service or user) performs an update.

```d2
direction: right
c: Client {
    class: client
}
dc: Database cluster {
    w: Primary {
      class: db
    }
    r1: Replica 1 {
      class: db
    }
    r2: Replica 2 {
      class: db
    }
}
c -> dc.w: Update
dc.w -> dc.r1: Sync
dc.w -> dc.r2: Sync
```

There are two main approaches: {{< term syncRep >}} and {{< term asyncRep >}}.

### Synchronous Replication

In {{< term syncRep >}}, the primary server waits until the update has been successfully
applied to both itself and its replicas before responding to the client.

```d2
shape: sequence_diagram
direction: right
c: Client {
    class: client
}
w: Primary {
  class: db
}
r: Replica {
  class: db
}
c -> w: 1. Update data
w -> r: 2. Replicate synchronously
w -> c: 3. Respond to client
```

This method is straightforward and safe, as it guarantees immediate consistency and maintains a reliable backup
if the primary server fails.

However, it doesn't support {{< term ha >}} effectively.
If a replica goes down, the primary must stop processing updates to avoid unsafe writes.

```d2
shape: sequence_diagram
direction: right
c: Client {
    class: client
}
w: Primary {
  class: db
}
r: Replica {
  class: generic-error
}
c -> w: 1. Update data
w -> r: 2. Fail to replicate {
    class: error-conn
}
w -> c: 3. Fail  {
    class: error-conn
}
```

Additionally, this approach increases response latency since clients must wait for the replication to complete.

### Asynchronous Replication

With {{< term asyncRep >}}, the primary server responds to the client immediately,
while replication occurs independently in the background.

```d2
shape: sequence_diagram
direction: right
c: Client {
    class: client
}
w: Primary {
  class: db
}
r: Replica {
  class: db
}
c -> w: 1. Update data
w -> c: 2. Respond to client
w -> r: Replicate data {
    style.animated: true
}
```

{{< callout type="info">}}
In practice, replication is handled in a **separate thread** or process and can sometimes complete
before the client receives the response.
{{< /callout >}}

This approach offers better availability since the primary can continue serving requests even if a replica fails.
It also reduces write latency. However, it introduces important challenges:

- **Temporary Inconsistency**: Updates may not immediately appear on all replicas, leading to temporary inconsistencies.
- **Potential Data Loss**: If the primary fails before finishing replication and its data is lost, incomplete clones are permanently lost.

```d2
shape: sequence_diagram
client: Client {
    class: client
}
w: Primary {
    class: db
}
r: Replica {
    class: db
}
client -> w: 1. Update data
w -> client: 2. Respond to client immediately
w -> w: 3. Crash before replicating {
    class: error-conn
}
```

## Quorum-based Consistency

**Quorum-based Consistency** balances the trade-offs between {{< term asyncRep >}} and {{< term syncRep >}}.

It uses a metric called {{< term quo >}}, defining how many replicas must confirm a read or write operation
before it’s considered successful.

### Write Quorum

A **Write Quorum (WQ)** specifies how many replicas must confirm a write synchronously.
The remaining replicas can be updated asynchronously.

For example, with three replicas and a `WQ of 1`:

- 1 replica is updated synchronously.
- 2 replicas are updated asynchronously (`Replicas - WQ = 2`).

```d2
shape: sequence_diagram
client: Client {
    class: client
}
w: Primary Server {
    class: db
}
r1: Replica 1 {
    class: db
}
r2: Replica 2 {
    class: db
}
r3: Replica 3 {
    class: db
}
client -> w: 1. Update data
w -> r1: 2. Replicate synchronously
w -> client: 3. Respond to client
w -> r2: Replicate asynchronously {
    style.animated: true
    style.bold: true
}
w -> r3: Replicate asynchronously {
    style.animated: true
    style.bold: true
}
```

### Read Quorum

A **Read Quorum (RQ)** defines how many servers must agree on a read operation before a response is returned.

For instance, with an `RQ of 2`, reading from `Replica 1` involves verifying with
two other servers to ensure **the latest value among them** is retrieved.

```d2
shape: sequence_diagram
client: Client {
    class: client
}
r1: Replica 1 {
    class: db
}
w: Primary {
    class: db
}
r2: Replica 2 {
    class: db
}
client -> r1: 1. Read data
r1 -> w: 2. Verify
r1 -> r2: 2. Verify
r1 -> client: 3. Respond the latest value
```

## Consistency Level

The combination of write and read quorums defines a database’s consistency level,
determining how strongly data consistency is enforced.

### Strong Consistency Level

**Strong Consistency** ensures all servers reflect the same data at any given moment.

To achieve this, the sum of `WQ + RQ` must be greater than or equal to the total number of replicas.
This guarantees **overlap** between read and write operations.

For example, in a cluster with 2 replicas, we define:

- `WQ` is 1, e.g., `Replica 1` is up-to-date.
- `RQ` is 1.

Now, imagine a read request initially reaches an inconsistent replica, such as `Replica 2`.
To ensure data accuracy, the read operation leverages the read quorum by querying
at least one consistent server, either `Replica 1` or the `Primary`, before returning a response to the client.

```d2
shape: sequence_diagram
client: Client {

    class: client
}
r2: Replica 2 {
    class: db
}
w: Primary {
    class: db
}
r1: Replica 1 (Up-to-date) {
    class: db
}
client -> r2: 1. Read data
r2 -> r1: 2. Use the quorum here
r2 -> w: 2. Or here
```

### Eventual Consistency Level

**Eventual Consistency** means that while data discrepancies might exist temporarily, all replicas will eventually converge.

Here, `WQ + RQ` is less than the number of replicas.

For example, we configure a cluster of 2 replicas as:

- `WQ` is 1, e.g., `Replica 1` is up-to-date.
- `RQ` is set to 0, meaning read operations can return results without verifying data with other servers.

Now, imagine a client performs an update on the primary server.
The server synchronously replicates the update to `Replica 1` before sending a response to the client,
while it replicates to `Replica 2` asynchronously.

```d2
shape: sequence_diagram
c: Client {
    class: client
}
p: Primary {
    class: db
}
r1: Replica 1 {
    class: db
}
r2: Replica 2 {
    class: db
}
c -> p: 1. Update data
p -> r1: 2. Replicate synchronously
p -> c: 3. Respond
p -> r2: 4. Replicate asynchronously {
    style.animated: true
    style.bold: true
}
```

If the client then reads data from `Replica 2` before the asynchronous replication is completed, it will retrieve outdated data.

```d2
shape: sequence_diagram
c: Client {
    class: client
}
p: Primary {
    class: db
}
r1: Replica 1 {
    class: db
}
r2: Replica 2 {
    class: db
}
c -> p: 1. Update data
p -> r1: 2. Replicate synchronously
p -> c: 3. Respond
p -> r2: 4. Replicate asynchronously {
    style.animated: true
    style.bold: true
}
c -> r2: 5. Read the old version as the previous step hasn't completed {
    class: error-conn
}
p -> r2: 4. Complete replicating {
    style.bold: true
}
```

Although **Strong Consistency** is safer and easier to reason about,
it comes at the cost of **reduced availability** and increased latency.
As the number of servers in the system grows,
coordinating updates and ensuring consistent reads becomes increasingly complex and time-consuming.

On the other hand,
{{< term eveCons >}} sacrifices immediate synchronization but guarantees that
all replicas will eventually converge to a consistent state.
If temporary inconsistencies are acceptable and can be resolved over time,
{{< term eveCons >}} provides faster responses and higher system availability.
In such cases, the **Read Quorum** is often set to zero to maximize performance and responsiveness.

### Standby Server

Regardless of whether you choose **Strong Consistency** or **Eventual Consistency**,
it’s strongly recommended to keep the **Write Quorum** at least 1.
As asynchronous replicas are inherently unreliable for immediate recovery,
having at least one synchronously updated, consistent server is crucial for safeguarding data.
These reliable, up-to-date servers are commonly known as **Standby Servers**.

```d2
direction: right
p: Primary Server {
  class: server
}
s: Standby Server {
  class: server
}
p -> s: Replicate synchronously {
  style.animated: true
}
```
