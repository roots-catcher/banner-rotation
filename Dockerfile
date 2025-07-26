# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o banner-rotation ./cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/banner-rotation .
COPY configs ./configs
EXPOSE 8080
ENV DB_URL=""
ENV KAFKA_BROKERS=""
ENV KAFKA_TOPIC=""
CMD ["./banner-rotation"]