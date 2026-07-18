# Build stage — context: repo root (shopcaovanson/)
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY pkg/seedcatalog /pkg/seedcatalog
COPY services/auth-service/go.mod services/auth-service/go.sum ./
RUN go mod download

COPY services/auth-service/ .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /auth-service .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /migrate ./cmd/migrate
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /login-report ./cmd/login-report
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /seed ./scripts/fake_data.go

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /auth-service /app/auth-service
COPY --from=builder /migrate /app/migrate
COPY --from=builder /login-report /app/login-report
COPY --from=builder /seed /app/seed
COPY services/auth-service/migrations /app/migrations

ENV TZ=Asia/Ho_Chi_Minh

EXPOSE 8081

CMD ["/app/auth-service"]
