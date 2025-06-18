---
title: Concurrency Control
weight: 30
next: distributed-database
---

Modern databases strive to improve performance by executing jobs **in parallel**.
While this multitasking approach boosts efficiency, it also introduces a unique set of concurrency challenges.

These concurrency issues don't typically throw exceptions or halt execution.
Instead, they quietly degrade data quality, even if the codebase itself remains logically sound.
As a result, applications must anticipate and handle these problems from the outset.

In this topic, we’ll explore concurrency control techniques,
one of the more difficult aspects of maintaining a reliable {{< term sql >}} database.
This knowledge isn’t limited to a single database, it’s also valuable for managing distributed transactions.

## Concurrent Transactions

### Transaction

In most cases, business functionality requires multiple database operations.
A **transaction** is a sequence of operations grouped together as a **single, indivisible unit**.

For example, consider a banking scenario where we want to withdraw `X` from account `A`. The process involves:

1. Verifying if the balance is sufficient: `Balance >= X`
2. Subtracting the amount from account `A`: `Balance = Balance - X`

Although these are two separate steps, they are wrapped into a single `Withdraw` transaction to maintain integrity.

### Race Condition

{{< term sql >}} databases allows multiple transactions to run simultaneously.
A **race condition** occurs when two or more transactions attempt to access and modify the same data concurrently.

In our withdrawal example,
if another transaction initiates a withdrawal while the first one is still in progress,
a race condition could emerge.
Both passed the verification step and then modified the balance concurrently, resulting in an invalid final value.

```d2
shape: sequence_diagram
w: Withdrawal (30)
a: Account (Balance = 50)
d: Withdrawal (40)
a {
  "50"
}
w -> a: Verify balance (50 > 30)
d -> a: Verify balance (50 > 40)
d -> a: Update balance = 50 - 40 = 10 {
    style.bold: true
}
a {
  "10"
}
w -> a: Update balance = 10 - 30 = -20 {
    style.bold: true
}
a {
  "-20?"
}
```

Notice what happens here.
The balance goes wrong without any error or exception,
the system silently produces incorrect results.

## ACID

{{< term acid >}} is a set of four properties that ensure database transactions are processed reliably:

- **Atomicity**
- **Isolation**
- **Consistency**
- **Durability**

It’s important to note that {{< term acid >}} is a consistency model, not a specific tool or library.
Developers must design their transactions to uphold these properties,
though most SQL systems provide mechanisms to support them.

### Atomicity

**Atomicity** ensures a transaction is treated as a **single, all-or-nothing operation**.
It can conclude in two possible ways:

- **Commit**: All changes are successfully applied.
- **Rollback**: No changes are applied, reverting the database to its previous state.

For instance, when transferring funds from account `A` to account `B`,
the transaction should only commit if both debit and credit operations succeed.

```d2
shape: sequence_diagram
a: Account A
t: Transaction
b: Account B
t -> t: Begin
t -> a: Decrease the balance (- X)
t -> b: Increase the balance (+ X)
t -> t: Commit {
    style.bold
}
```

If, for example, account `B` is blocked by the bank and the credit operation fails,
the transaction should **rollback** the debit operation to maintain data integrity.

```d2
shape: sequence_diagram
a: Account A
t: Transaction
b: Account B (Blocked)
t -> a: Decrease the balance (- X)
t -> b: Fail to increase the balance {
  class: error-conn
}
t -> t: Rollback {
  class: error-conn
}
```

Now you might ask:
does a rollback restore the original balance or simply cancel the decrease? The next property explains this.

### Isolation

**Isolation** guarantees that transactions operate independently and their intermediate states remain invisible to others.
Changes become visible only after a transaction **commits**.

Consider a situation where another transaction reads the balance while a transfer is in progress.
In this case, `Another Transaction` will see the original,
unaltered balance because uncommitted changes are isolated within the `Transfer Transaction`.

```d2
shape: sequence_diagram
t: Transfer Transaction
a: Account A
t1: Another Transaction
a {
  "50"
}
t -> a: Decrease the balance (- X)
a {
  "50 - X"
}
t1 -> a: Read the balance (50) {
  style.bold: true
}
t -> t: Commit
```

#### Physical Isolation

Recall how table rows are organized in the [Physical Layer]({{< ref "physical-layer" >}}) topic.
Uncommitted changes don’t overwrite existing data.
Instead, they create **new tuples** alongside the originals, tagged with metadata:

- Commit status.
- IDs of the transactions that inserted, updated, or deleted the tuple.

When a query runs,
the database determines which tuple to return based on the **transactional context** and this metadata.

Let's model this using our transfer example. Initially, there's a committed tuple:.

```d2
Account {
    grid-columns: 1
    grid-gap: 0
    a: "(Id = A, Balance = 50, Committed = True, Transaction Id = 1000)"
}
```

When a new transaction (`Id = 1001`) updates the balance, it creates a new, uncommitted tuple.
Within the same transaction, queries will read the latest version.

```d2
t: Transaction (Id = 1001)
Account {
    grid-columns: 1
    grid-gap: 0
    a: "(Id = A, Balance = 50, Committed = True, Transaction Id = 1000)"
    b: "(Id = A, Balance = 70, Committed = False, Transaction Id = 1001)" {
        style.fill: ${colors.i2}
    }
}
t -> Account.b
```

Based on [Tuple Chaining]({{< ref "physical-layer#tuple-chaining" >}}), other transactions will only see the most recently **committed** tuple.
E.g., the transaction with `Id = 1002` will ignore the dirty row.

```d2

t: Transaction (Id = 1002)
Account {
    grid-columns: 1
    grid-gap: 0
    a: "(Id = A, Balance = 50, Committed = True, Transaction Id = 1000)" {
        style.fill: ${colors.i2}
    }
    b: "(Id = A, Balance = 70, Committed = False, Transaction Id = 1001)"
}
t -> Account.a
```

Rolling back the transaction actually **removes** the uncommitted tuple.

### Consistency

The **consistency** property ensures that a transaction transforms the database from one valid state to another,
either by successfully committing or by rolling back.

But what defines a valid state?
It’s dictated by business rules, such as:

- An account’s balance must equal the total sum of its transactions.
- An account’s balance must not fall below zero.
- And so forth.

#### Trigger Problems

Many developers choose **SQL Triggers** to enforce business rules.
Personally, I avoid using this feature for several reasons:

- First, I prefer not to embed business logic directly into the database through trigger procedures.
This can make applications harder to debug and troubleshoot.
- Second, triggers execute on every update to the associated table,
which can significantly degrade database performance.

Instead, I prefer to enforce data integrity within the application using transactions.
This approach ensures that any potentially unsafe operation is preceded by explicit checks.
For example, a `Withdrawal` transaction should first verify that the user has sufficient balance before proceeding to deduct it.

### Durability

Once a transaction is committed, its changes become permanent, even in the event of a system crash.
This durability is typically achieved through logging mechanisms like
[Write-Ahead Log (WAL)]({{< ref "system-recovery#logging" >}}), which records changes before they’re applied to the database.

## Concurrency Phenomena

Concurrency phenomena are common issues that emerge from [race conditions](#race-condition).
They typically occur due to **inadequate isolation** between concurrent transactions,
leading to inconsistencies and violations of the {{< term acid >}} guarantees.

### Isolation Level

**Isolation Level** determines the degree to which a transaction is isolated from others.
To maintain consistency, transactions might not be fully isolated and may share certain states.
The less they share, the faster transactions can be executed in parallel execute.

Phenomena resulting from concurrent transactions are categorized into **predictable types**.
SQL databases provide different **isolation levels** to manage them:

- Each isolation level addresses one or more specific phenomena.
- Higher isolation levels encompass the capabilities of lower ones.

However, higher isolation levels reduce parallelism and impact performance.
Therefore, selecting the appropriate isolation level is essential to balance consistency and performance.

### Dirty Read

A **Dirty Read** occurs when a transaction reads uncommitted changes made by another transaction,
essentially accessing **dirty records** in progress.

For example, consider two withdrawal transactions:

- The first updates the account balance but later rolls back.
- The second reads the dirty, uncommitted value and updates the balance incorrectly.

```d2

shape: sequence_diagram
d: Withdrawal 1 (10)
a: Account (Balance = 50)
w: Withdrawal 2 (20)
a {
  "50"
}
d -> a: Update balance = 50 - 10 = 40
a {
  "40"
}
w -> a: Update balance = 40 - 20 = 20 {
    style.bold: true
}
d -> a: Rollback {
   class: error-conn
}
```

#### Read Committed Isolation Level

The **Read Committed** isolation level only allows reading committed data.

Let’s see how this resolves the previous issue.
In this case, the second withdrawal ignores the dirty data and
processes against the **last committed** value:

```d2
shape: sequence_diagram
d: Withdrawal 1 (10)
a: Account (Balance = 50)
w: Withdrawal 2 (20)
d -> a: Update balance = 50 - 10 = 40
w -> a: Update balance = 50 - 20 = 30 (not read dirty data) {
    style.bold: true
}
d -> a: Rollback {
   class: error-conn
}
```

This is the default isolation level in many SQL solutions,
and in most systems, reading uncommitted data is rare and generally discouraged.

### Unrepeatable Read

An **Unrepeatable Read** occurs when a transaction reads the same record **twice**,
but the value changes between reads due to another committed transaction.

For instance, imagine an account `A` wants to withdraw `X`. The system must:

1. Verify the account has sufficient funds `A >= X`.
2. Deduct the balance `A = A - X`.

If another transaction intervenes in between these steps and reduces the balance,
it could result in a negative balance.

```d2
shape: sequence_diagram
t: Withdrawal Transaction (40)
a: Account A (50)
t1: Another Withdrawal (30)
a {
  "50"
}
t -> a: Read and verifies the balance (50 > 40)
t1 -> a: Update the balance (50 - 30 = 20)
t1 -> a: Commit the balance {
    style.bold: true
}
a {
  "20"
}
t -> a: Update the balance (20 - 30 = -10?) {
    class: error-conn
}
t -> a: Commit the balance {
    style.bold: true
}
```

Ideally, one of these transactions should fail to preserve consistency.

#### Repeatable Read Isolation Level

The **Repeatable Read** isolation level ensures a transaction sees
only the data committed **before it began**, preventing unrepeatable reads.

At this isolation level, **locking** and **snapshotting** (versioning)
mechanisms work together to maintain consistency.
When multiple transactions access the same data, two main strategies are used:

##### Snapshot Isolation

**Snapshot Isolation** states that transactions see only the data version that existed when they started,
newer updates are **isolated** from them.

{{< callout type="info">}}
The mechanism behind this isolation is discussed [above](#physical-isolation).
{{< /callout >}}

For example, consider two transactions operating on the same account:

- `T1` reads the account balance twice.
- `T2` updates the balance between `T1`'s two reads.

Even though `T2` updates and commits the balance before `T1`'s second read,
`T1` still sees the old value because the update is isolated from its view.

```d2
shape: sequence_diagram
t1: Transaction 1 (T1)
a: Account A (50)
t2: Transaction 2 (T2)
t1 <- a: Read balance (50)
t2 -> a: Update the balance (50 - 40 = 10)
t2 -> a: Commit
t1 <- a: Read balance (50) {
  style.bold: true
}
```

Thus, transactions are guaranteed **repeatable reads**,
data remains consistent for the duration of the transaction, regardless of intermediate changes.

However, this approach works well only when there is a **single writer**.
Multiple read-only transactions can operate on outdated snapshots without issues,
but if **multiple writers** update the same data concurrently, conflicts will arise.

Returning to our earlier example, when two withdrawals happen concurrently;
If a withdrawal transaction continues based on outdated information,
ignoring intervening updates, the final balance will be incorrect.
Below, we can see how a later update can overwrite a previous one:

```d2
shape: sequence_diagram
t: Withdrawal Transaction (40)
a: Account A (50)
t1: Another Withdrawal (30)
a {
  "50"
}
t -> a: Read and verify the balance (50 > 40)
t1 -> a: Update the balance (50 - 30 = 20)
t1 -> a: Commit the balance {
    style.bold: true
}
a {
  "20"
}
t -> a: Update the balance (50 - 40 = 10) (previous update ignored) {
    class: error-conn
}
t -> a: Commit the balance {
    style.bold: true
}
a {
  "10"
}
```

##### Locking Mechanism

At its core, this mechanism ensures that a specific row or table can only be
accessed by **one transaction at a time** to avoid conflicts.
Transactions can acquire **locks** to prevent others from modifying or
accessing the data until the transaction completes.

For example, if a transaction locks an account's row,
other transactions must wait for the release before they can access the row.

```d2
shape: sequence_diagram
t1: Transaction 1
a: Account
t2: Transaction 2
t1 -> a: Acquire a lock
a {
  "LOCKED"
}
t2 -> a: "Acquire a lock"
t2 -> a: Wait... {
    style.bold: true
    style.stroke-dash: 3
}
t1 -> a: Access data
t1 -> a: Release the lock
a {
  "RELEASED"
}
t2 -> a: Access data
```

We commonly encounter two types of operations in database systems: **read** and **write**.
In a highly concurrent environment, using a single, general-purpose lock would be too restrictive.
To manage this more efficiently, we use two distinct types of locks that are **automatically acquired** when needed:

- **Shared Lock (SL):** Might be acquired for **read-only** operations.
  
{{< callout type="info" >}}
In **PostgreSQL**, a simple **SELECT** query does **not** automatically acquire a **Shared Lock**.
However, we can explicitly request one using `SELECT FOR UPDATE`.
{{< /callout >}}

- **Exclusive Lock (XL):** Required for **write** operations, such as **INSERT**, **DELETE**, or **UPDATE**.

We could dedicate thousands of pages to the intricacies of this locking mechanism,
but for now, here are some essential rules you should be familiar with:

1. An **Exclusive Lock (XL)** prevents other transactions from accessing the locked resource until it is released, causing them to wait.

    For example, if `Transaction 1` holds an **XL** on a row,
    then `Transaction 2` must wait until `Transaction 1` releases it before proceeding.

    ```d2
    shape: sequence_diagram
    t1: Transaction 1
    a: Table
    t1 -> a: Acquire an XL
    t2: Transaction 2
    t2 -> a: Acquire another lock (XL or SL)
    t2 -> a: Wait... {
        style.bold: true
        style.stroke-dash: 3
    }
    t1 -> a: Release the lock
    t2 -> a: Access
    ```

    On the other hand, if `Transaction 1` has acquired a **Shared Lock (SL)**,
    then `Transaction 2` must wait if it attempts to acquire an **Exclusive Lock (XL)** on the same data.

    ```d2
    shape: sequence_diagram
    t1: Transaction 1
    a: Table
    t1 -> a: Acquire a SL
    t2: Transaction 2
    t2 -> a: Acquire an XL
    t2 -> a: Wait... {
        style.bold: true
        style.stroke-dash: 3
    }
    t1 -> a: Release the lock
    t2 -> a: Access
    ```

2. **Shared Locks (SL)** don’t block each other.
Multiple read-only transactions can safely acquire **SL** on the same data concurrently,
since none of them intends to modify it.

    For example, both `Transaction 1` and `Transaction 2` can hold an
    **SL** on the same row at the same time without causing any waiting.

    ```d2
    shape: sequence_diagram
    t1: Transaction 1
    a: Table
    t1 -> a: Acquire a SL
    t2: Transaction 2
    t2 -> a: Acquire a SL
    t1 -> a: Read data
    t2 -> a: Read data
    t1 -> a: Release the lock
    t2 -> a: Release the lock
    ```

To summarize this behavior, here’s a simple compatibility matrix:

|                        | **Shared Lock (SL)** | **Exclusive Lock (XL)** |
|------------------------|----------------------|-------------------------|
| **Shared Lock (SL)**   | ✔️ (No block)        | ❌ (Block)            |
| **Exclusive Lock (XL)**| ❌ (Block)           | ❌ (Block)            |

###### Deadlock

By default, read operations don’t acquire a **Shared Lock (SL)** automatically because it's easy to cause deadlocks.
A deadlock occurs when two transactions are waiting for each other to release a lock, and neither can proceed.

Let’s walk through a scenario involving two concurrent withdrawals:

1. Both transactions read and verify the account balance, acquiring shared locks.
Since shared locks are compatible with one another,
both transactions can safely access the record concurrently.

```d2

shape: sequence_diagram
t1: Withdrawal Transaction 1 (40)
a: Account A (50)
t2: Withdrawal Transaction 2 (30)
t1 -> a: Verifies balance (50 > 40) - SL
t2 -> a: Verifies balance (50 > 30) - SL
```

2. Both transactions then attempt to update the balance.
One transaction, e.g., `Transaction 1`, proceeds first and tries to acquire an exclusive lock,
but it’s blocked because the other transaction still holds a shared lock.

```d2

shape: sequence_diagram
t1: Withdrawal Transaction 1 (40)
a: Account A (50)
t2: Withdrawal Transaction 2 (30)
t1 -> a: Verifies balance (50 > 40) - SL
t2 -> a: Verifies balance (50 > 30) - SL
t1 -> a: XL (Waiting for T2's SL...) {
    style.stroke-dash: 3
}
```

3. The second transaction then also attempts to acquire an **XL**,
and is blocked by the first transaction’s **SL**.
Now, both are stuck waiting on each other: a classic deadlock.

```d2

shape: sequence_diagram
t1: Withdrawal Transaction 1 (40)
a: Account A (50)
t2: Withdrawal Transaction 2 (30)
t1 -> a: Verifies balance (50 > 40) - SL
t2 -> a: Verifies balance (50 > 30) - SL
t1 -> a: XL (Waiting for T2's SL...)
t2 -> a: XL (Waiting for T1's SL...)
t1 <-> t2: Deadlock {
  class: error-conn
}
```

**But why not just release `Shared Locks` immediately after reading?**

In theory, a transaction assumes any value it reads remains **stable and consistent**
throughout its execution.
If a transaction were to release a lock too early, another transaction could modify that value,
potentially leading to incorrect or conflicting outcomes if the original transaction tries to access it again.

In other words,
a transaction should only release a lock when it’s
**no longer accessing or depending on that data**.

###### Concurrent Updates With Locking

Returning to the **Repeatable Read** isolation level,
when multiple transactions compete to update the same data, **locking** is necessary to guarantee correctness.

Let’s rewrite the previous deadlock example where transactions do not acquire shared locks.
In the update step, one of the transactions, e.g. `Transaction 1`, will proceed first and acquire an exclusive lock on the record,
while the other must wait.

```d2

shape: sequence_diagram
t1: Withdrawal Transaction 1 (40)
a: Account A (50)
t2: Withdrawal Transaction 2 (30)
t1 -> a: Verifies the balance (50 > 40)
t2 -> a: Verifies the balance (50 > 30)
t1 -> a: Updates the balance (50 - 40 = 10) - XL
t2 -> a: XL (Waiting for T1) {
    style.stroke-dash: 3
}
```

Now, under the isolation level, we divide this scenario into two cases:

1. **Transaction 1 encounters an error and rolls back**: In this case, `Transaction 2` can continue without issues.

```d2

shape: sequence_diagram
t1: Withdrawal Transaction 1 (40)
a: Account A (50)
t2: Withdrawal Transaction 2 (30)
t1 -> a: Verifies the balance (50 > 40) - Read
t2 -> a: Verifies the balance (50 > 30) - Read
t1 -> a: Updates the balance (50 - 40 = 10) - Update
t2 -> a: Wait {
  style.stroke-dash: 3
}
t1 -> t1: Rollback {
  class: error-conn
}
t2 -> a: Updates the balance (50 - 30 = 20) - Update {
  style.bold: true
}
t2 -> t2: Commit
```

2. **Transaction 1 successfully commits**:
Here, `Transaction 2` must roll back because it may have operated on stale data, and applying its changes could result in incorrect outcomes.

```d2

shape: sequence_diagram
t1: Withdrawal Transaction 1 (40)
a: Account A (50)
t2: Withdrawal Transaction 2 (30)
t1 -> a: Verifies the balance (50 > 40) - Read
t2 -> a: Verifies the balance (50 > 30) - Read
t1 -> a: Updates the balance (50 - 40 = 10) - Update
t2 -> a: Wait {
    style.stroke-dash: 3
}
t1 -> t1: Commit
t2 -> t2: Rollback {
    class: error-conn
}
```

In either case, only one transaction commits, and the other is expected to retry.

##### Row-level Conflicts

By combining **locking** and **snapshotting**, we achieve the **Repeatable Read** isolation level.

However, this approach only handles row-level conflicts.
It does not prevent higher-level anomalies happing at the table level,
such as [Phantom Read](#phantom-read) or [Write Skew](#write-skew)

### Phantom Read

The **Phantom Read** phenomenon is somewhat similar to **Unrepeatable Read**,
but occurs at the **table level**.
It refers to a situation where the set of rows matching a condition changes during the course of a transaction.

This happens when rows that satisfy a condition appear or disappear unexpectedly, like phantoms.
Consider an example from a banking system that processes loan applicants:

1. The system first counts the number of eligible accounts and calculates the total loan amount as: `3,000 * number of eligible accounts`.
2. It then creates a new loan program record with the calculated total.
3. Finally, the system loops through the eligible accounts to create individual loan accounts.

Meanwhile, another transaction inserts an additional account that meet the loan eligibility criteria.
As a result, the original transaction processes more accounts than initially counted, leading to a phantom read.

```d2
shape: sequence_diagram
t: Loan Transaction
a: Account Table
l: Loan Table
la: Loan Account Table
t1: Another Transaction
a {
    old: |||yaml
    (AccountId = 2, Income = 2000)
    (AccountId = 3, Income = 1500)
    |||
}
t -> l: 'Create new loan program' {
  style.bold: true
}
l {
    old: |||yaml
    (LoanProgramId = 2, Total = 2 * 3.000 = 6.000)
    |||
}

t1 -> a: Add account 5 {
    class: error-conn
}
a {
    new: |||yaml
    (AccountId = 2, Income = 2000)
    (AccountId = 3, Income = 1500)
    (AccountId = 5, Income = 4000)
    |||
}
t -> la: Create loan accounts {
  style.bold: true
}
la {
    new: |||yaml
    (AccountId = 2, LoanProgramId = 2, Income = 2000)
    (AccountId = 3, LoanProgramId = 2, Income = 1500)
    (AccountId = 5, LoanProgramId = 2, Income = 4000)
    |||
}
```

This results in an inconsistency between the `Loan` and `Loan Account` tables.
While the total loan is calculated as `6,000`, `3` loan accounts are created instead of the expected `2`.

The **Repeatable Read Isolation Level** cannot prevent this issue, as it locks **only individual rows**, not the entire table.
Thus, newly inserted rows remain outside its scope.

To address this, the **Serializable Isolation Level** must be used.
It also solves another phenomenon known as **Write Skew**, which we’ll explore next.

### Write Skew

**Write Skew** occurs when two concurrent transactions read overlapping data and update
different datasets based on their initial reads, ultimately violating business rules.

For example, in a banking system:

- An account qualifies for a loan if its balance is greater than `1000`.
- One transaction verifies the account balance and proceeds to create a loan record.
- Meanwhile, another transaction withdraws funds, lowering the balance to an invalid amount.

```d2
shape: sequence_diagram
t: Loan Transaction
a: "Account A (1100)"
l: Loan Account
t1: "Withdrawal Transaction (700)"
a {
  "1100"
}
t -> a: "Verifies the balance (1100 > 1000)"
t1 -> a: "Deduct the balance 1100 - 700 = 400" {
  style.bold: true
}
t1 -> a: Commit
a {
  "400"
}
t -> l: Creates a new loan account {
  class: error-conn
}
l {
  c: |||yaml
  (AccountId = 1, CheckedBalance = 1100)
  |||
}
```

A loan is issued even though the final balance (`400`) no longer meets the eligibility criteria.
Since the transactions do not compete for updates on the **same row**,
**Repeatable Read** alone can't prevent this anomaly.

#### Serializable Isolation Level

The **Serializable Isolation Level** is the highest isolation level.
It guarantees that transactions produce the same result as if they were executed one after another sequentially.

There are two strategies to implement serializable isolation: **Optimistic** and **Pessimistic**.

##### Optimistic Strategy

The **Optimistic Strategy** assumes that transactions will succeed without conflict,
using **strict locking** to enforce order.

###### Two-phase Locking

**Two-Phase Locking** requires that a transaction to happen in two phases:

1. **Growing Phase**: Acquire all the locks it needs, but not release any.
2. **Shrinking Phase**: Release locks but no longer acquire new ones.

For example, in a withdrawal transaction,
the transaction must acquire an exclusive lock initially.
Since it can't acquire new locks after the **Growing Phase** ends,
it acquires the **strictest** lock it will need.

```d2
shape: sequence_diagram
t: Transaction
a: Account
"1. Growing Phase": {
   t -> a: XL
}
"2. Shrinking Phase": {
   t -> a: Verifies the balance
   t -> a: Decreases the balance
   t -> a: Release XL
}
```

Consider the essence of **Two-phase Locking**:
If any **potential conflict** arises, the growing phase acts as a safeguard,
halting the transaction before it can generate any side effects.

Returning to the loan example:

- The withdrawal transaction must first acquire an exclusive lock
because it needs to update data afterward.
- The `Loan Transaction` acquires a shared lock and remains idle.

```d2

shape: sequence_diagram
t: Loan Transaction
a: "Account A (1100)"
t1: "Withdrawal Transaction (700)"
a {
  "1100"
}
t1 -> a: "Verifies the balance (1100 > 700) - XL"
t -> a: "Wait - SL" {
  style.stroke-dash: 3
  style.bold: true
}
t1 -> a: "Updates the balance to 1100 - 700 = 400"
t1 -> a: Commit and release lock {
  style.bold: true
}
a {
  "400"
}
t -> a: "Verifies the balance and fails (400 < 1.000)" {
  style.bold: true
}
```

While this ensures correctness, it can hinder parallelism,
especially for long-running transactions that lock many rows, reducing system throughput.

##### Pessimistic Strategy

In contrast, the **Pessimistic Strategy** allows transactions to run concurrently but guarantees
serializable behavior by **detecting and resolving conflicts dynamically**.

###### Predicate Locking

**Predicate Locking** is a mechanism used to detect conflicts.
It's not about waiting for locks on tables or rows, but logical predicates in queries.
If a conflict is detected, one transaction is allowed to proceed, while the others are aborted immediately.

Returning to the loan application example.
Suppose two transaction reads the same account (e.g., `A`),
the predicate (`AccountId = A`) is temporarily recorded in memory.

```d2
shape: sequence_diagram
t: Loan Transaction
a: Account A (1100)
t1: "Withdrawal Transaction (700)"
"Predicate AccountId = A" {
   t -> a: "Verifies the balance (1100 > 1000)"
   t1 -> a: "Verifies the balance (1100 > 700)"
}
```

Next, the `Withdrawal Transaction` tries to update the balance.
The database detects that there is an existing predicate (`AccountId = A`) conflicted with the update,
so it aborts **the least costly transaction** (the one that has done less work),
e.g., the `Loan Transaction`.

```d2

shape: sequence_diagram
t: Loan Transaction
a: "Account A (2.000)"
l: Loan
t1: "Withdrawal Transaction (700)"
"Predicate AccountId = A" {
   t -> a: "Verifies the balance (2.000 > 1.000)"
   t1 -> a: "Verifies the balance (1.000 > 700)"
}
"Conflict AccountId = A" {
   t1 -> a: "Updates the balance to 300"
   t -> t: Aborted (because of conflicting predicate) {
      class: error-conn
   }
   t1 -> a: Commit
}
```

What if the loan transaction comes first and commits?
If it releases the predicate lock immediately,
nothing can stop the withdrawal transaction, and we face the anomaly again.

```d2

shape: sequence_diagram
t: Loan Transaction
a: Account A (2.000)
l: Loan
t1: "Withdrawal Transaction (700)"
"Predicate AccountId = A" {
   t -> a: "Verifies the balance (2.000 > 1.000)"
   t1 -> a: "Verifies the balance (1.000 > 700)"
}
t -> l: Creates a new loan record
t -> t: Commit and release {
  style.bold: true
}
t1 -> a: "Updates the balance to 300" {
  class: error-conn
}
```

So, the predicate lock is actually released when **all overlapping** transactions complete.

```d2

shape: sequence_diagram
t: Loan Transaction
a: Account A (2.000)
l: Loan
t1: "Withdrawal Transaction (700)"
"Predicate AccountId = A" {
   t -> a: "Verifies the balance (2.000 > 1.000)"
   t1 -> a: "Verifies the balance (1.000 > 700)"
}
t -> l: Creates a new loan record
t -> t: Commit (the predicate lock is not released here) {
  style.bold: true
}
"Conflict AccountId = A" {
   t1 -> a: "Updates the balance to 300"
   t1 -> t1: Aborted (because of conflicting predicate) {
      class: error-conn
   }
}
```

**Predicate Locking** also supports range queries (e.g., `WHERE value > ...`, `WHERE value < ...`)
and helps prevent **Phantom Reads**.

However, **Predicate Locking** incurs runtime overhead:
Each operation requires tracking predicates and checking for conflicts,
which can significantly impact performance.

This **Pessimistic Strategy** minimizes blocking by enabling concurrency,
making it suitable for complex transactions.
But frequent transaction aborts and the cost of re-execution can offset the benefits.

### Setting Isolation Level

Ultimately, it’s up to developers to choose an appropriate isolation level.
Higher levels prevent more issues but can severely impact performance.

First and foremost, developers should:

- Identify potential anomalies for each transaction.
- Choose the **lowest possible isolation level** that safely prevents them.

There’s no guarantee that the selected isolation levels will be perfect!
Testing under various transaction scenarios is essential to
uncover subtle issues and fine-tune isolation choices.
