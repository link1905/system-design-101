---
title: Physical Layer
weight: 10
prev: sql-database
---

The way data is organized at the physical layer shapes the entire database workflow.

## Page

A **page** is a fixed-size container
(typically a few kilobytes; for example, {{< term postgres >}} uses 8kB pages by default)
that holds multiple data entries, which may be either rows or index records.

A page generally consists of three parts: the **Header**,
**Tuples**, and a **Line Pointer Array (LPA)**.

### Header

The header primarily stores metadata used for
data integrity, recovery, and other management purposes.
For now, we don't need to worry about its details.

```d2
Page {
    grid-gap: 0
    grid-columns: 1
    Header (Metadata)
}
```

### Tuple

The actual data resides in the **Tuples** section as *sequentially packed records*.

We call them *tuples* rather than *records* because tuples are *immutable*.
To update a tuple, the database creates a new version that effectively supersedes the previous one.

For example, the third tuple is an updated version of the first.

```d2
Page {
    grid-gap: 0
    grid-columns: 1
    Header (Metadata)
    t: Tuples (Data) {
        grid-gap: 0
        grid-columns: 1
        t1: (ID=1, Name="John") {
            class: partition
        }
        t2: (ID=3, Name="Charlie") {
            class: partition
        }
        t3: (ID=1, Name="John Doe") {
            class: partition
        }
    }
}
```

There are two main reasons for this immutability:

1. A new tuple may be larger or smaller than the existing one
   (for example, because of a resizable **TEXT** field).
   If we updated it in place, subsequent tuples might need to shift to accommodate the size change,
   similar to inserting or deleting an element in an array, which can be costly.

    ```d2
    grid-gap: 0
    grid-columns: 1
    Header (Metadata)
    t: Tuples {
    grid-rows: 1
    horizontal-gap: 100
    t: Old Tuples {
        grid-gap: 0
        grid-columns: 1
        t1: '(ID=1, Name="John")' {
            height: 50
        }
        t2: '(ID=3, Name="Charlie")' {
            height: 50
        }
        "..." {
            height: 50
            class: none
        }
    }
    ut: Updated Tuples {
        grid-gap: 0
        grid-columns: 1
        t1: '(ID=1, Name="John Doe")' {
        height: 100
        }
        t2: '(ID=3, Name="Charlie")' {
        height: 50
        }
    }
    t.t2 -> ut.t2: Moved
    t.t1 -> ut.t1: Increased
    }
    ```

2. It supports [Multi-version Concurrency Control (MVCC)](https://www.postgresql.org/docs/current/mvcc-intro.html),
   which allows transactions to access different versions of the same data concurrently.
   We'll explore this in more detail in the [Concurrency Control]({{< ref "concurrency-control#transaction" >}}) topic.

### Line Pointer Array (LPA)

Between the header and the tuple data lies the **Line Pointer Array (LPA)**,
which contains fixed-size pointers to the actual tuples on the page.
Each pointer can also track the state of its tuple, such as *Updated* or *Deleted*.

```d2
Page {
    grid-gap: 0
    grid-columns: 1
    Header (Metadata)
    l: Line Pointer Array {
        grid-gap: 0
        grid-rows: 1
        p1: "(1, Updated)" {
            class: partition
        }
        p2: "(2, Normal)" {
            class: partition
        }
        p3: "(3, Deleted)" {
            class: partition
        }
    }
    t: Tuples (Data) {
        grid-gap: 0
        grid-rows: 1
        t1: (ID=1, Name="John") {
            class: partition
        }
        t2: (ID=3, Name="Charlie") {
            class: partition
        }
        t3: (ID=1, Name="John Doe") {
            class: partition
        }
    }
    l.p1 -> t.t1
    l.p2 -> t.t2
    l.p3 -> t.t3
}
```

Because tuple sizes can vary, primarily due to fields such as text,
using fixed-size pointers in the **Line Pointer Array (LPA)** enables fast,
random access to any specific tuple.

Retrieving a pointer in this way is similar to accessing an element in an array:

```md
pointer[i] = page_address + pointer_size * (i - 1)
```

The primary purpose of the **LPA** is to keep tuple references stable.
Even if a row is relocated among the tuples within a page,
its corresponding line pointer remains unchanged, ensuring consistent references.

## Heap

A **heap** is simply a collection of pages representing a table.
The term **heap** essentially means *no particular structure*.

When inserting a record into a heap,
the database looks for an available slot in an available page.

For example:

```d2
grid-rows: 2
h: Heap {
    grid-gap: 50
    grid-columns: 2
    Page 1 {
        grid-gap: 0
        grid-columns: 1
        Header (Metadata)
        l: Line Pointer Array {
            grid-gap: 0
            grid-rows: 1
            p1: "(1, Normal)" {
                width: 300
            }
            p2: "(2, Normal)" {
                width: 300
            }
        }
        t: Tuples (Data) {
            grid-gap: 0
            grid-rows: 1
            t1: (ID=3, Name="John") {
                width: 300
            }
            t2: (ID=1, Name="Charlie") {
                width: 300
            }
        }
        l.p1 -> t.t1
        l.p2 -> t.t2
    }
    p2: Page 2 {
        grid-gap: 0
        grid-columns: 1
        Header (Metadata)
        l: Line Pointer Array {
            grid-gap: 0
            grid-rows: 1
            p1: "(1, Normal)" {
                width: 300
            }
            p2: " " {
                width: 300
            }
        }
        t: Tuples (Data) {
            grid-gap: 0
            grid-rows: 1
            t1: (ID=2, Name="Mary") {
                width: 300
            }
            t2: " " {
                width: 300
                style.fill: ${colors.i2}
            }
        }
        l.p1 -> t.t1
    }
}
i: "INSERT INTO table VALUES (4, 'Tom')"
i -> h.p2.t.t2
```

### Tuple ID (TID)

External elements, such as other tables or indexes, reference tuples using a **Tuple ID (TID)**.
A TID is a composite identifier consisting of **(page number, LPA index)**.

Because pages have a fixed size, they support efficient random access.
For example, to retrieve the tuple at `TID (page = 2, index = 1)`,
the database can jump directly to it:

`Heap.Pages[2].LPA[1] -> Tuple`.

```d2
grid-rows: 2
h: Heap {
    grid-gap: 50
    grid-columns: 2
    Page 1 {
        grid-gap: 0
        grid-columns: 1
        Header (Metadata)
        l: Line Pointer Array {
            grid-gap: 0
            grid-rows: 1
            p1: "(1, Normal)" {
                width: 300
            }
            p2: "(2, Normal)" {
                width: 300
            }
        }
        t: Tuples (Data) {
            grid-gap: 0
            grid-rows: 1
            t1: (ID=3, Name="John") {
                width: 300
            }
            t2: (ID=1, Name="Charlie") {
                width: 300
            }
        }
        l.p1 -> t.t1
        l.p2 -> t.t2
    }
    p2: Page 2 {
        grid-gap: 0
        grid-columns: 1
        Header (Metadata)
        l: Line Pointer Array {
            grid-gap: 0
            grid-rows: 1
            p1: "(1, Normal)" {
                width: 300
            }
            p2: "(2, Normal)" {
                width: 300
                style.fill: ${colors.i2}
            }
        }
        t: Tuples (Data) {
            grid-gap: 0
            grid-rows: 1
            t1: (ID=2, Name="Mary") {
                width: 300
            }
            t2: (ID=4, Name="Tom") {
                width: 300
            }
        }
        l.p1 -> t.t1
        l.p2 -> t.t2
    }
}
i: "TID (page = 2, index = 1)"
i -> h.p2.l.p2
```

### Heap Search

However, a heap isn't particularly useful for searching.
There are two practical ways to *access a tuple in a heap*:

1. Using its **TID**.
2. Scanning the entire heap, also known as a full table scan. 🥲

## Index

Unstructured heaps aren't ideal for fast tuple lookups. This is where an **Index** comes to the rescue.

In short, an index is an auxiliary data structure built alongside a table to enable efficient data retrieval.

Unlike tables, indexes are organized using specific data structures designed for fast lookups.
In this section, we'll focus on the most commonly used type: the **B-Tree Index**.

### B-Tree Index

A [B-Tree](https://www.geeksforgeeks.org/introduction-of-b-tree-2/) is an advanced form of a **Binary Search Tree**.
Instead of storing a single key per node like a binary tree, it stores multiple keys in each node,
increasing the [branching factor](https://en.wikipedia.org/wiki/Branching_factor) and reducing the overall height of the tree.

For example, a B-Tree might store up to five elements per node.
Each node can contain either actual values or pointers to child nodes.

```d2
grid-gap: 100
grid-rows: 2
e0: {
    width: 300
    class: none
}
e1: "" {
    grid-gap: 0
    grid-rows: 1
    e0: "" {
        width: 10
    }
    e1: "35" {
        width: 100
    }
    e2: "" {
        width: 10
    }
    e3: "60" {
        width: 100
    }
    e4: "" {
        width: 10
    }
}
e2: {
    width: 300
    class: none
}
e3: "" {
    grid-gap: 0
    grid-rows: 1
    width: 230
    e0: "10" {
        width: 115
    }
    e2: "20" {
        width: 115
    }
}
e4: "" {
    grid-gap: 0
    grid-rows: 1
    e0: "40" {
        width: 115
    }
    e1: "50" {
        width: 115
    }
}
e5: "" {
    grid-gap: 0
    grid-rows: 1
    e0: "88" {
        width: 110
    }
    e1: "90" {
        width: 110
    }
    e2: "100" {
        width: 110
    }
}
e1.e0 -> e3
e1.e2 -> e4
e1.e4 -> e5
```

In {{< term sql >}}, an **index page** corresponds to a B-Tree node.
Each page can point either to **child index pages** or to actual table records through their **TIDs**.

A lookup traverses this structure starting from the root page,
moving down through child pages until it locates the desired tuple.

```d2
i: Index {
    vertical-gap: 100
    grid-rows: 2
    r: Page 1 (Root) {
        grid-gap: 0
        grid-rows: 1
        t1: Id = 1, (Page = 1, Page Offset = 2)
        t2: Page 2 {
            style.fill: ${colors.i2}
        }
        t3: Id = 3, (Page = 1, Page Offset = 1)
        t4: Page 3 {
            style.fill: ${colors.i2}
        }
    }
    p2: Page 2 {
        grid-gap: 0
        t1: Id = 2, (Page = 2, Page Offset = 1) {
         width: 250
        }
    }
    p3: Page 3 {
        grid-gap: 0
        t1: Id = 4, (Page = 3, Page Offset = 1) {
         width: 250
        }
    }
    r.t2 -> p2
    r.t4 -> p3
}
h: Heap {
    grid-gap: 0
    grid-columns: 1
    p1: Page 1 {
        grid-gap: 0
        grid-columns: 1
        t1: (ID=3, Name="John") {
            width: 300
        }
        t2: (ID=1, Name="Charlie") {
            width: 300
        }

    }
    p2: Page 2 {
        grid-gap: 0
        grid-columns: 1
        t1: (ID=2, Name="John") {
            width: 300
        }
    }
    p3: Page 3 {
        grid-gap: 0
        grid-columns: 1
        t1: (ID=4, Name="Tom") {
            width: 300
        }
    }
}
i.r.t1 -> h.p1.t2
i.r.t3 -> h.p1.t1
i.p2.t1 -> h.p2.t1
i.p3.t1 -> h.p3.t1
```

Notably, indexes introduce a trade-off between read and write performance.

To maintain a B-Tree's **self-balancing** structure, operations such as inserts, deletes, and updates may trigger node splits or merges.
This maintenance overhead enables efficient reads but comes at the cost of slower writes, especially for tables with multiple indexes,
where each change may require corresponding updates to several index structures.

### HOT Update

Because tuples are immutable, an update creates a new tuple.
However, we don't necessarily need to update every associated index for every change.

An update can be treated as a **HOT (Heap-Only Tuple)** update and avoid index manipulation if:

- The update doesn't modify [indexed columns]({{< ref "query-optimization#index-only-scan" >}}).
- The new tuple remains on the same page as the original tuple.

#### Tuple Chaining

When multiple versions of a tuple reside on the same page, they can be *chained together*.
Each tuple carries **metadata**, such as its transaction state and a pointer to the next version.
This chain allows queries to resolve the correct visible version of a tuple.

For example, an obsolete tuple may contain a pointer to the next valid tuple,
allowing the database to traverse the chain and retrieve the latest applicable version.
As a result, external index entries that hold the original **TID** don't need to be updated.

```d2
i: Index {
   i1: TID (index = 1)
}
h: Heap {
   horizontal-gap: 0
   vertical-gap: 50
   grid-columns: 1
   l: LPA {
      grid-gap: 0
      grid-columns: 1
      t1: Pointer (index = 1) {
         width: 330
      }
   }
   t: Tuples {
      grid-gap: 0
      grid-columns: 1
      vertical-gap: 30
      t1: (State = DELETED, Id = 1, Name = Jnho)
      t2: (State = DELETED, Id = 1, Name = John)
      t3: (State = NORMAL, Id = 1, Name = Johnny) {
         style.fill: ${colors.i2}
      }
      t1 -> t2
      t2 -> t3
   }
   l.t1 -> t.t1
}
i.i1 -> h.l.t1
```

#### Vacuum Process

As old tuple versions become obsolete, meaning they are no longer visible to any active transaction, their space can be reclaimed.
The remaining tuples may also be rearranged during this process.
This is called the **Vacuum Process**.

To perform vacuuming, the database needs to track two things:
*active transactions* and *the oldest transaction associated with each tuple*.

For example, suppose the database determines that the first two tuples are no longer associated with any active transaction.
Their space can then be reclaimed, and the line pointer can be updated to reference the latest valid tuple.

```d2
grid-columns: 1
i: "Current transactions: T3,T4"
h: Heap {
    grid-rows: 1
    grid-gap: 100
    h1: Heap (version 1) {
        grid-gap: 0
        grid-columns: 1
        l: LPA {
            grid-gap: 0
           grid-columns: 1
            t1: Pointer (Offset = 1) {
               width: 480
            }
        }
        t: Tuples {
           grid-gap: 0
           grid-columns: 1
           t1: (MinTransaction = T1, State = DELETED, Id = 1, Name = Jnho) {
            style.fill: ${colors.i2}
           }
           t2: (MinTransaction = T2, State = DELETED, Id = 1, Name = John)
           t3: (MinTransaction = T3, State = NORMAL, Id = 1, Name = Johnny)
        }
        l.t1 -> t.t1
    }
    h2: Heap (version 2) {
        grid-gap: 0
        grid-columns: 1
        l: LPA {
            grid-gap: 0
            grid-columns: 1
            t1: Pointer (Offset = 1) {
               width: 480
            }
        }
        t: Tuples {
           grid-gap: 0
           grid-columns: 1
           t1: "...Removed"
           t2: "...Removed"
           t3: (MinTransaction = T3, State = NORMAL, Id = 1, Name = Johnny) {
            style.fill: ${colors.i2}
           }
        }
        l.t1 -> t.t3
    }
    h1 -> h2: Cleaned
}
```

Vacuuming is critical to database maintenance because long-running transactions can keep old tuple versions visible,
preventing their space from being reclaimed.

### Secondary vs Clustered Index

So far, we've assumed that records are stored in an unordered heap
and that indexes exist *separately from the table data*.
These are known as **Secondary Indexes**.

However, in some database engines, such as [MySQL InnoDB](https://dev.mysql.com/doc/refman/8.4/en/innodb-storage-engine.html),
table data is physically organized according to the primary key.
This structure is known as a **Clustered Index**.

Queries using the primary key can therefore access the table data directly,
without requiring a separate secondary index for those lookups.