---
title: CAP Theorem
weight: 5
---
 
In this topic, we mention {{< term cap >}},
which is the key tradeoff in distributed databases.

There are three common characteristics of distributed systems: {{< term cons >}}, {{< term av >}} and {{< partTol >}}.

## 1. Consistency (C)

{{< term cons >}} in {{< term cap >}} means that nodes in a cluster will possess the same data.

## 2. Availability (A)

**Availability (A)**, as discussed in the [Service Cluster](Service-Cluster.md#service-availability) section,
**Availability** ensures that a system is reachable and responsive at any given time.

{{< callout type="info">}}
I've just sparely introduced them here,
as we need more knowledge to clearly understand them in the sections below.
{{< /callout >}}

## 3. Partition Tolerance (P)

To understand this property. First, we need to know what a partition:

### Network Partition

**Network partitioning** (long for **Partition**) is a network failure that splits a cluster into isolated partitions,
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

In the [Decentralized Cluster](../decentralized-cluster/) topic, we'll see how to deal with this problem.
Currently, we should recognize that **Partition Tolerance (P)** in {{< term cap >}} refers to a system’s ability to continue functioning
despite network partitions between nodes.

## CAP Theorem

The **CAP theorem** states that a distributed database can provide **at most two** of the three characteristics.  
Therefore, there are three possible combinations: **AP**, **CP**, and **CA**.

### CA System

A **CA** system provide {{< term cons >}} and {{< term av >}} but {{< term partTol >}}, that type of system is possible but **impractical**.
A **CA** system cannot tolerate network partitions, meaning whenever a partition occurs,
despite a single node, the system will malfunction or stop working entirely,
both of the options are bad and unacceptable.

More importantly, network partitions are unavoidable and often happen,
a system that not providing the **Partition Tolerance (P)** property is good for nothing.

Thus, the battle remains between **AP** and **CP**.  
In the event of network partitions, the system must choose between **Consistency (C)** or **Availability (A)**.

### CP (Consistency over Availability) System

Let's reuse the example above,
imagine `B` and `C` are `A`'s replicas,
and the cluster is divided into 2 partitions.

```d2
g1: Partition 1 {
    sa: Server A {
      class: server
    }
    sb: Server B {
      class: server
    }
    sa -> sb {
      style.animated: true
    }
}
g2: Partition 2 {
    sc: Server C {
      class: server
    }
}
sa -> g2.sc {
  class: error-conn
  style.animated: true
}
```

In a **CP** system, we will make the `Partition 2` stop functioning.
Only the `Partition 1` takes over the cluster

```d2
c: Client {
  class: client
}
g1: Partition 1 {
    sa: Server A {
      class: server
    }
    sb: Server B {
      class: server
    }
    sa -> sb {
      style.animated: true
    }
}
g2: Partition 2 {
    sc: Server C {
      class: server
    }
}
sa -> g2.sc {
  class: error-conn
  style.animated: true
}
c -> g1
c -> g2 {
  class: error-conn
}
```

In this type of system, we prefer **Consistency** over **Availability**.
We stop `C` because it is isolated and **not received** changes from `A`,
in other words, it's inconsistent with `A`.

**Consistency** we're talking here may not be as what you're expecting.
In the topic, we've categorized into {{< term strCons >}} and {{< term eveCons >}},
demonstrating how data is replicated **between nodes**.

However, in this topic, we've seeing the consistency **between partitions** in the event of a network partition.
Let's say in {{< term eveCons >}},
replicated data come late or even being dropped,
but their servers are interconnected and willing to resolve the problems.
Meanwhile, in a network partition,
servers living in different partitions are unable to communicate,
any inconsistencies due to the disconnection remains until the cluster is recovered.

There is a misconception saying that an {{< eveCons >}} cluster can't provide **CP**.
**AP** or **CP** is how a database cluster is maintained under network partitions,
if it guarantees that different partitions not contain

ow long a network partition occurs is unpredictable,
### AP (Availability over Consistency) System

Let's reuse the example above,
imagine `B` and `C` are `A`'s replicas,
and the cluster is divided into 2 partitions.

```d2
g1: Partition 1 {
    sa: Server A {
      class: server
    }
    sb: Server B {
      class: server
    }
    sa -> sb {
      style.animated: true
    }
}
g2: Partition 2 {
    sc: Server C {
      class: server
    }
}
sa -> g2.sc {
  class: error-conn
  style.animated: true
}
```

In an **AP** system, both `Partition 1` and `Partition 2` keep working together despite being disconnected,
meaning any server in the cluster continues serving.


**AP** system will prefer {{< term availability >}}

```d2
g1: Partition 1 {
    sa: Server A {
      class: server
    }
    sb: Server B {
      class: server
    }
    sa -> sb {
      style.animated: true
    }
}
g2: Partition 2 {
    sc: Server C {
      class: server
    }
}
sa -> g2.sc {
  class: error-conn
  style.animated: true
}
```
n

To better understand the difference between **AP** and **CP** systems, let’s explore two examples.

<var name="ap-system" value="
%d2-import%
direction: right
db: Database cluster {
  s1: Server 1 {
    p: Main shard {
    }
  }
  s2: Server 2 {
    p: Replica shard {
    }
  }
  n: Network failure {
    class: generic-error
  }
  s1.p -> n: Replicates {
    style.animated: true
  }
  n -> s2.p: X {
    style.font-color: red
    style.stroke: red
    style.stroke-dash: 3
  }
}
"/>

A chat system is a good example of this case.

Suppose we have two users, `A` and `B`, whose conversation is stored on shard `P`.  
The primary fragment is on server `S1`, with a replica on server `S2`.  
Currently, `S2` is **disconnected** from `S1` due to a network partition.

```d2
%ap-system%
```

The system remains available.
When user `A` sends a message to the main shard `S1`, and user `B` happens to connect to the same server, `B` will read consistent data.

```d2
%ap-system%
ca: Client A {
  class: client
}
cb: Client B {
  class: client
}
ca -> db.s1.p: Updates
cb -> db.s1.p: Reads
```

However, if client `B` connects to `S2`,
it will read outdated messages because `S2` has not synchronized with `S1` during the partition.

```d2
%ap-system%
ca: User A {
  class: client
}
cb: User B {
  class: client
}
ca -> db.s1.p: Updates
cb -> db.s2.p: Reads outdated data
```

Once the network partition resolves, `S2` will synchronize with `S1`,
and the system will recover to a consistent state.

```d2
%ap-system%
ca: User A {
  class: client
}
cb: User B {
  class: client
}
db: {
  n.style.opacity: 0
  (s1.p -> n)[0].style.opacity: 0
  (n -> s2.p)[0].style.opacity: 0
  s1.p -> s2.p: Replicates {
    style.animated: true
  }
}
ca -> db.s1.p: Updates
cb -> db.s2.p: Reads consistent data
```

In this scenario, we prioritize **availability**, and temporary inconsistency is acceptable.  
When a client sends a request, the system always provides an associated response.

### CP (Consistency over Availability)

<var name="cp-system" value="
%ap-system%
"/>

A banking system serves as a prime example of this case.

Suppose we have a user, `A`, whose account data is stored on shard `P`.  
The main shard is located on server `S1`, with a replica on server `S2`.  
Currently, `S2` is **disconnected** from `S1` due to a network error.

```d2
%cp-system%
```

`S2` will enter an **unavailable** state and stop serving requests.
Why? Imagine if `S2` were to remain operational:

- User `A` initiates a payment request through `S1`, successfully reducing their balance.
- Later, `A` refreshes the application and connects to `S2`,
which still shows the **old balance** because it hasn’t synchronized with `S1`.

```d2
%cp-system%
ua: User A {
  class: client
}
ua -> db.s1.p: 1. Updates balance
ua <- db.s1.p: 2. Responds successfully
ua -> db.s2.p: 3. Gets balance
ua <- db.s2.p: 4. Gets the old balance
```

- This inconsistency could lead to severe issues, such as double spending—clearly unacceptable in a critical system like this.

By sacrificing the **availability** of `S2`, we ensure the **strong consistency** of the system.  
Even though `S2` is temporarily unavailable, the system guarantees that all data remains accurate and consistent.

## Conclusion

Choosing between **Consistency** and **Availability** is a crucial decision when designing a distributed database.  

[Database systems designed with traditional ACID guarantees in mind such as RDBMS choose consistency over availability,
whereas systems designed around the BASE philosophy,
common in the NoSQL movement for example, choose availability over consistency.](https://en.wikipedia.org/wiki/CAP_theorem)
