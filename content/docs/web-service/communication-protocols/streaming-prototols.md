---
title: Streaming Protocols
weight: 20
---
We've highlighted the most popular protocols in the [previous topic](../).
While those are versatile and suitable for various use cases, they aren't particularly optimized for streaming media data — large, continuous streams like video or audio. Given the increasing importance of media streaming today, let’s take a brief look at some protocols better suited for this purpose.

## WebRTC

{{< term webrtc >}} (Web Real-Time Communication) is a protocol that enables peer-to-peer communication for audio, video, and data.
The core principle of {{< term webrtc >}} is to connect clients directly without relying on a central server.

### Public Address

By default, clients sit behind network routers,
with their public addresses represented by the router’s address via
[NAT](https://en.wikipedia.org/wiki/Network_address_translation) and an ephemeral port.

```d2
direction: right
c: "Client (192.168.1.2)"{
    class: client
}
r: "Router (1.2.3.4)" {
    class: router
}
i: Internet {
    class: internet
}
c -> r
r -> i: "Represented by 1.2.3.4:80" {
    style.bold: true
}
```

In theory, if two clients can exchange their public addresses,
they may establish a direct connection.

```d2

grid-columns: 4
grid-gap: 200
c1: "Client 1 (192.168.1.2)"{
    class: client

}
r1: "Router 1 (1.2.3.4:80)" {
    class: router
}
r2: "Router 2 (4.5.6.7:90)" {
    class: router
}
c2: "Client 2 (192.168.1.2)"{
    class: client
}

c1 -> r1
c2 -> r2
r1 <-> r2: Communicate {
    style.animated: true
}
```

### STUN Server

**STUN** (Session Traversal Utilities for NAT) is a lightweight service that helps reveal a machine’s public address.

```d2
direction: right

c1: "Client 1 (192.168.1.2)"{
    class: client
}
s: STUN server {
    class: server
}
c1 -> s: What is my address? {
    style.bold: true
}
s -> c1: "1.2.3.4:80" {
    style.bold: true
}
```

Why not rely solely on the router’s address?
Because the nearest router might not be a public-facing one — sometimes it only serves a local network.

By using a **STUN** server,
clients can discover their public addresses and attempt to establish direct connections.
We’ll cover how they exchange these addresses in the next section.

```d2

shape: sequence_diagram
c1: "Client 1"{
    class: client
}
c2: "Client 2"{
    class: client
}
s: "STUN Server" {
    class: server
}
c1 <- s: "Address = 1.2.3.4:80"
c2 <- s: "Address = 4.5.6.7:90"
c1 <-> c2: Exchange address {
    style.bold: true
}
c1 <-> c2: "Connect"
```

### TURN Server

Sometimes, addresses provided by a **STUN** server aren’t enough.
If two clients haven’t communicated before,
many routers are configured to reject unfamiliar addresses.
If both clients reject each other, a direct connection becomes impossible.

```d2

shape: sequence_diagram
c1: "Client 1"{
    class: client
}
s: "STUN Server" {
    class: server
}

c2: "Client 2"{
    class: client
}
c1 <- s: "Address = 1.2.3.4:80"
c2 <- s: "Address = 4.5.6.7:90"
c1 -> c2: "Connect"
c2 -> c1: "Reject because 1.2.3.4:80 is strange" {
    class: error-conn
}
c2 -> c1: "Connect"
c1 -> c2: "Reject because 4.5.6.7:90 is strange" {
    class: error-conn
}
```

A **TURN** (Traversal Using Relays around NAT) server helps in these cases by acting as a relay server between clients.

For example, `Client 2` connects to a **TURN** server,
becoming a recognized destination, and then `Client 1` can send messages through the server.

```d2

shape: sequence_diagram
c1: "Client 1"{
    class: client
}
s: "TURN Server" {
    class: server
}
c2: "Client 2"{
    class: client
}
c2 -> s: Connect
c1 -> s: "Send message"
s -> c2: Forward messages fluently as this is a familiar address
```

### Interactive Connectivity Establishment (ICE)

**Interactive Connectivity Establishment (ICE)**
is responsible for identifying **potential pathways** for peer-to-peer connections.

Both clients (let’s call them `A` and `B`) gather possible ways to connect. These are called **ICE candidates** and usually include:

1. The local address

2. A public address obtained via a **STUN** server

3. **TURN** candidates for fallback when direct connections fail

```yaml
ICE A:
    Local address: 192.168.1.1
    Public address: 1.2.3.4:80
    TURN candidates: [3.3.3.3:70, 4.4.4.4:90]

ICE B:
    Local address: 192.168.1.1
    Public address: 4.5.6.7:90
    TURN candidates: [5.5.5.5:70, 4.4.4.4:90]
```

### Signaling

Finally, they will exchange their **ICE candidates** through **signaling** — a separate mechanism not handled by {{< term webrtc >}} itself.

- This could be a lightweight {{< term ws >}} server.

- Or done manually via QR codes, chat messages, etc.

Once candidates are exchanged, they attempt to connect using the **most efficient path**:
Local address (if on the same network) → Public address → TURN server (as a last resort).

```d2
shape: sequence_diagram

ca: Client A {
    class: client
}
s: Signalling
cb: Client B {
    class: client
}
cb -> s: Send candidates
s -> ca: Transmit B candidates
ca -> s: Send candidates
s -> cb: Transmit A candidates
ca <-> cb: Connect by the most efficient way
```

### WebRTC Use Cases

How do clients discover **STUN** and **TURN** servers?
Some large services (like {{< term gg >}}) offer public servers, often hardcoded into browsers for seamless user experiences.
However, it’s possible to deploy custom servers if needed.

{{< term webrtc >}} is a powerful protocol that removes the need for a central media server and pushes more responsibility onto the clients.
That said, it’s best suited for one-to-one scenarios. In complex applications, like group calls with hundreds of participants, it can quickly overwhelm end-user devices or home networks.

## HTTP Live Streaming (HLS)

{{< term hls >}} (HTTP Live Streaming) is a media streaming protocol developed by `Apple`
that delivers video and audio content effectively.
Unlike protocols such as {{< term ws >}}, which depend on a central,
continuous live server, {{< term hls >}} offers a resilient, distributed system.

It works through **segmentation**, splitting audio or video into small, independent segments (files), usually a few seconds long.

- These segments are stored independently (typically an [Object Store]({{< ref "media-storage#object-storage" >}})),
potentially different servers.
- A **Master Record** (e.g., a {{< term sql >}} row) manages and indexes these segments.

```d2
grid-rows: 2
m: Master Record {
  grid-rows: 1
  grid-gap: 0
  s1: "Segment 1 (Length = 5s)" {
    width: 300
  }
  s2: "Segment 2 (Length = 5s)" {
    width: 300
  }
  s3: "Segment 3 (Length = 3s)" {
    width: 300
  }
}
s: Storage {
    s1: Server 1 {
        s1: "Segment_1.mp4" {
            class: file
        }
        s2: "Segment_2.mp4" {
            class: file
        }
    }
    s2: Server 2 {
        s3: "Segment_3.mp4" {
            class: file
        }
    }
}

m.s1 -> s.s1.s1
m.s2 -> s.s1.s2
m.s3 -> s.s2.s3
```

To play a video, the user first **fetches the master record**.
When seeking a specific moment, only the necessary segments are downloaded.
For example, to watch the `11th` second, only `Segment_2.mp4` would be retrieved.
In fact, to maintain a smooth experience, several sequential segments are usually preloaded in advance.
