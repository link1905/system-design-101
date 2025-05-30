---
title: Deployment Strategies
weight: 20
---

We'll obstained knowledge of deployment options in the previous topics.
In this one,
we'll discuss how to deploy a service,
especially making changes to an existing one.

Suppose we're running a service with several instances.
There are many deployment strategies fit different use cases:

```d2
direction: right
s1: Service (v0.1) {
  i1: Instance 1 {
    class: server
  }
  i2: Instance 2 {
    class: server
  }
}
```

## Recreate

The concept is extremely straightforward.
We shut down the old version completely, and then deploy a new one.

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
s: Downtime without any instances
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

Different versions potentially lead to conflicts and inconsistency due to code changes.
But at a time, **Recreate** ensures that only one version is operating,
this approach is suitable for applications favoring **consistency**.

The biggest problem is no smooth transition between versions,
results in a long downtime.
During deployment, users cannot access the system until the new version is set up completely.
This is problematic for **highly available** services.

## Blue/Green

**Blue/Green** is similar to **Recreate**,
but it **fully prepares** the new version before actually deploying.

It swaps between two environments: **Blue** and **Green**.
At a time, there is only one **active** environment exposed to clients.

```d2
s: Service {
    grid-rows: 1
    s1: Blue Environment (Active) {
        i1: Instance 1 (v0.1) {
            class: server
        }
        i2: Instance 2 (v0.1) {
            class: server
        }
    }
    s2: Green Environment (Inactive) {
        style.fill: ${colors.i1}
    }
}
c: Client {
    class: client
}
c -> s.s1
```

When a new version is released,
it will be set up in the inactive environment.

```d2
s: Service {
    grid-rows: 1
    s1: Blue Environment (Active) {
        i1: Instance 1 (v0.1) {
            class: server
        }
        i2: Instance 2 (v0.1) {
            class: server
        }
    }
    s2: Green Environment (Inactive) {
        style.fill: ${colors.i1}
        i1: Instance 1 (v0.2) {
            class: server
        }
        i2: Instance 2 (v0.2) {
            class: server
        }
    }
}
c: Client {
    class: client
}
c -> s.s1
```

Once the new version is tested and stable,
the environment becomes active and traffic will be routed to it,
meanwhile the other becomes inactive.

```d2
s: Service {
    grid-rows: 1
    s1: Blue Environment (Inactive) {
        i1: Instance 1 (v0.1) {
            class: server
        }
        i2: Instance 2 (v0.1) {
            class: server
        }
    }
    s2: Green Environment (Active) {
        style.fill: ${colors.i1}
        i1: Instance 1 (v0.2) {
            class: server
        }
        i2: Instance 2 (v0.2) {
            class: server
        }
    }
    s1 -> s2: Transition
}
c: Client {
    class: client
}
c -> s.s2
```

If unexpected issues appear,
we can promptly move back to the old version without downtime.

```d2
direction: right
s: Service {
    grid-rows: 1
    s1: Blue Environment (Active) {
        i1: Instance 1 (v0.1) {
            class: server
        }
        i2: Instance 2 (v0.1) {
            class: server
        }
    }
    s2: Green Environment (Inactive) {
        style.fill: ${colors.i1}
        i1: Instance 1 (v0.2) {
            class: server
        }
        i2: Instance 2 (v0.2) {
            class: server
        }
    }
    s2 -> s1: Rollback
}
c: Client {
    class: client
}
c -> s.s1
```

Similar to **Recreate**,
we reduce conflicts by only operating one active version.
However, during deployment, users still access to the old version,
helps to ensure that users experience with minimized downtime.
Moreover, we can rapidly transition back the system when encountering errors.

But this is such an expensive approach.
Since we need to maintain two concurrent environments during deployment,
leading to double the infrastructure.
In practice,
we are recommended to use **Blue/Green** rather than **Recreate** if the excessive cost is under control.

## Rolling Update

Instead of shutting everything down simultaneously,
**Rolling Update** is a strategy gradually **replacing** instances.
We take down of the old instances one by one and replace with a new version.

```d2
grid-columns: 4
s1: "Service (v0.1)" {
    i1: "Instance 1 (v0.1)" {
      class: server
    }
    i2: "Instance 2 (v0.1)" {
      class: server
    }
}
s2: "Service (Rolling)" {
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
  i1: "Instance 1 (v0.2)" {
    class: server
  }
  i2: "Instance 2 (v0.2)" {
    class: server
  }
}
s1 -> s2 -> s3 -> s4
```

Like **Blue/Green**, this strategy ensures a fluid transition between versions,
because it retains some instances (even the old version) to operate.
But **Rolling Update** is cheaper because the number of available instances is minimized.

There two problems of this approach:

- If the new verion has problems,
we can't roll back to the previous one immediately.

- The biggest problem is **inconsistency**.
During deployment, users accessing to different versions will experience inconsistently.

## Canary

**Canary** is like a combination of **Rolling Update** and **Blue/Green**.
Instead of automatic deployment, we control how the deployment happens in depth.

The new version is experimentally deployed to a small group of users,
this is called **Canary Group**.

```d2
direction: right
s {
  s1: Service (v0.1) {
    i1: "Instance 1 (v0.1)" {
        class: server
    }
    i2: "Instance 2 (v0.1)" {
        class: server
    }
  }
  s2: Service (v0.2) {
    i1: "Instance 2 (v0.2)" {
        class: server
    }
  }
}
c: Client {
    class: client
}
c -> s.s1: 90% traffic
c -> s.s2: 10% traffic
```

We need to monitor the canary group for performance, errors, and user feedback.
If everything works well, the canary group will gradually expand and receives more traffic.

```d2
direction: right
s {
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
c: Client {
    class: client
}
c -> s.s1: 50% traffic
c -> s.s2: 50% traffic
```

This strategy is valuable when releasing new features incrementally.
It allows us to collect early user feedback and detect unexpected issues.
Addtionally, if the new version is problematic,
it's easy to revert back to the old one.

However, it comes with the added management and cost of maintaining the canary group.
Users also experience inconsistenly between different versions.
