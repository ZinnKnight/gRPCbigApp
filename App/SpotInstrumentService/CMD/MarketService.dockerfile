FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" \
    -o /bin/market-service ./App/SpotInstrumentService/CMD

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/market-service /bin/market-service

EXPOSE 50052 2113

ENTRYPOINT ["/bin/market-service"]