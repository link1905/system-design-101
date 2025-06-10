---
title: Distributed Transaction
weight: 20
---

{{< callout type="info" >}}
You may review the [Concurrency Control topic](Concurrency-Control.md) about transaction and consistency
before browsing through this one.
{{< /callout >}}

In a distributed environment,
transactions can be truly complicated with the participation of multiple servers (nodes).
Because of physical segregation, it's challenging to ensure distributed transactions to follow
what we talked in the [ACID](Concurrency-Control.md) topic.
Let's compare them a bit!

## Characteristics

### Atomicity

The shortage of **Atomicity** probably lead to permanent inconsistencies.
So, at least, a distributed transaction needs to guarantee **Atomicity**,
requiring it to be written in all nodes (commit) or failed entirely (rollback).

### Isolation

**Isolation** primarily decides the consistency and accuracy of data.
In a single node, it's more straightforward to control this property,
we can locally set isolation levels or use a lock strategy.
However, in a distributed system, achieving **Isolation** is extremely challenging.
Different nodes may implement their own concurrency control,
requiring a consensus mechanism to prevent serialization anomalies between them.

In essence, we have two approaches to implement a distributed transaction:

- **Strong Consistency**: In this system, we need a transaction to be immediately completed (committed) in relevant nodes.
To do that, a strict algorithm is required to ensure the **Isolation**,
it's frequently about [locking](Concurrency-Control.md#locking-mechanism) data.

```d2
grid-columns: 1
p: Transaction {
    class: process
}
c: Commit {
    width: 30
    height: 30
    shape: circle
}
n: "" {
    grid-rows: 1
    n1: Node A {
        class: db
    }
    n2: Node B {
        class: db
    }
    n3: Node C {
        class: db
    }
}
p -> c
c -> n.n1
c -> n.n2
c -> n.n3
```

- **Eventual Consistency**: Transactions are separated into phases committed at different moments.
In this model, we **don't** even implement **Isolation** because of the asynchronous behavior,
the system must be designed somehow to implicitly avoid inconsistencies.

```d2
grid-columns: 1
p: Transaction {
    class: process
}
n: "" {
    grid-rows: 1
    n1: Node A {
        class: db
    }
    n2: Node B {
        class: db
    }
    n3: Node C {
        class: db
    }
}
p -> n.n1: Commit
p -> n.n2: Commit
p -> n.n3: Commit
```
