---
title: Gossip Protocol
weight: 10
---

We'll see the first approach to maintain a {{< term p2p >}} cluster - {{< term gosProto >}}.

## Eager Reliable Broadcast

The **Eager Reliable Broadcast** is a communication protocol designed
for distributed systems to ensure that messages are broadcast to all participants.
In essence, this protocol requires each peer to continuously exchange information with all other peers.

```d2
c: Cluster {
  grid-rows: 2
  n1: Peer 1 {
    shape: server
  }
  n2: Peer 2 {
    shape: server
  }
  n3: Peer 3 {
    shape: server
  }
  n4: Peer 4 {
    shape: server
  }
  n1 -> n2
  n1 -> n3
  n1 -> n4
  n2 -> n1
  n2 -> n3
  n2 -> n4
  n3 -> n1
  n3 -> n2
  n3 -> n4
  n4 -> n1
  n4 -> n2
  n4 -> n3
}
```

This protocol ensures **Eager delivery**,
where a message is immediately sent to all recipients as soon as it is generated.
However, as more peers are added, the complexity of the cluster increases.  
This approach leads to resource and bandwidth inefficiencies, as members must exchange a significant amount of redundant information.

The {{< term gosProto >}} was introduced to reduce the resource consumption of
the **Eager Reliable Broadcast** protocol by minimizing the number of exchanges.

## Gossip Protocol

The **Gossip Protocol** [can be illustrated by the analogy of office workers spreading rumors](https://en.wikipedia.org/wiki/Gossip_protocol).
To make an announcement, a peer starts by communicating with a few random peers.
These peers then propagate the message to other peers. **Eventually**, all peers will receive the announcement.

```d2
grid-columns: 1

c0: Cluster {
    n1: Peer 1 {
        shape: server
    }
    n2: Peer 2 {
        shape: server
    }
    n3: Peer 3 {
        shape: server
    }
    n4: Peer 4 {
        shape: server
    }
}
c1: Cluster {
    n1: Peer 1 {
        shape: server
    }
    n2: Peer 2 {
        shape: server
    }
    n3: Peer 3 {
        shape: server
    }
    n4: Peer 4 {
        shape: server
    }
    n1 -> n2: 1
    n1 -> n3: 1
}
c2: Cluster {
    n1: Peer 1 {
        shape: server
    }
    n2: Peer 2 {
        shape: server
    }
    n3: Peer 3 {
        shape: server
    }
    n4: Peer 4 {
        shape: server
    }
    n2 -> n4: 2
    n3 -> n4: 2
}
```

For example, let’s look at how a cluster of members is maintained.  
Initially, a peer stores information (address, state, sharding, etc.) about other peers.

```d2
c: Cluster {
  a: A {
    shape: circle
    m: |||yaml
    B: UP
    |||
  }
  b: B {
    shape: circle
    m: |||yaml
    A: UP
    |||
  }
}
```

### Adding Node

[Bootstrapping nodes](https://en.wikipedia.org/wiki/Bootstrapping_node) serve as designated entry points for new nodes,  
which can be obtained using static addresses or **DNS**.

In our example, let’s say `A` is a bootstrapping node.  
When a new node `C` wants to join the cluster,  
it sends its information to and retrieves the cluster metadata from `A`.

```d2
c: Cluster {
  a: A {
    shape: circle
    m: |||yaml
    B: UP
    C: UP (added)
    |||
  }
  b: B {
    shape: circle
    m: |||yaml
    A: UP
    |||
  }
}
c: C {
    shape: circle
    m: |||yaml
    A: UP
    B: UP
    |||
}
c <-> c.a
```

Then, `A` informs `B` about the `C`’s participation.  
`B` becomes aware of `C` and adds it to its state list.

```d2
c: Cluster {
  a: A {
    shape: circle
    m: |||yaml
    B: UP
    C: UP
    |||
  }
  b: B {
    shape: circle
    m: |||yaml
    A: UP
    C: UP (added)
    |||
  }
  n1 -> n2
}
c: C {
    shape: circle
    m: |||yaml
    A: UP
    B: UP
    |||
}
```

Now, all the members recognize of `C`, that means it has joined the cluster.
Based on the cluster information, nodes can calculate and re-balance the data if necessary.

```d2
c: Cluster {
  n1: A {
    shape: circle
    m: |||yaml
    B: UP
    C: UP
    |||
  }
  n2: B {
    shape: circle
    m: |||yaml
    A: UP
    C: UP
    |||
  }
  n3: C {
    shape: circle
    m: |||yaml
    A: UP
    B: UP
    |||
  }
  n1 <-> n2 <-> n3 <-> n1
}

```

### Forwarding

With [consistent hashing](../../../#consistent-hashing) and the cluster metadata,  
any node can be an interface

- Serving read requests autonomously (or forwarding to replicas).
- Forwarding write requests to the appropriate nodes.

```d2

direction: right
c: Cluster {
  n1: Peer 1 {
    shape: server
  }
  n2: Peer 2 {
    shape: server
  }
}
client: Client {
  class: client
}
client -> c.n1: "1. Write a record"
c.n1 -> c.n1: "2. Calculate appropriate node" 
c.n1 -> c.n2: "3. Forward" 
```

## Gossip Round

**Gossip Round** is the foundation of the protocol.
**Periodically**, a node selects random peers to send its state (like a [heartbeat](../../../web-service/service-cluster/#heartbeat-mechanism))
spreading this information progressively across the entire cluster.
As the time goes by, **Gossip Rounds** will help maintain the cluster metadata between members.

For example, at the moment of `3`,
`A` starts gossiping to spread its state.
After receiving the information, `B` transmits it to `C`.

```d2
grid-columns: 2
vertical-gap: 100
c: "Cluster" {
  a: A {
    shape: circle
    m: |||yaml
    B: UP, Heartbeat = 1
    C: UP, Heartbeat = 1
    |||
  }
  b: B {
    shape: circle
    m: |||yaml
    A: UP, Heartbeat = 1
    C: UP, Heartbeat = 1
    |||
  }
  c: C {
    shape: circle
    m: |||yaml
    A: UP, Heartbeat = 1
    B: UP, Heartbeat = 1
    |||
  }
}
c3: "Cluster (A gossips)" {
  a: A {
    shape: circle
    m: |||yaml
    B: UP, Heartbeat = 1
    C: UP, Heartbeat = 1
    |||
  }
  b: B {
    shape: circle
    m: |||yaml
    A: UP, Heartbeat = 3
    C: UP, Heartbeat = 1
    |||
  }
  c: C {
    shape: circle
    m: |||yaml
    A: UP, Heartbeat = 3
    B: UP, Heartbeat = 1
    |||
  }
  a -> b: "A: UP"
  b -> c: "A: UP"
}
```

`B` also starts to gossip its state to other nodes at the moment of `3`.

```d2
c3: "Cluster" {
  a: A {
    shape: circle
    m: |||yaml
    B: UP, Heartbeat = 1
    C: UP, Heartbeat = 1
    |||
  }
  b: B {
    shape: circle
    m: |||yaml
    A: UP, Heartbeat = 3
    C: UP, Heartbeat = 1
    |||
  }
  c: C {
    shape: circle
    m: |||yaml
    A: UP, Heartbeat = 3
    B: UP, Heartbeat = 1
    |||
  }
}
c4: "Cluster (B gossips)" {
  a: A {
    shape: circle
    m: |||yaml
    B: UP, Heartbeat = 3
    C: UP, Heartbeat = 1
    |||
  }
  b: B {
    shape: circle
    m: |||yaml
    A: UP, Heartbeat = 3
    C: UP, Heartbeat = 1
    |||
  }
  c: C {
    shape: circle
    m: |||yaml
    A: UP, Heartbeat = 3
    B: UP, Heartbeat = 3
    |||
  }
  b -> a: "B: UP"
  b -> c: "B: UP" 
}
```

### Failure Detection

A real benefit of **Gossip Round** is determining the failure of servers.
A server marks another server as **DOWN** after it hasn't received any heartbeat within a certain period.

For example, if the heartbeat lifetime is constrained to 3 seconds,
then at the moment `5`, both `A` and `B` will assume that `C` is down, as its heartbeat has expired.

```d2
grid-columns: 2
vertical-gap: 100
c: "Cluster (HeartbeatLifetime = 3, Time = 5)" {
  n1: A {
    shape: circle
    m: |||yaml
    B: UP, Heartbeat = 3
    C: DOWN, Heartbeat = 1 (expired)
    |||
  }
  n2: B {
    shape: circle
    m: |||yaml
    A: UP, Heartbeat = 3
    C: DOWN, Heartbeat = 1 (expired)
    |||
  }
  n3: C {
    class: generic-error
  }
}
```

#### Temporary Promotion

When a node considers another node as corrupted,
it will stop forwarding data to the failed node,
but temporarily work with a **replica** instead.
Once the failed node revives, it can collaborate with its replicas to recover the data.

```d2
grid-columns: 2
vertical-gap: 100
e: Node failure {
   n1: Peer A {
    shape: circle
   }
   n2: Peer C (Primary) {
    class: generic-error
   }
   n3: Peer B (Replica) {
    shape: circle
   }
   n1 -> n2: Stop forwarding {
    class: error-conn
   }
   n1 -> n3: Failover
}
r: Recovery {
   n1: Peer A {
    shape: circle
   }
   n2: Peer C (Primary) {
    shape: circle
   }
   n3: Peer B (Replica) {
    shape: circle
   }
   n1 -> n2
   n2 <- n3: Pull data to recover
}
```

## Data Conflicts

The previous workflow is an **AP (Availability over Consistency)** setup.
To ensure high availability, the **Gossip Protocol** allows partitions to work autonomously.
In other words, different partitions can concurrently serve writing data.

Probably, that leads to a big question.
After the network heals, how do we merge and **resolve conflicts** from different partitions?
We have two effective approaches:

### Last Write Wins (LWW)

The first approach resolves conflicts based on timestamps.  
In short, a record keeps track of the latest timestamp when it was updated.  
In case of conflict, the newest record is selected.

```d2
s1: Peer 1 {
    shape: circle 
    r: |||yaml
    (Id = 123, Name = John, Updated = "00:00")  
    |||
}

s2: Peer 2 {
    shape: circle 
    r: |||yaml
    (Id = 123, Name = Mike, Updated = "00:03")  
    |||
}
s1 -> r
s2 -> r: Peer 2 wins
r: |||yaml
    (Id = 123, Name = Mike)  
|||
```

To apply this strategy, it is crucial to ensure that **clock synchronization** between servers is reliable.  
A node with an incorrect clock (e.g., ahead or behind real time) may assign incorrect timestamps,
leading to the wrong data being selected.

### Vector Clocks

**Vector Clocks** offer a more reliable approach, independent of clock synchronization.  
In short, **Vector Clocks** maintain conflicting versions of a record and allow **clients to resolve** conflicts autonomously.

Instead of a timestamp, records maintain a vector clock in the form of `[(Server, Version)]`.
When a server writes a record,
it will increases its respective version (or creates if not exists).

Let's say a record initially has the vector clock from its owning shard `Server 1`.

```d2
c: Cluster {
    p1: Server 1 {
        shape: circle
        r: |||yaml
        (Id = 123, Name = John, VL = [(Peer 1, 3)])  
        |||
    }
    p2: Server 2 {
        shape: circle
    }
    p3: Server 3 { 
       shape: circle 
    }
}
```

Then, a network partition occurs, the cluster is divided into three partitions.  
All partitions can update the record, which then expands the vector clock;

- `Server 1` updates and increases its version.
- `Server 3` updates and creates its own versions.

```d2
c: Cluster {
    p1: Partition 1 {
        p: Server 1 {
            shape: circle
           r: |||yaml
           (Id = 123, Name = John Doe, VL = [(Peer 1, 4)])  
           |||
        }
    }
    p2: Partition 2 {
         p: Server 2 {
         shape: circle
        r: |||yaml
        (Id = 123, Name = John, VL = [(Peer 1, 3)])
        |||
        }
    }
    p3: Partition 3 { 
         p: Server 3 {
            shape: circle
            r: |||yaml
            (Id = 123, Name = James, VL = [(Peer 1, 3), (Peer 3, 1)])  
            |||
        }
    }
    p1 <-> p2: {
        class: error-conn
    }
    p3 <-> p2 {
        class: error-conn
    }
}
```

A conflict is detected if two vectors can’t be merged. For example:

- `[(Peer 1, 4)]` is a descendant of `[(Peer 1, 3)]` and can overwrite it.
- `[(Peer 1, 4)]` and `[(Peer 1, 3), (Peer 3, 1)]` generate a conflict.

If any conflict is detected, multiple versions of the record are maintained during the merge process.

```d2
c: Cluster {
    p1: Peer 1 {
        shape: circle
        r1: |||yaml
        Version 1: (Id = 123, Name = John Doe, VL = [(Peer 1, 4)])  
        Version 2: (Id = 123, Name = James, VL = [(Peer 1, 3), (Peer 3, 1)])  
        |||
    }
    p2: Peer 2 {
        shape: circle
    }
    p3: Peer 3 { 
        shape: circle
    }
    p1 <-> p2: Overwrite
    p1 <-> p3: Conflict
}
```

#### Application-level Resolution

The resolution step is entrusted to **the application level**.  
This means the application may receive multiple versions of the data and must merge them accordingly.

For example, an application accepts changes from the first version of the conflict.

```d2

s: Application {
    class: process
}
d: Database {
    class: db
}

r1: |||yaml
Version 1: (Id = 123, Name = John Doe, VL = [(Peer 1, 4)])  
Version 2: (Id = 123, Name = James, VL = [(Peer 1, 3), (Peer 3, 1)])  
|||
r2: |||yaml
(Id = 123, Name = John Doe, VL = [(Peer 1, 4)])
|||
s <- r1 <- d
s -> r2: Update resolved record
r2 -> d1
d1: Database {
    class: db
}
```

Since the database is unaware of the business logic, leaving resolution to the application level is a safer solution.  
Therefore, the application should initially attach useful information to data to assist with the resolution process.
