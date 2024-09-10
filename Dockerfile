FROM golang:1.23 as builder

# Set the environment variables for cross-compilation
ENV GOOS=linux
ENV GOARCH=amd64

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code and k8s directory into the container
COPY . .
COPY .k8 ./.k8/

# Build the Go app and list contents of /app to verify
RUN go build -o main . && ls -la /app && chmod +x /app/main

# Use a minimal image for running the app
FROM alpine:latest
RUN apk --no-cache add ca-certificates

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the Pre-built binary file from the builder stage
COPY --from=builder /app/main .

# List the contents of the /root directory to verify
RUN ls -la /root && chmod +x /root/main

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the 
CMD ["/app/main"]

