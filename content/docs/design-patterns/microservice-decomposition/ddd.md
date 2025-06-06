---
title: Domain-Driven Design
weight: 10
---

## Domain And Domain Model

The **domain** of an organisation is the field it runs on, e.g. insurance, news, banking...
The organisation needs to solve problems originating from the domain,
for example,
a banking domain may have several problems:
`Loan Management`, `Saving Management`, `Customer Support`...

A **domain model** is a collection of all necessary insights and tools needed
to solve its domain's problems.
A domain model properly mentions many aspects and assets:
documents, data schemas, sourcecode, ...

```d2
d: Domain {
  grid-columns: 1
  "Loan Management"
  "Saving Management"
  "Customer Support"
}
dm: Domain Model {
  grid-columns: 1
  "Data schemas"
  "Sourcecode"
  "Documents"
  "Analyses"
}
dm -> d: Solve
```

## Shared Domain Model

**Domain-Driven Design (DDD)** suggests strategies to
provide a domain model **exactly reflecting** its domain.

### Business Reflection

**DDD** requires the domain model to be created based on a **tight collaboration**
between technical and domain experts, to close the gap between them as much as possible.

In a normal scenario,
developers have no clue about the business,
they develop the system based on knowledge **guided** by domain experts.
In other words,
the technical experts follow and reflect rules from the domain experts to build the domain model.

For example, a business requirement states that `users can withdraw their balance`,
we expect a class `User` with a method `Withdraw` implemented in the system.

However, information is transmitted back-and-forth and there is no strict rules in the system,
it's potential to have problems in **Language Translation**.

For example, it's a mismatch if developers implement a class `Client` with a method `DecreaseBalance`,
then terms need repetitively translating or implicitly understanding between organizational sections.

### Enterprise Model

Intuitively,
we can create a shared source of knowledge containing everything, called **Enterprise Model**.
All teams and employees share and use the rich model to avoid translations.

```
d: Domain experts {
  class: domain-expert
}
t: Technical experts {
  class: dev
}
s: Enterprise Model {
  t: |||yaml
  User: refer to system end-users
  Withdraw: users can withdraw their balance
  |||
}
d -> s
s -> s
```

Based on this model, people can share the knowledge and avoid inconsistencies.
However, in a huge environment,
**Enterprise Model** is easy to generate **Language conflicts**:
Different teams may mention the same term differently.

For example, in a banking system

- The `Saving` department mentions `User` as whom pour money into the system.
- The `Loan` department uses the term to refer whom borrow money from the system.

Basically, we can effectively complicate the model by adding more meanings for a term.
To retrieve the correct definition, we need to attach the term with its context

```yaml
User:
  Saving context: whom pour money into the system to receive interests
  Loan context: whom borrow money from the system
```

However, this makes the model become extremely challenging and confusing to use.
When referring to any term correctly,
people need to attach the context;
This explication will happen repetitively from place to place,
e.g., documents, code, meetings, etc.

Therefore, it is recommended to break the unified model into smaller manageable and isolated fragments,
helping reduce conflicts within the organization.
In the next subsections, we will see how a large model is separated into pieces.

## Ubiquitous Language

Domain experts have limited understanding of the technical language, and vice versa.
That potentially turns out business and technical people have their own jargon.

As discussed,
we want to set up a **shared dictionary** in the domain model
to minimize the cost of translations and the risk of misunderstandings.

```yaml
Account:
  Definition: A financial record that holds the balance and transaction history.
```

People (domain experts, developers, operators...) are encouraged to apply the dictionary
**everywhere**: documents, codebase, data schema, meetings, ... even direct conversations,
using inconsistent or conflicting words is prohibited.
Therefore, this shared dictionary is called **Ubiquitous Language (UL)**.

### Business Invariant

Alongside business concepts, **UL** should make clear of business invariants (rules).

```yaml
Account:
  Definition: A financial record that holds the balance and transaction history.
  Invariants:
  - An account should not send more money than its limit
```

### UL Evolution

An **UL** cannot be established once and in place permanently,
it has to **continuously** evolve.
While domain experts bring in business concepts,
technical experts help shape how these concepts are implemented effectively in the system.
For example, from the previous invariant, technical experts may ask

- *Is the limit daily, weekly or monthly?*
- *Should the system block transactions exceeding the limit immediately?*

Then, the **UL** is refined with more types of limits

```yaml
Account:
  Definition: A financial record that holds the balance and transaction history.
  Invariants:
  - An account must obey its daily limits

Daily Transfer Limit: The maximum amount a customer can transfer in 24 hours.
Soft Limit: Allows transactions beyond the `Daily Transfer Limit` but requires additional verification (e.g., OTP).
Hard Limit: Transactions exceeding this amount are automatically rejected.
```

As more conversations and collaborations happen, new vocabularies will be added to the **UL** gradually.
IT experts are in charge of capturing changes in the **UL** and
adapting the domain model accordingly (code refactoring, database restructuring, ...)

## Bounded Context

### Context

Context is the environmental factors driving the meaning of a piece of information.
For example, when using the term `User`,
we must explicitly specify the context `Saving` or `Loan`.

```yaml
User:
  Saving context: whom pour money into the system to receive interests
  Loan context: whom borrow money from the system
```

As more and more contexts appear,
a single unified **UL** across the organisation is challenging to deal with.
A term is potentially applied differently in different contexts,
we need to attach the context when referring to it.
More importantly, being aware of business context is error-prone,
developers must get well with the entire domain to ensure working with the correct context.

### Bounded Context {id = "bounded-context-details"}
To minimize conflicts, we'd better split a system into smaller contexts.
These contexts cover different subsets of problems
and don't overlap with each other, usually called `Bounded Context`.
![](./bounded-context.png)

A team is designated to a well-defined context to help the members breathe better.
They only need to learn and work with their context's business,
the outside world is carefree.
As a consequence, teams in different contexts can evolve autonomously,
e.g., a developer works at the `Saving` context,
then `Loan` or `Customer Support` contexts are none of his business

A bounded context has its own **domain model** to resolve the business problems.
In other words, the model only has the proper meaning if it lives inside the context.
Shifting a model to another context needs **additional translations**, otherwise the model will be misinterpreted.
For example, `Saving` and `Loan` models have the `Account` entity,
different contexts lead to different explanations
![](bounded-context-model.png)

### Bounded Context Realization
Realizing bounded contexts within an organization is paramount.
There is **no golden rule** to finish this job,
it heavily relies on experience and an effective collaboration with business experts

#### Business Capability
First and foremost, we need to get a clear and good understanding of the business.

We should actively read existing documents (user stories, design documents...) and observe business operations.
The key action is looking at the organization structure, helping us conveniently detect business capabilities.
```md
Board Of Director
│
├── Core Banking Operations
│   ├── Account Management
│   └── Savings Management
│
├── Loan Management
|   ├── Loan Management
|   └── Credit Scoring
│
│── Risk & Compliance
│   ├── Fraud Prevention
│   └── Regulatory Compliance
```

The cohesion of departments allows us to initially demarcate **broad** bounded contexts.
![](bounded-context-capability.png)

Then, we will identify and schedule interviews with the associated experts
to dive into these capabilities.

#### Experts Interviewing
While working with business experts, we need to capture the overall workflows with key activities.
Besides, we must pay attention to and take note of common business terms to enrich the `UL`.

The key activities and `UL` clues are helpful to detect granular bounded contexts
We can observe **linguistic conflicts** are arisen between workflows,
they are potentially signs of living in different bounded contexts

Let's take the `Loan Management` context above.
We have two processes of loan approval from
```d2
grid-rows: 2
r: Retail Loan {
  grid-rows: 1
  "Loan Application" -> "Credit Evaluation" -> "Sign Agreement"
}
c: Cooperate Loan {
  grid-rows: 1
  "Loan Application" -> "Request Financial Documents" -> "Credit Evaluation" -> "Sign Agreement"
}
```

The processes may generate semantic conflicts,
causing a loan approval to take care of contextual verification
- `Customer` can be an individual or a business
- `Credit Evaluation`
  - An individual is checked based on the credit score.
  - The system evaluates the creditworthiness of a business by checking its credit history, cash flow, and financial stability.

To improve autonomy and maintainability, we prefer small bounded contexts.
Therefore, we might separate the context into smaller sections
![](bounded-context-smaller-contexts.png)

#### Legacy Codebase
Inheriting a well-designed legacy system is also a huge advantage,
we may leverage the former employees' perspective.
The codebase may show small and manageable modules (or namespaces),
each of them is potentially a bounded context.

### Context Map
An organisation is barely impossible to operate with **completely** isolated bounded contexts.
Although it is ideal to have truly independent contexts, the coordination between them is inevitable.
The connections between contexts must be controlled to avoid unmanageable and haphazard models.

`Context Map` is a tool demonstrating relationships between bounded contexts.
In this map, we have some common types of relationships


#### Separate Ways
The first relationship is no relationship![](emoji-scared-cat.svg).
```d2
%d2-import%
a: Context A
b: Context B
```

This is the ideal setup,
allowing different bounded contexts to evolve autonomously.

#### Asymmetric Relationships
When a context (`upstream`) shared its model to another context (`downstream`),
it's called an asymmetric relationship.
```d2
%d2-import%
a: Upstream Context
b: Downstream Context
b <- a: Share
```

Please note that
the upstream context has no idea of the downstream one.
In other words, the upstream context can evolve freely,
but the downstream needs to capture changes from the source.

##### Conformist Pattern
`Conformist (CF)` pattern suggests that the downstream context
must conform to what the upstream shares, that means it adopts and reuses the shared model internally.
```d2
%d2-import%
a: Upstream Context {
  a: Account {
    shape: sql_table
    AccountId: string
    SSN: string
    Name: string
  }
}
b: Downstream Context {
  a: Account {
    shape: sql_table
    AccountId: string
    SSN: string
    Name: string
  }
}
b <- a: CF
b.a <-> a.a: Same model
```

This pattern allows quickly adapting from the upstream context.
Apart from simplicity and rapidity, it's unsafe as
changes from the upstream can propagate significant refactors in the downside.
In fact, `CF` is frequently used for small scopes or experimental purposes.

##### Anti-Corruption Layer Pattern
`Anti-corruption Layer (ACL)` is a pattern isolating models from outside contexts to minimize corruption.
Briefly, the downstream context needs to build a layer intercepting and transforming
outside information to its internal model.
```d2
%d2-import%
a: Upstream Context {
  a: Account {
    shape: sql_table
    AccountId: string
    Name: string
  }
}
b: Downstream Context {
  acl: "Anti-corruption Layer (ACL)"
  a: Account {
    shape: sql_table
    Id: string
    FirstName: string
    LastName: string
  }
  acl -> a: Transform
}
b.acl <- a
```

This brings in agility and protects the downstream context from propagated changes or errors.
`Anti-corruption Layer (ACL)` is always recommended if we have enough development effort.

#### Symmetric Relationships
When changes need mutually coordinating between relevant contexts,
it's called a symmetric relationship.

##### Partnership Pattern
`Partnership` is a pure symmetric pattern,
requiring relevant bounded contexts to communicate mutually.
```d2
%d2-import%
a: Context A
b: Context B
a <-> b
```

Despite segregation, contexts are still tightly coupled with each other.
Moreover, a context's members also need to learn about the other context (business rules, UL, ...),
although they live outside it.

##### Shared Kernel Pattern
`Shared Kernel Pattern` is frequently used to mitigate the `Partnership` pattern.
This patterns requires demarcating around shared models between the bounded contexts,
then changes outside the shared part can evolve independently.
Unavoidably, the teams must coordinate when making changes in this overlap.

For example, `Saving` and `Loan` share the same `account` entity
![](./bounded-context-shared-kernel.png)

Yet this pattern is not a panacea,
a large shared part still results in high coupling.

#### Mapping Benefits
Modeling context mapping helps understand how different teams and systems interact.
```d2
%d2-import%
a: Account Management
s: Savings Management
l: Loan Management
c: Credit Scoring
l <-> c: Shared Kernel
s <- a: CF
l <- a: ACL
c <- a: ACL
```


Relationships partially show the **level of coupling** between contexts.
Their spectrum of co-operation can be treated below.
```d2
%d2-import%
grid-columns: 1
l1: "" {
  class: none
  "Separate Ways" {
    width: 200
  }
}
l2: "" {
  class: none
  "Anti-Corruption Layer" {
    width: 400
  }
}
l3: "" {
  class: none
  "Conformist" {
    width: 600
  }
}
l4: "" {
  class: none
  "Shared Kernel" {
    width: 800
  }
}
l5: "" {
  class: none
  "Partnership" {
    width: 1000
  }
}
```

Thickly connected bounded contexts may hint a bad design.
For example, we've decided to separate a bounded context into 3 smaller sections.
However, at the end of the day, they're still highly coupled with each other.
We only gain no benefit from agility, but adding more management overhead.
```d2
%d2-import%
b: Bounded Context
c: "" {
  b1: Bounded Context A
  b2: Bounded Context B
  b3: Bounded Context C
  b1 <-> b2
  b2 <-> b3
  b1 <-> b3
}
b -> c
```

Moreover, in the implementation phase,
teams in different bounded contexts are expected to respect the context map.
Ignoring the context map can lead to tighter coupling, unclear ownership, and harder-to-maintain systems.

## Tactical Design
`Tactical Design` refers to the process of conceptually
designing the domain model within a bounded context.
In a context, there are many types of domain objects,
this process helps recognize and tell relationships between them.

### Transaction Script Pattern
`Transaction Script Pattern` is a common mistake of implementing `DDD`.
This pattern will establish `anaemic` models **only** containing state, getters and setters,
while business logic is placed on outer service classes

For example, when a user withdraws an amount,
`AccountService` checks the balance and creates the associated transaction
```d2
direction: right
o: AccountService {
    shape: class
    Withdraw(accountId, amount): ""
}
a: "BankAccount" {
    shape: class
    id: "string"
    balance: "number"
}
s: "Transaction" {
    shape: class
    id: "string"
    amount: "number"
    type: "(Deposit, Withdrawal)"
}
o -> a: 1. Checks balance
o -> s: 2. Creates a new transaction
o -> a: 3. Decreases balance and adds the transaction
```

This problem regularly comes from the expectation of desegregating state from logic.
Remember the most important principle of `DDD`, the technical side must reflect the business.
This pattern makes the system drift off the business, making it challenging to maintain.

Instead, we should place business [invariants](#business-invariant) inside domain objects.
For example, we want `BankAccount` object to control its own withdrawal processes,
not passing to an outside service.
```d2
direction: right
a: "BankAccount" {
    shape: class
    id: "string"
    balance: "number"
    Withdraw(amount): "" {
      style.bold: true
    }
}
s: "Transaction" {
    shape: class
    id: "string"
    amount: "number"
    type: "(Deposit or Withdrawal)"
}
a -> s: Creates
```

### Entity
An entity is a uniquely identifiable object.
That means two entities are equal if they possess the same identifier.

For example, in a banking system, we can specify
- An account with `AccountId`
- A transaction with `TransactionId`

### Value Object
Unlike `Entity`, `Value Objects` do not have a unique identifier,
comparing between them needs to consider their **attributes**.
They primarily serve for utility purposes, e.g., wrapping and reusing business logic.

For example,
letting the `Transaction` entity hold a single number is ambiguous,
because of currency ignorance.
Then, we create a `Value Object` specifying an amount of money
```d2
direction: right
s: "Transaction" {
    shape: class
    id: "string"
    type: "(Deposit, Withdrawal)"
}
c: "Currency" {
  shape: class
  amount: "number"
  unit: "(Dong, USD, Yen...)"
  Display();
}
s -> c: With
```

To compare between two `Currencies`,
we need to check their `unit` then `amount`,
all attributes in other words.

### Aggregate
An aggregate is the cluster of one or more entities to
ensure the [atomic consistency](Concurrency-Control.md#atomicity) of them.

Take the banking example.
If a `Transaction` is directly canceled without passing the associated `BankAccount`,
some operations will be missed, e.g., modifying the balance in the account.
```d2
direction: right
Saving Aggregate {
  a: "BankAccount" {
      shape: class
      id: "string"
      balance: "number"
  }
  s: "Transaction" {
      shape: class
      id: "string"
      amount: "number"
      type: "(Deposit, Withdrawal)"
      Cancel(): "" {
        style.bold: true
      }
  }
  a -> s
}
```

Don't try to refer the `BankAccount` from the `Transaction`!
`Cyclic reference` between entities is a bad practice and makes the codebase as hell,
we should completely avoid it.

#### Atomic Operation
We want the operation to be automatically executed.
In other words, all changes will be wrapped inside a single method.

To do that, we build an aggregate and designate a facade entity called `Aggregate Root`.
We only deal with the root to make sure operations to take effect on relevant entities.

For example,
when the `Cancel()` operation is called, the flow must originate from the `BankAccount` class
```d2
direction: right
Account Aggregate {
  a: "BankAccount" {
      shape: class
      id: "string"
      balance: "number"
      Cancel(transactionId): "" {
        style.bold: true
      }
  }
  s: "Transaction" {
      shape: class
      id: "string"
      amount: "number"
      type: "(Deposit, Withdrawal)"
      Cancel(): "" {
        style.bold: true
      }
  }
  a -> s
}
```

Guaranteeing consistency between the entities is the most important principle of `Aggregate`.
Changes are encapsulated within the root as **atomic** operations,
helping to escape the responsibility of maintaining consistency in different entities.

#### Referenced from root
A child entity can be directly accessed
if it's **referenced** from the root and the function is **self-contained**.
For example, the `CalculateAmountWithAnotherUnit()` function can be called directly
after retrieving the transaction from the account
```d2
Account Aggregate {
    a: "BankAccount" {
        shape: class
        id: "string"
        getTransaction(transactionId): Transaction {
            style.bold: true
        }
    }
    s: "Transaction" {
      shape: class
      id: "string"
      amount: "number"
      type: "(Deposit, Withdrawal)"
      CalculateAmountWithAnotherUnit(unit): "number" {
        style.bold: true
      }
    }
    a -> s: Ref
}
```

#### Referenced by identifier only
When an aggregate references an entity living in another aggregate,
it's supposed to maintain the identifier only.

For example, we have two aggregates,
a `BankAccount` can have some `SavingAccount`s.
From the `BankAccount`'s view, it only possesses saving account ids
instead of actual instances
```d2
Account Aggregate {
    a: "BankAccount" {
        shape: class
        id: "string"
        balance: "number"
        savingAccountIds: "[string]"
    }
}
Saving Aggregate {
    a: "SavingAccount" {
        shape: class
        id: "string"
        amount: "number"
    }
}
```

That's an entity instance belongs to an aggregate, and it's valid only if the aggregate root creates it.
That means the root can spawn another entity and make the old instance **detached** and invalid.
So, forcing entities to move around their aggregate is a safe approach.

#### Mediator Pattern
However, the identifier field is no use to manipulate data.
To perform changes on multiple aggregates, we will build a center mediator.
```d2
c: Mediator
a: Account Aggregate {
    a: "BankAccount" {
        shape: class
        id: "string"
        balance: "number"
        savingAccountIds: "[string]"
    }
}
s: Saving Aggregate {
    a: "SavingAccount" {
        shape: class
        id: "string"
        amount: "number"
    }
}
c -> s: 1. Close account
c -> a: 2. Update balance
```

As changes don't need to be atomic, to stimulate agility,
[](Distributed-Database.md#eventual-consistency-level) is recommended to implement the mediator.
For example, when a `SavingAccount` is closed,
an event `ClosedSavingAccount` is fired,
the associated `BankAccount` will take the deposit accordingly.
```d2
m: Mediator
a: Account Aggregate {
    a: "BankAccount" {
        shape: class
        id: "string"
        balance: "number"
        savingAccountIds: "[string]"
    }
}
s: Saving Aggregate {
    a: "SavingAccount" {
        shape: class
        id: "string"
        amount: "number"
    }
}
s -> m: ClosedSavingAccount
s <- m: ClosedSavingAccount
```

But that doesn't offer atomicity, why don't we cluster them as a single aggregate?
Technically, we prefer building aggregates as small as possible.
We've not mentioned the storage layer yet,
however, you may imagine that entities need persistently **stored**.
To perform manipulation in a program, we need to somehow fetch data from data stores,
map to entities and maintain them in memory.
A wide aggregate relates to a higher number of records,
consequently, the fetching process happens repetitively and consumes resources dramatically.

So, in any case, if eventual consistency is **acceptable**,
we're encouraged to build granular aggregates rather than the broad ones.

### Service
Services are referred to stateless classes only containing logic,
that mean service behaviours are based on entities' state.
There are two primary types of services

> The term `service` comes from the perspective of software development, not related to `Microservice`

#### Domain Service
`Domain services` implement **business logic**, but not naturally fitted in any entities or aggregates

Let's say we want to transfer money between two accounts.
Intuitively, it's unnatural to put this logic inside the `BankAccount` class.
Even in business, users can’t solely transfer money directly like that,
they need a coordinator standing in between (e.g., ATM).
```d2
direction: right
s: TransferService {
    shape: class
    transfer(fromAccountId, toAccountId, amount): "" {
        style.bold: true
    }
}
a1: "BankAccount A" {
    shape: class
    id: "string"
    balance: "number"
    withdraw(amount): "" {
      style.bold: true
    }
}
a2: "BankAccount B" {
    shape: class
    id: "string"
    balance: "number"
    deposit(amount): "" {
      style.bold: true
    }
}
s -> a1: 1. Withdraw
s -> a2: 2. Deposit
```

#### Application Service
`Application service` do **not** deal with domain logic but depending on domain objects.
This type of service usually offers utilities at the application layer
- Authentication and authorization
- Data aggregation from multiple aggregates
- [](Event-Sourcing.md#cqrs)

##### Application Layer
What does the application layer mean?
`Entity`, `ValueObject` and `Aggregate` are domain objects, their collaboration reflects how the domain works.
They are technology-agnostic, and can be implemented without the help of any framework or libraries.
On the other hand, the application layer is built on top of the domain layer,
not participating in the domain,
but offering utility elements to bring in the domain to the outside world.

For example, a client application wants to view transactions as pages.
Due to irrelevance to the business, we create an `Application service` to tackle it
```d2
direction: right
s: TransactionPagingService {
    shape: class
    getTransactions(userId, fromDate, toDate): "Transaction[]" {
        style.bold: true
    }
}
```

## Microservice Decomposition
We’ve researched a part of `DDD`, and now it’s time to discuss microservice decomposition

### Bounded Context As Microservice
First and foremost, a microservice mustn’t span multiple bounded contexts.
Bounded contexts are separated to reduce complexity for developers,
so mixing them again defeats the initial purpose

The first candidate for a microservice is bounded contexts.
They reflect business boundaries and help effectively organize and align teams.

Sometimes, a bounded context is often too large to map to a single microservice.
To enable fast and parallel development, we may attempt to find smaller targets.

### Aggregate As Microservice
`Well-defined Aggregates` are strong candidates for microservices.

Aggregates are highly cohesive and independent with atomic operations.
Different aggregates aren't supposed to be created together and referenced to each other,
that helps them decoupled in terms of storage and programming.

Moreover,
cooperative between aggregates (or microservices) can be performed in the eventual consistency manner,
ensuring high availability.

#### Domain Service As Microservice
`Domain services`, which often interact with some aggregates, are also a viable candidate for microservices.
In fact, domain service is usually built as a [distributed transaction coordinator](Distributed-Transaction.md).


### Microservice Verification
Rome wasn’t built in a day!
Before moving further, we must carefully verify the design

We’ve discussed some requirements of a good microservice system,
and you may want to [review them again](Microservice-Decomposition.md#microservice-requirements):
- `Cohesion`: All tightly coupled changes should be confined within a single microservice
- `Single Responsibility`: A microservice should have a single, well-defined responsibility
- `Loose Coupling`: Services should have minimal dependencies on each other, allowing them to evolve independently

In addition to these, non-functional requirements play a critical role in microservice decomposition:
- **Team size**: The size of a team depends on the human capital.
Services that are too large or too small can lead to an imbalance in effort distribution, resulting in some services lacking adequate maintainers.
- **Team competency**: Managing and developing a complex system requires much experience (e.g., observability, distributed transaction...).
It’s crucial to ensure that developers are capable of meeting these challenges.

However, things often go awry.
Microservice decomposition is a **continuous process**.
People often overcompensate by initially creating too many small services,
which leads to increased complexity and problems down the line.
In reality, splitting a service into smaller fragments is far easier
than merging them back into a unified whole.

## Clean Architecture

In terms of `DDD Bounded Context`, there are two types of events.
