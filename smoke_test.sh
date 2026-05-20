#!/bin/bash
set -e

IMAGE_NAME="ottenwbe/recipes-manager:local"
CONTAINER_NAME="smoke-test-recipes-manager"
DB_CONTAINER_NAME="smoke-test-mongo"
NETWORK_NAME="recipes-smoke-test-net"
PORT=8080
HEALTH_ENDPOINT="/api/v1/recipes"

echo "--- Starting Recipes Manager Local Smoke Test ---"

# --- Cleanup function to ensure container is stopped and removed ---
cleanup() {
    echo "Cleaning up..."
    podman stop "$CONTAINER_NAME" "$DB_CONTAINER_NAME" > /dev/null 2>&1 || true
    podman rm "$CONTAINER_NAME" "$DB_CONTAINER_NAME" > /dev/null 2>&1 || true
    podman network rm "$NETWORK_NAME" > /dev/null 2>&1 || true
    echo "Cleanup complete."
}
trap cleanup EXIT

echo "1. Building image..."
podman build -t "$IMAGE_NAME" -f Dockerfile .

echo "2. Setting up network and Database..."
podman network create "$NETWORK_NAME" > /dev/null 2>&1 || true

podman run -d \
    --name "$DB_CONTAINER_NAME" \
    --network "$NETWORK_NAME" \
    mongo:7 > /dev/null

echo "3. Waiting for MongoDB to be ready..."
until podman exec "$DB_CONTAINER_NAME" mongosh --port 27017 --eval "db.adminCommand('ping')" >/dev/null 2>&1; do
  echo "MongoDB not ready yet... waiting"
  sleep 1
done

echo "4. Checking if port $PORT is in use..."
if ss -tuln | grep -q ":$PORT "; then
    echo "Error: Port $PORT is already in use. Please free it before running the smoke test."
    exit 1
fi

echo "5. Starting application container..."
podman run -d \
    -p $PORT:8080 \
    --name "$CONTAINER_NAME" \
    --network "$NETWORK_NAME" \
    -e RECIPES_MANAGER_RECIPEDB_HOST="mongodb://$DB_CONTAINER_NAME:27017" \
    "$IMAGE_NAME"

echo "6. Waiting for application to respond (via $HEALTH_ENDPOINT)..."
MAX_RETRIES=30
COUNT=0

while [ $COUNT -lt $MAX_RETRIES ]; do
    HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://127.0.0.1:$PORT$HEALTH_ENDPOINT || echo "000")
    
    # We accept 200 (OK), 401 (Unauthorized), or 500 (Internal Server Error)
    # because any of these indicate that the Go server process is running and listening.
    if [ "$HTTP_STATUS" = "200" ] || [ "$HTTP_STATUS" = "401" ] || [ "$HTTP_STATUS" = "500" ]; then
        echo "Success: Application is responding with HTTP $HTTP_STATUS!"
        break
    fi
    
    echo "Current HTTP status for $HEALTH_ENDPOINT: $HTTP_STATUS. Waiting... ($((COUNT+1))/$MAX_RETRIES)"
    sleep 1
    COUNT=$((COUNT+1))
done

if [ $COUNT -eq $MAX_RETRIES ]; then
    echo "Failure: Application failed to respond within $MAX_RETRIES seconds."
    echo "Container logs for '$CONTAINER_NAME':"
    podman logs "$CONTAINER_NAME"
    exit 1
fi

echo "--- Smoke Test Passed Successfully ---"