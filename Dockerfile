# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM golang:1.27-bookworm AS build

# Install Node.js for the UI build and task for running the build pipeline.
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y nodejs && \
    npm install -g @go-task/task

WORKDIR /src

# Cache dependencies first.
COPY go.mod go.sum ./
RUN go mod download

COPY web/package.json web/package-lock.json* ./web/
RUN cd web && npm install

# Copy the rest of the source and build.
COPY . .
RUN task build

# ---- Runtime stage ----
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=build /src/demeter /app/demeter

ENTRYPOINT ["/app/demeter"]