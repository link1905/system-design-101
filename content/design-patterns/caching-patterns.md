---
title: Caching Patterns
weight: 30
prev: distributed-transaction
---

**Caching** is a crucial technique for optimizing system performance and conserving resources.
It involves temporarily storing and sharing data in a high-speed memory section.
This approach offers two main benefits:

- It avoids the need to retrieve data from slower physical storage.
- It allows the reuse of results from computationally intensive queries.

## Shared Cache

A common caching pattern involves sharing a cache among multiple servers.

Consider a web service as an example. If we want to cache a piece of data temporarily:

- Storing it locally on a single instance might make the service **stateful**.
This is because other instances might not be aware of the cached data, leading to inconsistencies.

```d2
s: Service {
    grid-rows: 1
    i1: Instance 1 {
        c: |||yaml
        A: 123
        |||
    }
    i2: Instance 2 {
        c: |||yaml
        A: 234
        |||
    }
}
```

- To address this, cached data can be moved to a dedicated shared store.
All instances will consistently serve the same data by accessing this central cache.

```d2
s: Service {
    i1: Instance 1 {
        class: server
    }
    i2: Instance 2 {
        class: server
    }
    db: Shared Cache {
        c: |||yaml
        A: 123
        |||
    }
    i1 <-> db
    i2 <-> db
}

```

### Distributed Cache

Cached data is often self-contained,
which allows for the creation of a distributed cache by sharding data across multiple servers.

```d2
Cache Cluster {
    s1: Cache Server 1 {
        c: |||yaml
        A: 123
        |||
    }
    s2: Cache Server 2 {
        c: |||yaml
        B: 234
        |||
    }
    s3: Cache Server 3 {
        c: |||yaml
        C: 113
        |||
    }
}
```

### Cache Compression

Cache components can be expensive due to their reliance on large amounts of fast memory.
**Compression** is an important, though often overlooked, method to reduce runtime costs.

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

### Cache Eviction

Due to the high cost of high-speed memory,
it's crucial to cache only necessary data.
This requires a cache eviction policy to remove older data and create space for new entries.

#### Least Recently Used (LRU)

The **Least Recently Used (LRU)** strategy is the most common approach.
When the cache reaches its capacity, the data that was least recently accessed is discarded.

```d2
direction: right
c1: Cache {
    c: |||yaml
    A: 
        value: 123
        lastAccessed: 00:03
    B: 
        value: 234
        lastAccessed: 00:10
    |||
}
c2: Cache {
    c: |||yaml
    B: 
        value: 234
        lastAccessed: 00:10
    |||
}
c1 -> c2: Evicted
```

This method is straightforward and widely used.
It is most effective when recent access patterns are a reliable indicator of future access.

#### Least Frequently Used (LFU)

**Least Frequently Used (LFU)** is applied when the access frequency is a better indicator of the data's access pattern.
In this case, the data with the lowest number of accesses is evicted.

```d2
direction: right
c1: Cache {
    c: |||yaml
    A: 
        value: 123
        accessCount: 100
    B: 
        value: 234
        accessCount: 11
    |||
}
c2: Cache {
    c: |||yaml
    A: 
        value: 123
        accessCount: 100
    |||
}
c1 -> c2: Evicted
```

From a programming standpoint,
**LFU** is more challenging and requires more resources to operate.
Basically, the choice between **LRU** and **LFU** should be based on the specific data access pattern.

---

Next, we will explore common patterns for effectively maintaining caches.

## Cache-aside (Lazy Loading)

Adapted from the [lazy loading pattern](https://en.wikipedia.org/wiki/Lazy_loading),
this strategy caches data only after it has been **recently read**.
In other words, the data must be initialized from the primary store for the first time,
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

## Write-through Cache

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

## Refresh-ahead Cache

Another common strategy is **Refresh-ahead**, which involves precomputing and caching data **periodically**.

For instance, in a game's ranking system,
directly sorting and paginating results from the primary user store for every query would be extremely costly.
Instead, the leaderboard can be computed and cached, say, every hour.

```d2
shape: sequence_diagram
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
```

The primary issue with this strategy is **staleness**; the data might be outdated between refresh intervals.
Despite this, **Refresh-ahead caching** is well-suited for computationally intensive data,
such as ranking systems, recommendation engines, and analytics dashboards.

## Write-Behind (Write-Back) Cache

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

## Client-Side Caching

In many scenarios, caching can be implemented on the client-side.
This helps reduce resource consumption on the server, particularly network bandwidth.

However, this approach should be used cautiously.
Client devices often have limited resources,
and implementing caching or performing heavy computations on them could significantly degrade the user experience.
