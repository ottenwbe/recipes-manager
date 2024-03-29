# Build container for the application.
# Eases up the process of building
# for different target architectures.
FROM docker.io/golang:1.21 AS builder
ARG APP=recipes-manager
COPY . /build

WORKDIR /build

# build the app binary
RUN make release

FROM docker.io/ubuntu:latest

ARG APP=recipes-manager
ENV GIN_MODE release

RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /app; mkdir -p /etc/recipes-manager

# Copy the app binary to /app
COPY --from=builder /build/${APP} /app/recipes-manager

# Make port 8080 available to the world outside this container
EXPOSE 8080

ENTRYPOINT ["/app/recipes-manager"]
