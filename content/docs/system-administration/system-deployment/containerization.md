---
title: Containerization
weight: 10
prev: system-deployment
---

**Containerization** is a lightweight virtualization method that has gained significant momentum in recent years,
establishing itself as a cornerstone of modern software development and deployment practices.

## Kernel

To understand **Containerization** thoroughly,
let's first examine the foundational elements upon which it is built.

Users should generally not interact directly with hardware.
Direct hardware access can lead to numerous problems related to security,
resource management, and hardware compatibility.

The **Kernel** is the core component of an operating system.
It acts as a bridge between **processes** (running applications) and the physical **hardware**.
To interact with hardware, user commands are first passed to and processed by the kernel.

```d2
grid-columns: 1
o: OS {
  grid-columns: 1
  c {
    class: none
    grid-rows: 1
    c1: Process 1 {
      class: process
    }
    c2: Process 2 {
      class: process
    }
  }
  k: Kernel {
    class: gw
  }
}
h: Hardware {
  class: hd
}
o.c.c1 <-> o.k
o.c.c2 <-> o.k
o.k <-> h
```

### Kernel Isolation

Typically, processes run within a **shared operating system (OS)** environment without inherent isolation.
This means they can access shared resources like the file system,
user accounts, environment variables, and network sockets.
A process with sufficient permissions could access almost anything on the machine,
leading to management overhead and security concerns.

```d2
m: Host OS {
  p {
    class: none
    p1: Process 1 {
      class: process
    }
    p2: Process 2 {
      class: process
    }
  }
  r {
    class: none
    f: File system {
      class: file
    }
    u: User system {
      class: client
    }
  }
  p.p1 -> r.f
  p.p2 -> r.f
  p.p1 -> r.u
  p.p2 -> r.u
}
```

The **Kernel** helps address these concerns by enabling isolation.
Essentially, it allows for the creation of **isolated sections** within the machine,
where each section has its own view of resources and cannot directly see or share resources with other sections.

**Containerization** leverages this kernel feature to construct
**virtual operating systems** that are based on the host's kernel.
Each such virtual OS, with its own properties (like users, groups, files, and processes), is called a **Container**.

Despite all containers relying on the same underlying host kernel,
they are **logically isolated** from one another and cannot mutually share resources by default.

```d2
m: Host OS {
  k: Kernel {
    class: gw
  }
  c1: Container 1 {
    Users {
      class: client
    }
    Files {
      class: file
    }
    Processes {
      class: process
    }
  }
  c2: Container 2 {
    Users {
      class: client
    }
    Files {
      class: file
    }
    Processes {
      class: process
    }
  }
  c1 -> k: Rely on
  c2 -> k: Rely on
}
```

This distinction also clarifies the difference between a virtual machine and a container.
A **virtual machine (VM)** relies on a [Hypervisor](https://en.wikipedia.org/wiki/Hypervisor) to virtualize hardware,
allowing it to run a **separate, full kernel** and requiring a complete installation of its own operating system.
VMs are strongly isolated and use the host's physical resources as mediated by the hypervisor.

```d2
Host OS {
  grid-columns: 1
  vm {
    vm1: Virtual Machine 1 {
      OS {
        Resources {
            class: resource
        }
        k: kernel {
            class: gw
        }
      }
    }
    vm2: Virtual Machine 2 {
      OS {
        k: kernel {
            class: gw
        }
        Resources {
            class: resource
        }
      }
    }
  }
  hy: Hypervisor {
    class: process
  }
  k: Host Kernel {
    class: gw
  }
  vm.vm1.OS.k -> hy
  vm.vm2.OS.k -> hy
  hy -> k
}
```

## Resource Isolation

### Namespace

**Namespace** is a **Linux Kernel** feature that plays a crucial role in isolating system resources for containers.
In brief, when a process runs within a specific **namespace**,
it can only detect and interact with the resources that are also part of that same namespace.

Various types of resources can be segregated using namespaces.
We will focus on the primary ones:

### User Namespace

By default, within a standard OS environment,
a user can typically see other users and groups on the machine because
they all share the same underlying OS user management system.

**User Namespace** allows for the creation of isolated user systems.
Users within different user namespaces will have distinct views of user and can even have **overlapping IDs** without conflict.

```d2
direction: right
OS {
  grid-rows: 1
  c1: User Namespace 1 {
    u1: User 1 {
        class: client
    }
  }
  c2: User Namespace 2 {
    u1: User 1 {
        class: client
    }
    u2: User 2 {
        class: client
    }
  }
}
```

#### Namespace Mapping

Technically, a user within a namespace (a namespaced user) is ultimately a user on the host system.
When this namespaced user attempts to access resources,
the kernel maps its namespaced ID to a corresponding host user ID to perform authorization checks.

For example,
`User 1` within `Namespace 1` is actually be mapped to
`User 100` on the host system when performing actions that require kernel-level permissions.

```d2
OS {
  grid-rows: 1
  horizontal-gap: 100
  c: Namespace 1 {
    u1: User 1 {
        class: client
    }
  }
  k: Kernel {
    c: |||yaml
    Namespace Mappings:
      Namespace 1:
        User 1: User 100
    |||
  }
  r: Host User {
    u1: User 100 {
        class: client
    }
  }
  c.u1 -> k: Access
  k -> r.u1: Map to
}
```

To possess a private user system,
a container creates its own user namespace.
Within this namespace, a user is effectively represented by a real, often non-privileged, user on the host.

```d2
OS {
  grid-rows: 1
  horizontal-gap: 100
  c: Container {
    u1: User root {
        class: client
    }
  }
  r: Host User {
    u1: User 100 {
        class: client
    }
  }
  c.u1 -> r.u1: Mapped
}
```

{{< callout type="info">}}
This also answers a common question: *Is the root user of a container also the host's root?*
It is **configurable**.
The container's root user can be mapped to a non-privileged user on the host or,
less securely, to the host's actual root user.
{{< /callout >}}

### Process Namespace

Similar to users, processes in the same OS can typically see each other by default.

```d2
direction: right
OS {
  rp: Host Processes {
    grid-rows: 1
    p1: Process 100 {
        class: process
    }
    p2: Process 101 {
        class: process
    }
  }
}
```

For better control and isolation, **Process Namespace** provides isolation of **Process IDs (PIDs)**.
This means that processes running in different process namespaces are isolated from each other and can have **overlapping PIDs**.

Again, a mapping occurs: a PID within a process namespace corresponds to a different PID on the host system.

```d2
OS {
    r: Host Processes {
        p1: Process 100 {
            class: process
        }
        p2: Process 101 {
            class: process
        }
    }
    c1: Process Namespace 1 {
        p1: Process 1 {
            class: process
        }
    }
    c2: Process Namespace 2 {
        p1: Process 1 {
            class: process
        }
    }
    c1.p1 -> r.p1
    c2.p1 -> r.p2
}
```

A question might arise here: *Can users in different user namespaces see the same process?*

It's important to recognize that different types of namespaces serve distinct purposes and operate independently.
Each namespace type helps segregate a particular aspect of the system.

For example, `Process 1` could be placed in:

- `User Namespace 1`: It can only see and manage users (`User 1` and `User 2`) defined within this user namespace.
- `Process Namespace 1`: It can only see processes (`Process 2`) existing within this process namespace.

```d2
OS {
    p: Process Namespace 1 {
      p1: Process 2 {
        class: process
      }
    }
    u: User Namespace 1 {
      u1: User 1 {
        class: client
      }
      u2: User 2 {
        class: client
      }
    }
    p1: Process 1 {
      class: process
    }
    p1 -> p
    p1 -> u
}
```

### Network Namespace

Each network namespace possesses its own independent set of network interfaces (which are virtual),
IP addresses, firewall rules, routing tables, etc.

A network namespace's settings apply to the processes running inside.
These virtual network stacks ultimately still rely on the host's physical network,
and their network messages are processed by the host's network settings **before** exiting the machine.

```d2
OS {
  n1: Network Namespace 1 {
    grid-rows: 1
    p1: Process 1 {
        class: process
    }
    s: Network settings {
      class: process
    }
  }
  h: Host network settings {
    class: process
  }
  n1.p1 -> n1.s
  n1.s -> h
}
```

#### Bridging

Essentially, different network namespaces, despite coexisting on the same machine,
must communicate with each other (and the outside world) via networking mechanisms.
Network bridging is a popular technique to connect network namespaces.

In brief, a virtual network interface called a **bridge** is created on the host.
This special interface acts as a virtual switch between different network namespaces
(and potentially the host's network), allowing them to connect with each other by **private IPs**.

```d2
OS {
  ns1: "Network Namespace 1" {
      ni: "Network Interface (192.168.1.1)" {
          class: ni
          width: 400
      }
  }
  ns2: "Network Namespace 2" {
      ni: "Network Interface (192.168.1.19)" {
          class: ni
          width: 400
      }
  }
  b: Bridge Interface {
      class: ni
  }
  b <-> ns1.ni {
    style.animated: true
  }
  b <-> ns2.ni {
    style.animated: true
  }
}
```

However, this setup is typically suitable for internal workloads.
These private network IPs (e.g., `192.168.1.1`, `192.168.1.2`) cannot be exposed directly to the public internet,
as the outside world has no knowledge of them.

#### Port Mapping

Another strategy for connecting network namespaces to external networks is **Port Mapping**.
This is similar in concept to [Network Address Translation (NAT)](https://en.wikipedia.org/wiki/Network_address_translation).

The host can be configured to forward network messages to specific namespaces based on their **destination port** number.

```d2
OS {
    ns1: "Network Namespace 1" {
        ni: "Network Interface (192.168.1.1)" {
            class: ni
            width: 400
        }
    }
    ns2: "Network Namespace 2" {
        ni: "Network Interface (192.168.1.19)" {
            class: ni
            width: 400
        }
    }
    b: Host Interface {
        class: ni
    }
    b -> ns1: 8888
    b -> ns2: 9999
}
c: Client {
    class: client
}
c -> OS.b
```

### Mount Namespace

Mounting is the process of attaching a directory (known as a **mount point**) to a storage device in the file system,
making the data on that device accessible via that directory path.

For example:

- `Device 1` is mounted to `/boot`.
- `Device 2` is mounted to `/var`.

```d2
o: OS {
  b: "/boot" {
    class: folder
  }
  v: "/var" {
    class: folder
  }
}
d1: Storage Device 1 {
  class: hd
}
d2: Storage Device 2 {
  class: hd
}
o.b <-> d1
o.v <-> d2
```

With **mount namespaces**, processes can have different views of the filesystem hierarchy.
For example:

- In `Mount Namespace 1`, the path `/var` might point to `Device 1`.
- Simultaneously, in `Mount Namespace 2`, the same path `/var` could point to `Device 2`.

In other words, despite using the same directory path,
processes in different mount namespaces can be working with entirely different data sections.

```d2
grid-rows: 2
o: OS {
  grid-rows: 1
  horizontal-gap: 100
  n1: Mount Namespace 1 {
      "/var" {
        class: folder
      }
  }
  n2: Mount Namespace 2 {
      "/var" {
        class: folder
      }
  }
}
d1: Storage Device 1 {
  class: hd
}
d2: Storage Device 2 {
  class: hd
}
o.n1 -> d1
o.n2 -> d2
```

#### Bind Mount

Conveniently, a directory on the host filesystem can itself be treated as
if it were a storage device and mounted to another directory location.
This process is known as a **bind mount**.

```d2
o: OS {
  grid-rows: 1
  horizontal-gap: 100
  v: "/docs" {
    class: folder
  }
  t: "/team/shared/docs" {
    class: folder
  }
  v -> t: Mount
}
```

Mount namespaces each have their own distinct view of the filesystem.
Bind mounts can be used to cleanly mount host directories into these namespaces,
allowing for **isolated views** while also enabling controlled data sharing between the host and containers
(or between containers).

For example:

- `/var` inside `Mount Namespace 1` points to `/app1` on the host.
- `/var` inside `Mount Namespace 2` points to `/app2` on the host.
- Both of them share the same `/etc` directory.
Changes in any namespace will result in the same host's folder.

```d2
OS {
  grid-rows: 2
  n {
    class: none
    grid-rows: 1
    n1: Mount Namespace 1 {
        grid-rows: 1
        h: "/var" {
          class: folder
        }
        e: "/etc" {
          class: folder
        }
    }
    n2: Mount Namespace 2 {
        grid-rows: 1
        e: "/etc" {
          class: folder
        }
        v: "/var" {
          class: folder
        }
    }
  }
  f: OS filesystem {
      grid-rows: 1
      u: "/app1" {
        class: folder
      }
      e: "/etc" {
        class: folder
      }
      a: "/app2" {
        class: folder
      }
  }
  n.n1.e -> f.e
  n.n1.h -> f.u
  n.n2.e -> f.e
  n.n2.v -> f.a
}
```

### Cgroups

**Namespaces** alone are not sufficient for building robust containers because
they primarily provide isolation of view but not necessarily of resource consumption.
Without constraints, processes could compete for system resources
(CPU, memory, bandwidth) or some processes could consume resources inefficiently,
impacting the others.

The **Linux Kernel** includes a feature called **control groups (cgroups)** to address this.
A cgroup can define limits for resources like CPU time and memory allocation.
Processes associated with a cgroup **cannot** use more resources than the limits defined for that cgroup.

```d2
OS {
  grid-rows: 1
  horizontal-gap: 100
  cg1: "CGroup 1 (CPU < 50%, RAM < 3GB)" {
    p1: Process 1 {
        class: process
    }
    p2: Process 2 {
        class: process
    }
  }
  cg2: "CGroup 2 (CPU < 30%, RAM < 1GB)" {
    p1: Process 3 {
        class: process
    }
  }
}
```

### Container Creation

A container encompasses one or more processes running within a virtualized OS environment,
effectively separating them from the host system and other containers.

Creating a container typically involves:

1. Creating the appropriate **namespaces** (user, process, network, mount, etc.) to
provide resource isolation from the perspective of the processes within the container.
2. Creating and configuring a **cgroup** to limit and account for the resources (CPU, memory, etc.) that the container's processes can consume.
3. Running the container processes inside these configured namespaces and under the control of the cgroup.

```d2
grid-columns: 1
n: {
  class: none
  grid-rows: 1
  u: User Namespace 1 {
    class: client
  }
  n: Network Namespace 1 {
    class: ni
  }
  m: Mount Namespace 1 {
    class: file
  }
  p: Process Namespace 1 {
    class: process
  }
  c: Cgroup 1 {
    class: resource
  }
}
c: Container 1 {
  p1: Process 1 {
    class: process
  }
  p2: Process 2 {
    class: process
  }
}
c -> n.u
c -> n.n
c -> n.m
c -> n.p
c -> n.c
```

### Container Vulnerabilities

Running multiple applications or processes directly on the same host OS without isolation can be risky.
If one process is attacked or behaves erratically,
it can potentially bring down other processes or even the entire machine
because they all share the host's resources.

Containers help to build isolated environments,
effectively providing each application with what appears to be its own operating system.
This greatly mitigates many problems that can arise when a container is under attack or misbehaves.

Compared to virtual machines, containers are **not completely isolated** because they still share the host kernel.
Misconfigured containers can potentially be compromised in ways that could allow an attacker to affect the host system,
such as:

- **Mapping the host’s root user to a container’s user:**  
If the root user inside a container is directly mapped to the root user on the host,
any compromise of the container's root account could grant an attacker unrestricted access to the host kernel and the entire server.

- **Mounting sensitive host folders into a container:**  
When sensitive host directories (such as `/etc/passwd`) are bind-mounted into a container,
an attacker who gains control of the container may be able to access or tamper with critical data on the host system.

## Distributed Containerization

Taking containerization a step further,
the seamless integration of containers across multiple physical or virtual machines allows
for the construction of elegant and scalable distributed environments.

```d2
Server 1 {
    Container 1 {
        User Service 1 {
            class: process
        }
    }
    Container 2 {
        User Store - Shard 1 {
            class: db
        }
    }
}
Server 2 {
    Container 1 {
        User Service 2 {
            class: process
        }
    }
}
Server 3 {
    Container 1 {
        User Store - Shard 2 {
            class: db
        }
    }
}
```

One of the most critical challenges in such distributed container environments
is enabling containers to communicate with each other **transparently**,
irrespective of the underlying physical network infrastructure and on which host a container is running.

### Network Overlaying

**Network Overlaying** is a technique widely used in containerization solutions
to address cross-host container communication:

- An **underlay network** refers to the **physical network** infrastructure,
comprising actual routers, switches, and physical links.

```d2
un: "Underlay Networks" {
    grid-columns: 2
    n1: Network 1 {
        grid-rows: 1
        m {
          grid-columns: 1
          class: none
          m1: Machine 1 {
              class: server
          }
          m2: Machine 2 {
              class: server
          }
        }
        r: Router {
            class: router
        }
        m.m1 <-> r <-> m.m2
    }
    n2: Network 2 {
        grid-rows: 1
        r: Router {
            class: router
        }
        m {
          grid-columns: 1
          class: none
          m3: Machine 3 {
              class: server
          }
          m4: Machine 4 {
              class: server
          }
        }
        m.m3 <-> r <-> m.m4
    }
    n1.r <-> n2.r
}
```

- An **overlay network** is a virtual network that is logically superimposed on an existing physical (underlay) network infrastructure.
This layer is specifically configured to obscure the underlying network complexity,
allowing devices to interact as though they are directly connected within the same network.

```d2
grid-columns: 1
on: "Overlay Network" {
  grid-rows: 1
  m1: Machine 1 {
      class: server
  }
  m2: Machine 2 {
      class: server
  }
  m3: Machine 3 {
      class: server
  }
  m4: Machine 4 {
      class: server
  }
  m1 <-> m2 <-> m3 <-> m4
}
un: "Underlay Networks" {
    grid-columns: 2
    n1: Network 1 {
        grid-rows: 1
        m {
          grid-columns: 1
          class: none
          m1: Machine 1 {
              class: server
          }
          m2: Machine 2 {
              class: server
          }
        }
        r: Router {
            class: router
        }
        m.m1 <-> r <-> m.m2
    }
    n2: Network 2 {
        grid-rows: 1
        r: Router {
            class: router
        }
        m {
          grid-columns: 1
          class: none
          m3: Machine 3 {
              class: server
          }
          m4: Machine 4 {
              class: server
          }
        }
        m.m3 <-> r <-> m.m4
    }
    n1.r <-> n2.r
}
un.n1 -> on
un.n2 -> on
```

### Cluster Address

Instead of relying directly on the physical network addresses of the hosts,
each container in the overlay network is typically assigned a unique IP address within
the cluster's virtual address space.
This is often called a **Cluster Address**.

```d2
Cluster {
    "Host 1 (1.1.1.1)" {
        "Container 1 (192.168.1.1)" {
          class: container
        }
        "Container 2 (192.168.1.2)" {
          class: container
        }
    }
    "Host 2 (2.2.2.2)" {
        "Container 3 (192.168.1.3)" {
          class: container
        }
    }
}
```

Two main questions need to be addressed for this to work:

**1. How do hosts (and containers) discover the cluster addresses of other containers?**

#### Cluster Controller

Maintaining the state of this overlay network is often handled similarly to a [distributed database cluster]({{< ref "distributed-database" >}}).
A central **controller** typically manages the overlay network's settings and synchronizes them across the cluster:

- The cluster's network state changes frequently (e.g., containers starting, stopping, moving),
necessitating strong consistency to avoid routing errors.
- A central controller offers a simpler, more powerful point for managing and inspecting the cluster's network

```d2
Cluster {
    m: Controller Node {
        c: |||yaml
        Host 1:
            Address: 1.1.1.1
            Container 1: 192.168.1.1
            Container 2: 192.168.1.2
        Host 2:
            Address: 2.2.2.2
            Container 3: 192.168.1.3
        |||
    }
    h1: "Host 1 (1.1.1.1)" {
        "Container 1 (192.168.1.1)" {
          class: container
        }
        "Container 2 (192.168.1.2)" {
          class: container
        }
    }
    h2: "Host 2 (2.2.2.2)" {
        "Container 3 (192.168.1.3)" {
          class: container
        }
    }
    m -> h1: Sync settings {
        style.animated: true
    }
    m -> h2: Sync settings {
        style.animated: true
    }
}
```

**2. How is a packet actually transmitted between containers on different physical hosts?**

#### Packet Encapsulation

Let's consider an example of sending a packet from `Container 1` (on `Host 1`) to `Container 3` (on `Host 2`):

- **Encapsulation**: The initial packet is created with the source address of `Container 1` and the destination address of `Container 3`.
When this packet reaches the networking stack of `Host 1`, it is **encapsulated**.
This means the original packet (with Cluster IPs) is wrapped inside
the physical IP address of `Host 1` as its source and the physical IP address of `Host 2` as its destination.
Physical network devices then transmit the packet normally across the cluster.

```d2
Cluster {
    h1: "Host 1 (1.1.1.1)" {
        c1: "Container 1 (192.168.1.1)" {
            i: Initial Packet {
                p: |||yaml
                Source: 192.168.1.1 (Container 1)
                Destination: 192.168.1.3 (Container 3)
                |||
            }
        }
        e: "Encapsulated Packet" {
            p: |||yaml
            Outer Source: 1.1.1.1 (Host 1)
            Outer Destination: 2.2.2.2 (Host 2)
            Source: 192.168.1.1 (Container 1)
            Destination: 192.168.1.3 (Container 3)
            |||
        }
    }
    h2: "Host 2 (2.2.2.2)" {
        c1: "Container 3 (192.168.1.3)" {
          class: container
        }
    }
    h1.c1.i -> h1.e: Encapsulated {
      style.bold: true
    }
}
```

- **Decapsulation**: When the encapsulated packet arrives at `Host 2`,
its networking stack recognizes it as an overlay packet.
The outer header (with physical IPs) is stripped off (**decapsulated**),
revealing the original inner packet (with Cluster IPs).
`Host 2` then forwards the original packet to `Container 3`.

```d2
Cluster {
  grid-rows: 1
  h1: "Host 1 (1.1.1.1)" {
      c1: "Container 1 (192.168.1.1)" {
        i: Initial Packet {
            p: |||yaml
            Source: 192.168.1.1 (Container 1)
            Destination: 192.168.1.3 (Container 3)
            |||
        }
      }
      e: "Encapsulated Packet" {
          p: |||yaml
          Outer Source: 1.1.1.1 (Host 1)
          Outer Destination: 2.2.2.2 (Host 2)
          Source: 192.168.1.1 (Container 1)
          Destination: 192.168.1.3 (Container 3)
          |||
      }
  }
  h2: "Host 2 (2.2.2.2)" {
    e: "Encapsulated Packet" {
        p: |||yaml
        Outer Source: 1.1.1.1 (Host 1)
        Outer Destination: 2.2.2.2 (Host 2)
        Source: 192.168.1.1 (Container 1)
        Destination: 192.168.1.3 (Container 3)
        |||
    }
    c1: "Container 3 (192.168.1.3)" {
      d: "Decapsulated Packet" {
          p: |||yaml
          Source: 192.168.1.1 (Container 1)
          Destination: 192.168.1.3 (Container 3)
          |||
      }
    }
  }
  h1.c1.i -> h1.e: 1. Encapsulated
  h1.e -> h2.e: 2. Physical transmitted
  h2.e -> h2.c1.d: 3. Decapsulated
}
```
