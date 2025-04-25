---
title: Consensus Protocol
weight: 20
---

We can observe that the **Gossip Protocol** basically does not afford [strong consistency](Distributed-Database.md#strong-consistency-level),
as the cluster can be divided into autonomous partitions, causing temporary inconsistencies.
It prefers **Availability** to **Consistency**, we will see this bias clearer in the [CAP topic](CAP-Theorem.md).

### Consensus

To support strong consistency, many systems use a protocol called **Consensus**.
**Consensus Protocol** enables a group of nodes to reach consensus on values,
even in the presence of failures (fault tolerance).

Let's say we have a cluster of nodes.
A new node is about to join the cluster,
in a **Consensus** cluster, it's unable to easily ask a random node (e.g. `Node A`) to do that

```d2
c: Cluster {
   a: Node A {
      shape: circle
   }
   b: Node B {
      shape: circle
   }
   c: Node C {
      class: generic-error
   }
}
n: D (New node) {
   shape: circle
}
n -> c.a: Join {
   class: error-conn
}
```

The cluster needs to reach a decision approving the new node.
Most of the time, a decision is made when **any majority of nodes** agrees on it,
e.g. node `A` and `B` consent `D` to join the cluster,
although `C` is going down, `D` still becomes a member.

```d2
c: Cluster {
   a: Node A {
      shape: circle
   }
   b: Node B {
      shape: circle
   }
   c: Node C (down) {
      class: generic-error
   }
   a <-> b: Approve
}
n: D (New node) {
   shape: circle
}
n -> c.a: Join
```

**Consensus Protocol** is just an abstract piece of theory with [strict requirements](https://en.wikipedia.org/wiki/Consensus_(computer_science)).
We won't go into the details as it will take a lot of time to make clear.
Instead, we will learn about a real implementation - [Raft Consensus Algorithm](https://raft.github.io/).

### Raft Consensus Algorithm

**Raft** is a consensus protocol designed to manage a replicated log in distributed systems.
**Raft** managing a cluster with two critical factors.

#### Majority

**Raft** is tolerant with network partitions as long as a **majority** of nodes can communicate.
For example, a **Raft** cluster encounters a network partition and becomes 2 partitions.
The partition `A` continues operating as it has a majority of members,
and `B` becomes unavailable.

```d2
c: Cluster {
   a: Node A {
      shape: circle
   }
   b: Node B {
      shape: circle
   }
   c: Node C {
      shape: circle
   }
   d: Node D {
      shape: circle
   }
   a <-> b <-> c <-> d
}
c: Network partition {
   a: Partition A {
      a: Node A {
         shape: circle
      }
      b: Node B {
         shape: circle
      }
      c: Node C {
         shape: circle
      }
      a <-> b <-> c
   }
   b: Partition B (Unavailable) {
      d: Node D {
         shape: circle
      }
   }
}
```

How about 2 partitions of 2 nodes?
Unfortunately, the cluster will become **unavailable totally**,
as we cannot provide major decisions any longer.
This problem is called **total network failure**.

```d2
c: Cluster {
   a: Node A {
      shape: circle
   }
   b: Node B {
      shape: circle
   }
   c: Node C {
      shape: circle
   }
   d: Node D {
      shape: circle
   }
   a <-> b <-> c <-> d
}
c: Network partition {
   a: Partition A {
      a: Node A {
         shape: circle
      }
      b: Node B {
         shape: circle
      }
      a <-> b
   }
   b: Partition B {
      c: Node C {
         shape: circle
      }
      d: Node D {
         shape: circle
      }
      c <-> d
   }
}
```

As a side benefit, a **Raft** cluster does not generate conflict,
as it makes sure there is always a partition to work at a time.

#### Leader

**Raft** resolves around election processes to elect a temporary leader.
That means, a `Raft` cluster has **at most one** (0 to 1) `Leader`, the others are `Followers`.
After a leader is selected, other nodes must respect and follow its demands.

In essence, a node becomes the leader if it's voted by a majority of nodes.
The leader continuously stays in power while it is still available,
an election will take place when the leader is **unknown**, when:

1. The cluster starts to work and it has no leader
2. Or the leader has just crashed

##### Heartbeat

How is the leader detected as crashed?
We define a common `timeout` for the cluster,
and let the leader periodically send **heartbeats** to the members.
If the last heartbeat of a node expires, the node supposes the leader as unavailable

E.g., `Node B` times out because its heartbeat has expired

```d2

c: Cluster with (Timeout = 2 seconds, Current time = 19:00:03) {
  n1: Node A (Good) {
    grid-columns: 1
    grid-gap: 0
    states: "Heartbeat = 19:00:02"
  }
  n2: Node B (Times out) {
    grid-columns: 1
    grid-gap: 0
    style.font-color: ${colors.e}
    states: "Heartbeat = 19:00:00"
    to: "Current time - Hearbeat = 3 seconds > Timeout (2 seconds)" {
      style.font-color: ${colors.e}
    }
  }
}
```

When a node times out, it transitions to a `Candidate` and starts an election.
That means any node of the cluster can autonomously initiate an election.

##### Election

To vote reliably, each node has a private number called `Term` representing its passed elections,
e.g., a node possesses the term of `3` means it has already passed `3` elections.
Therefore, a node only votes for notes holding a **larger** term, implying new elections.

Let's see how `Timeout`, `Heartbeat` and `Term` are used for elections.
Imagine we establish a cluster of three nodes with `Timeout = 3 seconds`
<var name="cluster" value="

c: 'Current time = 00:03, Timeout = 3s' {
  n1: Node A {
    grid-gap: 0
    grid-columns: 1
    State: 'Follower'
    Term: '0'
    Heartbeat: '00:00'
  }
  n2: Node B {
    grid-gap: 0
    grid-columns: 1
    State: 'Follower'
    Term: '1'
    Heartbeat: '00:01'
  }
  n3: Node C {
    grid-gap: 0
    grid-columns: 1
    State: 'Follower'
    Term: '0'
    Heartbeat: '00:01'
  }
}
"/>

```d2
%cluster%
```

<procedure>
<step>

A follower node times out.
The node marks it as a candidate, increases its **term** by 1 and initiates a voting process to other nodes.
E.g., `Node A` times out and starts a new election

```d2
%cluster%
c.n1: {
  label: "Node A times out"
  style.font-color: ${colors.e}
  State: 'Follower -> Candidate' {
    style.font-color: ${colors.e}
  }
  Term: '0 -> 1' {
    style.font-color: ${colors.e}
  }
}
```

</step>

<step>

A node only votes for a newer election

- If the node hasn't participated in the election (or its term < the candidate's term), it votes for the node and increases its term to the node's value.
- If the node has already passed the election (or its term >= the candidate's term), it ignores the request.

E.g., `Node B` ignores because its term is equal to the candidate's term

```d2
%cluster%
c: {
  n1: {
    label: "Node 1"
    style.font-color: ${colors.e}
    State: 'Candidate'
    Term: '1'
  }
  n2: {
    Term: {
      style.font-color: ${colors.e}
    }
  }
  n3: {
    Term: '0 -> 1' {
      style.font-color: ${colors.e}
    }
  }
  n2 -> n1: Ignores the voting
  n3 -> n1: Elects the candidate and updates term
}
```

</step>

<step>

A new leader must be voted from a **majority** of the nodes.
E.g., `Node A` wins the voting from the majority of nodes (`Node A`, `Node C`)

```d2
%cluster%
c: {
  n1: {
    label: "Node A (Leader)"
    style.font-color: ${colors.e}
    Term: '1'
    State: 'Candidate -> Leader' {
      style.font-color: ${colors.e}
    }
  }

  n3: {
    Term: '1'
  }
  n1 -> n2: Replicates log {
    style.animated: true
  }
  n1 -> n3: Replicates log {
    style.animated: true
  }
}
```

</step>

</procedure>

What if we have no majority or equal votes?
The process comes to a phase called **Split Vote**

- Some systems let the candidate initiate a new election by increasing its term and again requesting votes
- Or simply, some systems select between the candidates randomly
- ...

The term value in **Raft** also serves as a logical clock.
If a candidate has a higher term than others, it signifies that more time has passed

- Nodes time out and increase its term after a **shared** duration, a node with the lower term must live shorter the higher ones
- If a node escalates its term from a higher node, it must walk behind that node

#### Log Replication

After a leader is elected,
changes in the cluster must be saved in the leader's log first,
and further replicated to followers.

For example, a node wants to join a cluster.
It needs to contact the leader first,
this participation is later propagated to other nodes

```d2

c: Cluster {
   a: Leader {
      shape: circle
   }
   b: Follower 1 {
      shape: circle
   }
   c: Follower 2 {
      shape: circle
   }
}
n: D (New node) {
   shape: circle
}
n -> c.a: Join 
c.a -> c.b: Replicated {
   style.animated: true
}
c.a -> c.c: Replicated {
   style.animated: true
}
```
