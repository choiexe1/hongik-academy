#!/bin/bash
source ./deploy.env 2>/dev/null || true

DB_USER=${DB_USER:-academy}
DB_PASSWORD=${DB_PASSWORD:-academy123}
DB_NAME=${DB_NAME:-academy}
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}

GOOSE_DBSTRING="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

goose -dir ./migrations postgres "$GOOSE_DBSTRING" "$@"
