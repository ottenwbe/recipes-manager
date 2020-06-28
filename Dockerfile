FROM ubuntu:latest

ARG APP=go-cook
ENV GIN_MODE release

RUN mkdir -p /app; mkdir -p /etc/go-cook

# Copy the app binary to /app
COPY ${APP} /app/go-cook

# Make port 8080 available to the world outside this container
EXPOSE 8080

ENTRYPOINT ["/app/go-cook"]
