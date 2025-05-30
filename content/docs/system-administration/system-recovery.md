---
title: System Recovery
weight: 40
---

Distributed systems face various failures (node crashes, network partitions, etc.),
and they're unavoidable.
A distributed system must equip recovery mechanisms to ensure that
it can continue operating or restart from a known good state.

## Backward Recovery

Let's take an example of transferring money from account `A` to `B`.
Unfortunately, the application is down in between,

```yaml
UPDATE 1: A = A - amount
SYSTEM DOWN: ...
UPDATE 2: B = B + amount
```

How do we correct the transaction when the application fully recovers?

- **Backward**: We roll back the first update to **abort** the transaction.
- **Forward**: We try to execute the second step to **complete** the transaction.

**Forward Recovery** adds much management and development overhead.
Each operation may require a different strategy.

Most of the time, **Backward Recovery** is applied because of the simplicity.
A system can recover from any type of fault by easily moving back to a previously stable state.

## Write-Ahead Logging (WAL)

**Logging** is a common stategy used in many database solutions.
This approach requires logging **all** operations performed within the application into
is a logging file called **Write-Ahead Logging (WAL)**.
Before performing any update, the system must log it to the file first.

Let's take a **WAL** example.
The system crashes between a transaction;
To recover, we need to undo the first update (`A += 100`).

```yaml
UPDATE 1:
  id: account A
  action: SET balance = balance + 100
SYSTEM DOWN:
UPDATE 2 (not executed):
  record: account B
  action: SET balance = balance - 100
```

There are two notable benefits:

1. We have logged all updates before they're actually performed,
   thus, this approach brings about reliability.

2. Based on the logging file,
   we can replay and rollback data to any stable state,
   this feature is called **Point-in-time Recovery (PITR)**.

And it also brings out two problems:

1. However, a lot of operations can happen,
   a log line is for one operation,
   then **WAL** leads to exceptionally **high strogage**.

2. To recover data from a **WAL** file,
   we need to relay data from all log lines,
   taking extremely long to complete.

## Checkpointing

### Write-through

In reliable databases, when receiving any change,
they will save it to the physical disk immediately.
This behaviour is called **Write-through**.

```d2
direction: right
c: Client {
  class: client
}
d: Database {
  class: db
}
h: Hard Disk {
  class: hd
}
c -> d: Update
d -> h: Save
```

This ensures the reliability,
as physical data is always up-to-date.

However, hard disks are slow to work with.
Writing to disk is not often easy;
Changes may result in unexpedly complex and heavy staff,
such as serializing data formats, migrating data between files, etc.
Briefly, it leads to degraded performance as writes will take longer to complete.

### Write-back

This model constracts with **Write-through**,
where changes are first kept in memory and written to disk at **regular intervals** (e.g., after one minute, after 10 operations...).
A captured state is called a **checkpoint** or **snapshot**.

```d2
d: Database {
  class: db
}
m: Memory {
  class: cache
}
h: Hard Disk {
  class: hd
}
d -> m: Update
m -> h: Flush periodically {
  style.animated: true
}
```

Absolutely,
this comes with a better performance by reducing saving operations.
Changes can respond immediately after writing to the memory,
their durability will be handled in the background.

However, we easily observe the possibility of data loss,
data processed after the latest checkpoint will not be recovered in case of failure.
This model is benefical for use cases which data loss is acceptable, such as caching.

### Checkpointing And Logging

Many databases try to combine between **Logging** and **Checkpointing**.
Changes are wriiten to the logging file first and periodically flushed from the memory.

- Writing changes to structed data files is complex and require much computing effort.
  The **WAL** file is merely an linearly only-appended file,
  it's simple and lighweight for logging new lines.

- In case of failure,
  the **WAL** works as a reliable place for recovering data.

```d2
direction: right
s: Database {
  m: Memory {
      class: cache
  }
  wal: WAL {
      class: file
  }
  l: Hard Disk {
    class: hd
  }
  m -> l: Flush periodically {
    style.animated: true
  }
}
c: Client {
  class: client
}
c -> s.m: 1. Update in memory
s.m -> s.wal: 2. Log the operation
```

As we've said,
the **WAL** file can grow exceptionally and take a lot of time to recover.
Thanks to checkpoints,
we can remove the old log lines to significantly reduce the file's size.

For example, we have a **WAL** of 5 logs:

```d2
l: Log {
  c: |||yaml
  UPDATE 1
  UPDATE 2
  UPDATE 3
  UPDATE 4
  UPDATE 5
  |||
}
```

With a snapshot taken at the firth update.
We necessarily need to maintain log lines after that moment.
Morever, recovering data is more rapid
as we only need to replay from the snapshot (not the entire log).

```d2
l: Log {
  c: |||yaml
  UPDATE 4
  UPDATE 5
  |||
}
"SNAPSHOT at UPDATE 3"
```

## Data Reconciliation

Data reconciliation is the process of comparing and resolving discrepancies between two data sets.
This process is used a lot in distributed databases,
where we need to to compare between nodes, such as a primary and its replica,
to ensure data consistency in system.

### Hash Tree

However, comparing every piece of data in sets one by one is inefficient.
Thus, **Hash Tree** is a powerful data structure in this case.

Basically, in a **Hash Tree**, we use a certain hash function:

- Data is determined by its hash and lives in the leave nodes.
- A parent node is the hash from its children.

For example,
a tree of `[1, 2, 3, 4]` look like this:

```d2
grid-columns: 1
l1 {
  class: none
  r1: H(H(H(1) + H(2)) + H(H(3) + H(4)))
}
l2 {
  class: none
  r1: H(H(1) + H(2))
  r2: H(H(3) + H(4))
}
l3 {
  class: none
  r1: H(1)
  r2: H(2)
  r3: H(3)
  r4: H(4)
}
```

Based on the hashing instinct, two different nodes will result in different parents.
Traversing the trees, when detecting any two different nodes,
that means their leaves are different.

Suppose we need to compare with another set `[1, 2, 3]`.
From the root, we see the first right nodes are different,
then we dive into the branches to find the exact discrepancies.

```d2
grid-columns: 2
t1 {
  grid-columns: 1
  l1 {
    class: none
    r1: H(H(H(1) + H(2)) + H(H(3) + H(4)))
  }
  l2 {
    class: none
    r1: H(H(1) + H(2))
    r2: H(H(3) + H(4)) {
      style.fill: ${colors.e}
    }
  }
  l3 {
    class: none
    r1: H(1)
    r2: H(2)
    r3: H(3)
    r4: H(4)
  }
}
t2 {
  grid-columns: 1
  l1 {
    class: none
    r1: H(H(H(1) + H(2)) + H(H(3)))
  }
  l2 {
    class: none
    r1: H(H(1) + H(2))
    r2: H(H(3)) {
      style.fill: ${colors.e}
    }
  }
  l3 {
    class: none
    r1: H(1)
    r2: H(2)
    r3: H(3)
  }
}
```

This structure helps detect discrepancies efficiently with the **O(log n)** complexity,
but that requires maintaining an subsidiary structure.
