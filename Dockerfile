# Сборка
FROM golang:1.21-alpine AS builder

ARG REBUILD_TS

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Копируем go.mod и go.sum из папки sitechecker
COPY sitechecker/go.mod sitechecker/go.sum ./
RUN go mod download

# Копируем весь исходный код приложения
COPY sitechecker/ ./

# Собираем бинарник (используем REBUILD_TS для сброса кэша)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.BuildTimestamp=${REBUILD_TS}" -o sitechecker .

# Запуск
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Копируем бинарник
COPY --from=builder /app/sitechecker .

# Копируем static и templates из папки sitechecker
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates

EXPOSE 8080

CMD ["./sitechecker"]
