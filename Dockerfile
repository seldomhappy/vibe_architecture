# Stage 1: Builder
FROM golang:1.24-alpine AS builder

# Установка зависимостей для сборки
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Копируем go mod files
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Сборка бинарника
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o app \
    ./cmd/main.go

# Stage 2: Runtime
FROM alpine:3.18

# Установка CA сертификатов и timezone data
RUN apk --no-cache add ca-certificates tzdata

# Создаем непривилегированного пользователя
RUN addgroup -g 1001 appgroup && \
    adduser -D -u 1001 -G appgroup appuser

WORKDIR /app

# Копируем бинарник из builder
COPY --from=builder /build/app .

# Копируем конфиги (если нужны)
COPY --from=builder /build/config ./config

# Устанавливаем владельца
RUN chown -R appuser:appgroup /app

# Переключаемся на непривилегированного пользователя
USER appuser

# Expose порты
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Запуск с миграциями
ENV RUN_MIGRATIONS=true

CMD ["./app"]
