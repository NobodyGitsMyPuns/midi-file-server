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

# Set the Current Working Directory inside the container
WORKDIR /rootapp

# Copy the source code and k8s directory into the container
COPY . .
COPY .k8 ./.k8/

# Use a minimal image for running the app
FROM alpine:latest
RUN apk --no-cache add ca-certificates

# Build the Go application
RUN go build -o main .

# Expose port 8080
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
