FROM golang:1.26.2-alpine AS builder

WORKDIR /app
# Flattened structure: copy service code and proto to subfolder
COPY services/core-banking-system .
COPY proto ./proto

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
COPY services/core-banking-system/migrations ./migrations
COPY services/core-banking-system/scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh
EXPOSE 50052
ENTRYPOINT ["entrypoint.sh"]
