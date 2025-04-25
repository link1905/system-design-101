---
title: Decentralized Cluster
weight: 10
---

We have extensively discussed data sharding and replication.  
But how can we effectively combine them into a single virtual database?  
Maintaining a {{< term p2p >}} system is no easy feat.  
This architecture advocates a decentralized approach,  
aiming to remain highly available and fault-tolerant, with **no single point of failure**.

Without a shared master server,
cluster metadata (e.g., member addresses, sharding information, etc.)  
must be somehow both reliable and consistently shared among cluster members  
to enable tasks such as replication, sharding, and promotion.

```d2
p1: Peer 1 {
    class: server
}
p2: Peer 2 {
    class: server
}
p3: Peer 3 {
    class: server
}
p1 <-> p2 {
    style.animated: true
}
p2 <-> p3 {
    style.animated: true
}
p1 <-> p3 {
    style.animated: true
}
```

In this topic, let's dive into two common approaches maintaining a decentralized cluster: {{< term consProto >}} and {{< term gosProto >}}.

{{< callout type="info">}}
Honestly, I write them as I want to show you interesting knowledge.
However, they are heavily theoretical,
you might never touch them as a developer or system designer.
It's fine to end the **Peer-to-peer architecture** topic right here 🙂!
{{< /callout >}}
