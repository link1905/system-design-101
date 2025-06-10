---
title: Data Security
weight: 10
prev: system-administration
---

## Data Compliance

**Data Compliance** involves adhering to legal frameworks that protect sensitive data by regulating how organizations
collect, store, process, and share such information.

Several well-known frameworks are widely recognized and implemented:

- **General Data Protection Regulation (GDPR):** Applies to any organization worldwide that processes personal data of EU residents.
- **Health Insurance Portability and Accountability Act (HIPAA):** Protects personal health information (**PHI**) within the U.S. healthcare sector.
- **Payment Card Industry Data Security Standard (PCI DSS):** Ensures organizations securely accept, transmit, and store credit card information.

When operating in a specific region, country, or industry,
it is essential to thoroughly research and strictly follow applicable rules and regulations.
**Non-compliance** can result in severe financial penalties, legal consequences, and significant damage to an organization’s reputation.

In essence, the most effective ways to achieve data compliance include:

### Data Minimization

The best way to protect data is to avoid collecting it in the first place.
**Data Minimization** emphasizes collecting, processing,
and storing only the minimum amount of sensitive data necessary to fulfill a specific purpose.

Key practices include:

- Only retain user data when essential for business needs.
- Delete sensitive data when it is no longer needed (e.g., after account deletion).

### Data Segmentation

**Data Segmentation** involves dividing data into distinct segments within the system to reduce complexity, limit risk, and lower compliance costs.

Example: Under **GDPR**, only personal data falls under strict regulation.

- Customer personal data should be accessible to the **Customer Service** department only.
- If data must be shared with the **Analytics** team, it should be anonymized by removing personal identifiers like names or addresses.

```d2
c: Customer Service {
  c: |||yaml
  email: john@mymail.com
  name: John Doe
  problem: Failed to subscribe the service
  |||
}
a: Analytics Service {
  c: |||yaml
  problem: Failed to subscribe the service
  |||
}
c -> a: Remove identifiers
```

This approach not only reduces the attack surface, making unauthorized access more difficult,
but also limits the parts of the system subject to strict compliance requirements.

### Third-Party Services

Shifting responsibility to reputable third-party services can enhance compliance and security.

For example, **storing customer payment information** for automatic subscriptions involves rigorous security measures and regulatory challenges,
not only against external threats, but also to protect data from internal misuse.

If the necessary security controls cannot be guaranteed in-house,
it is best to use a trusted third party (such as **Stripe**) to handle these tasks.

---

Next, we’ll explore common techniques to protect data.

## Data Masking

**Data Masking** is the technique of obscuring sensitive data by replacing it with meaningless or partially hidden values.

For example,
**HIPAA** mandates that **Personal Health Information (PHI)** be de-identified before use in secondary contexts
(e.g., analytics or research) outside direct patient care.

- **Original PHI:**

```yaml
CitizenId: 1111-2222-9999
Symptom: trouble sleeping
Conclusion: asthenia
```

- **Masked PHI:**

```yaml
CitizenId: XXXX-XXXX-9999
Symptom: trouble sleeping
Conclusion: asthenia
```

In this example, the `CitizenId` is partially masked, only the last segment is visible.
This might enable operational tasks, such as a receptionist verifying patients by asking for the last four digits.
If full anonymity is needed, the entire field could be completely hidden.

The process is **non-reversible**; masked data cannot be restored to its original form.
It supports safe use of realistic data for testing, analytics, and sharing without exposing sensitive values.

## Data Tokenization

**Data Tokenization** involves replacing sensitive data with a unique, meaningless **token**.
The actual, sensitive data is securely stored in a separate, protected **Vault**.

Consider the development of a proprietary `Payment Service` designed specifically to securely store and process sensitive credit card details:

1. The client initiates a subscription by interacting with the `Subscription Service`.
2. The `Subscription Service` then forwards the client to the `Payment Service`.
3. The client submits their card information directly to the `Payment Service`.
4. The `Payment Service` generates a unique **token** (e.g., `TKN1234`) to represent the card details.
5. The `Subscription Service` receives only this **token**. It uses this **token** for future payment operations,
such as charging or refunding.

```d2
shape: sequence_diagram
c: Client { class: client }
s: Subscription Service { class: server }
p: Payment Service
v: Vault

c -> s: 1. Subscribe
s -> p: 2. Forward
c -> p: 3. Send card information
p {
  k: |||yaml
  (111222333, 123)
  |||
}
p -> v: 4. Generate random token
v {
  k: |||yaml
  TKN123: (111222333, 123)
  |||
}
p -> s: 4. Send token TKN1234
s -> p: 5. Charge with token TKN1234 {
  style.bold: true
}
```

This workflow centralizes the handling of sensitive data within the secure `Payment Service`, **minimizing exposure** elsewhere.
The `Subscription Service` can perform actions related to the card (e.g., process payments)
without ever possessing or accessing the actual card number.

If a **stolen token** is used in an unauthorized attempt to initiate a charge,
the transaction would still be processed by the legitimate `Payment Service`.
Any funds would be directed to the intended merchant, not the attacker,
because the token only has operational meaning within that secure payment system.

The effectiveness of this method is critically dependent on the robust security of the **Vault**.
The **Vault** must be maintained as a strictly isolated component with rigorously enforced,
tightly controlled access mechanisms.

## Cryptographic Hashing

**Cryptographic Hashing** is a security technique that protects data by converting it into a non-sensitive form using a hash function,
such as [MD5](https://en.wikipedia.org/wiki/MD5) or [SHA](https://en.wikipedia.org/wiki/SHA-1).

Hash functions are mathematically complex.
Even a minor change in the input (such as a single character) produces a completely different output.

For example, we hash two values with the **MD5** algorithm:

- `Hash_MD5("mypassword") → "0d21908a7454"`
- `Hash_MD5("mypassword1") → "5cc716f9be1a"`

A key feature is that hash values are non-reversible, meaning the original value cannot be derived from them.
As a result, hashed values are used only for **comparison and verification** purposes, not for retrieving the original data.
Saving user passwords is the most common use case.

Consider this example:

- User passwords should not be stored in plain text.
If the storage is breached, the password will be stolen.
Even without an external attack, employees with access could misuse it.

```yaml
Email: example@gmail.com
Password: mystrongpassword
```

- Instead, hash passwords (for example, using MD5) and saving only the hashed values are recommended:

```yaml
Email: example@gmail.com
Password: c924729b0e04eb0d21908a7454c0218a # MD5(mystrongpassword)
```

- When a user signs in, we hash the input and compare it with the stored value:

```yaml
UserInput: mystrongpassword → c924729b0e04eb0d21908a7454c0218a
```

Even if the database containing user credentials (the password store) is spitefully accessed,
the actual user passwords can remain unrevealed.

### Pattern Recognition

A common drawback of this approach is **Pattern Recognition**.
Since hash functions always return the same output for the same input,
attackers can use this consistency to infer original values.

### Rainbow Table

**Rainbow Table** is a well-known hacking technique based on pattern recognition.
It involves precomputing and storing popular passwords with their corresponding hash values.

For example,
an attacker might pre-calculate and catalogue the cryptographic hashes for a substantial collection of frequently used passwords:

```yaml
MD5Rainbow:
  helloworld: fc5e038d38a57032085441e7fe7010b0
  hello123: f30aa7a662c728b7407c54ae6bfd27d1
  mygooglepassword: 884f755c6750cb773cbb37589a9972bf
```

Consider a user store with the same hashing algorithm:

```yaml
user1:
  Email: user1@gmail.com
  Password: 884f755c6750cb773cbb37589a9972bf
user2:
  Email: user2@gmail.com
  Password: fc5e038d38a57032085441e7fe7010b0
```

By comparing these values, it’s clear that `user1`’s password is `mygooglepassword` and `user2`’s is `helloworld`.

### Salt

**Salt** is a random value added to the sensitive data before hashing.
This causes identical inputs to produce different hash results.

For example, when users have the same password,
hashing without salt produces identical hashes:

```yaml
user1:
  # MD5(mystrongpassword)
  Password: c924729b0e04eb0d21908a7454c0218a

user2:
  # MD5(mystrongpassword)
  Password: c924729b0e04eb0d21908a7454c0218a
```

When a random salt is added:

```yaml
user1:
  Salt: s1
  # MD5(mystrongpassword + s1)
  Password: 1e2381d9b7ef33eab1f79d392ceadc81

user2:
  Salt: s2
  # MD5(mystrongpassword + s2)
  Password: 53346d86b558b33653371c2083cd760b
```

Salts are stored alongside with user records, and passwords remain interpretable.

However,
hashing is a **resource-intensive operation**,
and the addition of **salt** makes it significantly harder for attackers to utilize **Rainbow Tables**.
Attackers are compelled to combine every potential password with each user’s unique salt,
greatly increasing the effort required to compromise the stored credentials
and giving the system more time to respond.

### Pepper

**Pepper** is a **hidden, shared value** used across records.

For example,
when a shared secret **pepper** (e.g., `p1`) is added,
the hashed password is then derived from `(password, pepper, salt)`.

- The **salt** works to ensure that the same password generates different hash outputs.
- The **pepper**'s role is to hide the method by which these hashed passwords are calculated.

```yaml
user1:
  # Secret pepper: p1
  Salt: s1
  # MD5(mystrongpassword + p1 + s1)
  Hash: 60558839fa98235fa8cd9bdfe633b240
```

Pepper requires the additional task of storing the secret securely.
As long as the pepper is kept protected, this method keeps user passwords undetectable.

## Data Encryption

**Data Encryption** is a security technique that relies on
cryptographic algorithms and consists of two main phases:

- **Encryption:** Uses a key to transform data into ciphertext.
- **Decryption:** Uses another key to revert the ciphertext to its original form.

```d2
grid-columns: 1
d: Data {
  c: |||yaml
  id: 123
  name: Stew
  |||
}
k {
  class: none
  grid-rows: 1
  k1: Encryption {
      class: pub-key
  }
  k2: Decryption {
      class: pub-key
  }
}
c: Ciphertext {
  bfc4aa58713836
}
d -> k.k1
k.k1 -> c
c -> k.k2
k.k2 -> d
```

There are two main types of encryption:

### Symmetric Encryption

**Symmetric Encryption** uses a single key for both encryption and decryption.

```d2
grid-rows: 1
horizontal-gap: 150
d: Data {
  c: |||yaml
  id: 123
  name: Stew
  |||
}
k: Symmetric Key {
  class: pub-key
}
ed: Ciphertext {
  bfc4aa58713836
}
d <-> k
k <-> ed
```

Mathematically, this approach is straightforward and extremely fast compared to the following method.

However, the single key is highly powerful.
In some cases, we may want to expose either the encryption or decryption capability to external parties, but not both.

### Asymmetric Encryption

**Asymmetric Encryption** uses two distinct keys,
each assigned to one phase of the process.

- One key is used to encrypt the data.
- The other key is used to decrypt the data.


```d2
grid-columns: 1
d: Data {
  c: |||yaml
  id: 123
  name: Stew
  |||
}
k {
  class: none
  grid-rows: 1
  k1: Encryption Key {
      class: pub-key
  }
  k2: Decryption Key {
      class: pub-key
  }
}
c: Ciphertext {
  bfc4aa58713836
}
d -> k.k1
k.k1 -> c: Encrypt
c -> k.k2
k.k2 -> d: Decrypt
```

One of the two keys will be published for consumers to use;
this is the **Public Key**.
The other key is retained privately and is called the **Private Key**.
Typically, the public key is used for encryption, as decryption is often considered more sensitive.

For example,
we might distribute the public key so that the encryption step can occur on the client side,
leaving the system responsible only for decryption.

```d2
direction: right
s: System {
  pri: Private Key {
    class: pri-key
  }
  pub: Public Key {
    class: pub-key
  }
}
c: Client {
  pub: Public Key {
    class: pub-key
  }
}
s.pub -> c.pub: Distributed {
  style.animated: true
}
```

However, because it relies on a pair of keys,
decryption with asymmetric encryption is mathematically much slower.
Whenever possible, we should prefer symmetric key encryption.

#### Signature Validation

**Signature Validation** is a widely recognized use case for asymmetric encryption.

This mechanism allows a system to **seal** data and
gives clients the means to verify that the data is authentic.

Suppose a system possesses both a private and public key,
and distributes the public key to a client.

```d2
direction: right
s: System {
  pri: Private Key {
    class: pri-key
  }
  pub: Public Key {
    class: pub-key
  }
}
c: Client {
  pub: Public Key {
    class: pub-key
  }
}
s.pub -> c.pub: Distributed {
  style.animated: true
}
```

If the system needs to securely share data,
it first encrypts the data with its private key before sharing it.
The client then uses the public key to decrypt and verify the data.

```d2
direction: right
s: System {
  pri: Private Key {
    class: pri-key
  }
  data: |||yaml
  id: 123
  name: Stew
  |||
  data -> pri
}

enc: "bfc4aa58713836"

c: Client {
  pub: Public Key {
    class: pub-key
  }
  data: |||yaml
  id: 123
  name: Stew
  |||
  pub -> data
}

enc -> c.pub: Decrypt
s.pri -> enc: Encrypt
```

As long as the private key remains protected, we can guarantee:

- **Authentication:** Only trusted sources can produce valid data,
since others cannot create valid ciphertext without the private key.
- **Immutability:** A given piece of data always produces a certain ciphertext.
Any modification results in a different ciphertext, and the public key cannot decrypt tampered data.

This concept is widely applied, for example:

- [Json Web Token (JWT)]({{< ref "iam#json-web-token-jwt" >}}).
- [SSL/TLS](https://en.wikipedia.org/wiki/Transport_Layer_Security).
- File distribution, providing assurance that distributed files are valid and unmodified.

Data encryption is a granular approach.
Key holders can independently manage the encryption/decryption process,
making it highly efficient in distributed environments.

## Key Management

Keys are essential to securing data, thus, they must be properly managed and protected.

### Key Store

Scattering keys throughout the system is not a good practice.
It is better to centralize key management,
providing a clear overview of all keys and their access permissions.
Building a centralized key store is therefore an effective solution.

#### Blackbox Keystore

A blackbox keystore can be implemented to expose only the necessary interfaces to other services,
for example, `EncryptData`, `DecryptData`, or `GetPublicKey`.

This approach is excellent for security and compliance, because secret keys are not exposed outside the store.

```d2
direction: right
s: Service {
    class: server
}
k: Key Store {
    class: pri-key
}
s -> k: Encrypt()
s -> k: Decrypt()
```

However, this method has drawbacks in terms of performance and availability.
Cryptographic operations can be resource-intensive,
so a centralized store handling all requests may become a bottleneck and a {{< term spof >}}.

#### Key Distribution

For certain cases,
keys may be distributed to clients (typically internal services),
enabling them to perform encryption and decryption locally.

```d2
direction: right

s: Outside Service {
  k: Key {
    class: pub-key
  }
}
ks: Key Store {
  class: pri-key
}
ks -> s.k: Distribute key {
  style.animated: true
}
```

Although this approach provides enhanced performance and flexibility,
it introduces challenges related to **data compliance**.
Consumers are required to securely store distributed keys and may encounter heightened compliance responsibilities as a result.

### Hold Your Own Key (HYOK)

**HYOK** means clients retain complete control over their encryption keys.
There is no need for a key store; clients are responsible for managing keys themselves.

This approach is especially useful when clients wish to conceal their data even from the backend system, as in **end-to-end encryption**.

```d2
direction: right
c: Client {
  class: client
}
s: Service {
  class: server
}
ed: Encrypted Data {
  bfc4aa58713836
}
c -> ed: Encrypt data before sending
ed -> s
ed <- s
c <- ed: Decrypt data after getting
```

This strategy requires additional support for securely sharing keys between client devices.
Since keys are stored locally, changing or losing devices can result in key, and therefore data loss.
