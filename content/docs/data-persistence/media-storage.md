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

**File Storage** refers to a solution that builds the storage as a live server,
helping scale a machine and its storage unit independently.

It's perfect for **high performant** and low-latency applications while
they can work with files directly through network.
Additionally, multiple clients (service or enduser) can rely on the same storage server.

```d2
%d2-import%
s1: Server 1 {
  class: server
}
d1: File Storage {
  class: hd
}
s2: Server 2 {
  class: server
}
s1 -> d1: Mount
s2 -> d1: Mount
```

Of course,
we should build replicas to enhance availability and performance.

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
 
## Object Storage
This is a new technology but gaining tremendous momentum recently. 
`Object Storage` organizes data as separate `objects` rather than files or blocks. 

`Object Storage` is **flat**, that means there is no relationship (directory, sibling...) between them.
To retrieve or create an object, we need to provide its **unique key**.
People tend to mimic the file system notation for naming keys e.g. `Team/Shared/File1.md`, `Users/Avatar/User1.png`...
```d2
Object Storage {
    o1: "Team/Shared/File1.md" {
        grid-gap: 0
        grid-columns: 1
        Data
    }
    o3: "Users/Avatar/User1.png" {
        grid-gap: 0
        grid-columns: 1
        Data
    }
}
```


### Bucket
Although we don't have the concept of directory, objects sharing the same prefix can be **logically** grouped.
For instance, `Team/README.md`, `Team/Rules.md` both exist in the `Team/` prefix. 
The `flat` structure is perfect for retrieving single objects. 
But it's bad for directory-like accesses (e.g., getting all objects of a prefix), 
because we’re building independent objects, grouping them requires scan operations.

A `bucket` is a **unique** and **isolated** container of objects inheriting the bucket settings.
With the concept of `bucket` and `prefix`, `Object Storage` looks like this
- A bucket is similar to a drive
- A prefix is similar to a directory
```d2
Object Storage {
  Developer Bucket {
    Team Prefix {
      grid-columns: 1
      "Team/README.md"
      "Team/Rules.md"
    }
  }
  Manager Bucket {
    Members Prefix {
      grid-columns: 1
      "Members/John/cv.pdf"
      "Members/Cait/avatar.png"
    }
    Documents Prefix {
      "Documents/stategy.txt"
    }
  }
}
```

#### Write Once Read Many - WORM
`Object Storage` only allows creating and deleting objects, an object is immutable once it's created. 
This model is called `WORM (Write Once Read Many)`.
Object is really *updated* by creating a new **version**. 
An object potentially has many versions, this approach is safe in terms of data retention.


### Object Store Cluster
Object storage is built for **availability and durability**, often designed for distributed systems, 
rather than supporting high-performant workloads. 

It divides the system into a metadata service and a storage service.
The metadata server hints actual data on the storage service.
```d2
%d2-import%
o: Object storage {
    m: Metadata service {
        class: db
    }
    s: Storage service {
        class: db
    }
}
```

An object is not mapped directly to a file.
Instead, it's divided into smaller parts (called `chunks`), distributed across multiple storage nodes.
Clients first ask the metadata service, then gradually assemble all the chunks from associated storage nodes.
```d2
m: Metadata service {
  o1: "Object_1 (metadata)" {
    shape: sql_table
    size: 100MB
  }
  o2: "Object_2 (metadata)" {
    shape: sql_table
    size: 75MB
  }
}
s: Storage service {
  n1: Node 1 {
    shape: circle
  }
  n2: Node 2 {
    shape: circle
  }
  n3: Node 3 {
    shape: circle
  }
}
m.o1 -> s.n1
m.o1 -> s.n2
m.o2 -> s.n2
m.o2 -> s.n3
```

More detailed, a chunk is not also a file.
A file takes up a data block (typically 4KB). 
If there are a lot of chunks less than the smallest size, 
many unused gaps will be left inside blocks. 
Instead, many chunks are bundled into a file to make use of system space.
```d2
grid-gap: 0
grid-columns: 1
f: File 1 {
    grid-gap: 0
    grid-rows: 1
    "Object 1 - Chunk 1"
    "Object 2 - Chunk 1"
}
f: File 2 {
    grid-gap: 0
    grid-rows: 1
    "Object 2 - Chunk 2"
    "Object 3 - Chunk 1"
}
```


The best advantage of `Object Storage` is its distributed nature.
Data has no relationship and is autonomously distributed to many nodes,
resulting in a highly available system. 
Furthermore, this helps load balancing between nodes, 
they can currently serve multipart of an object.

## Content Delivery Network - CDN
CDN plays a big role in serving media data. 
In short, a CDN contains two parts

### Caching Layer
CDN stands as a [read-through](Caching-Patterns.md#cache-aside-cache) caching layer before a data source

For example, a piece of data is initialized once, and it's efficiently retrieved through CDN later
```d2
%d2-import%
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

c -> cdn: 1. Requests a piece of data
cdn -> b: 2. Queries the backend
cdn -> c: 3. Caches the data before responding {
  style.bold: true
}
c -> cdn: 1. Requests the data again
cdn -> c: 2. Responds the cached data immediately {
  style.bold: true
}
```

### Backbone Network
Normally, data is exchanged through the public internet. 
A long distance between endpoints creates many network hops and high latencies.

In the background, a CDN is built on an **internal network** called `backbone` network.
This network consists of dedicated, high-speed fiber-optic links to significantly reduce the latency, 
rather than relying on the public internet. 
The backbone network resides in many regions, transmitting data within it is extremely fast.
When a client connects to CDN, he will be routed to the nearest server first, 
further forwarded to the target server. 
```d2
%d2-import%
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
cdn.s1 -> cdn.s2: 2. Forwards within the internal network {  
  style.bold: true
}
cdn.s2 -> cdn.s1
cdn.s1 -> cdn.s1: 3. Cache data
```

> A CDN provider (AWS, Cloudflare) is in charge of its own network.
> To reduce cost, some big tech (Facebook, Netflix) build their proprietary networks instead

### Usage
There are two common mistakes of using CDN
1. Remember that CDN means for caching data, frequently updated data is not suitable, we should consider directly querying the data source. 
If the source is far away, we can leverage the backbone network for efficiently replicating data to a closer position
```d2
%d2-import%
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
cdn -> o: Replicates {
  style.animated: true
}  
c -> o: Queries a nearby store 
```
2. If data is only distributed to nearby locations (e.g., within a country), the backbone network may be redundant.
We may develop a simple caching layer for a cheaper solution  

### Edge Computing
`Edge Computing` is a distributed computing paradigm based on CDN. 
In short, when we want to transfer a lot of data, 
we **preprocess** it in the nearest server (of a CDN) and send the result to the target server,
e.g., compressing, aggregating...
> `Edge` means edges of a network, near the sources of data

```d2
%d2-import%
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

`Edge Computing` reduces traffic bandwidth and network latency because data is preprocessed and optimized.