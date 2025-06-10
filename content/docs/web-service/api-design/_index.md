---
title: API Design
weight: 50
prev: load-balancer
next: api-pagination
---

## API (Application Programming Interface)

{{< term api >}} stands for **Application Programming Interface**,
which is a shared contract between processes that defines how they communicate over a network.

For example, consider two processes: `Client` and `Server`.

- When the `Client` sends a command `Hello!`, the `Server` responds with `Hi!` and nothing more.

- When the `Client` sends `Address?`, the `Server` responds with its `IP address`.

```d2
shape: sequence_diagram
s: Server {
    class: server
}
c: Client {
    class: client
}
c -> s: Hello!
s -> c: Hi!
c -> s: Address?
s -> c: 1.2.3.4
```

The complete set of these commands, together with other rules (such as authorization), constitutes an {{< term api >}}.
For example, here’s the API definition from the earlier example:

```yaml
api:
- command: Hello!
  response: Hi!
- command: Address?
  response: getAddress()
```

In this topic, we’ll explore how to design and document APIs effectively.

## REST (Representational State Transfer)

**API design** is a crucial part of system design.
Without a clear, consistent framework, a system with many components can quickly become a [big ball of mud](https://www.geeksforgeeks.org/big-ball-of-mud-anti-pattern/).

{{< term rest >}} (Representational State Transfer)
is an **architectural style** first introduced by [Roy Fielding](https://en.wikipedia.org/wiki/Roy_Fielding) in 2000.
It comprises a set of high-level principles promoting scalability, simplicity, and compatibility.

It's **not tied** to any specific protocol or framework, such as `HTTP` or `WebSocket`.
To clarify these principles, we will use [HTTP]({{< ref "communication-protocols" >}}) for the examples
in the following sections.

## Resource

A {{< term rest >}} service is made up of resources, which represent the data and services it exposes.

**Resources** represent database records, files, pages, or other internal data structures.
For example:

- The `user` resource comes from the `user` SQL table.
- The `images` resource comes from local files.

```d2
direction: right
f: Local files {
   class: file
}
d: user_table {
   shape: sql_table
   id: string
   name: string
}
s: REST service {
   u: "/user"
   i: "/images"
}
d <-> s.u
f <-> s.i
```

## 1. Statelessness

The first principle of {{< term rest >}} is [statelessness]({{< ref "service-cluster#stateless-service" >}}).
This means servers do not retain any session state between requests.

For example, if a user resource tracks a credit offset between calls, the server would have to maintain local state, making it **stateful**.

```d2
direction: right
c: Client {
    class: client
}
u: User Resource {
    class: server
}
s1: |||yaml
User:
    Id: 1234
    Credit: 1000
|||
s2: |||yaml
User:
    Id: 1234
    CreditOffset: -200
|||
c <- s1: 1st call
s1 <- u
c <- s2: 2nd call
s2 <- u
```

Instead, a stateless service returns complete records with each request, keeping interactions independent.

```d2
direction: right
c: Client {
    class: client
}
u: User Resource {
    class: server
}
s1: |||yaml
User:
    Id: 1234
    Credit: 1000
|||
s2: |||yaml
User:
    Id: 1234
    Credit: 800
|||
c <- s1: 1st call
s1 <- u
c <- s2: 2nd call
s2 <- u
```

## 2. Uniform Interface

The second principle is **Uniform Interface**.
{{< term rest >}} services should offer a consistent, standardized way for clients to interact with resources.

### Resource Identifier

Each resource is uniquely identified using a **Uniform Resource Identifier (URI)**.
In general, URIs are **structured hierarchically**, reflecting the relationships between resources, for example:

- A collection of resources, e.g., `/users`.
- A single resource, e.g., `/users/user_1234`.
- A nesting resource, e.g. `/users/user_1234/orders`.

### Resource Method

Resources allow both data retrieval and manipulation.
When a client requests a resource, it must include the intended action, known as **method**.

For instance:

```md
// method /resource_uri
LIST /users
GET_DATA /users/user_1234
REMOVE /users/user_1234
CHANGE_NAME /users/user_1234
```

In {{< term rest >}}, it’s recommended to use **nouns** for URIs, avoiding verbs like `/user/change_name`.
Actions should be expressed via request methods, not resource URIs.

#### HTTP Methods

[HTTP methods](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods) are widely used to implement {{< term rest >}} methods:

- **GET**: Retrieve a resource.
- **POST**: Create a new resource.
- **DELETE**: Remove a resource.
- **PUT**: Completely update a resource (the client sends the entire updated resource).
- **PATCH**: Partially update a resource (the client sends only the fields that need updating).

{{< callout type="info" >}}
You may follow [this link](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods) to learn more about HTTP methods.
{{< /callout >}}

Some methods, such as **POST**, **PUT**, and **PATCH**,
require a payload (or body) to execute.
For example, a request creating a new user needs to include the user details.

```http
POST /users HTTP/1.1

// Body
{
    "name": "John Doe",
    "age": 18
}
```

#### Partial Update

In practice,
allowing clients to send a completely updated version of a resource by **PUT**
can be bandwidth-wasteful and potentially unsafe.

In many cases, updates are only limited to specific parts of a resource.
Two effective approaches for handling this are:

1. [HTTP PATCH](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/PATCH):
   Using **PATCH**, clients can update only the included fields. This is both efficient and simple:

    ```http
    PATCH /users/1234 HTTP/1.1

    {
        "name": "My wonderful name"
    }
    ```

2. **Sub resource**: For more complex logic, e.g.,
   *a user can only change their name after a specific time period*.
   It’s better to separate the field as a new resource,
   this allows for finer control and more specific validation:

    ```http
    PUT /users/1234/name HTTP/1.1

    "My wonderful name"
    ```

#### Request Idempotency

Before wrapping this section, let's discuss a critical characteristic of requests - **Idempotency**.

1. **Idempotent**: A request is idempotent if perform it multiple times
   leaves the system unchanged after the first request, including:
    - **Read**: Does not manipulate resources, only retrieves data.
    - **Delete** and **Update**: Once a resource is deleted or updated, the following requests result in nothing.
      For example, a resource remains unchanged with the second update.

```d2
shape: sequence_diagram
direction: right
c: Client {
    class: client
}
u: "/users/1234"
u {
    f1: |||yaml
    Name: John
    |||
}
c -> u: Update Name = Doe
u {
    f2: |||yaml
    Name: Doe
    |||
}
c -> u: Update Name = Doe
u {
    f3: |||yaml
    Name: Doe
    |||
}
```

2. **Non-idempotent** requests result in different system states when they're made multiple times.
    - **Create**: Repeatedly creating a resource generates new and distinct data records.

```d2
shape: sequence_diagram
direction: right
c: Client {
    class: client
}
u: "/users"
c -> u: Create
u {
    "Name: Johnny, CreatedAt: 00:00"
}
c -> u: Create
u {
    "Name: Johnny, CreatedAt: 00:02"
}
```

Identifying the idempotency of a request is crucial for ensuring **request safety**.

- **Non-idempotent requests** can often be retried freely, as repeating them does not compromise the system.
- **Idempotent requests**, on the other hand, should be safeguarded using a **deduplication mechanism** to avoid unintended consequences.

For example, in a payment request, a unique key is used to identify a transaction.
Even if the user retries the payment multiple times, only the first attempt is processed.

```d2
shape: sequence_diagram
c: Client {
    class: client
}
p: Payment Service
c -> p: 1. Initiate a transaction
p {
    "Tran123: New"
}
p -> c: "Tran123" {
    style.bold: true
}
c -> p: "2. Process 'Tran123'"
p {
    "Tran123: Processing"
}
p -> p: Processing...
c -> p: 3. Process the transaction again (duplication) {
    style.bold: true
}
c <- p: Failed because the transaction is being processed {
    class: error-conn
}
```

The idempotency of a request depends on its **effect**, not just the method.
For example, an `update` request that cancels a `payment` might also create a new `payment cancellation` record. In this case, the overall action is no longer idempotent, since repeating the same request would generate additional resources.

By carefully understanding and designing for idempotency, we can build robust APIs that handle retries and duplicate requests gracefully, improving both reliability and client experience.

## 3. Self-descriptive Message

**Self-descriptive Message** is a key principle in {{< term rest >}},
ensuring that all messages (both requests and responses) contain enough information to interpret and use their content.

For example, a message representing a user might look like this.
The plain-text indicator guides how to read the **JSON** payload.

```
// Indicator
TYPE: JSON

// Payload
{
    "id": 1234,
    "name": "John Doe"
}
```

### Content Negotiation

**Content Negotiation** is a mechanism that allows the client and server side to agree on the format of a resource.
It enables the server to serve different representations of the same resource,
while clients can favor their preferred format.

{{< term http >}} frameworks process content negotiation through:

- **Accept** header in requests: Clients indicate their preferred formats.
- **Content-Type** header in responses: Specifies how to process the response.

For example, a `user` resource can conveniently be served as either **JSON** or **XML** data.

```d2
shape: sequence_diagram
jc: JSON Client
p: "/users/1234" {
    class: server
}
xc: XML Client
jc -> p: "Accept: application/json" {
   style.bold: true
}
p -> jc: "Content-Type: application/json"
jc {
   '{ "id": 1234 }'
}
xc -> p: "Accept: text/xml" {
   style.bold: true
}
p -> xc: "Content-Type: text/xml"
xc {
   "<user><id>1234</id></user>"
}
```

{{< callout type="info" >}}
**application/json** (JSON) or **text/xml** (XML) are HTTP conventions.
You may follow [this link](https://developer.mozilla.org/en-US/docs/Web/HTTP/MIME_types) to learn more about HTTP media types.
{{< /callout >}}

In a more complex use case, the `user` resource can be retrieved as:

- A simple version with minimal information to reduce computation and network bandwidth.
- A full representation with the most recent orders.

```d2
shape: sequence_diagram
jc: Simple Client
p: "/users/1234" {
    class: server
}
fc: Full Client
jc -> p: "Accept: application/vnd.user.simple+json" {
   style.bold: true
}
p -> jc
jc {
   '{ "name": "John Doe" }'
}
fc -> p: "Accept: application/vnd.user.full+json" {
   style.bold: true
}
p -> fc
fc {
    '{ "name": "John Doe", "order": { "orderCount": 86, "recentOrders": [ ] } }'
}
```

{{< callout type="info" >}}
**application/vnd** stands for a vendor-specific prefix in HTTP.
In practice, you may name whatever you like, but it should be consistent across resources.
{{< /callout >}}

Conveniently, we don’t need to create multiple resources for different shapes,
as it can make the server unnecessarily complex.
This capability can be also leveraged for [API Versioning](#api-versioning) in a later section.

## 4. Hypermedia As The Engine of Application State (HATEOAS)

{{< term hate >}} is a key principle of {{< term rest >}}.
Initially, the client needs minimal knowledge about the server,
{{< term hate >}} suggests that the server can dynamically guide clients move between related
resources based on the **hypermedia links** included in responses.

### Hypermedia links

For example, a user's orders might only contain the total number of orders with a link.
The user can then follow the link to retrieve the actual orders.

**GET /users/1234**:

```json
{
  "name": "John Doe",
  "order": {
    "orderCount": 86,
    // Hypermedia link
    "orders": {
      "link": "/users/1/orders",
      "method": "GET",
      "description": "Get all orders"
    }
  }
}
```

Accessing orders at `/users/1/orders`, each order contains additional information to further navigate the client to
get the detailed information.

**GET /user/1234/orders**:

```json
[
  {
    "orderId": 1,
    "totalPrice": 120,
    "links": [
      {
        "link": "/orders/1",
        "method": "GET",
        "description": "Get the order itself"
      },
      {
        "link": "/orders/1/cancellation",
        "method": "POST",
        "description": "Cancel the order"
      }
    ]
  }
]
```

Resources contain hypermedia links that can be followed to transition the application **from state to state**.
{{< term hate >}} makes the system more robust and self-discoverable,
meaning clients don't need to hardcode knowledge of available endpoints;
they are steadily guided by the backend.

### HATEOAS Or Not?

{{< term hate >}} is often considered the most challenging aspect of {{< term rest >}}.
Many systems choose to **hardcode** links on the client side to simplify development, viewing {{< term hate >}} as unnecessary overhead. Additionally, hypermedia links can noticeably increase the bandwidth consumption of responses.

Personally, I’ve rarely implemented {{< term hate >}}, except for certain convenient scenarios like pagination or linking to a detailed version of a resource.

For example, suppose we have a service providing `order` resource at `/users/{userId}/orders`.
If one day, the resource is moved to `/orders/{userId}`, a client relying on response-provided links would remain unaffected,
this is where {{< term hate >}} can help prevent disruptions.

However, this approach raises some concerns:

- If the server changes the structure of a resource, the client might still break and require adjustments.
- How do clients directly access a specific resource?
  Imagine creating an entry endpoint (e.g. `/index`) listing all available interfaces.
  If a client needs to reach a sub-resource, how would it determine its parent?
  It would be inefficient to traverse multiple layers and handle several responses just to reach a single resource.

Ultimately, I still rely on having a documented, up-to-date description of the active APIs.
For this reason, I’ve rarely witnessed the practical benefits of {{< term hate >}} and often choose to ignore it.

{{< term hate >}} may make sense in **Server-side Rendering (SSR)** scenarios, where the server fully controls and returns complete views (like {{< term html >}} pages).
But this tightly couples the server and client, which can become problematic when the backend needs to serve different types of clients.

## API Versioning

{{< term apiv >}} is the practice of managing changes to an API without breaking existing clients.
Clients can choose the version that suits them, enabling the server to evolve independently.

Generally, a new version should be introduced if:

- Functionality is removed, breaking compatibility.
- Response or request structures are changed.
- Integrity mechanisms are modified, e.g., authentication or authorization.

There are several ways to version an API:

1. Modifying {{< term uri >}} directly, e.g., `v1/users`, `v2/users`.
   This approach brings about visibility in the URL, making it easy to use and debug.
   However, it conceptually violates {{< term rest >}} principles since versions are not resources and should not be part of the {{< term uri >}}.
2. Inserting directly versions into requests, e.g., through the `Accept` header.
   This results in a clear and stable API hierarchy but more complex to implement and document.

### Version Deprecation

Managing multiple versions (`v1`, `v2`, `v3`, etc.) is challenging.
It is crucial to ensure **backward capability** and enforce all versions to produce consistent results.
Moreover, it makes the codebase grow dramatically.
Therefore, we should announce deprecated versions and encourage consumers to upgrade to latest versions,
including **deprecated points** (when to completely remove) and **migration guides**.
