---
title: Physical Layer
weight: 10
prev: sql-database
---

The way data is organized at the physical layer drives the entire database workflow.

## Page

A **page** is a fixed-size container
(typically a few kilobytes, e.g., {{< term postgres >}} defaults to 8kB)
holding multiple data entries, which can be either rows or index records.

A page generally consists of three parts: **Header**,
**Tuples**, and a **Line Pointer Array (LPA)**.

### Header

The header primarily holds metadata for
data integrity, recovery, and other management purposes.
For now, we don’t need to worry about its details.

```d2
Page {
    grid-gap: 0
    grid-columns: 1
    Header (Metadata)
}
```

### Tuple

The actual data resides in the **Tuples** section as **sequentially packed records**.

We call them *tuples* rather than *records* because tuples are **immutable**.
To update a tuple, the database creates a new version, essentially an overriding tuple.

For example, the third tuple is an updated version of the first one.

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

There are two main reasons for the immutability:

1. A new tuple might be larger or smaller than the existing one
   (for example, a resizable **VARCHAR** field).
   If we tried to update it in place, subsequent tuples would need to shift to accommodate size changes,
   much like deleting an element from an array, which can be costly.

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

2. To support [Multi-version Concurrency Control (MVCC)](https://www.postgresql.org/docs/current/mvcc-intro.html),
   where transactions can access different versions of the same data concurrently.
   We’ll cover this more deeply in the [Concurreny Control]({{< ref "concurrency-control#transaction" >}}) topic.

### Line Pointer Array (LPA)

Between the header and the tuple data lies the **Line Pointer Array (LPA)**,
which holds fixed-size pointers to the actual tuples on the page.
Each pointer also tracks the state of its tuple (e.g., **Updated**, **Deleted**).

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

Since tuple sizes can vary (primarily due to text fields),
utilizing fixed-size pointers within the **Line Pointer Array (LPA)** guarantees fast,
random access to any specific tuple.
Retrieving a pointer in this manner is analogous to accessing an element in an array:

```md
pointer[i] = page_address + pointer_size * (i - 1)
```

The primary function of the **LPA** is to maintain the stability of tuple references.
Even if a row is relocated among different tuples within the page,
its corresponding line pointer remains constant, ensuring consistent referencing.

## Heap

A **heap** is simply a collection of pages representing a table.
The term **heap** means essentially **no structure**.

When inserting a record into a heap,
the database looks for the first available slot in the first available page.
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

External elements, such as other tables, reference data tuples using a **Tuple ID (TID)**.
A TID is a composite identifier consisting of: **(page number, LPA index)**.

**Fixed-size pages** make random access efficient,
for example,
to retrieve the tuple at: `TID (page = 2, index = 1)`;
The database can jump straight to it via: `Heap.Pages[2].LPA[1] -> Tuple`.

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

However, a heap isn’t particularly helpful for searching.
There are two effective ways to access a tuple in a heap:

1. Using its **TID**.
2. Scanning the entire heap (i.e., a full table scan) 🥲.

## Index

Messy heaps aren’t ideal for fast tuple lookups, this is where the **Index** comes to the rescue.
In short, an index is an auxiliary data structure built alongside a table to enable efficient data retrieval.

Unlike tables, indexes are organized using specific data structures tailored for quick lookups.
In this section, we'll focus on the most commonly used type: the **B-Tree Index**.

### B-Tree Index

A [B-Tree](https://www.geeksforgeeks.org/introduction-of-b-tree-2/) is an advanced form of a **Binary Search Tree**.
Instead of storing one key per node like a binary tree, it groups multiple keys in a single node, increasing the
[branching factor](https://en.wikipedia.org/wiki/Branching_factor) and reducing the overall height of the tree.

For example, a B-Tree might store up to 5 elements per node.
Each node either contains actual values or pointers to its child nodes.

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
Each page can point either to **child index pages** or to actual table records via their **TIDs**.
A lookup operation traverses this structure, starting from the root page,
it moves down through child pages until it locates the desired tuplee.

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

Notably, using indexes introduces a trade-off between read and write performance.
To maintain a B-Tree’s **self-balancing** structure, operations such as inserts, deletes, or updates may trigger node splits or merges.
This overhead ensures efficient reads but comes at the cost of slower writes, especially in tables with multiple indexes,
where every change must be reflected across all relevant index structures.

### HOT Update

Given the immutability of tuples, when an update occurs, a new tuple is created.
However, we don't need to necessarily update all associated indices for every update.

An update can be treated as a **HOT (Heap-Only Tuple)** and does not require index manipulation if:

- The update doesn’t modify [indexed columns]({{< ref "query-optimization#index-only-scan" >}}).
- And the new tuple remains on the same page as the original tuple.

#### Tuple Chaining

When multiple versions of a tuple reside on the same page, they are **chained together**.
Each tuple carries **metadata**, such as its transaction state and a pointer to the next version.
This chain ensures queries can resolve to the correct visible version of a tuple.

For example, deleted tuples may include a pointer to the next valid tuple,
allowing the system to traverse and retrieve the most up-to-date version.

```d2
i: Index {
   i1: TID (index = 1)
}
h: Heap {
   horizontal-gap: 0
   vertical-gap: 50
   grid-columns: 1
   class: none
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
i -> h.l.t1
```

As old tuple versions become obsolete (no longer visible to any active transaction), their space is reclaimed.
The line pointer in the page is then updated to reference the latest valid tuple.
This cleanup happens without requiring changes to the index itself, keeping lookup operations consistent.

```d2
grid-columns: 1
i: Index {
    i1: (Offset = 1) {
        width: 1300
    }
}
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
               width: 330
            }
        }
        t: Tuples {
           grid-gap: 0
           grid-columns: 1
           t1: (State = DELETED, Id = 1, Name = Jnho) {
            style.fill: ${colors.i2}
           }
           t2: (State = DELETED, Id = 1, Name = John)
           t3: (State = NORMAL, Id = 1, Name = Johnny)
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
               width: 330
            }
        }
        t: Tuples {
           grid-gap: 0
           grid-columns: 1
           t1: "...Removed"
           t2: (State = DELETED, Id = 1, Name = John) {
            style.fill: ${colors.i2}
           }
           t3: (State = NORMAL, Id = 1, Name = Johnny)
        }
        l.t1 -> t.t2
    }
    h3: Heap (version 3) {
        grid-gap: 0
        grid-columns: 1
        l: LPA {
            grid-gap: 0
            grid-columns: 1
            t1: Pointer (Offset = 1) {
               width: 330
            }
        }
        t: Tuples {
           grid-gap: 0
           grid-columns: 1
           t1: "...Removed"
           t2: "...Removed"
           t3: (State = NORMAL, Id = 1, Name = Johnny) {
            style.fill: ${colors.i2}
           }
        }
        l.t1 -> t.t3
    }
    h1 -> h2: Cleaned
    h2 -> h3: Cleaned
}
i.i1 -> h.h1.l.t1
i.i1 -> h.h2.l.t1
i.i1 -> h.h3.l.t1
```

### Secondary vs Clustered Index

In this explanation, we've assumed records are stored in a disorganized heap
and that indexes exist **outside the table data**. These are known as **Secondary Indexes**.

However, in some database engines (like [MySQL InnoDB](https://dev.mysql.com/doc/refman/8.4/en/innodb-storage-engine.html)),
tables are physically organized based on their primary key, also called **Clustered Index**.
Queries using the primary key access the table directly,
no secondary index needed for these lookups.
