# Build stage — context: repo root (shopcaovanson/)
FROM golang:1.22-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git ca-certificates

COPY pkg/seedcatalog /pkg/seedcatalog
COPY services/order-service/go.mod services/order-service/go.sum ./
RUN go mod download

COPY services/order-service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /order-service .
RUN CGO_ENABLED=0 GOOS=linux go build -o /migrate ./cmd/migrate
RUN CGO_ENABLED=0 GOOS=linux go build -o /seed ./scripts/fake_data.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=builder /order-service /app/order-service
COPY --from=builder /migrate /app/migrate
COPY --from=builder /seed /app/seed
COPY services/order-service/migrations /app/migrations

EXPOSE 8083
CMD ["/app/order-service"]
