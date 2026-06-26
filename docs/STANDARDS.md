# Code Standards — Catalyst

> Dokumen ini adalah **single source of truth** untuk standar penulisan kode di proyek Catalyst.
> Setiap PR yang tidak mengikuti standar ini akan di-reject oleh Code Reviewer.
>
> Untuk memahami *mengapa* standar ini ada, baca [`CONCEPTS.md`](./CONCEPTS.md).
> Untuk memahami gambaran besar arsitektur, baca [`ARCHITECTURE.md`](./ARCHITECTURE.md).

---

## Daftar Isi

1. [Penamaan File](#1-penamaan-file)
2. [Package & Struktur Direktori](#2-package--struktur-direktori)
3. [Import Order](#3-import-order)
4. [Naming Convention](#4-naming-convention)
5. [Error Handling](#5-error-handling)
6. [Layer Rules](#6-layer-rules)
7. [Database Transaction](#7-database-transaction)
8. [Goroutine & Concurrency](#8-goroutine--concurrency)
9. [Logging](#9-logging)
10. [DTO & Validation](#10-dto--validation)
11. [HTTP Handler](#11-http-handler)
12. [Testing](#12-testing)
13. [Commit Message](#13-commit-message)

---

## 1. Penamaan File

| Layer | Pola Nama | Contoh |
|---|---|---|
| Service | `<entity>.service.go` | `product.service.go` |
| Repository | `<entity>.repository.go` | `product.repository.go` |
| Handler | `<entity>.handler.go` | `product.handler.go` |
| DTO | `<entity>.go` | `product.go` |
| Domain entity | `<entity>.go` | `product.go` |
| Domain interface | `port.go` | `port.go` |
| Domain error | `errors.go` | `errors.go` |
| Test | `<entity>_test.go` | `product_service_test.go` |

**Aturan:**
- Gunakan **lowercase** dan **underscore** sebagai pemisah kata dalam nama file.
- Jangan gunakan `camelCase` atau `PascalCase` pada nama file.

---

## 2. Package & Struktur Direktori

```
internal/
├── domain/           # package domain  — entity + interface + sentinel error
│   ├── errors.go
│   ├── port.go
│   ├── product.go
│   └── ...
├── service/          # package service  — business logic
│   ├── product.service.go
│   └── ...
├── repository/
│   ├── mysql/        # package mysql    — SQL implementation
│   └── redis/        # package redis    — cache implementation
├── delivery/
│   └── http/
│       ├── handler/  # package handler  — HTTP handler
│       ├── dto/      # package dto      — request/response struct
│       ├── middleware/
│       └── route/
├── infrastructure/   # package infrastructure — DB, Redis, Kafka init
├── event/            # package event    — Kafka event struct + producer interface
├── worker/           # package worker   — Kafka consumer / background worker
└── pkg/              # package <name>   — shared utility (response, dll)
```

**Aturan:**
- Setiap subdirektori = satu Go package. Nama package **harus sama** dengan nama direktori.
- `internal/domain` tidak boleh mengimport package dari layer manapun.
- Dependency hanya boleh mengalir ke dalam (inward): `handler → service → domain ← repository`.

---

## 3. Import Order

Kelompokkan import menjadi **3 blok** yang dipisahkan baris kosong, **diurutkan secara alfabetikal** dalam tiap blok:

```go
import (
    // 1. Standard library
    "context"
    "database/sql"
    "fmt"
    "log/slog"

    // 2. External / third-party
    "github.com/go-playground/validator/v10"
    "github.com/julienschmidt/httprouter"
    "go.opentelemetry.io/otel"

    // 3. Internal
    "github.com/assidik12/catalyst/internal/domain"
    "github.com/assidik12/catalyst/internal/service"
)
```

> ✅ `goimports` atau `gofmt` dapat menegakkan ini secara otomatis.

---

## 4. Naming Convention

### Variabel & Fungsi
| Jenis | Convention | Contoh |
|---|---|---|
| Variabel lokal | `camelCase` | `totalPrice`, `userID` |
| Fungsi/method exported | `PascalCase` | `GetProductById` |
| Fungsi/method unexported | `camelCase` | `handleServiceError` |
| Konstanta | `camelCase` (unexported) atau `PascalCase` (exported) | `defaultCacheDuration` |
| Interface | `PascalCase`, akhiran `-er` atau `-Service`/`-Repository` | `ProductService`, `Producer` |
| Struct konkret | `camelCase` (unexported) | `productService`, `productRepository` |

### Khusus
- **ID** ditulis `ID`, bukan `Id` atau `id` (kecuali dalam json tag: `"id"`).
- **URL** ditulis `URL`, bukan `Url`.
- **Receiver** menggunakan **1–2 huruf** dari nama tipe: `(p *productService)`, `(ph *ProductHandler)`.
- Hindari nama generik seperti `data`, `result`, `temp` — gunakan nama yang deskriptif.

---

## 5. Error Handling

### Sentinel Errors

Semua error yang dikembalikan service **wajib** menggunakan sentinel dari `internal/domain/errors.go`:

| Sentinel | HTTP Status | Kapan Digunakan |
|---|---|---|
| `domain.ErrNotFound` | 404 | Resource tidak ditemukan di database |
| `domain.ErrInvalidInput` | 400 | Input dari caller tidak valid atau constraint gagal |
| `domain.ErrUnauthorized` | 401 | Caller tidak punya izin |
| `domain.ErrConflict` | 409 | Duplicate resource atau stok habis |

### Wrap Error

```go
// ✅ BENAR — wrap dengan konteks dan sentinel
return fmt.Errorf("ProductService.GetByID: %w", domain.ErrNotFound)
return fmt.Errorf("%w: product id must be positive", domain.ErrInvalidInput)

// ❌ SALAH — string mentah tanpa sentinel
return errors.New("product not found")
return fmt.Errorf("invalid input")
```

### Error Mapping di Handler

Handler **wajib** menggunakan `errors.Is()` untuk memetakan error ke HTTP status code. Gunakan helper `handleServiceError` (lihat contoh di section Handler):

```go
// ✅ BENAR
switch {
case errors.Is(err, domain.ErrNotFound):
    response.NotFound(w, err.Error())
case errors.Is(err, domain.ErrInvalidInput):
    response.BadRequest(w, err.Error())
// ...
}

// ❌ SALAH — string comparison
if err.Error() == "not found" { ... }
```

### Jangan Swallow Error

```go
// ❌ SALAH — error diabaikan diam-diam
_ = p.cache.Set(ctx, key, data, ttl)

// ✅ BENAR — minimal di-log jika error non-critical
if err := p.cache.Set(ctx, key, data, ttl); err != nil {
    slog.Warn("failed to set cache", "key", key, "error", err)
}
```

---

## 6. Layer Rules

### Domain Layer (`internal/domain`)
- **Hanya boleh** berisi: entity struct, interface (port), sentinel error.
- **TIDAK BOLEH** mengimport package apapun selain stdlib (`errors`, `context`, `database/sql`).
- Interface didefinisikan di sini, implementasinya ada di layer lain.

### Service Layer (`internal/service`)
- **TIDAK BOLEH** mengimport `database/sql` driver secara langsung.
- Hanya bergantung ke `domain.XxxRepository` interface (bukan struct konkret `mysql.xxxRepository`).
- Seluruh business logic (kalkulasi harga, validasi stok, pemberian UUID) ada di sini.
- Validasi input menggunakan `go-playground/validator`, bukan validasi manual ad-hoc.

```go
// ✅ BENAR — bergantung ke interface
type productService struct {
    repo domain.ProductRepository // ← interface dari domain
}

// ❌ SALAH — bergantung ke implementasi konkret
type productService struct {
    repo *mysql.ProductRepository // ← struct dari package lain
}
```

### Repository Layer (`internal/repository`)
- Satu-satunya layer yang boleh mengandung query SQL dan operasi cache.
- Menerima `*sql.Tx` sebagai parameter **dari service**, bukan membuat transaksi sendiri.
- Memetakan `sql.ErrNoRows` ke `domain.ErrNotFound` sebelum dikembalikan ke caller.

```go
// ✅ BENAR — map error di repo
if err == sql.ErrNoRows {
    return domain.Product{}, domain.ErrNotFound
}
```

### Handler Layer (`internal/delivery/http/handler`)
- **TIDAK BOLEH** mengandung business logic atau query SQL.
- Tanggung jawab: decode request body → panggil service → encode response.
- Selalu gunakan `r.Context()` saat memanggil service (jangan buat context baru).
- Gunakan helper `handleServiceError` untuk memetakan error ke HTTP response.

---

## 7. Database Transaction

### Pola Wajib

```go
// ✅ BENAR — transaction dibuka di service, di-pass ke repo
tx, err := s.DB.BeginTx(ctx, nil)
if err != nil {
    return domain.Transaction{}, err
}
defer tx.Rollback() // no-op jika sudah Commit

// ... operasi repo menggunakan tx ...

if err := tx.Commit(); err != nil {
    return domain.Transaction{}, err
}
```

### Conditional UPDATE untuk Operasi Stok

**Selalu** gunakan atomic SQL UPDATE dengan constraint, bukan read-then-write:

```go
// ✅ BENAR — atomic, race-condition safe
q := `UPDATE products SET stock = stock - ? WHERE id = ? AND stock >= ?`
result, err := tx.ExecContext(ctx, q, qty, productID, qty)
rowsAffected, _ := result.RowsAffected()
if rowsAffected == 0 {
    return domain.ErrInvalidInput // insufficient stock
}

// ❌ SALAH — race condition: dua request bisa lolos cek stok bersamaan
product := repo.FindById(ctx, id)  // read
if product.Stock < qty { return err }
repo.DecrementStock(ctx, id, qty)  // write (terpisah dari read!)
```

---

## 8. Goroutine & Concurrency

### Fire-and-Forget

Goroutine yang dilepas (fire-and-forget) **wajib** membuat context baru dengan timeout sendiri. **Jangan** mewarisi context request HTTP yang bisa di-cancel oleh client.

```go
// ✅ BENAR — context independen dari request
go func() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := s.producer.PublishEvent(ctx, event); err != nil {
        slog.Error("failed to publish event", "error", err)
    }
}()

// ❌ SALAH — ctx dari request bisa di-cancel sebelum goroutine selesai
go func() {
    s.producer.PublishEvent(ctx, event) // ctx bisa expired
}()
```

### Singleflight untuk Cache Stampede

Gunakan `singleflight.Group` untuk operasi cache miss yang berat (DB query):

```go
res, err, _ := sf.Do(cacheKey, func() (any, error) {
    data, err := repo.FindById(ctx, id)
    if err != nil {
        return nil, err
    }
    cache.Set(ctx, cacheKey, data, defaultCacheDuration)
    return data, nil
})
```

### Shared State

- Akses shared state (counter, map) dari multiple goroutine **wajib** menggunakan `sync.Mutex` atau `sync/atomic`.
- Gunakan `go test -race` untuk mendeteksi race condition.

---

## 9. Logging

Gunakan **structured logging** via `log/slog` (stdlib). Hindari `fmt.Println` atau `log.Printf` di production code.

```go
// ✅ BENAR — structured, queryable
slog.Error("failed to publish event",
    "transaction_id", txID,
    "error", err,
)
slog.Info("server starting", "port", cfg.Port)

// ❌ SALAH — plain text, tidak queryable
log.Printf("Error publishing event: %v", err)
fmt.Println("Server starting on port 8080")
```

**Aturan:**
- Level `INFO` — event normal yang penting (server start, request masuk).
- Level `WARN` — sesuatu yang tidak normal tapi tidak memblokir operasi (cache miss yang berulang).
- Level `ERROR` — operasi gagal yang perlu investigasi.
- **Jangan** log data sensitif: password, JWT token, API key, PII.

---

## 10. DTO & Validation

### Struktur DTO

DTO hanya boleh berisi field untuk transport data. Tidak ada business logic.

```go
// internal/delivery/http/dto/product.go
package dto

type ProductRequest struct {
    Name        string  `json:"name"        validate:"required,min=3,max=100"`
    Price       int     `json:"price"       validate:"required,gt=0"`
    Description string  `json:"description" validate:"required"`
    Img         string  `json:"img"         validate:"required,url"`
    CategoryID  int     `json:"category_id" validate:"required,gt=0"`
    Stock       int     `json:"stock"       validate:"required,gte=0"`
}
```

**Aturan:**
- Selalu sertakan `validate` tag untuk field yang wajib divalidasi.
- `json:"-"` untuk field yang tidak boleh di-serialize/deserialize (misal: `IdempotencyKey` yang diambil dari header).
- DTO **tidak boleh** diimport oleh `internal/domain`.

### Validasi

Validasi struct dilakukan **di service layer** menggunakan `validator.Validate`:

```go
// ✅ BENAR — validasi di service
if err := s.validator.Struct(req); err != nil {
    return domain.Product{}, fmt.Errorf("%w: %s", domain.ErrInvalidInput, err.Error())
}
```

---

## 11. HTTP Handler

### Pola Standar

```go
// Handler struct — satu service per handler file
type ProductHandler struct {
    service service.ProductService
}

func NewProductHandler(service service.ProductService) *ProductHandler {
    return &ProductHandler{service: service}
}

// Method handler — urutan: decode → call service → encode response
func (ph *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    // 1. Decode
    var req dto.ProductRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.BadRequest(w, "invalid request body")
        return
    }

    // 2. Call service (gunakan r.Context())
    product, err := ph.service.CreateProduct(r.Context(), req)
    if err != nil {
        ph.handleServiceError(w, err)
        return
    }

    // 3. Encode response
    response.Created(w, product)
}

// Error mapping — selalu gunakan errors.Is
func (ph *ProductHandler) handleServiceError(w http.ResponseWriter, err error) {
    switch {
    case errors.Is(err, domain.ErrNotFound):
        response.NotFound(w, err.Error())
    case errors.Is(err, domain.ErrInvalidInput):
        response.BadRequest(w, err.Error())
    case errors.Is(err, domain.ErrUnauthorized):
        response.Unauthorized(w, err.Error())
    case errors.Is(err, domain.ErrConflict):
        response.Conflict(w, err.Error())
    default:
        response.InternalServerError(w, "internal server error")
    }
}
```

### Response HTTP Status Codes

| Kondisi | Status Code | Helper |
|---|---|---|
| Berhasil baca/update/delete | 200 | `response.OK(w, data)` |
| Berhasil create resource baru | 201 | `response.Created(w, data)` |
| Input tidak valid | 400 | `response.BadRequest(w, msg)` |
| Token tidak ada / invalid | 401 | `response.Unauthorized(w, msg)` |
| Resource tidak ditemukan | 404 | `response.NotFound(w, msg)` |
| Conflict (duplicate, stok habis) | 409 | `response.Conflict(w, msg)` |
| Error tak terduga | 500 | `response.InternalServerError(w, msg)` |

---

## 12. Testing

### Direktori

```
test/
├── service/          # unit test service layer (pakai testify/mock)
├── repository/       # unit test repository layer (pakai go-sqlmock)
├── handler/          # unit test handler layer (pakai httptest)
└── mocks/            # generated/manual mock untuk domain interface
```

### Aturan Wajib

1. Setiap test file **harus punya minimal 1 failure scenario** (bukan hanya happy path).
2. **Jangan** gunakan `time.Sleep()` untuk mensimulasikan timing — gunakan `context.WithTimeout` atau channel.
3. Test harus **deterministik** — tidak boleh flaky karena ordering atau timing.
4. Jalankan `go test -race ./...` untuk semua test yang melibatkan goroutine.
5. **Jangan** ubah production code hanya untuk membuat test bisa kompilasi — koordinasikan dengan Code Execution Agent.

### Nama Test

Format: `Test<TypeName>_<MethodName>_<Scenario>`

```go
func TestProductService_GetByID_Success(t *testing.T) { ... }
func TestProductService_GetByID_NotFound(t *testing.T) { ... }
func TestProductService_GetByID_ContextCancelled(t *testing.T) { ... }
```

### Run Commands

```bash
# Semua test
go test ./...

# Dengan race detector
go test -race ./...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Test spesifik
go test -v -run TestProductService ./test/service/...
```

---

## 13. Commit Message

Ikuti format **Conventional Commits**:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Tipe yang Digunakan

| Tipe | Kapan |
|---|---|
| `feat` | Fitur baru |
| `fix` | Bug fix |
| `refactor` | Perubahan kode tanpa fitur baru atau bug fix |
| `test` | Menambah atau memperbaiki test |
| `docs` | Perubahan dokumentasi saja |
| `chore` | Maintenance (update dependency, config, dll) |
| `perf` | Peningkatan performa |

### Contoh

```bash
feat(transaction): add idempotency key support
fix(product): prevent negative stock on concurrent decrement
test(user): add missing failure scenarios for user service
docs(standards): add commit message conventions
refactor(handler): extract handleServiceError to shared helper
```

**Aturan:**
- `<description>` ditulis dalam **Bahasa Inggris**, **imperative mood** ("add", bukan "added" atau "adds").
- Maksimum 72 karakter pada baris pertama.
- Scope menggunakan nama domain/layer: `product`, `transaction`, `user`, `handler`, `repo`, dll.

---

## Bahasa

| Konteks | Bahasa |
|---|---|
| Diskusi, PR description, komentar di Slack/GitHub | Bahasa Indonesia |
| Kode Go (nama variabel, fungsi, struct) | Bahasa Inggris |
| Komentar dalam kode (`//`) | Bahasa Inggris |
| Commit message | Bahasa Inggris |
| Nama file dan direktori | Bahasa Inggris |
