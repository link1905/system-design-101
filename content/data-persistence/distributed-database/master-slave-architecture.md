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

The most widely adopted form of the {{< term maSl >}} model is {{< term sql >}} databases,
as a single writer makes it easier to maintain strong consistency for [ACID transactions]({{< ref "concurrency-control#acid" >}}).

Because of this, **Multi-Master** setups are rarely used in this case:

- They can not asynchronously replicate as risking violating {{< term acid >}} principles.
- In the other hands, if the masters continuously collaborate to maintain {{< term acid >}},
they must compromise availability and actions spanning on many nodes will be extremely complex.

## Standby Promotion

Back to the {{< term maSl >}} model, the master handles all updates, becoming a {{< term spof >}} that can affect system availability.
To mitigate the impact of master failure, we can introduce a [Standby Server]({{< ref "distributed-database#standby-server" >}})
that is synchronously replicated from the master.
In the event of a failure, we can quickly **promote** the standby to become the new master.

## Centralized Cluster

The {{< term maSl >}} model is often deployed with a centralized registry,
typically a [KV store]({{< ref "nosql-database#key-value-store" >}}),
holding and intermediating members information within the cluster.

```d2
direction: right
db: Database cluster {
  s1: Master {
    class: server
  }
  s2: Replica 1 {
    class: server
  }
  s3: Replica 2 {
    class: server
  }
  r: Registry {
    class: server
  }
  s1 <-> r
  s2 <-> r
  s3 <-> r
}
```

The master can be chosen by several ways:

- Manually affiliated by administrators.
- Or voting process:
members can communicate through the store to obtain agreements.

Selecting the one has most up-to-date data is a common strategy.
For example:

- When the **Master** node becomes unresponsive.
- Other replicas promote itself to the registry to become the new master.
- `Replica 2` then becomes the new master as it holds newer data then `Replica 1`.

```d2
direction: right
db: Database cluster {
  s1: Master {
    class: error
  }
  s2: Replica 1 {
    class: server
  }
  s3: Replica 2 {
    class: server
  }
  r: Registry {
    class: server
  }
  s1 <-> r {
    class: error-conn
  }
  s2 -> r: "Last record at 00:10"
  s3 -> r: "Last record at 00:20"
}

db-pro: Database cluster {
  s2: New master (from Replica 1) {
    class: server
  }
  s3: Replica 2 {
    class: server
  }
  r: Registry {
    class: server
  }
  s2 <-> r
  s3 <-> r
}
db -> db-pro
```

### Reverse Proxy

Letting clients contact with all of servers does not make scene,
instead,
we should build a [reverse proxy]({{< ref "load-balancer#reverse-proxy-pattern" >}}) before them.
Since each server has a predefined role (master or replica), the proxy can:

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
  c: Proxy {
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
