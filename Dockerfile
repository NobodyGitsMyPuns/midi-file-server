FROM golang:1.23 as builder

# Set environment variables for cross-compilation
ENV GOOS=linux
ENV GOARCH=amd64

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app and verify the binary exists
RUN go build -o /app/main . && ls -la /app

# Use a minimal image for running the app
FROM alpine:latest
RUN apk --no-cache add ca-certificates

# Set the working directory to /app
WORKDIR /app

# Copy the built binary from the builder stage into /app
COPY --from=builder /app/main /app/main

# Ensure the binary is executable and verify it exists
RUN chmod +x /app/main && ls -la /app/main

COPY .env /app/.env
# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["/app/main"]
