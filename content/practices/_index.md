---
title: Practices
weight: 50
prev: design-patterns
next: booking-system
---

This final section transitions from theoretical concepts to practical application by examining several intricate systems from a design perspective.
The topics are organized into two distinct parts:

- **Abstract Design**: This section focuses on designing each system conceptually, ensuring the architecture is platform-agnostic and applicable across self-managed infrastructures and cloud providers like **AWS**, **Azure**, etc.
- **Implementation**: This part specifies how to deploy and maintain the systems on [AWS](https://aws.amazon.com/).
This section requires a good understanding of cloud computing and **AWS**. You may wish to skip it if you are not familiar with these topics.

We will discuss the following applications:

**1. [Booking System]({{< ref "booking-system" >}})**

- Raises awareness about challenges in large-scale systems,
such as [Data Distribution]({{< ref "peer-to-peer-architecture" >}}) and [Concurrency Control]({{< ref "concurrency-control" >}}).
- Involves working with a simple **multi-region** setup in **AWS**.

**2. [Chat System]({{< ref "chat-system" >}})**

- Explores how to develop a real-time application and maintain a [WebSocket]({{< ref "communication-protocols" >}}) cluster.
- Demonstrates how to manage system complexity with [Queuing]({{< ref "event-streaming-platform" >}}).

**3. [Video-on-demand And Livestreaming System]({{< ref "vod-system" >}})**

- Examines a system that deals heavily with [Media Storage]({{< ref "media-storage" >}}).
- Covers the design of a system that relies mainly on background processing.
