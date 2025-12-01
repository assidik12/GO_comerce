<div align="center">

# 🛒 Go E-Commerce REST API

### _Enterprise-grade RESTful API built with Clean Architecture_

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker)](https://www.docker.com/)
[![MySQL](https://img.shields.io/badge/MySQL-8.0-4479A1?style=for-the-badge&logo=mysql&logoColor=white)](https://www.mysql.com/)
[![Redis](https://img.shields.io/badge/Redis-7.0-DC382D?style=for-the-badge&logo=redis&logoColor=white)](https://redis.io/)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)

**[Features](#-fitur-utama)** • **[Architecture](#️-arsitektur)** • **[Quick Start](#-quick-start)** • **[Documentation](#-dokumentasi-api)** • **[Contributing](#-kontribusi)**

---

</div>

## 📖 Tentang Proyek

Sebuah RESTful API modern untuk aplikasi E-Commerce yang dibangun dengan **Go (Golang)** mengikuti prinsip **Clean Architecture**. Proyek ini menyediakan backend yang robust, scalable, dan mudah di-maintain untuk mengelola user, produk, dan transaksi.

### 🎯 Kenapa Proyek Ini?

- 🏗️ **Clean Architecture** - Pemisahan concern yang jelas untuk maintainability
- 🚀 **Production Ready** - Fully containerized dengan Docker & Docker Compose
- ⚡ **High Performance** - Redis caching untuk optimasi response time
- 🔐 **Secure** - JWT authentication & middleware protection
- 📦 **Easy Deployment** - One-command setup untuk development & production

---

## ✨ Fitur Utama

<table>
<tr>
<td width="50%">

### 👤 User Management

- ✅ User registration & authentication
- ✅ JWT-based authorization
- ✅ Password hashing with bcrypt
- ✅ Profile management

</td>
<td width="50%">

### 📦 Product Management

- ✅ CRUD operations for products
- ✅ Category management
- ✅ Redis caching (10 min TTL)
- ✅ Soft delete support

</td>
</tr>
<tr>
<td width="50%">

### 💳 Transaction Management

- ✅ Order creation & tracking
- ✅ Transaction history
- ✅ Status management
- ✅ Business logic validation

</td>
<td width="50%">

### 🛠️ Technical Features

- ✅ Auto database migration
- ✅ Input validation
- ✅ Error handling middleware
- ✅ API documentation (Swagger)

</td>
</tr>
</table>

---

## 🏗️ Arsitektur

Aplikasi ini menggunakan **Clean Architecture** dengan pemisahan layer yang jelas:

<div align="center">

![Clean Architecture Diagram](./docs/architecture.png)

</div>

### 📂 Struktur Folder

```
go-restfull-api/
├── cmd/                    # Application entrypoints
│   ├── api/               # Main application
│   └── injector/          # Dependency injection (Wire)
├── config/                # Configuration management
├── db/migrations/         # Database migrations
├── docs/                  # Documentation & Swagger
├── internal/
│   ├── delivery/         # Presentation layer (HTTP handlers, DTOs)
│   ├── domain/           # Business entities
│   ├── infrastructure/   # External services (MySQL, Redis)
│   ├── pkg/              # Shared utilities
│   ├── repository/       # Data access layer
│   └── service/          # Business logic layer
└── test/                  # Test files
```

---

## 🛠️ Tech Stack

<div align="center" width="80%" height="80%">
<table>
<tr>
<td align="center" width="25%">
<img src="https://go.dev/blog/go-brand/Go-Logo/PNG/Go-Logo_Blue.png" width="80" height="80" alt="Go"/>
<br><b>Go 1.22+</b>
<br>Core Language
</td>
<td align="center" width="25%">
<img src="https://www.mysql.com/common/logos/logo-mysql-170x115.png" width="80" height="80" alt="MySQL"/>
<br><b>MySQL 8.0</b>
<br>Primary Database
</td>
<td align="center" width="25%">
<img src="https://redis.io/wp-content/uploads/2024/04/Logotype.svg?auto=webp&quality=85,75&width=120" width="80" alt="Redis"/>
<br><b>Redis 7.0</b>
<br>Caching Layer
</td>
<td align="center" width="25%">
<img src="https://www.docker.com/wp-content/uploads/2022/03/vertical-logo-monochromatic.png" width="80" height="80" alt="Docker"/>
<br><b>Docker</b>
<br>Containerization
</td>
</tr>
</table>
</div>

### 📚 Dependencies & Libraries

| Category       | Library                                                                   | Purpose                       |
| -------------- | ------------------------------------------------------------------------- | ----------------------------- |
| **Router**     | [`julienschmidt/httprouter`](https://github.com/julienschmidt/httprouter) | High-performance HTTP router  |
| **Database**   | [`go-sql-driver/mysql`](https://github.com/go-sql-driver/mysql)           | MySQL driver for Go           |
| **Cache**      | [`redis/go-redis`](https://github.com/redis/go-redis)                     | Redis client for Go           |
| **Validation** | [`go-playground/validator`](https://github.com/go-playground/validator)   | Struct validation             |
| **JWT**        | [`golang-jwt/jwt`](https://github.com/golang-jwt/jwt)                     | JSON Web Token implementation |
| **Config**     | [`spf13/viper`](https://github.com/spf13/viper)                           | Configuration management      |
| **DI**         | [`google/wire`](https://github.com/google/wire)                           | Dependency injection          |
| **Migration**  | [`golang-migrate`](https://github.com/golang-migrate/migrate)             | Database migrations           |
| **Password**   | [`golang.org/x/crypto`](https://pkg.go.dev/golang.org/x/crypto)           | Bcrypt hashing                |

---

## 🚀 Quick Start

### 📋 Prerequisites

Pastikan sistem Anda telah menginstall:

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

Buat file `.env` di root directory:

```bash
# Windows (CMD)
type nul > .env

# Windows (PowerShell)
New-Item .env -ItemType File

# Linux/Mac
touch .env
```

Copy dan sesuaikan konfigurasi berikut ke file `.env`:

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
# JWT Configuration
# ================================
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
```

> ⚠️ **Security Warning**:
>
> - Ganti semua password dengan nilai yang strong untuk production
> - Pastikan `.env` sudah ada di `.gitignore`
> - Jangan commit `.env` ke repository

#### 3️⃣ Run Application

```bash
docker-compose up --build
```

Proses ini akan:

- 📦 Build Docker images untuk Go application
- 🗄️ Setup MySQL database dengan healthcheck
- 🚀 Setup Redis cache dengan healthcheck
- 🔄 Menjalankan database migrations secara otomatis
- ▶️ Start aplikasi pada port 3001

### ✅ Verifikasi

Setelah semua container berjalan, Anda akan melihat output:

```
✅ db-mysql-service      - healthy
✅ redis-cache-service   - healthy
✅ go-app-service        - running
```

Akses endpoints:

- 🌐 **API Base URL**: http://localhost:3001
- 📚 **API Documentation**: http://localhost:3001/api/v1/docs

---

## 📊 Services Overview

### 🐳 Docker Services

| Service            | Container Name        | Image              | Port        | Volume       | Description         |
| ------------------ | --------------------- | ------------------ | ----------- | ------------ | ------------------- |
| **go-app-service** | `go-app-service`      | Custom (built)     | `3001:3000` | -            | Main Go application |
| **db**             | `db-mysql-service`    | `mysql:8.0`        | `3307:3306` | `db-data`    | MySQL database      |
| **cache**          | `redis-cache-service` | `redis:7.0-alpine` | `6379:6379` | `redis-data` | Redis cache         |

### 🔌 Port Mapping

| Service | Internal Port | External Port | Access URL              |
| ------- | ------------- | ------------- | ----------------------- |
| Go API  | 3000          | 3001          | `http://localhost:3001` |
| MySQL   | 3306          | 3307          | `localhost:3307`        |
| Redis   | 6379          | 6379          | `localhost:6379`        |

### 💾 Data Persistence

- **MySQL Data**: Persisted in Docker volume `db-data`
- **Redis Data**: Persisted in Docker volume `redis-data`
- **Migrations**: Auto-run on container startup via `entrypoint.sh`

---

## 🔧 Configuration

### 🌍 Environment Variables

<details>
<summary><b>Click to expand full configuration reference</b></summary>

#### Application Settings

| Variable   | Default | Description                       |
| ---------- | ------- | --------------------------------- |
| `APP_PORT` | `3000`  | Port untuk aplikasi Go (internal) |

#### MySQL Settings

| Variable              | Required | Description                               |
| --------------------- | -------- | ----------------------------------------- |
| `MYSQL_HOST`          | ✅       | Database host (gunakan `db` untuk Docker) |
| `MYSQL_PORT`          | ✅       | Database port (default: `3306`)           |
| `MYSQL_USER`          | ✅       | Database username                         |
| `MYSQL_PASSWORD`      | ✅       | Database password                         |
| `MYSQL_DATABASE`      | ✅       | Database name                             |
| `MYSQL_ROOT_PASSWORD` | ✅       | MySQL root password                       |
| `DB_URL`              | ✅       | Full connection string untuk migrations   |

#### Redis Settings

| Variable         | Required | Description                               |
| ---------------- | -------- | ----------------------------------------- |
| `REDIS_HOST`     | ✅       | Redis host (gunakan `cache` untuk Docker) |
| `REDIS_PORT`     | ✅       | Redis port (default: `6379`)              |
| `REDIS_PASSWORD` | ✅       | Redis authentication password             |

#### Security Settings

| Variable     | Required | Description                           |
| ------------ | -------- | ------------------------------------- |
| `JWT_SECRET` | ✅       | Secret key untuk JWT token generation |

</details>

---

## 📚 Dokumentasi API

### 📖 Swagger Documentation

API documentation tersedia melalui Swagger UI:

**URL**: http://localhost:3001/api/v1/docs/

### 🔑 Authentication

API menggunakan **JWT (JSON Web Token)** untuk authentication:

1. Register user melalui endpoint `/register`
2. Login untuk mendapatkan JWT token
3. Include token di header: `Authorization: Bearer <your-token>`

### 📍 Endpoints Overview

<details>
<summary><b>Click to see available endpoints</b></summary>

#### User Endpoints

- `POST /api/users/register` - Register new user
- `POST /api/users/login` - Login user
- `GET /api/users/profile` - Get user profile (protected)
- `PUT /api/users/profile` - Update user profile (protected)

#### Product Endpoints

- `GET /api/products` - Get all products (with caching)
- `GET /api/products/:id` - Get product by ID (with caching)
- `POST /api/products` - Create new product (protected)
- `PUT /api/products/:id` - Update product (protected)
- `DELETE /api/products/:id` - Delete product (protected)

#### Transaction Endpoints

- `GET /api/transactions` - Get all transactions (protected)
- `GET /api/transactions/:id` - Get transaction by ID (protected)
- `POST /api/transactions` - Create new transaction (protected)

</details>

---

## 🧪 Testing

### Run Tests Locally

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -cover ./...

# Run specific test
go test -v ./internal/service/...
```

### Test Database Connection

```bash
go test -v ./test/connection.test.go
```

---

## 🐛 Troubleshooting

<details>
<summary><b>Common Issues & Solutions</b></summary>

### Issue: Container gagal start

**Solution**:

```bash
# Stop semua container
docker-compose down

# Remove volumes
docker-compose down -v

# Rebuild dan start ulang
docker-compose up --build
```

### Issue: Port already in use

**Solution**:

```bash
# Check port usage
netstat -ano | findstr :3001
netstat -ano | findstr :3307

# Kill process atau ubah port di .env dan docker-compose.yml
```

### Issue: Database migration failed

**Solution**:

```bash
# Check migration status dalam container
docker exec -it go-app-service /bin/sh
migrate -database "$DB_URL" -path db/migrations version

# Force version (hati-hati!)
migrate -database "$DB_URL" -path db/migrations force <version>
```

### Issue: Redis connection refused

**Solution**:

```bash
# Check Redis container
docker logs redis-cache-service

# Test Redis connection
docker exec -it redis-cache-service redis-cli
> AUTH your-redis-password
> PING
```

</details>

---

## 🎯 Caching Strategy

### Redis Implementation

- **Cache Layer**: Product repository
- **TTL**: 10 minutes
- **Cache Key Pattern**: `product:{id}`
- **Strategy**: Cache-aside (Lazy Loading)

### Cache Flow

```
1. Request product by ID
   ↓
2. Check Redis cache
   ↓
3a. Cache HIT → Return from Redis
3b. Cache MISS → Query MySQL → Store in Redis → Return
```

---

## 🚦 Development

### Local Development (without Docker)

<details>
<summary><b>Setup for local development</b></summary>

#### Prerequisites

- Go 1.22+
- MySQL 8.0
- Redis 7.0

#### Steps

1. Install dependencies:

```bash
go mod download
```

2. Setup local MySQL & Redis

3. Update `.env` dengan local configuration:

```env
MYSQL_HOST=localhost
REDIS_HOST=localhost
```

4. Run migrations:

```bash
migrate -database "mysql://user:pass@tcp(localhost:3306)/dbname" -path db/migrations up
```

5. Run application:

```bash
go run cmd/api/main.go
```

</details>

---

## 🤝 Kontribusi

Kontribusi sangat diterima! Silakan baca [Contribution Guidelines](./docs/CONTRIBUTING.md) untuk detail.

### 📝 How to Contribute

1. Fork repository ini
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

Jika proyek ini membantu Anda, berikan ⭐️ di [GitHub](https://github.com/assidik12/go-restfull-api)!

---

<div align="center">

**[Back to Top ⬆️](#-go-e-commerce-rest-api)**

Made with ❤️ using Go

</div>
