# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-s -w -X main.version=${VERSION:-dev}" \
    -o tempus .

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 tempus && \
    adduser -D -u 1000 -G tempus tempus

WORKDIR /home/tempus

# Copy binary from builder
COPY --from=builder /build/tempus /usr/local/bin/tempus

# Copy timezone data and other resources
COPY --from=builder /build/timezones /home/tempus/timezones
COPY --from=builder /build/locales /home/tempus/locales

# Set ownership
RUN chown -R tempus:tempus /home/tempus

# Switch to non-root user
USER tempus

# Set environment
ENV HOME=/home/tempus
ENV PATH=/usr/local/bin:$PATH

# Default command
ENTRYPOINT ["tempus"]
CMD ["--help"]

# Labels
LABEL org.opencontainers.image.title="Tempus"
LABEL org.opencontainers.image.description="ADHD-friendly ICS calendar event generator"
LABEL org.opencontainers.image.authors="Tempus Contributors"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.documentation="https://github.com/YOUR_USERNAME/tempus"
