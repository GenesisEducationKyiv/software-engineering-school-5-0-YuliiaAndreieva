FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY services/token/go.mod services/token/go.sum ./
COPY services/token/ ./

RUN go mod tidy
RUN go mod download
RUN go build -o main ./cmd/token-service/main.go

FROM alpine:3.19

RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /app/main .

CMD ["./main"] 