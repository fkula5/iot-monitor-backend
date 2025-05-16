FROM golang:1.24 AS builder

WORKDIR /app

# Copy go.mod and go.sum files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the sensor service
WORKDIR /app/services/sensor-service
RUN CGO_ENABLED=0 GOOS=linux go build -o sensor-service .

# Use a smaller base image for the final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/services/sensor-service/sensor-service .

# Set executable permissions
RUN chmod +x /app/sensor-service

# Command to run the executable
CMD ["/app/sensor-service"]