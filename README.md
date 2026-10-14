# Webform Builder

A multi-tenant backend for building web forms and collecting submissions.

An admin creates tenants. Each tenant creates forms, defines the fields as a JSON schema, and publishes a version. Anyone can then fetch the published form and submit data to it without logging in. The server validates every submission against the published schema, queues it in NATS JetStream, and a background worker stores it in PostgreSQL.

For the design and the reasoning behind it, see [ARCHITECTURE.md](ARCHITECTURE.md) and [TRADEOFFS.md](TRADEOFFS.md).

## What is in the box

| Part | What it does |
|---|---|
| API (`cmd/api`) | Go + Echo HTTP server. Handles login, tenants, forms, versions, and submissions. Runs the database migrations on startup. |
| Worker (`cmd/worker`) | Reads submissions from JetStream and writes them to PostgreSQL. |
| PostgreSQL 16 | Stores tenants, forms, form versions, and submissions. |
| NATS JetStream | Durable queue between the API and the worker. |
| Nginx | Reverse proxy in front of the API. This is the port you call. |

## Running it

### With Docker (recommended)

You need Docker with Docker Compose.

1. Create a `.env` file in the project root:

   ```env
   JWT_SECRET=change-me-to-a-random-string-of-32-chars-or-more
   JWT_EXPIRY_HOURS=24
   ADMIN_EMAIL=admin@example.com
   ADMIN_PASSWORD=change-me
   ```

   `JWT_SECRET` must be at least 32 characters, or the API refuses to start.

2. Start everything:

   ```bash
   docker compose up --build -d
   ```

3. Check that it is up:

   ```bash
   curl http://localhost/healthz
   ```

   You should get `{"status":"ok"}`.

The API is now available at `http://localhost`. If port 80 is taken, add `NGINX_PORT=8080` to `.env` and use `http://localhost:8080`.

To stop it, run `docker compose down`. Add `-v` to also delete the database and queue data.

### Without Docker for the Go code

You need Go 1.26 or newer. PostgreSQL and NATS still run in Docker.

1. Start the two dependencies:

   ```bash
   docker compose up -d postgres nats
   ```

2. Add the connection settings to the same `.env` file:

   ```env
   DATABASE_URL=postgres://webform:webform@localhost:5432/webform?sslmode=disable
   NATS_URL=nats://localhost:4222
   PORT=8080
   ```

3. Run the API and the worker in two terminals, from the project root:

   ```bash
   go run ./cmd/api
   ```

   ```bash
   go run ./cmd/worker
   ```

The API is now at `http://localhost:8080`. Start the API first, because it creates the tables.

### Settings

| Variable | Required | Meaning |
|---|---|---|
| `DATABASE_URL` | yes | PostgreSQL connection string. Set for you in Docker. |
| `NATS_URL` | yes | NATS address. Set for you in Docker. |
| `PORT` | no | API port. Defaults to `8080`. |
| `JWT_SECRET` | yes | Secret used to sign login tokens. At least 32 characters. |
| `JWT_EXPIRY_HOURS` | no | How long a token is valid. Defaults to `24`. |
| `ADMIN_EMAIL` | yes | Email of the single admin account. |
| `ADMIN_PASSWORD` | yes | Password of the admin account. |
| `NGINX_PORT` | no | Host port for Nginx in Docker. Defaults to `80`. |

## How login works

There are two kinds of users:

- **Admin.** One account, set by `ADMIN_EMAIL` and `ADMIN_PASSWORD`. The admin can create and list tenants and can access every tenant's data.
- **Tenant.** Created by the admin with an email and password. A tenant can only access its own forms and submissions.

Logging in returns a token. Send it on every `/api/...` request:

```http
Authorization: Bearer <token>
```

The `/public/...` endpoints and `/healthz` need no token.

## A quick walkthrough

These commands take you from nothing to a stored submission. They assume the Docker setup, so the base URL is `http://localhost`. Replace the `<...>` placeholders with the values from earlier responses.

**1. Log in as admin**

```bash
curl -X POST http://localhost/api/auth/admin/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"change-me"}'
```

Copy `token` from the response.

**2. Create a tenant**

```bash
curl -X POST http://localhost/api/tenants \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Acme","email":"owner@acme.com","password":"acme-pass-123"}'
```

Copy `id` from the response. This is the tenant ID.

**3. Log in as the tenant**

```bash
curl -X POST http://localhost/api/auth/tenant/login \
  -H "Content-Type: application/json" \
  -d '{"email":"owner@acme.com","password":"acme-pass-123"}'
```

**4. Create a form**

```bash
curl -X POST http://localhost/api/tenants/<tenant_id>/forms \
  -H "Authorization: Bearer <tenant_token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Contact us"}'
```

**5. Add a version with the form's fields**

```bash
curl -X POST http://localhost/api/forms/<form_id>/versions \
  -H "Authorization: Bearer <tenant_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "schema": {
      "fields": [
        {"name":"name",  "label":"Full name", "type":"text",   "required":true},
        {"name":"email", "label":"Email",     "type":"email",  "required":true},
        {"name":"age",   "label":"Age",       "type":"number", "required":false}
      ]
    }
  }'
```

**6. Publish that version**

```bash
curl -X POST http://localhost/api/forms/<form_id>/versions/<version_id>/publish \
  -H "Authorization: Bearer <tenant_token>"
```

**7. Fetch the public form (no token)**

```bash
curl http://localhost/public/forms/<form_id>
```

**8. Submit to it (no token)**

```bash
curl -X POST http://localhost/public/forms/<form_id>/submissions \
  -H "Content-Type: application/json" \
  -d '{"data":{"name":"Asha","email":"asha@example.com","age":29}}'
```

You get `202 Accepted`. The worker stores the submission a moment later.

**9. Read the submissions**

```bash
curl "http://localhost/api/forms/<form_id>/submissions?limit=20&offset=0" \
  -H "Authorization: Bearer <tenant_token>"
```

## API reference

All request and response bodies are JSON. Errors look like `{"error":"message"}`.

### Summary

| Method | Path | Who can call it | Purpose |
|---|---|---|---|
| GET | `/healthz` | anyone | Health check |
| POST | `/api/auth/admin/login` | anyone | Log in as admin |
| POST | `/api/auth/tenant/login` | anyone | Log in as a tenant |
| GET | `/api/auth/me` | admin, tenant | Show who the token belongs to |
| POST | `/api/tenants` | admin | Create a tenant |
| GET | `/api/tenants` | admin | List tenants |
| GET | `/api/tenants/:id` | admin, that tenant | Get one tenant |
| POST | `/api/tenants/:tenant_id/forms` | admin, that tenant | Create a form |
| GET | `/api/tenants/:tenant_id/forms` | admin, that tenant | List a tenant's forms |
| GET | `/api/forms/:id` | admin, owning tenant | Get one form |
| POST | `/api/forms/:id/versions` | admin, owning tenant | Create a form version |
| GET | `/api/forms/:id/versions` | admin, owning tenant | List a form's versions |
| POST | `/api/forms/:id/versions/:version_id/publish` | admin, owning tenant | Publish a version |
| GET | `/public/forms/:id` | anyone | Get the published form |
| POST | `/public/forms/:id/submissions` | anyone | Submit the form |
| GET | `/api/forms/:id/submissions` | admin, owning tenant | List a form's submissions |
| GET | `/api/submissions/:id` | admin, owning tenant | Get one submission |

A missing or invalid token returns `401`. A valid token without access to the resource returns `403`.

### Auth

**`POST /api/auth/admin/login`** and **`POST /api/auth/tenant/login`**

Request:

```json
{ "email": "owner@acme.com", "password": "acme-pass-123" }
```

Response `200`:

```json
{
  "token": "eyJhbGciOi...",
  "token_type": "Bearer",
  "expires_at": "2026-10-14T10:00:00Z",
  "role": "tenant",
  "tenant_id": "7b1c..."
}
```

`tenant_id` is only present for tenant logins. Wrong credentials return `401`.

**`GET /api/auth/me`**

Returns the `role`, `tenant_id`, and `expires_at` of the token you sent.

### Tenants

**`POST /api/tenants`** (admin)

```json
{ "name": "Acme", "email": "owner@acme.com", "password": "acme-pass-123" }
```

The password must be 8 to 72 characters. Returns `201` with the tenant:

```json
{
  "id": "7b1c...",
  "name": "Acme",
  "email": "owner@acme.com",
  "created_at": "2026-10-13T10:00:00Z"
}
```

**`GET /api/tenants?page=1&limit=20`** (admin)

Returns `{ "tenants": [...], "page": 1, "limit": 20, "total": 42 }`.

**`GET /api/tenants/:id`**

Returns one tenant, or `404` if it does not exist.

### Forms

**`POST /api/tenants/:tenant_id/forms`**

```json
{ "name": "Contact us" }
```

Returns `201` with the form. A new form has `"status": "draft"`.

```json
{
  "id": "c4e2...",
  "tenant_id": "7b1c...",
  "name": "Contact us",
  "status": "draft",
  "created_at": "2026-10-13T10:05:00Z",
  "updated_at": "2026-10-13T10:05:00Z"
}
```

**`GET /api/tenants/:tenant_id/forms?page=1&limit=20`**

Returns `{ "forms": [...], "page": 1, "limit": 20, "total": 3 }`, newest first.

**`GET /api/forms/:id`**

Returns the form with its version count and version IDs:

```json
{
  "form": {
    "id": "c4e2...",
    "tenant_id": "7b1c...",
    "name": "Contact us",
    "status": "published",
    "created_at": "...",
    "updated_at": "...",
    "version_count": 2,
    "version_ids": ["9a0f...", "51bd..."]
  }
}
```

### Form versions

A version holds the form's fields. Versions are never edited. To change a form, create a new version and publish it.

**`POST /api/forms/:id/versions`**

```json
{
  "schema": {
    "fields": [
      { "name": "email", "label": "Email", "type": "email", "required": true }
    ]
  }
}
```

Returns `201` with the version. Version numbers start at 1 and go up by one per form.

```json
{
  "id": "9a0f...",
  "form_id": "c4e2...",
  "version_number": 1,
  "schema": { "fields": [ ... ] },
  "created_at": "2026-10-13T10:10:00Z"
}
```

**`GET /api/forms/:id/versions?page=1&limit=20`**

Returns `{ "versions": [...], "page": 1, "limit": 20, "total": 2 }`, highest version number first.

**`POST /api/forms/:id/versions/:version_id/publish`**

No body. Marks the version as published, sets the form's status to `published`, and returns the version with `published_at` filled in.

### Public endpoints

**`GET /public/forms/:id`**

Returns the published version of the form, in the same shape as a version above. If several versions have been published, you get the one with the highest version number. Returns `404` if nothing is published.

**`POST /public/forms/:id/submissions`**

```json
{ "data": { "name": "Asha", "email": "asha@example.com", "age": 29 } }
```

Response `202`:

```json
{
  "message": "submission accepted",
  "submission": {
    "id": "f3a1...",
    "tenant_id": "7b1c...",
    "form_id": "c4e2...",
    "form_version_id": "9a0f...",
    "data": { "name": "Asha", "email": "asha@example.com", "age": 29 },
    "created_at": "2026-10-13T10:20:00Z"
  }
}
```

`202` means the submission is safely queued. It shows up in the read endpoints once the worker has saved it, usually within a moment.

If the data does not match the form, you get `400` with one message per field:

```json
{
  "error": "validation failed",
  "fields": {
    "email": "must be a valid email address",
    "name": "is required",
    "nickname": "is not a field of this form"
  }
}
```

### Submissions

**`GET /api/forms/:id/submissions?limit=20&offset=0`**

Note that this endpoint uses `offset`, not `page`. Newest first.

```json
{
  "form_id": "c4e2...",
  "submissions": [ ... ],
  "limit": 20,
  "offset": 0
}
```

**`GET /api/submissions/:id`**

Returns `{ "submission": { ... } }`, or `404` if it does not exist.

### Paging

`limit` defaults to 20 and is capped at 100 on every list endpoint.

## Form schema and validation

A schema is an object with a `fields` list. Each field has:

| Key | Meaning |
|---|---|
| `name` | The key used in the submitted `data`. |
| `type` | One of the types below. |
| `required` | `true` if the field must be filled in. |
| `label` | Display text. Stored and returned, not validated. |

Supported types:

| Type | Accepted value |
|---|---|
| `text`, `textarea`, `tel`, `phone`, `password` | A string |
| `email` | A valid email address |
| `number` | A number, or a string containing a number |
| `url` | An `http` or `https` URL |
| `date` | A string in `YYYY-MM-DD` format |
| `checkbox`, `boolean` | `true` or `false` |

The server also rejects any submitted field that is not in the schema. Other keys in a field definition, such as `min_length` or `max`, are stored with the schema but are not enforced yet.

## Project layout

```text
cmd/api          API entry point and route registration
cmd/worker       Submission worker entry point
internal/auth        Login, JWT tokens, access middleware
internal/tenant      Tenant endpoints
internal/form        Forms, versions, publishing
internal/submission  Submit, validate, publish to JetStream, read back
internal/messaging   NATS connection and stream setup
internal/database    PostgreSQL connection and migration runner
internal/config      Environment settings
migrations       SQL migrations, applied by the API on startup
nginx            Reverse proxy config
```
