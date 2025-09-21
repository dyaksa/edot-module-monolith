# Warehouse Service (Module Monolith) – Go + FastHTTP

Production‑ready Go service implementing a Module Monolith architecture layered with Clean Architecture principles. It balances the simplicity of a monolith with modular boundaries that make it easy to scale, test, and (selectively) extract modules later.

---

## Why a Module Monolith?

Instead of scattering logic across microservices prematurely, this codebase groups cohesive business capabilities into explicit modules (auth, product, order, stock/reservation, warehouse, transfer) while sharing a single deployable binary and database. Each module exposes:

- Domain contracts (entities + repository/usecase interfaces) in `domain`
- Application services in `usecase`
- Delivery (HTTP controllers/routes) in `api`
- Persistence implementations in `repository`

Clear interface boundaries let you:

- Enforce dependency direction (domain → usecase → adapters)
- Unit test modules in isolation using generated mocks (Mockery)
- Defer microservice extraction until it provides clear value

---

## High‑Level Architecture

| Layer              | Folders                          | Responsibility                                   |
| ------------------ | -------------------------------- | ------------------------------------------------ |
| Domain (Core)      | `domain/`                        | Entities, value objects, interfaces (pure Go)    |
| Use Cases          | `usecase/`                       | Orchestrate business rules & transactions        |
| Interface Adapters | `api/`, `repository/`, `worker/` | HTTP/controllers, DB persistence, async workers  |
| Infrastructure     | `infrastructure/`, `bootstrap/`  | DB clients, crypto, env, app wiring              |
| Shared Kernel      | `pkg/`                           | Reusable helpers (logging, pagination, response) |
| Tests & Mocks      | `mocks/`, `usecase/*_test.go`    | Auto‑generated repository mocks + unit tests     |

Dependency rule: Outer layers depend inward; inner layers have no outward knowledge.

---

## Module Breakdown

| Module      | Key Concepts / Responsibilities               |
| ----------- | --------------------------------------------- |
| Auth        | User registration, login, JWT creation        |
| User        | User entity + credential hashing via crypto   |
| Shop        | Shop registration and management              |
| Product     | SKU + stock entry creation, availability view |
| Stock       | Reservation, release, commit, movements       |
| Order       | Checkout, idempotency, order items linkage    |
| Warehouse   | Physical storage locations (activation state) |
| Transfer    | Inter‑warehouse stock movement lifecycle      |
| Idempotency | Safe replay protection for mutative endpoints |

---

## Domain Flows (Summaries)

1. Checkout

   1. Validate items & user
   2. Idempotency key (begin + store response on commit)
   3. Reserve stock → create order + items → commit reservations → append movements

2. Stock Release (Scheduled/Worker)

   - Scan expired reservations → release stock → append reversal movements

3. Warehouse Transfer

   - REQUESTED → (APPROVED) → IN_TRANSIT (reserve + outbound + commit) → COMPLETED (inbound + add stock)
   - Guard: cannot deactivate warehouse with active transfers

4. Product Creation
   - Create product row → initialize stock record in selected warehouse

---

## Project Structure

```txt
api/            # Controllers, middleware, routes
bootstrap/      # App assembly (env, db, wiring)
domain/         # Entities + interface definitions
infrastructure/ # DB clients, crypto, Kafka (if any), etc.
migrations/     # SQL schema evolution
mocks/          # Auto-generated mocks (Mockery)
pkg/            # Shared utilities (log, paginator, response, etc.)
repository/     # Persistence implementations (Postgres)
usecase/        # Application services (business orchestration)
```

---

## Local Development

### Prerequisites

- Go >= 1.23.x (1.24 ready)
- PostgreSQL
- Docker + Docker Compose (optional but recommended)

### Environment

```bash
cp .env.example .env
# Adjust DB credentials, JWT secret, rate limits, etc.
```

### Database Migrations

Migrations live in `migrations/` and are applied sequentially. If you integrate a tool like `golang-migrate`, add a Make target (see Enhancements section).

### Run (Direct)

```bash
make run
```

### Run (Docker Dev)

```bash
make setup-dev        # builds & starts services
# or manually
docker-compose up -d --build
```

### Production Compose

```bash
make compose-prod
```

### Build Binary

```bash
make build
```

---

## Testing

### Unit Tests

All use case tests reside beside implementations in `usecase/*_test.go` and rely on Mockery generated repository mocks under `mocks/repository`.

Run everything:

```bash
make test
# or
go test ./...
```

Target a single area:

```bash
go test ./usecase -run OrderUsecase
```

### Mock Generation (Mockery v3)

Configuration is handled via `.mockery.yml` (minimal v3-compliant). Regenerate mocks after interface changes:

```bash
mockery --all --with-expecter
```

> Note: Only supported YAML keys are used (v3 removed several v2-era fields). Additional features must be passed via CLI flags.

### Test Design Highlights

- Fake transactional DBs simulate `Transaction` boundaries without a real driver.
- Idempotency flow asserts repository interactions & stored responses.
- Transfer + stock release flows validate multi-step state transitions & failure paths.

---

## Logging & Observability

- Structured logging via custom wrapper in `pkg/log` (Logrus under the hood).
- Prometheus config placeholder in `docker/prometheus.yml` (extend to instrument app).
- Extend with OpenTelemetry exporters later without leaking infra concerns into domain.

---

## Security & Crypto

- `bootstrap/crypto.go` wires encryption + hashing utilities.
- JWT creation & parsing in `domain/jwt_custom.go` and helpers under `pkg/tokenutils`.
- Password hashing helpers under `pkg/passwordutils`.

---

## Rate Limiting & Middleware

Located in `api/middleware/`:

- `jwt_auth_middleware.go` – AuthN/JWT validation
- `ratelimit_middleware.go` – Request throttling (token bucket style)

Add global / route-scoped middleware in `api/route/route.go`.

---

## Data Model Overview (Selected Tables)

Migration files (incremental):

- Users / Shops / Warehouses
- Products, Product Stock, Stock Movements
- Orders, Order Items
- Reservations & Reservation Expiry logic
- Warehouse Transfers (+ items)
- Idempotency Requests

Suggested indexing: composite indexes for frequently filtered columns (e.g. `(product_id, warehouse_id)` for stock; `(warehouse_id, status)` for transfers).

---

## Extending a Module (Example: Adding Cancellation to Transfers)

1. Add method signature to `domain.WarehouseTransferUsecase`
2. Implement in `usecase/warehouse_transfer_usecase.go`
3. Add repository method if persistence changes required
4. Generate mocks (`mockery ...`)
5. Add unit tests first (!)
6. Wire endpoint in `api/route/` + controller

---

## Make Targets (Common)

| Target              | Purpose                        |
| ------------------- | ------------------------------ |
| `make run`          | Run locally (go run)           |
| `make build`        | Compile binary                 |
| `make test`         | Run all tests                  |
| `make setup-dev`    | Docker compose dev environment |
| `make compose-prod` | Start production services      |

> Add future: `make migrate-up`, `make migrate-down` once migration tool integrated.

---

## Potential Enhancements (Roadmap)

| Area          | Suggestion                                                            |
| ------------- | --------------------------------------------------------------------- |
| Observability | Add OpenTelemetry tracing & metrics wrappers                          |
| Migrations    | Introduce `golang-migrate` with Make targets                          |
| Validation    | Introduce request DTO validation layer (e.g. go-playground/validator) |
| Caching       | Add Redis for frequently accessed read models                         |
| Messaging     | Kafka patterns (outbox, async stock sync) already scaffolded dir wise |
| CI/CD         | GitHub Actions pipeline (lint, vet, test, build)                      |
| Security      | Add structured audit logging for stock mutations                      |
| Resilience    | Introduce circuit breaker / retry policy wrappers                     |

---

## Guidelines for Contributions

1. Keep domain interfaces stable; evolve via additive changes.
2. Never import outward (e.g. `usecase` must not import `api`).
3. Prefer pure functions in `domain` & thin orchestration in `usecase`.
4. Always add/update tests when changing business logic.
5. Keep mocks regenerated and committed when interfaces change.

---

## License

MIT License

---

## At a Glance (Quick Start)

```bash
cp .env.example .env
make setup-dev   # or: make run
go test ./...    # ensure green
mockery --all --with-expecter   # when interfaces change
```

Happy building — this module monolith is designed to grow with you until (and if) you truly need to slice it into services.

---

## Error Taxonomy & Central HTTP Mapping

This project uses a structured error model (`pkg/errx`) plus a centralized HTTP translation middleware.

### Goals

- Consistent machine-readable error codes
- Single place to map codes -> HTTP status
- Separation of internal messages vs public payload (extensible via `PublicMessage`)
- Composable wrapping with operations and metadata

### Core Types

`errx.AppError` fields:

- `Code` (`errx.Code`): stable enum string (e.g. `NOT_FOUND`, `VALIDATION_ERROR`)
- `Message`: developer-oriented message (also sent publicly for now)
- `Op`: operation or logical function boundary (e.g. `UserUsecase.Login`)
- `Err`: wrapped underlying error (not serialized)
- `Meta`: optional contextual key/values (serialized)
- `Transient`: hint for retry semantics (not serialized)

### Defined Codes

```
INVALID_ARGUMENT
VALIDATION_ERROR
NOT_FOUND
ALREADY_EXISTS
UNAUTHENTICATED
UNAUTHORIZED
PERMISSION_DENIED
CONFLICT
RATE_LIMITED
PRECONDITION_FAILED
INTERNAL_ERROR
SERVICE_UNAVAILABLE
TIMEOUT
```

### HTTP Mapping

Handled in `errx.HTTPStatus`:

- 400: INVALID_ARGUMENT, VALIDATION_ERROR
- 401: UNAUTHENTICATED, UNAUTHORIZED
- 403: PERMISSION_DENIED
- 404: NOT_FOUND
- 409: ALREADY_EXISTS, CONFLICT
- 412: PRECONDITION_FAILED
- 429: RATE_LIMITED
- 500: INTERNAL_ERROR (default)
- 503: SERVICE_UNAVAILABLE
- 504: TIMEOUT

### Producing Errors

Use the ergonomic constructor:

```go
return errx.E(errx.CodeValidation, "email invalid", errx.Op("UserUsecase.Register"), underlyingErr)
```

Add metadata:

```go
return errx.E(errx.CodeConflict, "user already exists").WithMeta("email", input.Email)
```

### Controller Pattern

Instead of writing JSON directly for failures:

```go
if err != nil {
      c.Error(err) // controller returns; middleware serializes
      return
}
```

If you need to wrap on boundary:

```go
if err != nil {
      c.Error(errx.E(errx.CodeInternal, "failed to persist user", errx.Op("UserController.Create"), err))
      return
}
```

### Middleware Registration

Add `ErrorMiddleware` early in the Gin stack:

```go
r := gin.New()
r.Use(middleware.LoggerMiddleware(logger))
r.Use(middleware.ErrorMiddleware(logger))
```

### Response Shape

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "invalid request payload",
    "meta": { "field": "email" }
  }
}
```

### Testing

`pkg/errx/httpmap_test.go` ensures code->status coverage.

### Migration Notes

- Existing `response_error` / `response_success` can coexist; migrate gradually.
- For legacy endpoints wrap: `c.Error(errx.ToAppError(err))`.
- Eventually remove `response_error` after full migration.

### Future Enhancements

- Internationalization of public messages via `PublicMessage`
- Structured logging correlation IDs in `Meta`
- gRPC status mapping if needed
