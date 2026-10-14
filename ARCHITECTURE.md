# Architecture & Implementation

## 1. Overview

This project is a multi-tenant webform builder and submission platform. A tenant creates a form, defines its fields and validation rules, and publishes a version of it. The published form is served publicly, visitors submit data, and the server validates each submission against the published schema before storing it. Tenants can then list and page through their submissions.

Four requirements drove the design:

- **Correctness.** A submission always keeps the exact form version it was accepted against.
- **Reliability.** An accepted submission is not lost when a component fails.
- **Scalability.** A burst of public submissions does not land directly on PostgreSQL.
- **Security.** Tenants cannot see each other's data, and user-generated content is treated as untrusted.

The working code covers the core path: create a form, publish it, submit to it publicly, and store the submission durably. Redis rate limiting, exports, production load balancing, and some operational pieces are designed here but not built. [Section 14](#14-what-is-built-and-what-is-only-designed) lists exactly which is which.

## 2. System layout


                           INTERNET
                              |
                              v
                     +----------------+
                     | Load Balancer  |
                     |    / Nginx     |
                     +-------+--------+
                             |
              +--------------+--------------+
              |              |              |
              v              v              v
        +-----------+  +-----------+  +-----------+
        | Go API #1 |  | Go API #2 |  | Go API #3 |
        |   Echo    |  |   Echo    |  |   Echo    |
        +-----+-----+  +-----+-----+  +-----+-----+
              |              |              |
              +--------------+--------------+
                             |
          +------------------+------------------+
          |                  |                  |
          v                  v                  v
   +-------------+    +-------------+    +-------------+
   | PostgreSQL  |    |    Redis    |    |    NATS     |
   | System of   |    | Rate limit  |    | JetStream   |
   | record      |    | Cache       |    | Durable     |
   +-------------+    +-------------+    | buffer      |
          ^                              +------+------+
          |                                     |
          |                                     v
          |                             +---------------+
          +-----------------------------| Submission    |
                                        | workers       |
                                        +---------------+


### Components

**Load balancer (Nginx).** The single public entry point. It spreads traffic across API instances, forwards the client IP, and is where TLS would terminate in production. The API is stateless, so any healthy instance can serve any request. A cloud load balancer could replace Nginx in production.

**Go API (Echo).** Handles tenant, form, and version management, publishing, serving public forms, validating submissions, and reading submissions back. It does not write submissions to the database itself. Once a submission passes validation, the API publishes it to JetStream and returns.

**PostgreSQL.** The system of record. It stores tenants, forms, form versions, and submissions. Form schemas and submission payloads are stored as JSONB because their shape differs per form.

**Redis.** Intended for distributed rate limiting and short-lived caching of published form schemas. It is never the source of truth. This is designed but not part of the working slice.

**NATS JetStream.** The durable buffer between the public API and the database. It persists submission events to file storage, redelivers any message that was not acknowledged, and lets messages be replayed.

**Submission worker.** Consumes events from JetStream, deserializes and checks each one, inserts it into PostgreSQL, and ACKs the message only after the insert succeeds. Workers scale horizontally.

## 3. Request flows

### Creating and publishing a form


POST /api/tenants                                          -> tenant
POST /api/forms                                            -> draft form
POST /api/forms/{id}/versions                              -> version 1
POST /api/forms/{id}/versions/{version_id}/publish         -> version 1 is published


A form and its versions are separate entities. Saving a new schema never overwrites an old one; it creates a new immutable version.

### Serving a public form

`GET /public/forms/{form_id}` looks up the currently published version and returns its schema. Draft versions are never exposed on this endpoint.

### Submitting


Visitor
   |  POST /public/forms/{id}/submissions
   v
Load balancer
   v
Go API
   +--> look up the form and its published version
   +--> validate the data against that version's schema
   +--> rate limit check (designed)
   v
NATS JetStream          <- API returns 202 Accepted here
   v
Submission worker
   v
PostgreSQL INSERT
   v
ACK the JetStream message


The API returns `202 Accepted` once JetStream has durably stored the event. The submission becomes queryable shortly after, when a worker has written it to PostgreSQL.

## 4. Why submissions are processed asynchronously

A public form can go from 10 submissions per second to 2,000 with no warning. If every request ran its own PostgreSQL transaction, the database would sit on the critical path of all public traffic, and a burst larger than the database can handle would turn into failed requests.

With JetStream in between, the API keeps accepting submissions while workers drain the queue at a rate the database can sustain. API replicas scale with incoming traffic, and workers scale separately based on queue depth and processing latency.

The alternatives and the costs of this choice are covered in [TRADEOFFS.md](TRADEOFFS.md).

## 5. How submissions are not lost

Three mechanisms work together.

**Durable storage.** Submission events are written to JetStream file storage. A submission is not treated as accepted while it exists only in application memory.

**ACK after persistence.** The worker ACKs a message only after the PostgreSQL insert succeeds. If a worker crashes before that, the message stays unacknowledged and JetStream redelivers it to another worker.

**Idempotent inserts.** The submission ID is generated by the API before publishing and is the primary key of the `submissions` table. The worker inserts with `ON CONFLICT (id) DO NOTHING`, so a redelivered message cannot create a duplicate row.

Together this gives at-least-once delivery with idempotent persistence.

### Failure scenarios

| What fails | What happens |
|---|---|
| API crashes before publishing | The request was never accepted. The client gets an error and can retry. |
| API publishes, then crashes before responding | The event is safe in JetStream. If the client retries, the retry gets a new submission ID, so it is stored twice. Production should accept a client-supplied idempotency key to prevent this. |
| Worker crashes before the insert | The message is not ACKed and JetStream redelivers it. |
| Worker crashes after the insert, before the ACK | The message is redelivered, the primary key blocks the duplicate, and the worker ACKs. |
| PostgreSQL is temporarily down | Workers cannot insert, so they do not ACK. JetStream holds the messages until the database is back. |
| Redis is down | No data is at risk because Redis holds nothing authoritative. The rate-limit behaviour should be an explicit choice. For a public endpoint I recommend failing closed or falling back to a conservative local limit, not silently allowing unlimited traffic. |

## 6. Form schemas and validation

A form schema is a JSON document stored in `form_versions.schema`:

```json
{
  "fields": [
    {
      "name": "name",
      "label": "Full Name",
      "type": "text",
      "required": true,
      "min_length": 2,
      "max_length": 100
    },
    {
      "name": "email",
      "label": "Email",
      "type": "email",
      "required": true
    },
    {
      "name": "age",
      "label": "Age",
      "type": "number",
      "required": true,
      "min": 18,
      "max": 100
    }
  ]
}
```

Adding or changing a field changes this document, not the database tables.

Validation happens on the server. The API loads the published schema and checks the submitted data against it: required fields, types, email format, minimum and maximum values, string lengths, and allowed values. The client is not trusted, since anyone can bypass browser-side validation by sending a request directly.

## 7. Versioning and republishing

Version immutability is the most important correctness property in the system.


Form
 +-- Version 1  (schema A)
 +-- Version 2  (schema B)
 +-- Version 3  (schema C)


Each submission stores the `form_version_id` it was validated against. If version 3 is published today, a submission made under version 1 still points to version 1. The invariant is:

```text
submission.form_version_id == the exact published version used to validate it
```

This stays true no matter how many times the form is republished. Submission data is never reinterpreted against a newer version, which keeps historical submissions reproducible and auditable.

### Making the published version explicit

For production, the `forms` table should carry a `current_published_version_id` column, and publishing should update that pointer in one transaction:

```sql
BEGIN;

UPDATE forms
SET current_published_version_id = $version_id,
    updated_at = NOW()
WHERE id = $form_id;

COMMIT;
```

With a single pointer there is never any doubt about which of several historical versions is live.

## 8. Data model

```text
tenants
  id            UUID PRIMARY KEY
  name          TEXT
  created_at    TIMESTAMPTZ

forms
  id            UUID PRIMARY KEY
  tenant_id     UUID  FK -> tenants.id
  name          TEXT
  status        TEXT
  created_at    TIMESTAMPTZ
  updated_at    TIMESTAMPTZ

form_versions
  id              UUID PRIMARY KEY
  form_id         UUID  FK -> forms.id
  version_number  INTEGER
  schema          JSONB
  created_at      TIMESTAMPTZ
  published_at    TIMESTAMPTZ
  UNIQUE (form_id, version_number)

submissions
  id               UUID PRIMARY KEY
  tenant_id        UUID  FK -> tenants.id
  form_id          UUID  FK -> forms.id
  form_version_id  UUID  FK -> form_versions.id
  data             JSONB
  created_at       TIMESTAMPTZ
```

### Indexes

```sql
CREATE INDEX idx_forms_tenant_id
ON forms(tenant_id);

CREATE INDEX idx_form_versions_form_id
ON form_versions(form_id);

CREATE INDEX idx_submissions_form_created
ON submissions(form_id, created_at DESC, id DESC);

CREATE INDEX idx_submissions_tenant_form_created
ON submissions(tenant_id, form_id, created_at DESC, id DESC);
```

Submission listings are filtered by form and ordered newest first, so the two submission indexes match that query's shape and let PostgreSQL read recent rows straight from the index.

### Connection pooling

The API uses `pgxpool`, which reuses connections and caps how many are open at once. The pool size has to be set with the whole deployment in mind: the number of API replicas and worker replicas multiplied by their pool sizes must stay under the database's connection limit.

## 9. Pagination and export

### Pagination

The working slice uses `limit` and `offset`:

```http
GET /api/forms/{form_id}/submissions?limit=20&offset=0
```

Offset pagination gets slower as the offset grows, because PostgreSQL still has to walk past every skipped row. For large submission tables, production should switch to keyset pagination:

```sql
WHERE form_id = $1
AND (created_at, id) < ($cursor_time, $cursor_id)
ORDER BY created_at DESC, id DESC
LIMIT $2;
```

### Export

A large export should not run inside an API request. The design is a background job: the API records an export job and queues it, a worker streams the submissions in batches into a CSV, uploads the file to object storage, and the user receives a signed download URL. Streaming in batches keeps memory use flat regardless of export size.

This is designed, not built.

## 10. Tenant isolation

Every tenant-owned resource carries a `tenant_id`, either directly (`forms.tenant_id`, `submissions.tenant_id`) or through its parent.

Management APIs must take the tenant from the authenticated context, never from a value the client sends, and every query must be scoped to it:

```sql
SELECT *
FROM forms
WHERE id = $1
AND tenant_id = $authenticated_tenant_id;
```

That is what stops one tenant from reading another tenant's forms or submissions.

The current slice passes tenant IDs explicitly to keep things simple. Production authentication and authorization middleware is designed but not fully implemented.

## 11. Security

**Which endpoints need authentication.** The admin endpoints should require it:


POST /api/forms
POST /api/forms/{id}/versions
POST /api/forms/{id}/versions/{version_id}/publish
GET  /api/forms/{id}/submissions


The public endpoints do not, but they need abuse protection instead:


GET  /public/forms/{id}
POST /public/forms/{id}/submissions


**XSS.** Form labels, descriptions, option labels, and submitted values all come from users. Anything that renders them must escape HTML, and user-controlled values must never be injected through unsafe template operations.

**SQL injection.** Every query is parameterized. No user input is concatenated into SQL.


query := `
    SELECT id, name
    FROM forms
    WHERE id = $1
`

db.QueryRow(ctx, query, formID)


**Payload limits.** The server should cap the request body size, the number of fields, field name length, string length, and schema size, so that an oversized payload is rejected early.

**Rate limiting.** Public submission endpoints should be rate limited through Redis so the limit holds across all API instances. A key such as `rate_limit:form:{form_id}:{client_key}` with a configurable limit (for example 100 requests per minute per form per IP) covers both per-form and per-client abuse. Designed, not built.

**Further abuse protection for production.** CAPTCHA or a challenge for suspicious traffic, a WAF, and bot detection.

## 12. Technology choices

| Technology | Why |
|---|---|
| Go | Good performance and concurrency with low memory use, a single binary to deploy, and a strong NATS client. It suits both the HTTP API and the workers. |
| Echo | A lightweight HTTP framework with simple routing and middleware. A good size for this project. |
| PostgreSQL | Transactions, foreign keys, mature indexing, and JSONB. It handles both the relational metadata and the dynamic data. |
| JSONB | Form schemas and submissions differ per form. A column per user-defined field would not be manageable. |
| NATS JetStream | Durable messages, acknowledgements, redelivery, and replay, with a lighter operational footprint than Kafka. |
| Redis | Fast key/value operations for rate limiting and short-lived caching. Deliberately not a system of record. |
| Nginx | A simple reverse proxy and load balancer. |

The comparison of JetStream with RabbitMQ, Kafka, and Redis Streams, and of PostgreSQL JSONB with MongoDB and per-form tables, is in [TRADEOFFS.md](TRADEOFFS.md).

Across these choices I favoured correctness, durability, a simple operational model, and horizontal scalability over maximum throughput. The stack shows the required properties without infrastructure the project does not need.

## 13. Scaling, observability, and testing

### Scaling

- **API.** Add replicas behind the load balancer. No submission state lives in API memory.
- **Workers.** Add workers on the same durable consumer. They scale independently of the API.
- **PostgreSQL.** Start with the right indexes, connection pooling, keyset pagination, and batch processing. At larger scale, add read replicas, partition the submissions table, archive old data, and move analytics and export workloads off the primary.
- **JetStream.** Tune stream configuration, retention policy, and storage size as volume grows.

### Observability

A production deployment should expose:

- **Metrics:** HTTP request count and latency, validation failures, rate-limit rejections, JetStream publish failures, consumer lag, worker processing latency, PostgreSQL errors, and submission success and failure counts.
- **Structured logs** carrying `request_id`, `tenant_id`, `form_id`, `submission_id`, and `form_version_id`. Submitted field values should not be logged, since they may be sensitive.
- **Health checks:** `GET /health` for liveness and `GET /ready` for readiness, where readiness verifies the required dependencies.

### Testing

- **Unit tests** for dynamic validation: required fields, invalid email, number minimum and maximum, and string length.
- **Version integrity test:** create and publish version 1, submit, create and publish version 2, then verify the earlier submission still points to version 1.
- **Integration tests** running the API, PostgreSQL, and NATS together.
- **Load test:** a Go load generator sends a burst at `POST /public/forms/{id}/submissions` (for example 1,000 submissions from 50 concurrent workers) and reports total requests, successes, failures, throughput, and duration. It sends a burst on purpose, because the point is to show that ingestion and persistence are decoupled.

## 14. What is built and what is only designed

### Built

- Go/Echo API
- PostgreSQL connection and migrations
- Tenant creation
- Form creation, form version creation, and publishing
- Public form retrieval
- Submission API
- JetStream publishing and the submission worker
- Durable submission storage in PostgreSQL
- Submission retrieval and listing
- Duplicate protection using the submission ID
- Dockerized PostgreSQL and NATS JetStream

### Designed, not fully built

- Redis distributed rate limiting
- Production authentication and authorization
- CAPTCHA and bot protection, WAF
- The export job pipeline and object-storage exports
- Production cloud load balancer
- Metrics and advanced monitoring
- Keyset pagination (the slice uses offset)
- Database partitioning and read replicas
- Dead-letter queue workflow
- Multi-region deployment
- A full frontend form builder

I have kept these two lists separate so that this document does not claim anything that was not implemented.

## 15. Summary

The platform is split into stages: form management, versioned form definitions, public ingestion, durable messaging, asynchronous persistence, and querying. That separation is what lets it absorb burst traffic, keep public load off PostgreSQL, hold on to every accepted submission, and preserve which form version each submission belongs to.

The working implementation proves the critical path end to end. The rest of this document describes what a larger production deployment would add.
