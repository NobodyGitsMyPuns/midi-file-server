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

# Build the Go app and verify the binary is created
RUN go build -o main . && ls -la main

# Run stage using a minimal Alpine image
FROM alpine:latest
RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /app

# Copy the pre-built binary from the builder stage and list contents
COPY --from=builder /app/main .

# Verify the binary is in the right place with correct permissions
RUN ls -la /app && chmod +x /app/main

# Expose port 8080
EXPOSE 8080

# Run the binary
CMD ["./main"]
