---
title: Service Cluster
weight: 20
next: communication-protocols
params:
  math: true
---

In the previous topic, we've talked about some basic facets of {{< term ms >}}.
Jumping to this one, we'll see how to operate microservices effectively to
ensure a reliable system.

## Cluster

Traditionally, we build and run a service as a **single process**.
This method might work initially,
nonetheless, if the machine or process crashes, the service is also down.

```d2
sv-no: Machine {
  Process {
    class: process
  }
}
```

Therefore, for resilience,
we need to deploy the service as a **cluster of multiple instances (processes)**,
ideally distributed across **different machines**.
This setup ensures that the service remains operational even if some instances crash.

For example, consider running a service cluster of two instances residing in different machines.
If one instance (or its machine) fails, the other instance still provides the service.

```d2
sv-no: Service {
  m1: Machine 1 {
   Instance 1 {
      class: process
   }
   Instance 2 {
      class: process
   }
  }
  m2: Machine 2 {
   Instance 3 {
      class: process
   }
  }
}
```

From this point onward, when we refer to a **service**,
we imply a service cluster comprising multiple instances.

## Service State

To maintain a service, we need to know its state.
Two key metrics are used to assess this: {{< term health >}} and {{< term av >}}.

### Healthy

{{< term health >}} refers to an instance's ability to perform its intended tasks.
The instance must determine its health state in one of two options:

- **Heathy** (**Up**): it's willing to accept and handle requests from users.
- **Unhealthy** (**Down**): The instance has encountered a problem (e.g., a disconnect from the database, hardware failure...)
  and is no longer serving requests.

#### Health Interface

Typically, an instance exposes a health interface, reporting its status:

- Consumers (end-users or other services) can perform {{< term hc >}} to confirm they are
  interacting with a healthy instance.

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
client -> service: 3. Cancel the request because the instance is unhealthy
```

- The system can also use this interface to **isolate** unhealthy instances.

#### Heartbeat Mechanism

A common technique to isolate unhealthy instances is the heartbeat mechanism.
In essence, a health checker ([Load Balancer]({{< ref "load-balancer" >}}) or
[DNS](https://en.wikipedia.org/wiki/Domain_Name_System)) **periodically** accesses the health interfaces
of instances to filter out the faulty ones.

For example, a health checker verifies the health status of instances every 5 seconds.
If an instance is found to be unhealthy, the checker will stop forwarding traffic to it.

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

{{< term av >}} is a critical metric that indicates the accessibility of a service from the **user perspective**.

For example, a service depends on another service that is currently down.
From the technical perspective, the target is the source of problem,
the service itself is still operational and healthy;
However, users don't care; they just request and see that the service is unavailable.

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

This metric is essential for outlining [Service Level Agreement (SLA)](https://en.wikipedia.org/wiki/Service-level_agreement).
Typically, it's calculated in two ways

#### Time-based Availability

The first approach is using the percentage between the **uptime** and total time over a period of time,
typically a year:

$Availability = \frac{Uptime}{Uptime + Downtime}$

For example, a service runs for a year with a downtime of approximately `3 days`,
the availability would be:

$Availability = \frac{362}{362 + 3} \approx 99\\%$

This approach assumes that requests are equally distributed over time.
However, it's less sensitive to **short outages**,
a service may receive much higher traffic than usual within downtimes and
make the final availability away from the real experience.

#### Request-based Availability

**Request-based Availability** suggests calculating availability based on the number of
**successful requests** compared to the total numbers of requests:

$Availability = \frac{Successful\ requests}{Total\ requests}$

For example, we have a service that successfully handles 1000 out of 1010 requests:

$Availability = \frac{1000}{1010} = 99\\%$

This approach gives a more precise result, but it possibly generates **bias**.

- More active users will affect the availability strongly,
  they can suppress the experience of less active ones.
- During downtime, users tend to retry and make a lot of junk requests,
  making availability much worse.

Thus, **Request-based Availability** is rarely used in public-serving services,
and more suited to internal workloads.

#### Aggregate Availability

In the {{< term ms >}} topic,
we've discussed some types of [design-time coupling]({{< ref "microservice#loose-coupling" >}})
negatively impact the development process.
In the operational environment,
**runtime dependencies** also emerge when services communicate over a network:

- **Location Coupling**: services need to know the address (IP, domain name...) of others.
- **Availability Coupling**: when a service calls another service,
  its availability will be impacted by that service.
  Let's focus on this critical one!

For instance,
the `Subscription Service` cannot complete its task without successfully communicating with the `Account Service`.
Therefore, if the `Account Service` is unavailable,
the `Subscription Service` will also be affected.

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

Thus, the final availability of a service is an aggregation from all relevant services.

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

This interdependency can be a huge issue when the communication between services forms a complex graph.
Some services will become a {{< term spof >}},
that means its corruption halts the entire system.
For example, in this map, if `D` or `E` is unavailable,
the entire chain stops working unexpectedly.

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
d: Service D {
   class: generic-error
}
e: Service E {
   class: server
}
a -> b
b -> d
c -> d
d -> e
```

#### Availability Decoupling

We've discussed the role of {{< term msg >}} to decouple a {{< term ms >}} system.
Helpfully, {{< term msg >}} also means in the runtime environment.

For example,
the `Subscription Service` becomes unavailable too as
it's afraid that the `Account Service` will miss its requests.

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

By introducing {{< term msg >}}, the `Subscription Service` can simply publish messages and continue its workflow without waiting for a response.
The `Account Service` can then process these messages whenever it is available, ensuring its availability no longer directly impacts the `Subscription Service`.
This approach decouples the services, promoting greater resilience and flexibility in the system.

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
a -> m: Publish continuously
b <- m: Consume
```

This approach is particularly beneficial in environments with numerous services. Services do not rely on each other;
even if some of them fail, the rest continue to function.

For example, in this diagram,
the final availability of `Service A` would be: $(SA) = SA (self) \times SB \times SC \times SD$

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
d: Service D {
   class: server
}
a -> b
a -> c
b -> d
c -> d
```

With {{< term msg >}},
it becomes $SA = SA (self) \times MessageBroker$

```d2
direction: down
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
d: Service D {
   class: server
}
a <-> m
b <-> m
c <-> m
d <-> m
```

Actually, we've **shifted** the complex interdependency to the broker.
Now, the system looks more manageable as the dependencies only end with one connection,
not a harmfully long chain.
Probably, the message broker becomes a dangerous {{< term spof >}},
requiring it to be highly available and fault-tolerant.

## Cluster Types

Basically, service clusters are categorized into two types: {{< term sl >}} and {{< term sf >}}.

- Stateless services only contain logic and do not store state between requests.
- Stateful services **store state** and make requests relate to each other.

### Stateless Service

A stateless service operates with instances that share **identical logic** and
don’t retain local data.

For instance, consider two instances of the `Account Service`.
Both instances query the same database and are designed to return identical results.
The specific instance a client connects to doesn't matter because each
instance has the same logic and data access patterns,
ensuring a **consistent response**.

```d2
direction: right
system {
    db: Database {
      class: db
    }
    s: Account Service {
      s1: Instance 1 {
        class: server
      }
      s2: Instance 2 {
        class: server
      }
    }
    s.s1 <- db: Query
    s.s2 <- db: Query
    c: Client {
        class: client
    }
    c <- s.s1: Get account data
    c <- s.s2: Get account data
}
```

This consistent behavior makes
scaling a stateless service be a piece of cake,
we simply increase or decrease the number of identical instances.

### Stateful Service

A stateful service, unlike a stateless one, has instances that may store **local state**.
As a result, different instances of the service may **behave differently** based on their local state.

Stateful services are often paired with **real-time features**,
which require maintaining client connections to push messages from the service side.
A common example of this is a chat application that holds client connections
(typically using [WebSocket]({{< ref "communication-protocols#websocket" >}})) for real-time messaging.

Consider a cluster of two instances,
if `Client A` connects to `Instance 1` and `Client B` connects to `Instance 2`,
they cannot chat with each other because different instances handle their own **socket connections**.

```d2
direction: right
c: Clients {
    ca: Client A {
      class: client
    }
    cb: Client B {
      class: client
    }
}
system: System {
    s1: Instance 1 {
      class: server
    }
    s2: Instance 2 {
      class: server
    }
}
c.ca <-> system.s1: Connecting
c.cb <-> system.s2: Connecting
```

#### Scaling Problem

Stateful services are more challenging to scale and generally **recommended avoiding**.
Simply increasing the number of instances is insufficient,
additional strategies are required to manage and **share state** across instances.

#### Centralized Cluster

The first approach is building a **shared store** between instances.

In the chat example, we introduce a shared component known as the `Presence Store`, which manages the mapping of users to their current server instances.
Whenever a user connects to the system, their server instance creates or updates a **presence record** in this store.
Service instances can effectively determine the location of any user,
enabling them to forward messages directly to the appropriate instance.

```d2
grid-columns: 3
horizontal-gap: 300
c: Clients {
  grid-columns: 1
  vertical-gap: 150
  class: none
  c1: Client 1 (C1) {
    class: client
  }
  c2: Client 2 (C2) {
    class: client
  }
}

s: Cluster {
  grid-columns: 1
  vertical-gap: 150
  s1: Instance 1 (I1) {
    class: server
  }
  s2: Instance 2 (I2) {
    class: server
  }
  s1 <-> s2: Forward {
    style.animated: true
  }
}
p: Presence Store {
  class: none
  grid-columns: 1
  t: |||yaml
  Client 1: Instance 1
  Client 2: Instance 2
  |||
  s: "Presence Store" {
      class: cache
  }
}
c.c1 -> s.s1
c.c2 -> s.s2
s.s1 <-> p.s
s.s2 <-> p.s
```

The simplicity of this approach makes it the preferred solution in many solutions.

However, this way **limits availability**.
When processing a message, the instance must rely on the connection store,
adversely impacting its availability.

#### Decentralized Cluster

To maximize availability, the system can eliminate the connection store and **deterministically map** users to instances.

Each instance is responsible for a specific group of users;
When a user connects to the system, it will be assigned to the owner instance.
For example, we define groups of users with the modulo operation `user id % the number of instances`.

```d2
grid-rows: 3
g1: "" {
    class: none
    grid-columns: 3
    horizontal-gap: 200
    c1: Client 1 (id = 1) {
        class: client
    }
    c3: Client 3 (id = 3) {
        class: client
    }
    c2: Client 2 (id = 2) {
      class: client
    }
}

s: Cluster {
  grid-rows: 1
  horizontal-gap: 300
  s1: Instance 1 {
    class: server
  }
  s2: Instance 0 {
    class: server
  }
}
g1.c1 -> s.s1: '1 % 2 = 1' {
  class: bold-text
}
g1.c3 -> s.s1: '3 % 2 = 1' {
  class: bold-text
}
g1.c2 -> s.s2: '2 % 2 = 0' {
  class: bold-text
}
```

Now, with messages containing `user id`,
instances can quickly specify where to forward them.
The cluster is far cleaner without any dependency,
the final availability is bounded around proprietary instances.

```d2
grid-rows: 3
g1: "" {
    class: none
    grid-columns: 3
    horizontal-gap: 200
    c1: Client 1 (id = 1) {
        class: client
    }
    c3: Client 3 (id = 3) {
        class: client
    }
    c2: Client 2 (id = 2) {
      class: client
    }
}

s: Cluster {
  grid-rows: 1
  horizontal-gap: 300
  s1: Instance 1 {
    class: server
  }
  s2: Instance 0 {
    class: server
  }
}
g1.c1 -> s.s1: '1 % 2 = 1' {
  class: bold-text
}
g1.c3 -> s.s1: '3 % 2 = 1' {
  class: bold-text
}
g1.c2 -> s.s2: '2 % 2 = 0' {
  class: bold-text
}
m: "" {
  class: none
  horizontal-gap: 300
  m1: "Message 1" {
    content: |||yaml
    userId: 1
    content: Hello client 1!
    |||
  }
  m2: "Message 2" {
    content: |||yaml
    userId: 2
    content: Hello client 2!
    |||
  }
}
m.m1 -> s.s1: '1 % 2 = 1' {
  class: bold-text
}
m.m2 -> s.s2: '2 % 2 = 0' {
  class: bold-text
}
```

However, this model is more complex than it appears.
It’s extremely challenging to develop and maintain:

- How do we track cluster information consistently across instances without relying on a central store?
- How can we adapt to a dynamic number of instances as they scale up or down?
- How do we ensure the cluster stays operational when some instances fail?

These questions highlight the inherent complexity of decentralized communication.
We’ll explore these challenges in greater depth in the [Distributed Database]({{< ref "distributed-database" >}}) topic, where a database cluster is treated as a stateful service.
