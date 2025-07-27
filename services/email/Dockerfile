FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./

COPY . .

WORKDIR /app

RUN go mod tidy
RUN go mod download

RUN go build -o main ./cmd/email-service/main.go

FROM alpine:3.19
RUN apk add --no-cache tzdata
WORKDIR /app
COPY --from=builder /app/main .
COPY .env .

EXPOSE 8081

CMD ["./main"] 