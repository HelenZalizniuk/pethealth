# Stage 1: Build the binary
FROM golang:1.25-alpine AS builder

# Install git and certificates (needed for Go modules and HTTPS)
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
# CGO_ENABLED=0 makes the binary static (perfect for Alpine/Distroless)
RUN CGO_ENABLED=0 GOOS=linux go build -o pethealth ./cmd/server/main.go

# Stage 2: Final lightweight image
FROM alpine:latest

# Add a non-root user for security (Senior practice!)
RUN adduser -D -u 1000 appuser
USER appuser

WORKDIR /home/appuser

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/pethealth .
# Copy config if it's a file, or ensure env vars are used
# COPY config.yaml . 

EXPOSE 8080

CMD ["./pethealth"]