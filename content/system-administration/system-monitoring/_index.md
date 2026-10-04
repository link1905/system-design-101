---
title: System Monitoring
weight: 30
prev: system-deployment
next: automatic-scaling
---

System monitoring is the process of **continuously observing** and analyzing the performance and overall health of servers, networks, or applications.
It involves tracking various metrics to ensure that systems function efficiently and reliably.

## Metrics

A metric is a **quantitative measurement** that provides valuable insights into the state and performance of a system. Metrics are generally divided into two main categories:

- **Hardware Metrics**: These focus on physical hardware performance and include measurements such as CPU usage, memory consumption (RAM), network activity, and disk performance.
- **Application Metrics**: These are dynamic, application-level measurements, such as the number of HTTP requests, concurrent connections, or average latency.

Tracking and managing metrics is an essential aspect of system administration.
Metrics serve several critical purposes:

- **Enabling [Automatic Scaling]({{< ref "automatic-scaling" >}}):** Metrics support automatic scaling strategies.
  By analyzing key metrics, systems can be dynamically scaled out (provisioning additional resources to maintain performance)
  or scaled in (deallocating excess resources to reduce costs) as needed.

- **Understanding System Performance Over Time:** Metrics provide clear insights into how a system performs over time, highlighting performance trends and identifying performance bottlenecks.
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
This challenge grows as the number of servers and metrics increases.

To handle this, a **time-series store** is typically used.
This type of **NoSQL database** is specifically designed to efficiently store and analyze data that varies over time, making it an ideal solution for monitoring services.

### Time Series

**Time-series** data is grouped based on a combination of **metric names** and **labels**.

For instance, consider the following samples:

```alloy
# <name> { <labels> } <timestamp> <value>

cpu_usage { region="na", name="cpu_1" } 00:15 0.8
cpu_usage { region="eu", name="cpu_2" } 00:30 0.68 
cpu_usage { region="na", name="cpu_1" } 00:45 0.7 
cpu_usage { region="na", name="cpu_2" } 01:00 0.8
```

A **time series** consists of a sequence of samples (pairs of timestamps and values) grouped by a unique combination of labels.
Based on the examples above, we can derive two distinct time series:

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

To address this issue, the store periodically (typically every few hours) processes the WAL to create storage files that are easier to query, known as **blocks**.
A **block** corresponds to a specific time range and contains all the time-series data within that range.

In a block, a lightweight index is built to make data retrieval faster and more efficient.
Entries belonging to the same series are clustered together, and the index points to their location in the block.

For example, consider this **WAL** with 2 time-series with 5 data points:

```text
cpu_usage 00:15 0.8
memory_usage 00:15 300MB 
cpu_usage 00:30 0.75
memory_usage 00:30 500MB 
cpu_usage 00:45 0.64
```

When this is processed into a storage block, the entries are reorganized by grouping entries belonging to the same series, and an index is created for quick access:

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
Each delta is added to the preceding timestamp to reconstruct the original sequence:

```text
1750235915 0.8
15 0.75  # +15 seconds
15 0.64  # +15 seconds
```

{{% callout type="info" %}}
For encoding values, the process involves complex bit operations.
For more details, see [this blog post on Prometheus's binary data encoding](https://fungiboletus.github.io/journey-prometheus-binary-data/).
{{% /callout %}}

### Compaction

Over time, maintaining individual blocks for each time range results in excessive storage consumption and slower queries,
as scanning across multiple blocks for a single time series becomes increasingly inefficient.

To address this, **compaction** is performed to merge smaller blocks into larger ones, reducing the total number of blocks:

- **Small blocks (e.g., one hour)** are compacted into **larger blocks (e.g., one day)**.
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

Since monitoring tools (for automatic scaling, visualization, etc.) often rely on **recent data**, different storage strategies can be applied to older samples.

### Downsampling

Historical samples within a small time range (e.g., one minute) can be compressed by summarizing values.
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

This reduces storage costs but sacrifices granularity, which may not be acceptable in some scenarios.

### Collector

The **collector** is a critical component dedicated to gathering metrics from machines and applications.
Its primary role is to continuously collect data and ensure its durability by storing it in a **time-series store**.

There are two primary paradigms for collecting metrics:

- **Push Model**: In this approach, an agent is installed within each service being monitored.
  At regular intervals, the agent collects the data and sends it directly to the centralized monitoring service.

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
This can be achieved through a central [service discovery mechanism]({{< ref "load-balancer#service-discovery" >}}).

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
It is also well-suited for **short-lived jobs**, as these can report their status immediately without waiting for the next collection interval.

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
This is especially important because most services report metrics frequently, and the push model can easily lead to traffic spikes in the monitoring system.

The push model provides a more straightforward implementation, while the pull model is often preferred when reliable **Service Discovery** is in place.
The choice between these models depends on the specific needs of the system.
