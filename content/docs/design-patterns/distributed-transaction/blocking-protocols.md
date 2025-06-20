---
title: Blocking Protocols
weight: 10
prev: distributed-transaction
---

First, we will explore protocols that ensure consistency by strictly locking data,
which necessitates deep interactions between participating components.

## Two-Phase Commit (2PC)

The **Two-Phase Commit (2PC)** protocol is a widely recognized solution for managing distributed transactions.
It enables a group of participants to collectively **commit** (make permanent) or **abort** (discard) a transaction in a coordinated manner.

As its name implies, **2PC** operates in two distinct phases.
The protocol requires a designated **coordinator**; other participating entities are referred to as **cohorts**.

When a transaction is initiated:

1. **Prepare Phase**: Initially, the coordinator instructs the cohorts to prepare for the transaction. Each cohort performs the necessary actions, such as verifying data, acquiring locks, but they **do not yet commit** these changes.

2. **Commit Phase**: The coordinator then makes a decision based on the responses received from all cohorts:

    - If **all** cohorts respond with `Yes` (indicating they are prepared),
    the coordinator instructs them to **commit** their dirty data, making the changes permanent.
    - If **any** cohort responds with `No` (or fails to respond, indicating it could not prepare),
    the coordinator instructs all cohorts to **abort** their dirty data, rolling back any provisional changes.

Let's illustrate this with an example of transferring money between different banks (from Bank `A` to Bank `B`):

{{% steps %}}

### Prepare

The coordinator sends a `Prepare` request to the account services of both Bank `A` and Bank `B`.

```d2
shape: sequence_diagram
c: Coordinator {
    class: process
}
aa: Server A {
    class: server
}
ab: Server B {
    class: server
}
"1. Prepare" {
    c -> aa: Prepare
    c -> ab: Prepare
}
```

These cohort services will verify account details, update balances provisionally,
**lock** the accounts' balances to prevent other operations from interfering during the transaction.
If both services can successfully prepare, they respond `Yes` back to the coordinator.

```d2
shape: sequence_diagram
c: Coordinator {
    class: process
}
aa: Server A {
    class: server
}
ab: Server B {
    class: server
}
"1. Prepare" {
    c -> aa: Prepare
    c -> ab: Prepare
    aa -> aa: Verify the account and lock the balance {
        style.bold: true
    }
    aa -> c: Yes
    ab -> ab: Verify the account and lock the balance {
        style.bold: true
    }
    ab -> c: Yes
}
```

### Commit

Observing that all cohorts are prepared for the transaction (having received `Yes` from all),
the coordinator sends them a `Commit` request. The cohorts then make their changes permanent.

```d2
shape: sequence_diagram
c: Coordinator {
    class: router
}
aa: Server A {
    class: server
}
ab: Server B {
    class: server
}
"1. Prepare" {
    c -> aa: Prepare
    c -> ab: Prepare
    aa -> aa: Verify the account and lock the balance
    aa -> c: Yes
    ab -> ab: Verify the account and lock the balance
    ab -> c: Yes
}
"2. Commit" {
    c -> aa: Commit {
        style.bold: true
    }
    c -> ab: Commit {
        style.bold: true
    }
}
```

{{% /steps %}}

Despite its straightforwardness, this process is susceptible to several evident problems.

### Coordinator Failure

Firstly, system failures are unavoidable.
How should the system handle a scenario where the coordinator fails after sending the `Prepare` requests
but before sending the final `Commit` or `Abort` decision?

```d2
shape: sequence_diagram
c: Coordinator {
    class: router
}
s1: Server 1 {
    class: server
}
s2: Server 2 {
    class: server
}
s3: Server 3 {
    class: server
}
"1. Prepare" {
    s1 -> c: "Yes"
    s2 -> c: "Yes"
    s3 -> c: "Yes"
}
c -> c: Crash {
    class: error-conn
}
```

When a cohort responds `Yes`, it transitions to a `Prepared` state,
**locking** some data and committing to finalize the transaction as per the coordinator's eventual instruction.
If the coordinator fails at this juncture, the participants enter an uncertain state and may **wait indefinitely**.
A participant cannot unilaterally decide whether to commit or abort because it lacks information about the status
of other participants and the coordinator's final decision.

```d2
shape: sequence_diagram
c: Coordinator {
    class: router
}
s1: Server 1 {
    class: server
}
s2: Server 2 {
    class: server
}
s3: Server 3 {
    class: server
}
"1. Prepare" {
    s1 -> c: "Yes"
    s2 -> c: "Yes"
    s3 -> c: "Yes"
}
"2. Commit" {
    c -> c: Crash {
        class: error-conn
    }
    s1 <-> s1: Blocked {
        class: error-conn
    }
    s2 <-> s2: Blocked {
        class: error-conn
    }
    s3 <-> s3: Blocked {
        class: error-conn
    }
}
```

Therefore, **2PC** is a **blocking protocol**.
The failure of a single node (the coordinator) can block the entire transaction **indefinitely**.
Cohorts remain in an uncertain state, holding locks, until the coordinator recovers.
This issue significantly degrades system availability.

### Cohort Cooperation

It's evident that the blocking problem arises largely because the coordinator is a {{< term spof >}}.
What if cohorts could interact with each other to resolve uncertainty?

One idea is that after a **timeout** period (waiting for the coordinator),
cohorts could communicate among themselves.
They might decide to commit if they achieve unanimity (all prepared cohorts agree to commit).

```d2
shape: sequence_diagram
c: Coordinator {
    class: router
}
s1: Server 1 {
    class: server
}
s2: Server 2 {
    class: server
}
s3: Server 3 {
    class: server
}
"1. Prepare" {
    s1 -> c: "Yes"
    s2 -> c: "Yes"
    s3 -> c: "Yes"
}
"2. Commit" {
    c -> c: Crash {
        class: error-conn
    }
    s3 <-> c: Timeout {
       class: error-conn
    }
    s1 <-> s3: Unanimous Yes
    s1 -> s1: Commit
    s2 -> s2: Commit
    s3 -> s3: Commit
}
```

Unfortunately, this doesn't fully resolve the availability issue.
If any cohort goes down along with the coordinator,
the remaining active cohorts might not achieve unanimity (as they can't confirm the state of the crashed cohort)
and would still be stuck in an uncertain state.

```d2
shape: sequence_diagram
c: Coordinator {
    class: router
}
s1: Server 1 {
    class: server
}
s2: Server 2 {
    class: server
}
s3: Server 3 {
    class: server
}
"1. Prepare" {
    s1 -> c: "Yes"
    s2 -> c: "Yes"
    s3 -> c: "Yes"
}
"2. Commit" {
    c -> c: Crash {
        class: error-conn
    }
    s1 -> s1: Crash {
        class: error-conn
    }
    s2 <-> s3: No unanimity because of no information from Server 1 {
        class: error-conn
    }
}
```

Moreover, in many implementations, the coordinator's logic is often hosted on one of the cohort machines.
This means its failure is equivalent to a coordinator and cohort failure, halting the entire system.

## Three-phase Commit (3PC)

**Three-Phase Commit (3PC)** is a variation of **2PC** designed to address some of its blocking issues.
In essence, **3PC** introduces an additional phase between the **Prepare** and **Commit** phases.
This extra step aims to ensure that all cohorts are aware of the consensus outcome of the transaction **before** they proceed to actually commit the data.

Now, a transaction unfolds in **three** phases:

{{% steps %}}

### Prepare Phase

Similar to 2PC, the coordinator asks cohorts if they are willing and able to accept the transaction. Cohorts respond `Yes` or `No`.

```d2
shape: sequence_diagram
c: Coordinator {
    class: router
}
s1: Server 1 {
    class: server
}
s2: Server 2 {
    class: server
}
s3: Server 3 {
    class: server
}
"1. Prepare" {
    c -> s1: "Prepare"
    c -> s2: "Prepare"
    c -> s3: "Prepare"
    s1 -> c: "Yes"
    s2 -> c: "Yes"
    s3 -> c: "Yes"
}
```

### PreCommit Phase

Based on the responses:

- If all cohorts voted `Yes`, the coordinator sends a `PreCommit` message to all cohorts.
- If any cohort voted `No` (or failed to respond), the coordinator sends an `Abort` message.

Cohorts receiving a `PreCommit` or `Abort` message acknowledge it by responding with an `ACK` (Acknowledgement) to the coordinator.

```d2
shape: sequence_diagram
c: Coordinator {
    class: router
}
s1: Server 1 {
    class: server
}
s2: Server 2 {
    class: server
}
s3: Server 3 {
    class: server
}
"1. Prepare" {
    c -> s1: "Prepare"
    c -> s2: "Prepare"
    c -> s3: "Prepare"
    s1 -> c: "Yes"
    s2 -> c: "Yes"
    s3 -> c: "Yes"
}
"2. PreCommit" {
    c -> s1: "PreCommit"
    c -> s2: "PreCommit"
    c -> s3: "PreCommit"
    s1 -> c: ACK
    s2 -> c: ACK
    s3 -> c: ACK
}
```

### Commit Phase

After receiving ACKs from **all** cohorts for the `PreCommit` message,
the coordinator sends a final `Commit` request to all cohorts, instructing them to make their changes permanent.

```d2
shape: sequence_diagram
c: Coordinator {
    class: router
}
s1: Server 1 {
    class: server
}
s2: Server 2 {
    class: server
}
s3: Server 3 {
    class: server
}
"1. Prepare" {
    c -> s1: "Prepare"
    c -> s2: "Prepare"
    c -> s3: "Prepare"
    s1 -> c: "Yes"
    s2 -> c: "Yes"
    s3 -> c: "Yes"
}
"2. PreCommit" {
    c -> s1: "PreCommit"
    c -> s2: "PreCommit"
    c -> s3: "PreCommit"
    s1 -> c: ACK
    s2 -> c: ACK
    s3 -> c: ACK
}
"3. Commit" {
    c -> s1: Commit
    c -> s2: Commit
    c -> s3: Commit
}
```

{{% /steps %}}

Instead of blocking indefinitely,
**3PC** incorporates a **timeout** mechanism.
If the coordinator becomes unresponsive before the final `Commit`,
cohorts can communicate mutually to reach a consensus:

- If **at least one** cohort has received the `PreCommit` request from the coordinator:
This implies that all cohorts must have voted `Yes` in the prepare phase. Therefore, the cohorts can safely decide to commit the transaction.
- If **no** cohort has received a `PreCommit` request:
This indicates that the transaction was likely aborted by the coordinator. Thus, the transaction is aborted.

Let's consider coordinator crashes during 3PC:

### Case 1: Coordinator crashes after sending at least one `PreCommit` message

```d2
shape: sequence_diagram
c: Coordinator {
    class: router
}
s1: Server 1 {
    class: server
}
s2: Server 2 {
    class: server
}
s3: Server 3 {
    class: server
}
"1. Prepare" {
    s1 -> c: "Yes"
    s2 -> c: "Yes"
    s3 -> c: "Yes"
}
"2. PreCommit" {
    c -> s1: PreCommit
    c -> c: Crash {
        class: error-conn
    }
}
"3. Commit (Timeout & Recovery)" {
    s1 <-> s3: 'Decide to commit because Server 1 received "PreCommit"' {
        style.bold: true
    }
    s1 -> s1: Commit
    s2 -> s2: Commit
    s3 -> s3: Commit
}
```

### Case 2: Coordinator crashes before sending any `PreCommit` message

```d2
shape: sequence_diagram
c: Coordinator {
    class: router
}
s1: Server 1 {
    class: server
}
s2: Server 2 {
    class: server
}
s3: Server 3 {
    class: server
}
"1. Prepare" {
    s1 -> c: "Yes"
    s2 -> c: "Yes"
    s3 -> c: "Yes"
}
"2. PreCommit" {
    c -> c: Crash {
        class: error-conn
    }
}
"3. Commit (Timeout & Recovery)" {
    s1 <-> s3: 'Decide to abort because no cohort received "PreCommit"' {
        style.bold: true
    }
    s1 -> s1: Abort
    s2 -> s2: Abort
    s3 -> s3: Abort
}
```

The **PreCommit phase** acts as a buffer, holding the final decision of the transaction.
Even if the coordinator fails after initiating the phase,
other cohorts can autonomously finalize the transaction based on whether any cohort reached the `PreCommit` state.

Unfortunately, **3PC** is not a perfect solution.
It does not guarantee consistency in the presence of [network partitions]({{< ref "peer-to-peer-architecture#network-partition" >}}).
Imagine a scenario where `Server 1` receives a `PreCommit` request but then gets partitioned from other cohorts.
During the timeout and recovery phase:

- `Server 1` (isolated) will proceed to commit the transaction.
- `Server 2` and `Server 3` will detect no `PreCommit` among themselves and decide to abort the transaction.

This leads to an inconsistent state across the system.

```d2
shape: sequence_diagram
c: Coordinator {
    class: router
}
s1: Server 1 {
    class: server
}
s2: Server 2 {
    class: server
}
s3: Server 3 {
    class: server
}
"1. Prepare" {
    s1 -> c: "Yes"
    s2 -> c: "Yes"
    s3 -> c: "Yes"
}
"2. PreCommit" {
    c -> s1: "PreCommit"
}
c -> c: Crash {
    class: error-conn
}
"3. Commit (Timeout in Partition 1)" {
    s1 <-> s1: 'Commit (global decision was PreCommit)' {
        style.bold: true
    }
}
"3. Commit (Timeout in Partition 2)" {
    s2 <-> s3: 'Abort (no PreCommit detected among themselves)' {
        style.bold: true
    }
}
```

The choice between **2PC** and **3PC** often involves a trade-off between **Availability** and **Consistency**:

- **2PC** favors consistency over availability.
If coordinator recovery is fast and the system can tolerate the potential downtime caused by blocking,
2PC offers a simpler solution.
- **3PC** aims to improve availability by being non-blocking in more failure scenarios.
However, resolving inconsistencies that can arise from network partitions in 3PC can be exceptionally intricate. Consequently, it is less commonly used in practice than 2PC.

## Use Cases

The primary advantage of **Phase Committing** protocols is their ability to achieve **strong consistency**.
Changes can be applied across participating services in a coordinated, seemingly simultaneous manner,
which helps prevent inconsistent states from being left in the system.

**Phase Committing** is typically applied in contexts where a service needs to update data across multiple data sources immediately, for example:

- Between different shards of a single logical database.

```d2
s: Service {
    class: server
}
d: Database {
    s1: Shard 1 {
        user: |||yaml
        UserId = 3, Name = John
        |||
    }
    s2: Shard 2 {
        user: |||yaml
        UserId = 7, Name = Doe
        |||
    }
}
s -> d.s1: Update user 3
s -> d.s2: Update user 7
```

- Between different types of data stores (e.g., a database and a message broker). In the context of an [Event Streaming Platform]({{< ref "event-streaming-platform" >}}), achieving **exactly-once delivery** semantics often requires additional mechanisms. **2PC** can be an effective way to ensure that an operation (like updating a database record) and publishing an associated event occur atomically:

```d2
s: Service {
    class: server
}
d: Database {
    class: db
}
m: Event Partition {
    class: mq
}
s -> d: Update a record
s -> m: Create the associated event
```

Both **2PC** and **3PC** are considered low-level distributed algorithms.
Requiring microservices to expose interfaces like `PrepareTransaction`, `CommitTransaction`, etc.,
for inter-service transactions can create **high coupling** between services.
Therefore, they are less frequently used for orchestrating transactions between distinct business services in a
{{< term ms >}} environment,
where high-level patterns like **Saga** are often preferred.
