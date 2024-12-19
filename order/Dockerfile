# Build stage
FROM golang:1.23 as builder

# Set the working directory
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o service main.go

# Final stage
#FROM debian:bullseye-slim
FROM ubuntu:20.04
# Set the working directory
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/service .

# Expose the application port
EXPOSE 8080

# Command to run the application
CMD ["./service"]