# Build both backend and frontend
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Build backend
WORKDIR /app-backend
COPY backend/ ./
COPY backend/go.mod backend/go.sum ./
RUN CGO_CFLAGS="-D_LARGEFILE64_SOURCE" CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o fableflow-backend .

# Build frontend
WORKDIR /app-frontend
COPY frontend/ ./
RUN CGO_CFLAGS="-D_LARGEFILE64_SOURCE" CGO_ENABLED=1 GOOS=linux go build -o fableflow-frontend .

# Final stage
FROM alpine:3.18 AS distro

# Install runtime dependencies
RUN apk add --no-cache \
    python3 \
    sqlite \
    ca-certificates \
    gettext \
    tzdata

# Create app user
RUN adduser -D -s /bin/sh fableflow

COPY --from=builder /app-backend/fableflow-backend /app/
COPY --from=builder /app-frontend/fableflow-frontend /app/
COPY frontend/templates /web/templates
COPY frontend/static /web/static

COPY config.yaml.template /app/
COPY docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# Create necessary directories
RUN mkdir -p /app/data /app/logs /app/tmp && \
    chown -R fableflow:fableflow /app

# Create application directories
RUN mkdir -p /ebooks /import /quarantine /database && \
    chown -R fableflow:fableflow /ebooks /import /quarantine /database

    # Switch to non-root user
USER fableflow

# Expose port
EXPOSE 8080 3000

# Health check
# HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
#     CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Set the entrypoint to the custom script
ENTRYPOINT ["docker-entrypoint.sh"]

# Run the backend
CMD ["/app/fableflow-backend"]
