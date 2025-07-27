FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN go mod tidy
RUN go build -o main ./cmd/token-service/main.go

FROM alpine:3.19

RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8083

CMD ["./main"] 