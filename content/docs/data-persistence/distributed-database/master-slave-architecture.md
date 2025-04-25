---
title: Master-slave Architecture
weight: 10
---

Many systems experience heavy read workloads, where read operations significantly outnumber write operations.
To meet this demand efficiently,
we can design a database architecture with a single writer (**Primary**) and multiple readers (**Replicas**).

- The writer propagates changes to the replicas.
- The replicas can serve read requests independently, offloading the primary and improving read scalability.

```d2
direction: right
w: Client (Write) {
    class: client
}
dc: Database cluster {
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

This setup is commonly known as the {{< term maSl >}} architecture (good name 🧐).

## Multi-master

Now, what happens if we allow **multiple writers**?
Would that significantly improve write performance?

```d2
dc: Database cluster {
    w1: Write server 1 {
      class: db
    }
    r1: Read server 1 {
      class: db 
    }
    r2: Read server 2 {
      class: db 
    }
    w2: Write sever 2 {
      class: db
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

Surprisingly, this paradigm **does not** significantly improve write throughput.
Every write operation must still be synchronized across all nodes in the cluster.
For example, when one master processes an update, it must replicate that change to the other masters.
This contrasts with read replicas, where each read request can be independently handled by a single replica.

```d2
sv: Client {
    class: client
}
dc: Database cluster {
    w1: Write server 1 {
      class: db
    }
    w2: Write sever 2 {
      class: db
    }
    w1 -> w2: Update
}
sv -> dc.w1: Update
```

The key advantage of a **Multi-Master** setup lies in higher availability.
If one master fails, others can continue to process writes, avoiding downtime.

### SQL Usage

The most widely adopted form of the {{< term maSl >}} model is {{< sqld >}},
as a single writer makes it easier to maintain strong consistency for [ACID transactions](../sql-database/concurrency-control/).

Because of this, **Multi-Master** setups are rarely used in practice.
They don’t offer enough benefits to justify their complexity:

- If the masters collaborate to maintain {{< term acid >}}, they must compromise availability.
- If they asynchronously replicate, they risk violating strong consistency.

## Replica Promotion

Back to the {{< term maSl >}} model, the master handles all updates, becoming a **single point of failure** that can affect system availability.
To mitigate the impact of master failure, we can introduce a [Standby Server](../#standby-server)
that is synchronously replicated from the master.
In the event of a failure, we can quickly **promote** the standby to become the new master.

```d2
grid-columns: 2
grid-gap: 100
d1: Database (normal) {
    grid-columns: 2
    horizontal-gap: 300
    m: Master {
      class: db
    }
    s: Standby {
      class: db
    }
    m -> s: Replicate synchronously {
      style.animated: true
    }
}
d2: Database (failure) {
    grid-columns: 2
    horizontal-gap: 300
    m: Master (crashes) {
      class: generic-error
    }
    s: Standby {
      class: db
    }
    s -> m: Promoted
}
```

## Centralized Cluster

The {{< term maSl >}} model is often deployed as a centralized cluster,  
with a **Coordinator** that acts as the cluster's entry point.

Since each server has a predefined role (master or replica), the **Coordinator** can:

- Route write requests to the master.
- Distribute (aka load balancing) read requests across replicas.

```d2
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

### Connection Pooling

Opening a new database connection is both slow and resource-intensive.
If each user request triggers a new connection, it leads to performance issues.

**Connection Pooling** is a common design pattern that caches and **reuses** existing connections.
A **Pool Manager** component is responsible for managing and sharing these connections.

This functionality is often integrated into the **Coordinator** to improve performance.

```d2
direction: right
s: Client 1 {
    class: client
}
s: Client  2 {
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
d: Database {
  class: db
}
s1 -> p.c1
s2 -> p.c2
p.c1 <-> d
p.c2 <-> d
```

## Problems

The {{< term maSl >}} model is simple and intuitive.
Each component has a well-defined role,
and the direct communication between nodes results in **low latency** and **fast responses**.

However, this simplicity conceals several **critical issues**,
most of which stem from the centralized control of the master server:

- The master becomes the {{< term spof >}}.
Its failure halts all write operations,
therefore, the {{< term maSl >}} model does not guarantee {{< term ha >}}.

- The master handles all write operations,
quickly becoming a **performance bottleneck**, especially in write-heavy applications.

Even if we can promote or recover the master extremely quickly,
there’s always a **downtime**, however brief.
And that downtime is **unpredictable**.
As a result, the database's availability affects its services, which cascades and impacts other services
([Aggregate Availability](../../web-service/service-cluster.md#aggregate-availability)).

In the next section,
we'll dig deeper into this challenge and explore a decentralized approach to building robust database clusters.
