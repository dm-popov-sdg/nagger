# Build stage
FROM golang:1.23-alpine AS builder

# Install ca-certificates for HTTPS
RUN apk update && apk add --no-cache ca-certificates && update-ca-certificates

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o nagger ./cmd/bot

# Final stage
FROM scratch

# Copy CA certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/nagger .

# Run the application
CMD ["./nagger"]
