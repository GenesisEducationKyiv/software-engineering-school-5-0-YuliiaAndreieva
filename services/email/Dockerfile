FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY proto/ ./proto/
COPY shared/ ./shared/
COPY services/email/go.mod services/email/go.sum ./

# Створюємо правильну структуру для shared та proto модулів
RUN mkdir -p ../../shared && cp -r ./shared/* ../../shared/ && \
    mkdir -p ../../proto && cp -r ./proto/* ../../proto/

COPY services/email/ ./

RUN go mod tidy
RUN go mod download
RUN go build -o main ./cmd/email-service/main.go

FROM alpine:3.19

RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /app/main .
COPY services/email/.env ./

CMD ["./main"] 