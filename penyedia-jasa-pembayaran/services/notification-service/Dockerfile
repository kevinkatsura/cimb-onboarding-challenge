FROM golang:1.26.2-alpine AS builder

WORKDIR /app
# Preserve structure: copy service code and proto to their respective folders
COPY proto ./proto
COPY services/notification-service ./services/notification-service

WORKDIR /app/services/notification-service

# Resolve dependencies
RUN go mod download

RUN go build -o /ns-consumer ./cmd/consumer
RUN go build -o /ns-migrate  ./cmd/migrate

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /ns-consumer /usr/local/bin/ns-consumer
COPY --from=builder /ns-migrate  /usr/local/bin/ns-migrate
WORKDIR /app/services/notification-service
COPY services/notification-service/migrations ./migrations
COPY services/notification-service/scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh
ENTRYPOINT ["entrypoint.sh"]
