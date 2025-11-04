
FROM golang:1.24-alpine AS builder

ARG SERVICE_PATH

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /app/server ./$SERVICE_PATH/main.go

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/server .

CMD ["./server"]