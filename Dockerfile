# Stage 1: Builder
FROM golang:alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod dan go.sum terlebih dahulu untuk caching dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh source code
COPY . .

# Build aplikasi dengan optimasi dari Makefile
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/build/shop-api ./cmd/api/main.go

# Stage 2: Runner (menggunakan alpine agar image akhir sangat kecil)
FROM alpine:latest

WORKDIR /root/

# Copy binary dari stage builder
COPY --from=builder /app/build/shop-api .

# Expose port (Cloud Run akan inject env PORT, default 8080)
EXPOSE 8080

# Jalankan aplikasi
CMD ["./shop-api"]