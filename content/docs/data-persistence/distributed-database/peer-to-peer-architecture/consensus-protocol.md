---
title: Consensus Protocol
weight: 20
---

We can observe that the **Gossip Protocol** essentially does not guarantee **Consistency**,
as the cluster can be divided into independent partitions.

## Consensus

To build a **CP (Consistency over Availability)** system, many architectures adopt a protocol called **Consensus**.
{{< term consProto >}} enables a group of nodes to agree on a value,
even in the presence of failures ({{< term fauTol >}}).

Consider a cluster of nodes.
When a new node attempts to join,
in a **Consensus-based** system, it cannot simply ask a random node (e.g., `Node A`) for admission:

```d2
n: D (New node) {
  class: server
}
c: Cluster {
  grid-rows: 1
  a: Node A {
    class: server
  }
  b: Node B {
    class: server
  }
  c: Node C {
    class: generic-error
  }
}
n -> c.a: Join {
  class: error-conn
}
```

Instead, the cluster must **reach a collective decision** to approve the new node.
This typically happens when a **majority of nodes** agree—for example,
if nodes `A` and `B` approve `D`'s entry, it succeeds even if node `C` is down.

```d2
n: D (New node) {
  class: server
}
c: Cluster {
  grid-rows: 1
  horizontal-gap: 100
  a: Node A {
    class: server
  }
  b: Node B {
    class: server
  }
  c: Node C (down) {
    class: generic-error
  }
  a <-> b: Approve
}
n -> c.a: Join
```

The **Consensus Protocol** is an abstract theoretical concept
with [strict requirements](https://en.wikipedia.org/wiki/Consensus_(computer_science)).
We won’t dive into all the theoretical details here.
Instead, we’ll focus on a practical implementation: the [Raft Consensus Algorithm](https://raft.github.io/)

## Raft Consensus Algorithm

**Raft** is a consensus protocol designed to manage a replicated log in distributed systems.
It manages a cluster using two key principles:

### Majority

**Raft** can tolerate network partitions as long as a majority of nodes remain connected.
If the cluster splits into multiple partitions, the partition containing a majority continues to operate, while the minority partitions become unavailable.

```d2
c: Cluster {
  grid-columns: 2
  horizontal-gap: 100
  a: Node A {
    class: server
  }
  b: Node B {
    class: server
  }
  c: Node C {
    class: server
  }
  a <-> b
  b <-> c: Disconnect {
    class: error-conn
  }
  a <-> c: Disconnect {
    class: error-conn
  }
}

np: Network partition {
  grid-rows: 2
  a: Partition 1 (Available) {
    grid-rows: 1
    a: Node A {
      class: server
    }
    b: Node B {
      class: server
    }
    a <-> b
  }
  b: Partition 2 (Unavailable) {
    style.font-color: ${colors.e}
    c: Node C {
      class: server
    }
  }
}
```

#### Loss Of Quorum

What if the cluster splits into **two equal partitions** (e.g., one node each)?
In that case, none of them can achieve majority—resulting in **total unavailability** for writes.
This situation is known as **Loss Of Quorum**.

```d2
c: Network partition {
  grid-rows: 1
  horizontal-gap: 100
  p1: Partition 1 (Unavailable) {
    style.font-color: ${colors.e}
    a: Node A {
      class: server
    }
  }
  p2: Partition 2 (Unavailable) {
    style.font-color: ${colors.e}
    a: Node B {
      class: server
    }
  }
  p3: Partition 3 (Unavailable) {
    style.font-color: ${colors.e}
    a: Node C {
      class: server
    }
  }
}
```

#### CP Design

This is how we achieve **CP (Consistency over Availability)**:
A **Raft** cluster ensures only one partition (the one with a majority) can operate at a time.
This guarantees that at any moment, there is a single writer—ensuring consistency.

## Raft Cluter

Let’s explore how to build a distributed cluster using the Raft algorithm.

### Leader Node

**Raft** revolves around the concept of electing a **temporary leader**.
At any given time, a **Raft** cluster has **at most one Leader** (or none). All other nodes are **Followers**.

A node becomes leader if it wins a majority vote.
It remains leader as long as it is reachable.

#### Heartbeat

How does the cluster detect that a leader has failed?
Each node uses a fixed **timeout** value. The leader must periodically send **heartbeats** to followers.
If a follower’s heartbeat expires, it assumes the leader is down.

Once a node suspects the leader has failed, it transitions to a **Candidate** and initiates an election.
Any node in the cluster can do this.

For example, `Node B` detects leader failure due to a heartbeat timeout,
it then becomes a **Candidate**.

```d2
c: Cluster (Timeout = 3 seconds, Current time = 00:04) {
  n1: Node A {
    states: |||yaml
    Heartbeat: 00:02
    Leader: UP
    State: Follower
    |||
  }
  n2: Node B {
    style.font-color: ${colors.e}
    states: |||yaml
    Heartbeat: 00:00
    Leader: DOWN
    State: Candidate
    |||
  }
}
```


### Election Process

In the **Raft** election process, nodes vote for a candidate based on its term number—a logical counter representing election rounds.

- For example, a node with `Term = 3` has participated in three election rounds.
- Nodes will only vote for candidates with higher terms, which ensures the system can always make progress.

Let’s walk through an example with **three nodes** and a **timeout of 3 seconds**.
Suppose the leader has just become corrupted.

```d2
c: 'Current time = 00:03, Timeout = 3s' {
  n1: Node A {
   c: |||yaml
   State: Follower
   Term: 0
   Heartbeat: 00:00
   |||
  }
  n2: Node B {
   c: |||yaml
   State: Follower
   Term: 1
   Heartbeat: 00:01
   |||
  }
  n3: Node C {
   c: |||yaml
   State: Follower
   Term: 1
   Heartbeat: 00:02
   |||
  }
}
```

{{% steps %}}

#### Step 1: Timeout and Candidacy

`Node A` times out, transitions to a **Candidate** state, increments its term, and requests votes:

```d2
c: 'Current time = 00:03, Timeout = 3s' {
  n1: Node A times out {
   c: |||yaml
   State: Follower -> Candidate
   Term: 0 -> 1
   Heartbeat: 00:00
   |||
  }
  n2: Node B {
   c: |||yaml
   State: Follower
   Term: 1
   Heartbeat: 00:01
   |||
  }
  n3: Node C {
   c: |||yaml
   State: Follower
   Term: 1
   Heartbeat: 00:02
   |||
  }
}
```

#### Step 2: Voting Rules

Nodes only vote for candidates with **higher terms**:

- If a node hasn't voted or receives a higher term than itself, it will vote.
- If another candidate matches or has a lower term, the node ignores the request.

`Node B` and `Node C` both ignore `Node A`’s vote request since their term is the same.
`Node A` does not receive a majority and returns to the **Follower** state.

```d2
c: 'Current time = 00:03, Timeout = 3s' {
  n1: Node A {
   c: |||yaml
   State: Candidate -> Follower
   Term: 1
   Heartbeat: 00:01
   |||
  }
  n2: Node B {
   c: |||yaml
   State: Follower
   Term: 1
   Heartbeat: 00:01
   |||
  }
  n3: Node C {
   c: |||yaml
   State: Follower
   Term: 1
   Heartbeat: 00:02
   |||
  }
  n2 -> n1: Ignores the voting
  n3 -> n1: Ignores the voting
}
```

At time **00:04**, `Node B` times out, becomes a **Candidate**, and initiates a new election.

```d2
c: 'Current time = 00:04, Timeout = 3s' {
  n1: Node A {
   c: |||yaml
   State: Follower
   Term: 1
   Heartbeat: 00:01
   |||
  }
  n2: Node B {
   c: |||yaml
   State: Follower -> Candidate
   Term: 1 -> 2
   Heartbeat: 00:01
   |||
  }
  n3: Node C {
   c: |||yaml
   State: Follower
   Term: 1
   Heartbeat: 00:02
   |||
  }
}
```

`Node B` now has a higher term, so the other nodes vote for it and update their own terms accordingly.

```d2
c: 'Current time = 00:04, Timeout = 3s' {
  n1: Node A {
   c: |||yaml
   State: Follower
   Term: 1 -> 2
   Heartbeat: 00:01
   |||
  }
  n3: Node C {
   c: |||yaml
   State: Follower
   Term: 1 -> 2
   Heartbeat: 00:02
   |||
  }
  n2: Node B {
   c: |||yaml
   State: Candidate
   Term: 2
   Heartbeat: 00:01
   |||
  }
  n1 -> n2: Vote for
  n3 -> n2: Vote for
}
```

#### Step 3: Leader Confirmation

A majority vote is required, so `Node B` becomes the new **Leader**.
After becoming leader, it sends regular heartbeat messages to all nodes to show it is alive.

```d2
c: 'Current time = 00:04, Timeout = 3s' {
  n1: Node A {
   c: |||yaml
   State: Follower
   Term: 2
   Heartbeat: 00:01 -> 00:04
   |||
  }
  n3: Node C {
   c: |||yaml
   State: Follower
   Term: 2
   Heartbeat: 00:02 -> 00:04
   |||
  }
  n2: Node B {
   c: |||yaml
   State: Candidate -> Leader
   Term: 2
   |||
  }
  n2 -> n1: Heartbeat at 00:04
  n2 -> n3: Heartbeat at 00:04
}
```

{{% /steps %}}

The term number is a reliable logical counter:

- Nodes increment their term after each timeout, so a node with a lower term must have participated in fewer election cycles than one with a higher term.
- If a node adopts a higher term from another node, it implies it has fallen behind and is updating its view of the election history.

### Split Vote

What if no candidate achieves majority?
This results in a **Split Vote**:

- The cluster may retry with new terms and timeouts.
- Some implementations may select a random leader.

### Log Replication

Once a leader is elected, all state changes go through it.
The leader writes changes to its log first and then replicates them to followers.

For example, when a new node wants to join, it contacts the leader:
The leader logs the change and then replicates the update to others.

```d2
direction: right
c: Cluster {
   a: Leader {
    class: server
   }
   b: Follower 1 {
    class: server
   }
   c: Follower 2 {
    class: server
   }
}
n: D (New node) {
  class: server
}
n -> c.a: Join
c.a -> c.b: Replicate {
   style.animated: true
}
c.a -> c.c: Replicate {
   style.animated: true
}
```
