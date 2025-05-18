FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./
# Optional go.sum if we have it
COPY go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /solidgo ./cmd/server

# Create a minimal image for running the application
FROM alpine:latest

# Add CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /solidgo .

# Expose the default port
EXPOSE 8080

# Command to run the executable
CMD ["./solidgo"] 