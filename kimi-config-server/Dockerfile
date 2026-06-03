# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git (needed for go-git)
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates git

# Copy binary and config repo
COPY --from=builder /app/server .
COPY --from=builder /app/config-repo ./config-repo

# Render requires PORT env, default to 8080
ENV PORT=8080
ENV CONFIG_REPO_PATH=./config-repo

EXPOSE 8080

CMD ["./server"]