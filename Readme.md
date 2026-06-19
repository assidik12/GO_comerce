# 🚀 Catalyst

**Enterprise-Grade Event-Driven Commerce Backend**

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![Architecture](https://img.shields.io/badge/Architecture-Clean%20Architecture-1e90ff?style=for-the-badge)](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
[![Event-Driven](https://img.shields.io/badge/Pattern-Event%20Driven-ff6b6b?style=for-the-badge)](https://en.wikipedia.org/wiki/Event-driven_architecture)

### What is Catalyst?

Catalyst demonstrates how to build a **production-grade e-commerce platform** 
using Go and microservices patterns. The project focuses on:

- 🏗️ **Clean Architecture** — Clear separation of concerns for maintainability
- ⚡ **Event-Driven Design** — Kafka-based async processing without data loss
- 🔐 **Transaction Safety** — Atomic operations, server-side validation
- 📊 **Real-time Performance** — Redis caching, singleflight cache stampede prevention
- 🎯 **Production Ready** — Graceful shutdown, structured logging, error handling

### Why "Catalyst"?

In chemistry, a catalyst **accelerates a reaction without being consumed**. 
Similarly, Catalyst is the infrastructure layer that makes your commerce 
system work reliably — transparent, fast, and essential. It's not the 
business logic; it's what makes good business logic possible.

---

## ✨ Key Features

<table>
<tr>
<td width="50%">

### 👤 User Management

- ✅ User registration & authentication
- ✅ JWT-based authorization
- ✅ Password hashing with bcrypt
- ✅ Profile management (CRUD)

</td>
<td width="50%">

### 📦 Product Management

- ✅ CRUD operations for products
- ✅ Category management
- ✅ Redis caching with auto-invalidation
- ✅ Cache stampede protection (singleflight)
- ✅ Pagination support

</td>
</tr>
<tr>
<td width="50%">

### 💳 Transaction Management

- ✅ Order creation & tracking
- ✅ Transaction history
- ✅ Idempotency Key support (Duplicate prevention)
- ✅ Transactional Outbox Pattern (Guaranteed Event Delivery)
- ✅ Business logic validation
- ✅ Event publishing to Kafka

</td>
<td width="50%">

### 🛠️ Technical Features

### 🔭 Observability & Testing

- ✅ Prometheus Metrics (`/metrics`)
- ✅ OpenTelemetry Tracing (Jaeger)
- ✅ Comprehensive Test Suite (>85% coverage)
- ✅ Concurrent Race-Condition Testing
- ✅ Error handling middleware

</td>
</tr>
</table>

---

## 🎯 Why This Project Matters

This project is not just another e-commerce API. It's a **reference implementation** 
demonstrating how to build reliable, scalable systems with Go. Each design decision 
solves real production problems:

### Problem: Data Integrity
**Scenario:** Two requests charge customer simultaneously → double charge disaster

**Solution:** Atomic database transactions. All changes succeed together or 
all rollback. No partial states.

### Problem: Performance Under Load
**Scenario:** 1000 concurrent requests for "iPhone 15" → database melts

**Solution:** Redis caching + singleflight. Only 1 database query, 999 requests 
get cached result instantly.

### Problem: System Reliability
**Scenario:** Email service is down → should transaction fail?

**Solution:** Event-driven architecture. Transaction completes, then async 
Kafka consumer handles email. If email fails, retry later. Transaction safe either way.

### Problem: Maintainability at Scale
**Scenario:** Switch database from MySQL to PostgreSQL → rewrite everything?

**Solution:** Clean architecture. Repository layer abstracts database. 
Only repository changes, service/handler untouched.

---

## 🚀 Use Cases

Catalyst is designed for:

- **E-commerce platforms** needing reliable order processing
- **Fintech applications** requiring atomic transactions
- **Marketplace systems** with inventory management
- **Subscription services** with event-based workflows

Learn how Catalyst handles these with pattern that scale to millions of users.

---

## 🏗️ Architecture

This application uses **Clean Architecture** integrated with **Event-Driven Architecture** to support a highly *scalable* and *resilient* commerce platform.

### Layered Architecture Diagram

<div align="center">

```text
       [Client Request / HTTP]
                 │
                 ▼
 ┌───────────────────────────────┐
 │       HTTP Handler (Delivery) │  ← Receives input, parses DTOs, returns HTTP Response
 └───────────────┬───────────────┘
                 │
                 ▼
 ┌───────────────────────────────┐  ← Business validation, price calculation, orchestration
 │      Service (Use Case)       │  
 │   [ Transaction Atomicity ]   │─────┐ (Publish Async Event)
 └───────────────┬───────────────┘     │
                 │                     ▼
                 ▼              ┌────────────────┐
 ┌───────────────────────────────┐      │ Apache Kafka   │  ← Event Broker for Asynchronous Tasks
 │   Repository (Data Access)    │      │ (Message Bus)  │  
 │   [ Cache Stampede Protect ]  │      └────────────────┘
 └───────────────┬───────────────┘             │
                 │                             ▼
                 │                      ┌────────────────┐
         ┌───────┴───────┐              │  Async Workers │  ← Notifications (Email), Third-party integrations
         ▼               ▼              └────────────────┘
 ┌──────────────┐ ┌──────────────┐
 │    Redis     │ │    MySQL     │
 │ (Multi-tier  │ │ (Source of   │
 │   Caching)   │ │    Truth)    │
 └──────────────┘ └──────────────┘
```

</div>

</div>

**Each layer has strict boundaries:**
- **Handler (Presentation)**: Focuses only on the HTTP layer, JSON to DTO deserialization, and mapping *Error Sentinels* to proper HTTP Status Codes.
- **Service (Business Logic)**: Unaware of specific databases or HTTP. Executes core *Use Cases* (e.g., deducting stock, verifying prices, publishing Events).
- **Repository (Data Access)**: Focuses on executing database queries and caching (Redis). This is where `Singleflight` is used to prevent *Cache Stampedes*.
- **Infrastructure**: Initialization of external dependencies (Database connections, Redis client, Kafka writer).

### 📂 Folder Structure

```
go-restfull-api/
├── cmd/
│   ├── api/                 # Main application entrypoint
│   └── injector/            # Dependency injection (Wire)
├── config/                  # Configuration management (Viper)
├── db/migrations/           # Database migrations (golang-migrate)
├── docs/                    # Documentation & Swagger specs
├── internal/
│   ├── delivery/
│   │   └── http/
│   │       ├── handler/     # HTTP handlers (Presentation layer)
│   │       ├── dto/         # Data Transfer Objects
│   │       ├── middleware/  # JWT Auth, Error handling
│   │       └── route/       # Route definitions
│   ├── domain/
│   │   ├── *.go            # Business entities (User, Product, Transaction)
│   │   └── event/          # Event payloads (OrderCreatedEvent)
│   ├── infrastructure/      # External service clients
│   │   ├── database.go     # MySQL connection
│   │   ├── redis.go        # Redis connection
│   │   └── kafka.go        # Kafka writer setup
│   ├── pkg/
│   │   ├── cache/          # Cache wrapper (abstraction)
│   │   └── response/       # Standardized HTTP responses
│   ├── producer/           # Kafka producers (OrderProducer)
│   ├── repository/
│   │   └── mysql/          # Data access layer (MySQL queries)
│   └── service/            # Business logic layer
└── test/                    # Integration & unit tests
```

---

## 🛠️ Tech Stack

<div align="center">
<table>
<tr>
<td align="center" width="20%">
<img src="https://go.dev/blog/go-brand/Go-Logo/PNG/Go-Logo_Blue.png" width="80" height="80" alt="Go"/>
<br><b>Go 1.22+</b>
<br>Core Language
</td>
<td align="center" width="20%">
<img src="https://www.mysql.com/common/logos/logo-mysql-170x115.png" width="80" height="80" alt="MySQL"/>
<br><b>MySQL 8.0</b>
<br>Primary Database
</td>
<td align="center" width="20%">
<img src="https://redis.io/wp-content/uploads/2024/04/Logotype.svg?auto=webp&quality=85,75&width=120" width="80" alt="Redis"/>
<br><b>Redis 7.0</b>
<br>Caching Layer
</td>
<td align="center" width="20%">
<img src="https://img.icons8.com/?size=100&id=k4fZIepXxmAZ&format=png&color=ffffff" width="80" alt="Kafka"/>
<br><b>Apache Kafka</b>
<br>Message Broker
</td>
<td align="center" width="20%">
<img src="https://upload.wikimedia.org/wikipedia/commons/3/38/Prometheus_software_logo.svg" width="80" height="80" alt="Prometheus"/>
<br><b>Prometheus & Jaeger</b>
<br>Observability
</td>
</tr>
</table>
</div>

### 📚 Dependencies & Libraries

| Category           | Library                                                                   | Purpose                           |
| ------------------ | ------------------------------------------------------------------------- | --------------------------------- |
| **Router**         | [`julienschmidt/httprouter`](https://github.com/julienschmidt/httprouter) | High-performance HTTP router      |
| **Database**       | [`go-sql-driver/mysql`](https://github.com/go-sql-driver/mysql)           | MySQL driver for Go               |
| **Cache**          | [`redis/go-redis`](https://github.com/redis/go-redis)                     | Redis client for Go               |
| **Message Broker** | [`segmentio/kafka-go`](https://github.com/segmentio/kafka-go)             | Pure Go Kafka client              |
| **Concurrency**    | [`golang.org/x/sync`](https://pkg.go.dev/golang.org/x/sync)               | Singleflight (cache stampede)     |
| **Validation**     | [`go-playground/validator`](https://github.com/go-playground/validator)   | Struct validation                 |
| **JWT**            | [`golang-jwt/jwt`](https://github.com/golang-jwt/jwt)                     | JSON Web Token implementation     |
| **Config**         | [`spf13/viper`](https://github.com/spf13/viper)                           | Configuration management          |
| **DI**             | [`google/wire`](https://github.com/google/wire)                           | Compile-time dependency injection |
| **Migration**      | [`golang-migrate`](https://github.com/golang-migrate/migrate)             | Database migrations               |
| **Password**       | [`golang.org/x/crypto`](https://pkg.go.dev/golang.org/x/crypto)           | Bcrypt hashing                    |
| **UUID**           | [`google/uuid`](https://github.com/google/uuid)                           | UUID generation                   |
| **Testing**        | [`stretchr/testify`](https://github.com/stretchr/testify)                 | Assertions & Mocking framework    |
| **Testing Mocks**  | [`DATA-DOG/go-sqlmock`](https://github.com/DATA-DOG/go-sqlmock)           | Mocking SQL driver behaviour      |

---

## 🚀 Quick Start

### 📋 Prerequisites

Ensure your system has the following installed:

- [Git](https://git-scm.com/) (v2.0+)
- [Docker](https://docs.docker.com/get-docker/) (v20.10+)
- [Docker Compose](https://docs.docker.com/compose/install/) (v2.0+)

### ⚙️ Installation

#### 1️⃣ Clone Repository

```bash
git clone https://github.com/assidik12/go-restfull-api.git
cd go-restfull-api
```

#### 2️⃣ Setup Environment Variables

Create a `.env` file in the root directory:

```bash
# Windows (CMD)
type nul > .env

# Windows (PowerShell)
New-Item .env -ItemType File

# Linux/Mac
touch .env
```

Copy and adjust the following configuration into your `.env` file:

```env
# ================================
# Application Configuration
# ================================
APP_PORT=3000

# ================================
# MySQL Database Configuration
# ================================
MYSQL_HOST=db
MYSQL_PORT=3306
MYSQL_USER=gouser
MYSQL_PASSWORD=gosecret123
MYSQL_DATABASE=go_ecommerce_db
MYSQL_ROOT_PASSWORD=rootsecret123

# Database URL for migrations
DB_URL=mysql://gouser:gosecret123@tcp(db:3306)/go_ecommerce_db?multiStatements=true

# ================================
# Redis Cache Configuration
# ================================
REDIS_HOST=cache
REDIS_PORT=6379
REDIS_PASSWORD=redissecret123

# ================================
# Kafka Configuration
# ================================
KAFKA_BROKER=message-broker:9092
KAFKA_HOST=message-broker
KAFKA_PORT=9092

# ================================
# JWT Configuration
# ================================
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
```

> ⚠️ **Security Warning**:
>
> - Change all passwords to strong values for production
> - Ensure `.env` is included in `.gitignore`
> - Do not commit `.env` to the repository

#### 3️⃣ Run Application

```bash
docker-compose up --build
```

This process will:

- 📦 Build Docker images for the Go application
- 🗄️ Setup MySQL database with healthcheck
- 🚀 Setup Redis cache with healthcheck
- 📨 Setup Apache Kafka & Zookeeper
- 🔄 Run database migrations automatically
- ▶️ Start the application on port 3001

### ✅ Verification

Once all containers are running, you will see output similar to:

```
✅ zookeeper              - healthy
✅ kafka                  - healthy
✅ db-mysql-service       - healthy
✅ redis-cache-service    - healthy
✅ go-app-service         - running
```

Access endpoints:

- 🌐 **API Base URL**: http://localhost:3001
- 📚 **API Documentation**: http://localhost:3001/api/v1/docs
- 📊 **Kafka Broker**: `localhost:9092`
- 🗄️ **MySQL**: `localhost:3307`
- 💾 **Redis**: `localhost:6379`

---

## 📊 Services Overview

### 🐳 Docker Services

| Service            | Container Name        | Image                    | Port(s)                  | Volume       | Description         |
| ------------------ | --------------------- | ------------------------ | ------------------------ | ------------ | ------------------- |
| **go-app-service** | `go-app-service`      | Custom (built)           | `3001:3000`              | -            | Main Go application |
| **db**             | `db-mysql-service`    | `mysql:8.0`              | `3307:3306`              | `db-data`    | MySQL database      |
| **cache**          | `redis-cache-service` | `redis:7.0-alpine`       | `6379:6379`              | `redis-data` | Redis cache         |
| **zookeeper**      | `zookeeper`           | `wurstmeister/zookeeper` | `2181:2181`              | -            | Kafka coordination  |
| **kafka**          | `kafka`               | `wurstmeister/kafka`     | `9092:9092`, `9093:9093` | `kafka-data` | Message broker      |

### 🔌 Port Mapping

| Service   | Internal Port | External Port | Access URL              | Description        |
| --------- | ------------- | ------------- | ----------------------- | ------------------ |
| Go API    | 3000          | 3001          | `http://localhost:3001` | HTTP REST API      |
| MySQL     | 3306          | 3307          | `localhost:3307`        | Database client    |
| Redis     | 6379          | 6379          | `localhost:6379`        | Cache client       |
| Kafka     | 9092          | 9092          | `localhost:9092`        | Kafka broker       |
| Zookeeper | 2181          | 2181          | `localhost:2181`        | Kafka coordination |

### 💾 Data Persistence

- **MySQL Data**: Persisted in Docker volume `db-data`
- **Redis Data**: Persisted in Docker volume `redis-data`
- **Kafka Data**: Persisted in Docker volume `kafka-data`
- **Migrations**: Auto-run on container startup via `entrypoint.sh`

---

## 🔧 Configuration

### 🌍 Environment Variables

<details>
<summary><b>Click to expand full configuration reference</b></summary>

#### Application Settings

| Variable   | Default | Description                       |
| ---------- | ------- | --------------------------------- |
| `APP_PORT` | `3000`  | Port for Go application (internal) |

#### MySQL Settings

| Variable              | Required | Description                               |
| --------------------- | -------- | ----------------------------------------- |
| `MYSQL_HOST`          | ✅       | Database host (use `db` for Docker) |
| `MYSQL_PORT`          | ✅       | Database port (default: `3306`)           |
| `MYSQL_USER`          | ✅       | Database username                         |
| `MYSQL_PASSWORD`      | ✅       | Database password                         |
| `MYSQL_DATABASE`      | ✅       | Database name                             |
| `MYSQL_ROOT_PASSWORD` | ✅       | MySQL root password                       |
| `DB_URL`              | ✅       | Full connection string for migrations   |

#### Redis Settings

| Variable         | Required | Description                               |
| ---------------- | -------- | ----------------------------------------- |
| `REDIS_HOST`     | ✅       | Redis host (use `cache` for Docker) |
| `REDIS_PORT`     | ✅       | Redis port (default: `6379`)              |
| `REDIS_PASSWORD` | ✅       | Redis authentication password             |

#### Kafka Settings

| Variable       | Required | Description                                |
| -------------- | -------- | ------------------------------------------ |
| `KAFKA_BROKER` | ✅       | Kafka broker address (format: `host:port`) |

#### Security Settings

| Variable     | Required | Description                           |
| ------------ | -------- | ------------------------------------- |
| `JWT_SECRET` | ✅       | Secret key for JWT token generation |

</details>

---

## 📚 API Documentation

### 📖 Swagger Documentation

API documentation is available via Swagger UI:

**URL**: http://localhost:3001/api/v1/docs/

### 🔑 Authentication

The API uses **JWT (JSON Web Token)** for authentication:

1. Register a user via the `/api/v1/users/register` endpoint
2. Login to get a JWT token via `/api/v1/users/login`
3. Include the token in the header: `Authorization: Bearer <your-token>`

### 📍 Endpoints Overview

<details>
<summary><b>Click to see available endpoints</b></summary>

#### User Endpoints

- `POST /api/v1/users/register` - Register new user
- `POST /api/v1/users/login` - Login user (returns JWT)
- `GET /api/v1/users/profile` - Get user profile (🔒 protected)
- `PUT /api/v1/users/profile` - Update user profile (🔒 protected)

#### Product Endpoints

- `GET /api/v1/products` - Get all products with pagination (cached ⚡)
- `GET /api/v1/products/:id` - Get product by ID (cached ⚡)
- `POST /api/v1/products` - Create new product (🔒 protected)
- `PUT /api/v1/products/:id` - Update product (🔒 protected, invalidates cache)
- `DELETE /api/v1/products/:id` - Delete product (🔒 protected, invalidates cache)

#### Transaction Endpoints

- `GET /api/v1/transactions` - Get all user transactions (🔒 protected)
- `GET /api/v1/transactions/:id` - Get transaction by ID (🔒 protected)
- `POST /api/v1/transactions` - Create transaction (🔒 protected, publishes event 📨)

</details>

---

## 🎯 Caching Strategy

### Redis Implementation

This application uses **Redis** to cache product data, reducing database load and improving response times.

#### Cache Specifications

- **Cached Endpoints**:
  - `GET /api/v1/products/:id` - Individual product details
  - `GET /api/v1/products?page=X` - Paginated product list
- **TTL (Time-To-Live)**: 10 menit
- **Cache Key Pattern**:
  - Detail: `product:detail:{id}`
  - List: `products:list:page:{page_number}`
- **Strategy**: Cache-Aside (Lazy Loading)

#### Cache Flow

```
┌─────────────────┐
│  Client Request │
└────────┬────────┘
         │
         ▼
┌─────────────────┐      ┌──────────────┐
│  Check Redis    │─────▶│  Cache HIT   │──┐
│     Cache       │      └──────────────┘  │
└────────┬────────┘                        │
         │ Cache MISS                      │
         ▼                                 │
┌─────────────────┐                        │
│  Query MySQL    │                        │
│    Database     │                        │
└────────┬────────┘                        │
         │                                 │
         ▼                                 │
┌─────────────────┐                        │
│  Store in Redis │                        │
│  (with 10m TTL) │                        │
└────────┬────────┘                        │
         │                                 │
         └─────────────────────────────────┘
                         │
                         ▼
                 ┌──────────────┐
                 │ Return Data  │
                 └──────────────┘
```

#### Cache Invalidation

Cache is automatically invalidated (deleted) on the following events:

- **Update Product**: Deletes `product:detail:{id}` and all list caches (`products:list:*`)
- **Delete Product**: Deletes `product:detail:{id}` and all list caches
- **Create Product**: Deletes all list caches to ensure new products appear

#### Performance Optimization

- **Singleflight Pattern**: Prevents **cache stampedes** by ensuring only one goroutine queries the database for the same key simultaneously.
- **Concurrent-Safe**: `CacheWrapper` is safe for concurrent use by multiple goroutines.

---

## 📨 Event-Driven Architecture

### Apache Kafka Integration

This application uses **Apache Kafka** as a message broker to handle asynchronous processes and improve system scalability.

#### Event Flow

```
┌──────────────────┐
│ Create Transaction│
│   (HTTP POST)     │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Save to MySQL DB │
│  (Transactional) │
└────────┬─────────┘
         │ Success
         ▼
┌──────────────────┐
│ Publish Event to │
│      Kafka       │──────┐
└────────┬─────────┘      │
         │                │
         ▼                │
┌──────────────────┐      │
│ Return Response  │      │
│   to Client      │      │
└──────────────────┘      │
                          │
         ┌────────────────┘
         │ Async Processing
         ▼
┌──────────────────┐
│ Kafka Consumer   │
│ (Background Job) │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Send Email /    │
│  Notification    │
└──────────────────┘
```

#### Kafka Topics & Events

| Topic           | Event Type          | Producer             | Consumer (Future)      | Description                          |
| --------------- | ------------------- | -------------------- | ---------------------- | ------------------------------------ |
| `order_created` | `OrderCreatedEvent` | `TransactionService` | `NotificationConsumer` | Published when a new transaction is created |

#### Event Payload: `OrderCreatedEvent`

```json
{
  "order_id": 123,
  "user_id": 456,
  "user_email": "user@example.com",
  "total_price": 150000.0,
  "created_at": "2024-12-06T14:30:00Z"
}
```

#### Why Kafka?

- ⚡ **Decoupling**: Services do not need to wait for slow processes (emails, notifications) to finish
- 🚀 **Scalability**: Consumers can be scaled independently
- 🔄 **Reliability**: Messages are persisted in Kafka until successfully consumed
- 📊 **Event Sourcing**: Logs all critical events for auditing and analytics

---

## 🧪 Testing (Comprehensive Suite)

This application includes enterprise-grade unit testing that simulates various conditions, such as connection interruptions, context cancellations, and cache misses, without affecting the production environment.

### 🏗️ Tools & Mocks Used
1. **[testify/mock](https://github.com/stretchr/testify):** Used in the *Service Layer* and *Handler Layer* to accurately mock repository and service interfaces without touching the actual database or Redis.
2. **[DATA-DOG/go-sqlmock](https://github.com/DATA-DOG/go-sqlmock):** Used to mock the SQL driver level and track query executions like `t.DB.BeginTx(ctx, nil)` and `Ping()` connectivity.
3. **httptest:** The standard `net/http/httptest` library is utilized in the *Delivery* layer (`middleware` & `handler`) to record HTTP scenarios (401 Unauthorized, 200 OK, Canceled Context, etc.).

### ▶️ Run the Test Suite

```bash
# Run all Unit Test scenarios
go test -v ./...

# Run Tests with real Code Coverage across all packages
go test -v -coverpkg=./... ./...

# Run specific scenarios per sub-package / domain (e.g., service validation area)
go test -v ./test/service/...
```

### ✨ Example Evaluation Scenarios:
- **Resiliency Testing**: Simulates forced Redis failure via an invalid dummy port (`localhost:9999`) to trigger immediate I/O timeouts, falling back to MySQL queries.
- **Context Cancellation**: Simulates `context.WithCancel()` specifically on HTTP request handlers.
- **Strict Data-Driven**: Transaction total price calculation is purely executed by backend queries and cannot be manipulated via JWT/Frontend parameters.
---

## 🐛 Troubleshooting

<details>
<summary><b>Common Issues & Solutions</b></summary>

### Issue: Container fails to start

**Solution**:

```bash
# Stop all containers
docker-compose down

# Remove volumes (⚠️ this will delete data!)
docker-compose down -v

# Rebuild and restart
docker-compose up --build
```

### Issue: Port already in use

**Solution**:

```bash
# Check port usage (Windows)
netstat -ano | findstr :3001
netstat -ano | findstr :9092

# Kill the process or change ports in .env and docker-compose.yml
```

### Issue: Kafka broker not reachable

**Solution**:

```bash
# Check Kafka container logs
docker logs kafka

# Verify Kafka is listening
docker exec -it kafka kafka-topics.sh --bootstrap-server localhost:9092 --list

# Check Zookeeper health
docker exec -it zookeeper zkServer.sh status
```

### Issue: Redis connection refused

**Solution**:

```bash
# Check Redis container
docker logs redis-cache-service

# Test Redis connection
docker exec -it redis-cache-service redis-cli
> AUTH redissecret123
> PING
```

### Issue: Database migration failed

**Solution**:

```bash
# Check migration status
docker exec -it go-app-service /bin/sh
migrate -database "$DB_URL" -path db/migrations version

# Force specific version (⚠️ use with caution!)
migrate -database "$DB_URL" -path db/migrations force <version>
```

</details>

---

## 🚦 Development

### Local Development (without Docker)

<details>
<summary><b>Setup for local development</b></summary>

#### Prerequisites

- Go 1.22+
- MySQL 8.0
- Redis 7.0
- Apache Kafka 3.0+

#### Steps

1. Install dependencies:

```bash
go mod download
```

2. Install Wire (to regenerate dependency injection):

```bash
go install github.com/google/wire/cmd/wire@latest
```

3. Setup local MySQL, Redis, & Kafka

4. Update `.env` with local configurations:

```env
MYSQL_HOST=localhost
REDIS_HOST=localhost
KAFKA_BROKER=localhost:9092
```

5. Run migrations:

```bash
migrate -database "mysql://user:pass@tcp(localhost:3306)/dbname" -path db/migrations up
```

6. (Optional) Regenerate Wire code if dependencies change:

```bash
cd cmd/injector
wire
```

7. Run application:

```bash
go run cmd/api/main.go
```

</details>

---

## 🗺️ Project Phases & Roadmap

This project is being built in structured phases to ensure enterprise-grade quality at each step.

### ✅ Phase 1: Reliability Hardening (Completed)
- **Outbox Pattern**: Events are saved to an `outbox_events` table within the same DB transaction as business data. A background worker (Relay) reliably publishes them to Kafka with retry mechanisms.
- **Idempotency Key**: Endpoints like Transaction Creation support `Idempotency-Key` headers to prevent double-charging or duplicate transactions during network retries.

### ✅ Phase 2: Observability (Completed)
- **Prometheus Metrics**: Added `MetricsMiddleware` to expose a `/metrics` endpoint, capturing RED metrics (Rate, Errors, Duration) for all HTTP endpoints.
- **OpenTelemetry & Jaeger**: Distributed tracing is injected via `TracingMiddleware`. Trace contexts are propagated automatically, allowing deep insight into database and cache performance.

### ✅ Phase 3: Comprehensive Test Coverage (Completed)
- **Unit & Integration Tests**: Implemented mock-based unit tests using `testify/mock` and `go-sqlmock`.
- **Concurrent Testing**: Verified race condition prevention in stock decrement logic using parallel goroutines.

### ⏳ Phase 4: AI Agent Integration (In Progress)
- [ ] Integration with Anthropic API for Smart Product Recommendations
- [ ] User mental-wellness driven personalization
- [ ] Contextual shopping assistance

### 📅 Phase 5: Future Enhancements (Backlog)
- [ ] Implement Kafka Consumer for email notifications (`NotificationConsumer`)
- [ ] Setup CI/CD pipeline (GitHub Actions)
- [ ] Add Swagger auto-generation via `swaggo`
- [ ] Implement gRPC endpoints for inter-service communication

---

## 🤝 Contributing

Contributions are highly appreciated! Feel free to open an issue or submit a pull request for improvements.

### 📝 How to Contribute

1. Fork this repository
2. Create feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to branch (`git push origin feature/AmazingFeature`)
5. Open Pull Request

---

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.

---

## 👨‍💻 Author

**Ahmad Sofi Sidik**

[![LinkedIn](https://img.shields.io/badge/LinkedIn-Connect-0077B5?style=for-the-badge&logo=linkedin)](https://www.linkedin.com/in/ahmad-sofi-sidik/)
[![GitHub](https://img.shields.io/badge/GitHub-Follow-181717?style=for-the-badge&logo=github)](https://github.com/assidik12)

---

## 🌟 Show Your Support

If this project helped you, please give it a ⭐️ on [GitHub](https://github.com/assidik12/go-restfull-api)!

---

<div align="center">

**[Back to Top ⬆️](#-go-e-commerce-rest-api)**

Made with ❤️ using Go • Powered by Clean Architecture

</div>
