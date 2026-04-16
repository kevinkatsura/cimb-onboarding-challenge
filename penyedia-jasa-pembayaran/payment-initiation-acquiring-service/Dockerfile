FROM golang:1.26.2-alpine AS builder

WORKDIR /src
# Copy the shared proto module and the service directory
COPY proto ./proto
COPY payment-initiation-acquiring-service ./payment-initiation-acquiring-service

WORKDIR /src/payment-initiation-acquiring-service
# Resolve dependencies
RUN go mod download

# Build the application
RUN go build -o /pias-api     ./cmd/api
RUN go build -o /pias-migrate ./cmd/migrate

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /pias-api     /usr/local/bin/pias-api
COPY --from=builder /pias-migrate /usr/local/bin/pias-migrate
WORKDIR /app
COPY payment-initiation-acquiring-service/migrations ./migrations
COPY payment-initiation-acquiring-service/scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh
EXPOSE 8080
ENTRYPOINT ["entrypoint.sh"]
