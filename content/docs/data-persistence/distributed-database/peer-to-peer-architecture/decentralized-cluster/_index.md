---
title: Decentralized Cluster
weight: 10
---

We have extensively discussed data sharding and replication.
But how can we effectively combine them into a single virtual database?
Maintaining a {{< term p2p >}} system is no easy feat.
This architecture advocates a decentralized approach,
aiming to remain highly available and fault-tolerant, with **no single point of failure**.

Without a shared master server as the [Master-slave](../../master-slave-architecture) model,
cluster metadata (e.g., member addresses, sharding information, etc.)
must be somehow both reliable and consistently shared among cluster members
to enable tasks such as replication, sharding, and promotion.

```d2
p1: Peer 1 {
    class: server
}
p2: Peer 2 {
    class: server
}
p3: Peer 3 {
    class: server
}
p1 <-> p2 {
    style.animated: true
}
p2 <-> p3 {
    style.animated: true
}
p1 <-> p3 {
    style.animated: true
}
```

## Distributed Properties

Before going further,
we need to learn about **CAP Theorom**,
which is the key tradeoff in distributed cluster.

There are three common characteristics of distributed systems: {{< term cons >}}, {{< term av >}} and {{< partTol >}}.

### 1. Consistency (C)

{{< term cons >}} in {{< term cap >}} means that nodes in a cluster will possess the same data
through a synchronization method, even eventual consistency.

### 2. Availability (A)

**Availability (A)**, as discussed in the [Service Cluster](../../web-service/service-cluster#service-availability) section,
**Availability** ensures that a system is reachable and responsive at any given time.
Please note that this only mentions the responsiveness of a system,
depsite returning errors and inconsistent data, it's still available.

{{< callout type="info">}}
We've just sparely talked about **Consistency** and **Availability** here,
as we need more knowledge to clearly understand them in the sections below.
{{< /callout >}}

### 3. Partition Tolerance (P)

To understand this property. First, we need to know what a partition:

#### Network Partition

**Network partitioning** (aka **Partition**) is a network failure that splits a cluster into isolated partitions,
preventing them from communicating with each other.

For example, we have a cluster of 3 servers.
They constantly coop with each other to maintain the cluster.

```d2
sa: Server A {
  class: server
}
sb: Server B {
  class: server
}
sc: Server C {
  class: server
}
sa <-> sb <-> sc <-> sa
```

Then, a network problem happens,
the connections between `C` with `A` and `B` are corrupted.
The failure divides the cluster into two partitions: `Partition 1 (A, B)` and `Partition 2 (C)`.

```d2
direction: right
c1: "" {
  sa: Server A {
    class: server
  }
  sb: Server B {
    class: server
  }
  sc: Server C {
    class: server
  }
  a <-> b
  c <-> a {
    class: error-conn
  }
  c <-> b {
    class: error-conn
  }
}
c2: "" {
  g1: Partition 1 {
      sa: Server A {
          class: server
      }
      sb: Server B {
          class: server
      }
      sa <-> sb
  }
  g2: Partition 2 {
      sc: Server C {
          class: server
      }
  }
}
c1 -> c2
```

**Partition Tolerance (P)** refers to a
system’s ability to continue functioning despite a network partition between nodes.

## CAP Theorem

The **CAP theorem** states that a distributed database can provide **at most two** of the three properties.  
Therefore, there are three possible combinations: **AP**, **CP**, and **CA**.

### CA System

A **CA** system provide {{< term cons >}} and {{< term av >}} but {{< term partTol >}},
that type of system is possible but **impractical**.
A **CA** system cannot tolerate network partitions, meaning whenever a partition occurs,
despite a single node, the system will malfunction or stop working entirely,
both of the options are bad and unacceptable.

More importantly, network partitions are unavoidable and often happen,
a system that not providing the **Partition Tolerance (P)** property is good for nothing.

Thus, the battle remains between **AP** and **CP**.  
In the event of network partitions, following {{< term cap>}},
the system must choose between **Consistency (C)** or **Availability (A)**.

### CP (Consistency over Availability) System

Let's see an example cluster of 2 servers:

- `A` holds `Shard 1`, `B` contains a replica of the shard.
- If clients write to `Shard 1` from `B`, B can forward the request to shard owner `A`.

```d2
client: Client {
    class: client
}
c: Cluster {
    sa: Server A {
        s: Shard 1
    }
    sb: Server B {
        s: Replica 1
    }
    sa.s -> sb.s {
        style.animated: true
    }
}
client -> c.sb: '1. Write to "Shard 1"'
c.sb -> c.sa: 2. Forward to the owner
```

Unfortunately, a partition happens, the cluster is divided into 2 partitions.
`Partition 2` can't contact with `A` (the owner of `Shard 1`);
therefore, when clients connnect to `Partition 2`, they can only read data from `Shard 1`.

```d2
c: Cluster {
    g1: Partition 1 {
        sa: Server A {
            s: Shard 1
        }
    }
    g2: Partition 2 {
        sb: Server B {
            s: Replica 1
        }
    }
    g1.sa.s <-> g2.sb.s: Disconnected {
        class: error-conn
        style.animated: true
    }
}
client: Client {
    class: client
}
client -> c.g1.sa : Read and write
client -> c.g2.sb: Read only
```

This is a **CP** setup.
In this system, we prefer **Consistency** over **Availability**,
sacrificing the `Partition 2` writing capability.

### AP (Availability over Consistency) System

Let's see how **AP** is implemented in the previous example.
Instead of stoping,
we will allows writing data to `Replica 1` on `B` to temporarily.

```d2
c: Cluster {
    g1: Partition 1 {
        sa: Server A {
            s: Shard 1
        }
    }
    g2: Partition 2 {
        sb: Server B {
            s: Replica 1
        }
    }
    g1.sa <-> g2.sb: Disconnected {
        class: error-conn
        style.animated: true
    }
}
client: Client {
    class: client
}
client -> c.g1.sa : Read and write
client -> c.g2.sb: Read and write
```

In this setup, **Avalability** is preferred to **Consistency**,
we try to maintain the full functionality of parititions.
The problem is,
data can be written differently in different partitions,
leading to inconsistencies.
If clients connect to different partitions,
they'll see different versions of data.

```d2
c: Cluster {
    g1: Partition 1 {
        sa: Server A {
            data: |||yaml
            user:
                id: 10
                name: John
            |||
        }
    }
    g2: Partition 2 {
        sb: Server B {
            data: |||yaml
            user:
                id: 10
                name: Doe
            |||
        }
    }
    g1.sa <-> g2.sb: Disconnected {
        class: error-conn
        style.animated: true
    }
}
c1: Client 1 {
    class: client
}
c2: Client 2 {
    class: client
}
c1 -> c.g1.sa : Write
c2 -> c.g2.sb: Write
```

**Consistency** we're talking here may not be as what you're expecting.
We've previously categorized into {{< term strCons >}} and {{< term eveCons >}},
demonstrating how data is replicated **between nodes**.
Briefly, if nodes are interconnected,
they can somehow synchronize data and ensure consistency.

However, in this topic, we've referring to the consistency **between partitions** in the event of a network partition.
Partitions are **unable to communicate**,
inconsistencies emerge at different partitions will stand still until the cluster is recovered.
Morever, changes can be conflicted with each other and need resolving, such as in the diagram above.

Choosing between **Consistency** and **Availability** is a crucial decision when designing a distributed database.  
Let's dive into two common approaches maintaining a decentralized cluster:

- {{< term gosProto >}} for **AP** systems.
- {{< term consProto >}} for **CP** systems.
