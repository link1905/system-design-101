---
title: Summary
weight: 60
---

Let's wrap up what we've achieved in this topic.
This topic provides an overview of key concepts in web service design and implementation:

**1. [{{< term ms >}}](../microservice/)**

- Explore the {{< term ms >}} architecture, its benefits, and how it contrasts with {{< term mono >}} systems.
- Learn about service coupling and its problems. How to use a {{< term mq >}} to decouple a large-scale system.

**2. [Service Cluster](../service-cluster/)**

- Understand how to manage clusters of services effectively, including more concepts like {{< term health >}}, {{< term ha >}}.
- Explore two types of service clusters: {{< term sl >}} and {{< term sf >}}

**3. [Communication Protocols](../communication-protocols/)**

- Dive into protocols like {{< term http >}}, {{< term ws >}}, and {{< term sse >}}. Learn how these protocols enable efficient communication between clients and servers, including streaming protocols like {{< term webrtc >}} and {{< term hls >}}.

**4. [{{< term lb >}}](../load-balancer/)**

- Discover how load balancers distribute traffic across multiple servers to ensure high availability and reliability,
including more concepts like load balancing algorithms, {{< term apigw >}}.
- Learn two common types of {{< term lb >}}: {{< term lb4 >}}, {{< term lb7 >}} and their usages.

**5. [API Design](../api-design/)**

- Understand the principles of RESTful APIs, including statelessness, uniform interfaces, and self-descriptive messages. Topics like API pagination, partial updates, and content negotiation are also covered.

Now, let's move to a deeper layer of {{< term sd >}} - **Data Persistance**.
