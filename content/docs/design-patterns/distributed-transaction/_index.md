---
title: Distributed Transaction
weight: 20
prev: event-driven-architecture
next: blocking-protocols
---

{{< callout type="info" >}}
It may be beneficial to review the [Concurrency Control topic]({{< ref "concurrency-control" >}})
concerning transactions and consistency before proceeding with this section.
{{< /callout >}}

A distributed transaction is a single,
logical transaction that spans multiple physical servers or nodes.

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

The physical separation of these nodes introduces challenges in ensuring that distributed transactions fully adhere to the
[ACID]({{< ref "concurrency-control#acid" >}}) properties (Atomicity, Consistency, Isolation, Durability).
This adherence necessitates tight collaboration and coordination between the participating servers.

In some scenarios, a deliberate decision might be made to relax strict **ACID** compliance to gain other benefits,
such as higher availability or lower coupling between services.

Regardless of the specific implementation details,
any distributed transaction algorithm must guarantee that changes across all involved nodes are either **committed together**
or **aborted together** (all or nothing).

Essentially, there are two primary approaches to consistency in distributed transactions:

- **Strong Consistency**: Systems aiming for strong consistency require that a transaction is **atomically committed** across all relevant nodes.
This means all parts of the transaction either succeed or fail as a single, indivisible unit.

    This model typically employs strict algorithms, often involving [locking]({{< ref "concurrency-control#locking-mechanism" >}}) mechanisms,
    to ensure the **Isolation** property, preventing transactions from interfering with each other.

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

- **Eventual Consistency**: In this model, transactions are often decomposed into phases that may be committed at **different times** across various nodes.

    Due to this asynchronous behavior, the **Isolation** property is typically not implemented in the same strict sense as in strong consistency models.
    Instead, the system must be designed to **implicitly avoid or resolve inconsistencies** over time,
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
p -> n.n1: Commit
p -> n.n2: Commit
p -> n.n3: Commit
```
