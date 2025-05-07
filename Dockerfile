# Build aşaması
FROM golang:1.20-alpine AS builder

# Gerekli paketleri yükle
RUN apk add --no-cache gcc musl-dev

# Çalışma dizini
WORKDIR /app

# Go modül dosyalarını kopyala ve bağımlılıkları indir
COPY go.mod go.sum ./
RUN go mod download

# Tüm kaynak kodları kopyala
COPY . .

# Uygulamayı derle
RUN go build -o todoapp .

# Final imaj
FROM alpine:latest

# Sertifikaları yükle
RUN apk --no-cache add ca-certificates

# Çalışma dizini
WORKDIR /root/

# Derlenen uygulamayı kopyala
COPY --from=builder /app/todoapp .

# Railway için port ayarla
ENV PORT=8080
EXPOSE 8080

# Uygulamayı başlat
CMD ["./todoapp"]
