# Build stage
FROM golang:1.23 AS builder

# Set the working directory
WORKDIR /app

# Copy the entire project to the build context
COPY . .

# Set up Go modules
WORKDIR /app/order
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o service main.go

# Final stage
#FROM debian:bullseye-slim
FROM gcr.io/distroless/static-debian11

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/order/service .

EXPOSE 8080

CMD ["./service"]