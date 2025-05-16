---
title: Communication Protocols
weight: 30
---

The communication protocol fundamentally shapes the way a service is built.
With so many types available, selecting the right one requires a thorough understanding of their workflows.

## Hypertext Transfer Protocol (HTTP)

**Hypertext Transfer Protocol (HTTP)** is built on top of **Transmission Control Protocol (TCP)**
and is widely regarded as the most common solution in many systems.

The concept is simple: clients send a request and receive an associated response immediately.

### HTTP/1.0

The initial version of {{< term http  >}} establishes a separate {{< term tcp >}} connection for each request.

```d2
shape: sequence_diagram
c: Client {
    class: client
}
s: Service {
    class: server
}
c <-> s: 1. Establish TCP connection {
  style.bold: true
}
c -> s: Request
c <- s: Response
c <-> s: Close the connection {
  style.bold: true
}
c <-> s: 2. Establish TCP connection {
  style.bold: true
}
c -> s: Request
c <- s: Response
c <-> s: Close the connection {
  style.bold: true
}
```

Creating a {{< term tcp >}} connection is **resource-intensive**,
especially when using [SSL]({{< ref "network-security#transport-layer-security-tls" >}}).
This becomes inefficient when clients need to make multiple requests simultaneously,
as numerous connections will be established as a result.

### HTTP/1.1

{{< term http1 >}} introduced an improvement by keeping a connection open for a short duration before disposing of it.
This behavior is controlled by the [Keep-Alive](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Keep-Alive)
header, which specifies the connection's lifespan.

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

{{< term http  >}} has some potential drawbacks:

- **Synchronous limitation**: {{< term http  >}} requires waiting for the request to complete,
  which is inefficient for long-running tasks better suited to the **asynchronous manner**.
- **One-way communication**: Requests always originate from the client side,
  with no mechanism for the server to actively send messages back.

However, the simplicity and lightweight make it beneficial in many scenarios.
This protocol is ideal for **simplifying communication** between the client and server sides, such as in:

- Client-facing services.
- Public APIs exposed to external systems.

## Polling

### Short Polling

To enable **bidirectional** communication with {{< term http  >}},
a naive approach involves having clients continuously request to the server side to pull new notifications.

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
c <- s: No
c -> s: Is anything new?
c <- s: Yes, abc123 has sent you a message
```

This approach is known as {{< term spoll >}}.
It is highly inefficient in terms of bandwidth,
hundreds of requests might be made just to retrieve a single notification.

### Long Polling

To improve efficiency, the server side should hold requests for a **short duration** before responding,
this brief retention significantly reduces the number of unnecessary requests.
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

{{< term lpoll >}} is a traditional method for real-time notifications from the server side.
Since requests originate from the client side, {{< term lpoll >}} is well-suited for:

- **Decoupling** the server from the client side and increasing its availability.
- **Back pressure-aware clients**,
  allowing them to control their polling behavior autonomously,
  such as setting delays between polls or specifying the number of messages per poll.

{{< term lpoll >}} is best implemented as a **stateless** service.
Connections are short-lived; clients can conveniently switch to any server to crawl data from a shared store.
For example, the instances of a stateless service share and poll the same store.

```d2
grid-rows: 2

s: Service {
  grid-rows: 1
  horizontal-gap: 150
  i1: Instance 1 {
    class: server
  }
  s: Shared Store {
    class: db
  }
  i2: Instance 2 {
    class: server
  }
  i1 <- s: 3. Pull
  i2 -> s: 2. Update
}
c: "" {
  class: none
  grid-rows: 1
  horizontal-gap: 200
  c: Client {
    class: client
  }
  o: Another service {
    class: server
  }
}

c.c <- s.i1: Periodically pull {
  style.animated: true
}
c.o -> s.i2: 1. Send message to the client
```

Despite being more efficient than {{< term spoll >}}, {{< term lpoll >}} remains **resource-intensive**,
often generating many redundant requests before retrieving any actual piece of data.

Furthermore, it doesn't fully provide the **real-time capability**.
Since clients decide when to pull data, making messages can’t be transmitted immediately after their creations.

## WebSocket

This is a more modern technology than {{< term lpoll >}}.
In short, a {{< term ws >}} server maintains **long-lived connections**,
allowing both sides to actively exchange messages through these connections.

```d2

shape: sequence_diagram
c: Client {
    class: client
}
s: WebSocket {
    class: server
}
c <-> s: Establish a connection
s --> c: Server send message
c --> s: Client send message
c <-> s: "...More actions on the connection..."
```

Basically, {{< term ws >}} offers better performance than {{< term lpoll >}} by exchanging messages only when necessary,
resulting in lower latency and reduced bandwidth usage.
It's excellent for bidirectional and low-latency communication, e.g., gaming service, chat service.

Some critical drawbacks of {{< term ws >}} include:

- **Availability**: the server side depends on the client side and worsens its availability.
- **Resource utilization**: a {{< term ws >}} connection is long-lived and tied to a specific server,
  making it bad for resource utilization.
  For example, a client relentlessly interacts with a fixed server,
  making others slack;
  although it's better to distribute and share the load among them.

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

Do you think maintaining long-lived connections makes a service stateful?
The answer is no!

The communication protocol doesn't represent this property,
{{< term sf >}} or {{< term sl >}} is actually based on **how we implement** the service.
Get back to the chat example in the [previous topic]({{< ref "service-cluster#stateful-service" >}}),
we've mentioned it as a stateful service due to keeping users on different servers.
```d2
direction: right
c: Clients {
    ca: Client A {
      class: client
    }
    cb: Client B {
      class: client
    }
}
system: System {
    s1: Instance 1 {
      class: server
    }
    s2: Instance 2 {
      class: server
    }
}
c.ca <-> system.s1: Connecting
c.cb <-> system.s2: Connecting
```

Let's approach from a different angle.
Instead of sending messages directly between instances,
we let them periodically pull from a shared store.
Now, it's stateless!
All instances perform the same;
it doesn't matter which one a client connects to.

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
  i1 <- s: Pull {
    style.animated: true
  }
  i2 <- s: Pull {
    style.animated: true
  }
}
c2: Client 2 {
    class: client
}
c1 <- s.i1
c2 <- s.i2
```

### Use Cases

In fact, people tend to use {{< term ws >}} for **real-time notification**,
when messages are delivered immediately after their creation.
An indirect paradigm (e.g., messaging, shared store) is impossible for
the task as it creates brief delays;
a direct and stateful model is mandatory.

## Server-Sent Events

As the name suggests, **Server-Sent Events (SSE)** is a **half-duplex** protocol,
that means it maintains **long-lived connections** yet
only allowing data to be sent from the server side.

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
s --> c: Send message
```

Behind the scenes, {{< term sse >}} is built on top of the {{< term http  >}} protocol.
Thus, developing and maintaining an SSE application is simpler than {{< term ws >}},
as it can leverage existing {{< term http  >}} tools, such as connection management and caching.

Additionally, a unidirectional connection incurs **less overhead** than a full-duplex connection.
{{< term sse >}} is recommended if the application only needs to send data from the server side,
e.g., live scores, news websites.

Similar to {{< term ws >}}, {{< term sse >}} also introduces the same problems about low availability
and resource balancing.

## Google Remote Procedure Call (gRPC)

{{< term grpc >}} is a modern technology developed by `Google`,
enabling both bidirectional and unidirectional
communication over **Remote Procedure Call (RPC)** and {{< term http2 >}} protocol.

### Remote Procedure Call (RPC)

Normally, to call an {{< term http >}} endpoint,
an application must handle various details — such as the URI, headers, and parameters — to construct a proper **request string**.
While this approach offers flexibility, it can also be complex and prone to errors.

```http
GET /docs?name=README&team=dev HTTP/2
```

In contrast, {{< term rpc >}} is more structured,
requiring both the client and server to agree on a **shared contract** representing exposed endpoints.
This contract is usually built as a shared library,
making the interaction convenient, like working with local functions.

For example, the `Chat Service` exposes a `Chat` function;
This exposure is wrapped as a native shared library.

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

Another advantage of {{< term rpc >}} is **fast serialization**.
Typically, {{< term json >}} and {{< term xml >}} are commonly used to exchange data due to
their versatility across many use cases,
but their serialization process is slow because they are text-based and unstructured.
With a prepared definition, {{< term rpc >}} can optimize by pre-generating
a byte-based efficient serializer, such as [Protocol Buffers](https://protobuf.dev/overview/).

One drawback of {{< term rpc >}} is the **coupling** it creates between the server and client sides.
Any change in the contract requires redeployment on both ends.
Therefore, {{< term grpc >}} is rarely used for public-facing applications,
when a server may serve multiple types of clients.

### HTTP/2

{{< term http1 >}} establishes a connection between the server and client,
with all data transferred **in order** through this pipeline.
To enhance, {{< term http2 >}} divides a connection into **independent streams**,
allowing multiple requests and responses to be sent concurrently.
For example:

- In the {{< term http1 >}} context, `dog.png` is only downloaded after `index.html` has been fetched.

- In the {{< term http2 >}} context, the requests are sent simultaneously through `Stream 1` and `Stream 2`,
and the resources can be downloaded together.

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
  c -- s: 1. Loads page request {
    style.stroke-dash: 3
    style.bold: true
  }
  c --> s: Stream 1: Requests index.html
  c --> s: Stream 2: Requests dog.png
  c -- s: 2. Response {
    style.stroke-dash: 3
    style.bold: true
  }
  c <-- s: Stream 1: Responds index.html
  c <-- s: Stream 2: Responds dog.png
}
```

Behind the scenes, it still uses a single {{< term tcp >}} connection, with each message tagged by a **Stream ID**.
Messages with the same **Stream ID** are reassembled together, allowing multiple streams to run in parallel over one connection.

### Use Cases {id="grpc_use_cases"}

Back to {{< term grpc >}}, it's a protocol built on top of {{< term http2 >}} and {{< term rpc >}},
making it highly efficient for transmitting **parallel requests** simultaneously.

Similar to {{< term ws >}} and {{< term sse >}}, {{< term grpc >}} also maintains **long-lived connections**.
However, the more complex the network connection is, the more resources it consumes.
{{< term grpc >}} generally requires more computing power to manage parallel transmissions and assemble responses.
If the service doesn't require parallelism but valuing in-ordered actions,
consider using {{< term ws >}} or {{< term sse >}} instead.

## Webhook

{{< term wh >}} is an effective protocol for handling **long-running requests**.
Its concept is similar to a function pointer in programming.

The client side registers callbacks (usually a {{< term url >}}) with the server,
later invoked to notify responses.
For example, if a client registers an address,
whenever the server needs to notify the client,
it will request to `site.com/callback`.

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

This approach is ideal for tasks with **unpredictable execution time**,
helping avoid resource waste due to long waits.
For example, in payment processing,
when a client pays,
it may pass through multiple banking systems (possibly different countries),
and that can take a long time to complete.

{{< term wh >}} is an elegant protocol.
It's highly efficient for real-time notification by reducing server load,
as data is sent only when events occur,
without the need for long-lived connections or polling mechanisms.

### Use Cases {id="webhook_use_cases"}

Miserably, this is impractical for serving end users,
as they typically don't have a **public address** for the callback purpose.

Furthermore, in this model,
the server side becomes the originator, and its availability is negatively impacted.

In practice, this protocol is often used to support external services, like **Stripe Payments**,
where the system interacts with numerous uncontrolled clients.
In such cases, solutions like a live {{< term ws >}} server or {{< term lpoll >}} would consume significant resources.
For internal workloads, however, {{< term lpoll >}} is typically preferred, as it helps maintain a more available and resilient system.
