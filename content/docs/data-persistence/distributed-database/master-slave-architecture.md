---
title: Master-slave Architecture
weight: 10
prev: distributed-database
next: peer-to-peer-architecture
---

Many systems experience heavy read workloads, where read operations significantly outnumber write operations.
To meet this demand,
we can design a database architecture with a single writer (**Primary**) and multiple readers (**Replicas**).

- The writer propagates changes to the replicas.
- The replicas can serve read requests independently, offloading the primary and improving read scalability.

```d2
grid-rows: 1
horizontal-gap: 100
w: Client (Write) {
    class: client
}
dc: Database cluster {
  direction: right
    w: Primary {
      class: db
    }
    r1: Replica 1 {
      class: db
    }
    r2: Replica 2 {
      class: db
    }
    w -> r1: Replicate {
      style.animated: true
    }
    w -> r2: Replicate {
      style.animated: true
    }
}

r: Client (Read) {
    class: client
}
w -> dc.w: Write
r -> dc.r1: Read
r -> dc.r2: Read
```

This setup is commonly known as the {{< term maSl >}} **Architecture** (good name 🧐).

## Multi-master

Now, what happens if we allow **multiple writers**?
Would that significantly improve write performance?

```d2
dc: Database cluster {
  grid-rows: 2
  grid-gap: 100
  w1: Master 1 {
    class: db
  }
  w2: Master 2 {
    class: db
  }
  r1: Replica 1 {
    class: db
  }
  r2: Replica 2 {
    class: db
  }
  w1 <-> w2 {
    style.animated: true
  }
  w1 -> r1 {
    style.animated: true
  }
  w1 -> r2 {
    style.animated: true
  }
  w2 -> r2 {
    style.animated: true
  }
  w2 -> r1 {
    style.animated: true
  }
}
```

However, this paradigm does not enhance write throughput.
Every write operation must still be synchronized across all nodes in the cluster.
This contrasts with read replicas, where each read request can be independently handled by a single replica.

The key advantage of a **Multi-Master** setup lies in higher availability.
If one master fails, others can continue to process writes, avoiding downtime.

### SQL Paradox

The most widely adopted form of the {{< term maSl >}} model is {{< term sql >}} databases,
as a single writer makes it easier to maintain strong consistency for [ACID transactions]({{< ref "concurrency-control#acid" >}}).

Because of this, **Multi-Master** setups are rarely used in practice.
They don’t offer enough benefits to justify their complexity:

- If the masters collaborate to maintain {{< term acid >}}, they must compromise availability.
- If they asynchronously replicate, they risk violating {{< term acid >}} principles.

## Standby Promotion

Back to the {{< term maSl >}} model, the master handles all updates, becoming a {{< term spof >}} that can affect system availability.
To mitigate the impact of master failure, we can introduce a [Standby Server]({{< ref "distributed-database#standby-server" >}})
that is synchronously replicated from the master.
In the event of a failure, we can quickly **promote** the standby to become the new master.

## Centralized Cluster

The {{< term maSl >}} model is often deployed as a centralized cluster,
with a **Coordinator** that acts as the cluster's entry point.

Since each server has a predefined role (master or replica), the **Coordinator** can:

- Route write requests to the master.
- Distribute (aka load balancing) read requests across replicas.

```d2
direction: right
db: Database cluster {
  w: Master {
    class: db
  }
  r1: Replica 1 {
    class: db
  }
  r2: Replica 2 {
    class: db
  }
  c: Coordinator {
    class: server
  }
  c -> w: "Write"
  c -> r1: "Read"
  c -> r2: "Read"

}
s: Client {
    class: client
}
s -> db.c
```

Moreover, if the **Master** node becomes unresponsive,
the **Coordinator** detects the failure and promptly promotes a server to take over its responsibilities.

```d2
direction: right
c: Coordinator {
  class: server
}
m: Master {
  class: generic-error
}
r: Standby Server {
  class: server
}
c -> m: Detect failure
c -> r: Promote to master
```

### Connection Pooling

Opening a new database connection is both slow and resource-intensive.
If each user request triggers a new connection, it leads to performance issues.

**Connection Pooling** is a fundamental design pattern that enables the **reuse** of database connections.
Database connections are not immediately terminated but instead maintained in a pool for subsequent use.
A **Pool Manager** component serves as the central authority responsible for managing and coordinating these shared connections.

This functionality is integrated into the **Coordinator** to improve performance.

```d2
grid-rows: 1
horizontal-gap: 100
c: Client {
    class: client
}
p: Coordinator (Pool Manager) {
  vertical-gap: 50
  grid-rows: 2
  c1: Connection 1 {
    class: conn
  }
  c2: Connection 2 {
    class: conn
  }
}
s: Servers {
  class: db
}
c -> p
p.c1 <-> s
p.c2 <-> s
```

## Problems

The {{< term maSl >}} model is simple and intuitive.
Each component has a well-defined role,
and the direct communication between nodes results in **low latency** and **fast responses**.

However, this simplicity conceals several **critical issues**,
most of which stem from the centralized control of the master server:

- The master becomes the {{< term spof >}}.
Its failure halts **all write operations**,
therefore, the {{< term maSl >}} model does not guarantee {{< term ha >}}.

- The master quickly becomes a **performance bottleneck**, especially in write-heavy applications.

In the next section,
we'll dig deeper into this challenge and explore a decentralized approach to building robust database clusters.
