FROM golang:1.22-alpine

WORKDIR /app

COPY go.mod . 
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o /my-gin-app

EXPOSE 8080

# Set environment variables for database connection
ENV DB_HOST=host.docker.internal
ENV DB_PORT=3306
ENV DB_NAME=golang_gin
ENV DB_USER=root
ENV DB_PASSWORD=
ENV DB_SSL_MODE=disable


CMD ["/my-gin-app"]
