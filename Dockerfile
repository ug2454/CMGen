# Build stage for the Go backend
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install FFmpeg and build dependencies
RUN apk add --no-cache ffmpeg build-base

# Copy go mod and sum files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o cmgen ./cmd/cmgen

# Build stage for the React frontend
FROM node:18-alpine AS frontend-builder

WORKDIR /app/web

# Copy package files first for better layer caching
COPY web/package*.json ./

# Install dependencies
RUN npm ci --only=production

# Copy frontend source code
COPY web/ .

# Build the frontend with production settings
RUN npm run build

# Final stage - small image with just what we need
FROM alpine:3.18

WORKDIR /app

# Install runtime dependencies only
RUN apk add --no-cache ffmpeg ca-certificates tzdata

# Create a non-root user for security
RUN adduser -D -h /app cmgen
USER cmgen

# Copy the Go binary from builder
COPY --from=builder --chown=cmgen:cmgen /app/cmgen .

# Copy the built frontend
COPY --from=frontend-builder --chown=cmgen:cmgen /app/web/build ./web/build

# Create a volume for input videos
VOLUME ["/videos"]

# Expose port for the web UI
EXPOSE 8080

# Command to run the application in web mode
ENTRYPOINT ["./cmgen"]
CMD ["--web"] 