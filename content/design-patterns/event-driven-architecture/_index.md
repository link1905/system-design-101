---
title: Event-Driven Architecture
weight: 10
prev: design-patterns
next: event-sourcing
---

## Service-Oriented Architecture (SOA)

**Service-Oriented Architecture (SOA)** is an architectural style centered on discrete services.
In this approach, a system is viewed as a collaboration of well-defined services.

Consider a user management system as an example.
Initially, an `Account Service` is responsible for creating new accounts.
A `User Service` then provides an interface for adding the corresponding user records.

```d2
direction: right
a: Account Service {
    "RegisterAccount()"
}
s: User Service {
    "AddUser(user)"
}
a -> s
```

If we later introduce a feature allowing users to delete their accounts,
the `User Service` expands to include an interface for deleting existing users.

```d2
direction: right
a: Account Service {
    "RegisterAccount()"
}
s: User Service {
    "AddUser(user)"
    "DeleteUser(userId)"
}
a -> s
```

The system evolves as its services expand and take on more responsibilities.
This demonstrates how services are central to building a system using **SOA**.

## Event-Driven Architecture (EDA)

An **event** represents a business fact that has occurred within the system,
such as `an order was created` or `an order was cancelled`.
In theory, the sequence of events reflects how a business evolves.

From this perspective, **Event-Driven Architecture (EDA)** advocates evolving a system around its events.

In our user management example,
a `UserCreated` event is triggered when a user registers an account.
A `UserCreated Handler` then handles this event and adds the corresponding user record.

```d2
direction: right
e: |||yaml
UserCreated:
  username: johndoe
  email: johndoe@mail.com
|||
h: UserCreated Handler {
    class: process
}
e -> h
h -> h: Add a user record
```

When the ability to delete users is introduced,
a `UserDeleted` event is introduced to support this requirement.
A new `UserDeleted Handler` is then developed to handle this event.

```d2
direction: right
c: "UserCreated" {
    e: |||yaml
    UserCreated:
        userId: user123
        username: johndoe
        email: johndoe@mail.com
    |||
}
d: "UserDeleted" {
    e: |||yaml
    UserDeleted:
        userId: user123
    |||
}
h1: UserCreated Handler {
    class: process
}
h2: UserDeleted Handler {
    class: process
}

c -> h1
d -> h2
```

Events are at the core of an **EDA** system.
The business is modeled as events and their transformations;
the system, comprising consumers and producers, is then developed to produce and handle these events.

## Event Collaboration

It is common for actions to involve multiple services.
For instance, when a new account is registered in the `Account Service`:

1. The `User Service` creates a new user record.
2. The `Notification Service` sends a welcome email.

### Orchestration

The first approach involves introducing an **orchestrator service**,
such as the `Account Service`, which performs these tasks sequentially.

```d2
shape: sequence_diagram
c: Client {
    class: client
}
a: Account Service {
    class: server
}
u: User Service {
    class: server
}
n: NotificationService {
    class: server
}
c -> a: "RegisterAccount()"
a -> u: "AddUser(user)"
a -> n: "SendEmail(email)"
```

Because everything is centralized and clearly defined in one place,
this approach is easier to understand and control,
particularly for managing complex interactions.
However, the orchestrator introduces dependencies between services and becomes a {{< term spof >}},
potentially reducing overall system availability and fault tolerance.

### Choreography

Conversely, **choreography** is an approach that relies purely on events:

1. The `Account Service` emits `AccountCreated` events.
2. The `User Service` consumes these events and, in turn, produces new `UserCreated` events.
3. The `Notification Service` then consumes `UserCreated` events to send welcome emails.

```d2
direction: right
a: Account Service {
    class: server
}
ac: AccountCreated {
    class: msg
}
u: User Service {
    class: server
}
uc: UserCreated {
    class: msg
}
n: NotificationService {
    class: server
}
a -> ac -> u -> uc -> n
```

In an **EDA** system, asynchronous communication is preferred.
This enables different parts of the system to react to events independently,
promoting loose coupling and fault tolerance.

However, this approach can become difficult to manage in complex workflows.
This complexity is the most significant challenge of **EDA**.
A single business operation might involve a long chain of events,
produced and handled in many different parts of the system.
This makes it difficult for developers to fully understand the business process and develop the system effectively.
The complexity is particularly pronounced in large systems,
which may handle a vast number of events (potentially hundreds or even thousands).

{{< callout type="info" >}}
This is a brief introduction to [Saga]({{< ref "distributed-transaction#saga" >}}).
We will discuss it in more detail in the corresponding section.
{{< /callout >}}
