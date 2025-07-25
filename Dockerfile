# syntax=docker/dockerfile:1
FROM golang:1.24.2-alpine

# Set the working directory inside the container
WORKDIR /app

# Install git (required if your go.mod uses any private/public repos)
RUN apk update && apk add --no-cache git

# Copy go module files and download dependencies first (helps with caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Copy the .env file into the container
COPY .env .env

# Build the Go binary
RUN go build -o main .

# Run the application
CMD ["./main"]
