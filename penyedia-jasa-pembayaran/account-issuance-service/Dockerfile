FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /ais-api     ./cmd/api
RUN go build -o /ais-migrate  ./cmd/migrate

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /ais-api     /usr/local/bin/ais-api
COPY --from=builder /ais-migrate /usr/local/bin/ais-migrate
WORKDIR /app
COPY migrations ./migrations
COPY scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh
EXPOSE 8080 50051
ENTRYPOINT ["entrypoint.sh"]
