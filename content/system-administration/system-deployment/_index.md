---
title: System Deployment
weight: 20
prev: system-protection
next: containerization
---

This section covers some fundamental aspects of system deployment.

## Deployment Options

Suppose we want to deploy a system on a server.
Let's examine the most common deployment options:

### Bare-Metal Deployment

This is the most basic deployment method,
in which system components run as native processes directly on the server's **operating system (OS)**.

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
  o: "" {
    class: os
  }
  h: Hardware {
    class: hd
  }
  p.p1 -> o
  p.p2 -> o
  o -> h
}
```

This model typically delivers **maximum performance** because
processes can access hardware resources with minimal overhead,
as there are no virtualization layers.

However, all processes share the same host OS.
This shared environment can be vulnerable to exploitation.

For instance, if a process under attack runs under an account with misconfigured permissions,
it could exploit those permissions to compromise other processes or even the entire server.

```d2
m: Server {
  grid-columns: 1
  p {
    class: none
    grid-rows: 1
    horizontal-gap: 150
    p1: Process 1 (Attacked) {
      class: process
    }
    p2: Process 2 {
      class: process
    }
  }
  o: "" {
    class: os
  }
  p.p1 -> p.p2: Compromise {
    class: error-conn
  }
  p.p1 -> o: Compromise {
    class: error-conn
  }
}
```

### Virtual Machines

**Virtual machines (VMs)** are used to create more strongly isolated environments.
A virtual machine emulates a physical computer but runs on a software layer called a **hypervisor**,
rather than directly on physical hardware.

Although VMs rely on the host OS (via the hypervisor),
each has its own independent OS and dedicated resources (files, users, processes, etc.),
ensuring they are **completely isolated** from one another.

```d2
m: Server {
  grid-columns: 1
  p {
    class: none
    grid-rows: 1
    v1: Virtual Machine 1 {
      grid-rows: 1
      horizontal-gap: 100
      r: Resources {
        class: hd
      }
      o: OS {
        class: os
      }
      o -> r: Manage
    }
    v2: Virtual Machine 2 {
      grid-rows: 1
      horizontal-gap: 100
      o: OS {
        class: os
      }
      r: Resources {
        class: hd
      }
      o -> r: Manage
    }
  }
  h: Hypervisor {
    class: process
  }
  o: Host OS {
    class: os
  }
  p.v1.o -> h
  p.v2.o -> h
  h -> o
}
```

System processes can then be deployed on different virtual machines.
This approach offers a significant security benefit:
even if some processes on one VM are compromised,
they cannot directly affect processes on other VMs or the underlying physical machine.

```d2
m: Server {
  grid-columns: 1
  p {
    class: none
    grid-rows: 1
    v1: Virtual Machine 1 {
      grid-rows: 1
      horizontal-gap: 100
      p1: Process 1 {
        class: process
      }
      o: "" {
        class: os
      }
      p1 -> o
    }
    v2: Virtual Machine 2 {
      grid-rows: 1
      horizontal-gap: 100
      o: "" {
        class: os
      }
      p2: Process 2 {
        class: process
      }
      p2 -> o
    }
  }
  h: Hypervisor {
    class: process
  }
  o: OS {
    class: os
  }
  p.v1.o -> h
  p.v2.o -> h
  h -> o
}
```

However, virtual machines are not optimal for resource efficiency.
Each VM runs its own **full operating system**, which consumes considerable CPU, memory, and storage resources.
Moreover, booting a virtual machine involves numerous setup steps and can take several minutes to complete.
This can increase the initialization time of system components, potentially affecting [system availability]({{< ref "service-cluster#availability" >}}).

In the next topic, we will explore a lighter-weight option known as **containerization**.

#### Multi-Tenancy

In practice, organizations often rely on cloud providers for their system infrastructure rather than building their own.
When a customer rents an entire physical machine and has full control over it,
this arrangement is known as **single-tenancy**, and the machine is called a **dedicated server**,
as no other customer can access that server.

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

By default, many cloud platforms offer a **multi-tenant** model.
In this setup, cloud providers divide a single physical machine into multiple virtual machines,
and each VM can be rented by a different customer.

```d2
s: Physical Server {
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

Ultimately, in a multi-tenant environment, the virtual machines still share the same underlying hardware.
This poses a potential risk of data leaks if there are flaws in the software isolation mechanisms.
Consequently, multi-tenancy is sometimes prohibited in systems with extremely strict security requirements.
