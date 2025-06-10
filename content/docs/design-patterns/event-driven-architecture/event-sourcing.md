---
title: Event Sourcing Pattern
weight: 10
---


{{< callout type="info" >}}
You may review the concept of [Event Streaming Platform]({{< ref "event-streaming-platform" >}}) if necessary.
{{< /callout >}}

In this topic, we're going to see a common pattern used in **EDA** systems - **Event Sourcing**.
This pattern helps to share data between teams based on a single source of truth.

## Data Coupling

**Data Coupling** is one of the most critical problems of `EDA`.
Events seldom contain enough information for consumers to handle,
the consumers must seek more from **external data sources**

For example, after receiving an `AccountBalanceChanged`,
the `Notification Service` needs to fetch the user information from the `UserService`
to send the email.

```d2
u: User Service {
    class: server
}
n: Notification Service {
    class: server
}
m: "AccountBalanceChanged"
n <- m
n <- u: Fetch the user information
```

There are some datasets standing in the heart of the business (e.g., user common information),
they’re widely accessed by a lot of services.
We can decouple infrastructure, codebase, workforce...
but data is dynamically born in a specific place.
Although we want the system as loose as possible,
to what extend, data coupling is **inevitable**.

However, letting this problem happen time after time will gradually
proliferate interdependency and defeat the agility of an **EDA** system.

### Service Interface

The most common way to share data is directly using service interfaces.
When we need any piece of information, be free to call the owner server.

This is also what we did in the previous example.
We make a call to the `UserService` with any `AccountBalanceChanged` event.

```d2
u: User Service {
    class: server
}
n: Notification Service {
    class: server
}
m: "AccountBalanceChanged"
n <- m
n <- u: Fetch the user information
```

Alongside with simplicity,
this approach comes with **strong consistency** due to contacting with the single data source.
However, we can fell a disadvantage instantly,
services are tightly coupled and harder to evolve.

#### Data Dichotomy

{{< callout type="info" >}}
I've got in [this useful **Confluent** blog](https://www.confluent.io/blog/data-dichotomy-rethinking-the-way-we-treat-data-and-services/),
you may want to have a look!
{{< /callout >}}

In principle, a service wants to encapsulate and **minimize sharing** its data,
it only exposes necessary interfaces.
However, a database supposes to **share** its data as much as possible.
In other words, placing a database behind a service leads to a data dichotomy.

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
s -> d2: Expose
```

When the service grows bigger,
it will contain more data requiring associated contact points.
The service is gradually drifted from its objectives, behaving as a database instead.

```d2
us: User Service {  
    class: server
}
i: "Interface" {
    "GetAllUsers();"
    "GetUserById(userId);"
    "GetUserByEmail(email);"
}
```

Moreover, as businesses often have some core data parts,
it's easy to fall into the bad practice of `God Service`,
when a service has a bunch of consumers.
Maintaining a god service is a nightmare,
it's strongly restricted, any changes possibly conduce collaborations with a lot of teams.

```d2
g: God Service {
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

Therefore, sharing data through service interface is not a flexible approach.
Though, we may find it helpful when the level of coupling between services
is small and controllable.

### Data Moving

Another approach of sharing is moving data from an owner service to consumers,
they can keep and process it **locally**.

```d2
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

Now, the consumer services can autonomously operate with copied fragments,
stimulating performance and availability.

However, in the diagram, we obviously make the interfaces extremely challenging,
we need to fetch data from the owner service and keep it **in-sync** with a synchronization mechanism.
Hopefully, **Event Steaming Platform** can help us solve the problem elegantly.

#### Data Moving With Event Streaming

**Event Steaming Platform** performs as a reliable event store,
helping evade from purely relying on service interfaces.
We can rely on it to move data between services due to the capability of:

1. **Event Durability**: services depend on existing events to initially build their store.
2. **Streaming**: services continuously capture changes to modify their store.

```d2
direction: right
m: Event store {
    class: mq
}
s1: Service {
    store: Datastore {
        class: db
    }
}
s1 <- m: 1. Build data from existing events
s1 <- m: 2. Capture changes from new events {
    style.animated: true
}
```

This is the foundation of the **Event Sourcing** pattern.
We will view it in depth in the next section.

## Event Sourcing

### Event

We're absolutely familiar with this term.
An **event** indicates a fact happened in the past,
e.g. `AccountBalanceChanged`, `AccountTransferred`, etc.

Event is primarily caused by internal components.
The primary responsibility is **notifying** (**fire-and-forget**),
event doesn't need a response or any further information.

```d2
direction: right
at: AccountTransferred {
    class: event
}
th: Transfer Handler {
    class: process  
}
ab: AccountBalanceChanged {
    class: event
}
bh: Balance Handler {
    class: process  
}
th -> at
at -> bh 
bh -> ab
```

### Reproducibility

**Event Sourcing** is a pattern suggesting **logging all events** within the system.
Based on the log, we can **reproduce** to retrieve the system's state at any moment.

For example, we have a log of a banking account.

```yaml
Account A:
    1-Deposit: 50 (Balance = 50)
    2-Withdrawal: 20 (Balance = 30)
    3-Withdrawal: 20 (Balance = 10)
```

While storing only the current balance is sketchy,
services can browse through the produced events to display the balance at any point of time.

```d2
e: Event Source {
    log: |||yaml
    Account A:
      1-Deposit: "50 -> Balance = 50"
      2-Withdrawal: "20 -> Balance = 30 (50 - 30)"
      3-Withdrawal: "20 -> Balance = 10 (30 - 20)"
    |||
}
s1: Account Service {
    "Balance = 10"
}
s1 <- e: Aggregate
```

This characteristic is a must to critical systems, especially finance,
when we need to show how critical values **vary overtime**.
Additionally, it helps prove the system's **reliability** across multiple versions,
as we can replay log entries with different versions and ensure their result are identical.

```d2
e: Event Source {
    class: mq
}
s1: Account Service - version 1.0 {
    20
}
s2: Account Service - version 1.1 {
    20
}
s1 <- e: Aggregate
s2 <- e: Aggregate
```

There are two common problems in this pattern:

- **Storage Growth**: a business operation possibly causes some events.
If we log everything happened in the system, the event log will grow dramatically
- **Increased Complexity**: events continuously evolve beside the business transformation.
Most importantly, we need to ensure events to be **seamlessly consumed** and integrated with system components.

## Storage Strategies

**Event Sourcing** leads to an extremely high data volume,
which daunting in terms of costly storage and performance downgrading.

### Snapshotting

**Snapshotting** a retention strategy where old events are compacted and removed from the system.

For example, we take a snapshot of `AccountBalanceChanged` events at a moment (`Event 2`).
Therefore, we can only retain later events, as earlier ones are necessarily important.

```yaml
Snapshotted Balance At Event 2: 20
Account A:
    # 1-Deposit: 50 (Balance = 50) Removed
    # 2-Withdrawal: 20 (Balance = 30) Removed
    3-Withdrawal: 20 (Balance = 10)
    4-Deposit: 10 (Balance = 20)
```

The retention duration of old events is varied based on each business.

- Some only requires few days or weeks.
- More critical ones need longer durations like several months and years.

### Cold Storage

For certain critical events,
we're supposed to retain events **indefinitely**.

In practice, a significant percentage of queries tend to focus on the most recent data.
As a result, maintaining all events in the same storage may be redundant,
when old pieces are rarely accessed.

We can productively migrate old events to a **much cheaper** storage (such as built on dirt-cheap **HDD drives**).
If necessary, historical events are accessed through the cheap store not the fast stream.

```d2
e: Event Source {
    class: mq
}
l: Local Storage (Fast SSD drives) {
    class: hd
}
c: Cold Storage (Cheap SSD drives) {
    class: hd
}
p <- l: New events
p <- c: Old events
```

## Event Evolution

**Event** needs to quickly transform and adapt to its business.
A flexible system doesn't only confidently evolve its events,
but also guarantees the compatibility of event handlers.

### Single Writer

The **Single Writer** principle recommends that events belonging to a **topic**
should be only originated by single writer (or service).

We want the topic to be autonomously managed by a team, deciding when to roll out of changes.
If many services can fire the same event topic, it's truly challenging to ensure the independent evolution.

### Additive Changes

The first approach of thriving event is only **adding** fields to the schema,
modifying or deleting existing fields is prohibited,
helps to make sure the compatibility of existing events.

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

v1 -> v2: "Add 'address'"
v2 -> v3: "Add 'addressDetailed'"
```

This approach works well with supplement changes completing event schemas.
However, business transformation is unpredictable,
the immutability constraint can make events unmanageable.
For examples, when an event modifies a field many times,
it grows unnecessarily bigger and gradually becomes absurd.

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

A more reasonable approach is **versioning event**.
In short, an event can live in different versions simultaneously.
The publisher is required to fire different versions,
dependent services freely pick their compatible version to operate

```d2
direction: right
p: Publisher {
   class: server
}
s: Event Source {
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
s1: Service A - using v1 {
    class: server
}
s2: Service B - using v2 {
    class: server
}
p -> s: v1 + v2
s1 -> s.v1
s2 -> s.v2
```

Despite different release milestones,
we need to ensure all versions to maintain the same historical data.
For example, the `v2` topic is introduced after the creation of the record `user1234`.
However, it still includes the historical record, as shown below:

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

But we don't want to retain this situation forever,
since maintaining multiple versions at once is bothersome and error-prone.
The publisher needs to mark a time before completely deprecate old versions,
helping consumer have time to prepare for the migration.

## Command Query Responsibility Segregation (CQRS)

**Event Sourcing** alone is extremely inefficient to query data,
like we need to aggregate everything before to retrieve any piece of data.
We will talk about a pattern regularly going together
with **Event Sourcing** to make it real powerful - **CQRS**.

### Command And Query

#### Command

A **command** is a request expecting to change the system state.
A command is typically synchronous with an obvious result,
you might think it as a normal function or API call,
e.g. `Transfer(toAccount, amount) -> result (failed or success)`

A command is originated from an actor,
e.g. end-users, staff or third-party applications.
It's usually the root cause of many events afterward

```d2
direction: right
e: User {
    class: user
}
th: Transfer Service {
    class: process  
}
at: AccountTransferred {
    class: event
}
ab: AccountBalanceChanged {
    class: event
}
e -> th: "Transfer(toAccount, amount)"
th -> at
at -> ab
```

#### Query

**Query** refers to requests **looking up** information
and generating no side effects in the system.
In other words, a query will not update the system state
e.g. `getTransaction(transactionId)`, `getUserAccount(userId)`, etc.

### Command Query Segregation

Basically, an application supports **Commands** (**read-write** operations) and **Queries** (**readonly** operations).
While **Command** is aligned with the business logic,
**Query** typically varies with different purposes.

For example, bank account transactions are divergent based on current perspective:

- `End-users` need only the most recent transactions.

```yaml
AccountNumber: 1234567890
RecentTransactions:
  - Date: 2024-12-10
    Type: Debit
    Amount: 50.00
  - Date: 2024-12-05
    Type: Credit
    Amount: 200.00
```

- `Analytical department staff` require all transactions in the last quarter.

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
  - Date: 2024-10-03
    Type: Credit
    Amount: 70.00
```

We see that a fragment of data can have many shapes (or views),
and it may be inefficient if we build a single store to serve them all.
Different views possibly require **dedicated techniques** (e.g., index, material view, or denormalized tables...)
or technologies (e.g., SQL, NoSQL).

### CQRS Pattern

**Command Query Responsibility Segregation (CQRS)** is a pattern
separating the **Command** from the **Query** side.

For example, we're maintaining an `SQL` database of banking accounts of transactions.

```d2
Account {
    shape: sql_table
    id: string
    balance: number
}
Transaction {
    shape: sql_table
    id: string
    fromAccount: string
    toAccount: string
    amount: number
}
```

- For `end-users`, we need to answer the most recent transactions.
However, this pagination task is not well-performed (we've explained in the [API Design](API-Design.md#re-querying) topic) in SQL.
Therefore, we build a **Key-value Store** caching the recent transactions by capturing created transactions from the primary database.

```d2
main: Main SQL database {
    class: db
}
balance: Key-value store {
    c: |||yaml
    accountId: [Recent transactions]
    |||
}
main -> balance: Transaction {
    style.animated: true
}
```

- The analytical department wants to run advanced search algorithms,
so we may build a separated **Search engine Store** for them.

```d2
main: Main SQL database {
    class: db
}
s: Search engine store {
    class: se
}
main -> s: Transaction {
    style.animated: true
}
```

When a store captures changes from another store, we call it as **Change Data Capture (CDC)**.
Now, the **CQRS** system may look like this:

```d2
s: "Account Service" {
    grid-rows: 3
    acc: Account Command {
        class: process
    }
    e: EndUser Query {
        class: process
    }
    a: Analytics Query {
        class: process
    }
}
db: "" {
    grid-rows: 3
    main: Main SQL database {
        class: db
    }
    kv: Key-value store {
        class: cache
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

We see that how data is read is irrelevant to the way it was written,
that's the essence of `CQRS`.
This pattern makes the most of it in the **eventual consistency** model:

- Scalability: the **Command** and **Query** are placed in different stores and independently scaled.
- Performance: varied views with different schemas or technologies efficiently serve for certain purposes.

However, **CQRS** may make the application overwhelmingly intricate.
For small systems, the cost of development and maintenance may outweigh the expected advantage.

### CQRS And Event Sourcing

Combined with **Event Sourcing**, we use:

- An event source for the **Command** side.
- The `Query` stores capture events to independently manage their state,
we don't need to aggregate states of **Event Sourcing** repetitively.

For example, `Account Query` infers a user's balance from its transactions,
and updates the value by consuming new transactions.

```d2
c: Account Command {
    class: process
}
e: Event Source {
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
    s: |||yaml
    Balance: 40
    ||| 
}
c -> e: 1. Event
q <- e: 2. Pull event {
    style.animated: true
}
q -> qs: 3. Build based on events 
```

To make it better, the **Query** side should periodically take snapshots.
Then, on recovery or initialization processes,
it can reproduce events from the latest snapshot instead of handling every event.

However, this combination does not provide [strong consistency](Distributed-Database.md#strong-consistency-level),
**Command** and **Query** must communicate through an asynchronous channel.
