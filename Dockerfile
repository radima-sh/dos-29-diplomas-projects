# ========== ЭТАП 1: СБОРКА ==========
FROM golang:1.21-alpine AS builder

ARG REBUILD_TS

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Копируем go.mod и go.sum ИЗ ПАПКИ sitechecker
COPY sitechecker/go.mod sitechecker/go.sum ./
RUN go mod download

# Копируем весь исходный код приложения
COPY sitechecker/ ./

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o sitechecker .

# ========== ЭТАП 2: ЗАПУСК ==========
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник
COPY --from=builder /app/sitechecker .

# Копируем static и templates ИЗ ПАПКИ sitechecker
# После COPY sitechecker/ ./ в builder, эти папки находятся в /app/static и /app/templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates

EXPOSE 8080

CMD ["./sitechecker"]
