---
title: Identity And Access Management (IAM)
weight: 20
---

We will now explore **Identity and Access Management (IAM)**, a fundamental and indispensable protective component for any system.

## Identity

Let's begin with the first component of an **IAM** system: **Identity**, also known as **Authentication**.

### User Session

The most straightforward method for implementing an identity system is through **User Sessions**.

Upon successful user sign-in, the system generates a temporary session code and sends it back to the user.
This code can then be attached to subsequent requests, allowing the user to access the system without repeatedly entering their credentials.

```d2
shape: sequence_diagram

c: Client {
  class: client
}
s: Identity Service {
  class: server
}

c -> s: Sign in with <username, password>
s -> c: Respond with a session code
c -> s: Request with the code
```

A stateful service can store user sessions directly in its local memory.

However, this approach presents challenges in a distributed environment.
For instance, if a client logs in through `Instance 1`, `Instance 2` will not recognize the client or be able to validate the session,
as it lacks any record of the session code.

```d2
direction: right

s: Identity Service {
  s1: Instance 1 {
    class: server
  }
  s2: Instance 2 {
    class: server
  }
}
c: Client {
  class: client
}

c -> s.s1: 1. Sign in
s.s1 -> c: 2. Respond with session code
c -> s.s2: 3. Fail to request with the code {
  class: error-conn
}
```

To develop a stateless authentication system,
sessions should be stored in a shared data store that all service instances can access.

With this setup, any server can validate a user's session code by querying the shared store.

```d2
direction: right

s: Identity Service {
  s1: Instance 1 {
    class: server
  }
  s2: Instance 2 {
    class: server
  }
  ss: Session Store {
    class: cache
  }
  s1 <-> ss
  s2 <-> ss
}
```

Nevertheless, in distributed environments, this strategy can lead to performance and availability issues,
as every session validation necessitates a query to the shared data store.

**User Sessions** are well-suited for stateful services where session management occurs locally.
A common example is in gaming systems,
where users often connect to a specific sticky instance that is responsible for local session validation.

### JSON Web Token (JWT)

**JSON Web Token (JWT)** is a widely used method for building authentication systems.

At its core, a **JWT** is an **immutable** credential that represents secure information about a user or entity.

A typical JWT looks like this `eyJhbGc.eyJzdWIiO.cThIIoDv`,
where the dots (`.`) separate three encoded sections: **header**, **payload**, and **signature**.

Removing the dots, you get three distinct encoded parts:

```json
{
  "header": "eyJhbGc",
  "payload": "eyJzdWIiO",
  "signature": "cThIIoDv"
}
```

Each section is encoded using [Base64](https://en.wikipedia.org/wiki/Base64),
starting from its original **JSON** representation:

```json
{
  // eyJhbGc
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  // eyJzdWIiO
  "payload": {
    "userId": "user1",
    "name": "John Doe",
    "role": "admin"
  },
  // cThIIoDv
  "signature": "cThIIoDv"
}
```

- **Header** specifies the cryptographic algorithm used to sign the token.
- **Payload** contains the customizable data, such as `role`, `username`, or other user-related information.
- **Signature** is generated from the payload and a **secret key**, securing the token against tampering.

#### Token Validation

Based on [signature validation]({{< ref "data-security#signature-validation" >}}),
**JWTs** rely on a **secret key** (secured server-side) for generating the **signature**.

Combined with the algorithm defined in the **header**, the signature is computed as:
`signature = alg(payload, secret key)`

Any modification to the **payload** results in a completely different **signature** due to cryptographic properties.

For example, when the system generates `signature = HS256(payload, secret)`:

- Every time the token is validated, its **signature** is recalculated and compared with the one present in the token.

```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "id": "user1",
    "role": "user"
  },
  "signature": "M5MDIy"
}
```

If an attacker modifies the role from `user` to `admin`, a valid signature cannot be generated without the secret key.

- Using the original signature with the altered payload leads to a verification failure, as they no longer match.

```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "id": "user1",
    "role": "admin"
  },
  "signature": "M5MDIy" // INVALID: does not match modified payload, should be "UD511yc"
}
```

As long as the **secret key** remains protected and undisclosed,
external entities cannot forge or tamper with tokens.

This approach also addresses the availability aspect of **User Session**:

- By sharing the secret key with other services,
tokens can be validated locally—no need to query a central authenticator.
- For instance, the `Identity Service` can distribute the key to the `User Service`.

```d2
shape: sequence_diagram
c: Client {
    class: client
}
a: Identity Service {
    class: server
}
o: User Service {
    class: server
}

a -> o: Distributes the secret key {
    style.animated: true
    style.bold: true
}
c -> a: 1. Authenticate
a -> c: 2. Respond a token
c -> o: 3. Request
o -> o: 4. Use the distributed key to validate {
    style.bold: true
}
```

#### Asymmetric Signature

Distributing the secret key to consumer services is risky because it allows them to independently generate malicious tokens.
To maintain security, consumer services should **not** be permitted to create new tokens.

By leveraging [asymmetric encryption](Data-Protection.md#asymmetric-encryption),
this risk is addressed by separating the key into two distinct parts:

- **Signing (private) key**: Used solely for generating tokens.
- **Authentication (public) key**: Used for verifying tokens.

Integrating a [Key store]({{< ref "data-security#key-management" >}}) further enhances control:

- The signing key is accessible only to the `Identity Service`.
- The authentication key can be distributed freely for token verification.

```d2
shape: sequence_diagram
c: Client {
    class: client
}
a: Identity Service {
    class: server
}
o: User Service {
    class: server
}
k: Key Management Service {
    class: kms
}

k -> o: Distribute the authentication key {
    style.animated: true
    style.bold: true
}
k -> a: Distribute the signing key {
    style.animated: true
    style.bold: true
}
c -> a: 1. Authenticate
a -> a: 2. Use the signing key to generate a new token {
    style.bold: true
}
a -> c: 3. Respond with the token
c -> o: 4. Request
o -> o: 5. Use the authentication key to validate {
    style.bold: true
}
```

This approach aligns with the [zero trust security model](https://en.wikipedia.org/wiki/Zero_trust_security_model), ensuring least privileged access and robust security across all services.

#### Token Expiry

Since we do not store **JWTs**, there is no mechanism to **invalidate** issued tokens directly.
To limit potential damage if a token is compromised,
**access tokens** are designed to be **short-lived**—typically lasting only 5 to 10 minutes.

Token lifespan is managed through special fields in the token payload,
and authentication processes must check these timestamps during verification:

```json
{
  "header": ...,
  "payload": {
    // Issued At: Time the token was created
    "iat": "2025-01-01 13:00",
    // Expiration: Time the token becomes invalid
    "exp": "2025-01-01 13:05"
  },
  "signature": ...
}
```

Short token lifespans, however, can create a poor user experience by requiring frequent logins.
To address this, the **refresh token** mechanism is used.

{{< callout type="info" >}}
Some critical systems (such as banking) opt to require users to re-authenticate each time,
forgoing refresh tokens entirely to maximize security.
{{< /callout >}}

#### Refresh Token

After a user signs in,
a **long-lived** is generated and stored securely within the system,
for example, **Facebook** uses a lifespan of 60 days.
The refresh token enables users to obtain new access tokens without having to re-authenticate each time their token expires.

```d2
shape: sequence_diagram
c: Client {
    class: client
}
a: Identity Service {
    class: server
}
store: Refresh Token Store {
    class: cache
}

c -> a: Sign in
a -> store: Generate and save refresh token {
    style.bold: true
}
c <- a: Respond access token + refresh token
c -> a: Re-authenticate with the refresh token {
    style.bold: true
}
a <- store: Check the refresh token {
    style.bold: true
}
c <- a: Respond new access token
```

The advantages of **Refresh Tokens** include:

- Eliminates the need for users to repeatedly enter their password, improving usability.
- Allows for **revocation**: removing a refresh token from the store immediately invalidates it.

It’s important to note this does **not** reintroduce the availability problem of **User Session**.
The shared storage is used **only** for validating refresh tokens during re-authentication (not for business actions).
In other words, the availability of business services is unaffected by operations involving refresh token validation.

```d2
direction: right

c: Client {
  class: client
}

s: {
  grid-rows: 2
  class: none
  s: Business Service {
    class: server
  }
  a: Identity Service {
    class: server
  }
}

c -> s: Request with access token
c -> a: Issue/refresh access token
```


### Federated identity

Identity is a fundamental requirement that often functions similarly across various systems.

Instead of developing an identity mechanism from scratch,
systems can utilize a trusted identity solution.

This approach, known as the **Federated Identity Pattern**,
involves relying on an independent service to handle user identity.

#### Identity Provider

Consider a scenario where a system intends to use an **Identity Provider (IdP)**,
like **Google**, as its login method.

A naive implementation might involve the system acting as an intermediary,
forwarding user credentials directly to the identity provider.

```d2
direction: right
s: System {
    class: server
}
i: Google {
    class: google
}
c: Client {
    class: client
}

c -> s: "1. Send email and password"
s -> i: "2. Authenticate on behalf of the user"
```

This workflow is inherently insecure
because the application handling the credentials could potentially misuse them.
Consequently, identity providers typically do not permit dependent applications
to authenticate directly on behalf of users in this manner.

```d2
direction: right
s: System {
    class: server
}
i: Google {
    class: google
}
c: Client {
    class: client
}

c -> s: "1. Send email and password"
s -> s: "2. Steal the credential" {
  class: error-conn
}
s -> i: "3. Attempt to authenticate on behalf of the user"
```

#### ID Token

Basically, identity providers require users to **interact directly** with them for authentication,
rather than through an intermediary application.

Let's examine this process in detail.

- When a user chooses to sign in to our system using a third-party platform, such as **Google**.

- The system redirects the user to the identity provider (e.g., **Google**) to sign in.

- Upon successful authentication by the provider, an **ID Token**,
typically in the [JSON Web Token (JWT)](#json-web-token-jwt) format,
is issued and sent back to the user's browser or client application

```d2
shape: sequence_diagram
u: User {
  class: user
}
cb: System {
  class: server
}
g: Google {
  class: google
}

u -> cb: "1. Initiate Sign in with Google"
cb -> u: "2. Redirect"
u -> g: "3. Sign in"
g -> u: "4. Issue ID Token"
```

The user then transmits this **ID Token** to the system.
The system must then verify the token's authenticity to confirm it was issued by a legitimate **Google** account.
Additionally, the system can inspect the token's payload to extract user information,
such as `email`, `account id`, and other details.

```d2
shape: sequence_diagram
u: User {
  class: user
}
s: System {
  class: server
}
g: Google {
  class: google
}

u -> s: "1. Initiate Sign in with Google"
s -> u: "2. Redirect"
u -> g: "3. Sign in"
g -> u: "4. Issue ID Token"
u -> s: "5. Send ID Token"
s -> s: "6. Verify token and extract user information"
```

How does the system verify the token?
This verification relies on [Asymmetric Encryption](#asymmetric-signature) principles:

- The **identity provider** (e.g., Google)
securely holds a **private signing key** used to generate new tokens.
- The provider also distributes a corresponding **public authentication key** to the system,
enabling it to verify the authenticity of tokens autonomously.

```d2
g: "Identity Provider" {
  s: "Signing Key (Private)" {
    class: pri-key
  }
}
s: "System" {
  a: "Authentication Key (Public)" {
    class: pub-key
  }
}
g.s -> s.a: "Key Distribution"
```

## Authorization

Once a user's identity is confirmed,
the next critical step is to determine what actions they are permitted to perform and which resources they can access within the system.
This process is known as **Authorization**. This section will explore fundamental concepts and methods for implementing **Authorization**.

### Access Policy

A core component in managing permissions is the **Access Policy**.
An access policy is a set of rules assigned to a user
that explicitly defines what actions it is permitted or denied to perform on specific resources.

For example, we can create distinct policies:

- One policy allowing user `John` to write to `Technical` documents,
- Another policy allowing `John` to read `Business` documents.

```d2
u1: "John" {
  class: client
}
u2: "Doe" {
  class: client
}
p1: "Policy 1: Allow Write Technical Docs" {
  |||yaml
  target: Technical documents
  allow: write
  |||
}
p2: "Policy 2: Allow Read Business Docs" {
  |||yaml
  target: Business documents
  allow: read
  |||
}
p1 -> u1: "Assigned to"
p2 -> u1: "Assigned to"
```

#### Least Privilege Principle

The **Principle of Least Privilege** is a fundamental and widely adopted cybersecurity concept.
In essence,
this principle dictates that users should only be granted the minimum level of access to perform their designated tasks,
and no more.

A key tenet of this principle is that **access is denied by default**;
a user's action is only permitted if explicitly approved by at least one applicable policy.

### Role

Managing permissions on an individual user basis can become exceedingly complex.
To simplify this, **Roles** are introduced.
A role groups users who share similar responsibilities,
allowing them to inherit a common set of permissions.

For example, all users assigned the `Developer` role might be granted permission to read `Technical` documents.

```d2
r: "Developer Role" {
  u1: "John" {
    class: client
  }
  u2: "Doe" {
    class: client
  }
}
p1: "Permission 1: Allow Read Technical Docs" {
  |||yaml
  target: Technical documents
  action: read
  |||
}
p1 -> r: "Assigned to"
```

A crucial aspect related to the Principle of **Least Privilege**.
If multiple policies apply, an explicit **denial policy** overrides any allow policies.
This ensures that specific restrictions can be enforced even if a user belongs to a role that generally grants broader access.

For instance, while the `Developer` role (`r`) might be granted permission (`p1`) to read `Technical` documents,
an explicit reject policy (`p2`) assigned specifically to user `John` (`r.u1`)
for those same documents will prevent him from performing that action.

```d2
r: "Developer Role" {
  u1: "John" {
    class: client
  }
  u2: "Doe" {
    class: client
  }
}
p1: "Permission 1: Allow Read Technical Docs (Role)" {
  |||yaml
  target: Technical documents
  action: read
  |||
}
p2: "Permission 2: Deny Read Technical Docs (John)" {
  |||yaml
  target: Technical documents
  action: reject
  |||
}
p1 -> r: "Assigned to Role"
p2 -> r.u1: "Assigned to John (Overrides Role Permission)"
```


#### Role Explosion

While roles simplify permission management,
relying solely on them can lead to a common issue known as **role explosion**.

Consider a scenario with `Technical` documents shared among `Developers`:
- `Lead Developers` require full access to them.
- Regular `Developers` can read and write them.
- `Intern Developers` should only be able to read them.

If a single, overly permissive `Developer` role is created,
it could lead to permission escalation.
The alternative is to create distinct roles like `Developer-Lead`, `Developer-Regular`, and `Developer-Intern`, and assign developers to these new roles accordingly.

As the organization grows and more variations are needed based on departments,
projects, or teams, the number of roles can multiply rapidly.
This proliferation results in significant management overhead.

### Attribute

To achieve finer-grained access control,
instead of relying solely on static role assignments,
we can tag users and resources with **attributes** that are then used in authorization decisions.

In our developer scenario, users within the `Developer` role can be tagged with attributes such as their `position`.
Thus, access policies become more dynamic by checking user attributes to make the final authorization decision.

```d2
r: "Developer Role" {
  u1: "John (Lead)" {
    c: |||yaml
    position: Lead
    |||
  }
  u2: "Doe (Intern)" {
    c: |||yaml
    position: Intern
    |||
  }
}
p1: "Policy 1: Full Access for Leads" {
  |||yaml
  target: Technical documents
  require:
    position: Lead
  action: full
  |||
}
p2: "Policy 2: Read Access for Interns" {
  |||yaml
  target: Technical documents
  require:
    position: Intern
  action: read
  |||
}
p3: "Policy 3: Write Access for Interns" {
  |||yaml
  target: Technical documents
  require:
    position: Intern
  action: write
  |||
}
p1 -> r: "Assigned to Role"
p2 -> r: "Assigned to Role"
p3 -> r: "Assigned to Role"
```

### Resourced-based Authorization

Thus far, our discussion has primarily focused on an administrative perspective where permissions are managed by creating policies and assigning them to identities (users or roles). This model is known as **Identity-Based Authorization**.

However, there are scenarios where regular users might need to control access to resources they own or manage, such as sharing their documents with specific colleagues.

Another authorization approach is **Resource-Based Authorization**.
In this model, permissions are attached directly to the resources themselves. The policies on the resource then specify which identities (users or roles) are granted or denied access.

For example, a specific `Technical document` could have a policy allowing the `Developer` role to `Read` it, and another policy granting user `Admin01` full control.

```d2
r: "Technical Documents (Resource)" {
  class: file
}
p1: "Permission 1: Developers can Read" {
  c: |||yaml
  target:
    type: role
    id: Developer
  action: Read
  |||
}
p2: "Permission 2: Admin01 has Full Access" {
  c: |||yaml
  target:
    type: user
    id: Admin01
  action: full
  |||
}
p1 -> r: "Attached to Resource"
p2 -> r: "Attached to Resource"
```

In practice, many robust authorization systems combine both **Identity-Based** and **Resource-Based** policies to offer comprehensive and flexible access control.

### Cross-system Resource Sharing

Consider a scenario where a system needs to share its resources with other applications.

For instance, imagine an application being developed that allows a user to upload files directly from their **Google Drive**.
In this situation, **Google Drive** must authorize the application to perform actions on the user's behalf.

```d2
u: User {
  class: client
}
s: Application {
  class: server
}
g: GoogleDrive {
  class: google
}
u -> s: Request to upload a file from Drive
s -> g: Access the file
```

Similar to the principles discussed in the [Identity Provider](#identity-provider) section,
it's crucial that the resource service (in this case, **Google Drive**)
**does not** permit the application to store the user's credentials,
such as passwords or comprehensive access tokens.
These credentials fully represent the user and could be misused to perform **any action** as that user.

Instead, a more secure approach involves creating distinct credentials specifically for the application.
These credentials are limited to the permissions explicitly granted by the user.

#### OAuth2.0

**OAuth2.0** is an authorization framework that allows third-party applications
to access user resources hosted on a service.

##### Basic Flow

The process typically begins when the application redirects the user to the resource service, such as **Google Drive**.
Here, the user signs in and grants specific permissions to the application.
Following this, the resource service issues an **Access Token** back to the user, which is limited to the granted permissions.

```d2
shape: sequence_diagram
u: User {
  class: client
}
s: Application {
  class: server
}
g: GoogleDrive {
  class: google
}
u -> s: Request
s -> u: Redirect to GoogleDrive
u -> g: Sign in and consent permissions
g -> u: Respond a limited access token {
  style.bold: true
}
```

The user then sends this token to the application. With this access token, the application can now access the resources shared by the user on the resource service.

```d2
shape: sequence_diagram
u: User {
  class: client
}
s: Application {
  class: server
}
g: GoogleDrive {
  class: google
}
u -> s: Request
s -> u: Redirect to GoogleDrive
u -> g: Sign in and consent permissions
g -> u: Respond an access token
u -> s: Send the token
s -> g: Access resources with the token
```

This outlines the fundamental flow of **OAuth 2.0**.
However, a potential security concern arises in this simplified model.
If the access token is sent directly to the frontend (e.g., the user's browser or device),
but the backend system is the actual entity that needs to access the resource,
the token becomes vulnerable. It should ideally be obtained and managed securely by the backend system to prevent exploitation.

```d2
shape: sequence_diagram
h: Hacker {
  class: hacker
}
u: User (Frontend) {
  class: client
}
s: Application (Backend) {
  class: server
}
g: GoogleDrive {
  class: google
}
g -> u: Respond an access token
h -> u: Steal the token here {
  class: error-conn
}
u -> s: Send the token
```

##### Authorization Code Grant

To mitigate the aforementioned security risk, the OAuth 2.0 framework introduces the concept of **authorization code grant**.
Instead of the resource service directly sending an access token to the user's browser, it sends back a temporary **Authorization Code**.

```d2
shape: sequence_diagram
u: User {
  class: client
}
s: Application {
  class: server
}
g: GoogleDrive {
  class: google
}
u -> s: Request
s -> u: Redirect to GoogleDrive
u -> g: Sign in and consent permissions
g -> u: Respond an authorization code {
  style.bold: true
}
```

The user then transmits this authorization code to the application (its backend system).
To gain access to the resources,
the application's backend must first use this code to exchange it for an actual
**Access Token** directly with the resource service.
This exchange happens server-to-server, away from the less secure frontend environment.

```d2
shape: sequence_diagram
u: User {
  class: client
}
s: Application {
  class: server
}
g: GoogleDrive {
  class: google
}
u -> s: Request
s -> u: Redirect to GoogleDrive
u -> g: Sign in and consent permissions
g -> u: Responds an authorization code
u -> s: Send the code
s -> g: Exchange an access token with the code {
  style.bold: true
}
```

With this flow, attackers who might intercept the communication between the user and the application would only obtain the short-lived authorization code,
not the actual access token.
However, this alone does not entirely solve the problem.
If attackers obtains the authorization code,
they could potentially still use it to exchange an access token themselves.
Therefore, this exchange mechanism is most effective when the entity exchanging the code is a **trusted target**.

##### Trusted Target

To ensure that only the legitimate application can exchange the authorization code for an access token,
the application must first register itself with the resource service.
This registration typically involves:

- Assigning a unique **Client ID** to the application.
- Providing the application with a **Client Secret**, which acts as a password for the application itself, used during the token exchange phase.

```d2
g: GoogleDrive {
  c: |||yaml
  app1:
   id: 123
   secret: 111222
  app2:
   id: 234
   secret: 222333
  |||
}
a1: Application 1 {
  c: |||yaml
  id: 123
  secret: 111222
  |||
}
a2: Application 2 {
  c: |||yaml
  id: 234
  secret: 222333
  |||
}
a1 -> g: Register
a2 -> g: Register
```

When redirecting the user to the resource service for authorization,
the application must include its **Client ID** in the redirection request (e.g., `/auth?client_id=123`).
This ensures that the resource service issues an authorization code specifically for that application.

During the access token exchange,
the application authenticates itself to the resource service using its **Client ID** and **Client Secret**,
along with the authorization code.
The resource service verifies all these components before issuing the access token.

```d2
shape: sequence_diagram
u: User {
  class: client
}
s: Application 1 {
  class: server
}
g: GoogleDrive {
  class: google
}
u -> s: Request
s -> u: "Redirect to /auth?client_id=123"
u -> g: Sign in and consent permissions
g -> u: Respond an exchange code
u -> s: Send the code
s -> g: Exchange an access token with (authorization code + client ID + client secret) {
  style.bold: true
}
g -> s: Respond an access token
```

This enhanced process ensures that only the registered application, possessing the correct Client ID and Client Secret, can successfully exchange the authorization code for an access token. Consequently, even if an authorization code is intercepted, it cannot be exploited by an unauthorized party.
