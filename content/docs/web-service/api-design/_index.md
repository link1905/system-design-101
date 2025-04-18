---
title: API Design
weight: 40
---


## API (Application Programming Interface)

{{< term api >}} abbreviates for Application Programming Interface,
it refers to shared contracts between processes to communicate over a network.

For example, we have two processes `Client` and `Server`

- When `Client` sends a command `Hello!`, `Server` will respond with `Hi!` but **nothing else**;
- When `Client` sends a command `Address?`, `Server` will send back its IP address.

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

The complete set of these commands and other rules (e.g., authorization) for
interacting with a process (or service) is called {{< term api >}}.
In this topic, we'll see how to produce this document effectively.

```yaml
api:
- command: Hello!
  response: Hi!
- command: Address?
  response: getAddress()
```

## REST (Representational State Transfer)

**API Design** is a critical step in **System Design**;
A system may contain a lot of components,
without a consistent and clear framework,
it's easy to become a [big ball of mud](https://www.geeksforgeeks.org/big-ball-of-mud-anti-pattern/).

{{< term rest >}} (Representational State Transfer)
is an **architectural style** first introduced by [Roy Fielding](https://en.wikipedia.org/wiki/Roy_Fielding) in 2000.
It comprises a set of high-level principles promoting scalability, simplicity, and compatibility.
It's **not tied** to any specific network protocols or frameworks, such as `HTTP` or `WebSocket`.
To clarify these principles, we will use [HTTP](../communication-protocols/) for examples
in the following sections.

## Resource

A {{< term rest >}} service is composed of a set of **resources**,
which expose its content to the outside world.

Resources themselves are not typically concrete,
but representing and transforming various forms of internal data:
database records, system files, site pages, etc.
For example:

- The `user` resource comes from the `user` SQL table.
- The `images` resource comes from local files.

```d2
%d2-import%
direction: up
s: REST service {
   u: "/user"
   i: "/images"
}
f: Local files {
   class: file
}
d: user_table {
   shape: sql_table
   id: string
   name: string
}
d <-> s.u
f <-> s.i
```

## 1. Statelessness

The first principle of {{< term rest >}} is [statelessness](../service-cluster.md#stateless-service).  
This means servers do not retain any session state between requests.

For example, a `user` resource transmitting credit offsets between the calls.
This **stateful** behavior will require the server to maintain local state.

```d2
%d2-import%
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

Instead, a **stateless** server should return complete records,
making requests are independent of each other.

```d2
%d2-import%
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

The second principle is `Uniform Interface`,
stating that a {{< term rest >}} service must provide a consistent and standardized way to interact with its resources.

### Resource Identifier

Each resource is uniquely identified using a **Uniform Resource Identifier (URI)** string.
In general, URIs are **structured hierarchically**, reflecting the relationships between resources, for example:

- A collection of resources, e.g., `/users`.
- A single resource, e.g., `/users/user_1234`.
- A nesting resource, e.g. `/users/user_1234/orders`.

### Resource Method

Resources not only return data but also allow manipulations.
When a client requests a resource, it must include the intended action, known as **method**.

For instance:

```http
# method /resource
GET /users
REMOVE /users/user_1234
CHANGE_NAME /users/user_1234
```

Therefore, in {{< term rest >}}, it is recommended to use **nouns** to name URIs.
We will avoid naming resources with verbs like `/change_username` or `/remove_user`.
Methods (or actions) should be attached to requests, not independent resources;

#### HTTP Methods

[HTTP methods](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods) are widely used to implement {{< term rest >}} methods:

- **GET**: Retrieve a resource.
- **POST**: Create a new resource.
- **DELETE**: Remove a resource.
- **PUT**: Completely update a resource (the client sends the entire updated resource).
- **PATCH**: Partially update a resource (the client sends only the fields that need updating).

> You may follow [this link](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods) to learn more about HTTP methods

Some methods, such as **POST**, **PUT**, and **PATCH**,
require a payload (or body) to execute.
For example, a request creating a new user needs to include the user details.

```HTTP
// Method /URI
POST /users
// Payload
{
    "name": "John Doe",
    "age": 18
}
```

##### Partial Update

In practice,
allowing clients to send a completely updated version of a resource by **PUT**
can be bandwidth-wasteful and potentially unsafe.

In many cases, updates are only limited to specific parts of a resource.
Two effective approaches for handling this are:

1. [HTTP PATCH](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/PATCH):
   Using **PATCH**, clients can update only the included fields. This is both efficient and simple:

```http
PATCH /users/1234
{
    "name": "My wonderful name"
}
```

2. **Sub resource**: For more complex logic, e.g.,
   when a user can only change their name after a specific time period.
   It’s better to separate the field as a new resource,
   this allows for finer control and more specific validation:

```http
PUT /users/1234/name
"My wonderful name"
```

#### Request Idempotency

Before wrapping this section, let's discuss a critical characteristic of requests - **Idempotency**.

1. **Idempotent**: A request is idempotent if perform it multiple times
   leaves the system unchanged after the first request.
    - **Read**: Does not manipulate resources, only retrieves data.
    - **Delete** and **Update**: Once a resource is deleted or updated, the following requests result in nothing.
      For example, a resource remains unchanged with the second update.

```d2
%d2-import%
shape: sequence_diagram
direction: right
c: Client {
    class: client
}
u: User Resource
u {
    f1: |||yaml
    User:
        Name: Johnny
    |||
}
c -> u: Update (Name = Doe)
u {
    f2: |||yaml
    User:
        Name: Doe
    |||
}
c -> u: Update (Name = Doe)
u {
    f3: |||yaml
    User:
        Name: Doe
    |||
}
```

2. **Non-idempotent** requests result in different system states when they're made multiple times
    - **Create**: Repeatedly creating a resource generates new and distinct data records.

```d2
%d2-import%
shape: sequence_diagram
direction: right
c: Client {
    class: client
}
u: User Resource
c -> u: Create
u {
    f1: |||yaml
    User:
        Name: Johnny
        CreatedAt: 00:00
    |||
}
c -> u: Create
u {
    f3: |||yaml
    User:
        Name: Johnny
        CreatedAt: 00:02
    |||
}
```

Identifying the idempotency of a request is crucial for ensuring **request safety**.

- **Non-idempotent requests** can often be retried freely, as repeating them does not compromise the system.
- **Idempotent requests**, on the other hand, should be safeguarded using a **deduplication mechanism** to avoid unintended consequences.

For example, in a payment request, a unique key is used to identify a transaction.
Even if the user retries the payment multiple times, only the first attempt is processed.

```d2
%d2-import%
shape: sequence_diagram
c: Client {
    class: client
}
p: Payment Service {
    class: client
}
c -> p: 1. Initiate a transaction
p -> c: 2. Respond with an idempotency key
c -> p: 2. Complete the transaction
p -> p: Processing...
c -> p: 3. Complete the transaction again (duplication)
c <- p: 4. Failed because the transaction is being processed {
    class: error-conn
}
```

The idempotency of a request depends on its **content**, not only its method.
For instance, an `update` request cancelling a `payment`
may also trigger the creation of a new `payment cancellation` resource,
making the overall action non-idempotent.

By understanding and implementing idempotency effectively,
we can build robust APIs that handle retries and duplicate requests gracefully.

## 3. Self-descriptive Messages

**Self-descriptive Messages** are a key principle in {{< term REST >}}, ensuring that messages (both requests and responses)
contain enough information to interpret and use their content,
e.g., **JSON**, **HTML page** or **PNG image**.

For example, a **JSON** message representing a user might look like this:

```json5
{
  // The content is in JSON format
  "type": "JSON",
  "content": {
    "id": 1234,
    "name": "John Doe"
  }
}
```

> It's pretty silly to retrieve a JSON indicator inside a JSON object.
> Actually, HTTP frameworks embed the type field in the **header** section, which contains plain texts.

### Content Negotiation

**Content Negotiation** is a mechanism that allows the client and server side to agree on the format of a resource.
It enables the server to serve different representations of the same resource,
while clients can favor their preferred format.

HTTP frameworks process content negotiation through:

- **Accept** header in requests: Clients indicate their preferred formats.
- **Content-Type** header in responses: Specifies how to process the response.

For example, a `user` resource can conveniently be served as either **JSON** or **XML** data.
> **application/json** (JSON) or **text/xml** (XML) are HTTP conventions.  
> You may follow [this link](https://developer.mozilla.org/en-US/docs/Web/HTTP/MIME_types) to learn more about HTTP media types.

```d2
%d2-import%
shape: sequence_diagram
jc: JSON Client
p: "/users/1234"
xc: XML Client
jc -> p: "Accept: application/json" {
   style.bold: true
}
p -> jc: "Content-Type: application/json" 
jc {
   json: |||json
   {
       "id": 1234,
       "name": "John Doe"
   }
   |||
}
xc -> p: "Accept: text/xml" {
   style.bold: true
}
p -> xc: "Content-Type: text/xml" 
xc {
   xml: |||xml 
   <user>
       <id>1234</id>
       <name>John Doe</name>
   </user>
   |||
}
```

In a more complex use case, the `user` resource can be retrieved as:

- A simple version with minimal information to reduce computation and network bandwidth.
- A full representation with the most recent orders.

> **application/vnd** stands for a vendor-specific prefix in HTTP.  
> In practice, you may name whatever you like, but it should be consistent across resources.

```d2
%d2-import%
shape: sequence_diagram
jc: Simple Client
p: "/users/1234" {
    class: server
}
xc: Full Client
jc -> p: "Accept: application/vnd.user.simple+json" {
   style.bold: true
}
p -> jc
jc {
   json: |||json
   {
       "id": 1234,
       "name": "John Doe"
   }
   |||
}
xc -> p: "Accept: application/vnd.user.full+json" {
   style.bold: true
}
p -> xc
xc {
   xml: |||json 
   {
       "id": 1234,
       "name": "John Doe",
       "order": {
         "orderCount": 86,
         "orders": [
           {
               "orderId": 1,
               "totalPrice": 120,
               "state": "completed"
           },
           {
               "orderId": 2,
               "totalPrice": 100,
               "state": "pending"
           }
         ]
       }
   }
   |||
}
```

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

```json5
//GET /users/1234
{
  "id": 1234,
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

```json5
//GET /users/1234/orders
[
  {
    "orderId": 1,
    "totalPrice": 120,
    "links": [
      {
        "rel": "self",
        "link": "/orders/1",
        "method": "GET",
        "description": "Get the order itself"
      },
      {
        "rel": "cancel",
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

{{< term hate >}} is considered the most challenging aspect of {{< term rest >}}.  
Many systems opt to **hardcode** links on the client side to simplify development, as {{< term hate >}} seems like an overhead.  
Moreover, hypermedia links largely extend the bandwidth usage of responses.

For my part, I've hardly implemented {{< term hate >}},
except for certain convenient tasks such as pagination or linking to the detailed version of a resource.

Let's say we have a server serving orders at `/users/{userId}/orders`.
One day, it moves the resource to another section, e.g. `/orders/{userId}`.
If the client side uses responded links, {{< term hate >}} helps it prevent the disruption.
But more curiosities come to our mind

- If the server tries to change the resource's structure,
  the client is also corrupted and needs to adapt.
- How can clients directly access a resource?
  Let's say we create an entrance endpoint (e.g. `/index`) listing all interfaces.
  If they want to access a sub resource, how do they know which is its parent?
  Moreover, it's extremely wasteful to go a long way, handle multiple responses to just reach a single resource.

At the end of the day,
I still need a document fully reflecting the running APIs.
Thus, I've hardly seen the real influence of {{< term hate >}} and
frequently ignored it.

I see that {{< term hate >}} is useful for **Server-side Rendering (SSR)**,
when the server side controls and returns complete views (e.g. `HTML` pages).
But that couples the server and client sides,
it can be problematic when the backend serves different client types.

## API Versioning

{{< term apiv >}} is the practice of managing changes to an API without breaking existing clients.  
Clients can choose the version that suits them, enabling the server to evolve independently.

Generally, a new version should be introduced if:

- Functionality is removed, breaking compatibility.
- Response or request structures are changed.
- Integrity mechanisms are modified, e.g., authentication or authorization.

There are several ways to version an API:

1. Modifying `URIs` directly, e.g., `v1/users`, `v2/users`.  
   This approach brings about visibility in the URL, making it easy to use and debug.
   However, it conceptually violates {{< term rest >}} principles since versions are not resources and should not be part of the
   `URI`.
2. Inserting directly versions into requests, e.g., through the `Accept` header.  
   This results in a clear and stable API hierarchy but more complex to implement and document.

### Version Deprecation

Managing multiple versions (`v1`, `v2`, `v3`, etc.) is challenging.  
It is crucial to ensure **backward capability** and enforce all versions to produce consistent results.
Moreover, it makes the codebase grow dramatically.
Therefore, we should announce deprecated versions and encourage consumers to upgrade,
including **deprecated points** (when to completely remove) and **migration guides**.