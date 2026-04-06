# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build all binaries
COPY . .
RUN go build -o /core-banking-api      ./cmd/api
RUN go build -o /core-banking-migrate   ./cmd/migrate
RUN go build -o /core-banking-seed      ./cmd/seed

# Runtime stage
FROM alpine:3.18
RUN apk add --no-cache ca-certificates

COPY --from=builder /core-banking-api      /usr/local/bin/core-banking-api
COPY --from=builder /core-banking-migrate   /usr/local/bin/core-banking-migrate
COPY --from=builder /core-banking-seed      /usr/local/bin/core-banking-seed
WORKDIR /app

# Copy runtime assets
COPY migrations ./migrations
COPY .env .env
COPY certs ./certs
COPY scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["entrypoint.sh"]
