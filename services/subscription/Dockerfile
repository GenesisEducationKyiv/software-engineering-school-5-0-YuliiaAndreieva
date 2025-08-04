FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY proto/ ./proto/
COPY services/subscription/go.mod services/subscription/go.sum ./
COPY services/subscription/ ./

RUN go mod tidy
RUN go mod download
RUN go build -o main ./cmd/subscription-service/main.go

FROM alpine:3.19

RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /app/main .

CMD ["./main"]