---
title: System Deployment
weight: 20
---

This section will cover some fundamental aspects of system deployment.

## Deployment Options

Suppose we aim to deploy a system on a server.
Let's examine the most common options for achieving this:

## Bare-metal Deployment

This is the most fundamental deployment method,
where system components run as native processes directly on the server's **Operating System (OS)**.
This model typically yields the **maximum performance** because
processes can access hardware resources with minimal overhead,
as there are no virtualization layers.

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

However, in this model, all processes share the same host OS.
This shared environment can be vulnerable to exploitation.

For instance, if an attacked process has incorrectly configured permissions for its executor,
an attacker could potentially abuse the OS to compromise other processes or even the entire server.

## Virtual Machine

**Virtual Machines (VMs)** are employed to create a more strictly isolated environment.
A virtual machine emulates a physical computer but runs on a software layer called a **Hypervisor**,
rather than directly on physical hardware.
Although VMs rely on the host's OS (via the hypervisor),
they each possess their own independent OS and dedicated resources,
ensuring they are **completely isolated** from one another.

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


However, virtual machines are not optimal in terms of resource efficiency.
Each VM runs its own **full operating system**, which consumes considerable CPU, memory, and storage resources.
Moreover, booting a virtual machine involves numerous setup steps and can take several minutes to complete.
This can prolong the initialization time for system components, potentially affecting [availability]({{< ref "service-cluster#availability" >}}).

In the next topic, we will explore a more lightweight option known as **Containerization**.

### Multi-tenant

{{< callout type="info" >}}
While not directly central to this topic, it's useful to explain the term **Multi-tenancy**, which is commonly used in cloud computing platforms.
{{< /callout >}}

In practice, organizations often rely on cloud providers for their system infrastructure rather than building it themselves.
When a customer rents an entire physical machine and has full control over it,
this is known as **Single-tenancy** or a **Dedicated Server**,
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

By default, many cloud platforms offer a **Multi-tenant** model.
In this setup, cloud providers divide a single physical machine into multiple virtual machines,
and each VM can be rented by a different customer.

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

Ultimately, in a multi-tenant environment, the virtual machines still share the same underlying hardware.
This poses a potential risk of data leaks if there are flaws in the software isolation mechanisms.
Consequently, multi-tenancy is sometimes prohibited in systems with extremely strict security requirements.
