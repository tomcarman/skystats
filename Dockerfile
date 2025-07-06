# Build the app
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . /app
RUN go build -o skystats ./core

# Run the app
FROM alpine:latest
WORKDIR /app/core
COPY --from=builder /app/skystats .
COPY --from=builder /app/web /app/web

CMD ["./skystats"]
