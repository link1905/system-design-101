---
title: Event Sourcing Pattern
weight: 10
prev: event-driven-architecture
next: distributed-transaction
---

{{< callout type="info" >}}
You may review the concept of an [Event Streaming Platform]({{< ref "event-streaming-platform" >}}) if necessary.
{{< /callout >}}

In this topic, we will explore a common pattern utilized in **EDA** systems:
**Event Sourcing**. This pattern facilitates data sharing between teams by relying on a single source of truth.

## Data Coupling

**Data Coupling** is one of the most significant challenges in **EDA**.
Events rarely contain all the information consumers need to process them,
forcing consumers to seek additional data from other data sources.

For instance, after receiving an `AccountBalanceChanged` event,
the `Notification Service` must fetch user information from the `User Service` to send an email.

```d2
direction: right
u: User Service {
    class: server
}
n: Notification Service {
    class: server
}
m: "AccountBalanceChanged" {
    class: msg
}
n <- m: 1. Consume
n <- u: 2. Fetch the user information
```

Certain datasets are central to the business (e.g., common user information) and are widely accessed by numerous services.
While we can decouple infrastructure, codebase, and workforce, data is inherently generated in specific locations.
Although the goal is to make the system as loosely coupled as possible, some degree of data coupling is inevitable.

### Service Interface

The most common method for sharing data is by directly using service interfaces.
When a piece of information is needed, a call is made to the service that owns the data.
This is what occurred in the previous example: a call is made to the `UserService` for every `AccountBalanceChanged` event.

```d2
direction: right
u: User Service {
    class: server
}
n: Notification Service {
    class: server
}
m: "AccountBalanceChanged" {
    class: msg
}
n <- m: 1. Consume
n <- u: 2. Fetch the user information
```

Along with its simplicity, this approach offers **strong consistency** because it interacts with a single data source.
However, a clear disadvantage is that services become tightly coupled and more difficult to evolve.

#### Data Dichotomy

{{< callout type="info" >}}
I found the term in this useful [Confluent blog post](https://www.confluent.io/blog/data-dichotomy-rethinking-the-way-we-treat-data-and-services/) that you might want to review.
{{< /callout >}}

In principle, a service aims to encapsulate its data and **minimize sharing**, exposing only necessary interfaces.
Conversely, a database is designed to **share** its data as widely as possible.
In other words, placing a database behind a service creates a data dichotomy.

```d2
grid-columns: 1
vertical-gap: 100
d: Database {
    class: db
}
d1: Internal data {
    width: 700
}
s: Service {
    class: server
}
d2: "" {
    class: none
    e1: "" {
        class: none
        width: 250
    }
    "Exposed data" {
        width: 200
    }
    e2: "" {
        class: none
        width: 250
    }
}
d -- d1
d1 -> s: Share
s -> d2: Encapsulate
```

As a service grows, it will encompass more data, requiring additional contact points.
The service gradually deviates from its original objectives and starts behaving more **like a database**.

```yaml
UserService:
    GetAllUser()
    GetUserById(userId)
    GetUserByEmail(email)
```

Moreover, since businesses often have core data,
it's easy to fall into the problematic practice of creating a **God Service** (a service with a multitude of consumers).
Maintaining a god service is challenging; it becomes highly restricted, and any modifications can necessitate collaboration with many teams.

```d2
g: God Service (Core data) {
    class: server
}
s1: Service 1 {
    class: server
}
s2: Service 2 {
    class: server
}
s3: Service 3 {
    class: server
}
s4: Service 4 {
    class: server
}
s1 <- g
s2 <- g
s3 <- g
s4 <- g
```

Therefore, sharing data through service interfaces is not a flexible approach.
However, it can be useful when the level of coupling between services is minimal and manageable.

### Data Moving

Another sharing strategy involves moving data from an owner service to consumers,
allowing them to keep and process it **locally**.

```d2
direction: right
u: UserService {
    db: UserDb {
        class: db
    }
}
n: NotificationService {
    db: Copied UserDb {
        class: db
    }
}
n.db <- u.db: Cloned
```

Now, consumer services can operate autonomously with copied data fragments,
which can enhance performance and availability.

This pattern makes the interaction between services become complex.
Data must be fetched from the owner service and kept **in-sync** using a synchronization mechanism.
Fortunately, an **Event Stream** can help address this problem elegantly.

#### Data Moving With Event Streaming

An **Event Stream** acts as a reliable event store,
reducing reliance on service interfaces.
It can be used to move data between services due to its capabilities of:

1. **Event Durability**: Services depend on existing events to initially build their local datastores.
2. **Streaming**: Services continuously capture changes to modify their local datastores.

```d2
direction: right
m: Event Stream {
    class: mq
}
s1: Service {
    store: Local store {
        class: db
    }
}
s1 <- m: Build local data from existing events
s1 <- m: Capture changes from new events {
    style.animated: true
}
```

This forms the foundation of the **Event Sourcing** pattern,
which we will examine in depth in the next section.

## Event Sourcing

### Event

We are quite familiar with this term.
An **event** signifies a fact that occurred in the past, such as `AccountBalanceChanged` or `AccountTransferred`.

Events are primarily triggered by internal components.
Their main responsibility is **notification**,
an event typically does not require a response or any further information.

### Reproducibility

**Event Sourcing** is a pattern that advocates for **logging all events** within the system.
Based on this log, we can **reproduce** the system's state at any given moment.

For example, consider the event log for a bank account:

```yaml
Account A:
    1-Deposit: 50
    2-Withdrawal: 20
    3-Withdrawal: 20
```

While storing only the current balance might seem insufficient,
services can browse through the produced events to display the balance at any point in time.

```d2
direction: right
e: Event Source {
    log: |||yaml
    Account A:
      1-Deposit: "50 -> Balance = 50"
      2-Withdrawal: "20 -> Balance = 30 (50 - 30)"
      3-Withdrawal: "20 -> Balance = 10 (30 - 20)"
    |||
}
s1: Account Service {
    "Current Balance = 10"
}
s1 <- e: Aggregate
```

This characteristic is essential for critical systems, especially in finance, where it's necessary to show how critical values **vary over time**.
Additionally, it helps prove the system's **reliability** across multiple versions,
as log entries can be replayed with different versions to ensure identical results.

```d2
horizontal-gap: 100
e: Event Source {
    log: |||yaml
    Account A:
      1-Deposit: "50 -> Balance = 50"
      2-Withdrawal: "20 -> Balance = 30 (50 - 30)"
      3-Withdrawal: "20 -> Balance = 10 (30 - 20)"
    |||
}
s1: Account Service - v1.0 {
    "Current Balance = 10"
}
s2: Account Service - v2.0 {
    "Current Balance = 10"
}
s1 <- e: Aggregate
s2 <- e: Aggregate
```

Two common challenges arise with this pattern:

- **Storage Growth**: A business operation can generate multiple events.
If every event in the system is logged, the event log can grow dramatically.
- **Increased Complexity**: Events continuously evolve alongside business transformations.
Crucially, it's necessary to ensure that events can be **seamlessly consumed** and integrated with system components.

## Storage Strategies

**Event Sourcing** can lead to an extremely high data volume,
which is daunting in terms of storage costs and potential performance degradation.

### Snapshotting

**Snapshotting** is a retention strategy where old events are compacted and removed from the system.

For example, if we take a snapshot of `AccountBalanceChanged` events at a specific moment (`Event 2`),
we only need to retain later events, as earlier ones become less critical for immediate state reconstruction.

```yaml
Snapshotted Balance At Event 2: 30
Account A:
    # 1-Deposit: 50 (Balance = 50) Removed
    # 2-Withdrawal: 20 (Balance = 30) Removed
    3-Withdrawal: 20 (Balance = 10)
    4-Deposit: 10 (Balance = 20)
```

The retention duration for old events varies based on business requirements:

- Some businesses may only require retention for a few days or weeks.
- More critical systems might need longer durations, such as several months or years.

### Cold Storage

For certain critical events, it may be necessary to retain them **indefinitely**.

However, in practice, a significant percentage of queries tend to focus on the most recent data.
Consequently, maintaining all events in the same high-performance storage may be redundant if older pieces are rarely accessed.

A productive approach is to migrate old events to **much cheaper** storage (such as that built on inexpensive **HDD drives**).
If necessary, historical events can be accessed through this cheaper storage rather than the fast stream.

```d2
e: Event Source {
    class: mq
}
l: Local Storage (Fast SSD) {
    class: hd
}
c: Cold Storage (Cheap HDD) {
    class: hd
}
e <- l: New events
e <- c: Old events
```

## Event Evolution

**Events** need to transform and adapt quickly to business changes.
A flexible system not only evolves its events confidently but also guarantees the compatibility of its event handlers.

### Single Writer

The **Single Writer** principle recommends that events belonging to a specific **topic** should only originate from a single writer (service).
This allows a topic to be autonomously managed by one team, which can then decide when to roll out changes.
If multiple services can publish to the same event topic, ensuring independent evolution becomes exceedingly challenging.

### Additive Changes

The primary approach for evolving events is by only **adding** new fields to the schema.
Modifying or deleting existing fields is prohibited to ensure the compatibility of existing events with older handlers.

```d2
direction: right
v1: |||yaml
AccountUpdated - v1:
    userId: 1234
    name: John Doe
|||
v2: |||yaml
AccountUpdated - v2:
    userId: 1234
    name: John Doe
    address: 1234 Hai Ba Trung HCMC
|||
v3: |||yaml
AccountUpdated - v3:
    userId: 1234
    name: John Doe
    address: 1234 Hai Ba Trung HCMC
    addressDetailed:
        country: Vietnam
        city: HCMC
        street: Hai Ba Trung
        district: 1
        number: 1234
|||
v1 -> v2
v2 -> v3
```

This approach works well for supplementary changes that complete event schemas.
However, business transformation is unpredictable, and the immutability constraint can make events unmanageable.

For instance, if an event modifies a field multiple times, it can grow unnecessarily large and gradually become nonsensical.

```yaml
AccountUpdated:
    userId: 1234
    name: John Doe
    address-1: 1234 Hai Ba Trung HCMC
    address-2:
        country: Vietnam
        city: HCMC
        street: Hai Ba Trung
        district: 1
        number: 1234
    address-3:
        latitude: 10
        longitude: 100
```

### Event Versioning

A more reasonable approach is **Event Versioning**.
In short, an event can exist in different versions simultaneously.
The publisher is required to emit different versions,
and dependent services can freely pick their compatible version to operate.

```d2
grid-rows: 1
p: Publisher {
   class: server
}
e: Event Source {
    grid-columns: 1
    v1: |||yaml
    Account Topic - v1:
        version: v1
        userId: 1234
        address: 1234 Hai Ba Trung HCMC
    |||
    v2: |||yaml
    Account Topic - v2:
        version: v2
        userId: 1234
        address:
            country: Vietnam
            city: HCMC
            street: Hai Ba Trung
            district: 1
            number: 1234
    |||
}
c {
    class: none
    grid-columns: 1
    s1: Service A - using v1 {
        class: server
    }
    s2: Service B - using v2 {
        class: server
    }
}
p -> e.v1
p -> e.v2
e.v1 -> c.s1
e.v2 -> c.s2
```

Despite different release milestones, it's necessary to ensure all versions maintain the same historical data.
For example, if the `v2` topic is introduced after the creation of the record `user1234`,
it must still include this historical record, as shown below:

```yaml
Account Topic - v1:
    - userId: 1234
      address: 1234 Hai Ba Trung HCMC
    # v2 is released
    - userId: 1235
      address: 1235 Ton Duc Thang HCMC

Account Topic - v2:
    - userId: 1234
      address:
        country: Vietnam
        city: HCMC
        street: Hai Ba Trung
        district: 1
        number: 1234
    # v2 is released
    - userId: 1235
      address:
        country: Vietnam
        city: HCMC
        street: Ton Duc Thang
        district: 1
        number: 1235
```

However, this situation should not be maintained indefinitely,
as managing multiple versions simultaneously is cumbersome and error-prone.
The publisher needs to set a timeline before completely deprecating old versions,
giving consumers adequate time to prepare for migration.

## Command Query Responsibility Segregation (CQRS)

**Event Sourcing** alone is extremely inefficient for querying data,
as it requires aggregating all events to retrieve any piece of data.
We will now discuss a pattern that regularly accompanies Event Sourcing to make it truly powerful: **Command Query Responsibility Segregation (CQRS)**.

### Command And Query

#### Command

A **command** is a request intended to change the system's state.
A command is typically synchronous and has a clear result (e.g., `Transfer(toAccount, amount) -> result (failed or success)`).
You can think of it as a normal function or API call.

Commands originate from an actor, such as an end-user, staff member, or a third-party application.
They are usually the root cause of many subsequent **events**.

```d2
direction: right
e: User {
    class: client
}
th: Transfer Service {
    class: process  
}
at: AccountTransferred {
    class: msg
}
ab: AccountBalanceChanged {
    class: msg
}
e -> th: "Transfer(toAccount, amount)"
th -> at
th -> ab
```

#### Query

A **query** refers to a request that looks up information and generates **no side effects** in the system.
In other words, a query will not update the system state (e.g., `getTransaction(transactionId)`, `getUserAccount(userId)`).

#### Command Query Segregation

Essentially, an application supports **Commands** (read-write operations) and **Queries** (read-only operations).
While **Commands** align with business logic, **Queries** typically vary based on different purposes.

For example, bank account transactions can be viewed differently depending on the perspective:

- `End-users` typically need only the most recent transactions.

```yaml
AccountNumber: 1234567890
RecentTransactions:
- Date: 2024-12-10
  Type: Debit
  Amount: 50.00
```

- `Analytical department staff` might require all transactions from the last quarter.

```yaml
AccountNumber: 1234567890
QuarterTransactions:
- Date: 2024-12-10
  Type: Debit
  Amount: 50.00
- Date: 2024-12-05
  Type: Credit
  Amount: 200.00
- Date: 2024-11-22
  Type: Debit
  Amount: 100.00
```

We observe that a single piece of data can have many shapes (or views),
and it may be inefficient to build a single store to serve all of them.
Different views might require **dedicated techniques** (e.g., indexes, materialized views, or denormalized tables) or technologies (e.g., SQL, NoSQL).

### CQRS Pattern

**Command Query Responsibility Segregation (CQRS)** is a pattern that separates the **Command** side (writes) from the **Query** side (reads).

For example, imagine maintaining an **SQL** database for banking accounts and transactions.

- For `end-users`, we need to provide the most recent transactions.
However, pagination tasks are not performed well in SQL (as explained in the [API Design]({{< ref "api-pagination#rowset-pagination" >}}) topic).
Therefore, we can build a **Key-value Store** that caches recent transactions by capturing newly created transactions from the primary database.

```d2
direction: right
main: Main SQL database {
    class: db
}
balance: Key-value store {
    c: |||yaml
    accountId: [Recent transactions]
    |||
}
main -> balance: Sync recent transactions {
    style.animated: true
}
```

- The `analytical department` may want to run advanced search algorithms,
so we might build a separate **Search Engine Store** for them.

```d2
direction: right
main: Main SQL database {
    class: db
}
s: Search engine store {
    class: se
}
main -> s: Sync {
    style.animated: true
}
```

When one store captures changes from another, this is known as **Change Data Capture (CDC)**.
The **CQRS** system might now look like this:

```d2
grid-rows: 2
s: "Account Service" {
    grid-columns: 3
    e: EndUser Query {
        class: process
    }
    acc: Account Command {
        class: process
    }
    a: Analytics Query {
        class: process
    }
}
db: "" {
    grid-columns: 3
    kv: Key-value store {
        class: cache
    }
    main: Main SQL database {
        class: db
    }
    g: Search engine store {
        class: se
    }
    main -> kv {
        style.animated: true
    }
    main -> g {
        style.animated: true
    }
}
s.acc -> db.main: Update
s.e <- db.kv: Query
s.a <- db.g: Query
```

We can see that how data is read is irrelevant to how it was written; this is the essence of **CQRS**.
This pattern is most effective in an **eventual consistency** model, offering:

- **Scalability**: The **Command** and **Query** sides are placed in different stores and can be scaled independently.
- **Performance**: Varied views with different schemas or technologies can efficiently serve specific purposes.

### CQRS And Event Sourcing

When combined with **Event Sourcing**:

- An event source is used for the **Command** side.
- The **Query** stores capture events to independently manage their current state.

For example, an `Account Query` store infers a user's balance from its transactions and updates this value by consuming new transaction events.

```d2
direction: right
grid-rows: 2
grid-gap: 200
c: Account Command {
    class: process
}
e: Event Source (Command Store) {
    s: |||yaml
    Account Events:
      1-Deposit: 50
      2-Deposit: 20
      3-Withdrawal: 30
    |||
}
q: Account Query {
    class: process
}
qs: Query Store {
    "Current Balance = 40"
}
c -> e: 1. Event
q <- e: 2. Pull event {
    style.animated: true
}
q -> qs: 3. Build based on events 
```

To improve this, the **Query** side should periodically take snapshots.
Then, during recovery or initialization processes,
it can reproduce events from the latest snapshot instead of processing every single event from the beginning.

However, **CQRS** can make an application overwhelmingly intricate.
For small systems, the cost of development and maintenance might outweigh the anticipated advantages.
Furthermore, this combination does not provide strong consistency;
**Command** and **Query** must communicate through an asynchronous channel.
