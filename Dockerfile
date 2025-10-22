# Build the app
FROM golang:1.25.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum /app
COPY core /app/core
COPY data /app/data
COPY docs /app/docs
RUN go mod download
ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown
RUN DATE="${DATE:-$(date -u +'%Y-%m-%dT%H:%M:%SZ')}" && \
    go build -ldflags "-s -w -X main.version=${VERSION:-dev} -X main.commit=${COMMIT:-none} -X main.date=${DATE}" -o skystats ./core


FROM node:20-alpine AS node

COPY ./web /app

SHELL ["sh", "-o", "pipefail", "-c", "-x"]
WORKDIR /app
RUN \
    npm install && \
    npm run build

LABEL org.opencontainers.image.source="https://github.com/tomcarman/skystats"
FROM ghcr.io/sdr-enthusiasts/docker-baseimage:base
#SHELL ["/bin/bash", "-o", "pipefail", "-c", "-x"]

ENV \
    S6_KILL_GRACETIME=100 \
    API_PORT=8080 \
    DOCKER_ENV=true

COPY --from=node /app/dist /app/dist
COPY --from=builder /app/skystats /app/core/skystats
COPY --from=builder /app/docs/logo /app/docs/logo
COPY migrations /app/migrations

COPY rootfs/ /
