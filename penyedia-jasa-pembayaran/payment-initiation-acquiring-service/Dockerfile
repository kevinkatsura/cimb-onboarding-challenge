FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /pias-api     ./cmd/api
RUN go build -o /pias-migrate ./cmd/migrate

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /pias-api     /usr/local/bin/pias-api
COPY --from=builder /pias-migrate /usr/local/bin/pias-migrate
WORKDIR /app
COPY migrations ./migrations
COPY scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh
EXPOSE 8080
ENTRYPOINT ["entrypoint.sh"]
