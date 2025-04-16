# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git for go mod download if needed
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the API binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./cmd/api/main.go

# Build the worker binary (if needed, uncomment the next line)
# RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/worker ./cmd/worker/main.go

# Final image
FROM alpine:latest

WORKDIR /app

# Copy the built binary from builder
COPY --from=builder /bin/api /bin/api

# Copy migrations and other necessary files if needed
COPY migrations ./migrations

# Copy environment file
COPY deployment/docker/.env .env

EXPOSE 8080

# Run the API server by default
ENTRYPOINT ["/bin/api"]

