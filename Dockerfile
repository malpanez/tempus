# Build stage
# Digest resolved 2026-09-07 from the multi-arch manifest index (serves
# linux/amd64 and linux/arm64, both required by release.yml).
FROM golang:1.26-alpine@sha256:ce864e7223ac17b1775e6fd0b4c0db580c2eb50e7953a427916379e4b92a1628 AS builder

# Version injected at build time (release.yml passes --build-arg VERSION=vX.Y.Z)
ARG VERSION=dev

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o tempus .

# Final stage
# Digest resolved 2026-09-07 from the multi-arch manifest index (serves
# linux/amd64 and linux/arm64, both required by release.yml).
FROM alpine:3.21@sha256:48b0309ca019d89d40f670aa1bc06e426dc0931948452e8491e3d65087abc07d

# Install dependencies, create user, and set up directories in one layer
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 tempus && \
    adduser -D -u 1000 -G tempus tempus

WORKDIR /home/tempus

# Copy all files from builder in one layer
COPY --from=builder /build/tempus /usr/local/bin/tempus
COPY --from=builder /build/timezones /home/tempus/timezones
COPY --from=builder /build/locales /home/tempus/locales

# Set ownership after all copies
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
LABEL org.opencontainers.image.description="Neurodivergent-friendly ICS calendar generator for ADHD, Autism, Dyslexia"
LABEL org.opencontainers.image.authors="Tempus Contributors"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.documentation="https://github.com/malpanez/tempus"
