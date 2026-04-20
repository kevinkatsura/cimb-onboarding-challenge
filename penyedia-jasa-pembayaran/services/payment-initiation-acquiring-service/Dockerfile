FROM golang:1.26.2-alpine AS builder

WORKDIR /app
# Flattened structure: copy service code and proto to subfolder
COPY services/payment-initiation-acquiring-service .
COPY proto ./proto

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
COPY services/payment-initiation-acquiring-service/migrations ./migrations
COPY services/payment-initiation-acquiring-service/scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh
EXPOSE 8080
ENTRYPOINT ["entrypoint.sh"]
