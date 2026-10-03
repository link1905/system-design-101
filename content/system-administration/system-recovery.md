---
title: System Recovery
weight: 40
prev: system-monitoring
next: design-patterns
---

Distributed systems inevitably experience failures such as node crashes or network partitions.
Robust recovery mechanisms are therefore crucial,
enabling systems to continue operating and recover effectively from unexpected problems.

## Backward Recovery

Consider a money transfer from account `A` to account `B`.
Unfortunately, the application crashes midway through the transaction:

```yaml
START TRANSACTION:
UPDATE 1: A = A - amount  # Executed
SYSTEM DOWN:
UPDATE 2: B = B + amount  # Not executed
```

When the application recovers, how do we correct this incomplete transaction?

- **Backward Recovery**: We undo the first update (`A = A - amount`) to **abort** the transaction and restore the system to its state before the transaction began.
- **Forward Recovery**: We attempt to execute the second step (`B = B + amount`) to **complete** the transaction.

**Forward Recovery** can introduce significant management and development overhead,
as each operation might require its own recovery strategy.

Consequently, **Backward Recovery** is more commonly applied due to its relative simplicity.
A system can recover from many types of faults by reverting to a previously known stable state.

## Write-Ahead Logging (WAL)

The **Write-Ahead Logging (WAL)** approach requires that **all** changes be recorded in a sequential log file **before**
they are applied to the actual data structures.

Let's revisit the money transfer.
Upon recovery, the system would examine the WAL and invalidate the first log entry.

```yaml
START TRANSACTION:
UPDATE 1:
  record: account A
  action: SET balance = balance + 100
SYSTEM DOWN:
```

There are two notable benefits of WAL:

1. **Reliability**: Since all updates are logged before they are applied to the main data,
this approach provides a high degree of reliability. If the system crashes, the WAL contains records of the intended changes.
2. **Point-in-Time Recovery (PITR)**: Based on the log file, the system can replay operations to reconstruct the state of the data up to any specific point in time covered by the logs.

However, WAL also introduces challenges:

1. **Storage Overhead**: A high volume of operations can lead to a very large WAL file,
as each operation typically generates at least one log entry.
This can result in high storage consumption.

2. **Recovery Time**: To recover data using a WAL file,
the system might need to replay a significant number of log entries,
which can be time-consuming.

## Checkpointing

### Durability

In reliable database systems,
a common practice is to write each change to the physical disk (persistent storage)
before acknowledging the operation to the client.

```d2
shape: sequence_diagram
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
h -> d: Successful
d -> c: Successful
```

This ensures maximum reliability and durability,
as the data on disk always reflects acknowledged changes.

However, hard disks are significantly slower than memory.
Writing to disk for every operation can be complex and slow,
especially if changes involve complex data serialization,
updates to multiple index structures, or data migration within files.
This can degrade performance, as write operations take longer to complete.

### Checkpoint

In this model, changes are first written to a fast in-memory buffer,
and the operation is quickly acknowledged to the client.
These changes are then written to the physical disk at **regular intervals** or based on certain triggers
(e.g., after a certain number of operations, when the buffer is full).
A captured state of the data flushed to disk is often referred to as a **checkpoint**.

```d2
shape: sequence_diagram
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
m -> d: Successful
m -> h: Flush snapshot {
  style.animated: true
}
```

This approach generally improves performance by reducing the number of slow disk writes.
Client requests can receive immediate responses after the data is written to memory,
while the actual persistence to disk happens asynchronously in the background.

However, the primary drawback is the increased risk of data loss.
If the system fails after an update has been written to memory but before it has been flushed to disk (i.e., before the next checkpoint),
that data will be lost.

This model is beneficial for use cases where some data loss is acceptable in exchange for higher performance,
such as in certain types of caching systems.

### Checkpointing and Logging

Writing changes to complex, structured data files on disk can be computationally expensive.
The WAL file is an **append-only sequential** file that is simple and inexpensive to write to,
making logging new entries very fast.

Many modern databases combine **Write-Ahead Logging** with **Checkpointing** to achieve both reliability and performance.
Changes are processed as follows:

1. Each change is first recorded as a log entry in the **WAL file** (ensuring durability of intent).
2. The change is then applied to an **in-memory version** of the data (allowing fast reads and writes).
3. Periodically, the in-memory data is **flushed to the main data files on disk** (checkpoint).

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
c -> s.m: Update
s.m -> s.wal: 1. Log the operation
s.m -> s.m: 2. Update in memory
```

As previously mentioned, the WAL file can grow very large and increase recovery time.
Checkpoints are crucial to truncate the WAL:
Old log entries preceding the last successful checkpoint can be removed or **archived**, significantly reducing the size of the active WAL file.

For example, suppose a WAL has five log entries:

```yaml
UPDATE 1
UPDATE 2
UPDATE 3
UPDATE 4
UPDATE 5
```

If a snapshot (checkpoint) is taken after `UPDATE 3` has been persistently stored in the main data files,
the system only needs to retain WAL entries from `UPDATE 4` onwards for future crash recovery.

```yaml
SNAPSHOT AT UPDATE 3
UPDATE 4
UPDATE 5
```

## Data Reconciliation

**Data Reconciliation** is the process of comparing two or more datasets to identify and resolve any discrepancies between them.

This process is frequently used in distributed database systems,
particularly to ensure data consistency between nodes,
such as a primary node and its replicas, or between different shards.

### Hash Tree

Comparing every piece of data in large datasets can be extremely inefficient.
A **hash tree**, also known as a **Merkle tree**,
is a data structure used to efficiently verify data consistency across different sources.

In a hash tree:

- Each leaf node typically represents a hash of an individual data block or record.
- Each non-leaf (parent) node is a hash of the concatenated hashes of its child nodes.
- This continues up to a single root hash, which represents the hash of the entire dataset.

A specific hash function (e.g., **SHA-256**) is used throughout the tree.
For example, a tree for a dataset `[1, 2, 3, 4]` might look like this (where `H()` is the hash function):

```d2
grid-columns: 1
l1 {
  class: none
  r1: H(H(H(1) + H(2)) + H(H(3) + H(4)))
}
l2 {
  class: none
  grid-rows: 1
  r1: H(H(1) + H(2))
  r2: H(H(3) + H(4))
}
l3 {
  class: none
  grid-rows: 1
  r1: H(1)
  r2: H(2)
  r3: H(3)
  r4: H(4)
}
l3.r1 -> l2.r1
l3.r2 -> l2.r1
l3.r3 -> l2.r2
l3.r4 -> l2.r2
l2.r1 -> l1.r1
l2.r2 -> l1.r1
```

Due to the nature of cryptographic hash functions,
if even a single bit of data represented by a leaf node differs between two trees,
their respective parent hashes will differ, and this difference will propagate all the way up to their root hashes.

Suppose we need to compare the set `[1, 2, 3, 4]` with another set `[1, 2, 3]` (missing `4`).
We then only need to correct the missing node and update the hashes along its path to the root.

```d2
grid-columns: 2
t1: Tree 1 {
  grid-columns: 1
  l1 {
    class: none
    grid-rows: 1
    r1: H(H(H(1) + H(2)) + H(H(3) + H(4))) {
      style.fill: ${colors.e}
    }
  }
  l2 {
    class: none
    grid-rows: 1
    r1: H(H(1) + H(2))
    r2: H(H(3) + H(4)) {
      style.fill: ${colors.e}
    }
  }
  l3 {
    class: none
    grid-rows: 1
    r1: H(1)
    r2: H(2)
    r3: H(3)
    r4: H(4) {
      style.fill: ${colors.e}
    }
  }
  l3.r1 -> l2.r1
  l3.r2 -> l2.r1
  l3.r3 -> l2.r2
  l3.r4 -> l2.r2
  l2.r1 -> l1.r1
  l2.r2 -> l1.r1
}
t2: Tree 2 {
  grid-columns: 1
  l1 {
    class: none
    grid-rows: 1
    r1: H(H(H(1) + H(2)) + H(H(3))) {
      style.fill: ${colors.e}
    }
  }
  l2 {
    class: none
    grid-rows: 1
    r1: H(H(1) + H(2))
    r2: H(H(3)) {
      style.fill: ${colors.e}
    }
  }
  l3 {
    class: none
    grid-rows: 1
    r1: H(1)
    r2: H(2)
    r3: H(3)
  }
  l3.r1 -> l2.r1
  l3.r2 -> l2.r1
  l3.r3 -> l2.r2
  l2.r1 -> l1.r1
  l2.r2 -> l1.r1
}
```

This structure allows for efficient detection of discrepancies,
often with a complexity related to **O(log N)** for finding the differing blocks,
rather than O(N) for a full comparison.
The trade-off is the overhead of constructing and maintaining this additional hash tree structure.
