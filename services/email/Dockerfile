FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY proto/ ./proto/
COPY shared/ ./shared/
COPY services/email/go.mod services/email/go.sum ./

# Змінюємо replace директиви для Docker контексту
RUN echo "Before awk:" && cat go.mod && \
    echo "Checking for replace directives:" && grep "replace" go.mod && \
    awk '{gsub("../../proto", "./proto"); gsub("../../shared", "./shared"); print}' go.mod > go.mod.tmp && \
    mv go.mod.tmp go.mod && \
    echo "After awk:" && cat go.mod && \
    echo "Checking for replace directives after:" && grep "replace" go.mod

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