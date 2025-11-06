---
title: Video-on-demand And Livestreaming System
weight: 40
---

This document outlines the design of a platform providing both video-on-demand (VOD) and live-streaming capabilities,
using a platform like **YouTube** as a prime example.

## Requirements

We will focus on the primary features:

**Serving VOD**:

- Users can upload videos to the system.
- The system must process and serve videos in various renditions (i.e., different bitrates and resolutions)
to accommodate different network conditions and devices.

**Real-time Live-streaming**:

- The system must allow a streamer to broadcast content to multiple viewers in real-time.
- Viewers must be able to replay moments from an ongoing or completed live stream session.

**Non-functional Requirements**:

- Streaming latency (both VOD and live) should be low, ideally under 5 seconds.
- The application must serve users primarily in **Southeast Asia** and the **United States**.

## System Overview

The system is logically divided into two main services:

- The **Video Service** handles the lifecycle of VOD content, from uploading to viewing.
- The **Streaming Service** manages real-time broadcasts, allowing users to stream content and viewers to watch it live.

```d2
direction: right
streamer: Streamer {
    class: client
}
s: System {
    s: Streaming Service {
        class: server
    }
    vs: Video Service {
        class: server
    }
}
v: Viewer {
    class: client
}
streamer -> s.s: cast
v <- s.s: view 
v <- s.vs: view
```

## Video Service

We will first address the design of the VOD component. The user's download or viewing experience fundamentally drives how the upload and processing pipeline should be built.

### Downloading (Video Consumption)

Serving large, monolithic video files directly to users is inefficient. A user might only watch a small portion of a video, making the download of an entire file slow and wasteful of bandwidth.

#### Chunking

To solve this, modern video platforms use adaptive streaming protocols like **HTTP Live Streaming (HLS)** or **Dynamic Adaptive Streaming over HTTP (DASH)**. These protocols work by breaking a video into small, sequential segments (e.g., 2-6 seconds in length). A **manifest file** is created to list these segments in the correct order.

- The client player first downloads the manifest file.
- It then requests the individual segments listed in the manifest to play the video.

```d2
grid-rows: 2
m: Manifest File {
    grid-rows: 1
    grid-gap: 0
    s1: "Segment 1 (Length = 5s)" {
        width: 300
    }
    s2: "Segment 2 (Length = 5s)" {
        width: 300
    }
    s3: "Segment 3 (Length = 3s)" {
        width: 300
    }
}
s: Storage {
    s1: "Segment_1.file" {
        class: file
    }
    s2: "Segment_2.file" {
        class: file
    }
    s3: "Segment_3.file" {
        class: file
    }
}
m.s1 -> s.s1
m.s2 -> s.s2
m.s3 -> s.s3
```

#### Adaptive Bitrate Streaming

To support a wide range of devices and network conditions, the original video is transcoded into multiple versions, each with a different resolution and bitrate. For each version, a set of segments and a corresponding manifest (playlist) are created. A **master manifest** file is then created to reference all the available quality levels.

The resulting file structure for an HLS stream might look like this:

```md
my-video.master.m3u8 (master playlist)
├── 1080p_playlist.m3u8
│   ├── segment_001.ts
│   ├── segment_002.ts
│   └── ...
├── 720p_playlist.m3u8
│   ├── segment_001.ts
│   ├── segment_002.ts
│   └── ...
```

When a viewer starts watching, the player downloads the master manifest, detects the user's available bandwidth, selects the most appropriate quality playlist, and begins downloading its segments. The player can dynamically switch to a higher or lower quality playlist during playback if network conditions change.

### Uploading

Instead of having the client upload a large video file to the Video Service, which then forwards it to storage, we can allow the client to upload directly to the storage solution to improve efficiency.

The upload process is as follows:

1. The user's client requests a secure, unique upload link from the **Video Service**. This link points to a unique location (like a folder) in the object storage.
2. The client's browser or application slices the video into fixed-length chunks and uploads these chunks in parallel to the provided link, increasing upload speed and reliability.
3. Once all chunks have been uploaded, a `COMPLETED` event is triggered.
4. The **Video Service** listens for this event and initiates the backend encoding pipeline.

```d2
shape: sequence_diagram
u: User {
    class: client
}
v: Video Service {
    class: server
}
s: Storage {
    class: resource
}
e: COMPLETED event {
    class: msg
}
u -> v: requests to upload a video
v -> u: responds with a unique link
u -> s: uploads chunks in parallel {
    style.animated: true
}
s -> e: completes uploading
v <- e: performs encoding
```

### Adaptive Bitrate Encoding Pipeline

Once the source video chunks are uploaded, the backend pipeline transforms them into the final, streamable formats. This involves three main steps.

{{% steps %}}

#### Transcoding

The raw source video is transcoded (converted) into multiple streams, each with a different bitrate and resolution (e.g., `1080p @ 4500 kbps`, `720p @ 2800 kbps`, etc.). This allows for adaptive bitrate streaming.

#### Segmentation

Each transcoded stream is then segmented. Storing every single frame of a video is inefficient. Instead, video compression techniques like **Group of Pictures (GOP)** are used. A GOP consists of:

- An **I-frame (Intra-coded frame)**: A full, complete picture, acting as a starting point.
- **P-frames (Predicted frames)**: A series of subsequent frames that only store the *changes* from the previous frame.

This means a player must start from an I-frame and apply the subsequent P-frames to reconstruct the video for that segment, significantly reducing storage size.

#### Packaging

Finally, for each transcoded and segmented version of the video, the service generates the corresponding manifest files (e.g., `.m3u8` for HLS, `.mpd` for DASH). It also creates the master manifest that points to all the individual quality-level manifests. This final step is called **Packaging**.

{{% /steps %}}

In summary, the video processing pipeline takes the uploaded source segments and runs them through this workflow:

```d2
direction: right
s: Source segments {
    class: file
}
t: Transcoding {
    grid-columns: 1
    c0: 1080p {
        class: file
    }
    c1: 720p {
        class: file
    }
}
seg: Segmentation {
    se: Segment {
        grid-columns: 1
        i: I-frame {
            class: file
        }
        p0: P-frame 0 {
            class: file
        }
        p1: P-frame 1 {
            class: file
        }
        p2: P-frame 2 {
            class: file
        }
    }
}
p: Packaging {
    m: Manifest files {
        class: file
    }
}
s -> t
t -> seg
seg -> p
```

## Streaming Service

Now, let's explore the live-streaming feature.
While our [Chat System]({{< ref "chat-system" >}}) design also involved real-time communication,
live-streaming operates on a different model with distinct challenges:

1. **One-to-Many Delivery**: Streaming follows a one-to-many broadcast model, where a single streamer's content is distributed to a large number of viewers.
2. **Continuous Media Data**: Unlike the intermittent text messages in a chat application, live-streaming requires the continuous, high-throughput delivery of media data (video and audio frames).

### Real-time Ingest

The first step in a live stream is ingesting the media from the streamer's device into our system.
This requires a different approach than the chunked uploads used for VOD.
For live content, we cannot wait for segments to be created on the client side;
we need to receive a continuous flow of data as it's being created.

This can be handled by stateful servers that use a real-time streaming protocol.
A modern choice for this is **Secure Reliable Transport (SRT)**, which is built on UDP for low latency but includes a reliability layer to detect and recover lost packets.

Conceptually, the streamer establishes a connection with the `Streaming Service` and transmits a continuous stream of media frames.
The service is responsible for receiving and processing this stream in real-time.

```d2
direction: right
str: Streamer {
    class: client
}
s: Streaming Service {
    class: server
}
str -> s: streams media {
    style.animated: true
}
```

### Playback and Adaptive Bitrate Streaming

While protocols like **SRT** or **RTMP** are excellent for ingesting the stream,
they are not ideal for delivering it to a large audience of viewers.
These protocols often require special client software and are less scalable than standard HTTP-based delivery.

To ensure wide compatibility, support adaptive bitrate switching, and enable features like rewinding a live stream,
we will convert the incoming real-time stream into an adaptive HTTP-based format like **HLS** or **DASH**.
This makes the viewing experience very similar to watching a VOD file.
The key difference is that for a live stream, the player must **periodically reload the manifest file** to discover newly available segments as they are created.

### The Live-Streaming Encoding Pipeline

The process of handling the incoming live stream mirrors the VOD encoding pipeline,
with the core steps being **Transcoding**, **Segmentation**, and **Packaging**.
However, these steps must be performed continuously and in real-time.

- **Transcoding**: The incoming frames are transcoded on the fly into multiple renditions (e.g., 1080p, 720p, 480p).
- **Segmentation**: The service buffers the transcoded frames and groups them into segments (e.g., 2-4 seconds long) using the **Group of Pictures (GOP)** format.
- **Packaging**: As soon as a new segment is created for each rendition,
the service immediately updates the corresponding manifest files. This makes the new segment available to viewers.

### Comprehensive Streaming Pipeline

The complete pipeline for a live stream session is a continuous loop of ingesting frames, processing them, and making them available to viewers who periodically check for updates.

```d2
direction: right
s: Source frames {
    class: frame
}
t: Transcoding {
    c0: 1080p {
        class: file
    }
    c1: 720p {
        class: file
    }
}
seg: Segmentation {
    se: Segment {
        i: I-frame {
            class: file
        }
        p0: P-frame 0 {
            class: file
        }
        p1: P-frame 1 {
            class: file
        }
        p2: P-frame 2 {
            class: file
        }
    }
}
p: Packaging {
    m: Manifest files {
        class: file
    }
}
v: Viewers {
    class: client
}
s -> t
t -> seg
seg -> p
v <- p: reloads {
    style.animated: true
}
```

## Implementation

This section covers the deployment of the video and live-streaming solution on **AWS**, focusing on a **multi-region** architecture to serve users in **Southeast Asia** and the **United States**.

### AWS Media Services

AWS offers a suite of pre-built services for media workflows, including:

- **AWS Elemental MediaConvert** for VOD file transcoding.
- **AWS Elemental MediaLive** for live stream ingesting and transcoding.
- **AWS Elemental MediaPackage** for packaging and originating content.

While these services simplify development and management, they can be expensive, especially at a large scale. To optimize for cost, this design will focus on a self-hosted solution using core AWS building blocks like EC2 and S3.

### Video Service Implementation

#### Separating Encoding from Core Service

Transcoding is a computationally intensive process, often requiring expensive GPU resources. To manage costs effectively, we will separate the VOD processing into two distinct services:

1. **Video Service**: Runs on reliable, on-demand instances (e.g., using Amazon ECS) and handles user-facing interactions like upload requests.
2. **Video Encoding Service**: Runs on cheap but unreliable **EC2 Spot Instances**.
Spot Instances are unused EC2 capacity offered at a significant discount.
While they can be terminated by AWS with short notice, they are perfect for fault-tolerant, asynchronous tasks like video encoding.

#### S3-Powered Upload Workflow

We will leverage several **Amazon S3** features to build an efficient and scalable upload process:

- **S3 Pre-signed URLs**: The Video Service generates a secure, temporary URL that grants the user's client permission to upload directly to a specific location in an S3 bucket.
- **S3 Multipart Upload**: The client can break the large video file into smaller parts and upload them in parallel, increasing speed and resilience.
- **S3 Event Notifications**: Once the multipart upload is complete, S3 automatically publishes an event to an **Amazon SQS (Simple Queue Service)** queue.
- The **Video Encoding Service** polls this SQS queue for new messages, retrieves the event, and begins the transcoding process on the newly uploaded file.

The complete workflow is as follows:

```d2
shape: sequence_diagram
u: User {
    class: client
}
v: Video Service {
    class: aws-ecs
}
e: Video Encoding Service {
    class: aws-ecs
}
s: S3 Bucket {
    class: aws-s3
}
q: SQS queue {
    class: aws-sqs
}
u -> v: requests to upload a video
v -> s: creates a presigned URL
v -> u: responds the URL
u -> s: multipart uploads {
    style.animated: true
}
s -> q: publishes CREATED event
e <- q: gets the event
e -> s: performs encoding
```

### Streaming Service Implementation

For live-streaming, the **Ingesting** and **Encoding** stages are tightly coupled, as frames must flow continuously from one to the other.

```d2
direction: right
i: Ingesting instance {
    class: server
}
t: Encoding instance {
    class: server
}
i -> t {
    style.animated: true
}
```

While this suggests a one-to-one relationship, we still want to run the expensive encoding tasks on EC2 Spot Instances.
Therefore, we will separate the workload into two services:

1. **Ingesting Service**: Runs on reliable instances to receive the incoming stream from the streamer.
2. **Streaming Encoding Service**: Runs on Spot Instances to handle the real-time transcoding, segmentation, and packaging.

Low-level streaming protocols like **SRT** or **RTMP** operate at Layer 4 and are not compatible with the Layer 7 Application Load Balancer (ALB).
Instead, we will use the **Network Load Balancer (NLB)** to distribute traffic:

- A **public-facing NLB** distributes incoming streams from publishers to the **Ingesting Service**.
- An **internal NLB** relays the stream from the **Ingesting Service** to the **Streaming Encoding Service**.

```d2
direction: right
s: Source {
    class: client
}
n0: Public Network Load Balancer {
    class: aws-elb
}
i: Ingesting Service {
    class: aws-ecs
}
n1: Internal Network Load Balancer {
    class: aws-elb
}
t: Streaming Encoding Service {
    class: aws-ecs
}
s -> n0
n0 -> i
i -> n1
n1 -> t
```

### Storage Layer

#### Cross-Region Replication and Delivery

A common pattern for delivering content globally is to place the object storage behind a **Content Delivery Network (CDN)** like **Amazon CloudFront**. This caches content closer to users, significantly reducing latency and S3 data transfer costs.

This leaves a key decision: should we use a single S3 bucket or two separate buckets, one in each region?

- **Single Bucket**: Simpler and cheaper, but results in higher latency for uploads and cache misses in the remote region.
- **Two Buckets**: Offers better availability (via failover) and lower latency for both uploads and cache misses.
However, it doubles storage costs and incurs cross-region replication fees.

For this project, we will prioritize user experience and availability by using **two separate S3 buckets**. We will use **S3 Cross-Region Replication (CRR)** to keep the buckets synchronized. To direct CloudFront to the closest bucket, we will use an **S3 Multi-Region Access Point** as the CloudFront origin. This access point automatically routes requests to the S3 bucket with the lowest latency for the end-user.

```d2
direction: right
cdn: Cloudfront {
    class: aws-cloudfront
}
m: Multi-region Access Point (least-latency) {
    class: gw
}
s3 {
    class: none
    grid-columns: 1
    vertical-gap: 100
    s3-ap: S3 Bucket ap-southeast-1 {
        class: aws-s3
    }
    s3-us: S3 Bucket us-east-1 {
        class: aws-s3
    }
    s3-ap <-> s3-us: CRR {
        style.animated: true
    }
}
cdn -> m
m -> s3
```

#### CloudFront Cache Configuration

We need different caching strategies for VOD and live content:

- **VOD**: Manifests and segments are immutable, so they can be cached in CloudFront with a long **Time-to-Live (TTL)**, such as one day.
- **Live-Streaming**: Manifests are updated frequently as new segments are added. We must configure a low TTL for these manifests (e.g., if segment duration is 6 seconds, the manifest TTL could be 3 seconds) to ensure viewers get updates promptly.

While frequent reloading of live manifests can increase CloudFront costs,
these files are typically small. Their size can be further reduced with techniques like **manifest compression**,
using **relative paths** for segments, and implementing **time-shifted viewing** (which removes old, expired segments from the manifest).

Notably, the **DASH** protocol handles this more efficiently than HLS by supporting templated segment URLs (e.g., `segment_$Number$.m4s`), allowing the player to predict the URL for the next segment without having to re-download the full manifest every time.
