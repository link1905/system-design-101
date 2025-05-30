---
title: Containerization
weight: 10
---

**Containerization** is a lightweight form of virtualization gaining significant momentum in recent years,
it's becoming a cornerstone of modern software development and deployment practices.


## Kernel

To understand **Containerization** in depth,
let's first the foundations to build it.

Users should not work with hardware directly,
that creates many problems about: security, resource management, hardware compatibility.
**Kernel** is the core component of an operating system, works a bridge between **processes** and the **hardware**;
To work with hardware, user commands are passed and processed through the kernel first.

```d2
grid-columns: 1
p: {
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

Normally,
processes runs in a **shared operating system (OS)** without any isolation,
e.g., users/groups, files, environment variables, network sockets...
a process with enough permission can access anything in the machine,
resulting in a management overhead.

```d2
m: OS {
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

**Kernel** helps segregate the concern by allowing isolation.
Essentially, we can build isolated sections in the machine,
they see and share nothing in between.

**Containerization** leverages this feature to build **virtual operating systems** based on the host's kernel.
Each virtual OS has its own properties (users, groups, files, processes...), called a **Container**.
Despite depending on the same kernel, containers are **logically isolated** and cannot share resources mutually.

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

This also tells apart between a virtual machine and a container.

- A virtual machine is based on a [Hypervisor](https://en.wikipedia.org/wiki/Hypervisor)
to virtualize a **separate kernel** and require a **full** installation of the OS.
Virtual machines are isolated and use the host's resources through the hypervisor.

```d2
%d2-import%
direction: right
h: Host {
    class: kernel
}
hy: Hypervisor {
    class: process
}
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
vm1.OS.k -- hy
vm2.OS.k -- hy
hy -- h
```

## Resource Isolation

### Namespace

**Namespace** is a **Linux Kernel** feature helping isolate system resources.
In brief, when a process runs in a specific **namespace**,
it can only detect and communicate with the resources within that namespace.

There are many types of resources we can separate,
we just focus on the primary ones:

### User Namespace

By default,
a user can see other users or groups in the machine
as they share the same OS.

**User Namespace** helps to create isolated user systems.
Users come from different namespaces will have different views and can have overlapped ids.

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

Technically, a namespaced user is just the host's user.
When accessing resources, the kernel will map it to the corresponding host user to perform authorization.
E.g., `User 1` in `Dev Namespace` is actually mapped to `User 100` when performing actions on the kernel.

{{< callout type="info">}}
This mapping is generally established with a special file located at `/proc/self/uid_map`.
{{< /callout >}}

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

To posses a private user systm,
a container creates its own user namespace,
where a user is effectively represented by a real user on the host.

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
It's configurable, the container's root user can be mapped to a trivial user or to the host's root.
{{< /callout >}}

### Process Namespace

Similar to users,
by default, processes in the same OS can see each other.

```d2
%d2-import%
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

For better control, **Process Namespace** provides isolation of **PID (Process Id)**.
In other words, processes in different namespaces are isolated and can have **overlapped IDs**.

Again, we also need to map a **Process Namespace** to the host's process.

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

Here, we might ask *do users in different user namespaces can see the same process?*.
Please acknowledge that different types of namespaces serve different purposes and **irrelevant mutually**.
Each namespace helps segregate an aspect, we can run a process in certain namespaces to isolate it.

For example, `Process 1` is put in:

- `User Namespace 1`: it can only see and manage users in this namespace
- `Process Namespace 1`: it can only see processes in this namespace

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

Each network namespace has its own set of network interfaces (virtual),
IP addresses, firewall, routing tables...

Network namespace apply their settings to the processes inside.
Probably, they still rely on the host,
their network messages are still processed by the host settings before going out of the machine.

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

Basically, different network namespaces must communicate through network, despite living in the same machine.
Network bridging is a popular approach to connect network namespaces through network.

Why do we need to separate but later connect them?
That's the philosophy of containerization - isolation and security.
Processes are only permitted to interact with network components within their namespace to
prevent malicious access to others.

In brief, we create a virtual interface called **bridge** on the host,
this special interface acts as a link between networks allowing them to behave as the same network.

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

But this is merely reasonable for internal workloads,
we cannot expose these networks publicly,
since the outer world has no idea of private IPs `192.168.1.1` or `192.168.1.19`.

#### Port Mapping

Another connecting strategy is **Port Mapping**, it's similar to [NAT](https://en.wikipedia.org/wiki/Network_address_translation).
A namespace will be paired with a fixed port on the host,
messages will be forwarded to the proper namespace by their target port.

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

Mounting is the process of attaching a directory (called a **mount point**) to a storage device,
making data on the device accessible from that directory.

For example
- `Device 1` is mounted to `/boot`
- `Device 2` is mounted to `/var`
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
o.b -> d1
o.v -> d2
```

With mount namespaces, we can switch between them to have different filesystem looks.

For example, in `Mount Namespace 1`,
`/var` points to the `Device 1`.
Meanwhile, `/var` actually points the `Device 2`.
In other words, despite having the same path, they work with different data sections.

```d2
%d2-import%
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

Conveniently, we can treat a directory as a storage device, and mount it to another directory.
This process is known as **bind mount**.

```d2
o: OS {
  v: "/docs"
  t: "/team/shared/docs"
  v -> t: Mount
}
```

Mount namespaces have their own view of filesystem,
we can cleanly mount host directories to isolate and share data between them.
For example,
- `/var` of `Mount Namespace 1` points to `/app1` on the host.
- `/var` of `Mount Namespace 2` points to `/app2` on the host.
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

**Namespace** is still not enough to build containers because of no constraint on processes,
they can compete with each other or consume resources inefficiently.

**Linux Kernel** adds a feature called **control groups (cgroups)**.
In brief, **cgroups** helps to limit the resource usage of a processes group: cpu, memory, network bandwidth...
A group contains some processes and **cannot** use more resources than its limits.

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

A container include some processes in a virtual OS,
helping seperating them from the host and other containers.

Creating the container means:
1. Creating appropriate namespaces (user, process, network, etc) and cgroup to isolate resources.
2. Running the processes inside these namespaces.

Running multiple processes in the same environment is dangerous!
When a process is attacked, it can pull down the others (or the entire machine) accidentally,
because processes are sharing the host's resources.
Containers help to build isolated environments, as if they have isolated operating systems,
mitigate a lot of problems when a process is under attack.

Compared to virtual machines,
containers are not completely isolated as still using the same kernel.
Misconfigured containers can be easily compromised to attack the entire host:

- Mapping the host's root user to a container's user.
An attacker can access the kernel to attack the entire server.
- Mounting a sensitive folder (`/etc/passwd`🥺) to a container.
Through bind mounting, an attaker can access its data malicously.
- ...

## Distributed Containerization

A step forward,
seamless integration of containers across multiple machines helps build an elegant distributed environment.

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

One of the most critical challenges is setting
distributed containers to communicate transparently,
ignoring of how the underlying infranstructure works.

### Network Overlaying

**Overlaying** is a technique widely used in containerization solutions.

An **underlay network** refers to a **physical network** with real devices.
Meanwhile, an **overlay network** is a **virtual network** built on top of the physical ones.
We will control traffic flow with the overlay layer and actually transmit data with the physical layer.

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

Instead of relying on the physical network’s addresses,
each container is marked with a unique address within the cluster called **Cluster Address**.

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

1. First, how do hosts know about the address of other containers?

This is similar to maintain a [database cluster]({{< ref "distributed-database" >}}).
A **Controller** server is built to manage addresses and forwarding rules of the overlay network,
rather than the decentralized style, because:

- The cluster's state is frequently modified and requires strong consistency.
- We can control the cluster more simply and deeply with the controller.

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


2. The second question is how a packet is transmitted between physical hosts?

Let's see an example of sending a packet from `Container 1` to `Container 3`
- **Encapsulation**: A packet is created with cluster addresses.
  When it comes to the host,
  it will be encapsulated with **outer headers** of the physical addresses for transmission.
- **Decapsulation**: when the packet arrives at the destination host,
  the outer address is discarded to get the initial address to forward to the appropriate container.

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
