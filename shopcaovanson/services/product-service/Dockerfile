FROM golang:1.22-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /product-service .
RUN CGO_ENABLED=0 GOOS=linux go build -o /migrate ./cmd/migrate
RUN CGO_ENABLED=0 GOOS=linux go build -o /seed ./scripts/fake_data.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=builder /product-service /app/product-service
COPY --from=builder /migrate /app/migrate
COPY --from=builder /seed /app/seed
COPY migrations /app/migrations

EXPOSE 8082
CMD ["/app/product-service"]
