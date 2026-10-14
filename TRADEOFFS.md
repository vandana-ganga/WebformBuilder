# Trade-offs

These are the three architectural decisions that shaped this project the most. For each one I describe what I chose, what else I considered, and what the choice costs.

The short version:

| Decision | Chosen | Main reason |
|---|---|---|
| Message buffer | NATS JetStream | Durable buffering with ACKs, redelivery and replay, without Kafka's operational weight |
| Dynamic storage | PostgreSQL with JSONB | Relational integrity and flexible schemas in one database |
| Submission processing | Asynchronous | Absorbs bursts and keeps a database outage from losing submissions |

## 1. NATS JetStream as the submission buffer

Public submissions go through a durable buffer before they reach the database:


Public API  ->  NATS JetStream  ->  Submission workers  ->  PostgreSQL


The public submission endpoint is the part of the system most likely to see sudden bursts, and I did not want every request to wait on PostgreSQL. JetStream persists each message, requires an explicit acknowledgement, redelivers anything that was not acknowledged, and can replay a stream. It also has a good Go client and is simple to run. The worker ACKs a message only after the submission has been written to PostgreSQL, so a crash mid-processing leads to a redelivery, not a lost submission.

### What I considered instead

**RabbitMQ** was the closest alternative. It has a mature work-queue model, solid acknowledgements, and built-in routing and dead-lettering. I passed on it because this project has no need for complex routing, and JetStream gives me the buffering and replay I need with less to set up. If the system grew many exchanges, routing keys, or elaborate queue topologies, RabbitMQ would be the better fit.

**Kafka** handles very high throughput, partitioning, and long retention, and it has a large ecosystem. It is also much more to operate than this project needs. I need a durable buffer and reliable workers, not an event-streaming platform. Kafka would make sense if submissions were also feeding many independent consumers, such as analytics, audit, and data pipelines, at very large scale.

**Redis Streams** offers durable streams, consumer groups, acknowledgements, and replay, and Redis is already in the stack for rate limiting and caching. That is the reason I did not use it. Putting the submission buffer on the same Redis means one component's failure takes out rate limiting, caching, and ingestion together. With JetStream, the submission pipeline has its own durability boundary.

## 2. PostgreSQL with JSONB for dynamic data

Form schemas and submission payloads vary per form, but everything around them is relational:


Tenant
  +-- Form
       +-- Form version
              +-- Submission


So the structure lives in ordinary tables (`tenants`, `forms`, `form_versions`, `submissions`), and only the parts that vary are stored as JSONB: `form_versions.schema` and `submissions.data`.

This gives me foreign keys, transactions, unique constraints, and mature indexing for the relationships, plus schema flexibility where the data is dynamic, all in one database.

### What I considered instead

**MongoDB** is a natural fit for dynamic documents, and storing form data in it would be simple. But the tenant, form, version, submission chain still needs integrity constraints, which PostgreSQL enforces directly. Keeping PostgreSQL for that metadata and adding MongoDB for the documents would mean running two databases, and nothing in this project justifies that.

**A table per form** (`form_123_submissions`, `form_456_submissions`, and so on) would give each form real columns. It would also mean a schema migration every time someone edits a form, a table count that grows with the number of forms, and much harder multi-tenant management. JSONB avoids all of that.

### What JSONB costs

Fields inside JSONB are not as cheap to query as fixed columns. If the product later needs frequent filtering on arbitrary form fields, that will need extra work: JSONB indexes, extracting commonly searched fields into columns, or a dedicated search or analytics store. For now I chose flexible storage and correctness over query speed on arbitrary fields.

## 3. Asynchronous submission processing

The API publishes each submission to JetStream and returns `202 Accepted`. A worker writes it to PostgreSQL afterwards.

The simpler design is to write to PostgreSQL inside the request and respond when the write completes. That has real advantages: fewer components, easier to reason about, and a submission is queryable the moment the client gets a response. The problem is that it puts PostgreSQL on the critical path of every public request. Bursts hit the database directly, database latency becomes API latency, and a short database outage means the public endpoint rejects submissions. API traffic and database writes also have to scale together.

Going through JetStream changes that:

- **Bursts are absorbed.** JetStream buffers submissions while workers drain them at a rate the database can handle.
- **Failures are isolated.** If PostgreSQL is down for a while, messages stay in the stream and workers retry when it comes back.
- **API and workers scale separately.** More API replicas for traffic, more workers for write throughput.
- **Public latency is lower.** The request ends once the message is durably published.

### What it costs

- **Eventual consistency.** A `202 Accepted` means the submission is safely queued, not that the row is visible in PostgreSQL yet.
- **More components.** NATS and a worker process now have to be deployed and kept running.
- **More to monitor.** Queue depth, consumer health, redelivery counts, and processing failures all need watching.

I accepted these costs because the assessment asks for exactly what the async design provides: handling bursty public load and high submission volume, surviving component failure, and not losing submissions.

## The principle behind all three

PostgreSQL is the source of truth, JetStream is the durable ingestion buffer, and the public API stays stateless. Ingestion, buffering, persistence, and querying are separate stages, so each can fail, recover, and scale without dragging the others with it.
