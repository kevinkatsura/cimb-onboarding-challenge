FROM golang:1.26.2-alpine AS builder

WORKDIR /app
# Flattened structure: copy service code and proto to subfolder
COPY services/account-issuance-service .
COPY proto ./proto

# Resolve dependencies
RUN go mod download

# Build the application
RUN go build -o /ains-api     ./cmd/api
RUN go build -o /ains-migrate  ./cmd/migrate

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /ains-api     /usr/local/bin/ains-api
COPY --from=builder /ains-migrate /usr/local/bin/ains-migrate
WORKDIR /app
COPY services/account-issuance-service/migrations ./migrations
COPY services/account-issuance-service/scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh
EXPOSE 8080 50051
ENTRYPOINT ["entrypoint.sh"]
