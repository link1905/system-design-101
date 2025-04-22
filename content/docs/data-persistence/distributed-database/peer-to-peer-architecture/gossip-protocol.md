---
title: Gossip Protocol
weight: 10
---

We'll explore the first approach to maintaining a {{< term p2p >}} cluster—{{< term gosProto >}}.

## Eager Reliable Broadcast Protocol

**Eager Reliable Broadcast** is a communication protocol in distributed systems that
ensures messages are broadcast to all participants.
In this protocol, each peer continuously exchanges information with every other peer.

```d2
c: Cluster {
  grid-rows: 2
  grid-gap: 100
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
  n1 -> n2 {
    style.animated: true
  }
  n1 -> n3 {
    style.animated: true
  }
  n1 -> n4 {
    style.animated: true
  }
  n2 -> n1 {
    style.animated: true
  }
  n2 -> n3 {
    style.animated: true
  }
  n2 -> n4 {
    style.animated: true
  }
  n3 -> n1 {
    style.animated: true
  }
  n3 -> n2 {
    style.animated: true
  }
  n3 -> n4 {
    style.animated: true
  }
  n4 -> n1 {
    style.animated: true
  }
  n4 -> n2 {
    style.animated: true
  }
  n4 -> n3 {
    style.animated: true
  }
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


For example, `Server 1` first informs `Server 2` and `Server 3` of a piece of information. They then propagate it to `Server 4`.
Eventually, all servers in the cluster acknowledge the information.

```d2
grid-columns: 1

c0: Cluster {
  grid-rows: 1
  p1: "" {
    class: none
    n1: Server 1 {
        class: server
    }
  }
  p2: "" {
    class: none
    grid-columns: 1
    n2: Server 2 {
        class: server
    }
    n3: Server 3 {
        class: server
    }
  }
  p3: "" {
    class: none
    n4: Server 4 {
        class: server
    }
  }
}
c1: Cluster {
  grid-rows: 1
  p1: "" {
    class: none
    n1: Server 1 {
        class: server
    }
  }
  p2: "" {
    class: none
    grid-columns: 1
    n2: Server 2 {
        class: server
    }
    n3: Server 3 {
        class: server
    }
  }
  p3: "" {
    class: none
    n4: Server 4 {
        class: server
    }
  }
  p1.n1 -> p2.n2
  p1.n1 -> p2.n3
}
c2: Cluster {
  grid-rows: 1
  p1: "" {
    class: none
    n1: Server 1 {
        class: server
    }
  }
  p2: "" {
    class: none
    grid-columns: 1
    n2: Server 2 {
        class: server
    }
    n3: Server 3 {
        class: server
    }
  }
  p3: "" {
    class: none
    n4: Server 4 {
        class: server
    }
  }
  p2.n2 -> p3.n4
  p2.n3 -> p3.n4
}
```

### Maintaining Cluster Membership

Each server maintains information such as addresses, states, and shard allocations of other nodes.

For example:

```d2
c: Cluster {
  a: "A (1.1.1.1)" {
    m: |||yaml
    B:
      State: UP
      Address: 2.2.2.2
    |||
  }
  b: "B (2.2.2.2)" {
    m: |||yaml
    A:
      State: UP
      Address: 1.1.1.1
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
grid-rows: 1
horizontal-gap: 200
c: "" {
  class: none
  c: "C (3.3.3.3)"
}
cl: Cluster {
  grid-rows: 1
  a: "A (1.1.1.1)" {
    m: |||yaml
    B:
      State: UP
      Address: 2.2.2.2
    |||
  }
  b: "B (2.2.2.2)" {
    m: |||yaml
    A:
      State: UP
      Address: 1.1.1.1
    |||
  }
}
c.c -> cl.a: Ask to join
```

```d2
grid-rows: 1
horizontal-gap: 200
c: "" {
  class: none
  c: "C (3.3.3.3)" {
    m: |||yaml
    A:
      State: UP
      Address: 1.1.1.1
    B:
      State: UP
      Address: 2.2.2.2
    |||
  }
}
cl: Cluster {
  grid-rows: 1
  a: "A (1.1.1.1)" {
    m: |||yaml
    B:
      State: UP
      Address: 2.2.2.2
    C:
      State: UP (Newly added)
      Address: 3.3.3.3
    |||
  }
  b: "B (2.2.2.2)" {
    m: |||yaml
    A:
      State: UP
      Address: 1.1.1.1
    |||
  }
}
c.c <- cl.a: Cluster information
```

`A` then informs `B` of `C`’s arrival, allowing `B` to update its view of the cluster.

```d2
grid-rows: 1
c: "" {
  class: none
  c: "C (3.3.3.3)" {
    m: |||yaml
    A:
      State: UP
      Address: 1.1.1.1
    B:
      State: UP
      Address: 2.2.2.2
    |||
  }
}
cl: Cluster {
  grid-rows: 1
  horizontal-gap: 200
  a: "A (1.1.1.1)" {
    m: |||yaml
    B:
      State: UP
      Address: 2.2.2.2
    C:
      State: UP
      Address: 3.3.3.3
    |||
  }
  b: "B (2.2.2.2)" {
    m: |||yaml
    A:
      State: UP
      Address: 1.1.1.1
    C:
      State: UP (Newly added)
      Address: 3.3.3.3
    |||
  }
  a -> b: Inform of C {
    class: bold-text
  }
}
```

Now, all members are aware of `C`, meaning it has successfully joined the cluster.
Based on this information, the system can redistribute data if necessary.

```d2
cl: Cluster {
  a: "A (1.1.1.1)" {
    m: |||yaml
    B:
      State: UP
      Address: 2.2.2.2
    C:
      State: UP
      Address: 3.3.3.3
    |||
  }
  b: "B (2.2.2.2)" {
    m: |||yaml
    A:
      State: UP
      Address: 1.1.1.1
    C:
      State: UP
      Address: 3.3.3.3
    |||
  }
  c: "C (3.3.3.3)" {
    m: |||yaml
    A:
      State: UP
      Address: 1.1.1.1
    B:
      State: UP
      Address: 2.2.2.2
    |||
  }
}
```

### Forwarding

With [consistent hashing](../../../#consistent-hashing) and shared metadata, any node can act as an interface:

- Serving read requests directly or forwarding them to respective replicas.
- Routing write requests to the appropriate node.

```d2
grid-rows: 2
client: Client {
  class: client
}
c: Cluster {
  direction: right
  n1: Server 1 {
    class: server
  }
  n2: Server 2 {
    class: server
  }
}
client -> c.n1: "1. Write a record"
c.n1 -> c.n1: "2. Calculate appropriate node"
c.n1 -> c.n2: "3. Forward"
```

## Gossip Rounds

**Gossip Rounds** form the backbone of the protocol.
Periodically, each node selects random peers and shares its current state (e.g., heartbeat),
progressively spreading updates throughout the cluster.

For example, at time `3`, `A` gossips its state. `B` receives it and forwards it to `C`.

```d2
grid-rows: 2
vertical-gap: 100
c: "Cluster (00:03)" {
  grid-rows: 1
  horizontal-gap: 100
  a: A {
    m: |||yaml
    B:
      State: UP
      Heartbeat: 1
    C:
      State: UP
      Heartbeat: 1
    |||
  }
  b: B {
    m: |||yaml
    A:
      State: UP
      Heartbeat: 1
    C:
      State: UP
      Heartbeat: 1
    |||
  }
  c: C {
    m: |||yaml
    A:
      State: UP
      Heartbeat: 1
    B:
      State: UP
      Heartbeat: 1
    |||
  }
}
c3: "Cluster (A gossips)" {
  grid-rows: 1
  horizontal-gap: 100
  a: A {
    m: |||yaml
    B:
      State: UP
      Heartbeat: 1
    C:
      State: UP
      Heartbeat: 1
    |||
  }
  b: B {
    m: |||yaml
    A:
      State: UP
      Heartbeat: 3 (New)
    C:
      State: UP
      Heartbeat: 1
    |||
  }
  c: C {
    m: |||yaml
    A:
      State: UP
      Heartbeat: 3 (New)
    B:
      State: UP
      Heartbeat: 1
    |||
  }
  a -> b: "A: UP" {
    style.bold: true
  }
  b -> c: "A: UP" {
    style.bold: true
  }
}
```

Later, `B` also gossips its state to others.

```d2
grid-rows: 2
c4: "Cluster (B gossips)" {
  grid-rows: 1
  horizontal-gap: 100
  a: A {
    m: |||yaml
    B:
      State: UP
      Heartbeat: 3 (New)
    C:
      State: UP
      Heartbeat: 1
    |||
  }
  b: B {
    m: |||yaml
    A:
      State: UP
      Heartbeat: 3
    C:
      State: UP
      Heartbeat: 1
    |||
  }
  c: C {
    m: |||yaml
    A:
      State: UP
      Heartbeat: 3
    B:
      State: UP
      Heartbeat: 3 (New)
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

If the heartbeat timeout is `3 seconds`, by time `5`, both `A` and `B` consider `C` down due to the absence of updates.

```d2
grid-columns: 2
vertical-gap: 100
c: "Cluster (HeartbeatLifetime = 3, Time = 5)" {
  n1: A {
    m: |||yaml
    B:
      State: UP
      Heartbeat: 3
    C:
      State: UP
      Heartbeat: 1 (Expired)
    |||
  }
  n2: B {
    m: |||yaml
    A:
      State: UP
      Heartbeat: 3
    C:
      State: UP
      Heartbeat: 1 (Expired)
    |||
  }
  n3: C {
    class: generic-error
  }
  n1 -> n3: Mark DOWN
  n2 -> n3: Mark DOWN
}
```

### Temporary Promotion

When a node is deemed unreachable, the detector stops forwarding data to it and switches to a replica.
Once the original node recovers, it synchronizes with the replica to restore consistency.

```d2
grid-columns: 2
vertical-gap: 100
e: Node failure {
   n1: Server A {
    class: server
   }
   n2: Server C (Primary) {
    class: generic-error
   }
   n3: Server B (Replica) {
    class: server
   }
   n1 -> n2: Stop forwarding {
    class: error-conn
   }
   n1 -> n3: Failover {
    style.animated: true
   }
}
r: Recovery {
   n1: Server A {
    class: server
   }
   n2: Server C (Primary) {
    class: server
   }
   n3: Server B (Replica) {
    class: server
   }
   n1 -> n2: Back to primary {
    style.animated: true
   }
   n2 <- n3: Pull data to recover
}
```

## Data Conflicts

This architecture emphasizes **Availability over Consistency (AP)**.
During network partitions, replicas can be promoted to serve writes temporarily.
However, once partitions heal, conflict resolution becomes necessary.

Two effective approaches to conflict resolution are:

### Last Write Wins (LWW)

This method uses timestamps to resolve conflicts.
Each record tracks the latest update time, and the most recent version wins during a merge.

```d2
s1: Server 1 {
    r: |||yaml
    Id: 123
    Name: John
    Updated: 00:01
    |||
}

s2: Server 2 {
    r: |||yaml
    Id: 123
    Name: Mike
    Updated: 00:03
    |||
}
s1 -> r
s2 -> r: Server 2 wins {
  style.bold: true
}
r: |||yaml
  Id: 123
  Name: Mike
|||
```

This strategy relies on **clock synchronization**. If clocks are skewed, incorrect records might be selected.

### Vector Clocks

A more robust alternative, **Vector Clocks**, avoids dependency on synchronized clocks.
Instead, each record tracks a version vector in the format `[(Server, Version)]`.
This allows servers to detect conflicting updates.

For example, a record initially receives its version number from its owner shard.
Other servers then replicate the record from the owner, preserving this version for consistency.

```d2
c: Cluster {
  p1: Server 1 {
      r: |||yaml
      Id: 123
      Name: John
      VL:
      - Server 1: 3
      |||
  }
  p2: Server 2 {
    r: |||yaml
    Id: 123
    Name: John
    VL:
    - Server 1: 3
    |||
  }
  p3: Server 3 {
    r: |||yaml
    Id: 123
    Name: John
    VL:
    - Server 1: 3
    |||
  }
  p1 -> p2
  p1 -> p3
}
```

Then, a network partition occurs, the cluster is divided into three partitions.
Each group may update the record independently, expanding the vector clock;

- `Server 1` updates and increases its version.
- `Server 3` updates and creates its own version.

```d2
grid-rows: 2
cs: "" {
  class: none
  grid-rows: 1
  horizontal-gap: 100
  c1: Client 1 {
    class: client
  }
  c2: Client 2 {
    class: client
  }
}
c: Cluster {
  grid-rows: 1
  p1: Partition 1 {
    s1: Server 1 {
      r: |||yaml
      Id: 123
      Name: John Doe
      VL:
      - Server 1: 4
      |||
    }
  }
  p2: Partition 2 {
    s2: Server 2 {
      r: |||yaml
      Id: 123
      Name: John
      VL:
      - Server 1: 3
      |||
    }
  }
  p3: Partition 3 {
    p: Server 3 {
      r: |||yaml
      Id: 123
      Name: John Wick
      VL:
      - Server 1: 3
      - Server 3: 1
      |||
    }
  }
}
cs.c1 -> c.p1: "SET Name = John Doe" {
  style.bold: true
}
cs.c2 -> c.p3: "SET Name = Wick" {
  style.bold: true
}
```

A conflict is identified when vectors cannot be merged deterministically:

- `[(Peer 1, 4)]` is a clear successor of `[(Peer 1, 3)]`.
- But `[(Peer 1, 4)]` and `[(Peer 1, 3), (Peer 3, 1)]` are concurrent and conflict.

The system maintains both versions:

```d2
c: Cluster {
    p1: Server 1 {
      r: |||yaml
      Id: 123
      Name: John Doe
      VL:
      - Server 1: 4
      |||
    }
    p2: Server 2 {
      r: |||yaml
      Id: 123
      Name: John
      VL:
      - Server 1: 3
      |||
    }
    p3: Server 3 {
      r: |||yaml
      Id: 123
      Name: John Wick
      VL:
      - Server 1: 3
      - Server 3: 1
      |||
    }

    p1 -> m1: Override
    p2 -> m1
    m1: |||yaml
    Id: 123
    Name: John Doe
    VL:
    - Server 1: 4
    |||
    m2: |||yaml
    Id: 123
    Version 1:
      Name: John Doe
      VL:
      - Server 1: 4
    Version 2:
      Name: John Wick
      VL:
      - Server 1: 3
      - Server 3: 1
    |||
    m1 -> m2: Conflict
    p3 -> m2: Conflict
}
```

#### Application-Level Resolution

The application is responsible for resolving conflicts. It receives all conflicting versions and decides how to merge them.

For instance, it may choose to accept the first version:

```d2
grid-rows: 1
horizontal-gap: 150
d: Database {
    class: db
}
r1: |||yaml
Id: 123
Version 1:
  Name: John Doe
  VL:
  - Server 1: 4
Version 2:
  Name: John Wick
  VL:
  - Server 1: 3
  - Server 3: 1
|||
s: Application {
    class: process
}
r2: |||yaml
Id: 123
Name: John Doe
VL:
- Server 1: 4
|||
s <- r1 <- d
s -> r2: Accept first version
```

Since the database lacks context about the business logic,
deferring resolution to the application ensures safer and more flexible conflict handling.
Thus, applications should include **additional metadata** in their records to assist in this process.
