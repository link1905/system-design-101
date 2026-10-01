---
title: Communication Protocols
weight: 30
prev: service-cluster
next: streaming-protocols
---

The communication protocol fundamentally shapes how a service is designed and implemented.
With so many options available, selecting the right one requires a thorough understanding of how each protocol works.

## Hypertext Transfer Protocol (HTTP)

**Hypertext Transfer Protocol (HTTP)** is built on top of **Transmission Control Protocol (TCP)**
and is one of the most widely used communication protocols in modern systems.

The concept is straightforward: a client sends a request and receives the corresponding response.

### HTTP/1.0

The initial version of {{< term http >}} establishes a separate {{< term tcp >}} connection for each request.

```d2
shape: sequence_diagram
c: Client {
    class: client
}
s: Service {
    class: server
}
c <-> s: Establish TCP connection {
  style.bold: true
}
c -> s: Request
c <- s: Response
c <-> s: Close the connection {
  style.bold: true
}
```

Establishing a {{< term tcp >}} connection is *resource-intensive*,
especially when using [SSL/TLS](https://en.wikipedia.org/wiki/Transport_Layer_Security).
This becomes inefficient when clients need to make multiple requests,
as a large number of connections must be established.

### HTTP/1.1

{{< term http1 >}} improved efficiency by keeping a connection open for a short period before closing it.
This behavior is controlled by the [Keep-Alive](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Keep-Alive)
header, which determines how long the connection remains active.

```d2
shape: sequence_diagram
c: Client {
    class: client
}
s: Service {
    class: server
}
c <-> s: Establish TCP connection (Keep-Alive = 10) {
  style.bold: true
}
c -> s: Request
c <- s: Response
c -> s: Request
c <- s: Response
c <-> s: ...
c <-> s: Close the connection after 10 seconds {
  style.bold: true
}
```

{{< term http >}} has several potential drawbacks:

- **Synchronous limitation**: {{< term http >}} generally requires the client to wait for a request to complete,
  making it inefficient for long-running tasks that are better handled *asynchronously*.
- **One-way initiation**: Requests always originate from the client,
  so the server cannot independently initiate communication with the client.

However, its simplicity and lightweight nature make it highly effective in many scenarios.
This protocol is particularly useful for **simplifying communication** between clients and servers, such as:

- Client-facing services.
- Public APIs exposed to external systems.

## Polling

### Short Polling

To approximate **bidirectional** communication using {{< term http >}},
a naive approach is to have clients continuously send requests to the server to check for new notifications.

```d2
shape: sequence_diagram
c: Client {
    class: client
}
s: Service {
    class: server
}
c -> s: Is anything new?
c <- s: No
c -> s: Is anything new?
c <- s: Yes, abc123 has sent you a message
```

This approach is known as {{< term spoll >}}.
It is highly inefficient in terms of bandwidth,
as hundreds of requests may be sent just to retrieve a single notification.

### Long Polling

To improve efficiency, the server can hold a request for a *short period* before responding.
This brief delay significantly reduces the number of unnecessary requests.
This pattern is known as {{< term lpoll >}}.

```d2
shape: sequence_diagram
c: Client {
    class: client
}
s: Service {
    class: server
}
o: Another service {
    class: server
}
c -> s: Is anything new?
s -> s: Hold the request for 10 seconds {
   style.bold: true
}
c <- s: Time out {
  style.bold: true
}
c -> s: Is anything new?
s -> s: Hold the request for 10 seconds {
   style.bold: true
}
o -> s: Send message to the client
c <- s: Respond to the client immediately {
  style.bold: true
}
```

{{< term lpoll >}} is a traditional technique for delivering near-real-time notifications from the server.
Because requests still originate from clients, {{< term lpoll >}} is well-suited for:

- **Decoupling** the server from the client. Clients determine when to retrieve messages,
  so the server does not need to "seek" active client connections within the system.
- **Backpressure-aware clients**,
  allowing clients to control their polling behavior independently,
  such as introducing delays between polls or specifying how many messages to retrieve per request.

{{< term lpoll >}} is best implemented using a *stateless service*.
Because connections are short-lived, clients can easily switch between service instances and retrieve data from a shared store.

For example, instances of a stateless service can share and poll the same store.

```d2
s: Service {
  vertical-gap: 100
  i1: Instance 1 {
    class: server
  }
  s: Shared Store {
    class: db
  }
  i2: Instance 2 {
    class: server
  }
  i1 <- s: Pull {
    style.animated: true
  }
  i2 -> s: 2. Update
}
c: Client {
  class: client
}
o: Another service {
  class: server
}
c <- s.i1: Pull {
  style.animated: true
}
o -> s.i2: 1. Send message to the client
```

However, this model does not provide true *real-time communication*.
Because clients determine when to retrieve data, messages cannot necessarily be delivered immediately after they are created.

## WebSocket

{{< term ws >}} is a newer technology than {{< term lpoll >}}.
In short, a {{< term ws >}} server maintains *long-lived connections*,
allowing *both sides* to actively exchange messages over the same connection.

```d2
shape: sequence_diagram
c: Client {
    class: client
}
s: WebSocket {
    class: server
}
c <-> s: Establish a connection
s --> c: Server sends message
c --> s: Client sends message
c <-> s: "...More actions on the connection..." {
  style.animated: true
}
```

In general, {{< term ws >}} provides lower latency than {{< term lpoll >}} because the connection is already established and ready for communication.
It is particularly well-suited for bidirectional, low-latency applications such as gaming and chat services.

Some important drawbacks of {{< term ws >}} include:

- **Availability and routing complexity**: When a service cluster distributes client connections across multiple instances,
  an instance receiving a message may need to locate and forward it to the instance holding the target client's connection.
  This additional coordination increases system complexity and can affect availability.
- **Resource utilization**: A {{< term ws >}} connection is long-lived and remains tied to a specific server,
  which can lead to uneven resource utilization.
  For example, one client may continuously interact with a single server while other instances remain underutilized,
  even though distributing the workload across instances would be more efficient.

```d2
c1: Client 1 {
    class: client
}
s: Service {
  direction: right
  i1: Instance 1 {
    class: server
  }
  i2: Instance 2 {
    class: server
  }
  i3: Instance 3 {
    class: server
  }
}
c1 -> s.i1: Tied to {
  style.animated: true
}
```

### Stateful Misconception

Does maintaining long-lived connections automatically make a service stateful?
The answer is no.

The communication protocol itself does not determine whether a service is {{< term sf >}} or {{< term sl >}}.
That property depends on *how the service is implemented*.

Returning to the chat example in the [previous topic]({{< ref "service-cluster#stateful-service" >}}),
we described it as a stateful service because user connections were maintained by specific servers.

```d2
direction: right
grid-rows: 1
c: Clients {
  class: none
  grid-columns: 1
  ca: Client A {
    class: client
  }
  cb: Client B {
    class: client
  }
}
system: System {
    grid-columns: 1
    s1: Instance 1 {
      class: server
    }
    s2: Instance 2 {
      class: server
    }
    s1 <-> s2
}
c.ca <-> system.s1: Connecting
c.cb <-> system.s2: Connecting
```

Now, consider a different approach.
Instead of forwarding messages directly between instances,
each instance periodically retrieves messages from a shared store.

The service can now be stateless.
Every instance behaves identically,
so it does not matter which instance a client connects to.

```d2
direction: left

c1: Client 1 {
    class: client
}
s: Service {
  i1: Instance 1 {
    class: server
  }
  s: Shared Store {
    class: db
  }
  i2: Instance 2 {
    class: server
  }
  i1 <- s {
    style.animated: true
  }
  i2 <- s {
    style.animated: true
  }
}
c2: Client 2 {
    class: client
}
c1 <- s.i1
c2 <- s.i2
```

In practice,
{{< term ws >}} services are often implemented as stateful services
because WebSocket is frequently used for *real-time communication*,
where messages should be delivered immediately after they are created.

A polling-based approach is less suitable for such workloads because it introduces delays, even if those delays are brief.
For strict real-time delivery, maintaining connection-related state is often necessary.

## Server-Sent Events

As the name suggests, **Server-Sent Events (SSE)** provides *unidirectional* communication.
It maintains a *long-lived connection* while allowing data to flow only from the server to the client.

```d2
shape: sequence_diagram
c: Client {
    class: client
}
s: SSE Service {
    class: server
}
c <-> s: Establish a connection
s --> c: Send message
s --> c: Send message
```

Behind the scenes, {{< term sse >}} is built on top of the {{< term http >}} protocol.
As a result, developing and maintaining an SSE application is generally simpler than using {{< term ws >}},
because it can leverage existing {{< term http >}} infrastructure and tooling.

Additionally, unidirectional communication typically incurs less overhead than full-duplex communication.
{{< term sse >}} is therefore a good choice when an application only needs to push data from the server to the client,
such as for live scores or news feeds.

Similar to {{< term ws >}}, {{< term sse >}} maintains long-lived connections
and therefore introduces similar challenges related to connection affinity and resource balancing.

## Google Remote Procedure Call (gRPC)

{{< term grpc >}} is a modern technology developed by `Google`
that supports both bidirectional and unidirectional communication
using **Remote Procedure Call (RPC)** over the {{< term http2 >}} protocol.

### Remote Procedure Call (RPC)

When calling a conventional {{< term http >}} endpoint,
an application must typically handle several details, such as the URI, headers, and parameters,
to construct a valid request.
Although this approach provides flexibility, it can also introduce complexity and increase the likelihood of errors.

```http
GET /docs?name=README&team=dev HTTP/2
```

In contrast, {{< term rpc >}} provides a more structured communication model,
requiring both the client and server to agree on a *shared contract* that defines the exposed operations.

This contract is typically used to generate client and server code,
making remote interactions resemble calls to local functions.

For example, suppose the `Chat Service` exposes a `Chat` function.
The service contract can define the function and its request and response types for consumers.

```proto
// Exchange schema
message ChatRequest {
  string content;
}
message ChatResponse {
  string messageId;
}

// Service definition
service ChatService {
  rpc Chat (ChatRequest) returns (ChatResponse);
}

// Service is called from the client side conveniently
var chatService = new ChatService();
var chatResponse = chatService.Chat(new ChatRequest("Hello Bro!"));
```

Another advantage of {{< term rpc >}} is efficient serialization.
Formats such as {{< term json >}} and {{< term xml >}} are commonly used for data exchange because of their flexibility and broad compatibility,
but text-based representations are generally larger and more expensive to process than compact binary formats.
With a predefined schema, {{< term rpc >}} frameworks can generate efficient binary serializers,
such as [Protocol Buffers](https://protobuf.dev/overview/).

One drawback of {{< term rpc >}} is its reliance on shared contracts and strict schemas, which can create friction for clients we don't control.
For this reason, {{< term grpc >}} is more commonly used for controlled service-to-service communication than for public-facing services.

### HTTP/2

{{< term http1 >}} uses a connection between the client and server,
with requests and responses transferred through that connection.

{{< term http2 >}} improves this model by dividing a connection into *independent streams*,
allowing multiple requests and responses to be transmitted concurrently.

For example:

- In the {{< term http1 >}} example, `dog.png` is requested only after `index.html` has been fetched.
- In the {{< term http2 >}} example, both requests can be transmitted concurrently through `Stream 1` and `Stream 2`,
  allowing the resources to be downloaded in parallel.

```d2
"HTTP/1.1" {
  shape: sequence_diagram
  c: Client {
    class: client
  }
  s: Server {
    class: server
  }
  c -> s: Request index.html
  c <- s: Respond index.html
  c -> s: Request dog.png
  c <- s: Respond dog.png
}
http2: "HTTP/2" {
  shape: sequence_diagram
  c: Client {
    class: client
  }
  s: Server {
    class: server
  }
  "1. Loads requests" {
    c --> s: Stream 1: Request index.html
    c --> s: Stream 2: Request dog.png
  }
  "2. Response" {
    c <-- s: Stream 1: Respond index.html
    c <-- s: Stream 2: Respond dog.png
  }
}
```

Behind the scenes, {{< term http2 >}} still uses a single {{< term tcp >}} connection,
with frames associated with individual **Stream IDs**.
Frames belonging to the same stream can be reassembled independently,
allowing multiple logical streams to share a single connection concurrently.

### Use Cases {id="grpc_use_cases"}

Returning to {{< term grpc >}},
it is built on top of {{< term http2 >}} and {{< term rpc >}},
making it highly effective for handling multiple concurrent requests and streams.

Similar to {{< term ws >}} and {{< term sse >}},
{{< term grpc >}} can also maintain *long-lived connections*.

However, more sophisticated communication patterns generally require additional processing and connection management.
{{< term grpc >}} may therefore consume more computational resources when handling large numbers of concurrent streams and messages.

If a service does not require RPC semantics or multiplexed request-response streams
and instead primarily needs ordered message exchange,
{{< term ws >}} or {{< term sse >}} may provide a simpler communication model, depending on whether communication needs to be bidirectional or server-to-client only.

## Webhook

{{< term wh >}} is an effective mechanism for handling *long-running or asynchronous operations*.
Conceptually, it resembles a callback function in programming.

The client registers a callback endpoint, usually a {{< term url >}}, with the server.
The server can later invoke that endpoint when an event occurs or a result becomes available.

For example, a client may register a callback address.
Whenever the server needs to notify the client,
it sends a request to `site.com/callback`.

```d2
shape: sequence_diagram
c: Client {
    class: client
}
s: Webhook server {
    class: server
}
cb: "site.com" {
  class: server
}
c --> s: 'Register "site.com/callback"'
s --> s: The client has a new notification
s --> cb: "/callback"
```

This approach is particularly useful for tasks with *unpredictable execution times*,
because it avoids wasting resources while clients wait for completion.

For example, during payment processing,
a transaction may pass through multiple banking systems, potentially across different countries,
and may therefore take an unpredictable amount of time to complete.

### Use Cases {id="webhook_use_cases"}

{{< term wh >}} provides an efficient event-driven communication model.
Data is transmitted only when an event occurs,
eliminating the need for continuous polling or long-lived client connections.

This approach is commonly used by intermediary systems that communicate with numerous trusted external clients,
such as **Stripe Payments**.

However, webhooks are generally impractical for directly serving end users,
because end-user devices typically do not expose a stable *publicly reachable address* that can receive callbacks.

Furthermore, because the server initiates requests to client-provided endpoints,
webhook systems must carefully address security concerns such as authentication, endpoint validation, and request verification.

## HTTP Live Streaming (HLS) {id=hls}

We've highlighted some of the most popular protocols in the previous sections.
While they are versatile and suitable for a wide range of use cases, they aren't specifically optimized for streaming media such as video and audio.

{{< term hls >}} is a media streaming protocol developed by **Apple** for efficiently delivering video and audio content.
Unlike protocols such as {{< term ws >}}, which typically rely on a persistent connection between the client and a server, {{< term hls >}} delivers media as a sequence of independent files over HTTP.
This design makes it well suited for distributed delivery through web servers and CDNs.

HLS works through **segmentation**, splitting audio or video into small, independent segments, typically a few seconds long.

- These segments can be stored independently and distributed across multiple servers.
- A **playlist**, usually an `.m3u8` file, describes the available segments and their locations.

```d2
grid-rows: 2
m: Playlist {
  grid-rows: 1
  grid-gap: 0
  s1: "Segment 1 (Length = 5s)"
  s2: "Segment 2 (Length = 5s)"
  s3: "Segment 3 (Length = 3s)"
}
s: Storage {
  grid-rows: 1
  s1: Server 1 {
    grid-rows: 1
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

To play a video, the client first **fetches the playlist** to determine which media segments are available and where to retrieve them.
When seeking to a specific point in the video, the client can request the segment containing that point instead of downloading the entire media file.

For example, with the segments shown above, seeking to the `11th` second would require `Segment_3.mp4`, since the first two segments cover the first 10 seconds.
In practice, clients usually buffer several sequential segments in advance to provide smooth, uninterrupted playback.

We'll discuss the storage aspect in more detail in a [later topic]({{< ref "media-storage" >}}).
