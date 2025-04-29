---
title: Decentralized Cluster
weight: 10
---

We have extensively discussed data sharding and replication.
Now, the question arises: **how can we effectively combine them into a single virtual database?**
Maintaining a {{< term p2p >}} (peer-to-peer) system is no simple task.
This architecture embraces decentralization, aiming for high availability and fault tolerance,
with **no single point of failure**.

Unlike the [Master-slave](../../master-slave-architecture) model,
where a shared master server coordinates the cluster,
a decentralized cluster must ensure that metadata (e.g., member addresses, sharding information, etc.)
is both reliable and consistently shared across all members.
This consistency is critical for enabling operations like replication, sharding, and promotion.

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

Before moving forward, we need to explore the **CAP Theorem**,
a fundamental trade-off that governs distributed systems.

Distributed systems are characterized by three key properties:
**Consistency**, **Availability** and **Partition Tolerance**.

### 1. Consistency (C)

**Consistency** in the context of {{< term cap >}}
means that all nodes see the same data at the same time,
either immediately or eventually through synchronization mechanisms.

### 2. Availability (A)

**Availability (A)** —
as discussed in the [Service Cluster](../../web-service/service-cluster#service-availability) section —
ensures that every request receives a (non-error) response,
even if the response contains outdated or inconsistent data.
In short, a system is considered **available** as long as it responds, regardless of accuracy.

{{< callout type="info">}}
We've only briefly touched on **Consistency** and **Availability** here;  
deeper explanations will follow in the sections below.
{{< /callout >}}

### 3. Partition Tolerance (P)

To understand **Partition Tolerance**, we first need to grasp the concept of a **Partition**:

#### Network Partition

A **network partition** occurs when failures split a cluster into isolated groups of nodes,  
preventing them from communicating with one another.

For example, consider a cluster of three servers that constantly cooperate to maintain synchronization:

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

Now imagine a network failure disrupts communication between `Server C` and the others:  
The cluster splits into two isolated partitions: `Partition 1 (A, B)` and `Partition 2 (C)`.

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

**Partition Tolerance (P)** is a system’s ability to continue functioning correctly despite these network partitions.

## CAP Theorem

The **CAP theorem** states that a distributed database can satisfy **only two**
of the following three properties simultaneously: **Consistency**, **Availability**, and **Partition Tolerance**.

Thus, practical systems must choose between three design patterns: **AP**, **CP**, or **CA**.

### CA System

A **CA** system provides **Consistency** and **Availability** but not **Partition Tolerance**.
In theory, this sounds ideal, but in practice, it’s **impractical**.

When a network partition occurs, a **CA** system would either stop working entirely or behave incorrectly —
both outcomes are unacceptable.
Since network partitions are inevitable in real-world environments,
a system that does not tolerate partitions is essentially unusable.

Thus, the real-world battle comes down to **AP** vs **CP**.  
In the presence of a partition, a distributed system must choose between **Consistency** and **Availability**.

### CP (Consistency over Availability) Systems

Consider a cluster of two servers:

- `A` hosts `Shard 1`; `B` maintains a replica of it.
- If clients write to `Shard 1` via `B`, `B` forwards the request to `A` (the shard owner).

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

Suppose a network partition occurs, separating `A` from `B`.
Now, clients connecting to `B` can **only read** from the replica — **writes are disabled** to preserve consistency.

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

This is a **CP** system:
it prioritizes **Consistency** over **Availability**, sacrificing write operations on isolated replicas.

### AP (Availability over Consistency) Systems

Now, let's modify the previous example to favor **Availability**.
Instead of disabling writes, `B` temporarily accepts writes even while partitioned from `A`.

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

In this **AP** system, partitions remain fully functional —
but at the cost of **Consistency**: different partitions may accept conflicting updates.

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

**Important:**  
Consistency here refers to **cross-partition consistency** during a network split,
not the usual node-to-node replication consistency.
Since partitions **cannot communicate**, inconsistencies persist until the cluster is healed.
After recovery, conflict resolution strategies must be applied

Choosing between **Consistency** and **Availability** is a fundamental decision when designing a distributed database.  
In the following sections, we will explore two major approaches for managing decentralized clusters:

- {{< term gosProto >}}.
- {{< term consProto >}}.
