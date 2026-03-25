# --- ЭТАП 1: Сборка (Builder) ---
# мы даем этому этапу имя "builder",
# чтобы ссылаться на него позже
FROM golang:1.26-alpine AS builder
WORKDIR /app
# Кэшируем зависимости
ENV GONOSUMDB="*"
ENV GONOPROXY=""
ENV GOFLAGS="-mod=mod"
COPY go.mod go.sum ./
RUN go mod download
# Копируем код и собираем.
COPY . .
# Собираем статический бинарник с именем telegram-service
RUN go build -o telegram-service cmd/main.go
# --- ЭТАП 2: Запуск (Production) ---
# Внимание! Новый FROM обнуляет всё, что было до этого
# Мы начинаем с чистого листа
FROM alpine:latest
RUN apk add --no-cache tzdata ca-certificates
WORKDIR /app
# Мы копируем НЕ с локальной машины, а из этапа "builder"
# И берем ТОЛЬКО один файл
COPY --from=builder /app/telegram-service .
# Запускаем
CMD ["./telegram-service"]
