# Сборка приложения
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

# Сборка исполняемого файла
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Финальный образ
FROM alpine:3.20 AS production

WORKDIR /app

# Установка зависимостей CA-сертификатов для HTTPS
RUN apk update && apk add --no-cache ca-certificates

# Копирование исполняемого файла из этапа сборки
COPY --from=builder /app/main .

# Копирование файлов миграции
COPY --from=builder /app/migrations ./migrations

# Открытие порта
EXPOSE 8080

# Определение команды запуска
ENTRYPOINT ["./main"]