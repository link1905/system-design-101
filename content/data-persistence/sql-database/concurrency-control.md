---
title: Concurrency Control
weight: 30
next: distributed-database
---

Modern databases strive to improve performance by executing operations **in parallel**.
While this concurrent approach improves efficiency, it also introduces a distinct set of concurrency challenges.

These concurrency issues do not typically throw exceptions or halt execution.
Instead, they can *silently compromise data quality*, even when the codebase itself appears logically sound.
As a result, applications must anticipate and handle these problems from the outset.

In this topic, we'll explore concurrency control techniques,
one of the more challenging aspects of maintaining a reliable {{< term sql >}} database.
This knowledge is not limited to a single database system. It is also valuable when managing distributed transactions.

## Concurrent Transactions

### Transaction

In many cases, business functionality requires multiple database operations.
A **transaction** is a sequence of operations grouped together as a **single, indivisible unit**.

For example, consider a banking scenario in which we want to withdraw `X` from account `A`. The process involves:

1. Verifying that the balance is sufficient: `Balance >= X`
2. Subtracting the amount from account `A`: `Balance = Balance - X`

Although these are two separate steps, they are wrapped in a single `Withdraw` transaction to maintain data integrity.

### Race Condition

{{< term sql >}} databases allow multiple transactions to run *simultaneously*.
A **race condition** occurs when two or more transactions access and modify the same data concurrently in a way that makes the outcome depend on execution timing.

In our withdrawal example,
if another transaction initiates a withdrawal while the first is still in progress,
a race condition can occur.
Both transactions may pass the verification step and then modify the balance concurrently, resulting in an invalid final value.

```d2
shape: sequence_diagram
w: Withdrawal (X = 30)
a: Account (Balance = 50)
d: Withdrawal (X = 40)
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
The balance becomes incorrect without producing an error or exception.
The system silently produces an invalid result.

## ACID

{{< term acid >}} is a set of four properties that help ensure database transactions are processed reliably:

- **Atomicity**
- **Isolation**
- **Consistency**
- **Durability**

It is important to note that {{< term acid >}} is a set of transaction properties, not a specific tool or library.
Developers must design their transactions to uphold these properties,
although most SQL systems provide mechanisms to support them.

### Atomicity

**Atomicity** ensures that a transaction is treated as a **single, all-or-nothing operation**.
It can end in two possible ways:

- **Commit**: All changes are successfully applied.
- **Rollback**: No changes are applied, restoring the database to its previous state.

For instance, when transferring funds from account `A` to account `B`,
the transaction should commit only if both the debit and credit operations succeed.

```d2
shape: sequence_diagram
a: Account A
t: Transferring Transaction (X)
b: Account B
t -> t: Begin
t -> a: Decrease the balance (- X)
t -> b: Increase the balance (+ X)
t -> t: Commit {
    style.bold
}
```

If, for example, account `B` is blocked by the bank and the credit operation fails,
the transaction should **roll back** the debit operation to preserve data integrity.

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

At this point, you might ask:
*Does a rollback restore the original balance, or does it simply cancel the decrease?*
The next property helps explain this behavior.

### Isolation

**Isolation** guarantees that transactions operate independently and that their intermediate states remain invisible to others.
Changes become visible only after a transaction commits.

Consider a situation in which another transaction reads the balance while a transfer is in progress.
In this case, `Another Transaction` sees the original,
unaltered balance because the uncommitted changes are isolated within the `Transfer Transaction`.

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
t1 -> a: Read the balance = 50 {
  style.bold: true
}
t -> t: Commit
```

#### Physical Isolation

Recall how table rows are organized in the [Physical Layer]({{< ref "physical-layer" >}}) topic.
Uncommitted changes do not overwrite existing data.
Instead, they create **new tuples** alongside the original ones, tagged with metadata such as:

- Commit status.
- The transaction ID that inserted, updated, or deleted the tuple.

When a query runs,
the database determines which tuple to return based on the **transactional context** and the associated metadata.

Let's model this using our transfer example. Initially, there is a committed tuple:

```d2
Account {
    grid-columns: 1
    grid-gap: 0
    a: "(Id = A, Balance = 50, Committed = True, Transaction Id = 1000)"
}
```

When a new transaction (`Id = 1001`) updates the balance, it creates a new **uncommitted tuple**.
Within the same transaction, queries read the latest version.

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

Based on [Tuple Chaining]({{< ref "physical-layer#tuple-chaining" >}}), other transactions see only the *most recently committed* tuple.
For example, the transaction with `Id = 1002` ignores the dirty row.

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

Rolling back the transaction effectively means *removing the uncommitted tuple*.

### Consistency

The **consistency** property ensures that a transaction transforms the database from *one valid state to another*,
either by successfully committing or by rolling back.

But what defines a valid state?
It is determined by business rules, such as:

- An account's balance must equal the total sum of its transactions.
- An account's balance must not fall below zero.
- And so forth.

#### Trigger Problems

Some developers use **SQL Triggers** to enforce business rules.
Personally, I avoid using this feature for several reasons:

- First, I prefer not to embed business logic directly in the database through trigger procedures.
  This can make applications more difficult to debug and troubleshoot.
- Second, triggers execute whenever the associated operation occurs,
  which can degrade database performance.

Instead, I prefer to enforce data integrity within the application using transactions.
This approach allows potentially unsafe operations to be preceded by explicit checks.
For example, a `Withdrawal` transaction should first verify that the user has a sufficient balance before deducting funds.

### Durability

Once a transaction is committed, its changes become permanent, even in the event of a system crash.
This durability is typically achieved through logging mechanisms such as
[Write-Ahead Log (WAL)]({{< ref "system-recovery#logging" >}}), which records changes before they are applied to the database.

## Concurrency Phenomena

Concurrency phenomena are common anomalies that can emerge from [race conditions](#race-condition).
They typically occur because of **inadequate isolation** between concurrent transactions,
leading to inconsistencies and violations of the {{< term acid >}} guarantees.

### Isolation Level

An **Isolation Level** determines the degree to which a transaction is isolated from other transactions.

Phenomena resulting from concurrent transactions can be categorized into **well-known types**.
SQL databases provide different **isolation levels** to manage them:

- Each isolation level addresses one or more specific phenomena.
- Higher isolation levels encompass the capabilities of lower ones.

However, higher isolation levels reduce parallelism and can affect performance.
Therefore, selecting an appropriate isolation level is essential for balancing consistency and performance.

Let's discuss common phenomena and the strategies used to address them.

## Dirty Read

A **Dirty Read** occurs when a transaction reads uncommitted changes made by another transaction,
effectively accessing **dirty records** whose changes are still in progress.

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

### Read Committed Isolation Level

The **Read Committed** isolation level allows transactions to read only committed data.

Let's see how this resolves the previous issue.
In this case, the second withdrawal ignores the dirty data and
operates against the **last committed** value:

```d2
shape: sequence_diagram
d: Withdrawal 1 (10)
a: Account (Balance = 50)
w: Withdrawal 2 (20)
d -> a: Update balance = 50 - 10 = 40
w -> a: Update balance = 50 - 20 = 30 (does not read dirty data) {
    style.bold: true
}
d -> a: Rollback {
   class: error-conn
}
```

This is the default isolation level in many SQL systems,
and in most applications, reading uncommitted data is an uncommon requirement.

## Unrepeatable Read

An **Unrepeatable Read** occurs when a transaction reads the same record **twice**,
but the value changes between those reads because another transaction commits an update.

For instance, imagine that account `A` wants to withdraw `X`. The system must:

1. Verify that the account has sufficient funds: `A >= X`.
2. Deduct the amount from the balance: `A = A - X`.

If another transaction intervenes between these steps and reduces the balance,
the operation could result in a negative balance.

```d2
shape: sequence_diagram
t: Withdrawal Transaction (30)
a: Account A (50)
t1: Another Withdrawal (30)
a {
  "50"
}
t -> a: Read and verify the balance (50 > 30)
t1 -> a: Update the balance (50 - 30 = 20)
t1 -> a: Commit the balance {
    style.bold: true
}
a {
  "20"
}
t -> a: Update the balance (20 - 30 = -10) {
    class: error-conn
}
a {
  "-10?"
}
t -> a: Commit the balance {
    style.bold: true
}
```

To preserve consistency, one of these transactions should ideally fail.

### Repeatable Read Isolation Level

The **Repeatable Read** isolation level ensures that a transaction sees
a single version of the data, preventing unrepeatable reads.

At this isolation level, **locking** and **snapshotting** (versioning)
mechanisms may be used to maintain consistency.
When multiple transactions access the same data, two main strategies are used.

#### Snapshot Isolation

**Snapshot Isolation** means that transactions see only the version of the data that existed when they started.
Newer updates are **isolated** from them.

{{< callout type="info">}}
The mechanism behind this isolation is discussed [above](#physical-isolation).
{{< /callout >}}

For example, consider two transactions operating on the same account:

- `T1` reads the account balance twice.
- `T2` updates the balance between `T1`'s two reads.

Even though `T2` updates and commits the balance before `T1` performs its second read,
`T1` still sees the old value because the newly created tuple from `T2` is isolated from its view.

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

Thus, transactions achieve **repeatable reads**:
the data remains consistent for the duration of the transaction despite intermediate changes made by others.

Multiple read-only transactions can operate on older snapshots without issues,
but conflicts can arise when **multiple writers** update the same data concurrently.

Returning to our earlier example, consider two withdrawals occurring concurrently.
If a withdrawal transaction continues based on outdated information
while ignoring intervening updates, the final balance can become incorrect.
The following example shows how a later update can overwrite an earlier one:

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
a {
  "10"
}
t -> a: Commit the balance {
    style.bold: true
}
```

#### Locking Mechanism

At its core, this mechanism coordinates access to a specific row or table
to avoid conflicting operations.
Transactions can acquire **locks** to prevent others from modifying or
accessing the data until the transaction completes.

For example, if a transaction locks an account row,
other transactions **must wait** for the lock to be released before they can access the row.

```d2
shape: sequence_diagram
t1: Transaction 1
a: Account
t2: Transaction 2
t1 -> a: Acquire a lock {
  style.bold: true
}
a {
  "LOCKED"
}
t2 -> a: "Acquire a lock"
t2 -> a: Wait... {
    style.bold: true
    style.stroke-dash: 3
}
t1 -> a: Access data
t1 -> a: Release the lock {
  style.bold: true
}
a {
  "RELEASED"
}
t2 -> a: Access data
```

Database systems commonly perform two fundamental types of operations: **reads** and **writes**.
In a highly concurrent environment, using a single general-purpose lock would be too restrictive.
To manage concurrency more efficiently, database systems commonly use two fundamental lock types:

- **Shared Lock (SL):** May be acquired for **read-only** operations.
- **Exclusive Lock (XL):** Required for **write** operations such as **INSERT**, **DELETE**, or **UPDATE**.

We could dedicate dozens of pages to the intricacies of locking,
but for now, here are some essential rules to understand:

1. An **Exclusive Lock (XL)** prevents other transactions from accessing the locked resource until the lock is released, causing them to wait.

    For example, if `Transaction 1` holds an **XL** on a row,
    `Transaction 2` must wait until `Transaction 1` releases it before proceeding.

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
    `Transaction 2` must wait if it attempts to acquire an **Exclusive Lock (XL)** on the same data.

    ```d2
    shape: sequence_diagram
    t1: Transaction 1
    a: Table
    t1 -> a: Acquire an SL
    t2: Transaction 2
    t2 -> a: Acquire an XL
    t2 -> a: Wait... {
        style.bold: true
        style.stroke-dash: 3
    }
    t1 -> a: Release the lock
    t2 -> a: Access
    ```

2. **Shared Locks (SL)** do not block one another.
   Multiple read-only transactions can safely acquire **SLs** on the same data concurrently,
   since none of them intends to modify it.

    For example, both `Transaction 1` and `Transaction 2` can hold an
    **SL** on the same row simultaneously without causing either transaction to wait.

    ```d2
    shape: sequence_diagram
    t1: Transaction 1
    a: Table
    t1 -> a: Acquire an SL
    t2: Transaction 2
    t2 -> a: Acquire an SL
    t1 -> a: Read data
    t2 -> a: Read data
    t1 -> a: Release the lock
    t2 -> a: Release the lock
    ```

To summarize this behavior, here is a simple compatibility matrix:

|                         | **Shared Lock (SL)** | **Exclusive Lock (XL)** |
|-------------------------|----------------------|-------------------------|
| **Shared Lock (SL)**    | ✔️ (No block)       | ❌ (Block)              |
| **Exclusive Lock (XL)** | ❌ (Block)           | ❌ (Block)              |

##### Deadlock

A deadlock occurs when two transactions wait for each other to release a lock, leaving neither able to proceed.

Let's walk through a scenario involving two concurrent withdrawals:

1. Both transactions read and verify the account balance, acquiring shared locks.
   Since shared locks are compatible with one another,
   both transactions can safely access the record concurrently.

    ```d2
    shape: sequence_diagram
    t1: Withdrawal Transaction 1 (40)
    a: Account A (50)
    t2: Withdrawal Transaction 2 (30)
    t1 -> a: Verify balance (50 > 40) - SL
    t2 -> a: Verify balance (50 > 30) - SL
    ```

2. Both transactions then attempt to update the balance.
   One transaction, for example `Transaction 1`, proceeds first and attempts to acquire an exclusive lock,
   but it is blocked because the other transaction still holds a shared lock.

    ```d2
    shape: sequence_diagram
    t1: Withdrawal Transaction 1 (40)
    a: Account A (50)
    t2: Withdrawal Transaction 2 (30)
    t1 -> a: Verify balance (50 > 40) - SL
    t2 -> a: Verify balance (50 > 30) - SL
    t1 -> a: XL (Waiting for T2's SL...) {
        style.stroke-dash: 3
    }
    ```

3. The second transaction then also attempts to acquire an **XL**
   and is blocked by the first transaction's **SL**.
   Both transactions are now waiting for each other, creating a classic deadlock scenario.

    ```d2
    shape: sequence_diagram
    t1: Withdrawal Transaction 1 (40)
    a: Account A (50)
    t2: Withdrawal Transaction 2 (30)
    t1 -> a: Verify balance (50 > 40) - SL
    t2 -> a: Verify balance (50 > 30) - SL
    t1 -> a: XL (Waiting for T2's SL...)
    t2 -> a: XL (Waiting for T1's SL...)
    t1 <-> t2: Deadlock {
      class: error-conn
    }
    ```

**But why not simply release `Shared Locks` immediately after reading?**

In this locking model, a transaction may depend on a value remaining **stable and consistent**
throughout its execution.
If a transaction releases a lock too early, another transaction could modify that value,
potentially causing incorrect or conflicting results if the original transaction accesses it again.

In other words,
a transaction should release a lock only when it is
**no longer accessing or depending on that data**.

##### Concurrent Updates With Locking

Returning to the **Repeatable Read** isolation level,
when multiple transactions compete to update the same data, **locking** is necessary to guarantee correctness.

Let's revise the previous deadlock example so that the transactions do not acquire shared locks.
During the update step, one of the transactions, such as `Transaction 1`, proceeds first and acquires an exclusive lock on the record,
while the other must wait.

```d2
shape: sequence_diagram
t1: Withdrawal Transaction 1 (40)
a: Account A (50)
t2: Withdrawal Transaction 2 (30)
t1 -> a: Verify the balance (50 > 40)
t2 -> a: Verify the balance (50 > 30)
t1 -> a: Acquire XL
t1 -> a: Update the balance (50 - 40 = 10)
t2 -> a: Acquire XL (Waiting for T1) {
    style.stroke-dash: 3
}
```

Under this model, we can divide the scenario into two cases:

1. **Transaction 1 encounters an error and rolls back**:
   In this case, `Transaction 2` can continue without issues.

    ```d2
    shape: sequence_diagram
    t1: Withdrawal Transaction 1 (40)
    a: Account A (50)
    t2: Withdrawal Transaction 2 (30)
    t1 -> a: Verify the balance (50 > 40) - Read
    t2 -> a: Verify the balance (50 > 30) - Read
    t1 -> a: Update the balance (50 - 40 = 10) - Update
    t2 -> a: Wait {
      style.stroke-dash: 3
    }
    t1 -> t1: Rollback {
      class: error-conn
    }
    t2 -> a: Update the balance (50 - 30 = 20) - Update {
      style.bold: true
    }
    t2 -> t2: Commit
    ```

2. **Transaction 1 successfully commits**:
   In this case, `Transaction 2` must roll back because it may have operated on stale data,
   and applying its changes could produce an incorrect result.

    ```d2
    shape: sequence_diagram
    t1: Withdrawal Transaction 1 (40)
    a: Account A (50)
    t2: Withdrawal Transaction 2 (30)
    t1 -> a: Verify the balance (50 > 40) - Read
    t2 -> a: Verify the balance (50 > 30) - Read
    t1 -> a: Update the balance (50 - 40 = 10) - Update
    t2 -> a: Wait {
        style.stroke-dash: 3
    }
    t1 -> t1: Commit
    t2 -> t2: Rollback {
        class: error-conn
    }
    ```

In either case, only one transaction commits, while the other must retry.
By combining **locking** and **snapshotting**, we achieve the **Repeatable Read** isolation level.

However, there are cases that this approach cannot resolve,
such as [Write Skew](#write-skew).

## Write Skew

**Write Skew** occurs when two concurrent transactions read overlapping data and update
different datasets based on their initial reads, ultimately violating business rules.

For example, consider a banking system:

- An account qualifies for submitting a loan application if its balance is greater than `1000`.
- One transaction verifies the account balance and proceeds to create an applicant record.
- Meanwhile, another transaction withdraws funds, reducing the balance below the required threshold.

```d2
shape: sequence_diagram
t: Application Transaction
a: "Account A (1100)"
l: Loan Applicant
t1: "Withdrawal Transaction (700)"
a {
  "1100"
}
t -> a: "Verify the balance (1100 > 1000)"
t1 -> a: "Deduct the balance 1100 - 700 = 400" {
  style.bold: true
}
t1 -> a: Commit
a {
  "400"
}
t -> l: Create a new applicant {
  class: error-conn
}
l {
  c: |||
  AccountId = 1, CheckedBalance = 1100
  |||
}
```

A new applicant is created even though the final balance (`400`) no longer satisfies the eligibility criteria.
Since the transactions do not compete for updates on the **same row**,
**Repeatable Read** alone cannot prevent this anomaly.

### Serializable Isolation Level

The **Serializable Isolation Level** is the highest isolation level.
It guarantees that transactions produce the same result **as if they had been executed sequentially, one after another**.

Serializable isolation can be implemented using **Optimistic** or **Pessimistic** strategies.

#### Pessimistic Strategy

The **Pessimistic Strategy** assumes that conflicts might occur and locks data before modifying it,
using **strict locking** to enforce execution order.

##### Two-Phase Locking

**Two-Phase Locking** requires a transaction to proceed through two phases:

1. **Growing Phase**: Acquire all required locks without releasing any.
2. **Shrinking Phase**: Release locks without acquiring any new ones.

For example, in a withdrawal transaction,
the transaction must initially acquire an exclusive lock.
Since it cannot acquire new locks after the **Growing Phase** ends,
it acquires the strictest lock it will eventually need.

```d2
shape: sequence_diagram
t: Transaction
a: Account
"1. Growing Phase": {
   t -> a: XL
   t -> a: Verify the balance
   t -> a: Decrease the balance
}
"2. Shrinking Phase": {
   t -> a: Release XL
}
```

The key idea behind **Two-Phase Locking** is that
if a potential conflict arises, the growing phase acts as a safeguard,
preventing the transaction from proceeding until the required locks are acquired.

Returning to the loan application example:

- The withdrawal transaction must first acquire an exclusive lock
  because it needs to update the data afterward.
- The `Application Transaction` attempts to acquire a shared lock and must wait.

```d2
shape: sequence_diagram
t: Application Transaction
a: "Account A (1100)"
t1: "Withdrawal Transaction (700)"
a {
  "1100"
}
t1 -> a: "Verify the balance (1100 > 700) - XL"
t -> a: "Wait - SL" {
  style.stroke-dash: 3
  style.bold: true
}
t1 -> a: "Update the balance to 1100 - 700 = 400"
t1 -> a: Commit and release lock {
  style.bold: true
}
a {
  "400"
}
t -> a: "Verify the balance and fail (400 < 1.000)" {
  style.bold: true
}
```

While this approach ensures correctness, it can **hinder parallelism**,
especially when long-running transactions lock many rows and reduce system throughput.

#### Optimistic Strategy

In contrast, the **Optimistic Strategy** allows transactions to run concurrently while guaranteeing
serializable behavior by **detecting and resolving conflicts dynamically**.

##### Predicate Locking

**Predicate Locking** is a mechanism used to detect conflicts.
Rather than operating only on individual tables or rows, it tracks logical predicates in queries.
If a conflict is detected, one transaction is allowed to proceed while the others are aborted.

Returning to the loan application example,
suppose two transactions read the same account, such as `A`.
The predicate (`AccountId = A`) is temporarily recorded in memory.

```d2
shape: sequence_diagram
t: Application Transaction
a: Account A (1100)
t1: "Withdrawal Transaction (700)"
"Predicate AccountId = A" {
   t -> a: "Verify the balance (1100 > 1000)"
   t1 -> a: "Verify the balance (1100 > 700)"
}
```

Next, the `Withdrawal Transaction` attempts to update the balance.
The database detects that an existing predicate (`AccountId = A`) conflicts with the update,
so it aborts **the least costly transaction**, meaning the transaction that has performed less work,
such as the `Application Transaction`.

```d2
shape: sequence_diagram
t: Application Transaction
a: "Account A (2.000)"
l: Applicant
t1: "Withdrawal Transaction (700)"
"Predicate AccountId = A" {
   t -> a: "Verify the balance (2.000 > 1.000)"
   t1 -> a: "Verify the balance (1.000 > 700)"
}
"Conflict AccountId = A" {
   t1 -> a: "Update the balance to 300"
   t -> t: Aborted (because of conflicting predicate) {
      class: error-conn
   }
   t1 -> a: Commit
}
```

What happens if the application transaction finishes first and commits?
If it releases the predicate lock immediately,
nothing prevents the withdrawal transaction from proceeding, and the anomaly can occur again.

```d2
shape: sequence_diagram
t: Application Transaction
a: Account A (2.000)
l: Applicant
t1: "Withdrawal Transaction (700)"
"Predicate AccountId = A" {
   t -> a: "Verify the balance (2.000 > 1.000)"
   t1 -> a: "Verify the balance (1.000 > 700)"
}
t -> l: Create a new record
t -> t: Commit and release {
  style.bold: true
}
t1 -> a: "Update the balance to 300" {
  class: error-conn
}
```

Therefore, the predicate lock is actually released only after **all overlapping** transactions have completed.

```d2
shape: sequence_diagram
t: Loan Transaction
a: Account A (2.000)
l: Loan
t1: "Withdrawal Transaction (700)"
"Predicate AccountId = A" {
   t -> a: "Verify the balance (2.000 > 1.000)"
   t1 -> a: "Verify the balance (1.000 > 700)"
}
t -> l: Create a new loan record
t -> t: Commit (the predicate lock is not released here) {
  style.bold: true
}
"Conflict AccountId = A" {
   t1 -> a: "Update the balance to 300"
   t1 -> t1: Aborted (because of conflicting predicate) {
      class: error-conn
   }
}
```

**Predicate Locking** also supports range queries such as `WHERE value > ...` and `WHERE value < ...`.
However, **Predicate Locking** introduces runtime overhead.
Operations require tracking predicates and checking for conflicts,
which can introduce significant runtime overhead.

This **Optimistic Strategy** minimizes blocking and enables greater concurrency,
which can make it suitable for complex transactions.
However, frequent transaction aborts and the cost of re-execution can offset these benefits.

## Setting Isolation Level

Ultimately, developers must choose an appropriate isolation level.
Higher isolation levels prevent more concurrency issues, but they can also significantly affect performance.

First and foremost, developers should:

- Identify potential anomalies for each transaction.
- Choose the **lowest possible isolation level** that safely prevents them.

There is no guarantee that an initially selected isolation level will be ideal.
Testing under a variety of concurrent transaction scenarios is essential for
uncovering subtle issues and refining isolation choices.
