---
title: Microservice
weight: 10
---

Let’s begin our journey with a concept that has become ubiquitous in recent years — {{< term ms>}}

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

Keep this concept in mind — a significant portion of this document will focus on the challenges and solutions
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
independent services—each handling a specific function.

For example, the microservice approach splits the previous system into three
**independent services**
and assigns them to different teams.

{{< filetree/container >}}
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
{{< /filetree/container >}}

Ideally, microservices are **isolated** and share no common dependencies.
This isolation allows teams to manage their services autonomously,
choose their own technology stacks, and independently deploy and test their code,
thereby speeding up development cycles.

### Microservice & Monolith

Is microservice inherently better than monolith architecture?
It depends on the context.

In a monolithic system, components communicate via **native function calls**.
In contrast, microservices should be isolated and broadly communicate **over a network**,
which introduces additional latency, more potential points of failure,
and increased complexity in testing and troubleshooting.
Moreover,
achieving complete autonomy in a microservice environment probably
leads to **duplication of code**.

However, in large-scale systems developed by dozens or hundreds of people,
the monolith approach can create bottlenecks and impede parallel development.
Briefly, {{< term ms >}} offer significant benefits from the **development perspective**
rather than runtime factors such as performance or reliability.

For example, if a monolithic system is slow under high traffic,
migrating to the microservice model is not a way of improving performance,
because network calls are not supposed to match the efficiency of native calls.

{{< callout type="info" >}}
Honestly, I’m not a big fan of **Microservice**, and I know many developers feel the same.
Once data leaves my service and travels over the network,
it opens the door to a host of unpredictable issues that eat up my time and energy.

That said, I dislike working with plenty of people even more.
When something breaks, I often have no idea who to turn to for answers.
And worse, I keep getting pulled into questions and problems that aren’t even part of my scope.
{{< /callout >}}

### Microservice & Horizontal Scaling

A common misconception is that a monolith system must reside on a single server using {{< term vs >}},
while a microservice system always requires {{< term hs >}}.

In reality, the development model is separate from operational strategies.
Both monolithic and microservice systems can be scaled either vertically or horizontally.

### Microservice Design

Designing a microservice-based system is a complex challenge.
Errors in the design process can result in an overly complicated architecture
that fails to deliver the benefits.
This concern is overwhelming for an open chapter,
we will see it in detail in [a later chapter](Microservice-Decomposition.md).

## Service Decoupling

### Tight Coupling

A significant challenge in {{< term ms >}} is tight coupling,
where isolated services become **overly dependent** on one another
and behave more like components of a monolithic system.

For example, when a user completes a subscription purchase,
the `Payment Service` calls the `Account Service` to update the account:

```d2
direction: right
system: System {
    acc: Account Service {
      class: server
    }
    p: Payment Service {
      class: server
    }
    p -> acc: 1. Update account subscription
}
```

Even though these services reside in separate codebases, they remain **implicitly dependent**.
Changes to the `Account Service`, such as interfaces or logic,
can have unintended consequences for the `Payment Service`,
requiring coordination and redeployment to prevent runtime errors and limiting service autonomy.

- The more consumer the `Account Service` has, the more coordination needs to happen.
- If the `Account Service` undergoes frequent changes,
  dependent services must constantly cope with it to maintain system integrity.

While **consolidating** services into a single unit might seem like a straightforward solution,
it risks creating a large, monolithic service — bringing back the very issues we sought to avoid.
{{< callout type="info" >}}
You may see it's silly to bolt services back after demarcating them.
Nowadays, this occurs a lot in many organizations.
That's because they've expected too much and produced excessively complex systems,
or more simply, they don't have enough workforces after downsizing.
🥺
{{< /callout >}}

Coupling between services is, to some extent, **unavoidable**.
Our goal should be to **minimize dependencies**
while ensuring that services remain as independent and loosely coupled as possible.

### Loose Coupling

**Loose Coupling**
involves minimizing dependencies between services so that changes in one service have little or no effect on others.
Services can be coupled in several aspects, including:

#### Temporal Coupling

**Temporal (or Sequential) Coupling** occurs when one service depends on another in a **particular sequence**.

For example, suppose the `Payment Service` initially calls the `Account Service` to update premium accounts:

```d2
direction: right
a: Payment Service {
    class: server
}
b: Account Service {
    class: server
}
a -> b: "UpdatePremium()"
```

Later, if the `Payment Service` also requires functions for upgrading or
canceling subscription plans,
the `Account Service` must expose additional functions:

```d2
direction: right
a: Payment Service {
    class: server
}
b: Account Service {
    class: server
}
a -> b: "UpdatePremium()"
a -> b: "UpgradePremiumPlan()"
a -> b: "CancelPremium()"
```

We observe that the `Payment Service` grasps the inner logic of the `Account Service`,
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
and the `Payment Service` is then required to **adapt** to send payment receipts to them:

```d2
direction: right
s1: System {
    acc: Account Service {
      class: server
    }
    p: Payment Service {
      class: server
    }
    p -> acc
}
s2: Adapted System {
    acc: Account Service {
      class: server
    }
    p: Payment Service {
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

Likewise, as additional services are added or removed,
the `Payment Service` is forced to **continuously adapt** to the topology.
However, managing these changes should not fall on the `Payment Service`,
instead, the responsibility for handling dynamic topology changes should lie with added or removed components.

#### Semantic Coupling

**Semantic Coupling** occurs when services share the same data structures and semantics.

For example, if the `Payment Service` receives a response from the `Account Service`,
it must understand the structure of that response.
If the `Account Service` modifies the structure, it must notify the `Payment Service` to prevent errors.

```d2
grid-rows: 1
horizontal-gap: 100
direction: right
a: Payment Service {
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

The {{< term ioc >}} principle can be implemented using a [Message Queue (MQ)](Event-Streaming-Platform.md).
A {{< term mq >}} is essentially an informative **message container** with two primary associates:

- **Publishers** publish messages.
- **Consumers** consume and process messages.

```d2
direction: right
p: Publisher {
  class: server
}
m: Message Queue {
  class: mq
}
c: Consumer {
  class: server
}
p -> m: Publish
c <- m: Consume
```

Integrating an MQ into the first coupling example:

- The `Payment Service` can publish account subscription messages to the queue.
- The `Account Service` can later retrieve these messages to update the associated accounts.

```d2
direction: right
acc: Account Service {
    class: server
}
p: Payment Service {
    class: server
}
mq: Message Queue {
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

Beneficially, this design moves us away from:

- [Temporal coupling](#temporal-coupling): The `Account Service` exposes only a minimal set of interfaces
  and adapts to handle various messages instead.
  Furthermore, the scope of the `Payment Service` is reduced, granting it more flexibility;
  Like so, even if the `Account Service` fails to process messages,
  the `Payment Service` continues to develop and deliver without disruption.

```d2
direction: right
system: System {
    acc: Account Service {
        class: server
    }
    p: Payment Service {
        class: server
    }
    mq: Message Queue {
        class: mq
    }
    p -> mq: Subscription Message
    p -> mq: Cancellation Message
    p -> mq: Upgrade Message
    acc <- mq: Consume
}
```

- [Topology coupling](#topology-coupling): Additional services, such as `Notification Service` and `Fraud Detection Service`,
  can autonomously read messages without requiring any changes from the `Payment Service`.

```d2
direction: right
system: System {
    acc: Account Service {
        class: server
    }
    p: Payment Service {
        class: server
    }
    mq: Message Queue {
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

- Both services depend on the message queue. The dependency is minimized and barely problematic,
  as the {{< term mq >}} exposes only basic `Publish()` and `Consume()` interfaces that rarely change.
- Both the publisher and consumer adhere to the same message schema, which is [Semantic Coupling](#semantic-coupling).

We've just gone through some types of coupling and made messaging look mighty.
However, keep in mind that there are more types of couplings,
e.g., technology dependency, data dependency, flow dependency (like [SAGA](Compensating-Protocols.md#saga)), ...
which may weaken messaging.

Occasionally, messaging may result in an **unnecessary overhead** and outweigh the benefits of decoupling.

- The indirect communication model results in **slower performance**,
  making it unsuitable for low-latency workloads.
- Asynchronous communication can lead
  to [temporary inconsistencies](Distributed-Database.md#eventual-consistency-level), since changes aren’t immediately
  reflected across services.
- Debugging may become more challenging as failures are asynchronous and harder to trace.

In summary, while coupling in a microservice architecture can’t be eliminated,
they can be reduced and managed more effectively through messaging.
