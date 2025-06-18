---
title: Query Optimization
weight: 20
params:
  math: true
---

To maintain data deduplication,
{{< term sql >}} normalizes data across multiple tables
and then joins them in queries afterward.
While this helps maintain data integrity, it often makes queries less efficient,
since they must interact with multiple tables,
which might reside in different files (or even on different servers).

In this topic, we’ll cover common techniques to improve query performance.
At its core, the guiding principle is simple: **minimize I/O operations as much as possible**.

## I/O Operation

**I/O (Input/Output)** refers to the process of transferring data between disk (storage) and memory (RAM).
There are two primary types of `I/O`:

- **Read I/O**: Moves data from disk to memory for access.
- **Write I/O**: Persists changes from memory back to disk.

Since interacting with disk storage is relatively slow,
an efficient database system must minimize I/Os and leverage memory caching whenever possible.

### Memory Layer

The memory structure closely mirrors the organization of data on disk.
However, rather than caching entire tables or indexes, only the **necessary pages** are cached in memory.

A common caching strategy is **LRU (Least Recently Used)**, which works like this:

- **Cache Miss**: If a required page isn’t already in memory, it’s loaded from disk.
- **Eviction**: If memory is full, the least recently accessed pages are evicted to make room for new ones.

## Indexing

In runtime, how an index is utilized depends heavily on the query context.
There are typically three common query patterns:

### Index Scan

This is the standard way to use an index.
It involves at least **two I/O operations** per record retrieval:

1. One to locate the index entry.
2. Another to fetch the actual tuple from the heap.

For example, to find a student with `Id = 3`:

- First, locate the index entry for `Id = 3`.
- Then, follow its pointer to retrieve the actual record..

```d2
grid-rows: 2
query: SELECT Name WHERE Id = 3 {
  shape: text
  near: top-center
  style: {
    font-size: 30
    bold: true
  }
}
d: {
    class: none
    i: Index {
        grid-gap: 0
        grid-columns: 1
        i1: (Id = 1)
        i2: (Id = 3)  {
            style.fill: ${colors.i2}
        }
    }
    h: Heap {
        grid-gap: 0
        grid-columns: 1
        p1: Page 1 {
            grid-gap: 0
            grid-columns: 1
            t1: (Id = 1, Name = John, GPA = 9) {
              width: 300
            }
        }
        p2: Page 2 {
            grid-gap: 0
            grid-columns: 1
            t1: (Id = 3, Name = Yuki, GPA = 8) {
              style.fill: ${colors.i2}
              width: 300
            }
        }
    }
    i.i2 -> h.p2.t1
}
query -> d.i.i2
```

In large datasets, this back-and-forth access pattern can become inefficient.
When querying a large number of records,
repeatedly jumping between index pages and heap pages can significantly impact performance.

#### Index-Only Scan

To improve this, we can attach **additional columns** to an index,
allowing queries to retrieve these values directly from the index without accessing the heap.

This technique is known as a **Covering Index**.
It reduces I/O by avoiding heap lookups for queries that only need the indexed columns.

For example, by including the `Name` field in the student index,
queries that retrieve only the `Name` can be resolved directly from the index
without accessing the underlying table.

```d2
grid-rows: 2
query: SELECT Name WHERE Id = 3 {
  shape: text
  near: top-center
  style: {
    font-size: 30
    bold: true
  }
}
d: {
    class: none
    i: Index {
        grid-gap: 0
        grid-columns: 1
        i1: (Id = 1, Name = John)
        i2: (Id = 3, Name = Yuki)  {
            style.fill: ${colors.i2}
        }
    }
    h: Heap {
        grid-gap: 0
        grid-columns: 1
        p1: Page 1 {
            grid-gap: 0
            grid-columns: 1
            t1: (Id = 1, Name = John, GPA = 9) {
              width: 300
            }
        }
        p2: Page 2 {
            grid-gap: 0
            grid-columns: 1
            t1: (Id = 3, Name = Yuki, GPA = 8) {
                style.fill: ${colors.i2}
                width: 300
            }
        }
    }
}
query -> d.i.i2
```

However, note that adding extra columns makes the index larger, increasing the storage cost.
It also complicates updates, any change to an included column requires updating the index entry as well,
making optimizations like [HOT updates]({{< ref "physical-layer#hot-update" >}}) unfeasible.

### Table Scan

A **Table Scan** bypasses indexes entirely and reads the **entire set of table pages** sequentially.

This approach is chosen when the database estimates that the query will return a large proportion of the table’s rows.
In such cases, a full scan is more efficient than navigating between index and heap pages repeatedly.

For example,
if most students have `GPA > 6`, the system might opt for a table scan to retrieve those records.

```d2
query: SELECT Name WHERE GPA > 6 {
  shape: text
  near: top-left
  style: {
    font-size: 30
    bold: true
  }
}
h: Heap {
    grid-gap: 0
    grid-columns: 1
    p1: Page 1 {
        grid-gap: 0
        grid-columns: 1
        t1: (Id = 1, Name = John, GPA = 10) {
          style.fill: ${colors.i2}
          width: 300
        }
        t2: (Id = 5, Name = Caitlyn, GPA = 7) {
          style.fill: ${colors.i2}
          width: 300
        }
    }
    p2: Page 2 {
        grid-gap: 0
        grid-columns: 1
        t1: (Id = 3, Name = Yuki, GPA = 8) {
          style.fill: ${colors.i2}
          width: 300
        }
        t2: (Id = 4, Name = Link, GPA = 6) {
          width: 300
        }
    }
}
query -> h.p1: 1. Read first page
query -> h.p2: 2. Read second page
```

### Bitmap Scan

A **Bitmap Scan** offers a hybrid strategy, useful when:

- It’s inefficient to repeatedly jump between an index and the table.
- And a full table scan would be excessive.

Since indexes are typically smaller than table data, loading index pages first allows the system to gather more information with fewer I/Os.

In a **Bitmap Index Scan**:

1. The system scans the index to identify qualifying records and marks the **pages** (not tuples) in a bitmap (`page number -> boolean`).
2. It then sequentially reads the marked pages and filters them for the matching tuples.

For example:

- The index might mark `Page 1` and `Page 2` for filtering.
- The database then reads these pages to extract the qualifying recordss.

```d2
query: SELECT Name WHERE Id < 7 {
  shape: text
  near: top-left
  style: {
    font-size: 30
    bold: true
  }
}
i: Index {
    grid-gap: 0
    grid-columns: 1
    i2: (Id = 1) -> (Page = 1, Offset = 1) {
        width: 400
        style.fill: ${colors.i2}
    }
    i2: (Id = 3) -> (Page = 2, Offset = 2) {
        style.fill: ${colors.i2}
    }
    i3: (Id = 5) -> (Page = 1, Offset = 2) {
        style.fill: ${colors.i2}
    }
    i5: (Id = 6) -> (Page = 2, Offset = 1)  {
        style.fill: ${colors.i2}
    }
    i4: (Id = 7) -> (Page = 3, Offset = 1)
}
b: Bitmap (1: Present, 0: Absent) {
    grid-gap: 0
    grid-rows: 2
    1: "Page 1" {
        width: 200
        style.fill: ${colors.i2}
    }
    2: "Page 2" {
        width: 200
        style.fill: ${colors.i2}
    }
    3: "Page 3" {
        width: 200
    }
    1m: "1" {
        width: 200
    }
    2m: "1" {
        width: 200
    }
    3m: "0" {
        width: 200
    }
}
query -> i: Scan
i -> b
```

**Bitmap Scan** can also combine multiple index conditions using **bitwise operations** on their respective bitmaps.
For instance:

- One index bitmap identifies pages with `Id < 7`.
- Another bitmap identifies pages with `Grade > 7`.
- The system performs a bitwise `AND` operation on the bitmaps to find pages matching both conditions, reducing unnecessary I/Os.

```d2

grid-columns: 1
query: SELECT Name WHERE Id < 6 AND Grade > 7 {
  shape: text
  near: top-center
  style: {
    font-size: 30
    bold: true
  }
}
data: "" {
    grid-columns: 2
    i1: Id Index {
        grid-gap: 0
        grid-columns: 1
        i1: (Id = 1) -> (Page = 1, Offset = 1) {
            width: 600
            style.fill: ${colors.i2}
        }
        i2: (Id = 3) -> (Page = 2, Offset = 2) {
            style.fill: ${colors.i2}
        }
        i3: (Id = 5) -> (Page = 1, Offset = 2) {
            style.fill: ${colors.i2}
        }
        i4: (Id = 6) -> (Page = 2, Offset = 1)
        i5: (Id = 7) -> (Page = 3, Offset = 1)
    }
    b1: Index Bitmap {
        grid-gap: 0
        grid-rows: 2
        width: 600
        1: "Page 1" {
            width: 200
            style.fill: ${colors.i2}
        }
        2: "Page 2" {
            width: 200
            style.fill: ${colors.i2}
        }
        3: "Page 3" {
            width: 200
        }
        1m: "1" {
            width: 200
        }
        2m: "1" {
            width: 200
        }
        3m: "0" {
            width: 200
        }
    }
    i2: Grade Index {
        grid-gap: 0
        grid-columns: 1
        i1: (Grade = 5) -> (Page = 1, Offset = 1) {
            width: 600
        }
        i2: (Grade = 8) -> (Page = 2, Offset = 2) {
            style.fill: ${colors.i2}
        }
        i3: (Grade = 5) -> (Page = 1, Offset = 2)
        i4: (Grade = 6) -> (Page = 2, Offset = 1)
        i5: (Grade = 8) -> (Page = 3, Offset = 1) {
            style.fill: ${colors.i2}
        }
    }

    b2: Grade Bitmap {
        grid-gap: 0
        grid-rows: 2
        width: 600
        1: "Page 1" {
            width: 200
        }
        2: "Page 2" {
            width: 200
            style.fill: ${colors.i2}
        }
        3: "Page 3" {
            width: 200
            style.fill: ${colors.i2}
        }
        1m: "0" {
            width: 200
        }
        2m: "1" {
            width: 200
        }
        3m: "1" {
            width: 200
        }
    }
    i1 -> b1
    i2 -> b2
}
bc {
    style.opacity: 0
    grid-columns: 3
    grid-gap: 0
    s1: {
        class: none
        width: 300
    }
    b: AND Bitmap {
        grid-gap: 0
        grid-rows: 2
        1: "Page 1" {
            width: 200
        }
        2: "Page 2" {
            width: 200
            style.fill: ${colors.i2}
        }
        3: "Page 3" {
            width: 200
        }
        1m: "0" {
            width: 200
        }
        2m: "1" {
            width: 200
        }
        3m: "0" {
            width: 200
        }
    }
}

query -> data.i1
query -> data.i2
data.b1 -> bc.b
data.b2 -> bc.b
```

### Query Planner

Most of the time, we don’t manually decide which query execution strategy to use.
Instead, a component called **Query Planner** estimates and selects the most efficient query model based on cost:

- **Table Scan**: Reads all or most rows in a table.
- **Index Scan**: Used when highly selective conditions return a small number of rows.
- **Bitmap Index Scan**: Combines multiple indexes or handles queries that return a large number of rows.

But how does a database estimate the number of rows a query will process when the criteria are unpredictable?
Behind the scenes, it relies on various techniques, such as histogram distributions.
Periodically, the database randomly samples records and collects statistical metrics
such as **Most Common Values (MVC)** and **histogram buckets**..

#### Most Common Value (MVC)

**Most Common Value (MVC)** refers to the values that occur most frequently within a column. The database regularly samples records and identifies their MVCs.

When a query condition matches one of these MVCs, the database can quickly estimate the number of rows it will affect.

For example, consider the following statistics for the `Grade` column of the `Student` table.
If a query filters `Grade = 7`, the database instantly predicts a frequency of `0.5` based on the recorded stats.

| Most common values | Most common frequencies |
|--------------------|-------------------------|
| 7                  | 0.5                     |
| 9                  | 0.3                     |

#### Histogram Bucket

But what about queries involving values outside the common ones, or range-based conditions?
That’s where histogram buckets come in,
they capture data distribution by dividing values into buckets containing roughly equal numbers of rows.
The buckets might differ in range, but each holds a similar number of records.

For example, with the values:

`[1, 2, 3, 3, 3, 4, 4, 5, 10, 10, 10, 10, 20]`

If we configure the histogram to hold approximately 4 rows per bucket, they might look like:

- `[1, 2, 3, 3, 3]`
- `[4, 4, 5, 10]`
- `[10, 10, 10, 20]`

Internally, the database treats values within each bucket as uniformly distributed.
In other words, it only tracks the range of each bucket, simplifying estimation:

- `[1, 3]`
- `[4, 10]`
- `[10, 20]`

Using this, the estimated number of rows for a query like `BETWEEN 15 AND 20` is calculated as:

$Estimated\ Rows = \frac{Query\ Range}{Bucket\ Range} \times Number\ Of\ Rows\ Per\ Bucket = \frac{20-15}{20-10} \times 4 = 2$

It’s important to understand that both MVC and histogram metrics are derived from **randomly sampled** records within the table.
As a result, the collected statistics may not perfectly represent the actual distribution of data across the entire table.
While this estimation process isn’t flawless,
it provides the database with enough insight to make reasonably informed decisions when selecting the most efficient query execution strategy.

## Partition

Partitioning involves splitting a table into smaller, more manageable pieces, called **partitions**.

For example, a `User` table could be split into `UserActive` and `UserInactive` based on the `Active` status.

```d2
u: "UserTable"
ua: "UserActivePartition" {
    t: |||md
    (Name = John, Active = True)
    |||
}
ui: "UserInactivePartition"  {
    t: |||md
    (Name = Naruto, Active = False)
    |||
}
u -> ui
u -> ua
```

Here, the main `User` table serves as a proxy to the underlying partitions,
enabling improved performance by directing queries to smaller, more targeted tables.

This strategy is especially effective when the partitioning column frequently appears in queries.
However, when it’s not commonly queried, partitioning can degrade performance because:

- Table scans now require accessing multiple partitions (increasing I/O).
- Updates that modify the partitioning column involve moving records between tables, typically slower than simple in-place updates

## Denormalization

The final strategy is **Denormalization**,
manually restructuring tables by adding aggregated columns to avoid repetitive joins or calculations.

For example, given `Student` and `SubjectParticipation` tables,
calculating a student's GPA requires aggregating all subject grades.
If this is a frequent operation, it can impact performance:

```d2
student {
    shape: sql_table
    Id: "1"
    Name: "John"
}
s1: SubjectParticipation {
    shape: sql_table
    StudentId: "1"
    SubjectId: "1"
    Grade: "3"
}
s2: SubjectParticipation {
    shape: sql_table
    StudentId: "1"
    SubjectId: "2"
    Grade: "4"
}
student -> s1
student -> s2
```

To optimize performance, we can precompute and store the GPA directly in the `Student` table.
Any changes in the `SubjectParticipation` table would then trigger a recalculation of this field.

```d2
student {
    shape: sql_table
    Id: "1"
    Name: "John"
    GPA: "3.5"
}
```

Denormalization enables faster reads and is a fundamental principle in many [NoSQL databases]({{< ref "nosql-database" >}}).
However, it should be used carefully.

- More complex and slower updates, since aggregated fields must be recalculated and synchronized across multiple places.
- Broader transactions involving more updates, increasing chances of [concurrent conflicts and locking]({{< ref "concurrency-control">}}).
