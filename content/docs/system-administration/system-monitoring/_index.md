---
title: System Monitoring
weight: 30
prev: system-deployment
next: automatic-scaling
---

System monitoring is the process of **continuously observing** and analyzing the performance and and overall health of servers, networks, or applications.
It involves tracking various metrics to ensure that systems function efficiently and reliably.

## Metrics

A metric is a **quantitative measurement** that provides valuable insights into the state and performance of a system. Metrics are generally divided into two main categories:

- **Hardware Metrics**: These focus on physical hardware performance and include measurements like CPU usage, memory consumption (RAM), network activity, and disk performance.
- **Application Metrics**: These are dynamic, application-level measurements, such as the number of HTTP requests, concurrent connections, or average latency.

Tracking and managing metrics is an essential aspect of system administration.
Metrics serve several critical purposes:

- **Enabling [Automatic Scaling]({{< ref "automatic-scaling" >}}):** Metrics play a pivotal role in deploying automatic scaling strategies.
  By analyzing key metrics, systems can be dynamically scaled out (provisioning additional resources to maintain performance)
  or scaled in (deallocating excess resources to reduce costs) with flexibility and precision.

- **Understanding System Performance Over Time:** Metrics provide clear insights into how a system performs over time, highlighting performance trends and pinpointing system breakpoints.
  With this information, maintainers can proactively enhance existing systems or, when necessary, redesign them to address underlying issues.

## Time-series Store

The most critical aspect of tracking metrics is the **storage layer**,
as metrics are generated at an extremely high frequency, leading to an immense volume of data.

For example, consider tracking the `CPU usage` of a server. If the server reports this metric every 15 seconds:

```alloy
cpu_usage 00:15 0.8
cpu_usage 00:30 0.75
cpu_usage 00:45 0.60 
cpu_usage 01:00 0.73
```

This results in approximately **6000 records** per day for a single metric on one server.
This challenge compounds when the system has numerous servers, each tracking multiple metrics.

To handle this, a **Time-series Store** is typically employed.
This type of **NoSQL database** is specifically designed to efficiently store and analyze data that varies over time, making it an ideal solution for monitoring services.

### Time-series

**Time-series** data is grouped based on a combination of **metric names** and **labels**.

For instance, consider the following samples:

```alloy
# <name> { <labels> } <timestamp> <value>

cpu_usage { region="na", name="cpu_1" } 00:15 0.8
cpu_usage { region="eu", name="cpu_2" } 00:30 0.68 
cpu_usage { region="na", name="cpu_1" } 00:45 0.7 
cpu_usage { region="na", name="cpu_2" } 01:00 0.8
```

A **time-series** consists of a sequence of samples (pairs of timestamps and values) grouped by a unique combination of labels.
Based on the examples above, we can derive two distinct time-series:

```alloy
cpu_usage { region="na", name="cpu_1" }:
- 00:15 0.8
- 00:45 0.7

cpu_usage { region="eu", name="cpu_2" }:
- 00:30 0.68
- 01:00 0.8
```

### Write-ahead Logging (WAL)

This type of store typically handles **write-heavy workloads**.
[WAL (Write-ahead Log)]({{< ref "system-recovery#write-ahead-logging-wal" >}}) is an effective strategy for boosting write performance.

When a new value is added, it is efficiently appended to the WAL file for durability,
ensuring that the system can recover data reliably in case of a failure.

```d2
direction: right
s: Time-series Store {
    m: Memory {
        class: cache
    }
    wal: WAL {
        class: file
    }
}
v: Value {
    class: request
}
v -> s.m: 1. Update in memory
s.m -> s.wal: 2. Log the operation
```

### Storage Block

**WAL** is especially useful because recently added data can be rapidly accessed through the memory layer.
However, querying data directly from this large, append-only file is inefficient.

To address this issue, the store periodically (typically every few hours) processes the WAL to create more queryable storage files, known as **blocks**.
A **block** corresponds to a specific time range and contains all the time-series data within that range.

In a block, a lightweight index is built to make data retrieval faster and more efficient.
Entries of the same series are clustered together, and the index points to their location in the block.

For example, consider this **WAL** with two time-series data points:

```text
cpu_usage 00:15 0.8
memory_usage 00:15 300MB 
cpu_usage 00:30 0.75
memory_usage 00:30 500MB 
cpu_usage 00:45 0.64
```

When this is processed into a storage block, the entries are reorganized by clustering the same series together, and an index is created for quick access:

```yaml
Index:
  cpu_usage:
    startIndex: 0
    count: 3
  memory_usage:
    startIndex: 3
    count: 2

Data:
  00:15 0.8
  00:30 0.75
  00:45 0.64
  00:15 300MB
  00:30 500MB
```

### Delta Encoding

In time-series data, timestamps are usually represented in [Unix format](https://www.unixtimestamp.com/),
which indicates the number of seconds since **January 1, 1970 (UTC)**:

```text
1750235915 0.8
1750235930 0.75
1750235945 0.64
```

To optimize storage, instead of storing every timestamp, we only retain the **first complete timestamp** and the deltas for subsequent data points.
The deltas are added to their preceding timestamps to reconstruct the original sequence:

```text
1750235915 0.8
15 0.75  # +15 seconds
15 0.64  # +15 seconds
```

{{% callout type="info" %}}
For encoding values, the process involves complex bit operations.
You can check out [this helpful blog on Prometheus's binary data encoding](https://fungiboletus.github.io/journey-prometheus-binary-data/).
{{% /callout %}}

### Compaction

Over time, maintaining individual blocks for each time range results in excessive storage consumption and slower queries,
as scanning across multiple blocks for a single time-series becomes increasingly inefficient.

To address this, **compaction** is performed to merge smaller blocks into larger ones, reducing the total number of blocks:

- **Small blocks (e.g., 1-hour)** are compacted into **larger blocks (e.g., 1-day)**.
- After compaction, the smaller blocks are deleted to save storage space.

```d2
direction: right
b1: "" {
  t: "Block 6-18-2025, 00:00 -> 6-18-2025, 00:59"
  c: |||yaml
  00:00 0.8
  # ...
  00:59 0.64
  |||
}

b2: "01:00 -> 22:59"

b3: "" {
  t: "Block 6-18-2025, 23:00 -> 6-18-2025, 23:59"
  c: |||yaml
  23:00 0.88
  # ...
  23:59 0.6
  |||
}

b1 -> b
b2 -> b
b3 -> b

b: "" {
  t: "Block 6-18-2025"
  c: |||yaml
  00:00 0.8
  # ...
  23:59 0.6
  |||
}
```

Since monitoring utilities (automatic scaling, visualization, etc) often rely on **recent data**, different storage strategies can be applied to older samples:

#### Compression

Historical samples within a small time range (e.g., 1 minute) can be compressed by summarizing values.
For example, averages of multiple data points can replace the original entries:

From:

```yaml
00:15 0.8
00:30 0.75
00:45 0.64
01:00 0.6
```

To:

```yaml
00:00 0.6975
```

This reduces storage costs but comes at the cost of losing granularity, which may not be acceptable in some scenarios.

#### Cold Storage

Older samples that are rarely accessed can be moved to inexpensive storage solutions (e.g., **HDDs**).
While querying old data, the store switches to the cold storage section:

```d2
direction: right

s: Time-series Store {
  class: db
}

h: Hot storage {
  class: cache
}

c: Cold storage {
  class: hd
}

s -> h
s -> c
```

## Monitoring Service

To ensure effective system oversight, it is recommended to implement a centralized monitoring service.
This setup provides a comprehensive, unified view of the system's state, significantly simplifying the management and diagnosis of monitored metrics.

A monitoring service must fulfill two core responsibilities: **collecting metrics** and **supporting their efficient querying**.

### Collector Path

The **collector** is a critical component dedicated to gathering metrics from system machines and applications.
Its primary role is to continuously collect data and ensure its durability by storing it in a **Time-series Store**.

There are two primary paradigms for collecting metrics:

- **Push Model**: In this approach, an agent is installed within each service being monitored.
At regular intervals, the agent collects and sends the data directly to the centralized monitoring service.  

  ```d2
  direction: right

  s {
    class: none
    s1: Service 1 {
      a: Agent {
        class: process
      }
    }
    s2: Service 2 {
      a: Agent {
        class: process
      }
    }
  }

  m: Monitoring Service {
      c: Push Collector {
        class: monitor
      }
      db: Time-series Store {
        class: db
      }
      c -> db: Save
  }

  s.s1.a -> m.c: Push data
  s.s2.a -> m.c: Push data
  ```

- **Pull Model**: In this model, services expose an interface (such as an endpoint) that reports their current state.
The monitoring service periodically queries these interfaces to gather the required data.  

  ```d2
  direction: right

  s {
    class: none
    s1: Service 1 {
      i: "/status"
    }
    s2: Service 2 {
      i: "/status"
    }
  }

  m: Monitoring Service {
      c: Pull Collector {
        class: monitor
      }
      db: Time-series Store {
        class: db
      }
      c -> db: Save
  }

  s.s1.i -> m.c: Pull data {
    style.animated: true
  }

  s.s2.i -> m.c: Pull data {
    style.animated: true
  }
  ```

In the **pull model**, the service needs to dynamically track current targets.
This can be achieved through a central [Service Discovery]({{< ref "load-balancer#service-discovery" >}}).

```d2
direction: right

s1: Service 1 {
  class: server
}

s2: Service 2 {
  class: server
}

s: Service Discovery {
  c: |||yaml
  Service 1: 1.1.1.1
  Service 2: 2.2.2.2
  |||
}

m: Monitoring Service (Pull) {
  class: server
}

s1 -> s: Register
s2 -> s: Register
m <- s: Read
```

Therefore, the push model is easier to work with.
It is also well-suited for **short-term jobs**, as these can report their statuses immediately without waiting.

```d2
direction: right
j: Job {
  class: process
}
m: Monitoring Service (Push) {
  class: server
}
j -> m: Report status
```

The main benefit of the pull model is **backpressure awareness**.
The monitoring service has complete control over when, how often, and what data it collects, making monitoring more manageable and resilient.
This is especially important given that reporting metrics occurs frequently across most services, and the push model can easily lead to traffic spikes in the monitoring system.

The push model provides a more straightforward implementation, while the pull model is often preferred when reliable **Service Discovery** is in place.
Choosing between these models is ultimately a design decision tailored to the specific needs of the system.

#### Distributed Cluster

A distributed time-series store can leverage [consistent hashing]({{< ref "peer-to-peer-architecture#consistent-hashing" >}})
to effectively distribute time-series data across a cluster of nodes.
This approach enhances both availability and performance by evenly balancing the storage and processing workload.

```d2
t: Time-series Store {
  s1: Server 1 {
    class: server
  }
  s2: Server 2 {
    class: server
  }
}

t1: 'cpu_usage { name="cpu_1" }'
t2: 'cpu_usage { name="cpu_2" }'

t1 -> t.s1
t2 -> t.s2
```

### Query Path

As discussed in the [time-series store section](#time-series-store),
querying directly the store can be inefficient,
as it often involves scanning multiple files with simple local indices.
Without a proper design, the query path could lead to performance bottlenecks or even system crashes.

#### Query Queueing

To handle spikes in query requests and maintain stability,
a **query queueing** mechanism can be used.
Incoming queries are stacked and executed at controlled intervals,
smoothing out potential performance spikes.

```d2
q: Query Service {
  q: Query Queue {
    grid-rows: 1
    q1: Query 1 {
      class: request
    }
    q2: Query 2 {
      class: request
    }
    q3: Query 3 {
      class: request
    }
    q4: Query 4 {
      class: request
    }
  }
}
```

A **Querier** component processes the queued queries at regular intervals to ensure orderly execution:

```d2
grid-columns: 1

q1: Query Service {
  grid-columns: 1
  q: Query Queue {
    grid-rows: 1
    q1: Query 1 {
      class: request
      style.opacity: 0.5
    }
    q2: Query 2 {
      class: request
      style.opacity: 0.5
    }
    q3: Query 3 {
      class: request
    }
    q4: Query 4 {
      class: request
    }
  }

  qu: Querier {
    class: process
  }

  qu -> q.q1
  qu -> q.q2
}

q2: Query Service {
  grid-columns: 1
  q: Query Queue {
    grid-rows: 1
    q3: Query 3 {
      class: request
      style.opacity: 0.5
    }
    q4: Query 4 {
      class: request
      style.opacity: 0.5
    }
  }

  qu: Querier {
    class: process
  }

  qu -> q.q3
  qu -> q.q4
}
```

#### Subqueries and Query Optimization

Queries can be split into **independent subqueries**, providing two main benefits:

- Subqueries can execute in parallel, improving performance.
- Redundant or overlapping requests in different queries can be merged and processed only once, saving resources.

For instance, two queries may be broken into optimized and merged subqueries:

```d2
grid-columns: 1

q {
  class: none
  q1: Query 1 {
    'cpu_usage { name="cpu_1" OR name="cpu_2" }'
  }
  q2: Query 2 {
    'cpu_usage { name="cpu_1" }'
  }
}

s {
  class: none
  s1: Sub Query 1 {
    'cpu_usage { name "cpu_1" }'
  }
  s2: Sub Query 2 {
    'cpu_usage { name "cpu_2" }'
  }
  s3: Sub Query 3 {
    'cpu_usage { name "cpu_1" }'
  }
}

sc: Final subqueries {
  s1: Query 1 {
    'cpu_usage { name "cpu_1" }'
  }
  s2: Query 2 {
    'cpu_usage { name "cpu_2" }'
  }
}

q.q1 -> s.s1
q.q1 -> s.s2
q.q2 -> s.s3
s.s1 -> sc.s1
s.s2 -> sc.s2
s.s3 -> sc.s1
```

#### Query Cache

Since loading storage blocks directly from the time-series store can be slow,
a **fast in-memory cache** should be integrated into the query service.

```d2
direction: right

q: Query Service {
  class: server
}

db: Time-series Store {
  class: db
}

c: In-memory Cache {
  class: cache
}

q -> db: Query
q -> c: Cache
```

Thus, recent samples need to be dynamically loaded from the time-series store,
while older blocks can be efficiently accessed through the cache,
reducing query latency.

#### Alerting

**Alerting** is a vital feature in monitoring systems,
enabling real-time notifications and rapid responses to critical system events.
When a significant event occurs, such as *CPU usage exceeding 70%* or *remaining memory dropping below 100MB*,
it’s essential to inform other parts of the system or stakeholders.

To manage this effectively, a dedicated component called the **Alert Controller** is designed.
Its primary role is to track and evaluate alerting rules configured for the system.

```d2
Alert Controller {
  c: |||yaml
  rule_cpu_usage:
    metric: cpu_usage
    when: > 70%
  rule_memory_usage:
    metric: memory_usage
    when: < 100MB
  |||
}
```

Periodically, the controller requests the query service to evaluate the defined rules.
If any rule is triggered, it sends a notification to a **message broker**,
which distributes the alert to multiple consumers (e.g., dashboards, alerting systems, or administrators).

```d2
direction: right
a: Alert Controller {
  class: process
}

q: Query Service {
  class: server
}

db: Time-series Store {
  class: db
}

m: Message Broker {
  class: mq
}

a -> q: Evaluate
q -> db: Query
a -> m: Notification
```
