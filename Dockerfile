# Distroless base image with libc and nonroot user (UID 65532, GID 65532)
FROM gcr.io/distroless/base-debian12:nonroot

# Build-time argument for versioning
ARG GIT_SHA=unknown

# Metadata labels
LABEL org.opencontainers.image.title="VibeFeeder" \
      org.opencontainers.image.description="RSS feed aggregator with AI-powered summaries" \
      org.opencontainers.image.source="https://github.com/tjanas94/vibefeeder" \
      org.opencontainers.image.revision="${GIT_SHA}" \
      org.opencontainers.image.vendor="Tomasz Janas" \
      org.opencontainers.image.licenses="MIT"

# Environment variables with production defaults
ENV SERVER_ADDRESS="0.0.0.0:8080" \
    LOG_LEVEL="error" \
    LOG_FORMAT="json" \
    TZ="UTC"

# Set working directory
WORKDIR /app

# Copy pre-built binary from host (assets are embedded in binary)
# The binary must be built on host using: task build
COPY --chown=65532:65532 dist/vibefeeder /app/vibefeeder

# Document the exposed port (informational)
EXPOSE 8080

# Run as nonroot user (already set by base image, explicit for clarity)
USER 65532:65532

# Start the application
CMD ["/app/vibefeeder"]
