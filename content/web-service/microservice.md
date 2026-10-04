---
title: Microservice
weight: 10
---

Let’s begin our journey with a concept that has become ubiquitous in recent years - {{< term ms >}}.

## System Scaling

**Scaling** refers to the process of **adjusting hardware resources** of a system. For example:

- When the system experiences high traffic, additional resources must be allocated to
  maintain optimal performance.
- Conversely, if the system is underutilized, reducing resources can help lower costs.

In general, scaling can be categorized into two types: {{< term vs >}} and {{< term hs >}}.

### Vertical Scaling

{{< term vs >}}, also known as **Scaling Up**, involves upgrading a server to improve its performance.

For example:

- If a server lacks memory, additional RAM can be installed.
- If a server operates slowly, upgrading its CPU (e.g., 1-core to 4-core) can enhance performance.

```d2
direction: right
server1: Server (1-core CPU) {
  class: server
  width: 100
  height: 100
}
server2: Scaled Server (4-core CPU) {
  class: server
  width: 200
  height: 200
}
server1 -> server2: "Vertical scale"
```

However, relying on a single server in a large system poses significant challenges:

- Hardware limitations: A server's capacity cannot be expanded indefinitely.
- **Single point of failure**: If the sole server fails, the entire system may come to a halt.

### Horizontal Scaling

Another method is {{< term hs >}}, also known as **Scaling Out**.

Instead of relying on a single server, {{< term hs >}} builds a system by **combining multiple smaller servers** and distributing the workload among them.

In this model, scaling means adjusting the number of servers rather than changing the resources of a single server.

For example, consider a system initially running on a server, `S1`. During a traffic spike, adding another server, `S2`, can help distribute the increased load instead of requiring an upgrade to `S1`.

```d2
direction: right
c1: "System" {
    server1: S1 {
        class: server
        width: 100
        height: 100
    }
}
c2: "Scaled System" {
    server1: S1 {
        class: server
        width: 100
        height: 100
    }
    server2: S2 {
        class: server
        width: 100
        height: 100
    }
}
c1 -> c2: Horizontal scale
```

This approach reduces the risk of a {{< term spof >}}, since if one server fails, the others can continue operating.

However, {{< term hs >}} comes with its own trade-offs:

- **Increased Complexity**: Managing multiple machines is inherently more complex than managing a single one.
- **Network Overhead**: Distributed systems rely heavily on network communication, which can introduce latency, increase security risks, and create additional points of failure.

### Distributed System

{{< term hs >}} is a fundamental principle behind {{< term ds >}}.
Simply put, a distributed system is a set of machines that closely collaborate over a network
to share resources.

```d2
cluster: "Distributed System" {
  grid-rows: 2
  grid-gap: 100
  s1: Server 1 {
      class: server
  }
  s2: Server 2 {
      class: server
  }
  s3: Server 3 {
      class: server
  }
  s1 <-> s2: {
    style.animated: true
  }
  s2 <-> s3: {
    style.animated: true
  }
  s1 <-> s3: {
    style.animated: true
  }
}
```

Keep this concept in mind! a significant portion of this document will focus on the challenges and solutions
associated with {{< term ds >}}.

## Microservice

Now, let's move the main part - {{< term ms >}}.

### Monolith Architecture

Traditionally, **Monolith Architecture** is the first choice of software engineering.
In this model, all features are implemented within a single codebase and separated as **modules**.
This approach provides simplicity and rapid initial development due to its centralized nature.

For example, a system with three modules might be structured as follows:

{{< filetree/container >}}
  {{< filetree/folder name="Project" >}}
    {{< filetree/folder name="Account Module" >}}
      {{< filetree/file name="Account.class" >}}
    {{< /filetree/folder >}}
    {{< filetree/folder name="Subscription Module" >}}
      {{< filetree/file name="Request.class" >}}
      {{< filetree/file name="SubscriptionPackage.class" >}}
    {{< /filetree/folder >}}
  {{< /filetree/folder >}}
{{< /filetree/container >}}

However, as the system grows, its flexibility diminishes.
In large systems maintained by multiple teams, sharing a single codebase can significantly slow down development because it requires tight coordination among teams. For example:

- **Risk-Averse Changes**: Teams may hesitate to modify shared components due to the risk of unintended consequences.
- **Lock-Step Deployment**: One team’s deployment may be delayed by issues in another team’s code.

To overcome these limitations, it is essential to minimize inter-team dependencies and enable teams to work independently and in parallel, with clearly defined responsibilities.

### Microservice Architecture

{{< term ms >}} is an architectural pattern that decomposes a system into smaller,
independent services, each responsible for a specific function.

For example, a microservice architecture can split the previous system into three
**independent services**, each owned by a different team.

{{< filetree/container >}}
  {{< filetree/folder name="Account Service" >}}
    {{< filetree/file name="Account.class" >}}
  {{< /filetree/folder >}}
  {{< filetree/folder name="Subscription Service" >}}
    {{< filetree/file name="Request.class" >}}
    {{< filetree/file name="SubscriptionPackage.class" >}}
  {{< /filetree/folder >}}
{{< /filetree/container >}}

Ideally, microservices operate with a high degree of **isolation**, minimizing shared dependencies such as codebases, databases, and technology stacks.

```d2
grid-columns: 1
t {
  grid-rows: 1
  horizontal-gap: 250
  class: none
  ta: Team A {
    class: group
  }
  tb: Team B {
    class: group
  }
}
s {
  grid-rows: 1
  class: none
  sa: Account Service {
    grid-rows: 1
    c: Codebase {
      class: code
    }
    db: Data schema {
      class: db
    }
  }
  sb: Subscription Service {
    grid-rows: 1
    c: Codebase {
      class: code
    }
    db: Data schema {
      class: db
    }
  }
}
t.ta -> s.sa: Maintain
t.tb -> s.sb: Maintain
```

This isolation empowers teams to manage their services with greater autonomy.
Teams can choose technology stacks appropriate to their needs and **independently deploy** and test their services.
Consequently, this autonomy can **accelerate development cycles** and enable teams to iterate more independently.

### Microservices vs. Monoliths

Is a microservice architecture inherently superior to a monolithic one?
The answer depends entirely on the context.

In a monolithic system, all modules reside within a single codebase and are typically deployed as a single unit.
This centralized structure is often **simpler to design, develop, test, and deploy,**
particularly for small to medium-sized projects.
Modules can also communicate directly within the same process,
resulting in **lower communication overhead and latency**.

By contrast, microservices are intentionally isolated and often communicate **across a network**.
This distributed model introduces additional latency, creates more potential points of failure,
and increases the complexity of monitoring, managing, and troubleshooting the system.
Furthermore, maintaining strict service autonomy can lead to **code duplication** across services.

However, as organizations grow, especially to dozens or hundreds of developers,
a monolithic architecture can become a development bottleneck.
A large, shared codebase may increase coordination overhead,
complicate deployments, and make it more difficult for teams to develop and release features independently.

Ultimately, microservices tend to provide their greatest benefits from an **organizational and development perspective**.
They enable independent deployments, flexible scaling, and clearer service ownership.
These advantages do not necessarily translate into better runtime characteristics,
such as raw performance or reliability, because distributed systems introduce their own operational costs and failure modes.

{{< callout type="info" >}}
Personally, I'm not a fan of **Microservices**, and I know many developers share this sentiment.
Once data leaves my service and crosses the network,
it becomes subject to a range of failures and uncertainties that can consume significant time and effort to diagnose.

That said, I find working within very large teams even more challenging.
When something breaks, ownership can become unclear,
and I may end up investigating problems that fall well outside my area of responsibility.

For me, that trade-off captures much of the appeal of microservices:
they exchange some technical simplicity for stronger boundaries between teams and their responsibilities.
{{< /callout >}}

### Microservice & Horizontal Scaling

A common misconception is that a monolithic system must run on a single server and rely on {{< term vs >}},
while a microservice architecture inherently requires {{< term hs >}}.

In reality, *architectural style and scaling strategy are separate concerns*.
Both monolithic and microservice systems can be scaled either vertically or horizontally,
depending on their requirements and deployment environments.

## Service Decoupling

### Tight Coupling

A significant challenge in {{< term ms >}} is **tight coupling**,
where isolated services become overly dependent on one another
and begin to behave more like components of a monolithic system.

For example, when a user completes a subscription purchase, the `Subscription Service` first retrieves the necessary account information from the `Account Service`.
After gathering these details, it then notifies the `Account Service` to update the account's status accordingly.

```d2
shape: sequence_diagram
s: Subscription Service {
  class: server
}
acc: Account Service {
  class: server
}
s <- acc: GetAccountInformation()
s -> acc: UpdateStatus()
```

Even though these services reside in separate codebases, they remain **implicitly dependent** on each other.
Changes to the `Account Service`, such as changes to its interfaces or logic,
can have unintended consequences for the `Subscription Service`,
requiring **coordination and redeployment** to prevent runtime errors and thereby limiting service autonomy:

- The more consumers the `Account Service` has, the more coordination is required.
- If the `Account Service` changes frequently,
  dependent services must continually adapt to maintain system integrity.

While **consolidating services into a single unit** might seem like a straightforward solution,
it risks creating a large service and reintroducing the very problems we sought to avoid with a monolithic system.

{{< callout type="info" >}}
It may seem counterproductive to merge services again after deliberately separating them.
Nevertheless, this happens frequently in many organizations.
One common reason is that teams initially overestimate the benefits of decomposition and end up creating excessively complex systems.
{{< /callout >}}

Coupling between services is, to some extent, **unavoidable**.
Our goal should therefore be to **minimize dependencies**
while keeping services as independent and loosely coupled as possible.

### Loose Coupling

**Loose Coupling**
involves minimizing dependencies between services so that changes in one service have little or no effect on others.

Services can be coupled in several ways, typically including:

#### Sequential Coupling

**Sequential Coupling** occurs when one service depends on another in a **particular sequence of interactions**.

For example, suppose the `Subscription Service` initially calls the `Account Service` to update an account's status:

```d2
direction: right
a: Subscription Service {
    class: server
}
b: Account Service {
    class: server
}
a -> b: "UpdateStatus()"
```

Later, if the `Subscription Service` also requires functionality for upgrading or
canceling subscription plans,
the `Account Service` must expose additional functions:

```d2
direction: right
a: Subscription Service {
    class: server
}
b: Account Service {
    class: server
}

a -> b: "UpdateStatus()"
a -> b: "UpgradePlan()"
a -> b: "CancelPlan()"
```

We can see that the `Subscription Service` must understand the internal responsibilities of the `Account Service`.
Whenever it requires additional behavior,
it effectively dictates how the `Account Service` must evolve.
As a result, the two services become tightly coupled,
increasing interdependency and reducing flexibility.

#### Topology Coupling

**Topology Coupling** refers to dependencies that arise from the arrangement
and interconnection of services.
When a service is added or removed, the **overall topology** changes,
potentially affecting other services.

For example, suppose we introduce a `Notification Service` and a `Fraud Detection Service`.
The `Subscription Service` must then **adapt** to send subscription information to these new services:

```d2
direction: down
s1: System {
    direction: right
    acc: Account Service {
      class: server
    }
    p: Subscription Service {
      class: server
    }
    p -> acc
}
s2: Adapted System {
    acc: Account Service {
      class: server
    }
    p: Subscription Service {
      class: server
    }
    n: Notification Service {
      style.stroke-dash: 3
      class: server
    }
    d: Fraud Detection Service {
      style.stroke-dash: 3
      class: server
    }
    p -> acc
    p -> n: Added {
        style.stroke-dash: 3
    }
    p -> d: Added {
        style.stroke-dash: 3
    }
}
s1 -> s2: Changed to
```

Similarly, as new services are introduced or existing ones are removed, the `Subscription Service` must **continually adapt** to changes in the system topology.
However, for greater agility and maintainability, the burden of managing such changes should not rest with the `Subscription Service`.
Instead, responsibility for adapting to topology changes should belong to the individual components being added or removed.

#### Semantic Coupling

**Semantic Coupling** occurs when services depend on shared data structures and semantics.

For example, if the `Subscription Service` receives a response from the `Account Service`,
it must understand the structure and meaning of that response.
If the `Account Service` modifies the structure, the `Subscription Service` must be updated accordingly to prevent errors.

```d2
grid-rows: 1
horizontal-gap: 100
direction: right
a: Subscription Service {
    class: server
}
r: Response {
  shape: sql_table
  id: string
  status: free|plus|pro
}
b: Account Service {
    class: server
}
a <- r
r -- b
```

Services must agree on a common contract in order to interact with one another,
so this form of dependency is **difficult to avoid entirely**.

### Inversion of Control (IoC)

The [Inversion of Control (IoC)](https://en.wikipedia.org/wiki/Inversion_of_control) principle can help
reduce coupling between services.

Consider the previous example.
Suppose users query their status through the `Account Service`.
The `Subscription Service` actively controls updates to account status.
In other words, the `Subscription Service` depends on the `Account Service`.

```d2
direction: right
a: Subscription Service {
    class: server
}
b: Account Service {
    class: server
}
u: User {
  class: client
}
a -> b: "UpdateStatus()"
u <- b: "GetStatus()"
```

Using {{< term ioc >}}, we can attempt to invert this dependency.
Instead, the `Account Service` requests the current subscription from the `Subscription Service` and determines the account status itself.
The `Subscription Service` no longer needs to update account status directly or depend on the `Account Service`.
Instead, the `Account Service` becomes the initiator of the interaction.

```d2
direction: right
a: Subscription Service {
    class: server
}
b: Account Service {
    class: server
}
u: User {
  class: client
}
a -> b: "GetCurrentSubscription()"
u <- b: "GetStatus()"
```

However, simply inverting the direction of the call does not solve the underlying problem.
The dependency, along with its associated drawbacks, still exists.
Next, we'll examine an indirect approach to implementing {{< term ioc >}} using {{< term msg >}}.

### Messaging

The {{< term ioc >}} principle can be implemented using {{< term msg >}}.
In this model, we introduce a **message broker** and three primary roles:

- **Publishers** publish messages.
- **Broker** stores and distributes messages.
- **Consumers** consume and process messages.

```d2
direction: right
p: Publisher {
  class: server
}
m: Message Broker {
  class: mq
}
c: Consumer {
  class: server
}
p -> m: Publish
c <- m: Consume
```

Applying {{< term msg >}} to the previous coupling example:

- The `Subscription Service` publishes subscription-related messages to the broker.
- The `Account Service` later consumes these messages and updates the corresponding accounts.

```d2
direction: right
acc: Account Service {
    class: server
}
p: Subscription Service {
    class: server
}
mq: Message Broker {
    class: mq
}
msg: |||yaml
messageType: AccountSubscription
userId: 123
|||
p -- msg: "Publish"
msg -> mq
acc <- mq: "Consume"
```

Following the {{< term ioc >}} principle,
the `Account Service` now **actively consumes** and processes messages
instead of being directly invoked by another service.
As a result, neither the `Subscription Service` nor the `Account Service` directly depends on the other.

#### Decoupling With Messaging

{{< term msg >}} helps reduce several forms of coupling:

- [Sequential coupling](#sequential-coupling): The `Account Service` exposes only a minimal set of interfaces
  and adapts internally to handle different messages.
  At the same time, the responsibilities of the `Subscription Service` are reduced, giving it greater flexibility.
  For example, even if the `Account Service` temporarily fails to process messages,
  the `Subscription Service` can continue operating without being directly disrupted.

```d2
direction: right
system: System {
    acc: Account Service {
        class: server
    }
    p: Subscription Service {
        class: server
    }
    mq: Message Broker {
        class: mq
    }
    p -> mq: Subscription Message
    p -> mq: Cancellation Message
    p -> mq: Upgrade Message
    acc <- mq: Consume
}
```

- [Topology coupling](#topology-coupling): Additional services, such as the `Notification Service` and `Fraud Detection Service`,
  can independently consume messages without requiring any changes to the `Subscription Service`.

```d2
direction: right
system: System {
    acc: Account Service {
        class: server
    }
    p: Subscription Service {
        class: server
    }
    mq: Message Broker {
        class: mq
    }
    n: Notification Service {
        class: server
    }
    d: Fraud Detection Service {
        class: server
    }
    p -> mq
    acc <- mq: Consume
    n <- mq: Consume
    d <- mq: Consume
}
```

Nevertheless, some dependencies still remain:

- Both services depend on {{< term msg >}}. Fortunately, this dependency is relatively small and rarely problematic,
  as **Message Brokers** typically expose simple `Publish()` and `Consume()` interfaces that change infrequently.
- Both publishers and consumers must adhere to the same message schema, resulting in [Semantic Coupling](#semantic-coupling).

In some situations, however, messaging introduces **unnecessary overhead** that may outweigh the benefits of decoupling:

- Indirect communication can result in **higher latency**, making messaging unsuitable for certain low-latency workloads.
- Asynchronous communication can lead
  to [temporary inconsistencies]({{< ref "distributed-database#eventual-consistency-level" >}}), since changes are not immediately
  reflected across services.
- Debugging can become more challenging because failures occur asynchronously and may be harder to trace across services.

In summary, while coupling in a microservice architecture cannot be eliminated entirely,
it can be reduced and managed more effectively through {{< term msg >}}.
