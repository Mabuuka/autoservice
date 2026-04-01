#!/usr/bin/env sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"

if [ -f "$ROOT_DIR/.env" ]; then
    set -a
    . "$ROOT_DIR/.env"
    set +a
fi

DB_USER="${POSTGRES_USER:-automaster_user}"
DB_NAME="${POSTGRES_DB:-automaster}"

cd "$ROOT_DIR"

echo "Seeding database '$DB_NAME' as user '$DB_USER'..."
docker compose exec -T db psql -v ON_ERROR_STOP=1 -U "$DB_USER" -d "$DB_NAME" < "$ROOT_DIR/automaster_postgresql_seed.sql"
echo "Seed completed successfully."
