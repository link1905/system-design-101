---
title: Load Balancer
weight: 40
---

We've briefly shown how to build a cluster of some instances in the [Service Cluster](service-cluster.md) topic.
In this lecture, we'll see how to expose a service to the outer world.

## Load Balancer

### Reverse Proxy Pattern

When we have a cluster of some instances,
they may live on different machines possessing different addresses.
Moreover, instances can be added or destroyed flexibly from time to time.
It's nearly impossible to let clients directly contact service instances.

**Reverse Proxy** is a pattern that exposes a system as **a single entry point**,
concealing the internal architecture.
Based on the pattern, we should place instances behind a proxy,
further forwarding traffic to them.
Probably, the proxy must be fixed and recognizable,
such as through **DNS**.

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

Proxying alone doesn't cut it.
To efficiently utilize resources,
we want a service to **equally** distribute traffic among its instances.

For example, an instance is stressed with `4 requests`,
while another instance is slack with only `1`.

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

Hence, we plug the balancing capability into the proxy component,
and call it a {{< term lb >}}.
For example,
a load balancer may use the [round-robin strategy](https://en.wikipedia.org/wiki/Round-robin_scheduling)
to fairly distribute traffic across a cluster.

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

{{< term lb >}} needs to be aware of the internal instances behind it.
The most common solution is implementing a central {{< term sd >}},
keeping track of all instances.
Load balancing solutions are typically bundled with this feature.

Briefly,
the {{< term lb >}} know nothing about a service,
the instances must register themselves with the {{< term lb >}}.

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

To ensure that only healthy instances are exposed,
the {{< term lb >}} periodically performs [health checks](Service-Cluster.md#heartbeat-mechanism) and removes corrupted ones.

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

There are several balancing algorithms for selecting an instance from the cluster.

### Round-robin

This is the most common and is often the **default option** in many solutions.
**`Round-robin`** places all instances in a ring and selects one in sequence.

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

This algorithm is suited for **short-lived** and approximate resource-consuming requests, e.g., {{< term http >}},
making the balancing task work properly.

However, because of the strict behavior,
it's problematic when requests are significantly different in resource usage.
For instance, if the second server is already overloaded with many requests,
the load balancer might still orderly distribute more traffic to it,
even though the first server has spare capacity.

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

The **Least Connections** algorithm selects a server with the fewest active client connections.
In other words,
the load balancer needs to locally keep track of how many **in-fly requests**
are being processed on instances.

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
lb.lb -> s.s2: Picks Instance 2 
```

Is it better than **Round-robin**?
Not necessarily!
The number of requests doesn't always reflect the true load.
For example, 10 requests of `Instance 1` only consume 1 MB of memory,
while 3 requests of `Instance 2` takes 100 MB.

This strategy is a viable candidate for **long-lived sessions** such as {{< term ws >}},
as clients and their requests are stuck to specific servers in the long term.
When the lifespan of user sessions is erratic,
picking servers in order easily leads to unbalance,
so we prefer **Least Connections**.

### Session Stickiness

A load balancing algorithm determines which server requests should be forwarded to.
However, this decision can be **overridden** with a feature called **Session Stickiness**.

When a client connects to the load balancer,
it receives a **stickiness key** in the first response:

1. The client locally saves the key.
2. For subsequent requests,
   the client includes the key to specify the target server.

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

Why do we need this feature?
For [stateful applications](service-cluster.md#stateful-service) like gaming services,
**Session Stickiness** is required to ensure clients meet the same server,
e.g., reconnecting to the same match after being disconnected.

However, it introduces a high risk of system imbalance
as stickiness keys override the configured {{< term lb >}} algorithm,
easily leading to unequal load distribution.

## Load Balancer Types

There are two common types of {{< term lb >}}: {{< term lb4 >}} and {{< term lb7 >}}.
They define which network layer the load balancing occurs at.

### OSI Review

Briefly, how a network message is processed within a machine can be explained in **7 layers**
in the [OSI model](https://www.cloudflare.com/learning/ddos/glossary/open-systems-interconnection-model-osi/).
![OSI Model](/images/osi_model_7_layers.png)

This layering style helps separate concerns.
Each layer handles a different responsibility, doesn't affect others, and can evolve autonomously.
You may follow the link to research it in depth,
in this topic, we solely focus on **Application**, **Transport** and **Network** layers.

#### Encapsulation

When a process wants to send a message to another process on a different machine.
The message will be steadily **encapsulated**, from a normal text to a network message:

- **Application Layer—7**: this layer actually refers to the application itself.
  The application formats the message with its **protocol**, e.g., [HTTP](../communication-protocols/#http-1-1).
- **Transport Layer—4**: this layer happens on the machine,
  it attaches the message with the application's **port**.
- **Network Layer—3**: the machine then assigns its **address** to the message.

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

Through each layer, the message is **enriched** with more network information.
However, you may note the parentheses, which used to imply encapsulation.
Layers are isolated from each other,
that means a lower layer can't understand and modify messages from the higher ones.

#### Decapsulation

When the recipient receives the message,
it handles **decapsulation** through the layers but **vice versa**.

- **Network Layer—3**: the machine reads and removes the **address** field from the message.
- **Transport Layer—4**: the machine reads the **port** and forwards to the appropriate application.
- **Application Layer—7**: the application interprets its **protocol** and handles the message

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

{{< term lb7 >}} refers to the application layer of the OSI model,
handling application protocols such as {{< term http >}} or {{< term ws >}}.

Standing at a high level allows the balancer to richly access much information,
e.g., HTTP params/headers/body.
The balancer processes **application-level messages**
before forwarding to the endpoint.
Technically, there are two connections established:

1. The client side and the load balancer.
2. The load balancer and the service.

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

**API Gateway Pattern** is a design pattern providing a **single entry point** for all public services.
That means before load balancers, we have another proxy called **API Gateway**.

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

However, load balancing is a generic task,
operating many load balancers comes with a management overhead.
We may conveniently combine load balancers and the gateway,
creating a shared load balancer across the system.

To achieve that, we define **routing rules** to distinguish between services.
Several factors can be used for routing, such as domain, HTTP path/headers/parameters.
For example:

- If the request path is `/a`, it is forwarded to the `Service A`.
- If the request path is `/b`, it is forwarded to the `Service B`.

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

{{< term sslt >}} is a big issue of {{< term lb7 >}}.
We've known that {{< term lb7 >}} reads application-level messages,
therefore,
it becomes impossible to work with the [{{< term ssl >}}](Network-Protection.md#transport-layer-security-tls) **end-to-end
encryption** protocol.
{{< term lb >}}, as an intermediary, can only capture encoded information
that is unhelpful in the balancing process.

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
the [SSL/TLS decryption](Network-Protection.md#tls-certificate) must be shifted to the {{< term lb >}} itself.
This process is known as {{< term sslt >}}.

For communication between the {{< term lb >}} and internal services, new connections are established.

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
    lb -> lb: 2. Decrypt payload {
      style.bold: true
    }
    lb -> sv: "3. Forward payload='Hello'"
}
c -> s.lb: 1. Send payload=13a8f5f167f4
```

##### Security Concern

This can be a security problem when data is exposed at the balancer, not the services themselves.
In certain data governance scenarios,
data is required to remain encrypted until reaching its endpoint,
although the communication between the internal service and the load balancer is encrypted.

Moreover, we should be careful when leveraging an **external** load balancing solution,
as {{< term sslt >}} will lead to data leakage outside the system.

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

{{< term lb4 >}} refers to the transport layer of the OSI model.
Due to residing at a low level, this type of {{< term lb >}}
does not provide detailed information about application-level data.
It can only make decisions based on two pieces of information: **address** and **port**.

In essence, a {{< term lb4 >}}
acts as a network router sitting between clients and internal services.
Once a client connects to a server,
it will constantly communicate with the same server as long as the connection is active.
Due to [packet segmentation](https://en.wikipedia.org/wiki/Packet_segmentation),
a large network packet can’t be transmitted as smaller segments.

For example,
an {{< term http >}} request is split into two network
segments ([aka TCP messages](https://en.wikipedia.org/wiki/Transmission_Control_Protocol)),
a {{< term lb7 >}} can integrate with the {{< term http >}}
protocol and effectively reassemble these segments to transmit the complete request.

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

Meanwhile, a {{< term lb4 >}} operates without awareness of the {{< term http >}} protocol.
It may naively distribute segments to different servers,
which is probably unacceptable.

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
    lb.s1 -> s1: Forward {
      class: error-conn
    }
    lb.s2 -> s2: Forward {
      class: error-conn
    }
}
c: Client {
  class: client
}
c -> s.lb.s1
c -> s.lb.s2
```

The only solution is forwarding all segments to a consistent target
until the connection is disrupted.

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

Why should we use a {{< term lb4 >}} instead of a {{< term lb7 >}}?

- It does not require {{< term sslt >}}, which can be a security concern.
- An {{< term lb4 >}} offers significantly better performance
  because it simply forwards messages without processing them.

Due to the instinctive stickiness, a {{< term lb4 >}} easily becomes **unbalanced**.
A client may send a lot of traffic to the sticky server,
causing it to become overloaded, while other servers remain underutilized.
It's a good setup for [stateful](Service-Cluster.md#stateful-service) and high-performance services,
such as gaming service.
