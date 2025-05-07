# Build aşaması
FROM golang:1.20-alpine AS builder

# Gerekli bağımlılıklar
RUN apk add --no-cache gcc musl-dev git

WORKDIR /app

# Önce modülleri kopyala
COPY go.mod go.sum ./
RUN go mod download

# Tüm kaynak dosyaları kopyala
COPY . .

# Binary'yi derle
RUN go build -o privia_backend .

# Runtime aşaması
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Binary ve statik dosyaları kopyala
COPY --from=builder /app/privia_backend .
COPY --from=builder /app/son.html . # HTML dosyasını kopyala

# Railway için gerekli ortam değişkeni
ENV PORT=8080
EXPOSE 8080

# Çalıştırma komutu
CMD ["./privia_backend"]