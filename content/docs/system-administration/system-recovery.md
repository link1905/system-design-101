---
title: System Recovery
weight: 40
---

Distributed systems inevitably face failures like node crashes or network partitions.
Robust recovery mechanisms are therefore crucial,
enabling the system to continue operating and recover effectively from such unexpected problems.

## Backward Recovery

Consider an example of transferring money from account `A` to account `B`.
Unfortunately, the application crashes midway through the transaction:

```yaml
START TRANSACTION:
UPDATE 1: A = A - amount  # Executed
SYSTEM DOWN:
UPDATE 2: B = B + amount  # Not executed
```

When the application recovers, how do we correct this incomplete transaction?
- **Backward Recovery**: We undo the first update (`A = A - amount`) to **abort** the transaction and restore the system to its state before the transaction began.
-  **Forward Recovery**: We attempt to execute the second step (`B = B + amount`) to **complete** the transaction.

**Forward Recovery** can introduce significant management and development overhead,
as each operation might require a unique recovery strategy.

Consequently, **Backward Recovery** is more commonly applied due to its relative simplicity.
A system can recover from many types of faults by reverting to a previously known stable state.

## Write-Ahead Logging (WAL)

The **Write-Ahead Logging (WAL)** approach mandates that **all** operations or changes
intended for the data are first recorded in a sequential log file **before**
the changes are applied to the actual data structures.

Let's revisit the money transfer.
Upon recovery, the system would examine the WAL and invalidate the first log entry.

```yaml
START TRANSACTION:
UPDATE 1:
  recored: account A
  action: SET balance = balance + 100
SYSTEM DOWN:
UPDATE 2 (not executed):
  record: account B
  action: SET balance = balance - 100
```


There are two notable benefits of WAL:
1. **Reliability**:
Since all updates are logged before they are actually performed on the main data,
this approach provides a high degree of reliability. If the system crashes, the WAL contains a record of what was intended.
2. **Point-in-Time Recovery (PITR)**: Based on the log file, the system can replay operations to reconstruct the state of the data up to any specific point in time covered by the logs.

However, WAL also introduces challenges:
1. **Storage Overhead**: A high volume of operations can lead to a very large WAL file,
as each operation typically generates at least one log entry.
This can result in high storage consumption.
2. **Recovery Time**: To retrive data from a WAL file,
the system might need to replay a significant number of log entries,
which can be time-consuming.

## Checkpointing

### Durability

In reliable database systems,
a common practice is that when the system receives any change,
it writes that change to the physical disk (persistent storage) immediately before confirming the operation to the client.

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
h -> d: Successful
d -> c: Successful
```

This ensures maximum reliability and durability,
as the data on disk is always up-to-date with acknowledged changes.
If the system crashes, no acknowledged data is lost.

However, hard disks are significantly slower than memory.
Writing to disk for every operation can be complex and slow,
especially if changes involve intricate data serialization,
updates to multiple index structures, or data migration within files.
This can lead to degraded performance, as write operations will take longer to complete.

### Checkpoint

In this model, changes are first written to a fast in-memory buffer,
and the operation is confirmed to the client quickly.
These changes are then written to the physical disk at **regular intervals** or based on certain triggers
(e.g., after one minute, after a certain number of operations, when the buffer is full).
A captured state of the data flushed to disk is often referred to as a **checkpoint** or **snapshot**.

```d2
direction: right
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
m -> h: Flush periodically {
  style.animated: true
}
```

This approach generally offers better performance by reducing the number of slow disk-writing operations.
Client requests can receive immediate responses after the data is written to memory,
while the actual persistence to disk happens asynchronously in the background.

However, the primary drawback is the increased risk of data loss.
If the system fails after an update has been written to memory but before it has been flushed to disk (i.e., before the next checkpoint),
that data will be lost.
This model is beneficial for use cases where some data loss is acceptable in exchange for higher performance,
such as in certain types of caching systems.

### Checkpointing And Logging

Writing changes to complex structured data files on disk can be computationally expensive.
The WAL file, being an **append-only sequential** file, is simple and lightweight to write to,
making logging new entries very fast.
Many modern databases effectively combine **Write-Ahead Logging** with **Checkpointing** to get the benefits of both reliability and performance.
Changes are:
1.  First, written as a log entry to the **WAL file** (ensuring durability of intent).
2.  Then, applied to an **in-memory version** of the data (allowing fast reads and writes).
3.  Periodically, the in-memory data is **flushed to the main data files on disk** (this is the checkpoint).

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


As previously mentioned, the WAL file can grow exceptionally large and make recovery times long. Checkpoints are crucial to:
- **Truncate the WAL**: Old log lines preceding the last successful checkpoint can be removed or archived,
significantly reducing the WAL file's active size.
- **Speed up Recovery**: In case of a crash, the system can restore its state from the latest checkpoint on disk and then only needs to replay WAL entries *after* that checkpoint to recover subsequent changes.

For example, if a WAL has 5 log entries:

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

Data reconciliation is the process of comparing two or more datasets to identify and resolve any discrepancies between them.
This process is frequently used in distributed database systems,
particularly for ensuring data consistency between nodes,
such as a primary node and its replicas, or between different shards.

### Hash Tree

Comparing every piece of data in large datasets can be extremely inefficient.
**Hash Tree**, also known as **Merkle Tree**,
is a powerful data structure used to efficiently verify the consistency of data between different sources.

In a Hash Tree:
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
l3.r1 -> l2.r1
l3.r2 -> l2.r1
l3.r3 -> l2.r2
l3.r4 -> l2.r2
l2.r1 -> l1.r1
l2.r2 -> l1.r1
```

Due to the nature of cryptographic hash functions,
if even a single bit of data in any leaf node is different between two trees,
their respective parent hashes will differ, and this difference will propagate all the way up to their root hashes.

Suppose we need to compare the set `[1, 2, 3, 4]` with another set `[1, 2, 3]` (missing `4`).

```d2
grid-columns: 2
t1: Tree 1 {
  grid-columns: 1
  l1 {
    class: none
    r1: H(H(H(1) + H(2)) + H(H(3) + H(4))) {
      style.fill: ${colors.e}
    }
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
    r1: H(H(H(1) + H(2)) + H(H(3))) {
      style.fill: ${colors.e}
    }
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
The trade-off is the overhead of constructing and maintaining this subsidiary hash tree structure.
