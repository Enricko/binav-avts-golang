FROM golang:1.22-alpine

# Install tzdata to handle time zones
RUN apk add --no-cache tzdata

# Set the time zone to Asia/Jakarta (UTC+7)
ENV TZ=Asia/Jakarta

# Set the working directory
WORKDIR /app

# Copy the Go application source code
COPY . .

# Copy SSL certificates
COPY certs/fullchain.pem /etc/letsencrypt/live/binav-avts.id/fullchain.pem
COPY certs/privkey.pem /etc/letsencrypt/live/binav-avts.id/privkey.pem

# Install dependencies and build the application
RUN go mod download
RUN go build -o main .

# Expose ports for HTTP and HTTPS
EXPOSE 443

# Run the Go application
CMD ["./main"]
