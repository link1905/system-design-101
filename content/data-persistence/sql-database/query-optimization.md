---
title: Query Optimization
weight: 20
params:
  math: true
---

To maintain data deduplication,
{{< term sql >}} normalizes data across multiple tables
and joins them when executing queries.
While this approach helps maintain data integrity, it can make queries less efficient
because they must access multiple tables,
which may reside in different files or even on different servers.

In this topic, we'll explore common techniques for improving query performance.
At its core, the guiding principle is simple: *minimize I/O operations as much as possible*.

## I/O Operation

**I/O (Input/Output)** refers to the process of transferring data between disk storage and memory (RAM).
There are two primary types of `I/O`:

- **Read I/O**: Transfers data from disk to memory for access.
- **Write I/O**: Persists changes from memory back to disk.

Since accessing disk storage is relatively slow,
an efficient database system should minimize I/O operations and leverage memory caching whenever possible.

### Memory Layer

The organization of data in memory closely mirrors its structure on disk.
However, rather than caching entire tables or indexes, the database caches only the **necessary pages**.

A common caching strategy is **LRU (Least Recently Used)**, which works as follows:

- **Cache Miss**: If a required page isn't already in memory, it is loaded from disk.
- **Eviction**: If memory is full, the least recently accessed pages are evicted to make room for new ones.

## Indexing

At runtime, how an index is used depends heavily on the query context.
There are typically three common query patterns:

### Index Scan

This is the standard way to use an index.
Retrieving a record typically involves at least **two I/O operations**:

1. One to locate the index entry.
2. Another to fetch the actual tuple from the heap.

For example, to find a student with `Id = 3`:

- First, locate the index entry for `Id = 3`.
- Then, follow its pointer to retrieve the corresponding record.

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

For large datasets, this back-and-forth access pattern can become inefficient.
When a query retrieves many records,
repeatedly jumping between index pages and heap pages can significantly degrade performance.

#### Index-Only Scan

To improve this, we can include **additional columns** in an index,
allowing queries to retrieve those values directly from the index without accessing the heap.

This technique is known as a **Covering Index**.
It reduces I/O by eliminating heap lookups for queries that require only values stored in the index.

For example, by including the `Name` field in the student index,
queries that retrieve only `Name` can be resolved directly from the index
without accessing the underlying tuple.

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

However, adding extra columns makes the index larger, increasing storage requirements.
It also makes updates more expensive because any change to an included column requires updating the index entry as well,
potentially preventing optimizations such as [HOT updates]({{< ref "physical-layer#hot-update" >}}).

### Table Scan

A **Table Scan** bypasses indexes entirely and sequentially reads the **entire set of table pages**.

This approach is typically chosen when the database estimates that a query will return a large proportion of the table's rows.
In such cases, scanning the table directly can be more efficient than repeatedly navigating between index and heap pages.

For example,
if most students have `GPA > 6`, the system might choose a table scan to retrieve those records.

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

A **Bitmap Scan** provides a hybrid strategy that is useful when:

- Repeatedly jumping between an index and the table would be inefficient.
- A full table scan would read too much unnecessary data.

Since indexes are generally smaller than the underlying table data,
scanning index pages first allows the database to gather information using fewer I/O operations.

In a **Bitmap Index Scan**:

1. The system scans the index to identify qualifying records and **marks the relevant pages** in a bitmap (`page number -> boolean`).
2. It then reads the marked pages and filters them to retrieve the matching tuples.

For example:

- The index might mark `Page 1` and `Page 2` for further filtering.
- The database then reads those pages to extract the qualifying records.

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

A **Bitmap Scan** can also combine **multiple index conditions** using **bitwise operations** on their respective bitmaps.
For instance:

- One index bitmap identifies pages containing records with `Id < 7`.
- Another bitmap identifies pages containing records with `Grade > 7`.
- The system performs a bitwise `AND` operation on the two bitmaps to identify pages that may satisfy both conditions, reducing unnecessary I/O.

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

Most of the time, we don't manually choose which query execution strategy to use.
Instead, a component called the **Query Planner** estimates the cost of different execution plans and selects the most efficient one.

Common strategies include:

- **Table Scan**: Reads all or most rows in a table.
- **Index Scan**: Typically used when highly selective conditions return a small number of rows.
- **Bitmap Index Scan**: Useful for combining multiple indexes or handling queries that return a moderate to large number of rows.

But how does a database estimate how many rows a query will process when query conditions are unpredictable?

Behind the scenes, it relies on statistical techniques such as frequency statistics and histogram distributions.
Periodically, the database **samples records** and collects metrics
such as **Most Common Values (MCVs)** and **histogram buckets**.

#### Most Common Values (MCVs)

**Most Common Values (MCVs)** are the values that occur most frequently within a column.
The database periodically samples records and identifies these frequently occurring values.

When a query condition matches one of these MCVs,
the database can quickly estimate the proportion of rows that are likely to satisfy the condition.

For example, consider the following statistics for the `Grade` column of the `Student` table.
If a query filters for `Grade = 7`, the database can estimate a frequency of `0.5` based on the recorded statistics.

| Most common values | Most common frequencies |
|--------------------|-------------------------|
| 7                  | 0.5                     |
| 9                  | 0.3                     |

#### Histogram Bucket

But what about queries involving less common values or range-based conditions?

This is where histogram buckets become useful.
They approximate the data distribution by dividing values into buckets containing **roughly equal numbers of rows**.
The buckets may cover different value ranges, but each contains a similar number of records.

For example, consider the following values:

`[1, 2, 3, 3, 3, 4, 4, 5, 10, 10, 10, 10, 20]`

If we configure the histogram to contain approximately four rows per bucket, the buckets might look like this:

- `[1, 2, 3, 3, 3]`
- `[4, 4, 5, 10]`
- `[10, 10, 10, 20]`

Internally, the database treats values within each bucket as **uniformly distributed**.
Rather than storing every individual value, it can represent each bucket using its range:

- `[1, 3]`
- `[4, 10]`
- `[10, 20]`

Using this approximation, the estimated number of rows for a query such as `BETWEEN 15 AND 20` can be calculated as:

$Estimated\ Rows = \frac{Query\ Range}{Bucket\ Range} \times Number\ Of\ Rows\ Per\ Bucket = \frac{20-15}{20-10} \times 4 = 2$

It's important to understand that both MCV and histogram statistics are derived from **sampled records** within the table.
As a result, the collected statistics may not perfectly represent the actual distribution of data across the entire table.

Although this estimation process isn't exact,
it gives the database enough information to make reasonably informed decisions when selecting an efficient query execution strategy.

## Partitioning

**Partitioning** involves splitting a table into smaller, more manageable pieces called **partitions**.

For example, a `User` table could be split into `UserActive` and `UserInactive` partitions based on the `Active` status.

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

Here, the main `User` table serves as a **proxy for the underlying partitions**,
allowing queries to target smaller and more relevant subsets of data.

This strategy is particularly effective when the partitioning column frequently appears in query conditions.
However, if the partitioning column is rarely used for filtering, partitioning can introduce additional overhead because:

- Table scans may need to access multiple partitions, increasing I/O.
- Updates that modify the partitioning column may require moving records between partitions, which is typically more expensive than a simple in-place update.

## Denormalization

The final strategy is **Denormalization**,
which involves deliberately restructuring data by storing derived or aggregated values to avoid repetitive joins or calculations.

For example, given `Student` and `SubjectParticipation` tables,
calculating a student's GPA requires **aggregating grades across all subjects**.
If this calculation is performed frequently, it can negatively affect query performance:

```d2
direction: right
student: Student {
    shape: sql_table
    Id: "1"
    Name: "John"
}
s1: SubjectParticipation {
    shape: sql_table
    Participation1: "StudentId=1,SubjectId=1,Grade=3"
    Participation2: "StudentId=1,SubjectId=2,Grade=4"
}
student -> s1
```

To optimize this operation, we can store the total score and number of subjects directly in the `Student` table.
Any relevant changes in the `SubjectParticipation` table would then require these aggregated fields to be updated as well.

```d2
student {
    shape: sql_table
    Id: "1"
    Name: "John"
    TotalScore: 7
    NumOfSubjects: 2
}
```

Denormalization enables faster reads and is a fundamental principle used by many [NoSQL databases]({{< ref "nosql-database" >}}).
However, it should be applied carefully.

Because duplicated or aggregated values must remain synchronized across multiple locations,
updates become more complex and potentially more expensive.
They may also require broader transactions involving additional writes,
increasing the likelihood of [concurrent conflicts and locking]({{< ref "concurrency-control">}}).
