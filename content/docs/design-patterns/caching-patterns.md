---
title: Caching Patterns
weight: 30
---

Caching plays a big role in maintaining system performance and saving resources.
We temporarily hold and share data in a fast memory section to:

- Avoid querying data from **slow** physical storage.
- Reuse results from **heavily computed** queries.

## Shared Cache

The most common pattern is sharing cache between servers.

Imagine you have a web service.
We want to store a piece of data temporarily:

- If we save it on a single instance locally,
the service becomes **stateful** because other instances may not know about that.

```d2
s: Service {
    i1: Instance 1 {
        "A=123"
    } 
    i2: Instance 2
}
```

- Then, we move shared data to a shared store.
The service becomes **stateless** now because they always serve the same data.

```d2
s: Service {
    i1: Instance 1 {
        class: server 
    }
    i2: Instance 2 {
        class: server 
    }
    db: Shared Store {
        "A=123"
    }
    i1 <-> db
    i2 <-> db
}
```

### Cache Compression

Cache components are expensive because of consuming large quantities of fast memory.
**Compression** is an important part to reduce the runtime cost, but usually neglected

Nothing huge, we compress data before caching and decompressing after retrieving.
Cache components are often busy with many clients,
we should leave the responsibility of compression on the client side.

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

Let's see some caching strategies and their usages:

### Cache-aside Cache

Adapted from [the lazy loading pattern](https://en.wikipedia.org/wiki/Lazy_loading),
we only cache data if it has been read recently.
In other words, we need to actually query a piece of data for the first time
and efficiently reuse the result for the next times.

For example,
a service fails to load cache from the `Cache Store`,
then it queries data from the primary store and caches it for further usages.

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
s -> c: 1. Get cache
c -> s: 2. Respond "Not found" {
    class: error-conn
}
s <-> db: 3. Query data
s -> c: 4. Cache data
```

This strategy is applied most of the time
because of simplicity and versatility.

The biggest problem of **Cache-aside Cache** is **inconsistency**.
Cached data is automatically deleted after a period of time,
but the source data can be updated and mismatched with it during its existence

Furthermore, when the system has many complex and heavy queries,
**cache miss** (data is absent from cache) penalty can be problematic.
The system can suffer a lot of concurrent cache misses accidentally.

### Write-through Cache

A more complex strategy of caching is **Write-through**.
Removing the laziness, we will actively cache data **beforehand**.
That means, when updating data, we also need to update the cache store correspondingly
to minimize **cache miss**.

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

Absolutely, this is far more challenging to develop and control.
Imagine we cache a complex query of several entities,
any changes in these entities will lead to a different result,
and the cache must be refreshed.
Moreover, the preparation is resource wasting if cached data is unused.

The write-through cache is especially used to:

- Ensure **consistency** for critical data,
because cached data is continuously synchronized with the data source.
- **Minimize cache miss** as cache is precomputed early,
extremely useful when cache misses are expensive.

### Refresh-ahead Cache

Another common strategy is **Refresh-ahead**,
precomputing and caching data **periodically**.

For example, you are developing a ranking system of a game.
Directly sorting and paging on the primary user store for each query is extremely expensive.
Hence, we should compute the currently temporary leaderboard every hour.

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

The biggest problem is staleness,
data is possibly stale between refreshes.

In fact, **Refresh-Ahead caching** is well suited for heavily computed data,
e.g., ranking system, recommendation system, analytics dashboards...

### Write-Behind Cache

Unlike the previous ones intending to read operations,
this strategy was born to improve **write performance**.
Instead of flushing update to the physical store immediately,
we will do it in batch.
The cache component is in charge of retaining in-flight updates,
and write them to the reliable storage after a threshold (e.g., a certain number of updates, some seconds, etc).

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
s -> c: 1. Update data
s -> c: 2. Update data
s -> c: 3. Update data
c -> db: 4. Flush the update
```

Properly, this is intolerant of data loss.
If the cache goes down before flushing data, we lose the incomplete updates forever.

**Write-Behind cache** is especially useful for write-heavy applications:

- Non-critical systems where data loss is unacceptable, e.g., metrics collection, user activity tracking. etc.
- Or data loss is recoverable, e.g., pouring data from sources to a [data lake](https://en.wikipedia.org/wiki/Data_lake).

### Client-Side Caching

Sometimes, we may leverage the client side to cache data.
This helps to reduce the resource usage on the server side, especially network bandwidth.

Be careful with this approach!
Client devices are weak and limited,
caching or heavy computation possibly result in significantly degrading user experience.

## Bloom Filter

**Bloom Filter** is a cache strategy helping quickly verify the existence of a piece of data.

In this method, we must setup:

1. A presence array of **bits** initializing with zeros.

| Index   | 0 | 1 | 2 | 3 | 4 | 5 |
| Absence | 0 | 0 | 0 | 0 | 0 | 0 |

2. A hash function mapping values to fall into the array, such as `value % array_size (6)`.

When a value is **added** or **verified**,
the associated element is marked as `1` to indicate its presence.

For example, we add the value `9`,
the element at position `9 % 6 = 3` is marked as `1`.

| Index   | 0 | 1 | 2 | 3 | 4 | 5 |
| Absence | 0 | 0 | 0 | 1 | 0 | 0 |

When a value is verified, if the associated element is off (zero),
we quickly state that the value does not exist.

For example, we verify the existence of `10`,
the element at position `10 % 6 = 4` is `0`,
that means `10` **does not exist**.

Because of **hashing collision**,
although the associated element is on, we **cannot** say that a value exists.

For example, we verify the existence of `15`,
the element at position `15 % 6 = 3` is `1` but it does not exist,
because its hash collides with `9`.

That's just it!
We can quickly determine the **absence** of a value.
This algorithm is typically implemented on the client side or a shared cache store
to reduce redundant network bandwidth and existence verifications.
