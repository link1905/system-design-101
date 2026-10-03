---
title: Stateless Service
---

A stateless service consists of instances that share **identical logic** and
do not retain instance-specific local state.

For example, consider two instances of the `Account Service`.
Both instances query the same database and are designed to produce identical results.
The specific instance a client connects to does not matter because every instance
uses the same logic and follows the same data-access pattern,
ensuring a **consistent response**.

```d2
direction: right
c: Client {
    class: client
}
s: Account Service {
  s1: Instance 1 {
    class: server
  }
  s2: Instance 2 {
    class: server
  }
}
db: Database {
  class: db
}
c <- s.s1: Get account data
c <- s.s2: Get account data
s.s1 <- db: Query
s.s2 <- db: Query
```

This consistency makes a stateless service relatively straightforward to scale:
we simply **increase or decrease the number of identical instances**.