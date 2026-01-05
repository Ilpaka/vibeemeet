#!/bin/sh

echo "Waiting for backend to be ready..."
MAX_RETRIES=60
RETRY_COUNT=0

while [ "$RETRY_COUNT" -lt "$MAX_RETRIES" ]; do
  if nc -z backend 8080 2>/dev/null; then
    echo "Backend is ready on port 8080"
    break
  fi
  RETRY_COUNT=$((RETRY_COUNT + 1))
  if [ $((RETRY_COUNT % 5)) -eq 0 ]; then
    echo "Waiting for backend... ($RETRY_COUNT/$MAX_RETRIES)"
  fi
  sleep 1
done

if [ "$RETRY_COUNT" -eq "$MAX_RETRIES" ]; then
  echo "Warning: Backend port check failed after $MAX_RETRIES attempts, but continuing..."
fi

echo "Starting nginx..."
exec nginx -g "daemon off;"