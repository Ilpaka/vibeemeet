#!/bin/bash

echo "🚀 Запуск Video Conference Application..."
echo ""

# Проверка Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker не установлен. Установите Docker и повторите попытку."
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose не установлен. Установите Docker Compose и повторите попытку."
    exit 1
fi

echo "✅ Docker найден"
echo ""

# Остановка существующих контейнеров
echo "🛑 Остановка существующих контейнеров..."
docker-compose down

echo ""
echo "🔨 Сборка и запуск контейнеров..."
docker-compose up --build -d

echo ""
echo "⏳ Ожидание запуска сервисов..."
sleep 10

# Проверка статуса
echo ""
echo "📊 Статус сервисов:"
docker-compose ps

echo ""
echo "✅ Приложение запущено!"
echo ""
echo "🌐 Откройте в браузере: http://localhost"
echo ""
echo "📝 Тестовые учетные данные:"
echo "   Email: ilya@example.com"
echo "   Пароль: password123"
echo ""
echo "📋 Просмотр логов: docker-compose logs -f"
echo "🛑 Остановка: docker-compose down"

