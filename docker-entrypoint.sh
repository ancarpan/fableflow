#!/bin/sh
set -e

# Set default values for environment variables
export FF_HOST="${FF_HOST:-0.0.0.0}"
export FF_PORT="${FF_PORT:-8080}"
export FF_SERVESTATIC="${FF_SERVESTATIC:-true}"
export FF_SCAN_DIR="${FF_SCAN_DIR:-/ebooks}"
export FF_IMPORT_DIR="${FF_IMPORT_DIR:-/import}"
export FF_QUARANTINE_DIR="${FF_QUARANTINE_DIR:-/quarantine}"
export FF_TMP_DIR="${FF_TMP_DIR:-/tmp}"
export FF_LOG_DIR="${FF_LOG_DIR:-/logs}"
export FF_DATABASE_PATH="${FF_DATABASE_PATH:-/database/ebooks.db}"

# Use envsubst to replace all ${VAR_NAME} placeholders
# in the template file with the actual environment variable values,
# and write the output to the final config file path.
envsubst < /app/config.yaml.template > /app/config.yaml

# Execute the main application command
exec "$@" -c /app/config.yaml
