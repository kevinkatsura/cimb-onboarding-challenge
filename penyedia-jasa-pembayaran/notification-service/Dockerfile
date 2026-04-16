FROM golang:1.26.2-alpine AS builder

WORKDIR /src
# Copy the service directory
COPY notification-service ./notification-service

WORKDIR /src/notification-service
# Resolve dependencies
RUN go mod download

# Build the application
RUN go build -o /ns-consumer ./cmd/consumer
RUN go build -o /ns-migrate  ./cmd/migrate

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /ns-consumer /usr/local/bin/ns-consumer
COPY --from=builder /ns-migrate  /usr/local/bin/ns-migrate
WORKDIR /app
COPY notification-service/migrations ./migrations
COPY notification-service/scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh
EXPOSE 8080
ENTRYPOINT ["entrypoint.sh"]
