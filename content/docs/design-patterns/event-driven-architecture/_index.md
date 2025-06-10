---
title: Event-Driven Architecture
weight: 10
---

## Service-oriented Architecture

**Service-oriented architecture (SOA)** is an architectural style that focuses on discrete services.
In other words, we treat the system as the collaboration of well-defined services.

For example, we develop a user management system.
First, the `Account Service` is used to create new accounts.
Then, the `User Service` exposes an interface for adding corresponding records.

```d2
a: Account Service {
    "RegisterAccount()"
}
s: User Service {
    "AddUser(user)"
}
a -> s
```

Later we allow users to delete their account,
the service so grows with another interface for deleting existing users.

```d2
a: Account Service {
    "RegisterAccount()"
}
s: User Service {
    "AddUser(user)"
    "DeleteUser(userId)"
}
a -> s
```

## Event-Driven Architecture (EDA)

An **event** is a business fact happening within the system,
e.g. `an order was created`, `an order was cancelled`, etc.

Theoretically speaking, the evolution of events tells the development of a business.
In perspective, **Event-Driven Architecture (EDA)** recommends evolving system around events.

In the example, a `UserCreated` event is fired when the user registers an account.
A `UserCreated Handler` simply captures and adds a user record.

```d2
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

Once deleting users is allowed,
an event `UserDeleted` is built for this requirement.
A new `UserDeleted Handler` is created to adapt to this event.

```d2
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

What are their producers and consumers? We don't care!
Events stand in the heart of an **EDA** system.
We treat the business as events and their transformations,
the system is then built to handle and adapt to events accordingly.

## Event Collaboration

It's usual to have actions relating to multiple services.
For example,
when a new account is registered in `Account Service`:

1. The `User Service` creates a new user record.
2. The `Notification Service` sends a welcome email.

### Orchestration

The first approach is introducing an **orchestrator service**, such as `Account Service`,
performing these tasks in order.

```d2
direction: right
a: Account Service {
    "RegisterNewAccount()"
}
u: User Service {
    "AddUser(user)"
}
n: NotificationService {
    "SendEmail(email)"
}
a -> u: 1. Create new user
a -> n: 2. Send welcome mail
```

Since everything is centralized and well-defined in one place,
this approach is easier to understand and control,
especially for managing complex interactions.
However, the orchestrator introduces a clear interdependency
and becomes a {{< term spof >}},
which can reduce overall system availability and fault-tolerance.

### Choreography

While **Choreography** is an approach purely depending on events:

1. `Account Service` fires `AccountCreated` events.
2. `User Service` captures them and produces new `UserCreated` events correspondingly.

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

Despite loose coupling and fault-tolerance, this approach becomes problematic when dealing with complex workflows.
This complexity is also the most critical issue of **EDA**.
A business operation can be a chain of many events, generated and causing effect in many places.
This implication makes it challenging for developers to capture the business and develop the system.
The complication is exceptional in big systems,
as they can deal with a huge number of events, up to hundreds or thousands of them.

{{< callout type="info" >}}
This is just an introduction of [Saga](Distributed-Transaction.md#saga).
We will discuss it more in the appropriate section.
{{< /callout >}}
