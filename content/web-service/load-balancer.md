---
title: Load Balancer
weight: 40
prev: communication-protocols
next: api-design
---

We previously introduced how to build a cluster of instances in the [Service Cluster]({{< ref "service-cluster" >}}) topic.
In this lecture, we'll explore how to expose a service to the outside world.

## Load Balancer

### Reverse Proxy Pattern

When running a cluster of service instances, the instances typically reside on different machines with distinct addresses.
Moreover, instances can be dynamically added or removed. As a result, it is impractical for clients to communicate directly with individual service instances.

A **Reverse Proxy** is a pattern that exposes a system through *a single entry point* while concealing its underlying internal structure.
Following this pattern, service instances are placed behind a proxy that forwards traffic to them.
The proxy should have a fixed and discoverable endpoint, often provided through **DNS**.

```d2
direction: right
c: Client {
  class: client
}
p: Proxy {
  class: lb
}
s: Service {
  s1: Instance 1 {
    class: server
  }
  s2: Instance 2 {
    class: server
  }
}
c -> p
p -> s.s1
p -> s.s2
```

### Load Balancing

Proxying alone isn't enough.
To utilize resources efficiently, we want to *distribute traffic evenly* across the service instances.

For example, one instance might be handling `4 requests` while another is processing only `1`, creating a clear imbalance.

```d2
direction: right
c: Client {
  class: client
}
p: Proxy {
  class: lb
}
s: Service {
  s1: Instance 1 {
    grid-rows: 1
    r1: Request 1 {
      class: request
    }
    r2: Request 2 {
      class: request
    }
    r3: Request 3 {
      class: request
    }
    r4: Request 4 {
      class: request
    }
  }
  s2: Instance 2 {
    r1: Request 5 {
      class: request
    }
  }
}
c -> p
p -> s.s1
p -> s.s2
```

To solve this problem, we add load balancing capabilities to the proxy component, which we refer to as a {{< term lb >}}.
For example, the load balancer can distribute traffic evenly across the cluster.

```d2
direction: right
c: Client {
  class: client
}
p: Proxy {
  class: lb
}
s: Service {
  s1: Instance 1 {
    grid-rows: 1
    r1: Request 1 {
      class: request
    }
    r2: Request 2 {
      class: request
    }
  }
  s2: Instance 2 {
    grid-rows: 1
    r3: Request 3 {
      class: request
    }
    r4: Request 4 {
      class: request
    }
  }
}
c -> p
p -> s.s1
p -> s.s2
```

### Service Discovery

A {{< term lb >}} needs to know which service instances are available behind it.
The most common approach is to implement a central {{< term svd >}} system that tracks all available instances.

In this setup, service instances must register themselves with the {{< term lb >}}, which would otherwise have no inherent knowledge of their existence.

```d2
direction: right
s: Service {
  s1: Instance 1 {
    class: server
  }
  s2: Instance 2 {
    class: server
  }
}
sd: Load balancer {
  lb: "" {
    class: lb
  }
  r: |||yaml
  Instance 1: 1.1.1.1
  Instance 2: 2.2.2.2
  |||
}
s.s1 -> sd.lb: Register
s.s2 -> sd.lb: Register
```

#### Health Check

To ensure that only healthy instances receive traffic,
the {{< term lb >}} periodically performs [health checks]({{< ref "service-cluster#heartbeat-mechanism" >}}) and removes unhealthy instances from the pool.

```d2
direction: right
c: Client {
  class: client
}
s: System {
  lb: Load balancer {
    lb: "" {
      class: lb
    }
    r: |||yaml
    Instance 1: 1.1.1.1, Healthy
    Instance 2: 2.2.2.2, Unhealthy
    |||
  }
  s1: Instance 1 {
    class: server
  }
  s2: Instance 2 {
    class: generic-error
  }
  lb.lb -> s1: Health check {
    style.animated: true
  }
  lb.lb -> s2: Stop forwarding {
    class: error-conn
  }
}
c -> s.lb
```

## Load Balancing Algorithms

Several algorithms can be used to select a service instance from a cluster.

### Round-robin

The **Round-robin** algorithm is one of the most common algorithms and is often the *default option* in many load balancing solutions.
It cycles through the list of instances in order, assigning each new request to the next instance in the sequence.

```d2
s1: System {
  grid-rows: 2
  lb: Load Balancer {
    class: lb
  }
  i {
    class: none
    grid-rows: 1

    s1: Instance 1 {
      class: server
    }
    s2: Instance 2 {
      class: server
    }
    s3: Instance 3 {
      class: server
    }
  }
  lb -> i.s1: 1st request
  lb -> i.s2: 2nd request
  lb -> i.s3: 3rd request
  lb -> i.s1: 4th request
}
```

This method works well for *short-lived requests of similar size*, such as {{< term http >}} requests.

However, problems can arise when workloads vary significantly.
For example, if `Instance 2` is already overwhelmed with ongoing requests, the load balancer will still continue sending it new requests when its turn arrives, even while other instances remain underutilized.

```d2

s1: System {
  grid-rows: 2
  lb: Load Balancer {
    class: lb
  }
  i: {
    class: none
    grid-rows: 1
    s1: Instance 1 {
      r: "Request" {
        class: request
      }
    }
    s2: Instance 2 {
      grid-columns: 3
      r3: "Request" {
        class: request
      }
      r1: "Ongoing Request" {
        class: request
      }
      r2: "Ongoing Request" {
        class: request
      }
    }
  }
  lb -> i.s1.r
  lb -> i.s2.r3: Send new request orderly {
    class: bold-text
  }
}
```

### Least Connections

The **Least Connections** algorithm selects the instance currently handling the fewest active connections.
This requires the load balancer to track the number of *in-flight requests* on each instance.

```d2
direction: right
s: Service {
  s1: Instance {
    class: server
  }
  s2: Instance 2 {
    class: server
  }
}
lb: Load Balancer {
  lb: "" {
    class: lb
  }
  r: |||yaml
  Instance 1: ActiveConnections=10
  Instance 2: ActiveConnections=3
  |||
}
lb.lb -> s.s2: Pick Instance 2
```

Is this better than **Round-robin**?
Not necessarily, because the number of active connections does not always reflect actual resource consumption.
For example, `10` requests on `Instance 1` might use only `1 MB` of memory, while `3` requests on `Instance 2` could consume `100 MB`.

This strategy is particularly effective for *long-lived sessions*, such as {{< term ws >}}, where client sessions remain connected to the same server for extended periods.
In such cases, **Round-robin** can easily produce an imbalance, making **Least Connections** a more suitable choice.

### Session Stickiness

Load balancing algorithms typically determine which server should handle each request. However, this behavior can be overridden using a feature called **Session Stickiness**.

When a client first connects, the load balancer assigns a **stickiness key** and returns it in the response:

1. The client stores the key locally.
2. For subsequent requests, the client includes the key, ensuring that it connects to the same instance.

```d2
shape: sequence_diagram
c: Client {
  class: client
}
lb: Load Balancer {
  class: lb
}
s0: Instance 1 (I1) {
  class: server
}
c -> lb: 1. Connect to the system
lb -> s0: 2. Pick I1 as the sticky instance
c <- lb: '3. Respond with stickiness key "I1"' {
  style.bold: true
}
c -> lb: '4. Use the key to connect to "I1"'
lb -> s0
```

**Why is this used?**

In many load balancing solutions, this feature is disabled by default.
However, for [stateful applications]({{< ref "service-cluster#stateful-service" >}}), such as multiplayer games or chat services, clients often need to interact consistently with the same server instance.
One example is reconnecting to the same session after a temporary disconnection.

However, this behavior comes at a cost.
**Session Stickiness** can easily lead to uneven load distribution because it overrides the load balancer's configured algorithm in favor of routing a client to a specific instance.

## Load Balancer Types

There are two common types of {{< term lb >}}: {{< term lb4 >}} and {{< term lb7 >}}.
They differ in the network layer at which load balancing occurs.

### OSI Review

A network message's journey through a machine can be described using the **7 layers** of the [OSI model](https://www.cloudflare.com/learning/ddos/glossary/open-systems-interconnection-model-osi/).

![OSI Model](/images/osi_model_7_layers.png)

This layered design helps separate concerns. Each layer has distinct responsibilities, operates independently, and can evolve autonomously.
In this topic, we'll focus solely on the **Application**, **Transport**, and **Network** layers.

#### Encapsulation

When a process sends a message to another machine, the message is progressively **encapsulated**, transforming from application data into a network message:

- **Application Layer (L7)**: The application formats the message according to its specific **protocol**, such as [HTTP]({{< ref "communication-protocols#http-1-1" >}}).
- **Transport Layer (L4)**: The machine attaches the **port number** to the message.
- **Network Layer (L3)**: The machine adds its **address** to the message.

```d2

m: Machine {
  grid-columns: 1
  a: Application {
    grid-columns: 1
    d: Message
    l7: Application Layer (L7) {
      style.fill: ${colors.i2}
    }
    http: "message='GET /docs?name=README&team=dev'"
    d -> l7
    l7 -> http
  }
  l4: Transport Layer (L4) {
      style.fill: ${colors.i2}
  }
  l4r: "port=8080 (message='GET /docs?name=README&team=dev')"
  l3: Network Layer (L3) {
      style.fill: ${colors.i2}
  }
  l3r: "ip=1.9.0.5 (port=8080 (message='GET /docs?name=README&team=dev'))"
  a.http -> l4
  l4 -> l4r
  l4r -> l3
  l3 -> l3r
}
```

As the message moves downward through the layers, it is **enriched** with additional networking information at each stage.
Notably, lower layers cannot interpret or modify the data encapsulated by higher layers.

#### Decapsulation

On the recipient side, the message undergoes **decapsulation**, moving upward through the layers:

- **Network Layer (L3)**: Reads and strips off the **address**.
- **Transport Layer (L4)**: Reads the **port number** and routes the message to the correct application.
- **Application Layer (L7)**: Interprets and processes the **protocol-specific message**.

```d2

m: Machine {
  grid-columns: 1
  a: Application {
    grid-columns: 1

    d: The original message
    l7: Application Layer (L7) {
      style.fill: ${colors.i2}
    }
    http: "message='GET /docs?name=README&team=dev'"
    d <- l7
    l7 <- http
  }
  l4: Transport Layer (L4) {
      style.fill: ${colors.i2}
  }
  l4r: "port=8080 (message='GET /docs?name=README&team=dev')"
  l3: Network Layer (L3) {
      style.fill: ${colors.i2}
  }
  l3r: "ip=1.9.0.5 (port=8080 (message='GET /docs?name=README&team=dev'))"
  a.http <- l4
  l4 <- l4r
  l4r <- l3
  l3 <- l3r
}
```

### Layer 7 Load Balancer

A {{< term lb7 >}} operates at the **Application Layer (L7)** of the OSI model, handling protocols such as {{< term http >}} and {{< term ws >}}.

Operating at this higher layer allows it to inspect *application-specific details*, such as HTTP headers, parameters, and message bodies, enabling more intelligent routing decisions.

Technically, two separate connections are established:

1. Between the client and the load balancer.
2. Between the load balancer and the service.

```d2
direction: right
c: Client {
    class: client
}
lb: L7 Load Balancer {
    class: lb
}
s: Service {
  grid-rows: 1
  s1: Instance 1 {
    class: server
  }
  s2: Instance 2 {
    class: server
  }
}
c <-> lb
lb <-> s
```

#### API Gateway Pattern

{{< term apigw >}} is a design pattern that provides a *single entry point* to a system's public services.

It can also provide load-balancing capabilities, allowing multiple services to share the same infrastructure while using **routing rules** to direct traffic based on criteria such as the domain, HTTP path, headers, or query parameters.

For example:

- Requests to `/a` are routed to `Service A`.
- Requests to `/b` are routed to `Service B`.

```d2
direction: right
lb: Load Balancer + Gateway {
  class: lb
}
auth: Service A {
  grid-rows: 1
  s1: Instance 1 {
    class: server
  }
  s2: Instance 2 {
    class: server
  }
}
user: Service B {
  grid-rows: 1
  s1: Instance 1 {
    class: server
  }
  s2: Instance 2 {
    class: server
  }
}

lb -> auth: /a {
  style.animated: true
  class: bold-text
}
lb -> user: /b {
  style.animated: true
  class: bold-text
}
```

#### SSL Termination

A major challenge with {{< term lb7 >}} is handling traffic encrypted with [SSL/TLS](https://en.wikipedia.org/wiki/Transport_Layer_Security).
Because {{< term lb7 >}} needs to inspect application-level data to make routing decisions, it cannot operate directly on **end-to-end encrypted** traffic.

```d2
direction: right
s: System {
  direction: right
    lb: Load Balancer {
        class: lb
    }
    sv: Service {
        class: server
    }
    lb -> sv
}
c: Client {
  class: client
}
c -> s.lb: payload=13a8f5f167f4 {
  class: bold-text
}
```

In other words, we cannot use a {{< term lb7 >}} while preserving complete end-to-end encryption.
To make L7 load balancing possible, **SSL/TLS decryption** must instead occur at the {{< term lb >}} itself.
This process is known as {{< term sslt >}}.

New connections are then established internally to forward plaintext traffic to the services.

```d2
grid-rows: 1
horizontal-gap: 300
c: Client {
  class: client
}
s: System {
    direction: right
    lb: Load Balancer {
        class: lb
    }
    sv: Service {
        class: server
    }
    lb -> lb: "2. Decrypt payload='Hello'"{
      style.bold: true
    }
    lb -> sv: "3. Forward payload='Hello'" {
      style.bold: true
    }
}
c -> s.lb: 1. Send payload=a8f5f167f4
```

##### Security Concern

This introduces a security risk because decrypted data resides at the load balancer, potentially exposing sensitive information.

In some compliance and data governance contexts, data must remain encrypted all the way to its destination service.
Additionally, using an *external* load balancing service for {{< term sslt >}} may expose decrypted data outside the trusted environment.

```d2
grid-rows: 2
horizontal-gap: 400
e1 {
  class: none
}
lbw {
  class: none
  lb: External load balancer {
    class: lb
  }
  lb -> lb: 2. SSL termination (data can be leaked here) {
    class: error-conn
  }
}
c: Client {
  class: client
}
s: System {
  sv: Service {
    class: server
  }
}

c -> lbw.lb: 1. Send HTTPs requests
lbw.lb -> s.sv: 3. Forward
```

### Layer 4 Load Balancer

A {{< term lb4 >}} operates at the **Transport Layer (L4)** of the OSI model.
It cannot inspect application-level content, so routing decisions are based solely on the *destination address and port*.

Essentially, a {{< term lb4 >}} acts like a network router between clients and services.
Once a client connects to a server, it continues communicating with *the same instance* for as long as the connection remains open.
This behavior arises from [packet segmentation](https://en.wikipedia.org/wiki/Packet_segmentation), where large messages are split into multiple network packets, also known as [TCP segments](https://en.wikipedia.org/wiki/Transmission_Control_Protocol).

For example, an {{< term http >}} request may be split into two network segments.
A {{< term lb7 >}} understands application protocols such as HTTP and can route requests to the appropriate targets.

```d2
grid-columns: 1
r {
  class: none
  grid-rows: 1
  h1: HTTP request 1
  h2: HTTP request 2
}
s: System {
  grid-columns: 1
  lb: L7 Load Balancer {
    grid-rows: 1
    s1: Segment 1
    s2: Segment 2
    s3: Segment 3
    s4: Segment 4
  }
  i: {
    grid-rows: 1
    class: none
    s1: Instance 1 {
      class: server
      h1: HTTP request 1
    }
    s2: Instance 2 {
      class: server
      h2: HTTP request 2
    }
  }
  lb.s1 -> i.s1.h1: Assemble
  lb.s2 -> i.s1.h1: Assemble
  lb.s3 -> i.s2.h2: Assemble
  lb.s4 -> i.s2.h2: Assemble
}
r.h1 -> s.lb.s1
r.h1 -> s.lb.s2
r.h2 -> s.lb.s3
r.h2 -> s.lb.s4
```

Conversely, a {{< term lb4 >}} is unaware of application-level protocols and may accidentally distribute segments of the same request across different servers, resulting in errors.

```d2
s: System {
    lb: L4 Load Balancer {
        s1: Segment 1
        s2: Segment 2
    }
    i: {
      class: none
      s1: Instance 1 {
        class: server
      }
      s2: Instance 2 {
        class: server
      }
    }
    lb.s1 -> i.s1: Forward
    lb.s2 -> i.s2: Forward
}
c: HTTP request
c -> s.lb.s1
c -> s.lb.s2
```

The solution is to forward all segments belonging to the same connection to the same server until the connection is closed.

Why choose a {{< term lb4 >}} over a {{< term lb7 >}}?

- It avoids {{< term sslt >}}, which can introduce security risks.
- It provides significantly better performance because it can simply forward packets without interpreting application-level content.

However, because of this connection stickiness, a {{< term lb4 >}} can easily become *unbalanced*: one server might receive a disproportionate amount of traffic while others remain underutilized.
Still, it is a solid choice for *stateful, high-performance services* such as multiplayer gaming backends.