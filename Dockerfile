FROM golang:1.23 as builder

# Set environment variables for cross-compilation
ENV GOOS=linux
ENV GOARCH=amd64

# Set the working directory inside the container
WORKDIR /app

# Display the current directory
RUN pwd && ls -la

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./

# Display the current directory and contents after copying dependencies
RUN pwd && ls -la

# Download all dependencies
RUN go mod download

# Copy the entire source code (including main.go)
COPY . .

# Display the current directory and contents after copying the entire source
RUN pwd && ls -la

# Build the Go app, placing the binary in the /app directory
RUN go build -o main . && chmod +x /app/main && pwd && ls -la /app/main  # Add permission and debug step

# Display the final state of the directory in the builder stage
RUN pwd && ls -la /app

# Use a minimal alpine image for the final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /app

# Display the current directory in the final stage before copying
RUN pwd && ls -la

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# Display the contents after copying the binary
RUN pwd && ls -la

# Ensure the binary has execute permissions (just in case)
RUN chmod +x /app/main && pwd && ls -la /app/main

# Expose the port your service is using
EXPOSE 8080

# Display the final structure right before running the binary
RUN pwd && ls -la /app

# Run the compiled Go binary
CMD ["/app/main"]
