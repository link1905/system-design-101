---
title: Network Security
weight: 20
---

Network is the biggest source of threats: virus, worm, trojan...
In essence, although how they work, they must sneak into machines to perform malicious actions.
Hence, the first action is to deter them from **intruding** illegally.
We've gone through many helpful techniques in this topic.

## Secure Transmission

By default, data transmission through the internet is plain.
Attackers can intrude the transmission line (e.g. through proxy, local router...) to steal data in between.
This attack is called [Man in the middle](https://en.wikipedia.org/wiki/Man-in-the-middle_attack)

```d2
direction: right
c: "" {
  m1: Machine 1 {
    class: server
  }
  m2: Machine 2 {
    class: server
  }
  d: Data {
    a: |||yaml
    Username: johndoe
    Password: mypassword
    |||
  }
  m1 -> d: Sends through internet
  d -> m2
}

a: Attacker {
  class: hacker
}
a <-> c.d: Access data in between
```

To avoid this attack, we must apply [end-to-end encryption](https://en.wikipedia.org/wiki/End-to-end_encryption).
Briefly, data is transferred as ciphertexts,
and **only** the two endpoints have enough cryptographic keys to process data.

For example,
a piece of data stays encrypted while moving between intermediary routers,
just the end machines can read it.

```d2
direction: right
m1: Machine 1 {
  a: |||yaml
  Username: johndoe
  Password: mypassword
  |||
}
r1: Router 1 {
  "bfc4aa58713836aae"
}
r2: Router 2 {
  "bfc4aa58713836aae"
}
m2: Machine 2 {
  a: |||yaml
  Username: johndoe
  Password: mypassword
  |||
}
m1 <-> r1 <-> r2 <-> m2
```

## Network Layer Protection

The first **E2E** approach is protecting on the network layer (layer 3) of the **OSI** model,
it operates on the machine instead of high-level applications.
In other words, data protection on this layer encrypt **everything** between the endpoints,
whatever their protocol or application is.

```d2
direction: right
m1: Machine 1 {
  class: server
}
m2: Machine 2 {
  class: server
}
m1 <-> m2: "Secure channel" {
  style.animated: true
}
```

This strategy is usually for building **Virtual Private Networks (VPNs)**,
e.g. a company encrypts all traffic between remote employees and the internal network.
We have some popular implementations: **IPSec**, **WireGuard**...

## Transport Layer Protection

Protecting everything is not always good.
An application may have its own strategy,
e.g., transmitting plain data (no encryption), using a custom algorithm for encryption...

`Transport Layer Protection` is a more granular approach,
protecting data on the transport layer (layer 4) of the `OSI` model.
In other words, this process happens at the application level,
applications control their own encryption process
```d2
%d2-import%
direction: right
m: Machine 1 {
  a1: Appication 1 {
    class: server
  }
   a1: Appication 2 {
    class: server
  }
  i: Netword interface {
    class: ni
  }
  a1 -> a1: "Encrypt data" {
    style.bold: true
  }
  a2 -> a2: "Encrypt data" {
    style.bold: true
  }
  a1 -> i: "Send encrypted data"
}
m2: Machine 2 {
   class: server
}
m.i -> m2: "Send encrypted data"
```

### Transport Layer Security (TLS)
`TLS` is the most well-known protocol for protecting the transport layer.
In fact, `TLS` operates on the application layer

> SSL is a former and deprecated version of TLS.
> Nowadays, people use `SSL/TLS` to refer to TLS

#### Session Establishment
`TLS` requires participants to maintain a secure session.
In short, before transmitting any piece of data,
they first securely exchange a **shared master key** for further **symmetrically** encrypting/decrypting data
```d2
%d2-import%
c: Client {
   k: Shared symmetric key {
      class: key
   }
}
s: Server {
   k: Shared symmetric key {
      class: key
   }
}
s <-> c
```

After a master key is consented,
it'll be used for all messages during the session.


### TLS1.2
`TLS1.3` is a newer and recommended version.
However, `TLS1.2` is still widely used, especially in legacy systems

#### Key Exchange {id="tls1.2-key-exchange"}
`TLS1.2` leverages **asymmetric** encryption to establish safe sessions between client and server.
The server needs to hold a pair of **public key** and **private key**
1. The client initiates a connection to the server.
   The server returns its **public key**
```d2
%d2-import%
shape: sequence_diagram
c: Client {
  class: client
}
s: Server (Private Key + Public Key) {
  class: server
}
"1. Server Public Key" {
  c -> s: Hello
  s {
    p: Server Public Key {
        class: key
        style.fill: ${colors.i}
        height: 60
        width: 200
    }
  }
  s -> c: Responds server public key
  c {
      p: Server Public Key {
        class: key
        style.fill: ${colors.i}
        height: 60
        width: 200
    }
  }
  s -- c {
    style.opacity: 0
  }
}
```
2. The client generates a temporary private key, called **master key**,
   encrypts it with [](Data-Protection.md#symmetric-encryption) by the server public key.
   The client sends the **encrypted master key** to the server
```d2
%d2-import%
shape: sequence_diagram
c: Client {
  class: client
}
s: Server (Private Key + Public Key) {
  class: server
}
"1. Server Public Key" {
  c -> s: Hello
  s {
    p: Server Public Key {
        class: key
        style.fill: ${colors.i}
        height: 60
        width: 200
    }
  }
  s -> c: Responds server public key
  c {
      p: Server Public Key {
        class: key
        style.fill: ${colors.i}
        height: 60
        width: 200
    }
  }
  s -- c {
    style.opacity: 0
  }
}
"2. Master Key" {
  c -> c: Generates master key
  c {
    pri: Master Key {
        class: key
        style.fill: ${colors.c}
        height: 60
        width: 200
    }
  }
  c -> c: Encrypts master key by the server public key
  c {
    enp: Encryped Master Key {
        class: kms
        style.fill: ${colors.e}
        height: 60
        width: 200
    }
  }
  c -> s: Sends the encrypted master key
  s {
    enp: Encryped Master Key {
        class: kms
        style.fill: ${colors.e}
        height: 60
        width: 200
    }
  }
  s -- c {
    style.opacity: 0
  }
}
```
3. The server uses its private key to **decrypt the master key**.
   Now both client and server keep the master key,
   they can exchange data securely by the key
```d2
%d2-import%
shape: sequence_diagram
c: Client {
  class: client
}
s: Server (Private Key + Public Key) {
  class: server
}
"1. Server Public Key" {
  c -> s: Hello
  s {
    p: Server Public Key {
        class: key
        style.fill: ${colors.i}
        height: 60
        width: 200
    }
  }
  s -> c: Responds server public key
  c {
      p: Server Public Key {
        class: key
        style.fill: ${colors.i}
        height: 60
        width: 200
    }
  }
  s -- c {
    style.opacity: 0
  }
}
"2. Master Key" {
  c -> c: Generates master key
  c {
    mas: Master Key {
        class: key
        style.fill: ${colors.c}
        height: 60
        width: 200
    }
  }
  c -> c: Encrypts master key by the server public key
  c {
    enp: Encryped Master Key {
        class: kms
        style.fill: ${colors.e}
        height: 60
        width: 200
    }
  }
  c -> s: Sends the encrypted master key
  s {
    enp: Encryped Master Key {
        class: kms
        style.fill: ${colors.e}
        height: 60
        width: 200
    }
  }
  s -- c {
    style.opacity: 0
  }
}
"3. Decrypt Master Key" {
  s -> s: Uses the server private key to decrypt the master key
  s {
    mas: Master Key {
        class: key
        style.fill: ${colors.c}
        height: 60
        width: 200
    }
  }
  s -> c: Confirms that the master key is obtained
}
Connection Was Established {
  s <-> c: ...(Exchange data with symmetric encryption and the master key)...
}
```

This model is secure **as long as** the server private key is protected.
The private key is shared between connections;
If somehow attackers can steal the key,
**all transmitted data** including historical conversations will be decrypted to exploit

### TLS1.3
`TLS1.3` is a newer and safer version.
Except for compatibility issues, we're recommended to apply this one

#### Diffie-Hellman
Diffie-Hellman is a cryptographic algorithm used for secure key exchange.

It's far easier by explaining with an example.
Let's say two machines have two **private keys** `a` and `b` respectively.
`Diffie-Hellman` helps to form a shared key securely between them.
1. First, we need to define a **public number**: `G`
2. They compute their **public key** with these numbers
- `A = a * G`
- `B = b * G`
3. The machines **exchange** their public key.
From the target's public key, a machine can get obtain the **shared key**
- `(A) Shared = a * B`
- `(B) Shared = b * A`

> Why is the magic behind? I have no idea ![](emoji-sad-cat.svg)

#### Key Exchange {id="tls1.3-key-exchange"}
`Diffie-Hellman` is the theoretical foundation of TLS1.3,
which helps to exchange keys securely

Let's see how a safe connection is established between a client and a server.
The procedure starts with a public parameter `G`
1. The client generate a temporary private key `a`,
   calculates the public key `A = a * G`, and sends it to the server
```d2
%d2-import%
shape: sequence_diagram
c: Client {
  class: client
}
s: Server (Private Key + Public Key) {
  class: server
}
"1. Client Public Key" {
  c -> c: Generate private key
  c -> c: Calculate public key
  c {
    pub: Client Public Key (A) {
        class: key
        style.fill: ${colors.i}
        height: 60
        width: 200
    }
  }
  c -> s: Send client public key
  s {
    clientPub: Client Public Key (A) {
        class: key
        style.fill: ${colors.i}
        height: 60
        width: 200
    }
  }
  s -- c {
    style.opacity: 0
  }
}
```
2. Similarly, the server also generate a private key `b`, calculates the public key `B = b * G`.
   The server computes the shared key as `S = b * A`
```d2
%d2-import%
shape: sequence_diagram
c: Client {
  class: client
}
s: Server (Private Key + Public Key) {
  class: server
}
"1. Client Public Key" {
  c -> c: Generate a private key (a)
  c -> c: Calculates a public key
  c {
    pub: Client Public Key (A) {
        class: key
        style.fill: ${colors.i}
        height: 60
        width: 200
    }
  }
  c -> s: Sends client public key
  s {
    clientPub: Client Public Key (A) {
        class: key
        style.fill: ${colors.i}
        height: 60
        width: 200
    }
  }
  s -- c {
    style.opacity: 0
  }
}
"2. Server Public Key" {
  s -> s: Generate a private key (b)
  s -> s: Calculates a public key
  s {
    pub: Server Public Key (B) {
        class: key
        style.fill: ${colors.b}
        height: 60
        width: 200
    }
  }
  s -> s: Compute the shared key (S)
  s {
    shared: Shared Key (S) {
        class: key
        style.fill: ${colors.c}
        height: 60
        width: 200
    }
  }
  s -> c: Sends server public key
  c {
    serverPub: Server Public Key (B) {
        class: key
        style.fill: ${colors.b}
        height: 60
        width: 200
    }
  }
  s -- c {
    style.opacity: 0
  }
}
```
3. Finally, the client computes the shared key `S = a * B`
```d2
%d2-import%
shape: sequence_diagram
c: Client {
  class: client
}
s: Server (Private Key + Public Key) {
  class: server
}
"1. Client Public Key" {
  c -> c: Generate a private key (a)
  c -> c: Calculates a public key
  c {
    pub: Client Public Key (A) {
        class: key
        style.fill: ${colors.i}
        height: 60
        width: 200
    }
  }
  c -> s: Sends client public key
  s {
    clientPub: Client Public Key (A) {
        class: key
        style.fill: ${colors.i}
        height: 60
        width: 200
    }
  }
  s -- c {
    style.opacity: 0
  }
}
"2. Server Public Key" {
  s -> s: Generate a private key (b)
  s -> s: Calculates a public key
  s {
    pub: Server Public Key (B) {
        class: key
        style.fill: ${colors.b}
        height: 60
        width: 200
    }
  }
  s -> s: Compute the shared key (S)
  s {
    shared: Shared Key (S) {
        class: key
        style.fill: ${colors.c}
        height: 60
        width: 200
    }
  }
  s -> c: Sends server public key
  c {
    serverPub: Server Public Key (B) {
        class: key
        style.fill: ${colors.b}
        height: 60
        width: 200
    }
  }
  s -- c {
    style.opacity: 0
  }
}
"3. Shared Key" {
  c -> c: Computes shared key
  c {
    shared: Shared Key (S) {
        class: key
        style.fill: ${colors.c}
        height: 60
        width: 200
    }
  }
  s -- c {
    style.opacity: 0
  }
}
```

Where is `G` from?
Actually, it's hardcoded into `TLS1.3` libraries,
the client and server side implicitly agree on the same value without transferring it back and forth.

We totally avoid the problem of TLS1.2, because there is no reused secret key on the server side,
each connection possesses a **temporary private key**.
If a private key is stolen, its conversation is exploited, the others are irrelevant.

### TLS Certificate
`TLS/SSL`, despite version, has a problem of `server identification`.
When an encryption key is responded,
the client is unable to detect whether this key belongs to the correct server.

Imagine that a client's network is attacked,
attackers redirect the client to a spiteful server instead of the credible one.
```d2
%d2-import%
direction: right
c: Client {
    class: client
}
w: Attacked Router {
    class: lb
}
s: Target Server {
    class: server
}
a: Attacker {
    class: hacker
}
c -> w: 1. Connects to the target server
w -> a: 2. Forwards to a harmful server
w -> s: Fails {
    class: error-conn
}
```

Instead of returning a plain public key,
the server wraps it inside a `TLS certificate` for further identification.
A certificate must be **signed** (registered) by an official `Certificate Authority (CA)` (Comodo, GoDaddy, Let's Encrypt...).
It may simply look like this
```yaml
Common Name (CN): www.example.com
Issuer: Let's Encrypt
Valid From: 2024-01-01 (YYYY-MM-DD)
Valid To: 2024-04-01 (YYYY-MM-DD)
Public Key: MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQ
```
- `Public Key`: the most important field, used to verify this cert is sent from a valid server
- Other metadata fields: common name (usually domain name), expiration information (`Valid From`, `Valid To`...), ...

#### Certificate Authority Authentication
It is possible to create **fake certificates**,
so we only accept certificates from the trustworthy issuers.

Leveraging from [](Data-Protection.md#signature-validation),
a `CA` uses a **private key** to sign its certificates.
Public keys of common `CAs` are preloaded into operating systems and browsers as part of their trusted root certificate store,
these keys are used to verify valid certificates **locally**

#### Challenge-Response Authentication
`Challenge-Response Authentication` is similar to [](Data-Protection.md#signature-validation),
but focuses on **validating machines** instead of data.
The idea is straightforward:
Clients hold a **public key**,
they try to challenge the server side to prove that it possesses the **associated private key**.

When signing a new certificate, the owner server persistently keeps a **private key**
matched with the certificate's **public key**.
Before exchanging keys, a verification phase takes place first
1. The client initiates a new connection with a **random challenge code**
```d2
%d2-import%
shape: sequence_diagram
c: Client {
    class: client
}
s: Server {
    class: server
}
Challenge code: {
  c."use your key to encrypt it"
}
c -> s: Initiates with a challenge code
Challenge code: {
  s."use your key to encrypt it"
}
```
2. The server encrypts the code with the **private key**,
   and responds its TLS certificate and the encrypted code
```d2
%d2-import%
shape: sequence_diagram
c: Client {
    class: client
}
s: Server {
    class: server
}
Challenge code: {
  c."Code: use your key to encrypt it"
}
c -> s: Initiates with a challenge code
Challenge code: {
  s."Code: use your key to encrypt it"
}
Encrypted challenge code: {
  s -> s: Encrypts the challenge code with private key
  s."Encrypted code: AAOCAQ8AMIIBCgKCAQ"
  s -> c: Responds cerficate and encrypted code
  c."Certificate (PublicKey: MIIBIjANBgkqh, Issuer: Let's Encrypt,...)"
  c."Encrypted code: AAOCAQ8AMIIBCgKCAQCertificate"
}
```
3. The client verifies the certificate by the CA's local key.
   Then, it decrypts the code by the certificate's public key to compare with the original value.
   The matching indicates a correct server,
   as without the private key, strangers cannot encrypt the challenge code properly.
```d2
%d2-import%
shape: sequence_diagram
c: Client {
    class: client
}
s: Server {
    class: server
}
Challenge code: {
  c."Code: use your key to encrypt it"
}
c -> s: Initiates with a challenge code
Challenge code: {
  s."Code: use your key to encrypt it"
}
Encrypted challenge code: {
  s -> s: Encrypts the challenge code with private key
  s."Encrypted code: AAOCAQ8AMIIBCgKCAQ"
  s -> c: Responds cerficate and encrypted code
  c."Certificate (PublicKey: MIIBIjANBgkqh, Issuer: Let's Encrypt,...)"
  c."Encrypted code: AAOCAQ8AMIIBCgKCAQCertificate"
}
Verfication: {
  c -> c: Verifies certificate with the CA's local key
  c -> c: Decrypts the challenge code to compare
  c."Decrypted code: use your key to encrypt it"
}
c <-> s: Exchange key
```

With the two verifications of certificate and server, connecting with trustworthy servers is guaranteed

### Mutual TLS (mTLS)
Sometimes, the client side needs also authenticating, e.g. microservice communication.
`Mutual TLS (mTLS)` is an extended version that ensures two-way authentication between a client and a server.

To do that, we need a dedicated `CA` only issuing certificates to **trusted** entities.
A connection establishment will involve in mutually checking certificates from both the client and server
```d2
%d2-import%
ca: Dedicated CA {
   class: kms
}
c: Client {
}
s: Server {
}
ca -> c: Certificate
ca -> s: Certificate
s <-> c: mTLS with valid certificates
```

The dedicated `CA` can verify and distribute certificates by
- Email verification for small number of users
- Integrate with [identity providers](Identity-And-Access-Management.md#federated-identity-pattern) (e.g. OAuth) for enterprise environments
