FROM golang:1.23 as builder

# Set environment variables for cross-compilation
ENV GOOS=linux
ENV GOARCH=amd64
ARG COPY_ENV=false

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if go.mod and go.sum aren't changed
RUN go mod download

# Copy the source code and k8s directory into the container
COPY . .
COPY .k8 ./.k8/

# Build the Go app and verify the binary exists
RUN go build -o main . && ls -la

# Use a minimal image for running the app
FROM alpine:latest
RUN apk --no-cache add ca-certificates

# Set the Current Working Directory inside the container
WORKDIR /app/

# Copy the built binary from /app in the builder stage
COPY --from=builder /app/main /app/

# Ensure the binary is executable
RUN chmod +x /app/main

# Conditionally copy the .env file if COPY_ENV is true and .env exists
ARG COPY_ENV
RUN if [ "$COPY_ENV" = "true" ] && [ -f /app/.env ]; then \
    echo "Copying .env file"; \
    cp /app/.env .; \
else \
    echo "Skipping .env file copy"; \
fi

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable from /app
CMD ["/app/main"]
