# 🚀 Catalyst

**Enterprise-Grade Event-Driven Commerce Backend**

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
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

| Problem | Solution |
|---|---|
| Two requests charge customer simultaneously → double charge | **Atomic database transactions** — all changes succeed or rollback |
| 1000 concurrent requests for "iPhone 15" → database melts | **Redis caching + singleflight** — 1 DB query, 999 get cached result |
| Email service is down → should transaction fail? | **Event-driven architecture** — transaction completes, Kafka retries later |
| Switch MySQL to PostgreSQL → rewrite everything? | **Clean architecture** — only repository changes, service/handler untouched |

---

## 🏗️ Architecture

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
 │   [ Transaction Atomicity ]   │─────┐ (Save to Outbox → Kafka)
 └───────────────┬───────────────┘     │
                 │                     ▼
                 ▼              ┌────────────────┐
 ┌───────────────────────────────┐      │ Apache Kafka   │
 │   Repository (Data Access)    │      │ (Message Bus)  │
 │   [ Cache Stampede Protect ]  │      └────────────────┘
 └───────────────┬───────────────┘             │
                 │                             ▼
                 │                      ┌────────────────┐
         ┌───────┴───────┐              │  Async Workers │
         ▼               ▼              └────────────────┘
 ┌──────────────┐ ┌──────────────┐
 │    Redis     │ │    MySQL     │
 │ (Multi-tier  │ │ (Source of   │
 │   Caching)   │ │    Truth)    │
 └──────────────┘ └──────────────┘
```

</div>

> 📖 **Deep dive**: See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) for complete architecture,
> tech stack, dependencies, folder structure, and project roadmap.

---

## 🚀 Quick Start

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) v20.10+
- [Docker Compose](https://docs.docker.com/compose/install/) v2.0+
- [Git](https://git-scm.com/) v2.0+

### 1. Clone & Configure

```bash
git clone https://github.com/assidik12/catalyst.git
cd catalyst
cp .env.example .env
```

### 2. Run

```bash
docker-compose up --build
```

### 3. Verify

| Service | URL |
|---|---|
| 🌐 API | http://localhost:3001 |
| 📚 Swagger UI | http://localhost:3001/api/v1/docs |
| 📈 Grafana | http://localhost:3002 (`admin`/`admin`) |
| 🕵️ Jaeger | http://localhost:16686 |
| 📊 Prometheus | http://localhost:9090 |

> 📖 **Full setup guide**: See [`docs/DEPLOYMENT.md`](docs/DEPLOYMENT.md) for Docker services,
> port mapping, environment variables, local development setup, and troubleshooting.

---

## 📚 API Endpoints

### 🔑 Authentication

The API uses **JWT** for authentication:
1. Register via `POST /api/v1/users/register`
2. Login via `POST /api/v1/users/login` → get JWT token
3. Include header: `Authorization: Bearer <token>`

### Endpoints

<details>
<summary><b>Click to expand</b></summary>

#### User
| Method | Endpoint | Auth | Description |
|---|---|---|---|
| POST | `/api/v1/users/register` | — | Register new user |
| POST | `/api/v1/users/login` | — | Login (returns JWT) |
| GET | `/api/v1/users/profile` | 🔒 | Get profile |
| PUT | `/api/v1/users/profile` | 🔒 | Update profile |

#### Product
| Method | Endpoint | Auth | Description |
|---|---|---|---|
| GET | `/api/v1/products` | — | List products (paginated, cached ⚡) |
| GET | `/api/v1/products/:id` | — | Get product by ID (cached ⚡) |
| POST | `/api/v1/products` | 🔒 | Create product |
| PUT | `/api/v1/products/:id` | 🔒 | Update product (invalidates cache) |
| DELETE | `/api/v1/products/:id` | 🔒 | Delete product (invalidates cache) |

#### Transaction
| Method | Endpoint | Auth | Description |
|---|---|---|---|
| GET | `/api/v1/transactions` | 🔒 | List user transactions |
| GET | `/api/v1/transactions/:id` | 🔒 | Get transaction by ID |
| POST | `/api/v1/transactions` | 🔒 | Create transaction (publishes event 📨) |

</details>

> 📖 **Interactive docs**: Swagger UI available at http://localhost:3001/api/v1/docs

---

## 🧪 Testing

```bash
# All tests
go test -v ./...

# With race detector
go test -race ./...

# Coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

> 📖 **Testing strategy**: See [`docs/STANDARDS.md`](docs/STANDARDS.md#12-testing)
> for test patterns, naming conventions, and coverage checklist.

---

## 🗺️ Roadmap

| Phase | Focus | Status |
|---|---|---|
| Phase 1 | Reliability — outbox pattern + idempotency key | ✅ Done |
| Phase 2 | Observability — OpenTelemetry + Prometheus | ✅ Done |
| Phase 3 | Test coverage — comprehensive suite + concurrent testing | ✅ Done |
| Phase 4 | AI agent integration — Anthropic API, goroutine/channel | ⏳ In Progress |
| Phase 5 | Future — Kafka consumers, CI/CD, gRPC, Swagger auto-gen | 📅 Backlog |

> 📖 **Detailed roadmap**: See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md#project-status--roadmap)

---

## 📖 Documentation

| Document | Description |
|---|---|
| [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) | Technical architecture, tech stack, dependencies, folder structure, roadmap |
| [`docs/CONCEPTS.md`](docs/CONCEPTS.md) | Design decisions — *why* we chose Clean Architecture, event-driven, etc. |
| [`docs/STANDARDS.md`](docs/STANDARDS.md) | Coding standards — naming, error handling, layer rules, testing, commits |
| [`docs/DEPLOYMENT.md`](docs/DEPLOYMENT.md) | Deployment guide, Docker services, env vars, observability, troubleshooting |
| [`docs/CONTRIBUTING.md`](docs/CONTRIBUTING.md) | How to contribute — PR workflow, branch naming, review process |

---

## 🤝 Contributing

Contributions are highly appreciated! See [`docs/CONTRIBUTING.md`](docs/CONTRIBUTING.md) 
for the full guide.

Quick start:
1. Fork this repository
2. Create feature branch (`git checkout -b feat/amazing-feature`)
3. Follow [coding standards](docs/STANDARDS.md)
4. Commit with [conventional commits](docs/STANDARDS.md#13-commit-message)
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

If this project helped you, please give it a ⭐️ on [GitHub](https://github.com/assidik12/catalyst)!

---

<div align="center">

**[Back to Top ⬆️](#-catalyst)**

Made with ❤️ using Go • Powered by Clean Architecture

</div>
