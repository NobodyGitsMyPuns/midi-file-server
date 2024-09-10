# Build stage using a Go image
FROM golang:1.23 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN go build -o main .

# Run stage using a minimal Alpine image
FROM alpine:latest
RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /app

# Copy the pre-built binary from the builder stage
COPY --from=builder /app/main .

# Ensure the binary is executable
RUN chmod +x /app/main

# Expose port 8080
EXPOSE 8080

# Run the binary
CMD ["./main"]
