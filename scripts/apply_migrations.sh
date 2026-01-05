#!/bin/bash
# Скрипт для применения всех миграций

set -e

DB_NAME="${POSTGRES_DB:-app_database}"
DB_USER="${POSTGRES_USER:-appuser}"
MIGRATIONS_DIR="migrations"

echo "Применение миграций базы данных..."

# Проверяем что директория существует
if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo "❌ Директория $MIGRATIONS_DIR не найдена"
    exit 1
fi

# Применяем миграции в порядке
for migration in $(ls -1 "$MIGRATIONS_DIR"/*.sql | sort); do
    echo "Применение: $migration"
    psql -U "$DB_USER" -d "$DB_NAME" -f "$migration"
done

echo "✅ Все миграции применены"

