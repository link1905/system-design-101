---
title: Booking System
weight: 20
prev: practices
next: chat-system
---

This document outlines the system design for a simple online hotel reservation platform, similar to **Booking.com**.

## Requirements

### Functional Requirements

- **Hotel Management**: Hoteliers must be able to manage their properties, including details and room information.
- **Searching**: Users must be able to search for hotels by **city** or **name**. Clicking a search result should navigate to the hotel's detailed page.
- **Booking**: Users must be able to book one or more available rooms of a specific type at a hotel.

### Non-functional Requirements

- **Consistency**: The system must prevent booking conflicts, such as the same room being booked twice for overlapping dates.
- **Availability**: The application will primarily serve two regions: **Southeast Asia** and the **United States**.

## System Overview

The system can be broken down into three main components:

- The **Hotel Management Service** allows hoteliers to manage their property information.
- The **Search Service** enables users to find available hotels.
- The **Booking Service** handles the room reservation process.

```d2
grid-columns: 1
c: {
  class: none
  grid-rows: 1
  h: Hotelier {
    class: client
  }
  user: User {
    class: client
  }
}
s: System {
  grid-rows: 1
  m: Hotel Management Service {
    class: server
  }
  s: Search Service {
    class: server
  }
  b: Booking Service {
    class: server
  }
}
c.h -> s.m: manages
c.user <- s.s: searches for stays
c.user -> s.b: books rooms
```

> The term **service** here refers to a logical system component, not necessarily a microservice.

The design of the storage layer is fundamental, as it influences the entire system's architecture. Therefore, we will analyze the data storage for each service first.

## Hotel Management Service

A relational database is a straightforward choice for managing hotel data.

```d2
direction: right
h: HOTEL {
  shape: sql_table
  hotel_id: UUID {constraint: Primary Key}
  name: VARCHAR
  city: VARCHAR
}
r: ROOM_TYPE {
  shape: sql_table
  hotel_id: UUID {constraint: Primary Key}
  room_id: INT {constraint: Primary Key}
  name: VARCHAR
  available: INT
}
r -> h: belongs to
```

A key access pattern is that users typically work with only one hotel at a time.
This locality allows us to leverage a [NoSQL store]({{< ref "nosql-database" >}}) (such as a **Document Store** or **Column-family Store**) to improve performance and availability.
We can partition the data by `hotel_id`, ensuring that a hotel and all its associated rooms reside on the same server.

```d2
s_0: server_0 {
  grid-columns: 1
  h: hotel_0 {
    shape: sql_table
    hotel_id: hotel_0 {constraint: PARTITION KEY}
    name: My hotel
  }
  r_0: room_0 {
    shape: sql_table
    hotel_id: hotel_0 {constraint: PARTITION KEY}
    room_id: 0
    name: one-bedroom
    available: 4
  }
}
s1: server_1 {
  grid-columns: 1
  h: hotel_1 {
    shape: sql_table
    hotel_id: hotel_1 {constraint: PARTITION KEY}
    name: Another hotel
  }
  r_0: room_0 {
    shape: sql_table
    hotel_id: hotel_1 {constraint: PARTITION KEY}
    room_id: 0
    name: one-bedroom
    available: 2
  }
  r_1: room_1 {
    shape: sql_table
    hotel_id: hotel_1 {constraint: PARTITION KEY}
    room_id: 2
    name: two-bedroom
    available: 3
  }
}
```

Furthermore, **NoSQL** stores often provide a schemaless design,
which is beneficial in the hospitality industry where properties can have a wide variety of attributes.

![Room example](room_example.png)

## Search Service

Our search requirements include finding hotels by **city** or by **name**.
However, our primary data store is partitioned by `hotel_id`, which is not efficient for these search queries.

### Complete Search

A simple approach is to create duplicated datasets to serve these queries directly. We would create two new tables:

- `HOTEL_BY_CITY` is partitioned by `city`.
- `HOTEL_BY_NAME` is partitioned by `name`.

```d2
grid-rows: 1
n: HOTEL_BY_NAME {
  shape: sql_table
  name: VARCHAR {constraint: PARTITION KEY}
  hotel_id: UUID {constraint: SORT KEY}
  city: VARCHAR
}
c: HOTEL_BY_CITY {
  shape: sql_table
  city: VARCHAR {constraint: PARTITION KEY}
  hotel_id: UUID {constraint: SORT KEY}
  name: VARCHAR
}
```

With this design, a search query can be directed to the exact data partition, significantly improving performance.
However, this comes at the cost of increased storage, as each hotel record is now replicated three times.
A major drawback is the poor user experience, as it requires users to type the exact hotel or city name.

### Prefix Search

To support a better **search-as-you-type** experience, we need to handle prefix searches.
Partitioning by the full name is insufficient for this.
Instead, we can distribute data based on the leading characters of names or cities.

For instance, names starting with `A-B` go to `server_0`, `C-D` go to `server_1`, and so on.

```d2
s0: server_0 {
  c: |||yaml
  A Hotel: hotel_1
  AB Hotel: hotel_3
  B Hotel: hotel_2
  |||
}
s1: server_1 {
  c: |||yaml
  C Hotel: hotel_4
  D Hotel: hotel_5
  |||
}
```

This allows a prefix search to be routed to a single node.
However, this approach is limited. For more advanced features like fuzzy search, spell correction, or relevance ranking,
a dedicated **Search Engine** is a better solution.

### Full-text Search

To provide a rich user experience with features like autocompletion and spell correction,
we can employ a [Search Engine]({{< ref "nosql-database#search-engine" >}}).
While powerful, search engines are complex and can be costly to manage.

A significant challenge with search engines is data sharding. There is no simple criterion to distribute full-text terms efficiently.
This means a single search query often needs to be broadcast to **multiple nodes**, and the results must be aggregated.
This **scatter-gather** pattern can be problematic, especially for pagination.

For example, to fetch a page of `30 results` from `5 partitioned servers`,
each server might need to return its top `30 results`.
The application then has to sort through 150 records to produce the final page.

For this project, we will use a **Search Engine** store to provide a rich experience.
To keep the search index synchronized with the primary database,
data will be asynchronously replicated from the **Hotel Store**.

```d2
direction: right
c: Hotel Store {
  class: db
}
s: Search Engine Store {
  class: se
}
c -> s: async replicated {
  style.animated: true
}
```

## Booking Service

To handle bookings, we first need a table to store booking information.

```d2
r: ROOM_TYPE {
  shape: sql_table
  hotel_id: UUID {constraint: PARTITION KEY}
  room_id: INT {constraint: SORT KEY}
  available: INT
}
b: BOOKING {
  shape: sql_table
  hotel_id: UUID {constraint: PARTITION KEY}
  room_id: INT {constraint: SORT KEY}
  booked: INT
  booked_at: TIMESTAMP
}
```

A booking is valid only if the number of rooms being booked (`booked`)
is less than or equal to the number of rooms available (`available`).
The booking transaction follows these steps:

```d2
shape: sequence_diagram
s: Booking Service {
  class: server
}
r: ROOM_TYPE
b: BOOKING
s <- r: checks available >= booked
s -> r: deduces available -= booked
s -> b: inserts a new record
```

This sequence, however, introduces potential concurrency conflicts when multiple users try to book the same room type simultaneously.

> To learn more, refer to this article on [Concurrency Control]({{< ref "concurrency-control" >}}).

- **Unrepeatable Read**: This occurs when two transactions read the same data, but one modifies it before the other completes. In our case, two users might both see that a room is available, but one user's booking could reduce the `available` count to zero, causing the second user's subsequent update to result in a negative `available` count.

    ```d2
    shape: sequence_diagram
    r: ROOM_TYPE (available = 3)
    u1: User 1 (booked = 3)
    u2: User 2 (booked = 1)
    u1 -> r: checks: available (3) >= booked (3)
    u2 -> r: checks: available (3) >= booked (1)
    u1 -> r: deduces: available = 3 - 3 = 0
    u2 -> r: deduces: available = 0 - 1 = -1 {
      class: error-conn
    }
    ```

- **Phantom Read**: This happens when new records are inserted into a table that match the `WHERE` clause of a query in another ongoing transaction. For example, two users could both verify availability and then proceed to create booking records. This could lead to the total number of `booked` rooms exceeding the original `available` count.

    ```d2
    shape: sequence_diagram
    r: ROOM_TYPE (available = 3)
    b: BOOKING
    u1: User 1 (booked = 3)
    u2: User 2 (booked = 1)
    u1 -> r: checks: available (3) >= booked (3)
    u2 -> r: checks: available (3) >= booked (1)
    u1 -> b: creates: BOOKING(booked = 3)
    u2 -> b: creates: BOOKING(booked = 1) {
      class: error-conn
    }
    ```

In traditional relational databases, the **Serializable** isolation level is typically required to prevent phantom reads.
However, in this specific scenario, the **Unrepeatable Read** conflict occurs first;
By resolving it, we consequently prevent the phantom read from happening.
Therefore, the **Repeatable Read** (or **Snapshot Isolation**) level is sufficient.

This brings us to our NoSQL implementation.
Since we partition `ROOM` records by `hotel_id`, all transactions for a single hotel (even if they involve multiple room types) are processed on the same partition.
Many **NoSQL** solutions can efficiently enforce transaction isolation at the single-partition level,
allowing us to prevent these concurrency issues effectively.

## Implementation

This section details how to deploy the designed booking system within the **AWS** environment.

### Multi-region Setup

The system is required to serve users primarily from two regions: **Southeast Asia (ap-southeast-1)** and the **United States (us-east-1)**.
While deploying to a single region reduces costs significantly,
it would result in high latency and a degraded experience for users far from that region.
A single-region deployment also introduces a single point of failure; an outage in the primary region would bring down the entire system.

Therefore, we will deploy the system independently in both regions.
This approach increases cost and complexity, primarily due to the need for an effective data synchronization mechanism between the regions. This consideration will be central to the design of the following components.

In the next sections, we need to take this demand into account.

### Database Layer

Our design utilizes two distinct data stores:

- A **Hotel Store** for managing hotel properties and processing bookings.
- A **Search Store** for enabling full-text searches on hotel names and cities.

#### Hotel Store

The **Hotel Store** must excel at two primary functions: partitioning data and handling concurrent write conflicts.
Several databases, including **MongoDB**, **Cassandra**, and **Amazon DynamoDB**, are capable of meeting these requirements.

While open-source solutions offer flexibility and easier migration to other providers,
**DynamoDB**, a proprietary, serverless NoSQL database from AWS,
providing significant advantages within the AWS ecosystem.
It reduces operational overhead, simplifies management, and offers seamless integration with other AWS services.
Given our focus on an AWS implementation, we will use **DynamoDB**.

**DynamoDB** supports a multi-region setup natively through its [Global Tables](https://aws.amazon.com/dynamodb/global-tables/) feature.
This feature employs an active-active replication model, where writes can occur in any region.
Conflicts are resolved using a [last writer wins]({{< ref "gossip-protocol#last-write-wins" >}}) strategy.

```d2
direction: right
u: us-east-1 {
  t: DynamoDB Table {
    class: aws-dynamodb
  }
}
a: ap-southeast-1 {
  t: DynamoDB Table {
    class: aws-dynamodb
  }
}
u.t <-> a.t: active-active replication {
  style.animated: true
}
```

##### Cross-region Conflicts

While **DynamoDB** supports serializable isolation for transactions within a single region,
this guarantee does not extend to Global Tables due to their asynchronous replication.
A transaction can complete successfully in one region but be subsequently overwritten by a newer change from another region (due to **last writer wins**), potentially leading to data inconsistencies.

> For more details, refer to the AWS documentation on [transactions in DynamoDB](https://docs.aws.amazon.com/amazonDynamoDB/latest/developerguide/transaction-apis.html).

To prevent this, we must enforce a single-writer principle,
ensuring that all modifications to a specific record occur in only one region. There are two primary ways to achieve this:

- **Designate a Primary Region**: Route all write operations to a single primary region (e.g., **us-east-1**),
while the other region (**ap-southeast-1**) serves only read operations.
This approach turns the primary region into a potential bottleneck and increases latency for users far from it.

    ```d2
    grid-columns: 1
    s: Hotel Service {
      class: server
    }
    db: {
      direction: right
      class: none
      u: us-east-1 {
        t: DynamoDB Table {
          class: aws-dynamodb
        }
      }
      a: ap-southeast-1 {
        t: DynamoDB Table {
          class: aws-dynamodb
        }
      }
      u.t <-> a.t: active-active replication {
        style.animated: true
      }
    }
    s <-> db.u.t: views and books hotels
    s <- db.a.t: views only
    ```

- **Regional Data Ownership**: Assign ownership of data to a specific region.
For example, **us-east-1** handles writes for hotels in western countries,
while **ap-southeast-1** handles writes for hotels in eastern countries.

    ```d2
    grid-columns: 1
    s: Hotel Service {
      class: server
    }
    db: {
      direction: right
      class: none
      u: us-east-1 {
        t: DynamoDB Table {
          class: aws-dynamodb
        }
      }
      a: ap-southeast-1 {
        t: DynamoDB Table {
          class: aws-dynamodb
        }
      }
      u.t <-> a.t: active-active replication {
        style.animated: true
      }
    }
    s <-> db.u.t: books western hotels
    s <- db.a.t: books eastern hotels
    ```

To implement the second, preferred approach, we add a `managed_in` field to our schema to control routing for booking operations.
The final data model for the **Hotel Store** looks like this:

```d2
h: HOTEL {
  shape: sql_table
  hotel_id: UUID {constraint: PARTITION KEY}
  name: VARCHAR
  city: VARCHAR
  managed_in: CHAR
}
r: ROOM_TYPE {
  shape: sql_table
  hotel_id: UUID {constraint: PARTITION KEY}
  room_id: INT {constraint: SORT KEY}
  name: VARCHAR
  available: INT
  managed_in: CHAR
}
b: BOOKING {
  shape: sql_table
  hotel_id: UUID {constraint: PARTITION KEY}
  room_id: INT {constraint: SORT KEY}
  booked: INT
  booked_at: TIMESTAMP
}
```

#### Search Store

We will deploy an **Amazon OpenSearch** cluster to serve as our full-text search store.
Data must be replicated from the **Hotel Store** to this one.
**DynamoDB** facilitates this with **Amazon OpenSearch Ingestion**.
This service creates a pipeline that pulls new records from **DynamoDB Streams** and automatically flushes them to an **OpenSearch** cluster.

```d2
direction: right
t: DynamoDB Streams {
  class: aws-dynamodb
}
p: OpenSearch Integration pipeline {
  class: pipeline
}
s: OpenSearch Cluster {
  class: aws-open-search
}
p <- t: ingests {
  style.animated: true
}
p -> s {
  style.animated: true
}
```

Since **OpenSearch** does not natively support a multi-region active-active setup,
we will maintain two separate search clusters, one in each region.
Each cluster will be replicated from its respective regional **DynamoDB** table.

```d2
grid-rows: 1
horizontal-gap: 250
u: us-east-1 {
  t: DynamoDB table {
    class: aws-dynamodb
  }
  s: OpenSearch cluster {
    class: aws-open-search
  }
  s <- t: ingests {
    style.animated: true
  }
}
a: ap-southeast-1 {
  t: DynamoDB table {
    class: aws-dynamodb
  }
  s: OpenSearch cluster {
    class: aws-open-search
  }
  s <- t: ingests {
    style.animated: true
  }
}
u.t <-> a.t: active-active replication {
  style.animated: true
}
```

### Web Service Layer

We will build two separate, stateless services, **Hotel Service** and the **Search Service**,
to manage the data stores, allowing them to scale independently. These can be quickly deployed using the following AWS services:

- **Amazon Elastic Container Service (ECS)**: To run the services as containerized applications.
- **Application Load Balancer (ALB)**: To distribute incoming traffic across the service instances.
- **Amazon Route 53**: To route user traffic to the region with the lowest latency using latency-based routing.

This setup will be duplicated in two VPCs, one in each region.

```d2
direction: right
r: Route53 {
  class: aws-route53
}
u: us-east-1 {
  direction: right
  vpc: VPC {
    lb: Load Balancer (ALB) {
      class: aws-elb
    }
    hs: Hotel Service (ECS) {
      class: aws-ecs
    }
    ss: Search Service (ECS) {
      class: aws-ecs
    }
    lb -> hs
    lb -> ss
  }
}
a: ap-southeast-1 {
  direction: right
  vpc: VPC {
    lb: Load Balancer (ALB) {
      class: aws-elb
    }
    hs: Hotel Service (ECS) {
      class: aws-ecs
    }
    ss: Search Service (ECS) {
      class: aws-ecs
    }
    lb -> hs
    lb -> ss
  }
}
r -> u.vpc.lb
r -> a.vpc.lb
```

#### Internal Accessing

To ensure secure and efficient communication between our services and the AWS-managed data stores, we will use **AWS PrivateLink**.

- **DynamoDB**: We can access DynamoDB privately and efficiently using **VPC Gateway Endpoints**,
which are fast and do not incur data processing charges.
- **OpenSearch**: Since the OpenSearch cluster is deployed within our VPC, the service instances can access it directly.
However, for the data ingestion pipeline to communicate with OpenSearch, we need to set up a **VPC Interface Endpoint**.

```d2
grid-rows: 1
vpc: VPC {
  grid-columns: 1
  t {
    class: none
    grid-rows: 1
    hs: Hotel Service (ECS) {
      class: aws-ecs
    }
    c: {
      class: none
    }
    endpoint: Gateway Endpoint {
      class: aws-private-link
    }
    hs -> endpoint: manages hotels
  }
  b {
    class: none
    grid-rows: 1
    ss: Search Service (ECS) {
      class: aws-ecs
    }
    search: Search store (OpenSearch) {
      class: aws-open-search
    }
    interface: Interface Endpoint {
      class: aws-private-link
    }
    ss <- search: searches
    search <- interface
  }
}
a: AWS Managed {
  grid-columns: 1
  t: Hotel store (DynamoDB) {
    class: aws-dynamodb
  }
  c: {
    class: none
  }
  p: Pipeline (OpenSearch Integration) {
    class: pipeline
  }
}
a.t -> a.p: ingests data {
  style.animated: true
}
a.p -> vpc.b.interface
vpc.t.endpoint -> a.t
```

#### Routing Bookings

To prevent the cross-region write conflicts discussed earlier,
booking requests for a hotel must be processed in its designated region.
When a user sends a booking request to the closest (lowest-latency) region,
the service must check if it is the correct region to handle the write. If not, there are two ways to proceed:

1. **Server-Side Cross-Region Request**: The receiving server performs a cross-region request on behalf of the user to the correct region's service.
This can be implemented using **VPC Peering** and an **Interface Endpoint** (as **Gateway Endpoints** do not support communication outside a VPC).
This approach minimizes user-perceived latency but incurs additional costs for cross-region data transfer and the **Interface Endpoint**.

    ```d2
    direction: right
    u: us-east-1 {
      vpc: VPC {
        s: Hotel Service {
          class: server
        }
      }
    }
    a: ap-southeast-1 {
      vpc: VPC {
        interface: Interface Endpoint {
          class: aws-private-link
        }
      }
      t: Hotel store (DynamoDB) {
        class: aws-dynamodb
      }
    }
    u.vpc <-> a.vpc: peering {
      style.animated: true
    }
    u.vpc.s -> a.vpc.interface: books western hotels
    a.vpc.interface -> a.t
    ```

2. **Client-Side Redirect**: The server responds with a redirect, instructing the client to send a new request to the target region.
This increases latency because the second request must travel over the public internet, but it avoids internal data transfer costs.

    ```d2
    shape: sequence_diagram
    c: Client {
      class: server
    }
    a: ap-southeast-1 {
      class: aws-vpc
    }
    u: us-east-1 {
      class: aws-vpc
    }
    c -> a: 1. books a western hotel
    a -> c: 2. redirect to us-east-1
    c -> u: 3. retries the request
    ```

To prioritize a seamless user experience, we will use the first approach (server-side cross-region request).
