---
title: Identity And Access Management (IAM)
weight: 20
---

We dive into the most basic and necessary protective component of every system — **Identity And Access Management (IAM)**.

## Authentication

We will get through the first component of an **IAM** system - **Identity** (or **Authentication**).

### User Session

The most straightforward implementation of an authentication system is **User Session**.
User session is nothing but a temporary code representing a **<username, password>** pair.
Users can leverage this code to avoid retyping credential information.

We can set up a stateful service by storing user sessions in local memory.
E.g., a client authenticates with `Server 1`, but `Server 2` knows nothing about this and fails to identify the client.

```d2
direction: right
s: Auth Service {
  s1: Auth Server 1 {
    class: server
  }
  s2: Auth Server 2 {
    class: server
  }
}
c: Client {
    class: client
}
c -> s.s1: 1. Sign in
c <- s.s1: 2. Respond session code
c -> s.s2: 3. Fail to request with the code {
  class: error-conn
}
```

To build a stateless authenticator, we migrate user sessions to a shared store.

```d2
shape: sequence_diagram
c: Client {
    class: client
}
s1: Auth Server 1 {
    class: server
}
s2: Auth Server 2 {
    class: server
}
store: Shared Store
c -> s1: 1. Sign in
s1 -> store: 2. Save new user session {
    style.bold: true
}
store {
  "session123: user1"
}
c <- s1: 3. "Respond 'session123'"
c -> s2: 4. "Request with 'session123'"
s2 <- store: 5. Verify with the shared store {
    style.bold: true
}
```

Despite the simplicity,
this approach is problematic in a distributed environment.
Decreased performance and availability, with every request,
we need to query the store to verify the request's session code.

```d2
direction: right
c: Client {
  class: client
}
s: Service {
  class: server
}
a: Authentication Service {
  class: server
}
c -> s: Request
s <-> a: Verify the session code
```

Programmatically, comparing between sessions (unique strings) is extremely fast.
It's beneficial for high-performed stateful services.
E.g., A game system forwards a user to a sticky server, and the user is **locally validated** in this server.

### JSON Web Token (JWT)

Nowadays, **JSON Web Token (JWT)** is a well-known approach for constructing an authentication system.

In short, **JWT** is simply an **immutable** credential.
You may see a JWT like this `eyJhbGc.eyJzdWIiO.cThIIoDv`, please note the dots.
Cutting off the dots, we get three encoded parts respectively: **header**, **payload** and **signature**.

```json
{
  "header": "eyJhbGc"
  "payload": "eyJzdWIiO",
  "signature": "cThIIoDv"
}
```

These parts are initially encoded with [Base64](https://en.wikipedia.org/wiki/Base64) from **JSON** values.

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

- **Payload** is the custom part of a token.
  We may use it to serve various purposes by wrapping user data, such as `role`, `username`...
- **Header** tells about the token's cryptographic algorithm.
- **Signature** is generated from the **payload** and used to validate the token.

#### Token Validation

Based on []({{< ref "data-security#signature-validation" >}}),
**JWT** requires a **secret key** (protected on the server side) for generating the **signature**.
Combined with the function defined in the **header** part, we can calculate `signature = alg(payload, secret key)`.

Thus, based on cryptographic instinct, once a token is created,
any changes in the **payload** lead to a completely different **signature**.

E.g. the system generates a `signature = HS256(payload, secret)`.
Whenever the token is verified, its **signature** will be recomputed to compare with the field in the token.

```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "id": "user1"
    "role": "user"
  },
  "signature": "M5MDIy"
}
```

An attacker tries to change the role from `user` to `admin`, resulting in a new signature.
However, he doesn't know the secret key, he cannot generate the valid signature.
If the old signature is kept, the **payload** is mismatched with its signature.

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
  "signature": "M5MDIy" // WRONG with the payload, it should be "UD511yc"
}
```

As long as the **secret** is protected and unrevealed,
external entities cannot modify or generate tokens maliciously.
Then, we resolved the availability problem of `User Session`.
We can distribute the secret key to other services to let them validate tokens without querying any authenticator.

E.g., `Auth Service` shares the secret key to `User Service`.

```d2
shape: sequence_diagram
c: Client {
    class: client
}
a: Auth Service {
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

Distributing the secret key to consumer services is dangerous,
as the key empowers them to generate malicious tokens independently.

We should not allow consumer services to generate new tokens.
Taking benefit of [asymmetric encryption](Data-Protection.md#asymmetric-encryption),
we separate the secret key into

- **Signing (private) key** generating tokens.
- **Authentication (public) key** verifying tokens.

Combined with a [Key store]({{< ref "data-security#key-management" >}})

- The signing key is only accessed by the `Authentication Service`.
- The authentication key is freely distributed.

```d2
shape: sequence_diagram
c: Client {
    class: client
}
a: Auth Service {
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
a -> a: 2. Use the singing key to generate a new token {
    style.bold: true
}
a -> c: 3. Respond the token
c -> o: 4. Request
o -> o: 5. Use the authentication key to validate {
    style.bold: true
}
```

This work is expanded to the [zero trust security model](https://en.wikipedia.org/wiki/Zero_trust_security_model),
ensuring least privileged access to resources.

#### Refresh Token

As we do not store **JWT**, there is no way to **invalidate** generated tokens.

In fact, tokens are **short-lived** (5-10 minutes) to mitigate malicious actions in case of token loss.
To do that, we create some fields to mark tokens' lifespan,
authentication processes needs to consider these timestamps on checking.

```json
{
  "header": ...,
  "payload": {
    // Issued at: Timestamp when the token was issued
    "iat": "2025-01-01 13:00",
    // Expiration: Timestamp when the token expires
    "exp": "2025-01-01 13:05"
  },
  "signature": ...
}
```

But it is bothersome for users to re-login frequently as tokens are quick to expire.
A concept of `Refresh Token` is used to ensure smooth experience.
The internal structure of a refresh token is similar to a normal JWT **header.payload.signature**.

{{< callout type="info" >}}
Some critical systems (e.g. banking) accept the drawback of requiring users to re-authenticate without any refresh mechanism.
{{< /callout >}}

However, refresh token is **long-lived** (e.g., **Facebook** uses the duration of 60 days) and persistently stored in the system.
After a successful login, the system returns an `access token` attached with a `refresh token`.
The client also needs to save the refresh token **locally**,
later on, it can use the token to issue new access tokens.

```d2
shape: sequence_diagram
c: Client {
    class: client
}
a: Auth Service {
    class: server
}
store: Refresh Token Store {
    class: cache
}
c -> a: 1. Sign in
a -> store: 2. Save refresh token {
    style.bold: true
}
c <- a: 3. Respond access token + refresh token
c -> a: 4. Access resource by the access token
c -> c: 5. The access token expires
c -> a: 6. Re-authenticate with the refresh token {
  style.bold: true
}
a <- store: 7. Check the refresh token {
    style.bold: true
}
c <- a: 8. Respond new access token
```

**Refresh Token** comes with some benefits:

- Allows re-authenticating without typing and sending password.
- Allows revocation which deletes a refresh token from the store to invalidate it.

You may be confused that we are moving back to the problem of `User Session`,
when we must verify credentials by an shared store.
Please note that, we only leverage refresh tokens for re-authentication, not business actions.
When an access token expires (the token itself contains enough information to check that), the client is expected to ask the `Authentication Service` for refreshing.
In other words, the availability of consumer services are not dependendly affected.

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
  a: Authentication Service {
    class: server
  }
}
c -> s: Make business requests
c -> a: Issue/refresh token
```

### Federated Identity

Authentication is a generic requirement, it frequently behaves the same in almost systems.
Instead of developing from scratch,
we may depend on a credible **Identity Provider (IdP)**, such as **Google**, or build a managed solution.
This pattern is called **Federated Identity Pattern**,
relying on an independent identity solution.

#### Identity Provider

For example, we want to rely on a social platform,
such as **Google** as the login method.
Naively, we can make the system act as a shim forwarding user credentials to the provider.

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

This workflow is dangerous, as applications can take advantage of user credentials in between.
Thus, an identity provider won't let dependent applications authenticate on behalf of users.

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
s -> i: "2. Authenticate on behalf of the user"
```

#### ID Token

Many frameworks were born to resolve this problem.
In general, they require users to **directly interact** with identity providers
instead of an intermediate application.

Let's see the procedure in details.
When a user wants to use a third-party platform to sign in to our system.
The process starts by forwarding the user to sign in with the identity provider,
e.g., **Google**.
After the user is authenticated successfully,
the provider will respond back to it an **ID Token** in form of [JWT](#json-web-token-jwt).

```d2
shape: sequence_diagram
u: User {
  class: user
}
cb: System {
  class: server
}
g: Google
  class: google
}
u -> cb: Sign in with Google
cb -> u: Forward to Google
u -> g: Sign in Google account
g -> u: Respond an ID token
```

The user sends the **ID token** to the system.
The system needs to check whether this is a valid token,
a valid token means this comes from a valid **Google** account.
The system aslo inspects the token to get further information in the token's payload, such as `email`, `account id`, etc.

```d2
shape: sequence_diagram
u: User {
  class: user
}
cb: System {
  class: server
}
g: Google
  class: google
}
u -> cb: Sign in with Google
cb -> u: Forward to Google
u -> g: Sign in Google account
g -> u: Respond an ID token
u -> s: Send the token
s -> s: Verify and use information from the token
```

How does the system can verify the token?
Based on [Asymmetric Encryption](#asymmetric-signature),

- The identity provider secretly hold a signing key for generating new tokens.
- Besides, the provider distributes an authentication key to the system for verifying tokens autonomously.

```d2
g: Google {
  s: Signing (Private) Key {
    class: pri-key
  }
}
s: System {
  a: Authentication (Public) Key {
    class: pub-key
  }
}
g.s -> s.a: Distribute
```

## Authorization

After a user is identified, we must determine its rights and capabilities in the system,
this step is called **Authorization**.
In this section, we'll see how to basically implement **Authorization**.

### Access Policy

We first view a component called **Access Policy**.
A policy is assigned to an entity to decide what it's able to perform.

For example,
we create a policy allowing user `John` to write `Technical` and read `Business` documents.

```d2
u1: "John" {
  class: client
}
u2: "Doe" {
  class: client
}
p1: Policy 1 {
  |||yaml
  target: Technical documents
  allow: write
  |||
}
p2: Policy 2 {
  |||yaml
  target: Business documents
  allow: read
  |||
}
p1 -> u1: Assigned to
p2 -> u1: Assigned to
```

#### Least Privilege Principle

**Least Privilege** is a cybersecurity principle that widely used.
In short, the principle ensures that users are granted enough permissions, no less, no more!
The first rule is denial is the default behaviour,
a user action must be explicitly approved through at least one policy.

### Role

However, there're can be a lot of permissions,
assigning for each of users is cumbersome.
We may group many users into a role to let them share the same permission set.

For example,
users with `Deverloper` role can read `Technical` documents.

```d2
r: Developer role {
  u1: "John" {
    class: client
  }
  u2: "Doe" {
    class: client
  }
}
p1: Permission 1 {
  |||yaml
  target: Technical documents
  action: read
  |||
}
p1 -> r: Assigned to
```

The second rule of **Least Privilege** is **any rejection** will abort the entire authorization process.
This helps restricts some special users in a role,
for example,
`Developers` are allowed to read `Technical` documents,
but `John` can't do that because of the explicit reject policy.

```d2
r: Developer role {
  u1: "John" {
    class: client
  }
  u2: "Doe" {
    class: client
  }
}
p1: Permission 1 {
  |||yaml
  target: Technical documents
  action: read
  |||
}
p2: Permission 2 {
  |||yaml
  target: Technical documents
  action: reject
  |||
}
p1 -> r: Assigned to
p2 -> r.u1: Assigned to
```

#### Role Explosion

Simply applying this paradigm will lead us to a problem called **role explosion**.

Let's say `Technical` documents shared with `Developers`.

- `Lead` developers have fully access to them.
- Regular `Developers` can read and write them.
- `Intern` developes can only read them.

If we create a shared powerful `Developer` role,
it probably leads to permission escalation.
Thus, we need to create `Developer`, `Developer-Lead`, `Developer-Intern` add the developers into new them respectively.
As the business grows with more aspects (departments, projects or teams), a lot of roles will be also needed,
resulting in management overhead.
It's often to see some roles will be established for few users.

### Attribute

To apply finer-grained access control,
we should tag users with attributes later helping in the authorization process.

In the previous example,
we may tag developers with their positions,
and create appropriate policies for the `Developer` role.
In the access policies, we need to check user attributes for the final decision.

```d2
r: Developer role {
  u1: "John" {
    c: |||yaml
    position: Lead
    |||
  }
  u2: "Doe" {
    c: |||yaml
    position: Intern
    |||
  }
}
p1: Permission 1 {
  |||yaml
  target: Technical documents
  require:
    position: Lead
  action: full
  |||
}
p2: Permission 2 {
  |||yaml
  target: Technical documents
  require:
    position: Intern
  action: read
  |||
}
p3: Permission 3 {
  |||yaml
  target: Technical documents
  require:
    position: Intern
  action: write
  |||
}
p1 -> r: Assigned to
p2 -> r: Assigned to
p3 -> r: Assigned to
```

### Resourced-based Authorization

Since the beginning,
we've only played the role of an administrator managing permissions.
We manage access by creating permissions and assign them to identities (users or roles),
this is called **Identity-based Authorization**.
Occasionally,
we want regular users to control access of their own resources or share resources with other users,
but this approach is unfriendly for them.

From a more granular angle,
we can implement authorization by assigning permissions to resources,
making the permissions' target is identities instead.
This is called **Resource-based Authorization**.

```d2
r: Technical documents {
  class: file
}
p1: Permission 1 {
  c: |||yaml
  target:
    type: role
    id: developer
  action: Read
  |||
}
p2: Permission 2 {
  c: |||yaml
  target:
    type: user
    id: admin01
  action: full
  |||
}
p1 -> r: Assigned to
p2 -> r: Assigned to
```

In fact, **Identity-based** and **Resource-based** authorization should be implemented together
to make the system more flexible.

### Cross-system Resource Sharing

Let's suppose that we need to share a system's resources with others.

For example,
we've developing an application allowing a user to upload their files from **GoogleDrive**.
In a way, **GoogleDrive** must approve the application to perform actions on behalf of the user.

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

Similar to what we did in the [Identity Provider](#identity-provider) section,
the resource service (**GoogleDrive**) must **not** let the application keep the user credentials (passwords, access tokens)
as they fully represent the user and therefore can be used to perform **any actions**.
Instead, we should create a different credential for the application,
which is limitted to granted permissions.

#### OAuth2.0

**OAuth2.0** is an authorization framework that allows third-party applications
to access user resources hosted on a service.

##### Basic Flow

First, the system redirects the resource service, such as **GoogleDrive**.
The user signs in and consents permissions giving to the application.
Then, the resource service responds back an **Access Token** limitted to granted permissions.

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
g -> u: Respond an access code {
  style.bold: true
}
```

The user sends the token to the application.
With the access token,
the system can access to resources shared by the user.

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

This is the basic flow of **OAuth2.0**.
However, we will encouter a security concern.
The code is responded to the frontend layer (browser or device),
but the actual actor accessing the resource is the backend system,
the access token should be only obtained by it,
otherwise it can be exploited here.

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

##### Exchange Code

To avoid the problem,
the concept of exchanging access tokens is introduced.
Instead of responding access tokens directly,
the resource service sends back an **Exchange Code**.

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
g -> u: Respond an exchange code {
  style.bold: true
}
```

The user needs to send the token to the application (backend system).
In order to access resources,
the application first uses the code to exchange an actual access token.

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
g -> u: Responds an exchange code
u -> s: Send the exchange code
s -> g: Exchange an access token with the code {
  style.bold: true
}
```

Now, attackers can only steal exchange codes, not actual access tokens.
However, this does change nothing as
the attacker can still use the code to obstain an access token.
Thus, exchanging tokens is only feasible for **trusted targets**.

##### Trusted Target

An application needs to intially register itself with the resource service:

- It's uniquely detected with an **id**.
- It's given with a **secret** to validate in the exchanging phase.

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

When redirecting users to the resource service,
an application must attach its **id** to the path,
such as `/auth/id=123` to produce its own exchange codes.

When exchanging access tokens,
the application uses its identifier (id and secret) to prove it a trusted target.
Thus, the application needs to send the code and its identifier,
the resource check them all before responding the token.

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
s -> u: "Redirect to /auth/id=123"
u -> g: Sign in and consent permissions
g -> u: Respond an exchange code
u -> s: Send the exchange code
s -> g: Exchange an access token with (exchange code + identifier) {
  style.bold: true
}
g -> s: Respond an access token
```

Now, the application is the only one can issue access token.
Even if exchange codes are unexpectedly stolen,
they can't be exploited.
