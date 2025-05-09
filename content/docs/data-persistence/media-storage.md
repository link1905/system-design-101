---
title: Media Storage
weight: 40
---

Media data (videos, images, documents...) is unstructured binary data.
Storing them inside a database is not a smart choice,
instead, we should directly read as a normal file system.

Media data is extremely critical for client-facing services,
it's even usually the backbone of many applications, e.g., Netflix, YouTube...
In this topic, we will go through some common ways to manage it

## File Storage

**File Storage** refers to a solution that builds the storage as a live file system,
helping scale a machine and its storage unit independently.

It's perfect for **high performant** and low-latency applications while
they can work with complete files directly through network.
Additionally, multiple clients (service or end-user) can rely on the same storage server.

```d2
s1: Server 1 {
  class: server
}
d1: File Storage {
  "Docs/" {
    "doc1.md" {
      class: file
    }
    "doc2.md" {
      class: file
    }
  }
  "Images/" {
    "img.png" {
      class: file
    }
  }
}
s2: Server 2 {
  class: server
}
s1 <-> d1 
s2 <-> d1
```

Of course,
we can build replicas to enhance availability and performance.

```d2
p: Primary File Server {
    class: server
}
r1: Replica 1 {
    class: server
}
r2: Replica 2 {
    class: server
}
p -> r1 {
    style.animated: true
}
p -> r2 {
    style.animated: true
}
```

However, this setup reminds us of the [Master-slave model](../distributed-database/master-slave-architecture).
The primary server becomes a single point of failure,
degrading the service availability.

## Object Storage

This is a new technology but gaining tremendous momentum recently.
The best advantage of **Object Storage** is its distributed nature.
File is now called **Object**, has no relationship and is autonomously distributed to many servers,
resulting in a highly available and fault-tolerant system.

Let's see how an **Object Storage** is implemented in attempts:

### Object Distribution

Similar to what we did in the [Peer-to-peer architecture](../distributed-database/peer-to-peer-architecture) topic.
We want to distribute objects to many servers,
allowing them to write changes concurrently.
To make it better, a server also has some replicas.

```d2
s1: Server 1 {
  "img.png (primary)" {
    class: file
  }
  "doc.md (rep)" {
    class: file
  }
}
s2: Server 2 {
  "doc.md (primary)" {
    class: file
  }
  "img.png (rep)" {
    class: file
  }
}
s1 <-> s2
```

#### Object Naming

However, this make the way of interacting with objects more complex.
Normally, we access objects with a tree-like pattern, such as `Docs/doc.md".
But for now, we don't have a centralized file system,
that means there is no relationship (directory, sibling...) between objects.

**Object Storage** tends to mimic the file system notation by including the **full path** in object keys,
such as:

```d2
s1: Server 1 {
  "Image/img.png": {
    class: file
  }
  "Docs/doc.md": {
    class: file
  }
}
```

But this is only a trick,
when performing folder operations like *listing all files in a folder*,
we still needs to scan and work with all relevant severs.

### Chunking

Distributing objects is not enough;
Unlike common records, objects can significantly vary in size.
Large objects will require more storage and computing power,
easily causing resource imbalances between servers.

```d2
s1: Server 1 {
  "doc.md (30KB)" {
    class: file
  }
}
s2: Server 2 {
  "img.png (200MB)" {
    class: file
  }
}
```

Continue to the division process,
we divide objects into fixed-size storage units called **chunks**.
Apart from balancing resources,
**Chunking** is also beneficial for reading/writing in parallel to different servers.

#### Object Chunking

The most straightforward approach to chunk objects is slice them to equal parts.

For example, a service is configured the chunk size to `100MB`,
an object of `200MB` is cut to 2 chunks potentially placed on different servers.

```d2
f: "img.png (200MB)" {
  class: file
}
o: Object Storage {
  grid-rows: 1
  s1: Server 1 {
    "img.png.chunk_1 (100MB)" {
      class: file
    }
  }
  s2: Server 2 {
    "img.png.chunk_2 (100MB)" {
      class: file
    }
  }
}
f -> o.s1
f -> o.s2
```

But this approach is problematic in certain cases:

- Let's say we want configure the chunk size to small to ensure balancing resources.
But it results in a high number of chunks in different server,
accessing objects will require much computing power from the servers.

```d2
f: "img.png (500MB)" {
  class: file
}
o: Object Storage {
  grid-rows: 1
  s1: Server 1 {
    "img.png.chunk_1 (100MB)" {
      class: file
    }
  }
  s2: Server 2 {
    "img.png.chunk_2 (100MB)" {
      class: file
    }
  }
  s3: Server 3 {
    "img.png.chunk_3 (100MB)" {
      class: file
    }
  }
  s4: Server 4 {
    "img.png.chunk_4 (100MB)" {
      class: file
    }
  }
  s5: Server 5 {
    "img.png.chunk_5 (100MB)" {
      class: file
    }
  }
}
f -> o.s1
f -> o.s2
f -> o.s3
f -> o.s4
f -> o.s5
```

- However, if we adapt the chunk size wider, reducing the number of chunks.
This solution is easy to create imbalances,
as small objects tend to be ignored in the chunking process.

```d2
f: "img.png (200MB)" {
  class: file
}
d1: "doc1.md (10MB)" {
  class: file
}
d2: "doc2.md (20MB)" {
  class: file
}
d3: "doc3.md (30MB)" {
  class: file
}
o: Object Storage {
  grid-rows: 1
  s1: Server 1 {
    "img.png.chunk_1 (100MB)" {
      class: file
    }
    "doc1.md.chunk_1 (10MB)" {
      class: file
    }
    "doc2.md.chunk_1 (20MB)" {
      class: file
    }
    "doc3.md.chunk_1 (30MB)" {
      class: file
    }
  }
  s2: Server 2 {
    "img.png.chunk_2 (100MB)" {
      class: file
    }
  }
}
f -> o.s1
d1 -> o.s1
d2 -> o.s1
d3 -> o.s1
f -> o.s2
```

#### Chunk Packing

Let's approach in a more controlled angle,
instead of relying on cutting user objects,
we will define and pack them into system-customized chunks.
In other words, a chunk is a **fixed-size file** containing many objects.

For example, we'll configure the chunk size to `100MB`.
If an object is larger than the size,
it's placed in multiple chunks.

```d2
f1: "img1.png (50MB)" {
  class: file
}
f2: "img2.png (100MB)" {
  class: file
}
f3: "img3.png (50MB)" {
  class: file
}
o: Object Storage {
  grid-rows: 1
  s1: Server 1 {
    grid-columns: 2
    c1: "chunk_1" {
      grid-rows: 2
      grid-gap: 0
      "img1.png (50MB)"
      "img2.png.chunk_1 (50MB)"
    }
    c2: "chunk_2" {
      grid-rows: 2
      grid-gap: 0
      "img2.png.chunk_2 (50MB)"
      "img3.png (50MB)"
    }
  }
}
f -> o.s1
f -> o.s2
```

Programmatically, we can implement by only appending data to existing chunks until they're full.
This approach is preferred in building **Object Storage** solutions.
We also use this for below sections.

### Erasure Coding

Absolutely, we need to replicate chunks on multiple servers as a backup in case of data loss.

Let's say we have 2 chunks on 2 servers.
To prevent data loss, we replicate each chunk to the other server.
The storage cost is **at least two times** the real usage;
If one of the servers go down, we can still recover the chunks.

```d2
s1: Server 1 {
  c1: "chunk_1" {
    class: file
  }
  c2: "chunk_2_replica" {
    class: file
  }
}
s2: Server 2 {
  c1: "chunk_2" {
    class: file
  }
  c2: "chunk_1_replica" {
    class: file
  }
}

s1.c1 -> s2.c2 {
  style.animated: true
}
s2.c1 -> s1.c2 {
  style.animated: true
}
```

[**Erasure Coding (EC)**](https://en.wikipedia.org/wiki/Erasure_code) is a data integrity technique commonly used in distributed systems.

Let's say we have 2 data chunks.
We can create 1 **parity block** by using an encoding algorithm.
It's complex to see how the parity is created, we may image a simple math like `chunk_1 + chunk_2 = parity`.

```d2
direction: right
c1: chunk_1 {
  class: file
}
c2: chunk_2 {
  class: file
}
c3: parity {
  class: file
}
c1 -> c2: "+"
c2 -> c3: "="
```

Now, if any of the blocks is corrupted, e.g., `chunk_2`, we can recover the original data like `chunk_2 = parity - chunk_1`.

```d2
direction: right
c1: chunk_1 {
  class: file
}
c2: chunk_2 {
  class: generic-error
}
c3: parity {
  class: file
}
c2 -> c3: "="
c3 -> c1: "-"
```

**Erasure Coding** is wider technique,
stating that if we maintains `m` number of parities,
we can recover the original data as long as **no more than `m` blocks** are lost.

For example,
if we have 3 chunks and 2 parities,
let's place them on different servers,
the data is probably safe although **any** 2 servers are corrupted.

```d2
direction: right
s1: Server 1 {
  c1: chunk_1 {
    class: file
  }
}
s2: Server 2 { 
  c2: chunk_2 {
    class: generic-error
  }
}
s4: Server 4 { 
  c4: parity_1 {
    class: file
  }
}
s5: Server 5 { 
  p1: parity_2 {
    class: generic-error
  }
}
```

Let's compare with the method of fully replicating data.
To achieve the previous power,
for each chunk, we will replicate it to 2 other servers.
That means, the system is tolerant to any 2 server failures.
We observe that the storage cost for redundancy data is doubly higher than **Erasure Coding**,
to provide the same guarantee.

```d2
s1: Server 1 {
  c1: "chunk_1" {
    class: file
  }
}
s2: Server 2 {
  c1: "chunk_2" {
    class: file
  }
}
s3: Server 3 {
  c1: "chunk_1_replica" {
    class: file
  }
  c2: "chunk_2_replica" {
    class: file
  }
}
s4: Server 4 {
  c1: "chunk_1_replica" {
    class: file
  }
  c2: "chunk_2_replica" {
    class: file
  }
}
```

In general, **Erasure Coding** is usually implemented by maintaining the number of parities to a half number of original data blocks.
Based on that, we beneficially reduce the storage cost of data protection approximately by half.

However,
**Erasure Coding** make write operations and data recovery processes slower and consumer more resources,
as it needs to deal with encoding/decoding handlers.

### Metadata Server

Move to the final aspect of this section.
In the [Distributed Database](../distributed-database) topic,
we've routed a record to an owner server by leveraging its unique key.
But now, everything is much more sophisticated,
object keys (or paths) are not as important as they are bundled into system chunks.

Thus, to maintain an **Object Storage** cluster,
we need to also build an additional **Metadata Server** alongside the real storage servers.
This server needs to keep track where to retrieve an object through its key,
it can be a simple **Key-value store** with `key -> [(containing server, chunk, position in chunk, size in chunk)]`.

```d2
Object Storage {
  m: Metadata Server {
    content: |||yaml
    doc.md:
      name: doc.md
      size: 25MB
      storage:
      - Server_1:
        - chunk_1:
          - position: 123
          - size: 5MB
        - chunk_2:
          - position: 0
          - size: 10MB
      - Server_2:
        - chunk1:
          - position: 0
          - size: 10MB
    |||
  }
  s1: Storage Server 1 {
    c1: "chunk_1" {
      class: file
    }
    c2: "chunk_2" {
      class: file
    }
  }
  s2: Storage Server 2 {
    c1: "chunk_1" {
      class: file
    }
  }
  m -> s1
  m -> s2
}
```

## CDN (Content Delivery Network)

{{< term cdn >}} plays a big role in serving media data.
In short, a {{< term cdn >}} contains two parts

### 1. Caching Layer

{{< term cdn >}} stands as a [read-through](Caching-Patterns.md#cache-aside-cache) caching layer before a data source.

For example, a piece of data is initialized once,
and it's quickly retrieved through the {{< term cdn >}} later.

```d2
shape: sequence_diagram
c: Client {
  class: client
}
cdn: CDN {
  class: cdn
}
b: Server {
  class: server
}

c -> cdn: 1. Request data
cdn -> b: 2. Query the backend
cdn -> cdn: 3. Cache {
  style.bold: true
}
cdn -> c: 4. Respond
c -> cdn: 5. Request the data again
cdn -> c: 6. Respond the cached data immediately {
  style.bold: true
}
```

### 2. Backbone Network

Normally, data is exchanged through the public internet.
A long distance between endpoints creates many network hops and high latencies.

In the background, a {{< term cdn >}} is built on an **internal network** called **Backbone Network**.
This network consists of dedicated, high-speed fiber-optic links to significantly reduce the latency,
rather than relying on the public internet.
The network resides in many regions, transmitting data within it is extremely fast.

When a client connects to {{< term cdn >}}, it will be routed to the nearest server first,
further forwarded to the target server.

```d2
cdn: CDN {
  s1: Server at Southeast Asia {
    class: server
  }
  s2: Server at North America { 
    class: server
  } 
}
c: Client in Vietnam {
  class: client
}

c -> cdn.s1: 1. Routed to the nearest server
cdn.s1 -> cdn.s2: 2. Forward within the internal network {  
  style.bold: true
}
cdn.s2 -> cdn.s1
cdn.s1 -> cdn.s1: 3. Cache data
```

> A {{< term cdn >}} provider (AWS, Cloudflare) is in charge of its own network.
> To reduce cost, some big tech (Facebook, Netflix) build their proprietary networks instead

### Usages

There are two common mistakes of using {{< term cdn >}}:

1. Remember that {{< term cdn >}} means for caching data,
frequently updated data is not suitable, we should consider directly querying the data source.
If the source is far away, we can leverage the **Backbone Network** for replicating data to a closer position.

```d2
cdn: CDN {
  class: cdn
}
c: Client in Vietnam {
  class: client
}
o: Replica in Singapore {
  class: db
}
s: Source in US {
  class: db
}
s -- cdn {
  style.animated: true
}
cdn -> o: Replicate {
  style.animated: true
}  
c -> o: Query a nearby store 
```

2. If data is only distributed to nearby locations (e.g., within a country), the **Backbone Network** may be redundant.
We may develop a simple caching service for a cheaper solution.

### Edge Computing

**Edge Computing** is a distributed computing paradigm based on **Backbone Network**.
In short, when we want to transmit a lot of data,
we should **preprocess** it in the nearest server (called **Edge Server**) before sending the result to the target server (called **Origin Server**),
e.g., compression, aggregation...

```d2
cdn: CDN {
  s1: Server at Southeast Asia {
    class: server
  }
  s2: Server at North America { 
    class: server
  } 
}
c: Client in Vietnam {
  class: client
}
s: Server in US {
  class: server
}
c -> cdn.s1: 1. Routed to the nearest server
cdn.s1 -> cdn.s1: 2. Preprocess data {
  style.bold: true
}
cdn.s1 -> cdn.s2: 3. Forwards the preprocessed data
cdn.s2 -> s
```

**Edge Computing** significantly reduces traffic bandwidth and network latency because data is preprocessed and optimized.
