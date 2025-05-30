---
title: System Deployment
weight: 20
---

In this topic,
we'll some basic aspects of deploying a system.

## Deploy Options

Suppose we're trying deploy a system on a server.
Let's see the most common options to do that:

## Bare-metal Deployment

This is the most basic way,
we essentially run the components as native processes on the server's **Operating System (OS)**.
This model results in the **maximum performance**,
because processes can access hardware directly without any virtual overhead.

```d2
m: Server {
  grid-columns: 1
  p {
    class: none
    grid-rows: 1
    p1: Process 1 {
      class: process
    }
    p2: Process 2 {
      class: process
    }
  }
  h: Hardward {
    class: hd
  }
  o: OS {
    class: os
  }
  p.p1 -> h
  p.p2 -> h
  h -> r
}
```

However, in this model, processes share the same host's OS,
making it prone to be exploited.
E.g., from an attacked process,
if the permissions of its executor are not configured correctly,
the attacker can abuse the OS to harm other processes, or even the entire server.

## Virtual Machine

Then, **Virtual Machine** is used to build a stricter environment.

A virtual machine is similar to a physical one but run on a software called **Hypervisor** (instead of real hardward).
Despite depending the host's OS,
virtual machines possess their own OS and resources,
they are **completely isolated** with each other.

```d2
m: Server {
  grid-columns: 1
  p {
    class: none
    grid-rows: 1
    v1: Virtual Machine 1 {
      r: Resources {
        class: hd
      }
      o: OS {
        class: os
      }
    }
    v2: Virtual Machine 2 {
      r: Resources {
        class: hd
      }
      o: OS {
        class: os
      }
    }
  }
  h: Hypervisor {
    class: process
  }
  o: OS {
    class: os
  }
  p1.v1 -> h
  p2.v2 -> h
  h -> o
}
```

Then, we can deploy the system's processes on different virtual machines.
This comes with a big security benefit,
even some processes are attacked,
they can't be used to affect on others and the physical machine.

```d2
m: Server {
  grid-columns: 1
  p {
    class: none
    grid-rows: 1
    v1: Virtual Machine 1 {
      p1: Process 1 {
        class: process
      }
    }
    v2: Virtual Machine 2 {
      p2: Process 2 {
        class: process
      }
    }
  }
  h: Hypervisor {
    class: process
  }
  o: OS {
    class: os
  }
  p1.v1 -> h
  p2.v2 -> h
  h -> o
}
```


However, virtual machine is not good in terms of resource efficiency.
Each virtual machine runs its own **full OS**, consuming significant CPU, memory, and storage.
Morever, booting a virtual machine requires many setups and takes minutes to complete,
making initializing system components longer (potentially affect [availability]({{< ref "service-cluster#availability" >}}).

In the next topic,
we'll learn about a more lightweight option called **Containerization**.

### Multi-tenant

Not really relevant to this topic,
but I want to explain the term **Multi-tenant** popularly used in cloud platforms.

In practice,
we rarely build but depend on a cloud provider for the system's infrastructure.

When we rent a whole physical machine and take full control of it,
it's called **Single-tenant** or **Dedicated Server**,
as no one else can touch your server.

```d2
s1: Dedicated Server 1 {
  class: server
}
s2: Dedicated Server 2 {
  class: server
}
u1: User 1 {
  class: client
}
u2: User 2 {
  class: client
}
u1 -> s1: Rent
u2 -> s2: Rent
```

By default, many cloud platforms provide the **Multi-tenant** model.
They divide a physical machine into multiple virtual machines,
each can be rented by a different customer.

```d2
s: Server {
  v1: Virtual Machine 1 {
    class: os
  }
  v2: Virtual Machine 2 {
    class: os
  }
}
u1: User 1 {
  class: client
}
u2: User 2 {
  class: client
}
u1 -> s.v1: Rent
u2 -> s.v2: Rent
```

At the end of the day,
the virtual machines still share the same hardware,
raising the risk of data leaks if there’s a flaw in software isolation.
**Multi-tenant** is ocasionally outright banned in some strictly security systems.
