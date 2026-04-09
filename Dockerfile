# --- Tahap 1: Build (Membuat file .exe / binary) ---
FROM golang:1.20-alpine AS builder

WORKDIR /app

# Kopi file dependensi dan download
COPY go.mod go.sum ./
RUN go mod download

# Kopi seluruh source code
COPY . .

# Build aplikasi Go menjadi file binary bernama "gateway-app"
RUN CGO_ENABLED=0 GOOS=linux go build -o gateway-app ./cmd/gateway/main.go

# --- Tahap 2: Run (Menjalankan aplikasi di OS yang super ringan) ---
FROM alpine:latest

WORKDIR /app

# Ambil hasil build dari Tahap 1
COPY --from=builder /app/gateway-app .
# Kopi file konfigurasi .env
COPY .env .

# Buka port 3000
EXPOSE 3000

# Perintah saat kontainer dinyalakan
CMD ["./gateway-app"]