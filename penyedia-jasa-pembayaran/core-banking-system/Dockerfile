FROM golang:1.26.2-alpine AS builder

WORKDIR /src
# Copy the shared proto module and the service directory
COPY proto ./proto
COPY core-banking-system ./core-banking-system

WORKDIR /src/core-banking-system
# Resolve dependencies
RUN go mod download

# Build the application
RUN go build -o /cbs-server  ./cmd/server
RUN go build -o /cbs-migrate ./cmd/migrate

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /cbs-server  /usr/local/bin/cbs-server
COPY --from=builder /cbs-migrate /usr/local/bin/cbs-migrate
WORKDIR /app
COPY core-banking-system/migrations ./migrations
COPY core-banking-system/scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh
EXPOSE 50052
ENTRYPOINT ["entrypoint.sh"]
