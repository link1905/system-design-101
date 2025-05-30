---
title: Rate Limiting
weight: 30
---

Before wrapping up this topic,
let's see a critical requirement of almost systems - **Rate Limiting**.

Basically, **Rate Limiting** restricts the amount of traffic an external entity can send to a system in a time period to ensure:

- **Fair usage**: prevents a single entity (user/system) or a small group from abusing resources.
- **Stable performance**: traffic spikes easily lead to degraded performance and frustating user experience,
restricting traffic help protects from overwhelming resource usage.
- **Mitigate attacks**: [Denial-of-service attack](https://en.wikipedia.org/wiki/Denial-of-service_attack), password brute forcing...

Let's explore two famous rate limiting strategies:

## Leaky Bucket

The first algorithm is **Leaky Bucket**.
Let's imagine we have a bucket having a hole at the bottom,
traffic goes into the bucket and leak out the hole.
Although how big the bucket is, the hole is **unchanged**.
In other words, despite a lot of incoming traffic,
we only allow traffic at a fixed rate

For example, a batch of 2 requests is handled for every second.

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

This approach ensures a **consistent rate** for processing traffic,
traffic exceeding the capacity is either delayed or dropped.
This rigid behaviour makes it become a straightforward and cheap strategy.

However, due to the instinct of smoothing out traffic bursts.
Discarding (or delaying) excessive traffic will significantly degrade user experience,
thus, it's not a right choice for applications frequently encountering **bursts of traffic**,
such as gaming services.

## Token Bucket

Instead of relying on a constant rate,
**Token Bucket** is a more flexible approach with tokens

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

A token is preserved for one request,
executing the request will pick up and delete the token from the bucket.

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

If there is no token left, excessive requests are delayed or discarded

```d2
b: Bucket {
  "No token left"
}
r3: Request 3 {
  class: request
}
r3 -> b: Discarded (or delayed) because of no token
```

**Periodically**, the bucket is filled with some (configurable) tokens.
If the bucket is full of tokens, new tokens are discarded.
For example,
2 tokens are added every second.

```d2
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

In this algorithm,
we still control the average trasmission rate over time with the filling rate as **Leaky Bucket**.
However, in this approach,
tokens can be stacked up (up to the bucket's size) to deal with traffic bursts.

The problem of this approach is to prepare the sytem to work effectively in case of traffic bursts.
Traffic bursts require system services consume more resources,
potentially leading to crashes.
Moreover, bursts in a service may result in other services,
and we need to make sure all of them are intolerant.

```d2
b: "" {
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

We can also deploy rate limiting on the client side.
However, client-side stategies are unsafe,
to what extend clients can interfere with and bypass them.
Thus, it works best to support reduce traffic when the system encouters heavy loads.

### Exponential Backoff

**Exponential Backoff** is a strategy preventing clients from intensely accessing the system.
When a client experiences transient errors or rate-limiting responses from a server,
it needs to postpone before retrying.
The duration is **exponentially increased** with each retry, e.g., `1s -> 2s -> 4s -> 8s`.

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
c -> s: Request
s -> c: Respond error {
  class: error-conn
}
c -> c: Wait for 2 second
c -> s: Request
s -> c: Respond error {
  class: error-conn
}
c -> c: Wait for 4 second
```

Why should it be exponential?
When the server returns an error, it is often due to heavy load or even crash.
After a few retries, exponential backoff results in significantly longer delays,
giving the server more time to recover from high workloads.
Linear backoff, on the other hand, may unexpectedly continue contributing to the server's heaviness.

### Circuit Breaker

**Circuit Breaker** is a pattern used to prevent unnecessary attempts to an unavailable service.
**Circuit Breaker** is designed for **long-term** issues,
when we determine that requests are likely to fail,
we will to abort them immediately to save resources.

The **Circuit Breaker** works as a proxy with **three states**.

1. **Closed**: Requests pass to the target service normally in this state.

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

2. **Open**: When the breaker experiences a number of failures exceeding a **predefined threshold**,
the circuit transitions to **Open** state cancelling requests immediately.
This state prevents from wasting resources on calls that are **likely to fail** and gives the target service time to recover.

```d2
direction: right
c: Client {
  c: |||yaml
  state: Open
  threshold: 3
  failures: 4
  |||
  r: Request {
    class: request
  }
}
t: Target Service {
    class: server
}
c.r -> t: Fail {
  class: error-conn
}
```

3. **Half-open**: After a period of time, the circuit breaker switches from **Open** to **Half-open**.
Some trial requests are passed to the target service:

- If **any** request fails,
the breaker assumes the fault is still happening and retains the **Open** state.

```d2
direction: right
c: Client {
  c: |||yaml
  state: Half-open
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

- If they **all** succeed, the circuit moves back to the **Closed** state to work regularly.

```d2
direction: right
c: Client {
  c: |||yaml
  state: Half-open -> Closed
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

The **Half-Open** state only allows a limited amount of traffic,
it helps the target service from being **flooded** and have more time to recover.

The combination of **Backoff Limit** and **Circuit Breaker** is often productive.
We continue retrying until reaching the failure threshold,
then the circuit breaker should be leveraged to throttle requests immediately
