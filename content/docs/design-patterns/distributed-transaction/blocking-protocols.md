---
title: Blocking Protocols
weight: 10
---

First, we find out about protocols that strictly lock data to ensure consistency,
requiring deep interactions between participants.

## Two-phase Commit - 2PC

**Two-phase Commit - 2PC** stands out as the most popular solution for distributed transactions.
It allows a set of participants
to either **commit** or **abort** a transaction in a coordinated way.

As the name suggests, **2PC** will happen in two phases.
**2PC** itself requires a **coordinator**, other participants are called **cohorts**.
When a transaction is initiated

1. **Prepare Phase**: First, the coordinator will ask cohorts to prepare for the transaction.
   They will perform necessary actions: verify, lock, update data... but **not commit yet** (temporary dirty data)

2. **Commit Phase**: The coordinator will decide based on responses from the cohorts

- If **all** respond **Yes**: The coordinator will request them to **commit** the dirty data.
- If **any** responds **No** (failed to prepare): The coordinator will request them to **abort** the dirty data.

Let's say we want to transfer an amount of money across different banks (`A` to `B`)

1. **Prepare**: The coordinator sends a `Prepare` request to account services of `A` and `B`.

```d2
shape: sequence_diagram
c: Coordinator {
    class: router
}
aa: Account Service (A) {
    class: server
}
ab: Account Service (B) {
    class: server
}
"1. Prepare" {
    c -> aa: Prepare
    c -> ab: Prepare
} 
```

2. These cohorts will verify, update,
   and **lock** the accounts' balance to prevent other interferences.
   If everything is smooth, they respond `Yes` back to the coordinator

```d2
shape: sequence_diagram
c: Coordinator {
    class: router
}
aa: Account Service (A) {
    class: server
}
ab: Account Service (B) {
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
```

3. **Commit**: The coordinator observes that all cohorts are prepared for the transaction,
   it starts to send them a `Commit` request

```d2
shape: sequence_diagram
c: Coordinator {
    class: router
}
aa: Account Service (A) {
    class: server
}
ab: Account Service (B) {
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
    c -> aa: Commit
    c -> ab: Commit
}
```

Due to straightforwardness,
we can instantly come up with some obvious problems happening in this process.

### Coordinator Failure

First, corruption is unavoidable.
How do we handle if the coordinator dies after sending any `Prepare`?

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

Responding `Yes` means that
cohorts are moving the `Prepare` state,
**locking** some pieces of data and willing to commit them.

The corruption of the coordinator causes participants to come an uncertain state and **wait indefinitely**, why?
A participant is confused to decide whether it should commit or abort data,
it has no information of the others.

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
    s1 <-> c: Blocked {
       class: i-conn
    }
    s2 <-> c: Blocked {
       class: i-conn
    }
    s3 <-> c: Blocked {
       class: i-conn
    }
}
```

Therefore, **2PC** is a **blocking protocol**, which blocks data to ensure consistency.
However, a single-node failure can block the entire process **indefinitely**.
The cohorts stand in the uncertain state until the coordinator is healthy again.
This problem significantly degrades the system performance and availability.

### Cohort Cooperation

We easily see that the blocking problem arises because the coordinator is the {{< term spof >}},
how about letting cohorts play with each other?
After **timing out**, cohorts will ask mutually and decide to commit if they get unanimity.

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

Unfortunately, we've not boosted up the availability yet.
If any sever goes down with the coordinator,
once again, the remaining cohorts get no unanimity and stay in the uncertain state.
Moreover, in fact, the coordinator is frequently implemented as a cohort also;
That means its corruption still halts the entire system.

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
    s3 -> s3: Crash {
        class: error-conn
    }
    s1 <-> s3: Unanimous Yes
    s1 <-> s3: No unanimity because of no information of Server 3 {
        class: error-conn
    }
    s1 -> s1: Abort
    s2 -> s2: Abort
    s3 -> s3: Abort
}
```

## Three-phases Commit - 3PC

**Three-phases Commit - 3PC** is a variation of **2PC**.
In short, **3PC** adds an extra phase between **Prepare** and **Commit**,
letting cohorts **know the final consensus** of the transaction before they actually commit data.

Now, a transaction happens within **three** phases:

1. **Prepare Phase** is similar to **2PC**, the coordinator ask cohorts whether they will accept the transaction.
2. **PreCommit Phase**: The coordinator will send a **PreCommit** (all `Yes`) or **Abort** (any `No`) signal.
The cohorts respond back an **ACK** (Acknowledgement) to the coordinator.
3. **Commit Phase**: After receiving **ACKs** from **all** cohorts, the coordinator finally requests them to commit.

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
    s1 -> c: ACK
    s1 -> c: ACK
}
"3. Commit" {
    c -> s1: Commit
    c -> s2: Commit
    c -> s3: Commit
}
```

Now, we don't block data indefinitely but adding a **timeout** mechanism.
If the **PreCommit** phase lasts too long (timeout),
cohorts will communicate mutually to form a **consensus**:

- **At least one** cohort has received the **PreCommit** request:
This means all the cohorts have accepted the transaction before (that's why we have the **PreCommit** request),
the cohorts will commit the transaction.
- No **PreCommit** is detected:
The transaction has been declined by some cohorts, so it's aborted.

We may imagine the **PreCommit** phase works as a shim maintaining the final result of the transaction.
Even if the coordinator has failed, other cohorts can decide to finalize the transaction autonomously.

Let's see what if the coordinator crashes in between `3PC`.
In this case, at least one **PreCommit** is fired:

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
"3. Commit (Timeout)" {
    s1 <-> s3: 'Decide to commit because of detecting a "PreCommit" in Server 1' 
    s1 -> s1: Commit
    s2 -> s2: Commit
    s3 -> s3: Commit
}
```

Or no **PreCommit** is detected:

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
"3. Commit (Timeout)" {
    s1 <-> s3: 'Decide to abort because of no "PreCommit"' 
    s1 -> s1: Abort
    s2 -> s2: Abort
    s3 -> s3: Abort
}
```

Unfortunately, this is not a panacea,
because **3PC** does not provide strong consistency in [network partitions](Peer-to-peer-Architecture.md#network-partition).
Imagine when the `Server 1` receives a **PreCommit** request but gets **partitioned** from the rest.
That leads to an inconsistency in the timeout phase:

- `Server 1` will commit the transaction.
- (`Server 2`, `Server 3`) will abort it because of no **PreCommit**.

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
    c -> c: Crash {
        class: error-conn
    }
}
"3. Commit (Network Partition 1)" {
    s1 <-> s1: 'Commit' 
}
"3. Commit (Network Partition 2)" {
    s2 <-> s3: 'Abort'
}
```

**2PC** and **3PC** is a choice between **Availability** and **Consistency**:

- **2PC** favors consistency. If we can recover the coordinator fast enough and
  the system is satisfied with the downtime, `2PC` is an easier solution
- Meanwhile, discarding inconsistencies and recovering data in **3PC** can be exceptionally intricate.
  Therefore, it is less commonly used than **2PC**.

The best advantage of **Phase Committing** is **strong consistency** and high performance.
Changes can be performed in participating services simultaneously,
helps prevent inconsistent states are left in the system.
**Phase Committing** is usually applied in the context that a service updates data on multiple data source, e.g.,

- Between different shards of a single system.

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

- Between different data stores.
In the [Event Streaming Platform](Event-Streaming-Platform.md) topic,
we've discussed that we must equip more additional tools to actually reach the **exactly-once delivery**,
**2PC** is an effective way to do that:

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

Both **2PC** and **3PC** are **low-level** algorithms.
A microservice needs to expose interfaces like `PrepareTransaction`, `CommitTransaction`... creates high couplings.
Thus, they're less used to perform transactions between services in a {{< term ms >}} environment.
