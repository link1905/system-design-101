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

- When the `Client` sends the command `Hello!`, the `Server` responds with `Hi!` and nothing else.
- When the `Client` sends `Address?`, the `Server` responds with its `IP address`.

The complete set of these commands, together with other rules such as authorization, constitutes an {{< term api >}}.
For example, here is the API definition from the previous example:

```yaml
api:
- command: Hello!
  response: Hi!
- command: Address?
  response: getAddress()
```

In this topic, we'll explore how to design and document APIs effectively.

## REST (Representational State Transfer)

**API design** is a crucial part of system design.
Without a clear and consistent framework, a system with many components can quickly become a [big ball of mud](https://www.geeksforgeeks.org/big-ball-of-mud-anti-pattern/).

{{< term rest >}} (Representational State Transfer)
is an *architectural style* first introduced by [Roy Fielding](https://en.wikipedia.org/wiki/Roy_Fielding) in 2000.
It consists of a set of high-level principles that promote scalability, simplicity, and compatibility.

It is *not tied* to any specific protocol or framework, such as `HTTP` or `WebSocket`.
To illustrate these principles, we will use [HTTP]({{< ref "communication-protocols" >}})
in the examples throughout the following sections.

## Resource

A {{< term rest >}} service consists of resources that represent the data and services it exposes.

**Resources** can represent database records, files, pages, or other internal data structures.
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

The first principle of {{< term rest >}} is [Statelessness]({{% ref "service-cluster#stateless-service" %}}).

{{% include "stateless-service" %}}

## 2. Uniform Interface

The second principle is **Uniform Interface**.
{{< term rest >}} services should provide a consistent, standardized way for clients to interact with resources.

### Resource Identifier

Each resource is uniquely identified using a **Uniform Resource Identifier (URI)**.
In general, URIs are **structured hierarchically** to reflect relationships between resources. For example:

- A collection of resources, e.g., `/users`.
- A single resource, e.g., `/users/user_1234`.
- A nested resource, e.g., `/users/user_1234/orders`.

### Resource Method

Resources support both data retrieval and manipulation.
When a client requests a resource, it must specify the intended action, known as a **method**.

For instance:

```md
// method /resource_uri
LIST /users
GET_DATA /users/user_1234
REMOVE /users/user_1234
CHANGE_NAME /users/user_1234
```

In {{< term rest >}}, it is recommended to use *nouns* for URIs and avoid verbs such as `/user/change_name`.
Actions should be expressed through request methods rather than resource URIs.

#### HTTP Methods

[HTTP methods](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods) are widely used to implement {{< term rest >}} methods:

- **GET**: Retrieve a resource.
- **POST**: Create a new resource.
- **DELETE**: Remove a resource.
- **PUT**: Completely update a resource, with the client sending the entire updated representation.
- **PATCH**: Partially update a resource, with the client sending only the fields that need to change.

{{< callout type="info" >}}
You may follow [this link](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods) to learn more about HTTP methods.
{{< /callout >}}

Some methods, such as **POST**, **PUT**, and **PATCH**,
require a payload, or body, to perform an operation.
For example, a request that creates a new user needs to include the user's details.

```http
POST /users HTTP/1.1

// Body
{
    "name": "John Doe",
    "age": 18
}
```

#### Partial Update

In practice, allowing clients to send a completely updated representation of a resource using **PUT**
can consume unnecessary bandwidth and may introduce additional risks.

In many cases, an update affects only specific parts of a resource.
Two effective approaches for handling partial updates are:

1. [HTTP PATCH](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/PATCH):
   With **PATCH**, clients can update only the fields included in the request.
   This approach is both simple and efficient:

    ```http
    PATCH /users/1234 HTTP/1.1

    {
        "name": "My wonderful name"
    }
    ```

2. **Subresource**: For more complex logic, such as when
   *a user can change their name only after a specific period*,
   it may be better to model the field as a separate resource.
   This approach allows finer control and more specific validation:

    ```http
    PUT /users/1234/name HTTP/1.1

    "My wonderful name"
    ```

#### Request Idempotency

Before concluding this section, let's discuss an important characteristic of requests: **Idempotency**.

1. **Idempotent**: A request is idempotent if performing it multiple times
   leaves the system in the same state as performing it once. Examples include:
    - **Read**: Does not modify resources and only retrieves data.
    - **Delete** and **Update**: Once a resource has been deleted or updated, repeating the same operation does not further change the resulting state.
    - For example, updating a user's name to `Doe` a second time produces the same state as the first update.

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

2. **Non-idempotent** requests can result in different system states when they are performed multiple times.
    - **Create**: Repeatedly creating a resource can generate new and distinct records.
    - For example, creating a new user named `Johnny` twice results in two separate records.

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

Understanding whether a request is idempotent is crucial for ensuring **request safety**.

- **Idempotent requests** can often be retried freely because repeating them does not compromise the system.
- **Non-idempotent requests**, on the other hand, should be protected with a *deduplication mechanism* to avoid unintended consequences.

For example, in a payment request, a unique key can be used to identify a transaction.
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
p -> c: "Return unique id Tran123" {
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
p -> p: "Tran123 is in-process"
c <- p: Failed because the transaction is being processed {
    class: error-conn
}
```

The idempotency of a request depends on its *effect*, not merely on the method used.
For example, an `update` request that cancels a `payment` might also create a new `payment cancellation` record.
In this case, the overall action is no longer idempotent because repeating the same request would generate additional resources.

By carefully understanding and designing for idempotency, we can build robust APIs that handle retries and duplicate requests gracefully, improving both reliability and the client experience.

## 3. Self-descriptive Message

A **Self-descriptive Message** is a key principle of {{< term rest >}},
ensuring that every message, whether a request or response, contains enough information to interpret and process its content.

For example, a message representing a user might look like this.
The plain-text indicator specifies how the **JSON** payload should be interpreted.

```text
// Indicator
TYPE: JSON

// Payload
{
    "id": 1234,
    "name": "John Doe"
}
```

### Content Negotiation

**Content Negotiation** is a mechanism that allows the client and server to agree on the representation format of a resource.
It enables the server to provide multiple representations of the same resource
while allowing clients to indicate their preferred format.

{{< term http >}} frameworks commonly implement content negotiation through:

- **Accept** header in requests: Clients indicate their preferred response formats.
- **Content-Type** header in responses: Specifies the format of the returned content and how it should be interpreted.

For example, a `user` resource can be served as either **JSON** or **XML**
based on the client's preference.

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
**application/json** (JSON) and **text/xml** (XML) are HTTP media types.
You may follow [this link](https://developer.mozilla.org/en-US/docs/Web/HTTP/MIME_types) to learn more about HTTP media types.
{{< /callout >}}

In a more complex use case, the `user` resource might be retrieved in different representations:

- A simple representation containing minimal information to reduce computation and network bandwidth.
- A full representation containing additional information, such as the user's most recent orders.

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
**application/vnd** is commonly used as a prefix for vendor-specific media types in HTTP.
In practice, you may define your own media types, but their naming should remain consistent across resources.
{{< /callout >}}

This approach conveniently avoids the need to create separate resources for different representations,
which could otherwise make the server unnecessarily complex.
The same capability can also be leveraged for [API Versioning](#api-versioning), discussed later.

## 4. Hypermedia As The Engine of Application State (HATEOAS)

{{< term hate >}} is a key principle of {{< term rest >}}.
Initially, the client needs only minimal knowledge of the server.
{{< term hate >}} proposes that the server dynamically guides clients between related
resources through **hypermedia links** included in responses.

### Hypermedia Links

For example, a user's representation might include only the total number of orders and a link to them.
The client can then follow that link to retrieve the actual orders.

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

When accessing the orders at `/users/1/orders`, each order can contain additional links
that guide the client toward more detailed information or available actions.

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

Resources contain hypermedia links that clients can follow to transition the application *from state to state*.
{{< term hate >}} makes a system more self-discoverable and can improve its adaptability,
because clients do not need to hardcode knowledge of every available endpoint;
instead, they are progressively guided by the backend.

### HATEOAS Or Not?

{{< term hate >}} is often considered one of the most challenging aspects of {{< term rest >}}.

{{< term hate >}} can work naturally in **Server-side Rendering (SSR)** scenarios,
where the server controls and returns complete views, such as {{< term html >}} pages,
and users navigate between them by following links.

Outside browser environments built around **JavaScript** and **HTML**,
such as backend services, mobile applications, and desktop applications,
consuming HATEOAS-style resources can be less natural and more difficult to implement.

Additionally, hypermedia links can noticeably increase response sizes and network bandwidth consumption.
For these reasons, clients often choose to *hardcode* API paths to simplify development
and instead rely on accurate, up-to-date API documentation.

## API Versioning

{{< term apiv >}} is the practice of managing changes to an API without breaking existing clients.
Clients can choose the version that suits their requirements, allowing the server to evolve independently.

Generally, a new version should be introduced when:

- Functionality is removed in a way that breaks compatibility.
- Request or response structures change incompatibly.
- Integrity mechanisms, such as authentication or authorization, are modified in incompatible ways.

There are several ways to version an API:

1. Modifying the {{< term uri >}} directly, e.g., `v1/users` and `v2/users`.
   This approach is widely used because the version is clearly visible in the URL,
   making the API straightforward to use, inspect, and debug.
   However, it can be viewed as conceptually inconsistent with {{< term rest >}} principles,
   since versions are not resources and therefore do not naturally belong in the {{< term uri >}}.

2. Specifying the version within the request, e.g., through the `Accept` header.
   This approach preserves a clean and stable resource hierarchy,
   but it can be more complex to implement, use, and document.

### Version Upgrading

Managing multiple versions such as `v1`, `v2`, and `v3` is challenging.
When releasing a new version, we need to maintain *backward compatibility*,
meaning the new service must continue supporting all versions that remain active.

```d2
direction: right
v1: API v1 {
    v1: "/v1"
}
v2: API v2 {
    grid-rows: 1
    v1: "/v1"
    v2: "/v2"
}
v1 -> v2: Upgraded
```

This can be difficult because all supported versions must continue producing consistent results.
Moreover, maintaining several versions can significantly increase the size and complexity of the codebase.

Therefore, deprecated versions should be announced clearly, and consumers should be encouraged to migrate to newer versions.
A deprecation plan should include *deprecation deadlines*, indicating when support will be removed,
and *migration guides*, explaining how consumers can upgrade.