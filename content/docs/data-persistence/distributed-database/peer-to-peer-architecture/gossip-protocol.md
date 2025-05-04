---
title: Gossip Protocol
weight: 10
---

We'll explore the first approach to maintaining a {{< term p2p >}} cluster—{{< term gosProto >}}.

## Eager Reliable Broadcast

**Eager Reliable Broadcast** is a communication protocol in distributed systems that
ensures messages are broadcast to all participants.
In this protocol, each peer continuously exchanges information with every other peer.

```d2
c: Cluster {
  grid-rows: 2
  n1: Server 1 {
    class: server
  }
  n2: Server 2 {
    class: server
  }
  n3: Server 3 {
    class: server
  }
  n4: Server 4 {
    class: server
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

This protocol follows an **eager delivery** model, where messages are immediately sent to all peers upon generation.
While effective, this approach becomes inefficient as the number of peers grows—
it consumes substantial bandwidth and system resources due to redundant message exchanges.

To address these inefficiencies, the {{< term gosProto >}} was introduced,
reducing resource usage by limiting the number of exchanges.

## Gossip Protocol

The **Gossip Protocol** is akin to how rumors spread in an office.
A peer starts by sharing a message with a few randomly selected peers.
These peers then forward the message to others, eventually reaching all nodes in the cluster.

```d2
grid-columns: 1

c0: Cluster {
    n1: Server 1 {
        class: server
    }
    n2: Server 2 {
        class: server
    }
    n3: Server 3 {
        class: server
    }
    n4: Server 4 {
        class: server
    }
}
c1: Cluster {
    n1: Server 1 {
        class: server
    }
    n2: Server 2 {
        class: server
    }
    n3: Server 3 {
        class: server
    }
    n4: Server 4 {
        class: server
    }
    n1 -> n2: 1
    n1 -> n3: 1
}
c2: Cluster {
    n1: Server 1 {
        class: server
    }
    n2: Server 2 {
        class: server
    }
    n3: Server 3 {
        class: server
    }
    n4: Server 4 {
        class: server
    }
    n2 -> n4: 2
    n3 -> n4: 2
}
```

### Maintaining Cluster Membership

Servers maintain information such as addresses, states, and shard allocations of other nodes.

For example:

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

### Adding a Node

[Bootstrapping nodes](https://en.wikipedia.org/wiki/Bootstrapping_node) serve as entry points for new nodes,
typically accessed via static IPs or **DNS**.

In this example, `A` acts as a bootstrap node.
When node `C` joins, it sends its information to `A` and receives the current cluster metadata.

```d2
cl: Cluster {
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
c <-> cl.a
```

`A` then informs `B` of `C`’s arrival, allowing `B` to update its view of the cluster.

```d2
cl: Cluster {
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

Now, all members are aware of `C`, meaning it has successfully joined the cluster.
Based on this information, the system can redistribute data if necessary.

```d2
cl: Cluster {
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

With [consistent hashing](../../../#consistent-hashing) and shared metadata, any node can act as an interface:

- Serving read requests directly or forwarding them to replicas.
- Routing write requests to the appropriate node.

```d2

direction: right
c: Cluster {
  n1: Server 1 {
    class: server
  }
  n2: Server 2 {
    class: server
  }
}
client: Client {
  class: client
}
client -> c.n1: "1. Write a record"
c.n1 -> c.n1: "2. Calculate appropriate node" 
c.n1 -> c.n2: "3. Forward" 
```

## Gossip Rounds

**Gossip Rounds** form the backbone of the protocol.
Periodically, each node selects random peers and shares its current state (e.g., heartbeat data),
progressively spreading updates throughout the cluster.

For example, at time `3`, `A` gossips its state. `B` receives it and forwards it to `C`.

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

Later, `B` also gossips its state to others.

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

One key advantage of **Gossip Rounds** is fault detection.
If a node doesn't receive a heartbeat from another node within a defined time window,
it marks that peer as **DOWN**.

If the heartbeat timeout is 3 seconds, by time `5`, both `A` and `B` consider `C` down due to the absence of updates.

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

When a node is deemed unreachable, the system stops forwarding data to it and instead uses a **replica**.
Once the original node recovers, it synchronizes with its replica to restore consistency.

```d2
grid-columns: 2
vertical-gap: 100
e: Node failure {
   n1: Server A {
    shape: circle
   }
   n2: Server C (Primary) {
    class: generic-error
   }
   n3: Server B (Replica) {
    shape: circle
   }
   n1 -> n2: Stop forwarding {
    class: error-conn
   }
   n1 -> n3: Failover
}
r: Recovery {
   n1: Server A {
    shape: circle
   }
   n2: Server C (Primary) {
    shape: circle
   }
   n3: Server B (Replica) {
    shape: circle
   }
   n1 -> n2
   n2 <- n3: Pull data to recover
}
```

## Data Conflicts

This architecture emphasizes **Availability over Consistency (AP)**.
During network partitions, nodes can continue serving writes independently.
However, once partitions heal, conflict resolution becomes necessary.

Two effective approaches to conflict resolution are:

### Last Write Wins (LWW)

This method uses timestamps to resolve conflicts.
Each record tracks the latest update time, and the most recent version wins during a merge.

```d2
s1: Server 1 {
    shape: circle 
    r: |||yaml
    (Id = 123, Name = John, Updated = "00:00")  
    |||
}

s2: Server 2 {
    shape: circle 
    r: |||yaml
    (Id = 123, Name = Mike, Updated = "00:03")  
    |||
}
s1 -> r
s2 -> r: Server 2 wins
r: |||yaml
    (Id = 123, Name = Mike)  
|||
```

This strategy relies on **clock synchronization**. If clocks are skewed, incorrect records might be selected.

### Vector Clocks

A more robust alternative, **Vector Clocks**, avoids dependency on synchronized clocks.
Instead, each record tracks a version vector in the format `[(Server, Version)]`.
This allows servers to detect conflicting updates.

For example, a record initially has a version from its owner shard.

```d2
c: Cluster {
    p1: Server 1 {
        shape: circle
        r: |||yaml
        (Id = 123, Name = John, VL = [(Server 1, 3)])  
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
Each group may update the record independently, expanding the vector clock;

- `Server 1` updates and increases its version.
- `Server 3` updates and creates its own versions.

```d2
c: Cluster {
    p1: Partition 1 {
        p: Server 1 {
            shape: circle
           r: |||yaml
           (Id = 123, Name = John Doe, VL = [(Server 1, 4)])  
           |||
        }
    }
    p2: Partition 2 {
         p: Server 2 {
         shape: circle
        r: |||yaml
        (Id = 123, Name = John, VL = [(Server 1, 3)])
        |||
        }
    }
    p3: Partition 3 { 
         p: Server 3 {
            shape: circle
            r: |||yaml
            (Id = 123, Name = James, VL = [(Server 1, 3), (Server 3, 1)])  
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

A conflict is identified when vectors cannot be merged deterministically:

- `[(Peer 1, 4)]` is a clear successor of `[(Peer 1, 3)]`.
- But `[(Peer 1, 4)]` and `[(Peer 1, 3), (Peer 3, 1)]` are concurrent and conflict.

The system maintains both versions:

```d2
c: Cluster {
    p1: Server 1 {
        shape: circle
        r1: |||yaml
        Version 1: (Id = 123, Name = John Doe, VL = [(Server 1, 4)])  
        Version 2: (Id = 123, Name = James, VL = [(Server 1, 3), (Server 3, 1)])  
        |||
    }
    p2: Server 2 {
        shape: circle
    }
    p3: Server 3 { 
        shape: circle
    }
    p1 <-> p2: Overwrite
    p1 <-> p3: Conflict
}
```

#### Application-Level Resolution

The application is responsible for resolving conflicts. It receives all conflicting versions and decides how to merge them.

For instance, it may choose to accept the first version:

```d2
s: Application {
    class: process
}
d: Database {
    class: db
}

r1: |||yaml
Version 1: (Id = 123, Name = John Doe, VL = [(Server 1, 4)])  
Version 2: (Id = 123, Name = James, VL = [(Server 1, 3), (Server 3, 1)])  
|||
r2: |||yaml
(Id = 123, Name = John Doe, VL = [(Server 1, 4)])
|||
s <- r1 <- d
s -> r2: Update resolved record
r2 -> d1
d1: Database {
    class: db
}
```

Since the database lacks context about the business logic,
deferring resolution to the application ensures safer and more flexible conflict handling.
Applications should include **additional metadata** in their records to assist in this process.
