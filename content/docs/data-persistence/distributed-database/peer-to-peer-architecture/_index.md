---
title: Peer-to-peer Architecture
weight: 20
---

In terms of high availability and resiliency, the {{< term maSl >}} model is not ideal,
since the master server hold too much power.

The {{< term p2p >}} architecture is a distributed model that promotes
operating a system through the cooperation of multiple servers,
each with **the same responsibility**, known as a **Peer** (server/node)

## Data Ownership

Conceptually, a database is divided into multiple **shards**, each processed by a peer.  
For example, a database might consist of 3 shards, each assigned to a different server.

```d2

classes: {
  part: {
    width: 200
  }
}
grid-columns: 1
db: Database {
  direction: right
  width: 600
  grid-gap: 0
  grid-columns: 3
  p1: Shard 1 {
    class: part
  }
  p2: Shard 2 {
    class: part
  }
  p3: Shard 3 {
    class: part
  }
}
sv: "" {
  direction: right
  grid-columns: 3
  s1: Server 1 {
    class: db
  }
  s2: Server 2 {
    class: db
  }
  s3: Server 3 {
    class: db
  }
}

db.p1 -> sv.s1
db.p2 -> sv.s2
db.p3 -> sv.s3
```

A peer is responsible for managing a part of the database.
Technically, we're not dividing the database into storage chunks and distributing them across servers.
Instead, we take a more granular approach,
where each record is **owned** by one peer, determined by its **unique key**.

A hash function is commonly used to map record keys to specific servers.

For example, consider user records, where the identifier is `user_id` (integer):

- The system currently has three servers, with `server_number` ranging from 0 to 2.
- The mapper function used is `owner_server = user_id % total_number_of_servers (3)`.

```d2

grid-columns: 1

db: Databases {
  grid-columns: 3
  s0: Server 0 {
    class: db
  }
  s1: Server 1 {
    class: db
  }
  s2: Server 2 {
    class: db
  }
}
re: "" {
  grid-columns: 3
  u1: User 0 {
    shape: sql_table
    id: 0 {constraint: primary_key}
  }
  u2: User 1 {
    shape: sql_table
    id: 1 {constraint: primary_key}
  }
  u3: User 4 {
    shape: sql_table
    id: 4 {constraint: primary_key}
  }
}
re.u1 -> db.s1: "0 % 3 = 0 (S0)"
re.u2 -> db.s2: "1 % 3 = 1 (S1)"
re.u3 -> db.s1: "4 % 3 = 1 (S1)"
```

However, this solution becomes problematic when the number of servers changes.
When new servers are added, the hash function must be updated,
and previously stored data may become inaccessible as it was hashed with the old function.

In the previous example, if we increase the number of servers to 4, the hash function changes to `user_id % 4`:

- To find `User 4`, we calculate its server as `4 (id) % 4 (num_of_servers) = 0 (Server 0)`,
but it has been placed on `Server 1` before.
- To resolve this issue, we would need to inefficiently **rehash** and adapt the entire database (all shards).

### Consistent Hashing

As we can see, the standard hashing technique above is insufficient.
Regardless of how the hash function looks,
there is always a direct dependency between the number of servers and the final output.
To overcome this limitation, we use a more flexible approach called {{< term ch >}},
which decouples the record keys from the number of servers by mapping them onto a fixed range.

It's hard to explain this theory.
Let’s walk through its procedure with an example:

{{% steps %}}

#### Virtual Ring

First, we **define a hash function** that returns a value between `0` and `N`:  
`hash(value) -> output in [0 to N]`.

This hash function creates a virtual ring, where values wrap around from 0 to N.

For example, we define a mapper `value % 100`, mapping to the range in `0 -> 99`.
Then, we bend the range to form a virtual ring

![Consistent Hashing Ring](consistent-hashing-ring.png)

#### Placing servers

**Place the servers on the ring** by hashing their IDs.  
Since the hash function produces results within the range of 0 to N, each server is guaranteed to have a specific position on the ring.

![Placing Server on Ring](consistent-hashing-placing-server.png)

#### Placing records

**Hash a record's key** using the same function and place it onto the ring.

![Placing Record on Ring](consistent-hashing-placing-record.png)

To find the **owner server**, we scan clockwise (or counterclockwise, depending on your preference) along the ring from the record’s position.  
The first server encountered during the scan is assigned as the record's owner.

![Finding the Closest Server](consistent-hashing-closest-server.png)

{{% /steps %}}

Does this method completely resolve the mismatch problem?  
No, it only **mitigates** the issue.

For example, if we add a new server, `S0`, we need to migrate data from adjacent servers `S1` and `S83`,
a part of `S1` will be migrated to `S0`

![Migration Example](consistent-hashing-migrate.png)

The advantage of {{< term ch >}} is that it does not require adapting the entire database;  
it only necessitates migrating a portion of the data.

### Virtual Nodes

One issue with {{< term ch >}} is imbalance.  
Record keys are unpredictable, and they can inadvertently cluster on certain servers, leaving others underutilized.  
This is known as the **Hotspot Problem**, where some nodes are heavily accessed while others remain idle.  
For example, `S2` might take up a significant portion of the database, preventing optimal resource utilization.

![Imbalance in Hashing](consistent-hashing-imbalance.png)

The first step in addressing this imbalance is to evaluate the quality of the hash function.  
In some cases, introducing a new mapper function may be necessary.  
However, this step may not always be required,  
as the root cause of the imbalance often lies in the data distribution.

Another cause of imbalance is when the ring is too large for a small number of servers,
leading to large gaps between them.
Intuitively, we can reduce these gaps by placing a server at **multiple points** on the ring, known as **virtual nodes**.

![Virtual Servers](consistent-hashing-virtual-servers.png)

This can be easily implemented by generating a list of virtual IDs for each physical server.  
However, more virtual nodes can cause a physical server standing in many places,
that means more servers will take part in the process of migrating data.

### Shard Replication

Letting shards exist on a single server is dangerous.  
If the server crashes and can’t recover, we lose the shard forever.  
Hence, a shard should have some clones residing in different places.

For each shard,  
we determine the primary shard by `Consistent Hashing` and pick some as replicas.  
The number of replicas for a shard is called the `replication factor`.  
A replica shard can also be used to query data autonomously to stimulate the read performance.

Now, a node grows bigger with a primary shard and some replicas.  
For example, a database with three shards and the replication factor is 2:

- The `Server 1` contains `Shard 1 master`, `Shard 2 replica`
- The `Server 2` contains `Shard 2 master`, `Shard 3 replica`
- The `Server 3` contains `Shard 3 master`, `Shard 1 replica`

```d2
classes: {
  part: {
      width: 315
  }
}
grid-columns: 1
db: Virtually original database {
  direction: right
  grid-gap: 0
  grid-columns: 3
  p1: Shard 1 {
    class: part
  }
  p2: Shard 2 {
    class: part
  }
  p3: Shard 3 {
    class: part
  }
}

peer: Peer-to-peer cluster {
  s1: "Server 1" {
    grid-gap: 50
    grid-columns: 1
    p1: Shard 1 master
    p2: Shard 2 replica
  }
  s2: "Server 2" {
    grid-gap: 50
    grid-columns: 1
    p1: Shard 2 master 
    p2: Shard 3 replica
  }
  s3: "Server 3" {
    grid-gap: 50
    grid-columns: 1
    p1: Shard 3 master
    p2: Shard 1 replica
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

#### How do we actually pick replicas?

A straightforward solution is picking the next servers on the ring.

![Picking Replicas](consistent-hashing-pick-replicas.png)

Some systems make it more robust by taking the infrastructure into account,
e.g., selecting replicas living in another data center or region,
preventing cases in which the entire data center encounters a disaster.

## Master-slave or Peer-to-peer?

`Master-Slave` brings about simplicity.  
It is particularly well-suited for read-heavy workloads.  
However, the master server becomes the system's single point of failure,  
and reliance on it significantly degrades the system’s availability.

`Peer-to-peer` provides a more flexible and highly available cluster.  
The cluster operates cleanly without any runtime dependencies.  
However, maintaining replicated data consistency across peers becomes increasingly difficult as the network scales.  
For highly coupled data stores like `SQL`, this approach can be challenging.  
Data is scattered across multiple servers, and actions like transactions or joins
across many servers over the network become extremely costly and, at times, impossible.

Furthermore, a decentralized cluster is often coupled with `Eventual Consistency`.  
Data on different servers may appear inconsistent at times,
or even look totally different (e.g., different versions during network partitions).  
However, over time, the system will reach consensus.

In fact, many SQL databases treat the `Master-Slave` model as their native setup.  
On the other hand, NoSQL databases, which avoid joins and transactions, use `Peer-to-peer` for high availability.
