# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy all source files
COPY . .

# Build the application (dependencies will be downloaded during build)
RUN CGO_ENABLED=0 GOOS=linux go build -o nagger ./cmd/bot

# Final stage - use alpine for timezone and ca-certificates support
FROM alpine:3.19

# Add ca-certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/nagger .

# Run the application
CMD ["./nagger"]
