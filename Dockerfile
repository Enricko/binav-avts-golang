# Start from the official Golang image
FROM golang:1.22-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache gcc musl-dev

# Set the working directory inside the container
WORKDIR /golang-app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# Start a new stage from scratch
FROM alpine:latest  

WORKDIR /golang-app

# Copy the pre-built binary file from the previous stage
COPY --from=builder /golang-app/main .

# Copy necessary directories and files
COPY --from=builder /golang-app/templates ./templates
COPY --from=builder /golang-app/public ./public
COPY --from=builder /golang-app/certs ./certs
COPY --from=builder /golang-app/.env .

# Ensure the public/uploads directory exists and has correct permissions
RUN mkdir -p /golang-app/public/uploads && chmod 755 /golang-app/public/uploads

# Set environment variable for upload directory
ENV UPLOAD_DIR=/golang-app/public/uploads

# Ensure correct permissions for the entire app directory
RUN chmod -R 755 /golang-app

# Create a startup script
RUN echo '#!/bin/sh' > start.sh && \
    echo 'while true; do' >> start.sh && \
    echo '    ./main' >> start.sh && \
    echo '    echo "Application crashed with exit code $?. Respawning.." >&2' >> start.sh && \
    echo '    sleep 1' >> start.sh && \
    echo 'done' >> start.sh && \
    chmod +x start.sh

# Expose the port the app runs on   
EXPOSE 443

# Command to run the startup script
CMD ["./start.sh"]