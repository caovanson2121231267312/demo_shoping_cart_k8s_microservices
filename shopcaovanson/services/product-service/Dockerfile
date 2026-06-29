# Build stage — context: repo root (shopcaovanson/)
FROM golang:1.22-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git ca-certificates

COPY pkg/seedcatalog /pkg/seedcatalog
COPY services/product-service/go.mod services/product-service/go.sum ./
RUN go mod download

COPY services/product-service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /product-service .
RUN CGO_ENABLED=0 GOOS=linux go build -o /migrate ./cmd/migrate
RUN CGO_ENABLED=0 GOOS=linux go build -o /seed ./scripts/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=builder /product-service /app/product-service
COPY --from=builder /migrate /app/migrate
COPY --from=builder /seed /app/seed
COPY services/product-service/migrations /app/migrations

EXPOSE 8082
CMD ["/app/product-service"]
