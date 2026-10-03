---
title: API Pagination
prev: api-design
next: data-persistence
---

Sometimes, an API returns a large amount of data that clients cannot process all at once.
This limitation may stem from hardware constraints (e.g., memory or network bandwidth) or application requirements (e.g., paginated responses).

In this topic, we'll explore common strategies for handling large datasets in API design.

## Chunking

**Chunking** is a practical technique for transferring large binary files.
It involves splitting a file into smaller parts, called chunks, allowing clients to request and process the data incrementally.

{{% callout type="info" %}}
This is similar to the approach discussed in the [HLS protocol]({{< ref "communication-protocols#hls">}}) for serving media data.
{{% /callout %}}

For example, consider a client downloading a 20MB file:

{{% steps %}}

### Metadata

The client first requests the file's metadata from the server, such as its name, type, and size.

```d2
shape: sequence_diagram
c: Client {
   class: client
}
s: Server {
   class: server
}
c -> s: Request file
s -> c: Metadata (myfile.png, 20MB)
```

### Chunks

The client determines an appropriate chunk size based on its capabilities, such as downloading the file in two 10MB chunks.
Once all chunks have been downloaded, the client reassembles them into the final file.

```d2
shape: sequence_diagram
c: Client {
   class: client
}
s: Server {
   class: server
}

c -> s: Request file
s -> c: "Metadata (myfile.png, 20MB)"
c <- s: "Download chunk 1 [0, 10]"
c <- s: "Download chunk 2 [11, 20]"
c -> c: "Reassemble file = chunk 1 + chunk 2"
```

{{% /steps %}}

**Chunking** can also be applied to file uploads.
Its benefits include:

- **Parallelism**: Chunks can be processed independently, enabling concurrent downloads or uploads.
- **Fault tolerance**: If a transfer fails, only the affected chunk needs to be retried.

## Pagination

When dealing with large collections of records, **Pagination** is a common technique for dividing data into manageable pages.

Paginated responses should provide navigation information, allowing clients to move between pages easily.

For example, a response may include both the requested content and supplementary pagination metadata:

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

- *Implement pagination from the start*, even if the dataset is initially small, because introducing pagination later may break API compatibility.

- *Allow clients to specify the page size*:
A fixed page size may provide a poor experience across clients with different requirements and display sizes.
However, the server should enforce a reasonable upper limit to prevent excessive resource consumption or abuse, such as **DDoS** attacks.

### Filtering And Sorting Problem

If pages directly mirror the underlying data,
we can use offset-based access to retrieve any page efficiently:

`pages[n] = records[n × page_size ... (n + 1) × page_size]`

This is similar to slicing a segment from an array.

However, pagination is often combined with filtering and sorting,
which complicates offset calculations because the resulting order no longer directly matches the underlying storage order.

For example,
the final paginated result may present a completely different view of the underlying data:

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

- **LIMIT** specifies the maximum number of records returned per page.
- **OFFSET = page number × page size** skips the records belonging to previous pages.

Here's what typically happens behind the scenes:

1. The database identifies rows that satisfy the `WHERE` condition.
2. **OFFSET** skips the qualifying rows that belong to previous pages.
3. **LIMIT** stops the query after the requested number of rows has been produced.

For example, to fetch 20 users over the age of 30 after skipping the first 100,000 qualifying users:

```sql
SELECT *
FROM users
WHERE age > 30
ORDER BY id
OFFSET 100_000 LIMIT 20
```

The database may need to traverse roughly 100,020 qualifying entries before it can return the final 20.

{{< callout type="info" >}}
This does not necessarily mean loading 100,020 full rows. With an appropriate index and execution plan, much of this work may involve traversing index entries instead.
For a deeper explanation, see [SQL Query Optimization]({{< ref "query-optimization" >}}).
{{< /callout >}}

In other words, even though only a small number of records is returned, the database still has to process or skip qualifying results up to the specified offset.
This can become a performance concern for deeper pages because the amount of work generally increases as the offset grows.

### Keyset Pagination

A more efficient alternative is to avoid **OFFSET** altogether and instead rely on the **last fetched key**.
This technique is known as **Keyset Pagination**.

In this approach, each page response includes a keyset value that the client uses to request the next page.

For example:

```json
{
 "page": {
   "keyset": 10,
   "nextPage": "/users?keyset=10"
 },
 "content": [
   // ...
 ],
}
```

The keyset is typically based on the table's primary key or another indexed and sortable field.
It is incorporated directly into the query's **WHERE** clause, replacing the **OFFSET**.

This allows the database to seek directly to the relevant position in an index instead of repeatedly skipping all preceding results, which can significantly improve performance for deep pagination.

```sql
SELECT *
FROM users
WHERE age > 30 AND id > keyset
ORDER BY id
LIMIT 10
```

However, this approach comes with several trade-offs:

1. It does not naturally support direct navigation to arbitrary pages, so clients generally move sequentially through the result set.
2. Changes to the dataset between requests can affect what the client observes. Depending on the ordering key and the type of modification, records may be skipped, duplicated, or appear only after the client **refreshes** its view.

Therefore, **Rowset Pagination** remains useful when flexible page navigation is more important than efficient traversal of deep result sets.

### Static Views

A **Static View** is an application of [Refresh-ahead caching]({{< ref "caching-patterns#refresh-ahead-caching" >}}).
When an application does not require real-time updates, paginated results can be **precomputed at scheduled intervals**.

This allows clients to access any page directly by its **page key** without triggering the full server-side computation for every request.

For example, pages might be regenerated every hour.
Clients can subsequently request any page from the latest precomputed snapshot:

```d2
grid-rows: 1
db0: Database (00:00) {
   r: |||json
   [
       "Page 0": [
           {
             "id": 1,
             "name": "John"
           }
       ]
   ]
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

One major advantage of **Static Views** is that they can support **personalized content delivery**.
For example, some feed-oriented systems precompute customized result sets for individual users, reducing real-time computation and improving perceived response times.

While **Static Views** can provide very fast read access, they also introduce notable limitations:

- **Inconsistency**: Between refresh cycles, pages may not reflect the most recent data.
- **Resource intensive**: Because pages are generated and stored in advance, resources may be consumed for pages that are rarely or never requested.
