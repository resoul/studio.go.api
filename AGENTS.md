# AGENTS.md: Architectural Roles and Boundaries

This document defines the responsibilities for each module in the project.
We follow **Clean Architecture** (Ports & Adapters) principles — business logic stays decoupled from
external tools like Ory, Gin, or GORM.

---

## 1. Project Structure

```text
.
├── cmd/
│   ├── root.go             # Cobra root command, WaitGroup injection
│   ├── serve.go            # HTTP server lifecycle — no route wiring, no business logic
│   └── migrate.go          # Migration CLI (up / down), never shares serve-time container
├── internal/
│   ├── domain/             # Entities, Repository Interfaces, Service Interfaces, Ports
│   │   ├── errors.go       # Sentinel errors (ErrNotFound, ErrConflict, …)
│   │   ├── events.go       # Message contracts (InviteEvent) — single source of truth
│   │   ├── mailer.go       # Mailer port (MailMessage, Mailer interface)
│   │   ├── profile.go
│   │   ├── user.go
│   │   └── workspace.go
│   ├── service/            # Business Logic (Use Cases)
│   │   ├── profile_service.go
│   │   └── workspace_service.go
│   ├── worker/             # Async background consumers (RabbitMQ)
│   │   └── invite_worker.go
│   ├── infrastructure/     # External adapters
│   │   ├── db/
│   │   │   ├── migrations/
│   │   │   ├── profile_repository.go
│   │   │   └── workspace_repository.go
│   │   ├── mailer/
│   │   │   └── smtp.go
│   │   ├── ory/
│   │   │   └── kratos.go
│   │   ├── rabbitmq/
│   │   │   └── client.go
│   │   └── storage/
│   │       └── minio.go
│   ├── di/
│   │   └── container.go    # Composition root — holds all singletons
│   ├── config/
│   │   └── config.go       # Env-based config (envconfig + godotenv)
│   └── transport/
│       └── http/
│           ├── router/
│           │   └── router.go       # All route wiring — single source of truth
│           ├── handlers/
│           │   ├── profile_handler.go
│           │   ├── workspace_handler.go  # Workspace CRUD + config
│           │   ├── invite_handler.go     # Invite lifecycle
│           │   └── member_handler.go     # Member management
│           ├── middleware/
│           │   └── auth.go
│           └── utils/
│               ├── cors.go
│               ├── errors.go       # MapError + RespondMapped
│               ├── respond.go
│               └── response.go
├── pkg/
├── configs/
└── deployments/
```

---

## 2. Layer Responsibilities

### A. CLI & Entry Point (`cmd/`)

- `serve.go` — initialises `di.Container`, wires repositories → services → handlers, starts the
  router and HTTP server, starts background workers in goroutines. **No route definitions here.**
- `migrate.go` — opens a raw GORM connection and runs `gormigrate` commands; never shares the
  serve-time container.
- **No business logic in `cmd/`** — delegate everything to the Service layer.

### B. Domain Layer (`internal/domain/`)

The "language" of the project. **No external dependencies allowed here.**

| File | Contents |
|------|----------|
| `errors.go` | Sentinel errors: `ErrNotFound`, `ErrConflict`, `ErrForbidden`, `ErrUnauthorized`, `ErrInvalidInput`, `ErrInviteExpired`, `ErrOwnerCannotBeRemoved` |
| `events.go` | `InviteEvent` — the message contract between service and worker |
| `mailer.go` | `MailMessage`, `Mailer` interface |
| `profile.go` | `Profile` entity, `ProfileRepository`, `ProfileService`, `UpdateProfileInput` |
| `user.go` | `User` entity, `UserRepository` |
| `workspace.go` | `Workspace`, `WorkspaceMember`, `WorkspaceInvite`, `UserWorkspaceConfig`, repository/service interfaces, all input structs, `Storage` port |

**Hard rules:**
- **PROHIBITED:** `map[string]interface{}` or `map[string]any` for entities or API payloads. Use named structs.
- **PROHIBITED:** importing from `service/`, `worker/`, `transport/`, or `infrastructure/`.

### C. Service Layer (`internal/service/`)

Business logic. Coordinates domain entities and repository/port interfaces.

**Hard rules:**
- Must NOT know about Gin, SQL, Ory SDKs, or SMTP internals.
- Returns `domain` sentinel errors (e.g. `domain.ErrNotFound`) — never raw strings that embed
  status codes.
- May depend on `*rabbitmq.Client` for event publishing (messaging bus, not a data store).
- Must NOT import from `worker/` or `transport/`.

#### WorkspaceService
- `InviteUser` — persists the invite, publishes a `domain.InviteEvent` to `workspace.invites`.
- `RemoveMember` — returns `domain.ErrOwnerCannotBeRemoved` when target is the workspace owner.
- If `*rabbitmq.Client` is `nil`, invite is saved but no event is published (graceful degradation).

### D. Worker Layer (`internal/worker/`)

Long-running background goroutines consuming RabbitMQ queues.

**Hard rules:**
- Started by `serve.go`, respect `context.Context` cancellation for graceful shutdown.
- Depend only on **Repository** and **Infrastructure** interfaces — never on handlers or services.
- Must NOT import from `transport/` or `service/`.
- A worker failure (e.g. email delivery) must never crash the process — log, nack, continue.

#### InviteWorker
- Consumes `workspace.invites`.
- Unmarshals `domain.InviteEvent` (single source of truth in `domain/events.go`).
- On send success: `msg.Ack`.
- On send failure: `msg.Nack(requeue=true)` — retry.
- On malformed payload: `msg.Nack(requeue=false)` — dead-letter.

### E. Transport Layer (`internal/transport/http/`)

#### Router (`router/router.go`)
Single source of truth for all route definitions. Receives handler structs and `*config.Config`
via constructor; returns a ready `*gin.Engine`. Adding a new endpoint means editing only this file.

#### Handlers

Split by resource to keep files small and focused:

| File | Routes |
|------|--------|
| `profile_handler.go` | `GET /user/me`, `PATCH /user/profile` |
| `workspace_handler.go` | Workspace CRUD, current workspace, config |
| `invite_handler.go` | Preview, accept, create, list, resend, revoke invites |
| `member_handler.go` | List members, remove member |

**Handler contract:**
1. Extract identity from context.
2. Bind and validate request (JSON or multipart).
3. Call the service method.
4. On error → `utils.RespondMapped(c, err)` — never hand-code status codes for domain errors.
5. On success → `utils.RespondOK` / `utils.RespondCreated` / `c.Status(204)`.

#### Error Mapping (`utils/errors.go`)

`MapError(err) HTTPError` converts domain sentinel errors to HTTP status codes:

| Domain error | HTTP status |
|---|---|
| `ErrNotFound` / `gorm.ErrRecordNotFound` | 404 |
| `ErrConflict` | 409 |
| `ErrUnauthorized` | 401 |
| `ErrForbidden` | 403 |
| `ErrInvalidInput` | 400 |
| `ErrInviteExpired` | 410 |
| `ErrOwnerCannotBeRemoved` | 422 |
| anything else | 500 |

Use `utils.RespondMapped(c, err)` in handlers — never duplicate this table in handler code.

#### Middleware (`middleware/auth.go`)
- Validates `ory_kratos_session` cookie via Kratos `FrontendAPI`.
- Injects `*ory.Identity` into `gin.Context["user"]`.
- Rejects unverified identities with `403 Forbidden`.

#### Utils
- `cors.go` — origin-whitelist CORS middleware.
- `respond.go` / `response.go` — `RespondOK`, `RespondError`, `RespondCreated`, `ErrorResponse`.
- `errors.go` — `MapError`, `RespondMapped`.

### F. Infrastructure Layer (`internal/infrastructure/`)

#### Database (`db/`)
- Implements `ProfileRepository` and `WorkspaceRepository` via GORM.
- **FORBIDDEN:** `db.AutoMigrate()` inside application startup.
- All schema changes live in `db/migrations/` as explicit `gormigrate` entries.
- Migrations run only via `api migrate up` / `api migrate down`.

#### Mailer (`mailer/smtp.go`)
- One SMTP adapter, one constructor: `New(*config.MailerConfig) (domain.Mailer, error)`.
- Behaviour adapts automatically to config:
    - No `MAILER_USERNAME` → no auth (Mailhog / local dev).
    - `MAILER_USERNAME` set → `smtp.PlainAuth` (STARTTLS providers).
    - `MAILER_PORT=465` → implicit TLS via `crypto/tls`.
- Assembles `multipart/alternative` MIME (HTML + auto-generated plain-text fallback).
- **To switch from Mailhog to a real provider** — update `.env` only; no code changes.

#### Ory (`ory/kratos.go`)
- Wraps the Ory Kratos Admin API SDK.
- Implements `domain.UserRepository` (`FindByID`, `GetIdentity`).
- Uses typed structs mirroring `identity.schema.json` — no raw maps for traits.

#### Storage (`storage/minio.go`)
- Implements `domain.Storage` via MinIO client.
- Manages buckets (`workspaces`, `profiles`): creates them and sets public-read policy on startup.
- `GetPresignedURL` uses `STORAGE_PUBLIC_BASE_URL` (configurable for local vs. cloud).

#### RabbitMQ (`rabbitmq/client.go`)
- Optional — if the broker is unavailable at startup, the container degrades gracefully.
- Exposes `Publish`, `DeclareQueue`, `Consume` behind a mutex-protected channel.
- `Consume` is used exclusively by workers — handlers and services only call `Publish`.

---

## 3. Dependency Injection (`di/Container`)

`Container` is the single composition root.

| Field | Type | Source |
|-------|------|--------|
| `Config` | `*config.Config` | `config.Init(ctx)` |
| `DB` | `*gorm.DB` | `postgres.Open(cfg.DB.DSN)` |
| `Storage` | `domain.Storage` | `storage.NewMinioStorage(cfg)` |
| `Mailer` | `domain.Mailer` | `mailer.New(&cfg.Mailer)` |
| `RabbitMQ` | `*rabbitmq.Client` | optional, degrades gracefully |

Services receive only the dependencies they need — **never the full container**.

- `WorkspaceService` receives `*rabbitmq.Client` (publish) and **not** `domain.Mailer`.
- `InviteWorker` receives `*rabbitmq.Client` (consume), `domain.WorkspaceRepository`, `domain.Mailer`.

---

## 4. Event Contract

`domain.InviteEvent` in `internal/domain/events.go` is the **single source of truth** for the
message published by `WorkspaceService` and consumed by `InviteWorker`.

```go
type InviteEvent struct {
    Token         string `json:"token"`
    WorkspaceID   string `json:"workspace_id"`
    WorkspaceName string `json:"workspace_name"`
    Email         string `json:"email"`
    Role          string `json:"role"`
    ExpiresAt     string `json:"expires_at"` // RFC3339
    InviteBaseURL string `json:"invite_base_url"`
}
```

> When adding a new async event, create its contract struct in `domain/events.go` first,
> then reference it from both the publishing service and the consuming worker.

---

## 5. Invite Flow (Async)

```
Handler (POST /workspaces/:id/invites)
  → WorkspaceService.InviteUser
      → repo.CreateInvite          (persist to DB)
      → rbmq.Publish               (enqueue domain.InviteEvent)
  ← 202 Accepted

InviteWorker (goroutine)
  → rbmq.Consume (workspace.invites)
      → json.Unmarshal → domain.InviteEvent
      → mailer.Send
      → msg.Ack / msg.Nack
```

---

## 6. Communication Patterns

1. **Strong Typing Everywhere** — all inter-layer data uses domain structs. No `map[string]any`.
2. **Dependency Injection** — all dependencies passed via `New…` constructors.
3. **Context Propagation** — `context.Context` threaded from handler down to GORM, Mailer, workers.
4. **Sentinel Errors** — services return `domain.Err*` values; the transport layer maps them to
   HTTP codes via `utils.MapError`. Never hand-code HTTP status codes for domain conditions.
5. **Graceful Degradation** — optional infrastructure (RabbitMQ) must not prevent startup.
   Workers start only when `container.RabbitMQ != nil`.
6. **Async Side Effects** — operations touching external systems (email, future webhooks) are
   published to a queue and handled by workers — never blocking the HTTP response.

---

## 7. Adding a New Resource (Checklist)

When introducing a new domain object (e.g. `Project`, `Asset`):

- [ ] Add entity + repository interface + service interface to `internal/domain/<resource>.go`
- [ ] Add sentinel errors to `domain/errors.go` if new failure modes are needed
- [ ] Add message contracts to `domain/events.go` if async events are needed
- [ ] Implement repository in `internal/infrastructure/db/<resource>_repository.go`
- [ ] Add migration in `internal/infrastructure/db/migrations/`
- [ ] Implement service in `internal/service/<resource>_service.go`
- [ ] Add handler file(s) in `internal/transport/http/handlers/<resource>_handler.go`
- [ ] Register routes in `internal/transport/http/router/router.go`
- [ ] Wire in `cmd/serve.go`
- [ ] Extend `utils/errors.go` mapper if new sentinel errors were added

---

## 8. Configuration Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_DSN` | — | PostgreSQL DSN (required) |
| `MAILER_FROM` | `no-reply@studio.localhost` | Envelope sender |
| `MAILER_HOST` | `localhost` | SMTP hostname |
| `MAILER_PORT` | `1025` | `1025`=Mailhog, `587`=STARTTLS, `465`=TLS |
| `MAILER_USERNAME` | — | Empty = no auth (Mailhog) |
| `MAILER_PASSWORD` | — | SMTP password |
| `STORAGE_PUBLIC_BASE_URL` | — | Public URL prefix for asset links |
| `RABBITMQ_URL` | `amqp://guest:guest@localhost:5672/` | Optional; omit to disable |
| `SERVER_DASHBOARD_URL` | `http://dashboard.studio.localhost` | Invite link base URL |

---

## 9. Technology Stack

| Concern | Tool |
|---------|------|
| CLI | [Cobra](https://github.com/spf13/cobra) |
| Web Framework | [Gin Gonic](https://github.com/gin-gonic/gin) |
| ORM | [GORM](https://gorm.io/) |
| Migrations | [gormigrate](https://github.com/go-gormigrate/gormigrate) |
| Auth | [Ory Kratos](https://www.ory.sh/kratos/) |
| Database | PostgreSQL |
| Storage | MinIO (S3-compatible) |
| Mailer | SMTP (`net/smtp` + `crypto/tls`) |
| Message Queue | RabbitMQ via [amqp091-go](https://github.com/rabbitmq/amqp091-go) |