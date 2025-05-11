---
title: Media Storage
weight: 40
---

Media data (videos, images, documents, etc.) consists of unstructured binary content.
Storing such data in a traditional database is not ideal.
Instead, it should be read and stored directly via a standard file system interface.

Media data is especially critical for client-facing services and often serves as the backbone of many applications, such as Netflix and YouTube.
In this section, we will explore common approaches to managing media data effectively.

## File Storage

**File Storage** refers to a system that exposes storage as a live file system,
allowing independent scaling of compute and storage resources.

It is well-suited for **high-performance** and low-latency applications, enabling them to access complete files directly over the network.
Moreover, multiple clients (either services or end-users) can share the same storage backend.

For example, two services share the same file storage system.

```d2
s1: Service 1 {
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
s2: Service 2 {
  class: server
}
s1 <-> d1
s2 <-> d1
```

To improve availability and performance, we can add replicas.

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

However, this setup resembles the [Master-Slave model](../distributed-database/master-slave-architecture), where the primary server becomes a single point of failure, reducing overall service availability.

## Object Storage

This is a new technology but gaining tremendous momentum recently.
The best advantage of **Object Storage** is its distributed nature.
File is now called **Object**, has no relationship and is autonomously distributed to many servers,
resulting in a highly available and fault-tolerant system.

Let's see how an **Object Storage** is implemented in attempts:

## Object Storage

Although a relatively new technology, **Object Storage** has been rapidly gaining popularity.

The key advantage of Object Storage lies in its **distributed architecture**. Files, now referred to as **objects**, are independent and autonomously distributed across multiple servers. This results in a highly available and fault-tolerant system.

Let’s explore how Object Storage is implemented in practice.

### Object Distribution

As discussed in the [Peer-to-peer architecture](../distributed-database/peer-to-peer-architecture), we aim to distribute objects across multiple servers. This allows concurrent writes and improves scalability. Additionally, each server maintains replicas to increase fault tolerance.

For example, two servers continuously replicate data to one another.

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

However, this distributed model complicates how we interact with objects.
Traditionally, we access files using a hierarchical path structure like `Team/Docs/doc.md`.
Without a centralized file system, there’s no inherent relationship between objects (e.g., directories or siblings).

To simulate the familiar file system structure, **Object Storage** systems often include the **full path** in the object’s key, for example:

```d2
s1: Server 1 {
  "Users/Image/img.png": {
    class: file
  }
  "Team/Docs/doc.md": {
    class: file
  }
}
s2: Server 2 {
  "Team/Docs/README.md": {
    class: file
  }
}
```

This is merely a naming convention.
Operations like _listing files in a folder_ still require scanning across multiple servers.

### Chunking

Simply distributing objects isn’t sufficient. Unlike typical database records, object sizes can vary dramatically. Large objects demand more storage and processing resources, which can lead to imbalances across servers.

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

To address this, we divide objects into fixed-size units called **chunks**.
Chunking not only balances resource usage but also allows parallel read/write operations across servers.

#### Object Chunking

The most straightforward chunking strategy is to divide objects into equal-sized parts.

For instance, if the configured chunk size is `100MB`,
a `200MB` object will be split into two chunks, which can be stored on different servers:

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

However, this approach has some trade-offs:

- Using small chunk sizes improves load balancing but results in too many chunks spread across servers.
  Retrieving an object then requires significant coordination and computing effort.
  For example, an object is distributed across five servers, requiring queries to all of them for retrieval.

```d2
f: "img.png (500MB)" {
  class: file
}
o: Object Storage {
  grid-rows: 1
  s1: Server 1 { "img.png.chunk_1 (100MB)" { class: file } }
  s2: Server 2 { "img.png.chunk_2 (100MB)" { class: file } }
  s3: Server 3 { "img.png.chunk_3 (100MB)" { class: file } }
  s4: Server 4 { "img.png.chunk_4 (100MB)" { class: file } }
  s5: Server 5 { "img.png.chunk_5 (100MB)" { class: file } }
}
f -> o.s1
f -> o.s2
f -> o.s3
f -> o.s4
f -> o.s5
```

- On the other hand, using large chunk sizes reduces the number of chunks
  but can create resource imbalances—small objects may be underutilized or ignored.
  For example, objects smaller than the chunk size are inefficiently stored on the same server.

```d2
f: "img.png (200MB)" { class: file }
d1: "doc1.md (10MB)" { class: file }
d2: "doc2.md (20MB)" { class: file }
d3: "doc3.md (30MB)" { class: file }
d4: "doc4.md (40MB)" { class: file }
o: Object Storage (chunk size = 100MB) {
  grid-rows: 1
  s1: Server 1 {
    "img.png.chunk_1 (100MB)" { class: file }
    "doc1.md.chunk_1 (10MB)" { class: file }
    "doc2.md.chunk_1 (20MB)" { class: file }
    "doc3.md.chunk_1 (30MB)" { class: file }
    "doc4.md.chunk_1 (40MB)" { class: file }
  }
  s2: Server 2 {
    "img.png.chunk_2 (100MB)" { class: file }
  }
}
f -> o.s1
d1 -> o.s1
d2 -> o.s1
d3 -> o.s1
d4 -> o.s1
f -> o.s2
```

#### Chunk Packing

To achieve better control, instead of slicing user objects arbitrarily,
we define **fixed-size system chunks** and **pack multiple objects** into each chunk.

In this model, a chunk is a system-level file containing multiple objects.
If an object exceeds the chunk size, it’s split across multiple chunks.

For instance, three objects are packed into two chunks, which are stored on two separate servers.

```d2
f1: "img1.png (50MB)" { class: file }
f2: "img2.png (100MB)" { class: file }
f3: "img3.png (50MB)" { class: file }
o: Object Storage {
  grid-rows: 1
  s1: Server 1 {
    c1: "chunk_1" {
      grid-rows: 2
      grid-gap: 0
      "img1.png (50MB)"
      "img2.png.chunk_1 (50MB)"
    }
  }
  s2: Server 2 {
    c1: "chunk_1" {
      grid-rows: 2
      grid-gap: 0
      "img2.png.chunk_2 (50MB)"
      "img3.png (50MB)"
    }
  }
}
f1 -> o.s1
f2 -> o.s1
f2 -> o.s2
f3 -> o.s2
```

In practice, chunks are filled by appending object data until the chunk reaches capacity.
This method is common in modern **Object Storage** systems and will be used in the following sections.

### Erasure Coding

To prevent data loss, we must replicate chunks across servers.
A simple way is to duplicate each chunk to another server:

```d2
s1: Server 1 {
  c1: "chunk_1" { class: file }
  c2: "chunk_2_replica" { class: file }
}
s2: Server 2 {
  c1: "chunk_2" { class: file }
  c2: "chunk_1_replica" { class: file }
}
s1.c1 -> s2.c2 { style.animated: true }
s2.c1 -> s1.c2 { style.animated: true }
```

However, this basic replication results in **2x storage overhead**.

#### Parity Blocks

[**Erasure Coding (EC)**](https://en.wikipedia.org/wiki/Erasure_code) offers a more storage-efficient alternative. For example, with 2 data chunks, we can generate 1 **parity block** using an encoding function. Think of it conceptually as: `parity = chunk_1 + chunk_2`

```d2
direction: right
c1: chunk_1 { class: file }
c2: chunk_2 { class: file }
c3: parity { class: file }
c1 -> c2: "+"
c2 -> c3: "="
```

If one chunk is lost (e.g., `chunk_2`), it can be recovered as: `chunk_2 = parity - chunk_1`

```d2
direction: right
c1: chunk_1 { class: file }
c2: chunk_2 { class: generic-error }
c3: parity { class: file }
c2 -> c3: "="
c3 -> c1: "-"
```

With `m` parity blocks, we can tolerate loss of up to `m` chunks.
For example, with 3 data chunks and 2 parities, data remains safe even if **any two servers** fail:

```d2
direction: right
s1: Server 1 { c1: chunk_1 { class: file } }
s2: Server 2 { c2: chunk_2 { class: generic-error } }
s4: Server 4 { c4: parity_1 { class: file } }
s5: Server 5 { p1: parity_2 { class: generic-error } }
```

In comparison, using full replication for the same level of fault tolerance would require 3 total copies per chunk:

```d2
s1: Server 1 { c1: "chunk_1" { class: file } }
s2: Server 2 { c1: "chunk_2" { class: file } }
s3: Server 3 {
  c1: "chunk_1_replica" { class: file }
  c2: "chunk_2_replica" { class: file }
}
s4: Server 4 {
  c1: "chunk_1_replica" { class: file }
  c2: "chunk_2_replica" { class: file }
}
```

**Erasure Coding** typically uses a parity-to-data ratio of 1:2,
reducing storage overhead by roughly half compared to full replication.

However, **Erasure Coding** increases **write latency and resource usage**,
due to the encoding and decoding processes involved.

### Metadata Server

Let's move to the final aspect of this section.
In the [Distributed Database](../distributed-database) topic,
we routed a record to its owning server using a unique key.
However, in Object Storage, the situation is more complex—object keys (or paths) are no longer central, as objects are bundled into system-managed chunks.

To manage an **Object Storage** cluster effectively,
we must introduce a dedicated **Metadata Server** in addition to the actual storage servers.
This server is responsible for tracking where each object resides based on its key.
It can be implemented as a simple **Key-value store**, mapping keys to metadata like:
`key -> [(server, chunk, position within chunk, size within chunk)]`.

For example, a file is mapped on the **Metadata Server** to its actual storage locations.

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

{{< term cdn >}} plays a crucial role in delivering media content efficiently.
In essence, a {{< term cdn >}} is composed of two main components:

### 1. Caching Layer

A {{< term cdn >}} functions as a [read-through caching layer](Caching-Patterns.md#cache-aside-cache) positioned in front of the data source.

For example, once a piece of data is initialized, it can be quickly retrieved from the CDN in subsequent requests:

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
cdn -> c: 6. Respond with cached data immediately {
  style.bold: true
}
```

### 2. Backbone Network

Typically, data is transferred over the public internet.
However, long distances between endpoints result in many network hops and increased latency.

Behind the scenes, a {{< term cdn >}} is built on an **internal high-speed network**,
known as the **Backbone Network**.
This network consists of dedicated fiber-optic links across regions,
offering significantly faster transmission than the public internet.

When a client connects to the CDN, their request is first routed to the nearest CDN server,
which may then forward it internally to the target server:

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
cdn.s1 -> cdn.s2: 2. Forward through the internal network {
  style.bold: true
}
cdn.s2 -> cdn.s1
cdn.s1 -> cdn.s1: 3. Cache data
```

> Major CDN providers like AWS and Cloudflare operate their own backbone networks.
> Some large tech companies (e.g., Facebook, Netflix) even build
> proprietary networks to optimize performance and reduce costs.

### Usages

There are two common misuses of CDNs:

1. **Caching frequently updated data**:
   CDNs are designed for caching static or infrequently changing content.
   For dynamic data that updates often, it's better to query the source directly.
   If the source is geographically distant, consider using the **Backbone Network** to replicate data to a nearby location:

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
c -> o: Query the nearby replica
```

2. **Overengineering local deployments**:
   If content is only accessed within a limited geographic area (e.g., within one country), using a full CDN and Backbone Network might be overkill.
   A simpler, localized caching system may offer a more cost-effective solution.

### Edge Computing

**Edge Computing** is a distributed computing model that builds upon the CDN's **Backbone Network**.
The key idea is to **preprocess** data at the closest possible server (called the **Edge Server**)
before sending it to the main server (the **Origin Server**).

This preprocessing can include operations like compression, filtering, or aggregation.

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
cdn.s1 -> cdn.s2: 3. Forward preprocessed data
cdn.s2 -> s
```

**Edge Computing** dramatically reduces both bandwidth usage and latency by optimizing data closer to the client before transmission.
