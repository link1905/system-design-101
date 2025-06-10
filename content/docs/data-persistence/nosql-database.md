---
title: NoSQL Database
weight: 30
prev: distributed-database
---

{{< term sql >}} has long been the primary choice for data storage.
Its general-purpose design makes it suitable for many use cases, but not all.
To address the rapid growth of data and its diverse requirements,
many alternative systems have been developed that outperform {{< term sql >}} in specific scenarios.
These are collectively referred to as {{< term nosql >}} (Not Only SQL).

{{< term sql >}} relies on the unified [Relational Model](https://www.geeksforgeeks.org/relational-model-in-dbms/),
while {{< term nosql >}} encompasses various types of databases with different data models
tailored to specific purposes.

## NoSQL Characteristics

Despite having diverse data models, {{< term nosql >}} databases often share several common characteristics.

### Schemaless

{{< term nosql >}} databases are typically **schemaless**,
meaning they don’t require predefined data schemas. This flexibility is possible because:

- Some systems treat data abstractly as a sequence of binary values, without enforcing any structure.
- Others infer schema dynamically from the data as it's inserted, rather than requiring it upfront.

This makes {{< term nosql >}} especially useful for
handling flexible or irregularly structured data, something that {{< term sql >}} struggles with.

For example, consider customer records with varying attributes.

- In {{< term sql >}}, we need to define a wide table with many columns, which often results in storing numerous **NULL** values (**NULL** still takes up storage space).
- Adding a new attribute requires altering and restructuring the table, which is inefficient.

```d2
customer01: |||yaml
id: 1
name: Mike
|||
customer02: |||yaml
id: 2
income: 10000
job: developer
|||
customer03: |||yaml
id: 3
name: Steve
school: ABC University
gpa: 3.5
|||
```

### Horizontal Scalability

NoSQL databases are designed for **horizontal scalability**.
They avoid complex relationships and instead rely on [sharding]({{< ref "peer-to-peer-architecture#shard-replication" >}}) to
distribute data across multiple servers.

For example, user records can be split across different servers:

```d2
grid-rows: 2
s1: Server 1 {
  customer01: |||yaml
  id: 1
  name: Mike
  |||
}
s2: Server 2 {
  customer02: |||yaml
  id: 2
  income: 10000
  job: developer
  |||
}
s3: Server 3 {
  customer03: |||yaml
  id: 3
  name: Steve
  school: ABC University
  gpa: 3.5
  |||
}
s1 <-> s2
s2 <-> s3
s3 <-> s1
```

#### No Relationships

**NoSQL** databases generally do not support relational features such as **foreign keys** or **joins**.

This design choice stems from their distributed architecture, data is often spread across multiple servers.
Attempting to join records stored on different nodes can lead to **performance bottlenecks** and **reduced availability**.

When using NoSQL, we need to shift our data modeling mindset from normalization to **denormalization**,
the goal is to create **self-contained records** that are fully queryable from a single server.
We'll explore this concept further in the [Document Store](#document-store) section.

### Eventual Consistency

Many {{< term nosql >}} databases prefer
[**Eventual Consistency**]({{< ref "distributed-database#eventual-consistency-level" >}}),
trading off strict consistency for better availability and performance.
They often maintain multiple replicas that converge over time.

#### No ACID

{{< term nosql >}} databases are designed for high availability and fault tolerance but generally do not fully support ACID transactions.

In particular, maintaining the **Isolation** property across multiple nodes is prohibitively expensive.
Ensuring strict transactional behavior would require constant coordination and message exchange over the network,
leading to a complex, tightly coupled system that contradicts **NoSQL**'s design goals.

As a result, **NoSQL** is generally not well-suited for applications that demand
strong consistency and strict transactional guarantees.

## Key-value Store

We start with the simplest model: the **Key-Value Store**.

In this model, each record is identified by a **unique key**, and accessed via two basic operations:

- `Put(key, value)`
- `Get(key) => value`

This is similar to building a [Hash Table](https://www.geeksforgeeks.org/hash-table-data-structure/)
that operates as a **shared process**.

Example:

```d2
s: Service {
  class: server
}
k: Key-value Store {
  c: |||yaml
  cache_page_0: "<html>This is page 0</html>"
  user_0_devices: "['apple_998', 'android_110']"
  |||
}
s -> k: "Get('cache_page_0')"
k -> s: "Respond '<html>This is page 0</html>'"
```

### Use Cases {#use-cases-kv}

**Key-value stores** are ideal when data naturally fits the key-value model.
Absolutely, they shouldn't be used for non-key queries like aggregations.

Internally based on **Hash Table** and memory, they perform extremely fast key-based lookups.
That makes them perfect for use cases like [distributed caching]({{< ref "caching-patterns" >}}).

```d2
s1: Server 1 {
  class: server
}
s2: Server 2 {
  class: server
}
c: Distributed Cache {
  class: cache
}
s1 <-> c
s2 <-> c
```

## Document Store

A **Document Store** organizes data around **documents**,
each one representing a single record, typically in [JSON](https://www.json.org/json-en.html) format.

```d2
s1: Student Document 1 {
  c: |||json
  {
    "id": "student_a",
    "name": "Student A"
  }
  |||
}
s2: Student Document 2 {
  c: |||json
  {
    "id": "student_b",
    "name": "Student B",
    "class_id": "class_a",
    "gpa": 8
  }
  |||
}
```

Similar documents are grouped into **collections**, making management and retrieval simpler.
For example, `student_collection` and `room_collection`:

```d2
student_collection: {
  grid-rows: 2
    student A: |||json
{
    "id": "student_a",
    "name": "Student A"
}
|||
    student B: |||json
{
    "id": "student_b",
    "name": "Student B",
    "class_id": "class_a",
    "gpa": 8
}
|||
}
room_collection: {
    class A: |||json
{
    "id": "class_a",
    "position": "4th floor"
}
|||
}
```

At a high level, documents are conceptually similar to SQL **rows**, and collections resemble **tables**.
However, there are some key differences:

- Documents within the same collection can have different schemas.
  For example, `student_b` has more fields than `student_a`.
  This schema flexibility means structural changes to one document do not impact the rest of the collection.

- There's no native concept of relationship.
  Instead, references are made using plain values, e.g., `student_b` refers to `class_a` by storing its id as a plain string.

### Data Denormalization {id="doc_denormalize"}

**Document Stores** are typically distributed, documents within the same collection may reside on different nodes.
In such an environment, joining records would require querying multiple servers,
which can severely impact performance and availability.

For example, imagine we want to track student registrations across multiple classes.
In a relational model, this data is usually normalized into two separate tables: student and registration.
To obtain a student’s class registrations, we would join these tables using a shared key.
However, executing such joins often requires scanning multiple servers to gather all relevant records,
making complex queries resource-intensive and costly.

```d2
direction: right
s1: Server 1 {
  s: student {
      shape: sql_table
      id: stu123
      name: Mike
  }
}
s2: Server 2 {
  r: registration {
      shape: sql_table
      student_id: stu123
      class_id: class123
  }
}

s3: Server 3 {
  r: registration {
      shape: sql_table
      student_id: stu123
      class_id: class123
  }
}
s2.r.student_id -> s1.s.id
s3.r.student_id -> s1.s.id
```

In a **Document Store**, however, we often embed related data to keep records **self-contained**.
Instead of splitting into two collections, we can include a list of class IDs directly within the `student` document:

```d2
student: {
    shape: sql_table
    id: stu123
    name: Mike
    classIds: "['class123', 'class234']"
}
```

This design makes reads more efficient in distributed environments by ensuring that read requests are directed to a **single server**.
However, a single value may appear repeatedly across multiple documents,
meaning that updates must be applied to each occurrence individually,
an operation that can be both costly and prone to errors.

### Value-Based Search

Under the hood, many **Document Stores** use storage structures similar to those in **SQL**,
such as **Heap** and **B-tree**.
This enables indexing not just on document ids, but also on arbitrary fields.

For example, consider the following document.
We can perform fast queries on any indexed field, such as `name` or `gpa`.

```json
{
  "id": "stu123",
  "name": "Mike",
  "gpa": 8.4
}
```

Actually, there's no magic involved;
When querying by a non-key field,
the system needs to scan multiple servers internally, because it does not know in advance where these records are stored.

### Use Cases {id="use-cases_doc"}

In a way, a **Document Store** feels like a more flexible,
schema-relaxed version of {{< term sql >}}.
**Document Stores** are particularly effective when:

- Records have complex and variable structures, such as product catalogs with diverse attributes.
- We need searchable fields beyond primary keys, enabling rich queries based on values.

## Column-Oriented Store

### Row-Oriented Model

In {{< term sql >}} databases, data is typically stored in a **row-oriented** layout,
where each complete row is stored contiguously on disk.
For example, a simple student table in an {{< term sql >}} database might look like this:

```md
student1, Steve, 3.5
student2, Mike, 3.0
student3, John, 2.5
```

This format is ideal for scenarios where most or all columns of a row are frequently accessed together.

However, it's inefficient for analytical queries that require only a subset of columns.
For example, calculating the `average GPA` involves reading entire rows into memory,
even though we only need the third column.
This results in excessive I/O and wasted memory due to unnecessary data loading.

{{< callout type="info" >}}
You can refer to the [Physical Layer of SQL]({{< ref "physical-layer" >}}) for more insights on this.
{{< /callout >}}

### Column-Oriented Model

A **Column-Oriented Store** takes the opposite approach: it groups and stores data **by column** instead of by row.
Rewriting the student data in columnar format:

```md
student1, student2, student3
Steve, Mike, John
3.5, 3.0, 2.5
```

Each column is stored as a contiguous block of values.
So, when calculating the `average GPA`, the database only reads the third block,
skipping irrelevant data and reducing I/O overhead and memory usage.

### Use Cases {id="cs-use-cases"}

**Column-Oriented Stores** are well-suited for **analytical workloads**,
particularly those in the [Online Analytical Processing (OLAP)](https://en.wikipedia.org/wiki/Online_analytical_processing) domain.
These workloads typically involve reading a few columns across a large number of rows,
perfect for columnar optimization.

However, this design is not ideal for [Online Transaction Processing (OLTP)](https://en.wikipedia.org/wiki/Online_transaction_processing) use cases,
where full row access is frequent.
For example, updating a single record in a table with 100 columns might require
100 separate I/O operations, since each column is stored in a different location.

Due to this overhead, column stores generally:

- Perform best on read-heavy, analytical use cases.
- Recommend batch imports instead of row-by-row inserts.

## Column-Family Store

A **Column-Family (CF)** store organizes data by grouping related columns together.
Each data row is represented as a set of columns, called a **column family**.

Let’s clarify the concept with a comparison.

In {{< term sql >}}, data is stored row-by-row according to a strict schema.
To access the `GPA` of a student, we need to refer to the value at the **third column** within the row:

```md
student1, Steve, 3.5, England
student2, NULL, 3, Vietnam
```

In contrast, a **Column-Family Store** treats each row as a flexible map of key-value pairs.
Column names are explicitly stored with each row:

```yaml
row-1:
  id: student1
  name: Steve
  gpa: 3.5

row-2:
  id: student2
  name: Mike
  gpa: 3
  dob: 09-12-2000
  address: HCMC
```

### Benefits of Column Families

- **No NULLs**: There is no need to store **NULL** values.
- **Schema Flexibility**: Adding or removing columns affects only the relevant rows, not the whole table.

### Log-Structured Merge Trees (LSM)

Column-Family Stores are typically backed by a **Log-Structured Merge Tree (LSM)** architecture,
optimized for **high write throughput** and large-scale storage.

#### Memory Layer

In memory, LSM combines two principles:

1. [Write-Behind Caching]({{< ref "caching-patterns#write-behind-caching" >}}):
   Changes are temporarily stored in an in-memory **MemTable**.
   When the **MemTable** reaches a size threshold, it is flushed to disk to save data.
2. [Write-Ahead Logging (WAL)]({{< ref "system-recovery#logging" >}}):
   Writes are **immediately logged** to disk via a **WAL** for durability.
   If the system crashes before flushing, the **WAL** ensures no data is lost.

```d2
direction: right
s: Store {
    m: MemTable {
        class: cache
    }
    wal: WAL {
        class: file
    }
    l: Hard drive {
      class: hd
    }
    m -> l: Flush periodically {
      style.animated: true
    }
}
c: Client {
    class: client
}
c -> s.m: 1. Update in memory
s.m -> s.wal: 2. Log the operation
```

#### Storage Layer

In the storage layer, *LSM** structures data into multiple **levels**:

- Each level consists of several immutable **Sorted String Tables (SSTables)**.
- An **SSTable** stores sorted key-value pairs, allowing [fast binary search](https://en.wikipedia.org/wiki/Binary_search) for efficient lookups.

Consider an example of an **LSM tree** with two levels, where records are stored as `key=value` pairs.
In this structure, it's possible for the same key to appear across multiple levels,
each representing a different version of the record.

```d2
s: Store {
    grid-columns: 1
    l0: Level 1 {
      s0: SSTable 0 {
        grid-gap: 0
        grid-rows: 1
        "a=100"
        "b=2000"
        "c=50"
      }
    }
    l1: Level 2 {
        s0: SSTable 0 {
          grid-gap: 0
          grid-rows: 1
          "a=120"
          "b=2000"
        }
        s1: SSTable 1 {
          grid-gap: 0
          grid-rows: 1
          "c=100"
          "d=50"
          "e=60"
        }
    }
}
```

#### Compaction & Merging

In the background,
**periodic compaction** is performed, pushing data down to deeper levels.
This design enables extremely fast writes,
updates are first written to the memory layer, then flushed and reorganized asynchronously in the storage layer.

Let’s walk through an example to understand how this process works in practice:

{{% steps %}}

##### Step 1: Flushing MemTable

Once the MemTable fills up, it’s flushed as a new **SSTable** into `Level 1`.

```d2
s: Store {

  grid-rows: 3
  m: MemTable {
      "a=110"
      "d=70"
  }
  l1: Level 1 {
    s0: SSTable 0 (Existing) {
        "a=100"
        "b=2000"
        "c=50"
    }
    s1: SSTable 1 (Newly Flushed) {
      style.fill: ${colors.i1}
      "a=110"
      "d=70"
    }
  }
  l2: Level 2 {
      s0: SSTable 0 {
          "a=120"
          "c=2000"
      }
  }
  m -> l1.s1: Flush {
    style.bold: true
  }
}
```

##### Step 2: Merging Within Level 1

`Level 1`'s **SSTables** are compacted and merged to discard duplicates.

For instance, the previous insertion triggers merging at `Level 1`.

- `Level 1` merges between `SSTable 0` and `SSTable 1`.
- `SSTable 1` is newer and overwrites `SSTable 0`.

```d2
s: Store {
    grid-columns: 1
    m: MemTable
    l1: Level 1 {
        s0: SSTable 0 (Created 00:01) {
          style.fill: ${colors.i2}
          "a=100"
          "b=2000"
          "c=50"
        }
        s1: SSTable 1 (Created 00:02) {
          style.fill: ${colors.i2}
          "a=110"
          "d=70"
        }
        m: Merged SSTable {
          style.fill: ${colors.i1}
          "a=110"
          "b=2000"
          "c=50"
          "d=70"
        }
        s0 -> m
        s1 -> m
    }
    l2: Level 2 {
        s0: SSTable 0 {
          "a=120"
          "c=2000"
        }
    }
}
```

##### Step 3: Level Promotion

If `Level 1` becomes full, it'll be promoted to the `Level 2`.
`Level 2` similarly merges its **SSTables** to discard duplicates.

```d2
s: Store {
    grid-columns: 1
    m: MemTable
    l1: Level 1 {
      m: SSTable 0 {
        style.fill: ${colors.i2}
        "a=110"
        "b=2000"
        "c=50"
        "d=70"
      }
    }
    l2: Level 2 {
        s0: SSTable 0 {
          "a=120"
          "c=2000"
        }
        s1: SSTable 1 (Newer) {
          style.fill: ${colors.i1}
          "a=110"
          "b=2000"
          "c=50"
          "d=70"
        }
        m: "Merged SSTable" {
          "a=110"
          "b=2000"
          "c=50"
          "d=70"
        }
        s0 -> m
        s1 -> m
    }
    l1.m -> l2.s1: Moved to
}
```

The final result looks clean, `Level 2` is the only one containing data.

```d2
s: Store {
    grid-columns: 1
    m: MemTable
    l1: Level 1 (Empty)
    l2: Level 2 {
      s0: SSTable 0 {
        "a=110"
        "b=2000"
        "c=50"
        "d=70"
      }
    }
}
```

{{% /steps %}}

**Why do we need levelling?**

Each level holds increasingly larger and older data sets.
By prioritizing searches in shallower levels, **LSM** achieves cache-like behavior,
where **recently written** data is accessed faster.

```d2
grid-rows: 3
vertical-gap: 10
l1: "" {
  class: none
  l: "Level 1" {
    width: 300
  }
}
l2: "" {
  class: none
  "..." {
    width: 500
  }
}
l3: "" {
  class: none
  l: "Level N" {
    width: 1000
  }
}
l1.l -> l3.l: Older and larger
```

### Use Cases {id="use-cases_cf"}

**Column-Family Stores** are highly suitable for:

- **Write-heavy** workloads.
- **Key-based** access patterns.

However, they’re not ideal for:

- **Write-once workloads**: Constantly unique data leads to deep LSM trees with minimal compaction benefits.
- **Value-based queries**: Since data is indexed by keys, querying non-key fields is inefficient.

## Search Engine

We now arrive at the final type of data store in this discussion: the **Search Engine**.
As its name implies, a **Search Engine** is purpose-built to support search operations,
especially **text-based searches**.

### Inverted Index

At its core, a search engine relies on the **Inverted Index** data structure.

Traditional indices map from **keys** to **values**. For instance, in a student database:

```yaml
fromIdToName:
  student01: Mike
  student02: John
  student03: Mike
```

An **Inverted Index** reverses the usual direction of key-value mapping.
Instead of mapping from a unique key to a value, it maps from a value (which is often not unique) back to one or more keys:

```yaml
fromNameToIds:
  Mike: ["student01", "student03"]
  John: ["student02"]
```

This reversal is especially powerful for text search, as it allows us to quickly find *where* a given word or value appears.

### Full-Text Search

In a conventional database, search operations are performed against raw fields as inserted.

Take the example of an inverted index of books:

```yaml
fromTitleToIds:
  Little Prince: ["book01"]
  Little Women: ["book02"]
```

To search books by title,
we'd need to scan through all the title indices with a basic matching tools
like SQL's **LIKE**, which is inefficient.

#### Full-Text Store

To enable fine-grained search, search engines tokenize text fields into individual **terms**,
indexing each one separately.
These terms and their mappings form the **Full-text Store**, a type of inverted index.

```yaml
terms:
  Little: ["book01", "book02"]
  Prince: ["book01"]
  Women: ["book02"]
```

```d2
grid-rows: 2
i: Full-text Store {
  grid-rows: 1
  grid-gap: 0
  e1: "Little" {
    width: 200
  }
  e2: "Prince" {
    width: 200
  }
  e3: "Women" {
    width: 200
  }
}
d: Data {
  grid-rows: 1
  grid-gap: 0
  b1: |||yaml
  book01:
    title: Little Prince
  ||| {
    width: 300
  }
  b2: |||yaml
  book02:
    title: Little Women
  ||| {
    width: 300
  }
}
i.e1 -> d.b1
i.e2 -> d.b1
i.e1 -> d.b2
i.e3 -> d.b2
```

Now, searching for a term like `Prince` instantly returns all associated records via the inverted index,
no scanning required.

### Text Analysis

**Search Engine** has a component called **Analyzer**
splitting and processing text into consistent, searchable terms.

#### Insertion Analysis

When data is inserted, it is passed through an analyzer which tokenizes and transforms it
(e.g., lowercasing, removing punctuation):

```d2
direction: right
o: "Little Prince"
t: Terms {
  grid-gap: 0
  grid-rows: 1
  t2: little
  t3: prince
}
a: Analyzer {
  class: process
}
o -> a: Insert
a -> t
```

`Little Prince` becomes `[little, prince]`, ready for indexing.

#### Query Analysis

The same analyzer is applied to search queries to ensure consistency.

For example, if a user searches for `litTLe`, which may not exactly match the version stored in the index,
the analyzer needs to process both the query and the stored data in the same way, producing a consistent value.

```d2
direction: right
o: "litTLe"
t: "little"
a: Analyzer {
  class: process
}
o -> a: Search
a -> t
```

**Search Engine** can also use advanced analyzers to handle synonyms, stemming, or multiple languages, further enriching the search experience.

### Use Cases {id="use-case-se"}

Search engines are typically **not** used as the primary database. Reasons include:

- **High storage overhead**: Indexing every term from raw text can consume significant space, often exceeding the size of the data itself.
- **Limited non-text capabilities**: Operations like aggregations, transactional updates, or structured queries are often better handled by traditional databases.

Instead, they shine as **search satellites**,
optimized for querying, layered on top of primary databases like {{< term sql >}} or **Document Stores**:

```d2
direction: right
main: "Primary Database" {
  class: db
}
s: "Search Engine" {
  class: se
}
main -> s: Replicate {
  style.animated: true
}
```

In this architecture, changes are replicated to the search engine,
enabling full-text search without compromising the performance or integrity of the main database.

## Other NoSQL Databases

Beyond the stores we’ve covered, there are still many other types of **NoSQL databases**,
such as **Graph**, **Time Series**, and more.

While we won’t explore all of them here, here’s a useful framework you can use when learning a new type of database:

- **What problem does it aim to solve?**
  Every database emerges to address specific challenges. Understand the motivation behind its design is a must.

- **What is its data model?**
  Is it graph-based, time-based, or something else? How is the model implemented in practice?

- **What are its typical use cases?**
  Consider what kinds of applications benefit most from this database. That often reveals its strengths and limitations.
