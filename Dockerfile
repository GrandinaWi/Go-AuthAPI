# 1. Сборка
FROM golang:1.25-alpine AS builder

WORKDIR /app

# зависимости
COPY go.mod go.sum ./
RUN go mod download

# код
COPY . .

# сборка бинарника
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o app ./cmd/api

# 2. Рантайм
FROM alpine:3.20

WORKDIR /app

# сертификаты для HTTPS
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/app .

EXPOSE 8080

CMD ["./app"]
