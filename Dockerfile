# Этап сборки
FROM golang:1.25-alpine AS builder

# Устанавливаем необходимые пакеты для CGO (для SQLite)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite

WORKDIR /root/

# Копируем бинарник
COPY --from=builder /app/main .

# Копируем миграции
COPY --from=builder /app/migrations ./migrations

EXPOSE 4000

CMD ["./main"]
