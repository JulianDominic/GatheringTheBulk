# Build Stage
FROM golang:1.25-alpine AS builder

# Set working directory
WORKDIR /app

# Install necessary build tools if needed
# (modernc.org/sqlite is CGO-free, so we don't need gcc)
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go app
# CGO_ENABLED=0 ensures a static binary for minimal alpine images
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server/main.go

# Production Stage
FROM alpine:latest

# Install CA certificates for Scryfall API calls
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .

# Copy web assets
COPY web/ ./web/

# Create persistent directories
RUN mkdir -p data uploads

# Environment variables
ENV PORT=8080
ENV DATABASE_DSN=/app/data/inventory.db

# Expose the application port
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
