FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY proto/ ./proto/
COPY services/weather/go.mod services/weather/go.sum ./
COPY services/weather/ ./

RUN go mod tidy
RUN go mod download
RUN go build -o main ./cmd/weather-service/main.go

FROM alpine:3.19

RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /app/main .

CMD ["./main"] 