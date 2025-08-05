FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY shared/ ./shared/
COPY services/gateway/go.mod services/gateway/go.sum ./

# Створюємо правильну структуру для shared модуля
RUN mkdir -p ../../shared && cp -r ./shared/* ../../shared/

COPY services/gateway/ ./

RUN go mod tidy
RUN go mod download
RUN go build -o main ./cmd/gateway-service/main.go

FROM alpine:3.19

RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /app/main .
COPY services/gateway/.env ./

CMD ["./main"]