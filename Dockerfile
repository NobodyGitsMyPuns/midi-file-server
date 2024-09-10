FROM golang:1.23 as builder

# Set environment variables for cross-compilation
ENV GOOS=linux
ENV GOARCH=amd64

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the entire source code (including main.go)
COPY . .

# Build the Go app, placing the binary in the /app directory
RUN go build -o main . && chmod +x /app/main && ls -la /app/main  # Add permission and debug step

# Use a minimal alpine image for the final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# Ensure the binary has execute permissions (just in case)
RUN chmod +x /app/main

# Expose the port your service is using
EXPOSE 8080

# Run the compiled Go binary
CMD ["/app/main"]
