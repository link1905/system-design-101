---
title: Containerization
weight: 10
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
h: Hardware {
  class: hw
}
c.c1 <-> k
c.c2 <-> k
k <-> h
```

### Kernel Isolation

Typically, processes run within a **shared operating system (OS)** environment without inherent isolation.
This means they can potentially access shared resources like the file system,
user accounts, environment variables, and network sockets.
A process with sufficient permissions could access almost anything on the machine,
leading to management overhead and security concerns.

```d2
m: Host OS {
  f: File system {
    class: resource
  }
  u1: User {
    class: user
  }
  p1: Process 1 {
    class: process
  }
  p2: Process 2 {
    class: process
  }
  p1 <-> f: Access
  p2 <-> f: Access
  p1 <- u1: Read info
  p2 -> u1: Update info
}
```

The **Kernel** helps address these concerns by enabling isolation.
Essentially, it allows for the creation of isolated sections within the machine,
where each section has its own view of resources and cannot directly see or share resources with other sections.

**Containerization** leverages this kernel feature to construct
**virtual operating system environments** that are based on the host's kernel.
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
      class: resource
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
      class: resource
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
allowing it to run a **separate, full guest kernel** and requiring a complete installation of its own operating system.
VMs are strongly isolated and use the host's physical resources as mediated by the hypervisor.

```d2
Host OS {
  grid-columns: 1
  vm {
    vm1: Virtual Machine 1 {
      OS {
        k: kernel {
            class: kernel
        }
        Apps {
            class: process
        }
      }
    }
    vm2: Virtual Machine 2 {
      OS {
        k: kernel {
            class: kernel
        }
        Apps {
            class: process
        }
      }
    }
  }
  hy: Hypervisor {
    class: process
  }
  k: Kernel {
    class: kernel
  }
  vm.vm1 -> hy
  vm.vm2 -> hy
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
Users within different user namespaces will have distinct views of user and group IDs and can even have overlapping IDs without conflict.

```d2
direction: right
OS {
  c1: User Namespace 1 {
    u1: User 1 {
        class: user
    }
  }
  c2: User Namespace 2 {
    u1: User 1 {
        class: user
    }
    u2: User 2 {
        class: user
    }
  }
}
```

#### Namespace Mapping

Technically, a user within a namespace (a namespaced user) is still ultimately a user on the host system.
When this namespaced user attempts to access resources,
the kernel maps its namespaced ID to a corresponding host user ID to perform authorization checks.

For example,
`User 1` within a `Dev Namespace` is actually be mapped to
`User 100` on the host system when performing actions that require kernel-level permissions.

```d2
OS {
  r: Host Users {
    u1: User 100 {
        class: user
    }
  }
  c: Dev Namespace {
    u1: User 1 {
        class: user
    }
  }
  k: Kernal {
    c: |||yaml
    Dev Namespace:
      User 1: User 100
    |||
  }
  c.u1 -> k: Access
  k -> r.u1: Map to
}
```

{{< callout type="info">}}
This mapping is generally established with a special file located at `/proc/self/uid_map`.
{{< /callout >}}

To possess a private user system,
a container creates its own user namespace.
Within this namespace, a user is effectively represented by a real, often non-privileged, user on the host.

```d2
OS {
  r: Root User Namespace {
    u1: User 100 {
        class: user
    }
  }
  c: Container {
    u1: Root {
        class: user
    }
  }
  c.u1 -> r.u1
}
```

{{< callout type="info">}}
This also answers a common question: *Is the root user of a container also the host's root?*
It is configurable.
The container's root user can be mapped to a non-privileged user on the host or,
less securely, to the host's actual root user
{{< /callout >}}

### Process Namespace

Similar to users, processes in the same OS can typically see each other by default.

```d2
direction: right
OS {
  rp: Host Processes {
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
This means that processes running in different process namespaces are isolated from each other and can have **overlapping PIDs**
(e.g., multiple containers can each have a process with PID 1).

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
A process can be run within specific combinations of namespaces to achieve desired isolation.

For example, `Process 1` could be placed in:

- `User Namespace 1`: It can only see and manage users defined within this user namespace.
- `Process Namespace 1`: It can only see processes existing within this process namespace.

```d2
OS {
    p: Process Namespace 1
    u: User Namespace 1 {
      u1: Root {
          class: user
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

Network namespace settings apply to the processes running inside that namespace.
These virtual network stacks ultimately still rely on the host's physical network,
and their network messages are processed by the host's network settings before exiting the machine.

```d2
OS {
  h: Host Network Settings {
    class: process
  }
  n1: Network Namespace 1 {
    s: Settings {
      class: process
    }
    p1: Process 1 {
        class: process
    }
  }
  n1.p1 -> n1.s
  n1.s -> h
}
```

#### Bridging

Essentially, different network namespaces, despite coexisting on the same machine,
must communicate with each other (and the outside world) via networking mechanisms.
Network bridging is a popular technique to connect network namespaces.

Why separate them only to connect them later?
This aligns with the containerization philosophy of isolation and security.
Processes are restricted to interacting only with network components within their own namespace,
preventing malicious or accidental access to others.

In brief, a virtual network interface called a **bridge** is created on the host. This special interface acts as a virtual switch or link between different network namespaces (and potentially the host's network), allowing them to behave as if they are on the same local network segment.

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
  b <-> ns1
  b <-> ns2
}
```

However, this setup is typically suitable for internal workloads.
These private network IPs (e.g., `192.168.1.1`, `192.168.1.2`) cannot be exposed directly to the public internet,
as the outside world has no knowledge of these private IP addresses.

#### Port Mapping

Another strategy for connecting network namespaces to external networks (and to each other via the host) is **Port Mapping**.
This is similar in concept to [Network Address Translation (NAT)](https://en.wikipedia.org/wiki/Network_address_translation).
A service running inside a container (and thus its network namespace) can be mapped to a specific port on the host machine.
External traffic directed to that host port will then be forwarded to the appropriate container.

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
    b: Host network interface {
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

Mounting is the OS process of attaching a directory (known as a **mount point**) to a storage device in the file system,
making the data on that device accessible via that directory path.

For example:
- `Device 1` is mounted to `/boot`.
- `Device 2` is mounted to `/var`.

```d2
o: OS {
  b: "/boot"
  v: "/var"
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
o: OS {
    n1: Mount Namespace 1 {
        "/var"
    }
    n2: Mount Namespace 2 {
        "/var"
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
  v: "/docs"
  t: "/team/shared/docs"
  v -> t: Mount
}
```

Mount namespaces each have their own distinct view of the filesystem.
Bind mounts can be used to cleanly mount host directories into these namespaces,
allowing for isolated views while also enabling controlled data sharing between the host and containers,
or between containers.

For example,
- `/var` inside `Mount Namespace 1` points to `/app1` on the host.
- `/var` inside `Mount Namespace 2` points to `/app2` on the host.
- Both of them share the same `/etc` directory.
Changes in any namespace will result in the same host's folder.

```d2
OS {
    f: OS filesystem {
        grid-columns: 1
        u: "/app1"
        a: "/app2"
        e: "/etc"
    }
    n1: Mount Namespace 1 {
        grid-columns: 1
        h: "/var"
        e: "/etc"
    }
    n2: Mount Namespace 2 {
        grid-columns: 1
        v: "/var"
        e: "/etc"
    }
    n1.e -> f.e
    n1.h -> f.u
    n2.e -> f.e
    n2.v -> f.a
}
```

### Cgroups

**Namespaces** alone are not sufficient for building robust containers because
they primarily provide isolation of *view* but not necessarily of *resource consumption*.
Without constraints, processes could compete for system resources
(CPU, memory, I/O) or one process could consume resources inefficiently,
impacting others or the host.

The **Linux Kernel** includes a feature called **control groups (cgroups)** to address this.
A cgroup can define limits for resources like CPU time, memory allocation, network bandwidth, and disk I/O.
Processes associated with a cgroup **cannot** use more resources than the limits defined for that cgroup.

```d2
OS {
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

Creating a container involves:
1. Creating the appropriate **namespaces** (user, process, network, mount, etc.) to
provide resource isolation from the perspective of the processes within the container.
2. Creating and configuring a **cgroup** to limit and account for the resources (CPU, memory, etc.) that the container's processes can consume.
3. Running the intended application processes inside these configured namespaces and under the control of the cgroup.

```d2
n: {
  class: none
  grid-rows: 1
  User Namespace 1 {
    class: user
  }
  Network Namespace 1 {
    class: ni
  }
  Mount Namespace 1 {
    class: file
  }
  Process Namespace 1 {
    class: process
  }
  Cgroup 1 {
    class: resource
  }
}
c: Container 1 {
  p1: Process 1 {
    class: process
  }
}
c.p1 -> n
```

Running multiple applications or processes directly on the same host OS without isolation can be risky.
If one process is attacked or behaves erratically,
it can potentially bring down other processes or even the entire machine
because they all share the host's resources.
Containers help to build isolated environments,
effectively providing each application with what appears to be its own operating system.
This greatly mitigates many problems that can arise when a process is under attack or misbehaves.

Compared to virtual machines, containers are not *completely* isolated because they still share the host system's kernel.
Misconfigured containers can potentially be compromised in ways that could allow an attacker to affect the host system:

- **Mapping the host's root user to a container's user**:
If the root user inside a container is mapped directly to the root user on the host,
a compromise of the container's root could grant an attacker full access to the host kernel and the entire server.
- **Mounting sensitive host folders into a container**:
Through bind mounting, if a sensitive host directory (e.g., `/etc/passwd`, `/var/run/docker.sock`)
is made accessible inside a container without proper restrictions,
an attacker gaining control of the container could maliciously access or modify data on the host.

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
is enabling containers to communicate with each other transparently,
irrespective of the underlying physical network infrastructure and on which host a container is running.

### Network Overlaying

**Network Overlaying** is a technique widely used in containerization solutions
to address cross-host container communication.

- An **underlay network** refers to the **physical network** infrastructure,
comprising actual routers, switches, and physical links.
- An **overlay network** is a **virtual network** that is built on top of one or more underlay networks.
Traffic flow is managed at the overlay layer,
while the actual data packets are transmitted over the physical underlay layer.

```d2
grid-columns: 1
on: "Overlay Network" {
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
}
un: "Underlay Networks" {
    grid-columns: 2
    n1: Network 1 {
        grid-rows: 1
        m1: Machine 1 {
            class: server
        }
        r: Router {
            class: router
        }
        m2: Machine 2 {
            class: server
        }
        m1 <-> r <-> m2
    }
    n2: Network 2 {
        grid-rows: 1
        m1: Machine 3 {
            class: server
        }
        r: Router {
            class: router
        }
        m2: Machine 4 {
            class: server
        }

        m1 <-> r <-> m2
    }
}
un.n1 -> on
un.n2 -> on
```


### Cluster Address

Instead of relying directly on the physical network addresses of the host machines,
each container in an overlay network is typically assigned a unique IP address within
the cluster's virtual address space.
This is often called a **Cluster IP**.

```d2
Cluster {
    "Host 1 (1.1.1.1)" {
        "Container 1 (192.168.1.1)"
        "Container 2 (192.168.1.2)"
    }
    "Host 2 (2.2.2.2)" {
        "Container 3 (192.168.1.3)"
    }
}
```

There are two questions needing revolving:

Two main questions need to be addressed for this to work:

**1. How do hosts (and containers) discover the cluster addresses of other containers, especially those on different hosts?**

This is often managed similarly to how a [distributed database cluster]({{< ref "distributed-database" >}}) maintains its state.
A central **Controller** component is typically used to manage the overlay network's addresses, routing rules, and overall state.
A strongly consistent component is often preferred because:

- The cluster's network state changes frequently (containers starting, stopping, moving)
and requires strong consistency to avoid routing errors.
- A controller provides a simpler and more powerful point for managing and introspecting the cluster's network.

```d2
Cluster {
    m: Metadata Store {
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
        "Container 1 (192.168.1.1)"
        "Container 2 (192.168.1.2)"
    }
    h2: "Host 2 (2.2.2.2)" {
        "Container 3 (192.168.1.3)"
    }
    m <-> h1: Sync settings {
        style.animated: true
    }
    m <-> h2: Sync settings {
        style.animated: true
    }
}
```


**2. How is a packet actually transmitted between containers on different physical hosts?**

Let's consider an example of sending a packet from `Container 1` (on Host 1) to `Container 3` (on Host 2):

- **Encapsulation**: The initial packet is created by `Container 1` with the source Cluster IP of `Container 1` and the destination Cluster IP of `Container 3`.
When this packet reaches the networking stack of `Host 1`, it is **encapsulated**.
This means the original packet (with Cluster IPs) is wrapped inside
the physical IP address of `Host 1` as its source and the physical IP address of `Host 2` as its destination.

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
        c1: "Container 3 (192.168.1.3)"
    }
    h1.c1.i -> h1.e: 1. Encapsulated
}
```

- **Decapsulation**: When the encapsulated packet arrives at `Host 2`,
its networking stack recognizes it as an overlay packet.
The outer header (with physical IPs) is stripped off (decapsulated),
revealing the original inner packet (with Cluster IPs).
`Host 2` then forwards the original inner packet to the correct local `Container 3`.

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
        c1: "Container 3 (192.168.1.3)"
        d: "Decapsulated Packet" {
            p: |||yaml
            Source: 192.168.1.1 (Container 1)
            Destination: 192.168.1.3 (Container 3)
            |||
        }

    }
    h1.c1.i -> h1.e: 1. Encapsulated
    h1.e -> h2.d: 2. Physical transmitted
    h2.d -> h2.c1: 3. Forwarded
}
```
