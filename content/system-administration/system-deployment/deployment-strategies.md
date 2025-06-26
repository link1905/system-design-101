---
title: Deployment Strategies
weight: 20
next: system-monitoring
---

We have previously explored various deployment options.
In this section, we will delve into the specifics of deploying a service,
with a particular focus on how to update an existing one.

Imagine we are running a service with several instances:

```d2
s1: Service (v0.1) {
  grid-rows: 1
  i1: Instance 1 {
    class: server
  }
  i2: Instance 2 {
    class: server
  }
}
```

There are multiple deployment strategies, each suited to different use cases:

## Recreate Deployment

The concept is extremely straightforward:
the old version of the service is completely shut down before the new version is deployed.

```d2
grid-rows: 1
s1: Service (v0.1) {
    i1: Instance 1 {
        class: server
    }
    i2: Instance 2 {
        class: server
    }
}
s: Downtime without any instances {
  i: {
    class: none
  }
}
s2: Service (v0.2) {
  i1: Instance 1 {
    class: server
  }
  i2: Instance 2 {
    class: server
  }
}
s1 -> s: Shut down
s -> s2: Deploy
```

Running different versions simultaneously can potentially lead to conflicts and inconsistencies due to code changes.
The **Recreate** strategy ensures that only one version is operating at any given time,
making this approach suitable for applications that prioritize **consistency** above all else.

The most significant drawback is the lack of a smooth transition between versions,
which results in a **long downtime**.
During the deployment,
users cannot access the system until the new version is fully set up.
This is problematic for services requiring **high availability**.

## Blue/Green Deployment

**Blue/Green** deployment is similar to the **Recreate** strategy in that only one version is live at a time,
but it significantly reduces downtime by **fully preparing** the new version before making it active.

This strategy involves managing two identical production environments,
traditionally named **Blue** and **Green**.
At any moment, only one environment is **active** and exposed to client traffic.

Initially, let's say the Blue environment is active:

```d2
s: Service {
    grid-rows: 1
    s2: Green Environment (Inactive) {
      style.fill: ${colors.i1}
    }
    s1: Blue Environment (Active) {
      i1: Instance 1 (v0.1) {
          class: server
      }
      i2: Instance 2 (v0.1) {
          class: server
      }
    }
}
c: Client {
    class: client
}
c -> s.s1
```

When a new version is ready for release,
it is deployed to the inactive environment (Green).

```d2
s: Service {
    grid-rows: 1
    s2: Green Environment (Inactive) {
        style.fill: ${colors.i1}
        i1: Instance 1 (v0.2) {
            class: server
        }
        i2: Instance 2 (v0.2) {
            class: server
        }
    }
    s1: Blue Environment (Active) {
        i1: Instance 1 (v0.1) {
            class: server
        }
        i2: Instance 2 (v0.1) {
            class: server
        }
    }
}
c: Client {
    class: client
}
c -> s.s1
```

Once the new version in the Green environment is thoroughly tested and deemed stable,
traffic is routed it.
The Green environment becomes active, and the Blue environment becomes inactive.

```d2
c: Client {
    class: client
}
s: Service {
    grid-rows: 1
    horizontal-gap: 120
    s2: Green Environment (Active) {
        style.fill: ${colors.i1}
        i1: Instance 1 (v0.2) {
            class: server
        }
        i2: Instance 2 (v0.2) {
            class: server
        }
    }
    s1: Blue Environment (Inactive) {
        i1: Instance 1 (v0.1) {
            class: server
        }
        i2: Instance 2 (v0.1) {
            class: server
        }
    }
    s1 -> s2: Transition {
      style.bold: true
    }
}
c -> s.s2
```

If any unexpected issues arise with the new version,
traffic can be promptly redirected back to the old version in the Blue environment,
minimizing downtime and impact.

```d2
direction: right
s: Service {
    grid-rows: 1
    horizontal-gap: 120
    s2: Green Environment (Inactive) {
        style.fill: ${colors.i1}
        i1: Instance 1 (v0.2) {
            class: server
        }
        i2: Instance 2 (v0.2) {
            class: server
        }
    }
    s1: Blue Environment (Active) {
        i1: Instance 1 (v0.1) {
            class: server
        }
        i2: Instance 2 (v0.1) {
            class: server
        }
    }
    s1 <- s2: Rollback {
      style.bold: true
    }
}
c: Client {
  class: client
}
c -> s.s1
```

This strategy reduces conflicts by operating only one active version at a time.
However, unlike **Recreate**, users continue to access the old version during the deployment of the new one,
ensuring minimized downtime.
Furthermore, it allows for rapid rollback if errors are encountered.

The main disadvantage is cost.
Since two full production environments need to be maintained concurrently,
it can lead to double the infrastructure expenses.
In practice, **Blue/Green** is generally recommended over **Recreate** if the associated costs are manageable.

## Rolling Update

Instead of shutting down everything simultaneously,
**Rolling Update** is a strategy that **gradually replaces** instances of the old version with instances of the new version.

```d2
grid-columns: 1
s1: "Service (v0.1)" {
  grid-rows: 1
  i1: "Instance 1 (v0.1)" {
    class: server
  }
  i2: "Instance 2 (v0.1)" {
    class: server
  }
}
s2: "Service (Rolling)" {
  grid-rows: 1
  i1: "Instance 1 (v0.1)" {
    class: generic-error
  }
  i1n: "Instance 1 (v0.2)" {
    class: server
  }
  i2: "Instance 2 (v0.1)" {
    class: server
  }
  i1 -> i1n
}
s3: "Service (Rolling)" {
  grid-rows: 1
  i1: "Instance 1 (v0.2)" {
    class: server
  }
  i2: "Instance 2 (v0.1)" {
    class: generic-error
  }
  i2n: "Instance 2 (v0.2)" {
    class: server
  }
  i2 -> i2n
}
s4: "Service (v0.2)" {
  grid-rows: 1
  i1: "Instance 1 (v0.2)" {
    class: server
  }
  i2: "Instance 2 (v0.2)" {
    class: server
  }
}
s1 -> s2 -> s3 -> s4
```

This strategy ensures a **fluid transition** between versions because it always retains some instances to serve traffic.
However, it's generally cheaper than **Blue/Green** because the total number of concurrently available instances is minimized.

There are two primary problems with this approach:

- **Rollback Complexity**: If the new version has problems,
rolling back to the previous version is not immediate.
It typically involves another rolling update process to revert the changes.

- **Inconsistency**: The most significant issue is potential **inconsistency**.
During the deployment process,
some users might be routed to instances running the old version,
while others are routed to instances running the new version.
This can lead to inconsistent user experiences if there are breaking changes or significant differences between versions.

## Canary Deployment

**Canary** deployment is a strategy offering more fine-grained control over the deployment process.
The new version is **deployed experimentally** to a small subset of users,
known as the **Canary Group**.

For example,
we route `10%` of traffic (canary group) to the new version.

```d2
direction: right
s {
  class: none
  grid-rows: 1
  s1: Service (v0.1) {
    i1: "Instance 1 (v0.1)" {
        class: server
    }
    i2: "Instance 2 (v0.1)" {
        class: server
    }
  }
  s2: Service (v0.2) {
    i1: "Instance 1 (v0.2)" {
        class: server
    }
  }
}
c {
  class: none
  grid-rows: 1
  c: Client {
      class: client
  }
  ca: Canary group {
      class: client
  }
}
c.c -> s.s1: 90% traffic
c.ca -> s.s2: 10% traffic
```

The canary group is closely monitored for performance metrics, errors, and user feedback.
If everything works well and the new version is stable,
the canary group is gradually expanded,
and an increasing percentage of traffic is routed to the new version.

```d2
direction: right
s {
  class: none
  grid-rows: 1
  s1: Service (v0.1) {
    i1: "Instance 1 (v0.1)" {
        class: server
    }
    i2: "Instance 2 (v0.1)" {
        class: server
    }
  }
  s2: Service (v0.2) {
    i1: "Instance 1 (v0.2)" {
        class: server
    }
    i2: "Instance 2 (v0.2)" {
        class: server
    }
  }
}
c {
  class: none
  grid-rows: 1
  c: Client {
      class: client
  }
  ca: Canary group {
      class: client
  }
}
c.c -> s.s1: 50% traffic
c.ca -> s.s2: 50% traffic
```

This strategy is particularly valuable when releasing new features incrementally or when there's a higher risk associated with the new version.
It allows teams to collect early user feedback and detect unexpected issues with minimal impact.

Additionally, if the new version proves problematic,
it's relatively easy and quick to revert by simply routing all traffic back to the old version.

However,
canary deployments come with the added complexity of managing traffic splitting and maintaining (at least temporarily) multiple versions.
