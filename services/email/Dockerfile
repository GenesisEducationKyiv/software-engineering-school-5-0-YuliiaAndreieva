FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY proto/ ./proto/
COPY services/email/go.mod services/email/go.sum ./
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