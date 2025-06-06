---
title: Rate Limiting
weight: 30
---

**Rate Limiting** is an essential component in every system.
At its core, **Rate Limiting** ensures controlled access to a system by limiting the volume of
traffic an external entity can send in a specific time period.
Its primary purposes include:

- **Fair usage:** Prevents any single entity (user/system) or group from monopolizing resources.
- **Stable performance:** Protects systems from performance degradation caused by traffic spikes, improving user experience.
- **Attack mitigation:** Helps defend against [Denial-of-Service attacks (DoS)](https://en.wikipedia.org/wiki/Denial-of-service_attack), password brute-forcing, and more.

This article dives into popular strategies employed for implementing rate limiting effectively:

## Leaky Bucket

The **Leaky Bucket** works much like its name suggests: imagine a bucket with a hole at the bottom.
Requests flow into the bucket and **leak** at a constant rate. Regardless of the intensity of incoming traffic, only a fixed volume is processed.

For example,
incoming requests are queued,
with a limit of 2 requests can exit per second.

```d2
grid-columns: 3
s1: 00:00 {
   grid-columns: 1
   m: "Bucket" {
      grid-columns: 2
      m1: Request 1 {
         class: request
      }
      m2: Request 2 {
         class: request
      }
      m3: Request 3 {
         class: request
      }
      m4: Request 4 {
         class: request
      }
   }
}
s2: 00:01 {
   grid-columns: 1
   m1: "Bucket" {
      grid-columns: 2
      height: 450
      m3: Request 3 {
         class: request
      }
      m4: Request 4 {
         class: request
      }
   }
   m2: "Processed" {
      grid-columns: 2
      m1: Request 1 {
         class: request
      }
      m2: Request 2 {
         class: request
      }
   }
   m1 -> m2
}
s3: 00:02 {
   grid-columns: 1
   m1: "Bucket" {
      height: 450
   }
   m2: "Processed" {
      grid-columns: 2
      m3: Request 3 {
         class: request
      }
      m4: Request 4 {
         class: request
      }
   }
   m1 -> m2
}
```

This approach reliably maintains a **consistent processing rate** for traffic.
Excess requests are either queued (thereby delayed) or discarded,
rendering it an inherently straightforward and cost-effective strategy for implementation.

However, its primary function is to smooth out traffic bursts.
Consequently, discarding or delaying excessive traffic can significantly **degrade the user experience**.
Therefore,
it is not an ideal choice for applications that frequently encounter bursts of traffic,
such as online gaming services.

## Token Bucket

The **Token Bucket** is similar but more flexible than the **Leaky Bucket**.

- Each token represents the permission to process a request.
- Tokens are added to the bucket at a steady rate.

Imagine we have a bucket of tokens.

```d2
bucket: Bucket {
  t1: Token 1 {
    class: token
  }
  t2: Token 2 {
    class: token
  }
}
```

When a request arrives, it picks up a token:

```d2
b: Bucket {
  t1: Token 1 {
    class: token
    style.opacity: 0.5
  }
  t2: Token 2 {
    class: token
    style.opacity: 0.5
  }
}
r1: Request 1 {
  class: request
}
r2: Request 2 {
  class: request
}
r1 -> b.t1: Take
r2 -> b.t2: Take
```

If tokens are exhausted, requests are delayed or discarded:

```d2
direction: right
b: Bucket {
  "No token left"
}
r3: Request 3 {
  class: request
}
r3 -> b: Discarded (or delayed) because of no token
```

**Periodically**, the bucket receives a configurable number of tokens.
If the bucket reaches its capacity, any excess tokens are discarded.
For example, 2 tokens are added every second.

```d2
b0: Bucket (00:00) {
  t1: Token 1 (Filled) {
    class: none
    height: 130
  }
}
b1: Bucket (00:01) {
  t1: Token 1 (Filled) {
    class: token
  }
  t2: Token 2 (Filled) {
    class: token
  }
}

b2: Bucket (00:02) {
  t1: Token 1 {
    class: token
  }
  t2: Token 2 {
    class: token
  }
  t3: Token 3 (Filled) {
    class: token
  }
  t4: Token 4 (Filled) {
    class: token
  }
}
```

This algorithm regulates the average data transmission rate over time by managing its **token filling rate**,
a mechanism conceptually similar to the **Leaky Bucket** algorithm.

A key distinction, however, is that this approach permits tokens to **accumulate** during periods of lower traffic (slack rounds),
up to the predetermined capacity of the bucket.
This accumulated reserve of tokens then enables the system to effectively absorb and manage sudden bursts of traffic by
allowing temporary transmission rates higher than the average.

A key challenge with this method is preparing the system to operate effectively during such bursts. Traffic bursts compel system services to consume additional resources, potentially leading to crashes. Furthermore, as bursts in one service can propagate to others, it is vital to ensure that all affected services are intolerant.

```d2
direction: right
b: "Traffic bursts" {
  class: burst
}
s: System {
  s1: Service 1 {
    class: server
  }
  s2: Service 2 {
    class: server
  }
  s3: Service 3 {
    class: server
  }
  s1 -> s2
  s1 -> s3
}
b -> s.s1
```

## Client-side Limiting

While rate limiting mechanisms can be implemented on the client side,
such strategies are inherently **unreliable** and considered unsafe from a security perspective.
This is because client-side controls are susceptible to manipulation or complete bypass by the end-user.

Consequently, client-side rate limiting should only be employed in a **supplementary capacity**,
supporting more robust server-side enforcement, rather than serving as a primary security measure

### Exponential Backoff

**Exponential Backoff** is a strategy that prevents clients from accessing the system too intensely.
When a client encounters transient errors or rate-limiting responses from a server,
it should pause before retrying.
This pause duration is **exponentially increased** with each subsequent retry,
for example: `1s -> 2s -> 4s -> 8s`.

```d2
shape: sequence_diagram
c: Client {
  class: client
}
s: Server {
  class: server
}
c -> s: Request
s -> c: Respond error {
  class: error-conn
}
c -> c: Wait for 1 second
c -> s: Retry
s -> c: Respond error {
  class: error-conn
}
c -> c: Wait for 2 second
c -> s: Retry
s -> c: Respond error {
  class: error-conn
}
c -> c: Wait for 4 second
```

Why should the backoff be exponential?
When a server returns an error, it often indicates a heavy load or even a crash.
After a few retries, exponential backoff introduces significantly longer delays.
This gives the server more time to recover from high workloads.
Linear backoff, in contrast, might inadvertently continue to contribute to the server's heavy load.

### Circuit Breaker

The **Circuit Breaker** pattern is particularly designed for handling **long-term issues**.
When it's determined that requests are likely to fail,
the **Circuit Breaker** aborts them immediately,
thereby conserving resources.

This pattern operates as a proxy and manages requests through **three distinct states**:

1. **Closed**: In this state, requests are routed to the target service as usual.

```d2
direction: right
c: Client {
  c: |||yaml
  state: Closed
  threshold: 3
  failures: 0
  |||
}
t: Target Service {
    class: server
}
c -> t {
  style.animated: true
}
```

2. **Open**: If the number of failures surpasses a **predefined threshold**,
the circuit breaker transitions to the **Open** state.
In this state, all requests are immediately cancelled.
This prevents resource wastage on calls that are likely to fail and provides the target service with an opportunity to recover.

```d2
direction: right
c: Client {
  c: |||yaml
  state: Open
  threshold: 3
  failures: 4
  |||
}
t: Target Service {
    class: server
}
c -> t: Fail {
  class: error-conn
}
```

3. **Half-Open**: After a designated timeout period,
the circuit breaker transitions from the **Open** state to **Half-Open**.
In this state, a limited number of trial requests are allowed to pass through to the target service:

* If **any** of these trial requests fail,
the breaker presumes the underlying fault persists and reverts to the **Open** state.

```d2
direction: right
c: Client {
  c: |||yaml
  state: Half-Open -> Open
  threshold: 3
  failures: 4
  |||
  r1: Request 1 {
    class: request
  }
  r2: Request 2 {
    class: request
  }
}
t: Target Service {
  class: server
}
c.r1 -> t: Successful
c.r2 -> t: Failed {
  class: error-conn
}
```

* If **all** trial requests succeed,
the circuit breaker transitions back to the **Closed** state, resuming normal operation.

```d2
direction: right
c: Client {
  c: |||yaml
  state: Half-Open -> Closed
  threshold: 3
  failures: 4
  |||
  r1: Request 1 {
    class: request
  }
  r2: Request 2 {
    class: request
  }
}
t: Target Service {
  class: server
}
c.r1 -> t: Successful
c.r2 -> t: Successful
```

The **Half-Open** state permits only a restricted volume of traffic, which helps prevent the target service from being **overwhelmed** and allows it additional time to recover.

Combining **Exponential Backoff** and **Circuit Breaker** strategies is often effective.
Retries (with exponential backoff) may continue until the Circuit Breaker's failure threshold is reached,
at which point the breaker activates to throttle requests immediately.
