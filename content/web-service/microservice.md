---
title: Microservice
weight: 10
---

Let’s begin our journey with a concept that has become ubiquitous in recent years - {{< term ms>}}

## System Scaling

Scaling refers to the process of adjusting a system’s hardware resources. For example:

- When the system experiences high traffic, additional resources must be allocated to
  maintain optimal performance.
- Conversely, if the system is underutilized, reducing resources can help lower costs.

In general, scaling can be categorized into two types: {{< term vs >}} and {{< term hs >}}.

### Vertical Scaling

{{< term vs >}}, also known as **Scaling Up**, involves upgrading a server to improve its performance.

For example:

- If a server lacks memory, additional RAM can be installed.
- If a server operates slowly, upgrading its CPU can enhance performance.

```d2
direction: right
server1: Server (1 CPU) {
  class: server
  width: 100
  height: 100
}
server2: Scaled Server (3 CPUs) {
  class: server
  width: 200
  height: 200
}
server1 -> server2: "Vertical Scale"
```

However, relying on a single server in a large system poses significant challenges:

- **Hardware Limitations**: A server’s capacity cannot be expanded indefinitely.
- **Single point of failure**: If the sole server fails, the entire system may come to a halt.

### Horizontal Scaling

Due to the limitations of {{< term vs >}}, many opt for {{< term hs >}} (aka **Scaling Out**).
Instead of relying on one server,
{{< term hs >}} builds a system by combining multiple **smaller servers** using fewer resources.

For example, consider a system initially built with two servers.
Scaling in this model means increasing the number of servers rather than enhancing a
single server’s resources.
For example, during a traffic spike, adding new servers (e.g., `Server 2` and `Server 3`) can alleviate the load.

```d2
direction: right
c1: "System" {
    server1: Server 1 {
        class: server
        width: 100
        height: 100
    }
}
c2: "Scaled System" {
    server1: Server 1 {
        class: server
        width: 100
        height: 100
    }
    server2: Server 2 {
        class: server
        width: 100
        height: 100
    }
    server3: Server 3 {
        class: server
        width: 100
        height: 100
    }
}
c1 -> c2: Horizontal Scale
```

This approach allows for infinite resource scaling by provisioning separate machines.
It also eliminates the risk of {{< term spof >}},
since if one server fails, others can continue operating.

However, {{< term hs >}} comes with its own trade-offs:

- **Increased Complexity**: Managing multiple machines is inherently much more complex than a single one.
- **Network Problems**: Operating a distributed system requires extensive network communication, which may lead to
  reduced performance, security vulnerabilities, and potential network failures.

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
    {{< filetree/folder name="Payment Module" >}}
      {{< filetree/file name="Request.class" >}}
      {{< filetree/file name="Transaction.class" >}}
    {{< /filetree/folder >}}
    {{< filetree/folder name="Notification Module" >}}
      {{< filetree/file name="Email.class" >}}
      {{< filetree/file name="PushNotification.class" >}}
    {{< /filetree/folder >}}
  {{< /filetree/folder >}}
{{< /filetree/container >}}

However, as the system grows, its flexibility diminishes.
In large systems maintained by multiple teams,
sharing a single codebase can significantly slow development due to the need for tight coordination.
For instance:

- Teams hesitating to modify shared parts due to the risk of unintended consequences.
- Even minor changes cause the entire system to be redeployed.
- **Lock-step Deployment**: One team’s readiness to deploy can be delayed by issues in another team’s code.

To overcome these limitations,
it is essential to minimize inter-team dependencies
and allow teams to work in parallel with clearly defined responsibilities.

### Microservice Architecture

{{< term ms >}} is an **architectural pattern** that decomposes a system into smaller,
independent services, each handling a specific function.

For example, the microservice approach splits the previous system into three
**independent services** and assigns them to different teams.

{{< filetree/container >}}
  {{< filetree/folder name="Account Service" >}}
    {{< filetree/file name="Account.class" >}}
  {{< /filetree/folder >}}
  {{< filetree/folder name="Payment Service" >}}
    {{< filetree/file name="Request.class" >}}
    {{< filetree/file name="Transaction.class" >}}
  {{< /filetree/folder >}}
  {{< filetree/folder name="Notification Service" >}}
    {{< filetree/file name="Email.class" >}}
    {{< filetree/file name="PushNotification.class" >}}
  {{< /filetree/folder >}}
{{< /filetree/container >}}

Ideally, microservices operate in complete **isolation**, sharing no common dependencies such as codebase, databases, or specific technologies.

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
  sa: Microservice A {
    grid-rows: 1
    c: Codebase {
      class: code
    }  
    db: Data schema {
      class: db
    }
  }
  sb: Microservice B {
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

This fundamental isolation empowers teams to manage their services with full autonomy.
It grants them the freedom to select their own technology stacks and to **independently deploy** and test their code.
Consequently, this autonomy significantly **speeds up development cycles** and fosters greater agility.

### Microservice & Monolith

Is a microservice architecture inherently superior to a monolithic one?
The answer depends entirely on context.

In monolithic systems, all modules reside within a single codebase.
This centralized structure makes it simpler to design, develop, and deploy,
particularly for small to medium-sized projects.
Modules communicate directly and efficiently (in-process),
resulting in lower latency and higher performance.

By contrast, microservices are intentionally isolated and must interact **across a network**.
This distributed model introduces added latency, creates more potential points of failure,
and increases the complexity of monitoring, managing, and troubleshooting the system.
Additionally, striving for complete autonomy in microservices can often result in **code duplication** across services.

However, as organizations scale, especially those with dozens or hundreds of developers;
A monolithic architecture can become a bottleneck,
hindering parallel development and making it difficult for teams to work independently.

Ultimately, microservices tend to provide the most value from a **development perspective**,
enabling independent deployments, flexible scaling, and clearer team ownership,
rather than offering benefits for runtime characteristics like raw performance or reliability.

{{< callout type="info" >}}
To be honest, I’m not a fan of **Microservice**, and I know many developers share this sentiment.
Once data leaves my service and crosses the network,
it’s exposed to a host of unpredictable issues that can drain significant amounts of time and energy.

That said, I find working with very large teams even more challenging.
When things break, it’s often unclear where to turn for help,
and I frequently get dragged into problems that fall outside my area of responsibility.
{{< /callout >}}

### Microservice & Horizontal Scaling

A common misconception is that a monolith system must reside on a single server using {{< term vs >}},
while a microservice system always requires {{< term hs >}}.

In reality, the development model is separate from operational strategies.
Both monolithic and microservice systems can be scaled either vertically or horizontally.

## Service Decoupling

### Tight Coupling

A significant challenge in {{< term ms >}} is tight coupling,
where isolated services become **overly dependent** on one another
and behave more like components of a monolithic system.

For example, when a user completes a subscription purchase, the `Subscription Service` first retrieves the necessary account information from the `Account Service`.
After gathering these details, it then notifies the `Account Service` to update the user’s account status accordingly.

```d2
direction: right
system: System {
    s: Subscription Service {
      class: server
    }
    acc: Account Service {
      class: server
    }
    s <- acc: "1. GetAccountInformation()"
    s -> acc: "2. UpdateSubscription()"
}
```

Even though these services reside in separate codebases, they remain **implicitly dependent**.
Changes to the `Account Service`, such as interfaces or logic,
can have unintended consequences for the `Subscription Service`,
requiring coordination and redeployment to prevent runtime errors and limiting service autonomy:

- The more consumers the `Account Service` has, the more coordination needs to happen.
- If the `Account Service` undergoes frequent changes,
  dependent services must constantly cope with it to maintain system integrity.

While **consolidating** services into a single unit might seem like a straightforward solution,
it risks creating a large, monolithic service bringing back the very issues we sought to avoid.
{{< callout type="info" >}}
You may see it's silly to bolt services back after demarcating them.
Nowadays, this occurs a lot in many organizations.
That's because they've initially expected too much and produced excessively complex systems.
{{< /callout >}}

Coupling between services is, to some extent, **unavoidable**.
Our goal should be to **minimize dependencies**
while ensuring that services remain as independent and loosely coupled as possible.

### Loose Coupling

**Loose Coupling**
involves minimizing dependencies between services so that changes in one service have little or no effect on others.

Services can be coupled in several aspects, typically including:

#### Sequential Coupling

**Sequential Coupling** occurs when one service depends on another in a **particular sequence**.

For example, suppose the `Subscription Service` initially calls the `Account Service` to update subscriptions:

```d2
direction: right
a: Subscription Service {
    class: server
}
b: Account Service {
    class: server
}
a -> b: "UpdateSubscription()"
```

Later, if the `Subscription Service` also requires functions for upgrading or
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

a -> b: "UpdateSubscription()"
a -> b: "UpgradePlan()"
a -> b: "CancelPlan()"
```

We observe that the `Subscription Service` grasps the inner logic of the `Account Service`,
every time it needs something,
it dictates the `Account Service` to accommodate that.
The services are tightly coupled with each other,
increasing interdependency and reducing flexibility.

#### Topology Coupling

**Topology Coupling** refers to dependencies that arise from the arrangement
and interconnection of services.
When a service is added or removed, the **overall topology** changes and
can impact other services.

For example, suppose we add a `Notification Service` and a `Fraud Detection Service`,
and the `Subscription Service` is then required to **adapt** to send payment receipts to them:

```d2
direction: right
s1: System {
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
      s: "" {
        class: server
      }
    }
    d: Fraud Detection Service {
      style.stroke-dash: 3
      s: "" {
        class: server
      }
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

Similarly, as new services are introduced or existing ones are removed, the `Subscription Service` must **constantly adapt** to these new topologies.
However, for greater agility and maintainability, the burden of managing such changes should not rest with the `Subscription Service`.
Instead, the responsibility for handling dynamic topology adjustments should belong to the individual components being added or removed.

#### Semantic Coupling

**Semantic Coupling** occurs when services share the same data structures and semantics.

For example, if the `Subscription Service` receives a response from the `Account Service`,
it must understand the structure of that response.
If the `Account Service` modifies the structure, it must notify the `Subscription Service` to prevent errors.

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
  fullName: string
}
b: Account Service {
    class: server
}
a <- r
r -- b
```

Services need to agree on a common contract if they want to interact with each other,
this dependency seems **barely avoidable**.

### Inversion of Control (IoC)

The [Inversion of Control (IoC)](https://en.wikipedia.org/wiki/Inversion_of_control) principle can help
reduce coupling effectively.

Consider a car driving program:

- An `Engine` class controls the wheels.
- A `Controller` is necessary to direct the engine.

Typically, the `Controller` might actively invoke the `Engine`.
In other words, the `Controller` depends on the `Engine`.

```d2
direction: right
e: Engine {
  shape: class
  Drive(direction): ""
}
c: Controller {
  shape: class
  engine: Engine
}
c -> e: Send direction to run {
  class: bold-text
}
```

Using {{< term ioc >}}, we try to invert the dependency.
Now, the `Engine` drives the car by requesting the current direction from the `Controller`;
That means it depends on the `Controller`.

```d2
direction: right
c: Controller {
    shape: class
    GetCurrentDirection(): ""
}
e: Engine {
    shape: class
    controller: Controller
    Drive(): ""
}
e <- c: Get direction to run {
  class: bold-text
}
```

But purely inverting like this is no use,
the dependency and its problems are still there.
We'll see an indirect approach to implement {{< term ioc >}} called {{< term msg >}}.

### Messaging

The {{< term ioc >}} principle can be implemented using {{< term msg >}}.
We essentially build an informative **message broker** with two primary associates:

- **Publishers** publish messages.
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

Integrating {{< term msg >}} into the first coupling example:

- The `Subscription Service` can publish account subscription messages to the broker.
- The `Account Service` can later retrieve these messages to update the associated accounts.

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
messageType: Account subscription
userId: 123
|||
p -- msg: "Publish"
msg -> mq
acc <- mq: "Consume"
```

By the {{< term ioc >}} principle,
the `Account Service` **actively consumes** and processes messages
rather than being directly invoked by another service.
In other words, its role is inverted,
from being called to a caller.

#### Decoupling With Messaging

Beneficially, {{< term msg >}} moves us away from:

- [Sequential coupling](#sequential-coupling): The `Account Service` exposes only a minimal set of interfaces
  and adapts to handle various messages instead.
  Furthermore, the scope of the `Subscription Service` is reduced, granting it more flexibility;
  Like so, even if the `Account Service` fails to process messages,
  the `Subscription Service` continues to develop and deliver without disruption.

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

- [Topology coupling](#topology-coupling): Additional services, such as `Notification Service` and `Fraud Detection Service`,
  can autonomously read messages without requiring any changes from the `Subscription Service`.

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

Nevertheless, we still encounter some dependencies

- Both services depend on {{< term msg >}}. Luckily, the dependency is minimized and barely problematic,
  as **Message Brokers** expose only basic `Publish()` and `Consume()` interfaces that rarely change.
- Both the publisher and consumer adhere to the same message schema, which is [Semantic Coupling](#semantic-coupling).

Occasionally, messaging may result in an **unnecessary overhead** and outweigh the benefits of decoupling.

- The indirect communication model results in **slower performance**,
  making it unsuitable for low-latency workloads.
- Asynchronous communication can lead
  to [temporary inconsistencies]({{< ref "distributed-database#eventual-consistency-level" >}}), since changes aren’t immediately
  reflected across services.
- Debugging may become more challenging as failures are asynchronous and harder to trace.

In summary, while coupling in a microservice architecture can’t be eliminated,
they can be reduced and managed more effectively through {{< term msg >}}.
