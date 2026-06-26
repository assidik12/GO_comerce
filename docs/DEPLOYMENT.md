# Deployment & Operations Guide — Catalyst

> Panduan ini mencakup **setup**, **deployment**, **monitoring**, dan **troubleshooting**
> untuk proyek Catalyst. Ditujukan untuk DevOps engineer dan developer yang mengelola
> environment production/staging.
>
> Untuk arsitektur dan tech stack, lihat [`ARCHITECTURE.md`](./ARCHITECTURE.md).

---

## Daftar Isi

1. [Prerequisites](#prerequisites)
2. [Docker Deployment](#docker-deployment)
3. [Local Development (tanpa Docker)](#local-development)
4. [Services Overview](#services-overview)
5. [Environment Variables](#environment-variables)
6. [Observability](#observability)
7. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Docker Deployment
- [Docker](https://docs.docker.com/get-docker/) v20.10+
- [Docker Compose](https://docs.docker.com/compose/install/) v2.0+
- [Git](https://git-scm.com/) v2.0+

### Local Development
- Go 1.22+
- MySQL 8.0
- Redis 7.0
- Apache Kafka 3.0+
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI
- [Google Wire](https://github.com/google/wire) CLI (optional, untuk regenerate DI)

---

## Docker Deployment

### 1. Clone & Configure

```bash
git clone https://github.com/assidik12/catalyst.git
cd catalyst
cp .env.example .env
```

Edit `.env` sesuai kebutuhan. Lihat [Environment Variables](#environment-variables) untuk referensi lengkap.

> ⚠️ **Security**: Ganti semua password default sebelum deploy ke production.
> Pastikan `.env` ada di `.gitignore` dan tidak di-commit.

### 2. Start All Services

```bash
docker-compose up --build
```

Proses ini akan:
- 📦 Build Docker image untuk Go application
- 🗄️ Setup MySQL dengan healthcheck
- 🚀 Setup Redis dengan healthcheck
- 📨 Setup Apache Kafka & Zookeeper
- 📈 Setup Prometheus, Jaeger, dan Grafana
- 🔄 Jalankan database migration secara otomatis
- ▶️ Start application di port 3001

### 3. Verify

```bash
# Cek semua container running
docker-compose ps

# Cek logs aplikasi
docker-compose logs -f go-app-service

# Health check manual
curl http://localhost:3001/api/v1/products?page=1
```

### 4. Stop

```bash
# Stop semua service (data preserved)
docker-compose down

# Stop dan hapus semua data (⚠️ destructive)
docker-compose down -v
```

---

## Local Development

### 1. Install Dependencies

```bash
go mod download
```

### 2. Install CLI Tools

```bash
# Wire (dependency injection code generator)
go install github.com/google/wire/cmd/wire@latest

# golang-migrate (database migration)
# Lihat: https://github.com/golang-migrate/migrate/tree/master/cmd/migrate
```

### 3. Setup External Services

Pastikan MySQL, Redis, dan Kafka sudah berjalan secara lokal, atau jalankan hanya infrastructure via Docker:

```bash
# Jalankan hanya DB, Redis, Kafka (tanpa Go app)
docker-compose up db cache zookeeper message-broker -d
```

### 4. Configure Environment

```bash
cp .env.example .env
```

Update `.env` untuk local:
```env
MYSQL_HOST=localhost
MYSQL_PORT=3307
REDIS_HOST=localhost
KAFKA_BROKER=localhost:9093
JAEGER_ENDPOINT=http://localhost:14268/api/traces
```

### 5. Run Migration

```bash
migrate -database "mysql://gouser:gosecret123@tcp(localhost:3307)/go_ecommerce_db?multiStatements=true" -path db/migrations up
```

### 6. (Optional) Regenerate Wire

Jika dependency berubah:

```bash
cd cmd/injector
wire
cd ../..
```

### 7. Run Application

```bash
go run cmd/api/main.go
```

---

## Services Overview

### Docker Services

| Service | Container | Image | Port(s) | Volume | Description |
|---|---|---|---|---|---|
| **Go App** | `go-app-service` | Custom (built) | `3001:3000` | — | Main application |
| **MySQL** | `db-mysql-service` | `mysql:8.0` | `3307:3306` | `db-data` | Primary database |
| **Redis** | `redis-cache-service` | `redis:7.0-alpine` | `6379:6379` | `redis-data` | Cache layer |
| **Zookeeper** | `zookeeper-service` | `wurstmeister/zookeeper` | `2181:2181` | — | Kafka coordination |
| **Kafka** | `kafka-broker-service` | `wurstmeister/kafka` | `9092:9092`, `9093:9093` | `kafka-data` | Message broker |
| **Prometheus** | `prometheus-service` | `prom/prometheus:latest` | `9090:9090` | — | Metrics collector |
| **Jaeger** | `jaeger-service` | `jaegertracing/all-in-one` | `16686`, `14268`, `4317`, `4318` | — | Distributed tracing |
| **Grafana** | `grafana-service` | `grafana/grafana:latest` | `3002:3000` | — | Monitoring dashboards |

### Access URLs

| Service | URL | Credential |
|---|---|---|
| API | http://localhost:3001 | — |
| Swagger UI | http://localhost:3001/api/v1/docs | — |
| Grafana | http://localhost:3002 | `admin` / `admin` |
| Jaeger UI | http://localhost:16686 | — |
| Prometheus | http://localhost:9090 | — |

### Data Persistence

| Volume | Service | Description |
|---|---|---|
| `db-data` | MySQL | Database files |
| `redis-data` | Redis | Cache data |
| `kafka-data` | Kafka | Message logs |

Migrations dijalankan otomatis saat container startup via `entrypoint.sh`.

---

## Environment Variables

### Application

| Variable | Default | Description |
|---|---|---|
| `APP_PORT` | `3000` | Internal port (exposed via Docker as 3001) |

### MySQL

| Variable | Required | Description |
|---|---|---|
| `MYSQL_HOST` | ✅ | `db` (Docker) / `localhost` (local) |
| `MYSQL_PORT` | ✅ | `3306` (Docker) / `3307` (local via Docker mapping) |
| `MYSQL_USER` | ✅ | Database username |
| `MYSQL_PASSWORD` | ✅ | Database password |
| `MYSQL_DATABASE` | ✅ | Database name |
| `MYSQL_ROOT_PASSWORD` | ✅ | MySQL root password |
| `DB_URL` | ✅ | Full connection string for migrations |

### Redis

| Variable | Required | Description |
|---|---|---|
| `REDIS_HOST` | ✅ | `cache` (Docker) / `localhost` (local) |
| `REDIS_PORT` | ✅ | `6379` |
| `REDIS_PASSWORD` | ✅ | Redis authentication password |

### Kafka

| Variable | Required | Description |
|---|---|---|
| `KAFKA_BROKER` | ✅ | `message-broker:9092` (Docker) / `localhost:9093` (local) |
| `KAFKA_HOST` | ✅ | Kafka hostname |
| `KAFKA_PORT` | ✅ | Kafka port |

### Security

| Variable | Required | Description |
|---|---|---|
| `JWT_SECRET` | ✅ | Secret key for JWT. Generate with: `openssl rand -base64 32` |

### Observability

| Variable | Required | Description |
|---|---|---|
| `JAEGER_ENDPOINT` | ✅ | `http://jaeger:14268/api/traces` (Docker) / `http://localhost:14268/api/traces` (local) |

---

## Observability

### Prometheus Metrics

Prometheus scrapes metrics dari endpoint `/metrics` pada Go application.

**Konfigurasi** ada di `config/prometheus.yml`:
```yaml
global:
  scrape_interval: 5s

scrape_configs:
  - job_name: "catalyst_backend"
    static_configs:
      - targets: ["go-app-service:3000"]
```

**Key Metrics:**
- `http_requests_total` — traffic rate (TPS)
- `http_request_duration_seconds` — latency (P50, P90, P99)
- Go runtime metrics (memory, goroutines) via `promauto`

### Grafana Dashboards

- **URL**: http://localhost:3002
- **Login**: `admin` / `admin`
- Prometheus sudah di-provision otomatis sebagai data source via `config/grafana/provisioning/`

### Distributed Tracing (Jaeger)

- **URL**: http://localhost:16686
- Setiap HTTP request membuat **Root Span**
- Business logic berat (misal `TransactionService.Save()`) membuat **Child Span**
- Waterfall view menunjukkan durasi tiap operasi (DB query, Kafka publish, cache lookup)

Tracing di-inject via `TracingMiddleware` dan menggunakan OpenTelemetry SDK.

---

## Troubleshooting

### Container fails to start

```bash
docker-compose down
docker-compose down -v        # ⚠️ hapus semua data
docker-compose up --build
```

### Port already in use

```bash
# Windows
netstat -ano | findstr :3001
netstat -ano | findstr :9092

# Linux/Mac
lsof -i :3001

# Solution: kill process atau ubah port di .env dan docker-compose.yml
```

### Kafka broker not reachable

```bash
docker logs kafka-broker-service
docker exec -it kafka-broker-service kafka-topics.sh --bootstrap-server localhost:9092 --list
docker exec -it zookeeper-service zkServer.sh status
```

### Redis connection refused

```bash
docker logs redis-cache-service
docker exec -it redis-cache-service redis-cli
> AUTH redissecret123
> PING
```

### Database migration failed

```bash
docker exec -it go-app-service /bin/sh
migrate -database "$DB_URL" -path db/migrations version

# Force specific version (⚠️ use with caution)
migrate -database "$DB_URL" -path db/migrations force <version>
```

### Application crashes on startup

```bash
# Cek logs
docker-compose logs go-app-service

# Common causes:
# 1. MySQL belum ready — cek healthcheck
# 2. Kafka belum ready — cek depends_on dan healthcheck
# 3. .env tidak lengkap — bandingkan dengan .env.example
# 4. Migration belum jalan — cek entrypoint.sh
```

### High latency / slow responses

1. **Cek Jaeger** — apakah ada span yang lambat?
2. **Cek Prometheus** — apakah ada lonjakan `http_request_duration_seconds`?
3. **Cek Redis** — apakah cache miss rate tinggi?
4. **Cek MySQL** — apakah ada slow query? `SHOW PROCESSLIST;`
