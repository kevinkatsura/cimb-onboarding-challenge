FROM golang:1.25.0-alpine AS builder

WORKDIR /app

# Enable CGO for some dependencies if needed, or disable
ENV CGO_ENABLED=0
ENV GOOS=linux

# Copy all the workspace
COPY go.work go.work.sum ./
COPY proto ./proto
COPY account-information-service ./account-information-service
COPY vendor ./vendor

WORKDIR /app/account-information-service
RUN go build -mod=vendor -o main cmd/api/main.go

# Run stage
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/account-information-service/main .
COPY --from=builder /app/account-information-service/migrations ./migrations

EXPOSE 8081 50055
CMD ["./main"]
