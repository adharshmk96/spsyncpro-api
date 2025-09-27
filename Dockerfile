# Start with the official Golang image as the builder
FROM golang:1.25-alpine AS builder

# Set the current working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -o main .

# Start a new stage from scratch
FROM alpine:latest

# Set the current working directory inside the container
WORKDIR /root/

# Copy the compiled executable from the builder stage
COPY --from=builder /app/main .

# Expose the port that the application listens on
EXPOSE 8080

# Command to run the executable
CMD ["./main", "serve", "--port", "8080"]
