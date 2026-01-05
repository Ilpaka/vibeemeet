# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies including CGO dependencies for codecs
RUN apk add --no-cache git gcc g++ pkgconfig libvpx-dev opus-dev

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled for codecs
# Use dynamic linking with proper library paths
ENV CGO_CFLAGS="-I/usr/include"
ENV CGO_LDFLAGS="-L/usr/lib"
RUN CGO_ENABLED=1 GOOS=linux go build -a -o server ./cmd/server

# Final stage
FROM alpine:latest

# Install runtime dependencies for codecs
RUN apk --no-cache add ca-certificates tzdata netcat-openbsd wget \
    libvpx libvpx-dev opus opus-dev

# Create symlinks for library version compatibility if needed
# The binary might be linked against libvpx.so.9 but Alpine has .so.11
RUN cd /usr/lib && \
    if ls libvpx.so.* 1> /dev/null 2>&1 && [ ! -f libvpx.so.9 ]; then \
        LATEST=$(ls -1 libvpx.so.* | head -1) && \
        ln -sf $(basename $LATEST) libvpx.so.9 && \
        ln -sf $(basename $LATEST) libvpx.so; \
    fi && \
    ldconfig

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/server .

# Copy entrypoint script and ensure Unix line endings
COPY docker-entrypoint.sh .
RUN sed -i 's/\r$//' docker-entrypoint.sh && \
    chmod +x docker-entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/app/docker-entrypoint.sh"]

