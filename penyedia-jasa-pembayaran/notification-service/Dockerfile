FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /ns-consumer ./cmd/consumer
RUN go build -o /ns-migrate  ./cmd/migrate

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /ns-consumer /usr/local/bin/ns-consumer
COPY --from=builder /ns-migrate  /usr/local/bin/ns-migrate
WORKDIR /app
COPY migrations ./migrations
COPY scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh
EXPOSE 8080
ENTRYPOINT ["entrypoint.sh"]
