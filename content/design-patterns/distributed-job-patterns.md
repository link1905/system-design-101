---
title: Distributed Job Patterns
weight: 40
---

In contrast to long-term services, a job is a **short-term** process dedicated to a specific workload.

Jobs can range from simple tasks, like *compressing an image*, to computationally intensive operations,
such as *processing a large dataset*, which demand significant computing power.

In distributed environments,
jobs are often distributed across multiple machines to optimize the use of system resources.
Let's explore some common patterns for achieving this.

## Data Processing

Jobs are predominantly used for data processing.
There are two primary patterns for this: **Streaming** and **Batching**.

### Streaming

The streaming pattern is straightforward.
Jobs are executed with the assistance of an event stream, requiring the continuous maintenance of worker services that consume and process events.
This pattern enables **near-realtime adaptation**, as events are processed almost immediately after they are generated.

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

However, **Streaming Processing** incurs the cost of maintaining fault-tolerant and highly available workers,
which increases operational overhead.
Furthermore, achieving rapid adaptation often necessitates costly resource allocation to manage peak loads and prevent bottlenecks.

### Batching

An alternative data processing style is batching.
In this approach, data accumulates to a predefined threshold (e.g., 100MB of data or every 12 hours) before it is processed.

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

Generally, **Batch Processing** is simpler and more cost-effective for several reasons:

- Batch jobs typically handle a predictable volume of data, simplifying resource allocation and performance tuning.
- From a programming perspective, processing large amounts of data at once is often advantageous,
allowing for the application of efficient algorithms to large datasets.

In practice, the cost-effectiveness of batching can be quite significant.
When a system does not require near-realtime capabilities, **Batching** should be a serious consideration.

### Kappa Architecture

The **Kappa Architecture** brings these two paradigms together by utilizing a single event stream for all processing needs.

- **Streaming workers** continuously consume and process incoming events in real time, enabling immediate reactions and up-to-date system states.
- Meanwhile, **batching services** periodically replay events from the same stream, allowing for comprehensive, scheduled data computations.

For example, in an e-commerce system:

- The `Recommendation Service` instantly processes user interactions to personalize the experience.
- The `Dashboard Service` periodically (e.g., daily) update analytical metrics based on the accumulated data.

```d2
direction: right

o: Recommendation Service {
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
  style.animated: true
}
```

## Directed Acyclic Graph (DAG)

A **Directed Acyclic Graph (DAG)** is a graph structure characterized by:

- **Directed**: Edges have a single direction.
- **Acyclic**: Edges do not form any cycles.

**DAGs** can represent task scheduling, dividing a system into smaller, manageable workers.

For example, an e-commerce website might allow sellers to upload product information (metadata, videos, images, etc.).
The system can then branch into different processing flows to leverage parallel processing capabilities,
with a product record being created only after all associated assets are processed.

```d2
direction: right
p: Uploaded Data {
    class: resource
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
p -> v: Encode
p -> i: Encode
v -> ps
i -> ps
p -> ps: Create with the encoded assets
```

### Event Buffering

To maximize parallelism, workflows can be choreographed using collaborative events.
When steps in a workflow are distinct, they can be implemented as separate workers.
This enhances system flexibility, as each worker focuses solely on its responsibility and can be scaled independently.

```d2
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
e -> v: ProductUploadedEvent
v -> e: VideoEncodedEvent
e -> i: ProductUploadedEvent
i -> e: ImageEncodedEvent
e -> ps: ProductUploadedEvent + VideoEncodedEvent + ImageEncodedEvent
```

How can we coordinate the capture of three distinct events? Buffering events locally until all joined events arrive is an effective solution.
The `Product Creator`, for instance, would pause the execution until it receives all necessary events.

```d2
direction: right
e: Event Stream {
    class: mq
}
ps: Product Creator {
    class: process    
}
j: Join By ProductId
e -> j: ProductUploaded
e -> j: VideoEncoded
e -> j: ImageEncoded
j -> ps
```

### Stateless Worker

Jobs frequently need to query external datasets.

Consider a job that sends emails to users. The events triggering this job might only contain a `userId`.
Consequently, the job must query a `UserService` to retrieve the user's email address.

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

This statelessness promotes strong consistency,
contributing to the development of lightweight workers.

#### Claim-Check Pattern

Dealing with large-sized events presents a critical challenge.
Overwhelming the event stream with such events can lead to significantly degraded performance.

The **Claim-Check** pattern offers a solution by recommending the separation of large events into two distinct parts:

- **Claim-Check Token**: A unique token, representing the event, is transmitted through the event stream.
- **Payload**: The actual (heavy) data associated with the token is stored in a shared data store.

Upon receiving a claim-check token from the stream,
consumers become responsible for retrieving the corresponding full payload from the shared store using that token.

```d2
grid-rows: 1
horizontal-gap: 200
m: Event {
    class: msg
}
s {
    class: none
    grid-columns: 1
    e: Event Stream {
        class: mq
    }
    s: Payload Store {
        class: db
    }
}
c: Consumer {
    class: process
}
m -> s.s: 1. Store payload
m -> s.e: 2. Claim-check token
c <- s.e: 3. Pull the token
c <- s.s: 4. Get the payload
```

It's important to acknowledge that this pattern introduces management overhead for the payload store.
Therefore, consider implementing the **Claim-Check** pattern primarily when our events are so large that
they genuinely cause performance issues or instability.

### Stateful Worker

Stateless workers excel in resource efficiency and maintaining consistency by offloading state management to centralized data sources.
However, relying solely on external data providers can slow down workers, increase latency, and create critical dependencies, affecting both performance and availability.
This issue becomes even more pronounced when a worker needs to interact with multiple data providers.

As discussed in the [Event Sourcing]({{< ref "event-sourcing" >}}) topic,
worker services can instead depend on a shared event stream to build their own local data stores.

For instance, a `Mail Worker` can maintain a local store of user email addresses by listening to and capturing `UserInformationUpdated` events.
With this local store in place, the worker can swiftly retrieve the required user emails, enabling it to efficiently complete tasks without repeatedly relying on external data sources.

```d2
direction: right
us: User Service {
    class: server
}
e: User Stream {
    class: mq
}
u: UserInformationUpdated {
    class: msg
}
m: Mail Worker {
    m: Local mail store {
        class: db
    }
}
us -> e {
    style.animated: true
}
e -> u {
    style.animated: true
}
u -> m.m: Build {
    style.animated: true
}
```

Naturally, stateful services like this require more resources for initialization and maintenance.
Additionally, the asynchronous model only supports **eventual consistency**.
This can occasionally lead to issues, such as when a user changes their email,
but the `Mail Worker` processes this update slowly, resulting in an email being sent to the old mailbox.

## MapReduce

**MapReduce** is a programming paradigm designed for processing large datasets **in parallel** across a distributed cluster.

{{< callout type="info" >}}
This concept was originally introduced by `Google` and is widely used in big data frameworks.
{{< /callout >}}

**MapReduce** advocates for separating a job into two sequential steps:

1. **Map**: This stage initially processes input data to produce intermediate **key-value pairs**.
2. **Reduce**: This stage aggregates and **reduces** the volume of data from the previous step, based on unique keys.

Let's clarify with an example.
Suppose we have a large set of web server logs and want to count the access frequency for each page.

{{% steps %}}

### Map

First, the dataset is divided into small chunks, and each chunk is assigned to a mapping worker.
The mappers generate a list of key-value pairs by grouping their respective page accesses.

```d2
s: Source log {
    data: |||yaml
    # 1st chunk
    /index.html
    /about.html
    /index.html

    # 2nd chunk
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

### Reduce

This step begins after the mappers have completed.
Reducing workers pull data from the mappers,
using a specific **hash function** (akin to [data sharding]({{< ref "distributed-database#data-ownership" >}})) to ensure that data with the same key is processed by the same reducer.
The final result is calculated by aggregating these intermediate pairs.

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
    w1: Reducing Worker 1 {
        data: |||yaml
        (/index.html, 2)
        (/index.html, 1)
        |||
        output: |||yaml
        (/index.html, 3)
        |||
        data -> output: Aggregate
    }
    w2: Reducing Worker 2 {
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

{{% /steps %}}

This process focuses on data grouped by **unique keys**,
enabling data associated with different keys to be distributed and processed in parallel on different workers.

A critical challenge with **MapReduce** is its significant use of network bandwidth and data duplication.
Careful consideration should be given to the applicability of **MapReduce**.
Small datasets are not ideal for showcasing the benefits of parallelism;
in such cases, performance might even degrade due to extensive data movement.
