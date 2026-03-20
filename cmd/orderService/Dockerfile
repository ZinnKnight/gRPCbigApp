FROM golang:1.26.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ../../dockerfiles/dockerOrderFile .

RUN CGO_ENABLED=0 GOOS=linux go build -o /order-service ./cmd/orderService/

FROM alpine:3.21

RUN apk --no-cache add ca-certificates

COPY --from=builder /order-service .

EXPOSE 50051 2112

CMD ["./order-service"]