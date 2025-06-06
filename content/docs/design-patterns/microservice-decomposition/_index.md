---
title: Microservice Decomposition
weight: 10
---

We've basically mentioned the [Microservice]({{< ref "microservice" >}}) architecture in the first topic.
In this one,
we'll dive deep on more theories and see how to design a {{< term ms >}} system.

## Microservice Perspective

### Monolith Architecture

Traditionally,
IT systems are usually built around the {{< term mono >}} architecture.
In brief,
we try to stack everything inside a single codebase.
Developers, despite working at different teams,
must share the same assets including sourcecode, database, technologies, etc.

```d2
ta: Team A {
  class: group
}
tb: Team B {
  class: group
}
tc: Team C {
  class: group
}
s: Shared {
  c: Codebase {
    class: code
  }
  d: Database {
    class: db
  }
}
ta -> s
tb -> s
tc -> s
```

Nowadays, business is changed and transformed at a very rapid rate.
An IT system is born to resolve a business problem,
when the target problem evolves, the IT system needs to capture the changes promptly.
Unfortunately,
large systems often fail to meet demand due to the heavy collaboration required between different teams or departments.
Any change in one part can potentially introduce errors in others,
clumsy coordination is necessary to ensure the system's integrity.

For instance, when a massive codebase is shared across teams,
modifying programmatic classes may require consulting all relevant teams to confirm their acceptance of the changes.
This challenge is even greater for teams located in different countries or regions.


### Microservice Architecture

To address this, we need strategies to **minimize dependencies** within the system.
{{< term ms >}} comes to the rescue,
we dive the big codebase into smaller fragments (for granular business problems)
and assign each for a certain team.

```d2
ta: Team A {
  s: Codebase A {
    class: code
  }
  d: Database A {
    class: db
  }
}
tb: Team B {
  s: Codebase B {
    class: code
  }
  d: Database B {
    class: db
  }
}
tc: Team C {
  s: Codebase C {
    class: code
  }
  d: Database C {
    class: db
  }
}
```

An excellent microservice system would generate significant momentum
in evolving different services **independently** and **rapidly**:

1. **Cohesion**: A microservice resolves around a specific business problem,
that means changes are usually coupled within the service and do not affect the others.
2. **Parallelism**: Microservice helps enable smaller, focused teams to work on individual services in parallel.
3. **Independent evolution**: Typically, microservices share nothing in between (codebase, databases, technology stack...).
They can evolve independently of other services, according to their specific needs.

In conclusion,
{{< term ms >}} is an architecture primarily designed to address human-oriented and business challenges rather
than technical issues (system performance, system availability, etc).
A microservice system does not help reduce development effort is good for nothing.
Always keep this perspective in mind!


### Microservice Challenges

The separation of Microservice comes with many **inevitable** challenges.

#### Poor Performance

Actually, services are independent processes or machines,
native function calls are no longer possible.
Despite protocol, the communication between services must occur though **network**,
leading to high latency and poor performance

#### Sourcecode Duplication

A microservice typically maintains an isolated codebase to minimize dependencies between teams.
However, this isolation often leads to reinventing the wheel,
as different teams may unknowingly duplicate implementations.

To keep it productive, services can share specific modules under certain conditions:

- Minimal Coupling: Modules that are rarely updated can be shared between services.
- Common Tasks: Shared modules should address universally agreed-upon tasks, such as monitoring or authentication.
- Exchanged Schemas: For example, shared gRPC definitions for service communication.

However, sharing usually results in disrupting the agility and going against the benefits of {{< term ms >}}.
Thus, most of the time, we should minimize it as much as possible.

#### Tight Coupling

Once a service needs to cope with other services to complete its business,
we call it a **Coupling**.

```d2
s1: Service A {
  class: server
}
s2: Service B {
  class: server
}
s1 -> s2: Coupling
```

Couplings probably create shared sections in between:
data, API contracts, shared libraries, etc.
They will restrict services from autonomous development as

- Teams must tightly coordinate when working on the shared parts.
- Teams hesitate to modify shared parts due to the risk of unintended consequences.

A misdesigned system can result in tightly coupled services,
fading out the {{< ms >}} advantages.

{{< callout type="info" >}}
We've discussed in depth in [the first topic]({{< ref "microservice#service-decoupling" >}}).
You may have a review before continuing as I don't want to repeat it here.
{{< /callout >}}

## Microservice Recommendations

In common,
we have some best practices for designing an effective {{< term ms >}} system.
Actually, producing a perfect solution solving all of them is unbelievable,
but we try to reduce the margins as much as possible.

### Single Responsibility Principle (SRP)

**Single responsibility principle (SRP)** states that
[gathering together the things that change for the same reasons](https://en.wikipedia.org/wiki/Single-responsibility_principle).
Briefly, a microservice should be in charge of a single and well-defined scope.

For example,
a `Customer Management Service` attempting to handle everything related to customers:
billing, customer analytics, customer support...

- This wide service spanning multiple problems needs business knowledge in different domains,
  making it harder for developers to approach and work with.

- More importantly,
  the wide service requries many efforts to maintain,
  invalidating {{< term ms >}} benefits such as parallelism and autonomous evolution.

Instead, small services are suggested to speed up development and minimise coordination.

### Cohesion

**Cohesion** is a property stating that
all changes tightly **coupled** with each other should be bounded within a single service.

#### Loose Coupling

**Loose Coupling** expects that designing services in a system
so that they have **minimal or none** dependencies on each other.
**Cohesion** is the most powerful weapon to achieve it.
A cohesive service can handle its task independently and leave minimal dependency in the system.

#### Lock-step Deployment

When microservices are not cohesive enough,
a business change will impact on different services.

A common causal is **Lock-step Deployment**,
when different services must be deployed simultaneously to ensure the system effectiveness.
The delay between services significantly slows down the evolution of the entire system

```d2
s: "" {
    class: none
    grid-rows: 3
    sa: Service A {
        class: server
    }
    sb: Service B {
        class: server
    }
    sc: Service C {
        class: server
    }
}
d: "Release"
s.sa -> d
s.sb -> d
s.sc -> d
```

Teams should be capable of autonomously managing their development and delivery strategies.
Ideally, they would better deploy confidently without requiring any coordination.
Conversations like, `Hi, we're deploying a new API version. Please check the documentation and update your service`,
should be discarded!

```d2
s: "" {
    class: none
    grid-rows: 3
    sa: Service A {
        class: server
    }
    sb: Service B {
        class: server
    }
    sc: Service C {
        class: server
    }
}
d: "" {
    a: "A Release"
    b: "B Release"
    c: "C Release"
}
s.sa -> d.a
s.sb -> d.b
s.sc -> d.c
```

We easily observe see that:

- **SRP** requires establishing small services to enable small teams and speed up development process.
- Meanwhile, **Cohesion** expects services to be wide to diminish coordinations, allowing teams to evolve autonomously.

Thus, in the design process,
we want to build **right-size** services ensuring both facets.

## Decomposition Strategies

Now, we move to the main part - **Microservice Decomposition**.

### Decomposition By Business Capability

#### Business Capability

A business capability describes **what** an organization does to deliver values to its customers.
For example, an e-commerce system might include the following capabilities:

- `Product Management`: Managing product catalogs.
- `Order Management`: Handling customer orders from creation to delivery.
- `Customer Support`: Resolving customer queries.

Business capabilities are generated by **business experts** at a high level,
who encounter problems, relying on the system to solve them instead of attempting themselves.
Whether implemented on paper or through a high-end system,
as a microservice or a monolith, business capabilities remain agnostic of these implementation factors.
So, building a service for each business capability is a **business-oriented** approach.

```d2
s: System {
   p: Product Service {
      class: server
   }
   c: Customer Service {
      class: server
   }
   o: Order Service {
      class: server
   }
}
```

Apart from the simplicity, this approach is safe for the **Cohesion** requirement,
as it's naturally upheld by aligning with well-defined business scopes.

However, this approach is not well-suited for large systems.
The responsibility of maintaining a system lies on technicians,
and a pure business-driven viewpoint can easily result in an inefficient IT system.

For example, `Order Management` is a broad business capability with many tasks, e.g. fulfilment, payment, stock, etc.
It would be detrimental if it's directly mapped to a single microservice,
potentially violating the **SRP**.

### Domain-Driven Design

IT experts, who develop and maintain system,
needs considering in the process of designing a productive system.
Therefore, a smart approach is **bridging** between technical and business angles.
In the next subtopic,
we will explore a popular software design pattern towards the idea called **Domain-Driven Design**.
