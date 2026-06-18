FROM golang:1.22-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /chat-service .
RUN CGO_ENABLED=0 GOOS=linux go build -o /migrate ./cmd/migrate
RUN CGO_ENABLED=0 GOOS=linux go build -o /seed ./scripts/fake_data.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=builder /chat-service /app/chat-service
COPY --from=builder /migrate /app/migrate
COPY --from=builder /seed /app/seed

EXPOSE 8084
CMD ["/app/chat-service"]
