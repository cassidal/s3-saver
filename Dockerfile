# Build stage
FROM golang:1.25-alpine AS builder

# Установка swag для генерации Swagger документации
RUN go install github.com/swaggo/swag/cmd/swag@latest

WORKDIR /app

# Копируем go mod файлы
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Генерируем Swagger документацию
RUN swag init -g cmd/s3-saver/main.go -o ./docs

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o s3-saver ./cmd/s3-saver

# Runtime stage
FROM alpine:latest

# Устанавливаем CA сертификаты для HTTPS запросов к S3
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем скомпилированный бинарник из builder stage
COPY --from=builder /app/s3-saver .

# Копируем конфигурационные файлы
COPY --from=builder /app/config ./config

# Устанавливаем правильные права доступа для конфигурационных файлов
RUN chmod -R 755 /root/config && \
    chmod 644 /root/config/local.yaml

# Открываем порт
EXPOSE 8080

# Переменные окружения по умолчанию (можно переопределить)
ENV APP_CONFIG_PATH=/root/config/local.yaml

# Запускаем приложение
CMD ["./s3-saver"]

