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
- The replicas can serve read requests independently, reducing the load on the primary and improving read scalability.

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

This setup is commonly known as the {{< term maSl >}} **architecture** (good name 🧐).

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

The {{< term maSl >}} model is most widely adopted in {{< term sql >}} databases,
as a single writer makes it easier to maintain strong consistency for [ACID transactions]({{< ref "concurrency-control#acid" >}}).

Because of this, **multi-master** setups are rarely used in this context:

- They cannot replicate asynchronously without risking violations of {{< term acid >}} principles.
- On the other hand, if the masters continuously coordinate to maintain {{< term acid >}},
they must compromise availability, and operations spanning multiple nodes become extremely complex.

## Centralized Cluster

The {{< term maSl >}} model is often deployed with a centralized registry,
that stores and facilitates the exchange of cluster membership information.

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
    class: db
  }
  s1 <-> r
  s2 <-> r
  s3 <-> r
}
```

The master can be selected in several ways:

- Through manual assignment by administrators.
- Through a voting process:
members can communicate through the store to reach agreement.

Selecting the node with the most up-to-date data is a common strategy.
For example:

- The **master** node becomes unresponsive.
- The replicas put themselves forward through the registry to become the new master.
- `Replica 2` then becomes the new master because it holds newer data than `Replica 1`.

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
    class: db
  }
  s1 <-> r: Down {
    class: error-conn
  }
  s2 -> r: "Last record at 00:10"
  s3 -> r: "Last record at 00:20"
}

db-pro: Database cluster {
  s2: "New master (from Replica 2)" {
    class: server
  }
  s3: Replica 1 {
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

Having clients connect directly to all servers does not make sense.
Instead,
we should place a [reverse proxy]({{< ref "load-balancer#reverse-proxy-pattern" >}}) in front of them.
Since each server has a predefined role (master or replica), the proxy can:

- Route write requests to the master.
- Distribute read requests across replicas to balance the load.

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
and the direct communication between nodes results in low latency.

However, this simplicity conceals several issues,
most of which stem from the centralized control of the master server:

- The master becomes the {{< term spof >}}.
Its failure halts all write operations;
therefore, the {{< term maSl >}} model does not guarantee {{< term ha >}}.

- The master quickly becomes a performance bottleneck, especially in write-heavy applications.

In the next section,
we'll examine this challenge in more detail and explore a decentralized approach to building robust database clusters.
