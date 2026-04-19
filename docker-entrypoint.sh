#!/bin/sh
set -e

# Set default values for environment variables
export FF_HOST="${FF_HOST:-0.0.0.0}"
export FF_PORT="${FF_PORT:-8080}"
export FF_BACKEND_ADDR="${FF_BACKEND_ADDR:-http://backend:8080}"
export FF_SCAN_DIR="${FF_SCAN_DIR:-/ebooks}"
export FF_IMPORT_DIR="${FF_IMPORT_DIR:-/import}"
export FF_QUARANTINE_DIR="${FF_QUARANTINE_DIR:-/quarantine}"
export FF_TMP_DIR="${FF_TMP_DIR:-/tmp}"
export FF_LOG_DIR="${FF_LOG_DIR:-/logs}"
export FF_DATABASE_PATH="${FF_DATABASE_PATH:-/database/ebooks.db}"

# Determine container mode based on environment variable or command
CONTAINER_MODE="${CONTAINER_MODE:-backend}"
CONFIG_FILE="${FF_CONFIG_FILE:-/app/config.yaml}"

if [ "$CONTAINER_MODE" = "frontend" ]; then
    echo "Starting FableFlow frontend (Caddy proxy)..."

    envsubst < /app/Caddyfile.template > /app/Caddyfile
    exec /app/caddy run --config /app/Caddyfile --adapter caddyfile
else
    echo "Starting FableFlow backend..."

    if [ -f "$CONFIG_FILE" ] && [ -n "$FF_CONFIG_FILE" ]; then
        echo "Using custom config: $CONFIG_FILE"
    else
        echo "Generating config from template..."
        envsubst < /app/config.yaml.template > /app/config.yaml
        CONFIG_FILE="/app/config.yaml"
    fi

    exec /app/fableflow-backend -c "$CONFIG_FILE"
fi
