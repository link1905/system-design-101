---
title: Load Balancer
weight: 40
---

We previously introduced how to build a cluster of instances in the [Service Cluster]({{< ref "service-cluster" >}}) topic.
In this lecture, we’ll explore how to expose a service to the outside world.

## Load Balancer

### Reverse Proxy Pattern

When running a cluster of service instances, these instances typically reside on different machines with distinct addresses. Moreover, instances can be dynamically added or removed. As a result, it’s impractical for clients to directly communicate with individual service instances.

A **Reverse Proxy** is a pattern that exposes a system through **a single entry point**, concealing the underlying internal structure. Following this pattern, service instances are placed behind a proxy that forwards traffic to them. This proxy should be a fixed and discoverable endpoint, often achieved through **DNS**.

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

Proxying alone isn’t enough.
To efficiently utilize resources, we want to **distribute traffic evenly** across the service instances.

For example, one instance might be handling `4 requests` while another processes only `1` — clearly an imbalance.

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
    r1: Request 1 {
      class: request
    }
  }
}
c -> p
p -> s.s1
p -> s.s2
```

To solve this, we add a load balancing capability to the proxy component, which we refer to as a {{< term lb >}}.
For example, a load balancer might use a [round-robin strategy](https://en.wikipedia.org/wiki/Round-robin_scheduling) to evenly distribute traffic across the cluster.

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
lb: Load Balancer {
  class: lb
}
c: Client {
  class: client
}
c -> lb
lb -> s.s1: 1. Request 1
lb -> s.s2: 2. Request 2
lb -> s.s1: 3. Request 3
```

### Service Discovery

A {{< term lb >}} needs to be aware of the available service instances behind it.
The most common approach is to implement a central {{< term svd >}} system to track all instances. Load balancers often come bundled with this feature.

In this setup, instances must register themselves with the {{< term lb >}}, which otherwise has no inherent knowledge of their existence.

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

### Health Check

To ensure only healthy instances receive traffic,
the {{< term lb >}} periodically performs [health checks]({{< ref "service-cluster#heartbeat-mechanism" >}}) and removes unhealthy ones from the pool.

```d2
direction: right
system: System {
  s: Cluster {
    lb: Load balancer {
      lb: "" {
        class: lb
      }
      r: |||yaml
      Instance 1: 1.1.1.1, Healthy
      Instance 2: 2.2.2.2, Healthy
      |||
    }
    s1: Instance 1 {
      class: server
    }
    s2: Instance 2 {
      class: server
    }
    lb.lb -> s1: Check periodically {
      style.animated: true
    }
    lb.lb -> s2: Check periodically {
      style.animated: true
    }
  }
}
```

## Load Balancing Algorithms

Several algorithms can be used to select a service instance from a cluster.

### Round-robin

The **Round-robin** algorithm is the most common and often the **default option** in many load balancing solutions.
It cycles through the list of instances in order, assigning each new request to the next instance in sequence.

```d2
direction: right
s1: System {
  lb: Load Balancer {
    class: lb
  }
  s1: Instance 1 {
    class: server
  }
  s2: Instance 2 {
    class: server
  }
  s3: Instance 3 {
    class: server
  }
  lb -> s1: 1st request
  lb -> s2: 2nd request
  lb -> s3: 3rd request
  lb -> s1: 4th request
}
```

This method works well for **short-lived, similarly sized requests**, such as {{< term http >}} requests.

However, if the workload varies significantly, problems can arise.
For example, if `Instance 2` is already overwhelmed with ongoing requests, the load balancer will still continue to send it new requests in turn, while other instances may be underutilized.

```d2
direction: right
s1: System {
  lb: Load Balancer {
    class: lb
  }
  s1: Instance 1 {
    r: "Request" {
      class: request
    }
  }
  s2: Instance 2 {
    grid-columns: 3
    r1: "Request" {
      class: request
    }
    r2: "Request" {
      class: request
    }
    r3: "Request" {
      class: request
    }
  }
  lb -> s1
  lb -> s2: Send new request orderly {
    class: bold-text
  }
}
```

### Least Connections

The **Least Connections** algorithm selects the instance currently handling the fewest active connections.
This requires the load balancer to track the number of **in-flight requests** on each instance.

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
Not necessarily — because the number of active connections doesn’t always reflect the actual resource consumption.
For example, `10` requests on `Instance 1` might use just `1 MB` of memory, while `3` requests on `Instance 2` could consume `100 MB`.

This strategy shines for **long-lived sessions** (like {{< term ws >}} connections), where client sessions persist on the same server for extended periods.
In such cases, **Round-robin** can easily lead to imbalance, making **Least Connections** a better choice.

### Session Stickiness

Load balancing algorithms typically decide which server should handle each request. However, this can be overridden with **Session Stickiness**.

When a client first connects, the load balancer assigns a **stickiness key** and returns it in the response:

1. The client stores this key locally.
2. For subsequent requests, the client includes the key, ensuring it connects to the same instance.

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
```

**Why is this necessary?**
For [stateful applications]({{< ref "service-cluster#stateful-service" >}}) like multiplayer games or chat services, clients often need to consistently interact with the same server instance — for example, reconnecting to the same match or session after a temporary disconnection.

**However**, this comes at a cost.
Session stickiness can easily lead to uneven load distribution, as it bypasses the load balancer’s configured algorithm in favor of sticking with a specific instance.

## Load Balancer Types

There are two common types of {{< term lb >}}: {{< term lb4 >}} and {{< term lb7 >}}.
They define which network layer the load balancing occurs at.

### OSI Review

Briefly, a network message’s journey through a machine can be explained via **7 layers** in the [OSI model](https://www.cloudflare.com/learning/ddos/glossary/open-systems-interconnection-model-osi/).
![OSI Model](/images/osi_model_7_layers.png)

This layered design helps separate concerns — each layer has distinct responsibilities, operates independently, and can evolve autonomously.
In this topic, we’ll focus solely on the **Application**, **Transport**, and **Network** layers.

#### Encapsulation

When a process sends a message to another machine, it gets steadily **encapsulated**, transforming from plain text into a network message:

- **Application Layer (L7)**: the application formats the message using its specific **protocol** (e.g., [HTTP]({{< ref "communication-protocols#http-1-1" >}})).
- **Transport Layer (L4)**: the machine attaches the **port number** to the message.
- **Network Layer (L3)**: the machine adds its **address** to the message.

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

As the message moves down, it’s **enriched** with networking information at each layer.
Notably, lower layers cannot interpret or modify the data encapsulated by higher layers — maintaining isolation.

#### Decapsulation

On the recipient side, the message undergoes **decapsulation**, moving upward through the layers:

- **Network Layer (L3)**: reads and strips off the **address**.
- **Transport Layer (L4)**: reads the **port number** and routes to the correct application.
- **Application Layer (L7)**: interprets and processes the **protocol-specific message**

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

A {{< term lb7 >}} operates at the **Application Layer (L7)** of the OSI model, handling protocols like {{< term http >}} or {{< term ws >}}.

This high-level position allows it to inspect application-specific details, like HTTP headers, parameters, and message bodies — letting it make intelligent routing decisions.
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

{{< term apigw >}} is a design pattern providing a **single entry point** for all external clients.
It acts as a proxy ahead of load balancers:

```d2
direction: right
g: API Gateway {
  class: gw
}
la: Load Balancer (A) {
    class: lb
}
a: Service A {
  grid-rows: 1
  s1: Instance A1 {
    class: server
  }
  s2: Instance A2 {
    class: server
  }
}
lb: Load Balancer (B) {
    class: lb
}
b: Service B {
  grid-rows: 1
  s1: Instance B1 {
    class: server
  }
  s2: Instance B2 {
    class: server
  }
}
g -> la
la -> a
g -> lb
lb -> b
```

Operating multiple load balancers increases management complexity.
A more preferred solution combines the gateway and load balancer, sharing infrastructure and using **routing rules** to direct traffic based on criteria like domain, HTTP path, headers, or query parameters.

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

A major challenge with {{< term lb7 >}} is handling encrypted traffic via [{{< term ssl >}}]({{< ref "network-security#transport-layer-security-tls" >}}).
Since {{< term lb7 >}} needs to read application-level data to make decisions, it cannot work directly with end-to-end encryption.

```d2
direction: right
s: System {
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

In other words,
we can't use {{< term lb7 >}} to ensure **complete** end-to-end encryption.
To make it work,
the **SSL/TLS decryption** must be shifted to the {{< term lb >}} itself.
This process is known as {{< term sslt >}}.

New connections are then established internally to forward plaintext traffic to services

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
c -> s.lb: 1. Send payload=13a8f5f167f4
```

##### Security Concern

This introduces a security risk: decrypted data resides at the load balancer, potentially exposing sensitive information.

In some compliance and data governance contexts, data must remain encrypted all the way to its destination service.

Additionally, using an **external** load balancer for {{< term sslt >}} can lead to data leakage outside your trusted environment.

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
It cannot inspect application-level content — routing decisions are based solely on the **destination address and port**.

Essentially, a {{< term lb4 >}} acts like a network router between clients and services.
Once a client connects to a server, it keeps communicating with the same instance as long as the connection stays open.

This problem arises from [packet segmentation](https://en.wikipedia.org/wiki/Packet_segmentation), where large messages are split into multiple network packets (or [TCP segments](https://en.wikipedia.org/wiki/Transmission_Control_Protocol)).

For example,
an {{< term http >}} request is split into two network
segments ([aka TCP messages](https://en.wikipedia.org/wiki/Transmission_Control_Protocol)).
A {{< term lb7 >}} can understand protocols like HTTP and reassemble segmented requests before forwarding them.

```d2
direction: right
s: System {
    lb: L7 Load Balancer {
        s1: Segment 1
        s2: Segment 2
        s3: Segment 3
        s4: Segment 4
        h1: HTTP request 1
        h2: HTTP request 2
        s1 -> h1
        s2 -> h1
        s3 -> h2
        s4 -> h2
    }
    s1: Instance 1 {
       class: server
    }
    s2: Instance 2 {
       class: server
    }
    lb.h1 -> s1: Combine {
      class: bold-text
    }
    lb.h2 -> s2: Combine {
      class: bold-text
    }
}
c: Client {
  class: client
}
c -> s.lb.s1
c -> s.lb.s2
c -> s.lb.s3
c -> s.lb.s4
```

Conversely, a {{< term lb4 >}} is unaware of application protocols and may accidentally distribute segments of the same request to different servers — leading to errors.

```d2

direction: right
s: System {
    lb: L4 Load Balancer {
        s1: Segment 1
        s2: Segment 2
    }
    s1: Instance 1 {
       class: server
    }
    s2: Instance 2 {
       class: server
    }
    lb.s1 -> s1: Forward
    lb.s2 -> s2: Forward
}
c: Client {
  class: client
}
c -> s.lb.s1
c -> s.lb.s2
```

The solution is to forward all segments of a connection to the same server until it disconnects.

```d2
direction: right
s: System {
    lb: L7 Load Balancer {
      s1: Segment 1
      s2: Segment 2
      s3: Segment 3
    }
    s1: Instance 1 {
       class: server
    }
    s2: Instance 2 {
       class: server
    }
    lb.s1 -> s1
    lb.s2 -> s1
    lb.s3 -> s1
    lb -> s2: Unused {
      class: error-conn
    }
}
c: Client {
  class: client
}
c -> s.lb.s1
c -> s.lb.s2
c -> s.lb.s3
```

Why choose a {{< term lb4 >}} over a {{< term lb7 >}}?

- It avoids {{< term sslt >}}, which can be a security risk.
- It delivers significantly better performance, since it simply forwards packets without interpreting them.

However, due to this **sticky connection behavior**, a {{< term lb4 >}} can easily become **unbalanced** — one server might receive a disproportionate load while others stay underutilized.
Still, it’s a solid choice for **stateful, high-performance services** like multiplayer gaming backends.
