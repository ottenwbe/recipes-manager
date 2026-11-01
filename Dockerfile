# Get version from git tag or use default
ARG APP_VERSION="unknown"

# Build container for the application.
# Eases up the process of building for different target architectures.
FROM docker.io/golang:1.26-alpine AS builder

ARG APP=recipes-manager
ARG APP_VERSION

# Metadata for builder stage
LABEL maintainer="Beate Ottenwaelder <ottenwbe.public@gmail.com>" \
      description="Recipes Manager - Go backend service" \
      version="${APP_VERSION}" \
      stage="builder"

# Install build dependencies
RUN apk add --no-cache make git

COPY . /build
WORKDIR /build

# build the app binary
RUN make release APP_OUTPUT=${APP}

# Final production image
FROM docker.io/alpine:3.24.0

ARG APP=recipes-manager
ARG APP_VERSION
ENV GIN_MODE release

# Metadata for production stage
LABEL maintainer="Beate Ottenwaelder <ottenwbe.public@gmail.com>" \
      description="Recipes Manager - Production container" \
      version="${APP_VERSION}" \
      stage="production"

RUN apk add --no-cache ca-certificates tzdata

RUN mkdir -p /app; mkdir -p /etc/recipes-manager

# Copy the app binary to /app
COPY --from=builder /build/${APP} /app/recipes-manager

# Make port 8080 available to the world outside this container
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

ENTRYPOINT ["/app/recipes-manager"]
