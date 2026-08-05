#!/bin/sh
set -e

echo "Running database migrations..."
/bin/goose -dir ./migrations postgres "$DATABASE_URL" up

echo "Starting server..."
exec /bin/server

