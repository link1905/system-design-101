---
title: Service Cluster
weight: 20
next: communication-protocols
params:
  math: true
---

In the previous topic, we discussed some fundamental aspects of {{< term ms >}}.
In this topic, we'll explore how to operate microservices effectively to
build a reliable system.

## Cluster

Traditionally, we build and run a service as a *single process*.
This approach may work initially.
However, if the process or its host machine crashes, the entire service becomes unavailable.

```d2
sv-no: Machine {
  Process {
    class: process
  }
}
```

Therefore, to improve resilience,
we need to deploy the service as a *cluster of multiple instances (processes)*,
ideally distributed across *different machines*.
This setup allows the service to remain operational even if some instances or machines fail.

For example, consider a service cluster with multiple instances distributed across two machines.
If one instance, or even its entire machine, fails, the remaining instances can continue providing the service.

```d2
sv-no: Service {
  grid-rows: 1
  m1: Machine 1 {
    grid-rows: 1
    direction: right
    Instance 1 {
      class: process
    }
    Instance 2 {
      class: process
    }
  }
  m2: Machine 2 {
    grid-rows: 1
    direction: right
    Instance 3 {
      class: process
    }
    Instance 4 {
      class: process
    }
  }
}
```

From this point onward, when we refer to a **service**,
we mean *a cluster comprising multiple instances*.

## Service State

To operate a service reliably, we need to understand its state.
Two key metrics help us assess it: {{< term health >}} and {{< term av >}}.

### Health

{{< term health >}} refers to an instance's ability to perform its intended tasks.
An instance generally reports one of two health states:

- **Healthy (Up)**: It is ready to accept and handle requests.
- **Unhealthy (Down)**: The instance has encountered a problem, such as a database disconnection or hardware failure, and can no longer serve requests.

#### Health Interface

Typically, an instance exposes a health interface that reports its current status:

- Consumers, typically other services, can perform {{< term hc >}} to ensure they are
  communicating with a healthy instance.

```d2
shape: sequence_diagram
client: Consumer {
  class: client
}
service: Service Instance 1 {
  class: server
}
client -> service: 1. "Check '/health'"
service -> client: 2. Unhealthy {
  class: error-conn
}
client -> client: 3. Cancel the request because the instance is unhealthy
```

- The system can also use this interface to *isolate* unhealthy instances.

#### Heartbeat Mechanism

A common technique for isolating unhealthy instances is the heartbeat mechanism.
Essentially, a health checker, such as a [Load Balancer]({{< ref "load-balancer" >}})
or [DNS Server](https://en.wikipedia.org/wiki/Domain_Name_System), *periodically checks*
the health interfaces of service instances to identify and exclude faulty ones.

For example, suppose a health checker verifies the health of every instance every 5 seconds.
If an instance is found to be unhealthy, the checker stops forwarding traffic to it.

```d2
direction: right
s: System {
  s1: Instance 1 (Healthy) {
    class: server
  }
  s2: Instance 2 (Unhealthy) {
    class: generic-error
  }
  c: Health Checker {
    class: checker
  }
  c -> s1: Check health {
   style.animated: true
  }
  c -> s2: Check health {
    class: error-conn
   style.animated: true
  }
}
client: Consumer {
  class: client
}
client -> s.c: Only access Instance 1
```

### Service Availability

{{< term av >}} is a critical metric that represents the accessibility of a service from the *user's perspective*.

For example, suppose a service depends on another service that is currently unavailable.
From a technical perspective, the dependency is the source of the problem,
while the service itself may still be running and healthy.
However, users do not care which component failed. They simply observe that the requested functionality is unavailable.

```d2
direction: right
u: User {
  class: client
}
s: Service {
  class: server
}
t: Target Service {
  class: generic-error
}
u -> s: Unavailable
s -> t {
  class: error-conn
}
```

This metric is essential when defining a [Service Level Agreement (SLA)](https://en.wikipedia.org/wiki/Service-level_agreement).
Availability is commonly calculated in two ways:

#### Time-based Availability

The first approach calculates the proportion of *uptime* relative to the total operating time
within a given period, typically a year:

$Availability = \frac{Uptime}{Uptime + Downtime}$

For example, if a service runs for one year with approximately `3 days` of downtime,
its availability is:

$Availability = \frac{362}{362 + 3} \approx 99\\%$

This approach assumes that requests are distributed relatively evenly over time.
However, it is less sensitive to *short but high-impact outages*.
A service may receive significantly more traffic during an outage than usual,
causing the calculated availability to differ from the actual user experience.

#### Request-based Availability

**Request-based Availability** calculates availability based on the proportion of
*successful requests* relative to the total number of requests:

$Availability = \frac{Successful\ requests}{Total\ requests}$

For example, suppose a service successfully handles 1000 out of 1010 requests:

$Availability = \frac{1000}{1010} = 99\\%$

This approach can provide a more representative result, but it may also introduce *bias*:

- Highly active users have a disproportionately large impact on the metric,
  potentially overshadowing the experience of less active users.
- During outages, users often retry failed requests repeatedly,
  generating additional traffic that can make availability appear significantly worse.

Therefore, **Request-based Availability** is generally less suitable for public-facing services
and more appropriate for controlled internal workloads.

#### Aggregate Availability

In the {{< term ms >}} topic,
we discussed several forms of [design-time coupling]({{< ref "microservice#loose-coupling" >}})
that can negatively affect the development process.
In production environments,
*runtime dependencies* also emerge when services communicate over a network:

- **Location Coupling**: A service must know how to locate another service, such as through an IP address or domain name.
- **Availability Coupling**: When one service synchronously depends on another,
  its own availability is affected by the availability of that dependency.
  This is the critical form of coupling we'll focus on here.

For example,
the `Subscription Service` cannot complete its task without successfully communicating with the `Account Service`.
Therefore, if the `Account Service` becomes unavailable,
the `Subscription Service` is also affected.

```d2
direction: right
a: Subscription Service (Unavailable) {
   class: server
}
b: Account Service (Unavailable) {
   class: generic-error
}
a -> b {
    class: error-conn
}
```

Thus, the effective availability of a service depends on the availability of all services required to complete its operation.

$Availability = S (self) \times S1 \times S2 \times ... \times Sn$

```d2
direction: down
S: Service {
  class: server
}
S1: Service 1 {
  class: server
}
S2: Service 2 {
  class: server
}
Sn: Service n {
  class: server
}
S -> S1
S -> S2
S -> "..."
S -> Sn
```

This interdependency can become a serious problem when service communication forms a complex dependency graph.
Some services may become a {{< term spof >}},
meaning that their failure can disrupt a large portion, or even all, of the system.
For example, in this dependency graph, if `C` becomes unavailable,
the entire request chain may stop functioning.

```d2
direction: right
a: Service A {
   class: server
}
b: Service B {
   class: server
}
c: Service C {
   class: generic-error
}
d: Service D {
   class: server
}
a -> c
b -> c
c -> d
```

#### Availability Decoupling

We previously discussed how {{< term msg >}} can decouple services in a {{< term ms >}} architecture.
Fortunately, {{< term msg >}} can also reduce coupling at runtime.

For example,
the `Subscription Service` may become unavailable when the `Account Service` is down
because it cannot guarantee that requests will reach the `Account Service`.

```d2
direction: right
a: Subscription Service {
   class: server
}
b: Account Service {
   class: generic-error
}
a -> b {
  class: error-conn
}
```

By introducing {{< term msg >}}, the `Subscription Service` can publish a message and continue its workflow without waiting for an immediate response.
The `Account Service` can process the message later when it becomes available again.
As a result, the availability of the `Account Service` no longer directly determines the availability of the `Subscription Service`.

This approach decouples the services at runtime, improving the system's resilience and flexibility.

```d2
direction: right
m: Message Broker {
   class: mq
}
a: Subscription Service {
   class: server
}
b: Account Service {
   class: generic-error
}
a -> m: Publish {
  style.animated: true
}
b <- m: Consume {
  style.animated: true
}
```

This approach is particularly valuable in systems containing many services.
Services no longer depend directly on one another's availability;
even if some services fail temporarily, the remaining services can continue operating.

For example, in the following diagram,
the effective availability of `Service A` is:

$(SA) = SA (self) \times SB \times SC$

```d2
direction: right
a: Service A {
   class: server
}
b: Service B {
   class: server
}
c: Service C {
   class: server
}
a -> b
a -> c
```

With {{< term msg >}},
the equation becomes:

$SA = SA (self) \times MessageBroker$

```d2
direction: right
m: Message Broker {
   class: mq
}
a: Service A {
   class: server
}
b: Service B {
   class: server
}
c: Service C {
   class: server
}
a <-> m
b <-> m
c <-> m
```

In practice, we have *shifted* the complex interdependencies between services to the message broker.
The overall architecture becomes easier to manage because dependencies are now concentrated around the broker.

However, this introduces new trade-offs:

- The message broker can become a critical {{< term spof >}}, so it must itself be highly available and fault-tolerant.
- {{< term msg >}} also introduces several additional challenges, which were discussed in detail in the [previous topic]({{< ref "microservice#decoupling-with-messaging" >}}).

## Cluster Types

Service clusters can generally be categorized into two types: {{< term sl >}} and {{< term sf >}}.

- Stateless services contain application logic but do not retain state between requests.
- Stateful services *retain state*, causing requests or connections to depend on information stored by a particular instance.

### Stateless Service

{{% include "stateless-service" %}}

### Stateful Service

Unlike a stateless service, a stateful service contains instances that may retain *local state*.
As a result, different instances of the same service may *behave differently* depending on the state they currently hold.

Stateful services are often associated with *real-time features*,
which require maintaining persistent client connections so the server can push messages to connected clients.
A common example is a chat application that maintains client connections,
typically through [WebSocket]({{< ref "communication-protocols#websocket" >}}), for real-time messaging.

Consider a cluster containing two instances.
If `Client A` connects to `Instance 1` while `Client B` connects to `Instance 2`,
the two clients cannot communicate directly through local connections alone because each instance manages its own **socket connections**.

```d2
grid-rows: 1
c: Clients {
  grid-rows: 2
    ca: Client A {
      class: client
    }
    cb: Client B {
      class: client
    }
}
system: System {
  grid-rows: 2
    s1: Instance 1 {
      class: server
    }
    s2: Instance 2 {
      class: server
    }
}
c.ca <-> system.s1: WebSocket {
  style.animated: true
}
c.cb <-> system.s2: WebSocket {
  style.animated: true
}
```

#### Scaling Problem

Stateful services are more difficult to scale and are generally best avoided when state can reasonably be externalized.
Simply increasing the number of instances is not sufficient.
Additional mechanisms are required to coordinate and *share state* across instances.

#### Centralized Cluster

One approach is to introduce a *shared store* that is accessible to all instances.

In the chat example, we introduce a shared component called the `Presence Store`,
which maintains a mapping between users and the server instances they are currently connected to.
Whenever a user connects to the system, the corresponding server instance creates or updates a *presence record* in this store.

By querying the shared store, service instances can determine where a particular user is connected
and forward messages to the appropriate instance.

```d2
grid-rows: 1
c: Clients {
  grid-columns: 1
  vertical-gap: 100
  c1: Client 1 {
    class: client
  }
  c2: Client 2 {
    class: client
  }
}

s: Cluster {
  grid-columns: 1
  vertical-gap: 100
  s1: Instance 1 {
    class: server
  }
  s2: Instance 2 {
    class: server
  }
  s1 <-> s2: Forward {
    style.animated: true
  }
}
p: Presence Store {
  grid-columns: 1
  class: cache
  t: |||yaml
  Client 1: Instance 1
  Client 2: Instance 2
  |||
}
c.c1 -> s.s1
c.c2 -> s.s2
s.s1 <-> p.t
s.s2 <-> p.t
```

The simplicity of this design makes it a practical and commonly preferred solution.

However, this approach can *reduce availability* because the service now depends on an external shared store.
Suppose we are developing a *low-level service*, such as a data store itself.
In that case, the service may need to operate as a terminal component without relying on another centralized dependency.

In the [Distributed Database]({{< ref "distributed-database" >}}) topic,
where a database cluster is treated as a stateful service,
we'll discuss decentralized architectures in which instances can coordinate and operate without relying on a central store.
