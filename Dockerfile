# =================
# Tahap 1: Builder
# =================
# Menggunakan base image Go versi 1.24-alpine untuk proses build.
FROM golang:1.24-alpine AS builder

LABEL maintainer="Ahmad Sofi Sidik <github.com/assidik12>"

# Menginstall paket-paket yang dibutuhkan untuk build
RUN apk add --no-cache git build-base

# Menentukan direktori kerja di dalam container
WORKDIR /app

# Copy file go.mod dan go.sum terlebih dahulu untuk memanfaatkan cache Docker.
COPY go.mod go.sum ./

# Men-download semua dependencies yang terdaftar di go.mod
RUN go mod download

# Menginstall golang-migrate/migrate
RUN go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1

# Copy seluruh sisa source code proyek ke dalam container
COPY . .

# --- PERUBAHAN DI SINI ---
# Melakukan build aplikasi Go dengan menunjuk ke path entrypoint yang baru.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/main ./cmd/api

# =================
# Tahap 2: Final Image
# =================
# Menggunakan base image alpine:latest yang sangat ringan untuk production.
FROM alpine:latest

# Menginstall netcat untuk healthcheck di entrypoint
RUN apk add --no-cache netcat-openbsd

WORKDIR /app

# Copy binary 'migrate' yang sudah di-build dari tahap 'builder'
COPY --from=builder /go/bin/migrate /usr/local/bin/

# Copy binary aplikasi utama yang sudah di-build dari tahap 'builder'
COPY --from=builder /app/main .

COPY .env .

# Copy folder migrasi
COPY ./db/migrations ./db/migrations

COPY ./docs ./docs

# Copy script entrypoint
COPY ./entrypoint.sh .

# Memberikan izin eksekusi pada script entrypoint
RUN chmod +x ./entrypoint.sh

EXPOSE 3000

ENTRYPOINT ["/app/entrypoint.sh"]
