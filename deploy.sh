#!/bin/bash
set -e

# Определяем команду docker-compose
if command -v docker-compose &> /dev/null; then
    COMPOSE_CMD="docker-compose"
elif docker compose version &> /dev/null; then
    COMPOSE_CMD="docker compose"
else
    echo "Ошибка: docker-compose не найден!"
    exit 1
fi

echo "Используем: $COMPOSE_CMD"

HOST_IP="${1:-$(hostname -I | awk '{print $1}')}"
echo "IP: $HOST_IP"

# Создаём .env.prod
cat > .env.prod << EOF
HOST_IP=${HOST_IP}
POSTGRES_PASSWORD=videoconf_secret_pass_2024
JWT_ACCESS_SECRET=prod-access-$(openssl rand -hex 16)
JWT_REFRESH_SECRET=prod-refresh-$(openssl rand -hex 16)
LIVEKIT_API_KEY=prodkey
LIVEKIT_API_SECRET=prodsecret$(openssl rand -hex 8)
EOF

echo "Останавливаем старые контейнеры..."
$COMPOSE_CMD -f docker-compose.prod.yml --env-file .env.prod down 2>/dev/null || true

echo "Собираем и запускаем..."
$COMPOSE_CMD -f docker-compose.prod.yml --env-file .env.prod up -d --build

sleep 15

echo "Статус:"
$COMPOSE_CMD -f docker-compose.prod.yml --env-file .env.prod ps

echo ""
echo "Готово! Открой: http://${HOST_IP}:18080"
