# Build stage for the Go backend
FROM golang:1.20-alpine AS builder

WORKDIR /app

# Install FFmpeg and build dependencies
RUN apk add --no-cache ffmpeg build-base

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o cmgen ./cmd/cmgen

# Build stage for the React frontend
FROM node:18-alpine AS frontend-builder

WORKDIR /app/web

# Copy package files
COPY web/package*.json ./

# Install dependencies
RUN npm install

# Copy frontend source code
COPY web/ .

# Build the frontend
RUN npm run build

# Final stage
FROM alpine:latest

WORKDIR /app

# Install FFmpeg
RUN apk add --no-cache ffmpeg

# Copy the Go binary from builder
COPY --from=builder /app/cmgen .

# Copy the built frontend
COPY --from=frontend-builder /app/web/build ./web/build

# Expose port for the web UI
EXPOSE 3000

# Command to run the application
CMD ["./cmgen", "--web"] 