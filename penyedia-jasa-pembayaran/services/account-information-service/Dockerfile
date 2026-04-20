FROM golang:1.26.2-alpine AS builder

WORKDIR /app
# Preserve structure: copy service code and proto to their respective folders
COPY proto ./proto
COPY services/account-information-service ./services/account-information-service

WORKDIR /app/services/account-information-service

# Resolve dependencies
RUN go mod download

# Build the application
RUN go build -o /ais-api     ./cmd/api
RUN go build -o /ais-migrate  ./cmd/migrate

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /ais-api     /usr/local/bin/ais-api
COPY --from=builder /ais-migrate /usr/local/bin/ais-migrate
WORKDIR /app/services/account-information-service
COPY services/account-information-service/migrations ./migrations
COPY services/account-information-service/scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh
EXPOSE 8080 50051
ENTRYPOINT ["entrypoint.sh"]
