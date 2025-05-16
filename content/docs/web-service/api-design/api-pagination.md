---
title: API Pagination
---

Sometimes, an API returns a large amount of data that clients can’t process all at once.
This limitation might stem from hardware constraints (e.g., memory, network bandwidth) or application requirements (e.g., paginated responses).

In this topic, we'll explore common strategies for handling large datasets in API design.

## Large Binary File

A typical case involves downloading large binary files.
Low-level protocols like **FTP** or **SMB** aren't ideal here because we're designing a high-level interface.

### Chunking

**Chunking** is a practical solution in this scenario.
It involves splitting a file into smaller parts (chunks), enabling clients to request and handle data in portions.

For example, consider a client downloading a 20MB file:

1. The client first requests metadata from the server (name, type, size, etc.).

```d2
shape: sequence_diagram
c: Client {
    class: client
}
s: Server {
    class: server
}
c -> s: Request file
s -> c: Metadata (myfile.png, image, 20MB)
```

2. The client determines the number of chunks based on its capabilities (e.g., two 10MB chunks).
Once all chunks are downloaded, they're reassembled into the final file.

```d2
shape: sequence_diagram
c: Client {
    class: client
}
s: Server {
    class: server
}

c -> s: Request file
s -> c: "Metadata (myfile.png, image, 20MB)"
c <- s: "Download chunk 1 [0, 10]"
c <- s: "Download chunk 2 [11, 20]"
c -> c: "Reassemble file = chunk 1 + chunk 2"
```

**Chunking** also applies to file uploads.
Its benefits include:

- **Parallelism**: Chunks can be processed independently, enabling concurrent downloads or uploads.
- **Fault tolerance**: If a transfer fails, only the affected chunk needs to be retried.

## Pagination

When dealing with large collections of records, **Pagination** is a recommended technique to divide data into manageable pages.

Paging responses should provide navigation links, enabling clients to move between pages.

For example, a response includes both the content and supplementary pagination details.

```json
{
 // Page information
  "page": {
    "currentPage": 1,
    "pageSize": 10,
    "elementsCount": 30,
    "pagesCount": 3,
    "prevPage": "/users?page=0&size=10",
    "nextPage": "/users?page=2&size=10"
  },
  // Content
  "content": [
    { "id": 1234, "name": "John Doe" },
    { "id": 1345, "name": "Micheal" }
    // ... more records
  ],
}
```

### Best Practices

- **Implement pagination from the start**, even if the dataset is small initially.
Adding pagination later could break API compatibility.

- **Allow clients to specify page size**.
Fixed sizes can lead to poor user experience across devices with varying display sizes.
However, the server should enforce a reasonable upper limit to prevent abuse (e.g., **DDoS**).

### Filtering And Sorting Problem

If pages directly mirror the underlying data,
we can use offset-based access to quickly retrieve any page: `pages[n] = records[n × page_size ... (n + 1) × page_size]` (think of it like slicing a segment from an array).

However, pagination is often combined with filtering and sorting,
which complicates offset calculations because the filtered or sorted results no longer match the original storage order.

```d2
direction: right
d: Database {
    d: |||json
    [
        {
          "id": 1,
          "name": "John",
          "age": 20
        },
        {
          "id": 2,
          "name": "Micheal",
          "age": 17
        },
        {
          "id": 3,
          "name": "Abraham",
          "age": 25
        }
    ]
    |||
}
q: Query age > 18 AND sorted by name {
    d: |||json
    [
        {
          "id": 3,
          "name": "Abraham",
          "age": 25
        },
        {
          "id": 1,
          "name": "John",
          "age": 20
        }
    ]
    |||
}
d -> q
```

### Rowset Pagination

The most straightforward and flexible approach is to re-execute the query for each page request.

In {{< term sql >}}, we typically use the **LIMIT** and **OFFSET** clauses to paginate results:

- **LIMIT** defines the number of records per page.
- **OFFSET = page number × page size** skips records from previous pages.

For example, to fetch the third page (with a page size of 10) of users over 30:

```sql
SELECT *
FROM users
WHERE age > 30
OFFSET 20 LIMIT 10
```

Here’s what happens behind the scenes:

1. The database first retrieves **all rows** that match the **WHERE** condition.
2. Then, **OFFSET** is applied to skip the rows belonging to the previous pages.
3. Finally, **LIMIT** is used to select the rows for the requested page.

In other words, even though only a small portion of the records is returned,
the database still processes the entire result set up to the specified offset.

This can become a performance concern on large tables,
as it consumes unnecessary I/Os and processing resources for each paginated request.

> For a deeper explanation, see [SQL Query Optimization]({{< ref "query-optimization" >}}).

### Keyset Pagination

The **OFFSET** clause is applied **after the entire result set is generated**.
Although rows before the specified offset aren't returned, resources are still consumed to retrieve and validate them.

A more efficient alternative is to avoid using **OFFSET** altogether and instead rely on the **last fetched key** — a technique known as **Keyset Pagination**.

In this approach, each page response includes a keyset value, which the client uses to request the next page.

For example:

```json
{
  "page": {
    "keyset": 10,
    "nextPage": "/users?keyset=10"
  },
  "content": [
    {
      "id": 10,
      "name": "Micheal"
    }
  ],
}
```

The keyset is usually the table’s primary key (or any indexed, sortable field).
We incorporate it directly into the query’s **WHERE** clause, replacing the **OFFSET**.
This allows the database to skip over earlier records during the filtering stage, improving performance by reducing unnecessary processing.

```sql
SELECT *
FROM users
WHERE age > 30 AND id > :keyset
LIMIT 10
```

**However, this method comes with trade-offs:**

1. It doesn’t support direct navigation to arbitrary pages — clients must move sequentially through pages.
2. If new records are inserted **before** a client’s keyset, their view of the paginated data may drift out of sync with the current dataset, requiring a refresh to realign.

Thus, due to agility, **Rowset Pagination** is still a more preferred approach.

### Static Views

**Static View** is a [Refresh-ahead caching]({{< ref "caching-patterns#refresh-ahead-caching" >}}) implementation.
When an application doesn’t require real-time updates, pages can be **precomputed at scheduled intervals**.
This allows clients to quickly access any page by **page key**, without triggering additional server-side computation.

For example, pages might be **regenerated** every hour.
Clients can subsequently request any page, assured of its immediate availability:

```d2
db0: Database (00:00) {
    r: |||json
    {
        "Page 0": [
            {
              "id": 1,
              "name": "John"
            }
        ]
    }
    |||
}
db5: Database (01:00) {
    r: |||json
    [
        "Page 0": [
            {
              "id": 1,
              "name": "John"
            },
            {
              "id": 2,
              "name": "Micheal"
            }
        ]
    ]
    |||
}
db10: Database (02:00) {
    r: |||json
    [
        "Page 0": [
            {
              "id": 1,
              "name": "John"
            },
            {
              "id": 2,
              "name": "Micheal"
            }
        ],
        "Page 1": [
            {
              "id": 3,
              "name": "Abraham"
            }
        ]
    ]
    |||
}
```

One major advantage of **Static Views** is their ability to support **personalized content delivery**.
Many social media platforms use this approach to pre-generate customized feeds for different users, improving perceived performance and reducing real-time computation load.

While **Static Views** offer the fastest access times among pagination strategies, they come with notable limitations:

- **Inconsistency**: Between refresh cycles, pages might not reflect the most current data.
- **Limited flexibility**: Since pages are generated based on fixed, predetermined criteria, they can’t adapt to dynamic search queries or custom filters.
