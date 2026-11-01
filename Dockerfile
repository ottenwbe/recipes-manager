# Get version from git tag or use default
ARG APP_VERSION="unknown"

# Build container for the application.
# Eases up the process of building for different target architectures.
FROM docker.io/golang:1.27.3-alpine AS builder

ARG APP=recipes-manager
ARG APP_VERSION

# Install build dependencies
RUN apk add --no-cache make git ca-certificates

COPY . /build
WORKDIR /build

# build the app binary
RUN make release APP_OUTPUT=${APP}

# Final production image
FROM docker.io/alpine:3.24.3

ARG APP=recipes-manager
ARG APP_VERSION
ENV GIN_MODE=release

# Metadata for production stage
LABEL maintainer="Beate Ottenwaelder <ottenwbe.public@gmail.com>" \
      description="Recipes Manager - Production container" \
      version="${APP_VERSION}" \
      stage="production"

# Create non-root user and group for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata dumb-init wget

# Create application directories
RUN mkdir -p /app && mkdir -p /etc/recipes-manager && \
    chown appuser:appgroup /app /etc/recipes-manager

# Copy the app binary to /app
COPY --from=builder --chown=appuser:appgroup /build/${APP} /app/recipes-manager

# Make port 8080 available to the world outside this container
EXPOSE 8080/tcp

# Security: Run as non-root user
USER appuser:appgroup

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

# Use dumb-init as PID 1 to handle signals properly
ENTRYPOINT ["dumb-init", "--"]
CMD ["/app/recipes-manager"]
