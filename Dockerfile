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

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o sitechecker .

# Запуск
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник
COPY --from=builder /app/sitechecker .

# Копируем static и templates из папки sitechecker
# После COPY sitechecker/ ./ в builder, эти папки находятся в /app/static и /app/templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates

EXPOSE 8080

CMD ["./sitechecker"]
