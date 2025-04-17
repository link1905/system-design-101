---
title: Load Balancer
weight: 40
---

We've briefly how to build a cluster of some instances in the [](service-cluster.md) topic.
In this lecture, we'll see how to expose a service to the outer world.

## Load Balancer

### Reverse Proxy Pattern

When we have a cluster of some instances,
they may live in different machines possessing different addresses.
Moreover, instances can be added or destroyed flexibly from time to time.
It's nearly impossible to let clients directly contact with service instances.

`Reverse Proxy` is a pattern that exposes a system as **a single entrypoint**,
concealing the internal architecture.
Based on the pattern, we should place instances behind a proxy component,
further forwarding traffic to them.
Probably, the proxy must be fixed and recognizable,
such as through `DNS`.

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

For example, an instance is under stress of 4 requests,
meanwhile another instance is slack with only 1

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
    r: Request 1
    r: Request 2
    r: Request 3
    r: Request 4
  }
  s2: Instance 2 {
    r: Request 1
  }
}
c -> p
p -> s.s1
p -> s.s2
```

Hence, we plug the balancing capability into the proxy component,
and call it as a `%lb%`.
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
lb -> s.s1: Request 1
lb -> s.s2: Request 2
lb -> s.s1: Request 3
```

### Service Discovery

`%lb%` needs to be aware of the internal instances behind it.  
The most common solution is implementing a central `%sd%`,
keeping track of all instances.
Load balancing solutions are typically bundled with this feature.

Briefly,
the `%lb%` know nothing about a service,
the instances must register themselves with the `%lb%`.

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
  class: cache
}
s.s1 -> sd: Register
s.s2 -> sd: Register
```

### Health Check

To ensure that only healthy instances are exposed,  
the `%lb%` periodically performs [health checks](Service-Cluster.md#heartbeat-mechanism) and removes corrupted ones.

```d2
direction: right
system: System {
    s: Cluster {
      sd: Load balancer {
        class: cache
      }
      s1: Instance 1 {
        class: server
      }
      s2: Instance 2 {
        class: server
      }
      sd -> s1: Check periodically {
        style.animated: true
      }
      sd -> s2: Check periodically {
        style.animated: true
      }
    }
}
```

## Load Balancing Algorithms

There are several load balancing algorithms for selecting an instance from the cluster.

### Round-robin

This is the most common and is often the **default option** in many solutions.

`Round-robin` places all instances in a ring and selects one in sequence.

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

This algorithm is suited for short-lived and equal resource-consuming requests, e.g. `HTTP`,
making the balancing task work properly.

However, because of strict behaviour,
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
    s1: Instance 1 (I1)
    s2: Instance 2 (I2) {
        grid-columns: 3
        r1: "Request"
        r2: "Request"
        r3: "Request"
    }
    lb -> s1
    lb -> s2: Request 
}
```

### Least Connections

The `Least Connections` algorithm selects a server with the fewest active client connections.
In other words,
the load balancer needs to locally keep track of how many **in-fly requests**
are being processed on instances.

```d2

direction: right
s: Service {
    s1: Instance 1 is handling 3 requests {
        class: server
    }
    s2: Instance 2 is handling 1 requests {
        class: server
    }
    s3: Instance 3 is hanlding 10 requests {
        class: server
    }
}
lb: Load Balancer {
    class: lb
}
lb -> s.s2: Picks Server 2 
```

Is it better than `Round-robin`?
Not necessarily!
The number of requests doesn't always reflect the true load.
For example, 10 requests of `Instance 3` only consume 1MB RAM,
while 1 request of `Instance 2` takes 100MB RAM.

This strategy is a good candidate for **long-lived sessions** such as `WebSocket`,
as clients and their requests are stuck to specific servers in the long term.
When the lifespan of user sessions is erratic,
picking servers in order easily leads to unbalance,
so we prefer `Least Connections`.

### Session Stickiness

A load balancer (`%lb%`) algorithm determines which server a request should be forwarded to.
However, this decision can be **overridden** using a feature called `Session Stickiness`.

When a client connects to the load balancer,
it receives a **stickiness key** in the first response:

1. The client locally retains the key.
2. For subsequent requests,
   the client includes the key to specify the target server.

```d2

direction: right

c: Client {
  class: client
}
s: System {
    lb: Load Balancer {
        class: lb
    }
    s1: Instance 1 (I1) {
        class: server
    }
    s2: Instance 2 (I2) {
        class: server
    }
    lb -> s1: 2. Forwards requests
    lb -> s2
}

c -> s.lb: 1. Connect to the system
c <- s.lb: '3. Respond with stickiness key "I1"'
c -> s.lb: '4. Use the key to connect to I1'
```

Why do we need this feature?
For [stateful applications](Service-Cluster.md#stateful-service) like gaming services,
`Session Stickiness` is leveraged to ensure clients to meet the same server.

However, it introduces a high risk of system imbalance,
as stickiness keys override the configured `%lb%` algorithm,
easily leading to unequal load distribution.

## Load Balancer Types

There are two common types of `%lb%`: `Layer 4 (L4)` and `Layer 7 (L7)`.  
They define which **network layer** the load balancing occurs at.

### OSI Review

In brief, how a network message is processed within a machine can be explained **7 layers**
in the [OSI model](https://www.cloudflare.com/learning/ddos/glossary/open-systems-interconnection-model-osi/).
![](https://cf-assets.www.cloudflare.com/slt3lc6tev37/6ZH2Etm3LlFHTgmkjLmkxp/59ff240fb3ebdc7794ffaa6e1d69b7c2/osi_model_7_layers.png)

This layering style helps separate concerns.
Each layer handles a different responsibility, doesn't affect others and can evolve autonomously.
You may follow the link to research it in depth,
on this topic we sparely focus on `Application`, `Transport` and `Network` layers.

#### Message Sender

When a process wants to send a message to another process on a different machine.
The message will be steadily **encapsulated**, from a normal text to a network message:

- `Application Layer — 7`: this layer actually refers to the application itself.
  The application formats the message with its **protocol**, e.g., [HTTP](Communication-Protocols.md#http-1-1)
- `Transport Layer — 4`: this layer happens on the machine,
  it attaches the message with the application's **port**.
- `Network Layer — 3`: the machine then assigns its **address** to the message.

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

#### Message Recipient

When the recipient receives the message,
it handles **decapsulation** through the layers but **vice versa**.

- `Network Layer — 3`: the machine reads and removes the **address** field from the message.
- `Transport Layer — 4`: the machine reads the **port** and forwards to the appropriate application.
- `Application Layer — 7`: the application interprets its protocol and handles the message

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

`Layer 7 %lb%` refers to the application layer of the OSI model,
handling application protocols such as `HTTP` or `WebSocket`.

Standing at a high level allows the balancer to **richly access** much information,
e.g., HTTP params, headers, body.
The balancer processes and distributes **application-level messages**,
there are two connections are established:

- The client side and the load balancer
- The load balancer and the server side

```d2

direction: right
c: Client {
    class: client
}
lb: Load Balancer {
    class: lb
}
s: Service {
    class: server
}
c <-> lb
lb <-> s
```

#### API Gateway Pattern

`API Gateway Pattern` is a design pattern providing a **single entry point** for all public services.
That means, before load balancers, we have another proxy called `API Gateway`.

```d2

direction: right
g: API Gateway {
  class: gw
}
la: Load Balancer (A) {
    class: lb
}
a: Service A {
  class: server
}
lb: Load Balancer (B) {
    class: lb
}
b: Service B {
  class: server
}
g -> la 
la -> a
g -> lb
lb -> b
```

However, load balancing is a generic workload,
operating many load balancers comes with a management overhead.
We may conveniently combine load balancers and `API Gateway`,
creating a shared load balancer across the system.

To achieve that, we define **routing rules** to distinguish between services.
Several factors can be used for routing, such as domain, HTTP path/headers/parameters.
For example:

- If the request path is `/a`, it is forwarded to the `Service A`.
- If the request path is `/b`, it is forwarded to the `Service B`.

```d2

direction: right
system: System {
    s: Cluster {
      auth: Service A {
          s1: Instance 1 {
            class: server
          }
          s2: Instance 2 {
            class: server
          }
      }
      user: Service B {
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
      lb -> auth.s1: /a
      lb -> auth.s2: /a
      lb -> user.s1: /b
      lb -> user.s2: /b
    }
}
```

#### SSL Termination

`SSL Termination` is a big issue of `Layer 7 %lb%`.
We've known that `Layer 7 %lb%` read application-level messages,
therefore,
it becomes impossible to work with the [SSL/TLS](Network-Protection.md#transport-layer-security-tls) **end-to-end
encryption** protocol.
`%lb%`, as an intermediary, can only capture encoded information
that is unhelpful in the balancing process.

```d2

m: Machine {
  grid-columns: 1
  a: Application {
    grid-columns: 1
    l7: Application Layer (L7) {
      style.fill: ${colors.i2}
    }
    http: "payload=a8f5f167f44f4964e6c998dee827110c"
    l7 <- http
  }
}
```

In other words,
we can't use `Layer 7 %lb%` to ensure **complete** end-to-end encryption.
To make it work,
the [SSL/TLS decryption](Network-Protection.md#tls-certificate) must be shifted to the `%lb%` component.
This process is known as `SSL Termination`.

For communications between the `%lb%` and internal services, new connections are established.

```d2

direction: right
s: System {
    lb: Load balancer {
        class: lb
    }
    lb -> lb: 2. Decrypts HTTPS requests (SSL termination)
    sv: Service {
        class: server
    }
    lb -> sv: 3. Forward
}
c: Client {
  class: client
}
c -> s.lb: 1. Send HTTPs requests
```

This can be a security problem when data is exposed at the balancer, not the services themselves.
In certain data governance scenarios,
data is required to remain encrypted until reaching its endpoint,
although the communication between the internal service and the load balancer is encrypted.

Moreover, we should be careful when leveraging an **external** load balancing solution,
as `SSL Termination` will lead to data leakage outside the system.

```d2

direction: right
s: System {
    sv: Service {
        class: server
    }
}
lb: External load balancer {
    class: lb
}
c: Client {
  class: client
}
c -> s.lb: 1. Send HTTPs requests
lb -> lb: 2. SSL termination (data can be leaked here) {
  class: error-conn
}
lb -> s.sv: 3. Forward
```

or manage **data governance**,
it is **safer** because client requests remain encrypted at the L4 load balancer level.

### Layer 4 Load Balancing

`Layer 4 %lb%` refers to the transport layer of the OSI model.
Due to residing at a low level, this type of `%lb%`
does not provide detailed information about application-level data.
It can only make decisions based on two pieces of information: `address` and `port`.

In essence, a `Layer 4 %lb%`
acts as a network router sitting between clients and internal services.
Once a client connects to a server,
it will constantly communicate with the same server as long as the connection is active.
Due to [packet segmentation](https://en.wikipedia.org/wiki/Packet_segmentation),
a large network packet can’t be transmitted as smaller segments.

For example,
an HTTP request is split into two network
segments ([aka TCP messages](https://en.wikipedia.org/wiki/Transmission_Control_Protocol)),
a `Layer 7 %lb%` can integrate with the HTTP
protocol and effectively reassemble these segments to transmit the complete request.

```d2

direction: right
s: System {
    lb: L7 Load Balancer {
        s1: Segment 1
        s2: Segment 2
        h: HTTP request
        s1 -> h
        s2 -> h
    }
    s: Instance 1 {
       class: server
    }
    s1: Instance 2 {
       class: server
    }
    lb.h -> s: 3. Combine and send
}
c: Client {
  class: client
}
c -> s.lb.s1: 1. Send the first segment
c -> s.lb.s2: 2. Send the second segment
```

Meanwhile, a `Layer 4 %lb%` operates without awareness of the HTTP protocol.
It may naively distribute segments to different servers,
which is probably unacceptable.

```d2

direction: right
s: System {
    lb: L7 Load Balancer {
        s1: Segment 1
        s2: Segment 2
    }
    s1: Instance 1 {
       class: server
    }
    s2: Instance 2 {
       class: server
    }
    lb.s1 -> s1: Forwards {
      class: error-conn
    }
    lb.s2 -> s2: Forwards {
      class: error-conn
    }
}
c: Client {
  class: client
}
c -> s.lb.s1: 1. Sends the first segment
c -> s.lb.s2: 2. Sends the second segment
```

The only solution is forwarding all of them to a consistent target
until the connection is disrupted.

```d2

direction: right
s: System {
    lb: L4 Load Balancer {
        class: lb
    }
    s1: Server 1 {
        class: server
    }
    s2: Server 2 {
        class: server
    }
    lb -> s1: 2. Forward continuously {
      style.animated: true 
    }
    lb -- s2: Does not use as long as the connection is alive {
        style.stroke-dash: 3
    }
}
c: Client {
  class: client
}
c -> s.lb: 1. Connect to the system
```

Why should we use a L4 load balancer instead of a L7 version?

- It does not require `SSL Termination`, which can be a security concern.
- An L4 load balancer offers significantly better performance
  because it simply forwards messages without processing them.

Due to the instinctive stickiness, a `Layer 4 %lb%` easily becomes **unbalanced**.
A client may send a lot of traffic to the sticky server,
causing it to become overloaded, while other servers remain underutilised.
It's a good setup for [stateful](Service-Cluster.md#stateful-service) and high-performance services,
such as gaming service.
