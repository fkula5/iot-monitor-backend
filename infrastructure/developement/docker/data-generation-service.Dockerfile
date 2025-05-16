FROM golang:1.24 AS builder

WORKDIR /app

# Copy go.mod and go.sum files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the data generation service
WORKDIR /app/services/data-generation-service
RUN CGO_ENABLED=0 GOOS=linux go build -o data-generator .

# Use a smaller base image for the final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/services/data-generation-service/data-generator .

# Set executable permissions
RUN chmod +x /app/data-generator

# Command to run the executable
CMD ["/app/data-generator"]