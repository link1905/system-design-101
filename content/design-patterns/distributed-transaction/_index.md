---
title: Distributed Transaction
weight: 20
prev: event-driven-architecture
next: blocking-protocols
---

{{< callout type="info" >}}
Consider reviewing the [Concurrency Control topic]({{< ref "concurrency-control" >}})
for background on transactions and consistency before proceeding with this section.
{{< /callout >}}

A distributed transaction is a single logical transaction that spans multiple physical servers or nodes.

```d2
t: Transaction {
    class: process
}
s {
    class: none
    s1: Server 1 {
        class: server
    }
    s2: Server 2 {
        class: server
    }
    s3: Server 3 {
        class: server
    }
}
t -> s.s1
t -> s.s2
t -> s.s3
```

The physical separation of these nodes makes it challenging to ensure that distributed transactions fully satisfy the
[ACID]({{< ref "concurrency-control#acid" >}}) properties (atomicity, consistency, isolation, and durability).
Satisfying these properties requires close coordination among the participating servers.

In some scenarios, systems deliberately relax strict **ACID** compliance to gain other benefits,
such as higher availability or looser coupling between services.

There are two primary approaches to consistency in distributed transactions:

- **Strong Consistency**: Systems aiming for strong consistency require that a transaction is **atomically committed** across all relevant nodes.
    This means that all parts of the transaction either succeed or fail as a single, indivisible unit.

    This model typically uses strict algorithms, often involving [locking]({{< ref "concurrency-control#locking-mechanism" >}}) mechanisms,
    to ensure the **isolation** property and prevent transactions from interfering with one another.

```d2
grid-columns: 1
p: Transaction {
    class: process
}
c: Commit {
    height: 30
}
n: "" {
    grid-rows: 1
    n1: Node A {
        class: server
    }
    n2: Node B {
        class: server
    }
    n3: Node C {
        class: server
    }
}
p -> c
c -> n.n1
c -> n.n2
c -> n.n3
```

- **Eventual Consistency**: In this model, transactions are often divided into phases that may be committed at **different times** across nodes.

    Because of this asynchronous behavior, the system must be designed to **implicitly avoid or resolve inconsistencies** over time,
    eventually reaching a consistent state across all nodes.

```d2
grid-columns: 1
p: Transaction {
    class: process
}
n: "" {
    grid-rows: 1
    n1: Node A {
        class: server
    }
    n2: Node B {
        class: server
    }
    n3: Node C {
        class: server
    }
}
p -> n.n1: 1. Commit
p -> n.n2: 2. Commit
p -> n.n3: 3. Commit
```
