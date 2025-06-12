---
title: Caching Patterns
weight: 30
---

Caching is a crucial technique for optimizing system performance and conserving resources.
It involves temporarily storing and sharing data in a high-speed memory section.
This approach offers two main benefits:

- It avoids the need to retrieve data from slower physical storage.
- It allows the reuse of results from computationally intensive queries.

## Shared Cache

A common caching pattern involves sharing a cache among multiple servers.
Consider a web service as an example. If we want to cache a piece of data temporarily:

- Storing it locally on a single instance might makes the service **stateful**.
This is because other instances might not be aware of the cached data, leading to inconsistencies.

```d2
s: Service {
    grid-rows: 1
    i1: Instance 1 {
        "A: 123"
    }
    i2: Instance 2
}
```

- To address this, shared data can be moved to a dedicated shared store.
All instances will consistently serve the same data by accessing this central cache.

```d2
s: Service {
    i1: Instance 1 {
        class: server
    }
    i2: Instance 2 {
        class: server
    }
    db: Shared Store {
        "A: 123"
    }
    i1 <-> db
    i2 <-> db
}

```

### Cache Compression

Cache components can be expensive due to their reliance on large amounts of fast memory.
**Compression** is an important, though often overlooked, method to reduce runtime costs.

The process is straightforward: data is compressed before being cached and decompressed after retrieval.
Since cache components are frequently busy serving many clients,
it's generally better to assign the responsibility of compression and decompression to the client-side.

```d2
shape: sequence_diagram
s: Service {
    class: server
}
db: Cache Store {
    class: cache
}
s -> s: Compress data
s -> db: Cache
s <- db: Retrieve
s <- s: Decompress data
```

## Caching Strategies

Let's explore some common caching strategies and their typical use cases:

### Cache-aside (Lazy Loading)

Adapted from the [lazy loading pattern](https://en.wikipedia.org/wiki/Lazy_loading),
this strategy caches data only after it has been recently read.
In other words, the data must be queried from the primary store for the first time,
and then the result is efficiently reused for subsequent requests.

For example, if a service attempts to load data from the `Cache Store` and doesn't find it (a **cache miss**),
it then queries the data from the primary store and caches it for future use.

```d2
shape: sequence_diagram
s: Service {
    class: server
}
c: Cache Store {
    class: cache
}
db: Primary Store {
    class: db
}
s -> c: Get cache
c -> s: Cache miss {
    class: error-conn
}
s <- db: Query data
s -> c: Cache data
```

This strategy is widely applied due to its simplicity and versatility.
However, the main drawback of **Cache-aside** is potential **inconsistency**.
Cached data is typically evicted (deleted) after a certain period.
During its lifetime in the cache, the source data in the primary store might be updated,
leading to a mismatch between the cached version and the source.

Furthermore, if the system handles many complex and resource-intensive queries,
the penalty for a **cache miss** (when requested data is not found in the cache) can be significant.
The system might experience numerous concurrent cache misses, leading to performance degradation.

### Write-through Cache

A more complex caching strategy is **Write-through**.
This approach abandons laziness and actively caches data **beforehand**.
When data is updated in the primary store, it is also simultaneously updated in the cache store.

```d2
shape: sequence_diagram
s: Service {
    class: server
}
db: Primary Store {
    class: db
}
c: Cache Store {
    class: cache
}
s -> db: 1. Update data
s -> c: 2. Update cache
s <- c: 3. Retrieve cache
```

Developing and managing a write-through cache is considerably more challenging.
Imagine caching the result of a complex query involving several data entities.
Any change in these entities would alter the query result, necessitating a refresh of the cache.
Moreover, this preparatory caching can be resource-intensive if the cached data is ultimately not used.

The write-through cache is particularly useful for:

- Ensuring **consistency** for critical data, as the cache is continuously synchronized with the data source.
- **Minimizing cache misses**, especially when cache misses are computationally expensive, because the cache is precomputed.

### Refresh-ahead Cache

Another common strategy is **Refresh-ahead**, which involves precomputing and caching data **periodically**.

For instance, in a game's ranking system,
directly sorting and paginating results from the primary user store for every query would be extremely costly.
Instead, the leaderboard can be computed and cached, say, every hour.

```d2
shape: sequence_diagram
client: Client {
    class: client
}
s: Game Service {
    class: server
}
db: Primary Store {
    class: db
}
c: Cache Store {
    class: cache
}
s <- db: Rank data
s -> c: Update the leaderboard
s -> s: Wait 1 hour {
    style.bold: true
}
s <- db: Rank data
s -> c: Update the leaderboard
client -> c: Query the leaderboard
```

The primary issue with this strategy is **staleness**; the data might be outdated between refresh intervals.
Despite this, **Refresh-ahead caching** is well-suited for computationally intensive data,
such as ranking systems, recommendation engines, and analytics dashboards.

### Write-Behind (Write-Back) Cache

Unlike the previous strategies that focus on read operations, the **Write-Behind** strategy is designed to improve **write performance**.

Instead of immediately writing updates to the physical store, changes are batched and written asynchronously.
The cache component temporarily holds these in-flight updates and flushes them to the reliable storage after a certain threshold is met (e.g., a specific number of updates or a time interval).

```d2
shape: sequence_diagram
s: Service {
    class: server
}
c: Cache Store {
    class: cache
}
db: Primary Store {
    class: db
}
s -> c: Update data
s -> c: Update data
s -> c: Update data
c -> db: Flush the updates
```

A significant risk with this approach is potential data loss.
If the cache system fails before flushing the data, any unwritten updates will be lost permanently.
**Write-Behind cache** is particularly useful for write-heavy applications, such as:

- Non-critical systems where some data loss is tolerable (e.g., metrics collection, user activity tracking).
- Systems where data loss is recoverable
(e.g., transferring data from various sources to a [data lake](https://en.wikipedia.org/wiki/Data_lake)).

### Client-Side Caching

In many scenarios, caching can be implemented on the client-side.
This can help reduce resource consumption on the server, particularly network bandwidth.

However, this approach should be used cautiously.
Client devices often have limited resources,
and implementing caching or performing heavy computations on them could significantly degrade the user experience.

## Bloom Filter

A **Bloom Filter** is a probabilistic data structure that functions as a caching strategy to rapidly determine
if an element is not part of a set.
This makes it especially valuable for sidestepping costly lookups for items already known to be missing.

To set up a Bloom Filter, we need two main components:

1. A **bit array** (an array of bits), where all bits are initially set to zero. This array indicates the presence of elements.

    | Index | 0 | 1 | 2 | 3 | 4 | 5 |
    |-------|---|---|---|---|---|---|
    | Value | 0 | 0 | 0 | 0 | 0 | 0 |

2. A **hash function** that maps input values to indices within this bit array.
A straightforward example of such a function is `value % array_size`.

When an element is **added** to the set or its existence is checked:

- The hash function processes the element.
- The bits at the indices generated by the hash function are set to `1`.

- For instance, if we add a user with `userId = 9` and our hash function is `value % 6`.
The calculation `9 % 6` yields `3`. As a result, the bit at index `3` is changed to `1`.

    | Index | 0 | 1 | 2 | 3 | 4 | 5 |
    |-------|---|---|---|---|---|---|
    | Value | 0 | 0 | 0 | **1** | 0 | 0 |

When **checking** if an element is in the set.
If the bit at the index produced by the hash function is `0`,
then the element is **guaranteed to be absent** from the set.

- For example, to check if `userId = 10` is present.
The calculation `10 % 6` results in `4`.
Since the bit at index `4` is `0`, we can conclude that `user 10` does not exist.

A crucial aspect of Bloom Filters is their potential to generate **false positives** due to **hash collisions**.
This means that if the bit at the relevant index is `1`,
the filter suggests that the element *might* be present, but this is not certain.

- For example, to check if `userId = 15` is present.
- The calculation `15 % 6` gives `3`. The bit at index `3` is `1`.
However, in this scenario, `15` was never actually added to the set; its hash value collides with that of `9`.

This algorithm is commonly implemented within a shared cache store.
Its purpose is to minimize unnecessary network traffic and reduce the overhead associated with performing existence checks on the primary data store.
