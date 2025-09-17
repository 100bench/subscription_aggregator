# Сборка приложения
FROM golang:1.22-alpine AS builder

WORKDIR /app

ENV GOTOOLCHAIN=auto

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

# Сборка исполняемого файла
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Финальный образ (используем тот же golang:alpine, чтобы не тянуть alpine отдельно)
FROM golang:1.22-alpine AS production

WORKDIR /app

# Копирование исполняемого файла из этапа сборки
COPY --from=builder /app/main .

# Копирование файлов миграции
COPY --from=builder /app/migrations ./migrations

# Открытие порта
EXPOSE 8080

# Определение команды запуска
ENTRYPOINT ["./main"]