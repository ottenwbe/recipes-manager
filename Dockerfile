# Build container for the application.
# Eases up the process of building
# for different target architectures.
FROM docker.io/golang:1.25.8-alpine AS builder
ARG APP=recipes-manager

# Install build dependencies
RUN apk add --no-cache make git

COPY . /build
WORKDIR /build

# build the app binary
RUN make release

FROM docker.io/alpine:3.23.3

ARG APP=recipes-manager
ENV GIN_MODE release

RUN apk add --no-cache ca-certificates tzdata

RUN mkdir -p /app; mkdir -p /etc/recipes-manager

# Copy the app binary to /app
COPY --from=builder /build/${APP} /app/recipes-manager

# Make port 8080 available to the world outside this container
EXPOSE 8080

ENTRYPOINT ["/app/recipes-manager"]
