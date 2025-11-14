# syntax=docker/dockerfile:1.4
FROM golang:1.24-alpine AS builder
LABEL authors="vizuth"

WORKDIR /build

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./

# Скачиваем зависимости с кэшированием
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Копируем исходный код
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY configs/ ./configs/
COPY migrations/ ./migrations/

# Собираем приложение с кэшированием сборки
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build -o avito-service ./cmd/main.go

# Финальный образ
FROM gcr.io/distroless/base-debian12 AS runner

WORKDIR /app

# Копируем конфигурацию, миграции и бинарник
COPY --from=builder /build/configs/ ./configs/
COPY --from=builder /build/migrations/ ./migrations/
COPY --from=builder /build/avito-service .

CMD ["./avito-service"]