FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" \
    -o /bin/order-service ./App/OrderService/CMD

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/order-service /bin/order-service

EXPOSE 50051 2112

ENTRYPOINT ["/bin/order-service"]