FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /bin/market-service ./CMD/MarketServiceCMD

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/market-service /bin/market-service

ENTRYPOINT ["/bin/market-service"]