# --- Tahap 1: Build (Membuat file executable) ---
FROM golang:1.20-alpine AS builder

# Set direktori kerja di dalam container
WORKDIR /app

# Salin file manajemen dependensi dan unduh
COPY go.mod go.sum ./
RUN go mod download

# Salin seluruh source code proyek Anda
COPY . .

# Build aplikasi Go menjadi file binary bernama "gateway-app"
RUN CGO_ENABLED=0 GOOS=linux go build -o gateway-app ./cmd/gateway/main.go

# --- Tahap 2: Run (Menjalankan aplikasi di OS yang super ringan) ---
FROM alpine:latest

WORKDIR /app

# Ambil hasil build dari Tahap 1
COPY --from=builder /app/gateway-app .
# Salin file konfigurasi .env (dan folder crypto-config jika ada)
COPY .env .
# COPY crypto-config/ ./crypto-config/  <-- Buka komentar ini jika sertifikat Fabric ada di folder ini

# Buka port API Anda
EXPOSE 3000

# Perintah utama saat container dinyalakan
CMD ["./gateway-app"]