#!/bin/bash
# Скрипт для проверки схемы базы данных

set -e

DB_NAME="${POSTGRES_DB:-app_database}"
DB_USER="${POSTGRES_USER:-appuser}"

echo "Проверка схемы базы данных..."

# Список обязательных таблиц
REQUIRED_TABLES=(
    "users"
    "user_sessions"
    "rooms"
    "room_invites"
    "room_participants"
    "waiting_room_entries"
    "chat_messages"
    "participant_stats"
    "user_settings"
    "user_video_profiles"
    "audit_log"
)

MISSING_TABLES=()

for table in "${REQUIRED_TABLES[@]}"; do
    if ! psql -U "$DB_USER" -d "$DB_NAME" -tAc "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '$table');" | grep -q t; then
        MISSING_TABLES+=("$table")
        echo "❌ Таблица '$table' отсутствует"
    else
        echo "✅ Таблица '$table' существует"
    fi
done

if [ ${#MISSING_TABLES[@]} -eq 0 ]; then
    echo ""
    echo "✅ Все обязательные таблицы существуют!"
    exit 0
else
    echo ""
    echo "❌ Отсутствуют таблицы: ${MISSING_TABLES[*]}"
    echo "Запустите миграцию: migrations/002_create_missing_tables.sql"
    exit 1
fi

