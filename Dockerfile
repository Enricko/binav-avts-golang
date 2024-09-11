# Start from the official Golang image
FROM golang:1.22-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache gcc musl-dev

# Set the working directory inside the container
WORKDIR /app

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

# Install necessary packages and set timezone
RUN apk --no-cache add ca-certificates mysql-client tzdata && \
    cp /usr/share/zoneinfo/Asia/Jakarta /etc/localtime && \
    echo "Asia/Jakarta" > /etc/timezone && \
    apk del tzdata

WORKDIR /app

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Explicitly copy SSL certificates
COPY certs/fullchain.pem /app/certs/fullchain.pem
COPY certs/privkey.pem /app/certs/privkey.pem

# Copy any other necessary files (like templates or static assets)
COPY templates templates/
COPY public public/

# Copy the .env file
COPY .env .

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

# Set the environment variable for the time zone
ENV TZ=Asia/Jakarta

# Command to run the startup script
CMD ["./start.sh"]