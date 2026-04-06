# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN go build -o /core-banking ./cmd/main.go

# Runtime stage
FROM alpine:3.18
RUN apk add --no-cache ca-certificates

COPY --from=builder /core-banking /usr/local/bin/core-banking
WORKDIR /app

# Copy migrations and .env into runtime image
COPY migrations ./migrations
COPY .env .env
COPY certs ./certs

EXPOSE 8080

CMD ["core-banking"]
