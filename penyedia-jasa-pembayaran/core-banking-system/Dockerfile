FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /cbs-server  ./cmd/server
RUN go build -o /cbs-migrate ./cmd/migrate

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /cbs-server  /usr/local/bin/cbs-server
COPY --from=builder /cbs-migrate /usr/local/bin/cbs-migrate
WORKDIR /app
COPY migrations ./migrations
COPY scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh
EXPOSE 50052
ENTRYPOINT ["entrypoint.sh"]
