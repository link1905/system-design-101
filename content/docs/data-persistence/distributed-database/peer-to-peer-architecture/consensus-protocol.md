---
title: Consensus Protocol
weight: 20
---

We can observe that the **Gossip Protocol** essentially does not guarantee **Consistency**,  
as the cluster can be divided into independent partitions.

## Consensus

To build a **CP (Consistency over Availability)** system, many architectures adopt a protocol called **Consensus**.
{{< term consProto >}} enables a group of nodes to agree on a value,
even in the presence of failures (Fault Tolerance).

Consider a cluster of nodes.
When a new node attempts to join,
in a **Consensus-based** system, it cannot simply ask a random node (e.g., `Node A`) for admission:

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

Instead, the cluster must **reach a collective decision** to approve the new node.  
This typically happens when a **majority of nodes** agree—for example,
if nodes `A` and `B` approve `D`'s entry, it succeeds even if node `C` is down.

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

The **Consensus Protocol** is an abstract theoretical concept
with [strict requirements](https://en.wikipedia.org/wiki/Consensus_(computer_science)).
We won’t dive into all the theoretical details here.
Instead, we’ll focus on a practical implementation: the [Raft Consensus Algorithm](https://raft.github.io/)

## Raft Consensus Algorithm

**Raft** is a consensus protocol designed to manage a replicated log in distributed systems.
It manages a cluster using two key principles:

### Majority

**Raft** can tolerate network partitions as long as a **majority** of nodes remain connected.
If the cluster splits into some partitions, the one containing a majority continues to operate,
while the minority partitions becomes unavailable.

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
   a: Partition 1 {
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
   b: Partition 2 (Unavailable) {
      d: Node D {
         shape: circle
      }
   }
}
```

#### Loss Of Quorum

What if the cluster splits into **two equal partitions** (e.g., 2 nodes each)?  
In that case, **neither** can achieve majority—resulting in **total unavailability** for writes.  
This situation is known as **Loss Of Quorum**.

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

#### CP Design

This is how we achieve **CP (Consistency over Availability)**:
A **Raft** cluster ensures only **one partition** (the one with a majority) can operate at a time.
This guarantees that at any moment, there is a **single writer**—ensuring consistency.

### Leader Node

**Raft** revolves around the concept of electing a **temporary leader**.
At any given time, a **Raft** cluster has **at most one Leader** (or none). All other nodes are **Followers**.

A node becomes leader if it wins a majority vote.
It remains leader as long as it is reachable. If it becomes unavailable, a new election is triggered.

#### Heartbeat

How does the cluster detect that a leader has failed?
Each node uses a shared **timeout** value. The leader must periodically send **heartbeats** to followers.
If a follower’s heartbeat expires, it assumes the leader is down.

For example, `Node B` detects leader failure due to a heartbeat timeout:

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
    to: "Current time - Heartbeat = 3 seconds > Timeout (2 seconds)" {
      style.font-color: ${colors.e}
    }
  }
}
```

Once a node suspects the leader has failed, it transitions to a **Candidate** and initiates an election.
Any node in the cluster can do this.

### Election Process

In an election, nodes vote for a candidate based on a **term number**—a logical counter of elections.

- A node with `Term = 3` has experienced three election rounds.
- Nodes only vote for candidates with **higher terms**, ensuring progress.

Let’s walk through an example with 3 nodes and a `Timeout = 3 seconds`:

```d2
c: 'Current time = 00:03, Timeout = 3s' {
  n1: Node A {
   shape: circle
   c: |||yaml
   State: Follower
   Term: 0
   Heartbeat: 00:00
   |||
  }
  n2: Node B {
   shape: circle
   c: |||yaml
   State: Follower
   Term: 1
   Heartbeat: 00:01
   |||
  }
  n3: Node C {
   shape: circle
   c: |||yaml
   State: Follower
   Term: 0
   Heartbeat: 00:01
   |||
  }
}
```

{{% steps %}}

#### Step 1: Timeout and Candidacy

`Node A` times out, becomes a **Candidate**, increases its term, and requests votes:

```d2
c: 'Current time = 00:03, Timeout = 3s' {
  n1: Node A times out {
   shape: circle
   c: |||yaml
   State: Follower -> Candidate
   Term: 0 -> 1
   Heartbeat: 00:01
   |||
  }
  n2: Node B {
   shape: circle
   c: |||yaml
   State: Follower
   Term: 1
   Heartbeat: 00:01
   |||
  }
  n3: Node C {
   shape: circle
   c: |||yaml
   State: Follower
   Term: 0
   Heartbeat: 00:01
   |||
  }
}
```

#### Step 2: Voting Rules

Nodes vote only for candidates with **higher terms**:

- If a node hasn't voted or the term is smaller, it votes.
- If the term is the same or higher, it ignores the request.

`Node B` ignores the vote (equal term), `Node C` votes and updates its term:

```d2
c: 'Current time = 00:03, Timeout = 3s' {
  n1: Node A {
   shape: circle
   c: |||yaml
   State: Candidate
   Term: 1
   Heartbeat: 00:01
   |||
  }
  n2: Node B {
   shape: circle
   c: |||yaml
   State: Follower
   Term: 1
   Heartbeat: 00:01
   |||
  }
  n3: Node C {
   style.font-color: ${colors.i2}
   c: |||yaml
   State: Follower
   Term: 0 -> 1
   Heartbeat: 00:01
   |||
  }
  n2 -> n1: Ignores the voting
  n3 -> n1: Elects the candidate and updates term
}
```

#### Step 3: Leader Confirmation

A **majority** vote is required.  
`Node A` gets votes from itself and `Node C`—it becomes the **Leader**.

```d2
c: 'Current time = 00:03, Timeout = 3s' {
  n1: Node A (Leader) {
   shape: circle
   c: |||yaml
   State: Candidate -> Leader
   Term: 1
   Heartbeat: 00:01
   |||
  }
  n2: Node B {
   shape: circle
   c: |||yaml
   State: Follower
   Term: 1
   Heartbeat: 00:01
   |||
  }
  n3: Node C {
   style.font-color: ${colors.i2}
   c: |||yaml
   State: Follower
   Term: 1
   Heartbeat: 00:01
   |||
  }
}
```

Then, the leader periodically transmits heartbeat messages to indicate its continued availability.

```d2
c: 'Current time = 00:04, Timeout = 3s' {
  n1: Node A (Leader) {
   shape: circle
   c: |||yaml
   State: Candidate -> Leader
   Term: 1
   Heartbeat: 00:04
   |||
  }
  n2: Node B {
   shape: circle
   c: |||yaml
   State: Follower
   Term: 1
   Heartbeat: 00:04
   |||
  }
  n3: Node C {
   style.font-color: ${colors.i2}
   c: |||yaml
   State: Follower
   Term: 1
   Heartbeat: 00:04
   |||
  }
  n1 -> n2: Heartbeat at 00:04
  n1 -> n3: Heartbeat at 00:04
}
```

{{% /steps %}}

Why is **Term** reliable?

- Nodes time out and increment their term after a **shared timeout** interval;
thus, a node with a lower term must have been active for a shorter duration than one with a higher term.
- If a node adopts a higher term from another node, it effectively acknowledges that the other node is ahead in time and must follow its lead.

#### Split Vote

What if no candidate achieves majority?

This results in a **Split Vote**:

- The cluster may retry with new terms and timeouts.
- Some implementations use randomized delays to break ties.

### Log Replication

Once a leader is elected, all state changes go through it.  
The leader writes changes to its log first and then **replicates** them to followers.

For example, when a new node wants to join, it contacts the leader:  
The leader logs the change and then replicates the update to others.

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
c.a -> c.b: Replicate {
   style.animated: true
}
c.a -> c.c: Replicate {
   style.animated: true
}
```
