FROM golang:1.26.2-alpine AS builder

WORKDIR /src
# Copy the shared proto module and the service directory
COPY proto ./proto
COPY account-issuance-service ./account-issuance-service

WORKDIR /src/account-issuance-service
# Resolve dependencies
RUN go mod download

# Build the application
RUN go build -o /ais-api     ./cmd/api
RUN go build -o /ais-migrate  ./cmd/migrate

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /ais-api     /usr/local/bin/ais-api
COPY --from=builder /ais-migrate /usr/local/bin/ais-migrate
WORKDIR /app
COPY account-issuance-service/migrations ./migrations
COPY account-issuance-service/scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh
EXPOSE 8080 50051
ENTRYPOINT ["entrypoint.sh"]
