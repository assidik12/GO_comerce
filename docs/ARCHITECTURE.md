# Technical Architecture — Catalyst

## Design Philosophy

Catalyst follows **Clean Architecture** + **Event-Driven Architecture** pattern, 
designed to handle real-world e-commerce constraints:

1. **Data Integrity** — Atomic transactions, no double-charging
2. **Performance** — Multi-layer caching with stampede prevention
3. **Reliability** — Graceful degradation, async fallback, message durability
4. **Maintainability** — Clear separation of concerns, testable layers

---

## Tech Stack

| Category | Technology | Version | Purpose |
|---|---|---|---|
| **Language** | Go | 1.22+ | Core language |
| **Database** | MySQL | 8.0 | Primary data store (source of truth) |
| **Cache** | Redis | 7.0 | Multi-layer caching |
| **Message Broker** | Apache Kafka | 2.8+ | Async event processing |
| **Observability** | Prometheus + Grafana | latest | Metrics & dashboards |
| **Tracing** | OpenTelemetry + Jaeger | latest | Distributed tracing |
| **Container** | Docker + Docker Compose | 20.10+ | Development & deployment |

### Dependencies & Libraries

| Category | Library | Purpose |
|---|---|---|
| **Router** | [`julienschmidt/httprouter`](https://github.com/julienschmidt/httprouter) | High-performance HTTP router |
| **Database** | [`go-sql-driver/mysql`](https://github.com/go-sql-driver/mysql) | MySQL driver for Go |
| **Cache** | [`redis/go-redis`](https://github.com/redis/go-redis) | Redis client for Go |
| **Message Broker** | [`segmentio/kafka-go`](https://github.com/segmentio/kafka-go) | Pure Go Kafka client |
| **Concurrency** | [`golang.org/x/sync`](https://pkg.go.dev/golang.org/x/sync) | Singleflight (cache stampede prevention) |
| **Validation** | [`go-playground/validator`](https://github.com/go-playground/validator) | Struct validation |
| **JWT** | [`golang-jwt/jwt`](https://github.com/golang-jwt/jwt) | JSON Web Token |
| **Config** | [`spf13/viper`](https://github.com/spf13/viper) | Configuration management |
| **DI** | [`google/wire`](https://github.com/google/wire) | Compile-time dependency injection |
| **Migration** | [`golang-migrate`](https://github.com/golang-migrate/migrate) | Database migrations |
| **Password** | [`golang.org/x/crypto`](https://pkg.go.dev/golang.org/x/crypto) | Bcrypt hashing |
| **UUID** | [`google/uuid`](https://github.com/google/uuid) | UUID generation |
| **Metrics** | [`prometheus/client_golang`](https://github.com/prometheus/client_golang) | Prometheus metrics |
| **Tracing** | [`go.opentelemetry.io/otel`](https://pkg.go.dev/go.opentelemetry.io/otel) | OpenTelemetry SDK |
| **Rate Limit** | [`golang.org/x/time`](https://pkg.go.dev/golang.org/x/time) | Rate limiter |
| **Testing** | [`stretchr/testify`](https://github.com/stretchr/testify) | Assertions & mocking |
| **SQL Mock** | [`DATA-DOG/go-sqlmock`](https://github.com/DATA-DOG/go-sqlmock) | SQL driver mock |

---

## Layered Architecture

```
┌─────────────────┐
│  HTTP Handler   │ ← Presentation (DTOs, error mapping)
├─────────────────┤
│    Service      │ ← Business Logic (calculations, validation)
├─────────────────┤
│  Repository     │ ← Data Access (SQL, caching)
├─────────────────┤
│  Infrastructure │ ← External Services (DB, Redis, Kafka)
└─────────────────┘
```

Each layer:
- **Depends only on the layer below** (or domain)
- **Exposes only what's necessary** (via interfaces)
- **Is independently testable** (via mocks)

### Folder Structure

```
catalyst/
├── cmd/
│   ├── api/                 # Main application entrypoint
│   └── injector/            # Dependency injection (Wire)
├── config/                  # Configuration management (Viper, Prometheus, Grafana)
├── db/migrations/           # Database migrations (golang-migrate)
├── docs/                    # Documentation & Swagger specs
├── internal/
│   ├── delivery/
│   │   └── http/
│   │       ├── handler/     # HTTP handlers (Presentation layer)
│   │       ├── dto/         # Data Transfer Objects
│   │       ├── middleware/  # JWT Auth, Metrics, Tracing
│   │       └── route/       # Route definitions
│   ├── domain/              # Business entities, interfaces, sentinel errors
│   ├── dto/                 # Shared DTOs (if needed across layers)
│   ├── event/               # Kafka event structs + producer interface
│   ├── infrastructure/      # External service init (DB, Redis, Kafka)
│   ├── pkg/                 # Shared utilities (response helpers)
│   ├── repository/
│   │   ├── mysql/           # SQL implementation
│   │   └── redis/           # Cache implementation
│   ├── service/             # Business logic layer
│   └── worker/              # Kafka consumers / background workers (outbox relay)
├── test/                    # Unit & integration tests
├── ai_agent/                # AI agent system prompts
├── docker-compose.yml
├── Dockerfile
└── go.mod
```

## Event-Driven Processing

```
User Request → Service validates → DB transaction → Commit
                                        ↓
                        (If successful) → Save to Outbox table (same TX)
                                        ↓
                        Outbox Relay Worker → Publish to Kafka
                                        ↓
                        Async Consumer → Sends email, updates inventory, etc.
```

**Key principles:**
- The transaction commits FIRST, outbox event saved in the **same** transaction.
- Outbox Relay guarantees **at-least-once delivery** to Kafka.
- If Kafka fails, the relay retries — no event is lost.

## Caching Strategy

```
Request → Check Redis
            ↓
         Cache hit? → Return immediately (5-10ms)
            ↓
         Cache miss? → Single flight prevents thundering herd
            ↓
         Query database → Cache result → Return
```

Singleflight prevents 1000 concurrent requests causing 1000 DB hits.
Instead: 1 DB hit, 999 requests wait for result, all get same answer.

## Performance Characteristics

| Scenario | Latency |
|---|---|
| Cached product detail | 5-10ms |
| DB hit (cache miss) | 50-150ms |
| Full transaction | 200-500ms |
| With Kafka publish | <100ms extra (async) |

---

## Transaction Atomicity

### The Problem
```
Transaction started
  ├─ Charge customer Rp500.000
  ├─ Decrement inventory by 1
  └─ Crash here! ← Customer charged but inventory never decremented
```

### The Solution (Catalyst)
```
Database transaction started
  ├─ Lock rows
  ├─ Charge customer Rp500.000
  ├─ Decrement inventory by 1
  ├─ Save outbox event (same TX)
  ├─ All success? → Commit (atomic)
  └─ Any error? → Rollback (everything reverted)
  
Result: No half-state. Either all succeed or nothing happens.
```

---

## Server-Side Validation

### The Problem
```
Client sends: POST /transactions
{
  "products": [{"product_id": 1, "qty": 2}],
  "total_price": 100  // ← Attacker sets to Rp100 instead of Rp1.000.000
}
```

### The Solution (Catalyst)
```go
// Service calculates price from database
product := repo.FindById(1)     // Returns price: Rp500.000
totalPrice := product.Price * qty  // 500.000 × 2 = 1.000.000
// Client's total_price ignored entirely
```

Client cannot manipulate pricing. Server is the source of truth.

---

## Project Status & Roadmap

### Current Status

**Already implemented:**
- ✅ Clean Architecture 4-layer compliance (100%)
- ✅ Atomic stock decrement via conditional SQL UPDATE
- ✅ Graceful shutdown with connection cleanup
- ✅ Async Kafka publish post-commit (via outbox pattern)
- ✅ JWT authentication + RBAC
- ✅ Redis caching + singleflight (cache stampede prevention)
- ✅ Structured logging via `slog`
- ✅ Transactional Outbox Pattern (guaranteed event delivery)
- ✅ Idempotency Key support (duplicate prevention)
- ✅ Prometheus metrics (`/metrics` endpoint)
- ✅ OpenTelemetry distributed tracing (Jaeger)
- ✅ Comprehensive test suite (>85% coverage)
- ✅ Concurrent race-condition testing

### Roadmap

| Phase | Focus | Status |
|---|---|---|
| Phase 1 | Reliability — outbox pattern + idempotency key | ✅ Done |
| Phase 2 | Observability — OpenTelemetry (Jaeger) + Prometheus | ✅ Done |
| Phase 3 | Test coverage — comprehensive unit/integration + concurrent stock test | ✅ Done |
| Phase 4 | AI agent integration — Anthropic API, goroutine/channel, rate limiting | ⏳ In Progress |
| Phase 5 | Future — Kafka consumer (notifications), CI/CD, gRPC, Swagger auto-gen | 📅 Backlog |
