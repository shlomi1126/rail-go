# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download -x

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o rail-go ./cmd/bot

# Final stage
FROM alpine:latest

WORKDIR /app

# Install timezone data
RUN apk add --no-cache tzdata

# Set timezone to Jerusalem
ENV TZ=Asia/Jerusalem

# Copy the binary from builder
COPY --from=builder /app/rail-go .
# Copy the .env file
COPY .env .

# Create a non-root user
RUN adduser -D -g '' appuser
USER appuser

# Expose port (if needed)
EXPOSE 8080

# Run the application
CMD ["./rail-go"]
