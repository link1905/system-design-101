---
title: Distributed Job Patterns
weight: 40
---

Unlike long-term services,
**Job** is a short-term process dedicating to a specific workload.
A job can be small as *compressing an image*,
sometimes it can be expensive with high computing capability as *processing a large dataset*.

In a distributed environment,
we will distribute jobs on multiple machines to make the most of the system resources.
We will take a look at some popular patterns to do that.

## Data Processing

Most of the time, jobs are used to process data.
There are two patterns of processing data: **Streaming** and **Batching**.

### Streaming

Needless to say about this pattern.
Jobs are executed with the help of an event stream,
we need to maintain some worker services continuously pulling and processing events.

This pattern brings the near-realtime adaption,
events are fired and then processed immediately.

```d2
grid-columns: 3
e1: Event 1 {
    class: event
    height: 100
}
e2: Event 2 {
    class: event
    height: 100
}
w: Worker { 
    class: process
    height: 200
}
r1: Result 1 {
    class: event
    height: 100
}
r2: Result 2 {
    class: event
    height: 100
}

e1 -> r1: Process
e2 -> r2: Process
```

**Streaming Processing** comes with the cost of maintaining fault-tolerant and highly available workers,
increasing operational overheads.
Furthermore,
achieving the fast adaption often involves costly allocating resources to handle peak loads and avoid bottlenecks.

### Batching

Another style of processing data is batching.
We stack up data to a threshold before actually handling it, e.g. after 100MB of data or every 12h.

```d2
grid-columns: 3
grid-gap: 100
e1: Event 1 {
    class: event
    height: 100
}
e2: Event 2 {
    class: event
    height: 100
}
w: Worker Service { 
    class: process
    height: 200
}
r1: Result {
    class: event
    height: 200
}
e1 -> w
e2 -> w
w -> r1: Process
```

In general, **Batch Processing** is easier and cost-effective:

- Batch jobs often process a predictable amount of data,
making resource allocation and performance tuning straightforward.
- In terms of programming, processing a lot of data at once is typically beneficial,
when we can apply efficient algorithms on large datasets.

In practice, the cheapness is so impressive.
When your system doesn't require near-realtime power,
**Batching** should be seriously taken into account.

### Kappa Architecture

However, they're not two sides of the same coin.
**Kappa Architecture** combines both of them in a single event stream.

As usual, streaming workers pull events and process them continuously.
Besides, due to event durability, batching services periodically compute data by replaying from the stream.

For example, an e-commerce system with:

- `Order Service`: adapts to order creations instantly.
- `Dashboard Service`: computes analytical metrics from the order book **daily**.

```d2
direction: right
o: Order Service {
    class: server
}
d: Dashboard Service {
    class: server
}
e: Event Stream {
    class: mq
}
o <- e: Pull instantly
d <- e: Replay orders daily {
    style.stroke-dash: 3
}
```

## Directed Acyclic Graph (DAG)

**Directed Acyclic Graph (DAG)** is a graph structure in which

- **Directed**: The edges have only one direction.
- **Acyclic**: The edges form no cycles.

**DAG** can be used to represent task scheduling, dividing the system into smaller workers.

For example, an e-commerce website allows sellers to upload product information (metadata, video, images...).
The system can branch into different flows to make use of the parallel power,
a product record is actually created after the assets are processed.

```d2
direction: right
p: Uploaded Data {
    class: storage
}
v: Video Encoder {  
    class: process
}
i: Image Encoder {
    class: process
}
ps: Product Creator {
    class: process    
}
p -> v: Encode video
p -> i: Encode images
v -> ps
i -> ps
p -> ps
ps -> ps: Create the product record with the encoded assets
```

### Event Buffering

To make use of parallelism, we may choreograph the workflow with collaborative events.
When the steps are distinct, we may build them as separate workers.
This makes the system more flexible, workers only focus on its responsibility, and can be scaled independently.

```d2
direction: right
e: Event Stream {
    class: mq
}
v: Video Encoder {  
    class: process
}
i: Image Encoder {
    class: process
}
ps: Product Creator {
    class: process    
}
e -> v: ProductUploaded
v -> e: VideoEncoded
e -> i: ProductUploaded
i -> e: ImageEncoded
e -> ps: ProductUploaded + VideoEncoded + ImageEncoded
ps -> ps: Create the product record referencing the encoded assets
```

How do we suppose to capture three events together?
Buffering events locally to wait for their joined arrival is an effective solution,
the `ProduceCreator` stands still before getting the necessary events.

```d2
direction: right
e: Event Stream {
    class: mq
}
ps: Product Creator {
    class: process    
}
j: Join {
    class: join
    width: 50
}
e -> j: ProductUploaded
e -> j: VideoEncoded
e -> j: ImageEncoded
j -> ps
```

### Claim-Check Pattern

Querying external datasets is not a rare requirement of jobs.

Let's say we have a job which sends emails to users,
its attended events only contain `userId`.
Consequently, it needs to ask the `UserService` about user emails.

```d2
direction: right
e: Event Stream {
    class: mq
}
m: Mail Worker {
    class: process
}
u: User Service {
    class: server
}
event: |||yaml
MailEvent:
    userId: user123
    content: Hello
|||
e -> event
event -> m
m <- u: Get user email
```

This is called **Claim-Check** pattern

- **Claim**: build lightweight events with minimal metadata (e.g. `userId`, `orderId`...).
- **Check**: cause consumers to retrieve full payloads from other services.

The **Claim-Check** pattern ensures statelessness and strong consistency,
helping build lightweight workers.
However, this pattern doesn’t only make the worker slower but also creates a dependency on the data providers,
reducing their performance and availability,
especially problematic when a worker depends on many providers.

### Stateful Worker

We've discussed this in the [Event Sourcing]({{< ref "event-sourcing" >}}) topic.
Worker services can depend on a shared event stream to build their own local stores instead.

For example,
`Mail Worker` builds a local of user emails by capturing `UserInformationUpdated` events.

```d2
u: UserInformationUpdated {
    class: msg
}
e: User Stream {
    class: mq
}
m: Mail Worker {
    m: Local mail store {
        class: db
    }
}
u -> e
e -> m.m: Build {
    style.stroke-dash: 3
}
```

Of course, stateful services like this require more resources to initialize and maintain.
Additionally, the asynchronous model only supports **eventual consistency**,
which can an issue occasionally,
e.g. a user has changed email, but the `Mail Worker` slowly captures this event,
leading to send to the old mailbox unexpectedly.  

## MapReduce

**MapReduce** is a paradigm designed to process large data sets **in parallel**
across a distributed cluster.

{{< callout type="info" >}}
This concept was originally introduced by `Google` and is widely used in big data frameworks.
{{< /callout >}}

**MapReduce** recommends separating a job into two sequential steps:

1. **Map**: this part is used to initially process input data to provide intermediary **key-value pairs**.
2. **Reduce**: this part aggregates and **reduces** the volume of the previous step based on unique keys.

Let's make it clear with an example.
We have a large set of web server logs, and we want to count how many times each page was accessed.

- **Map**: First, the dataset is sliced into chunks, each chunk is assigned to a mapping worker.
The mappers produce a list of key-value pairs by grouping their page.

```d2
s: Source log {
    data: |||yaml
    /index.html
    /about.html
    /index.html
    /contact.html
    /index.html
    /about.html
    |||
}
m: "1. Map" {
    w1: Mapper 1 {
        data: |||yaml
        /index.html
        /about.html
        /index.html
        |||
        output: |||yaml
        (/index.html, 2)
        (/about.html, 1)
        |||
        data -> output: Map
    }
    w2: Mapper 2 {
        data: |||yaml
        /contact.html
        /index.html
        /about.html
        |||
        output: |||yaml
        (/index.html, 1)
        (/about.html, 1)
        (/contract.html, 1)
        |||
        data -> output: Map
    }
}
s -> m.w1.data
s -> m.w2.data
```

- **Reduce**: This step initiates after the mappers complete.
Reducing workers pull data from mappers,
selecting with a certain **hash function** (like [Data Ownership]({{< ref "distributed-database#data-ownership" >}}))
to let data having the same key go to the same reducer.
The final result is calculated by aggregating the intermediary pairs.

```d2
s: Source log {
    data: |||yaml
    /index.html
    /about.html
    /index.html
    /contact.html
    /index.html
    /about.html
    |||
}
m: "1. Map" {
    w1: Mapper 1 {
        data: |||yaml
        /index.html
        /about.html
        /index.html
        |||
        output: |||yaml
        (/index.html, 2)
        (/about.html, 1)
        |||
        data -> output: Compute
    }
    w2: Mapper 2 {
        data: |||yaml
        /contact.html
        /index.html
        /about.html
        |||
        output: |||yaml
        (/index.html, 1)
        (/about.html, 1)
        (/contract.html, 1)
        |||
        data -> output: Compute
    }
}
r: "2. Reduce" {
    w1: Reducer 1 {
        data: |||yaml
        (/index.html, 2)
        (/index.html, 1)
        |||
        output: |||yaml
        (/index.html, 3)
        |||
        data -> output: Aggregate
    }
    w2: Reducer 2 {
        data: |||yaml
        (/about.html, 1)
        (/about.html, 1)
        (/contract.html, 1)
        |||
        output: |||yaml
        (/about.html, 2)
        (/contract.html, 1)
        |||
        data -> output: Aggregate
    }
}
s -> m.w1.data
s -> m.w2.data
m.w1.output -> r.w1.data
m.w2.output -> r.w1.data
m.w2.output -> r.w2.data
m.w1.output -> r.w2.data
```

We see that the process focuses on data grouped by unique keys,
making data of different keys can be distributed and processed on different workers in parallel.

A critical problem of **MapReduce** is significantly using network bandwidth and cloning data.
You may carefully consider the applicability of **MapReduce**,
small datasets are not good targets for highlighting the power of parallelism,
the performance is even downgraded sometimes because of moving data a lot.
