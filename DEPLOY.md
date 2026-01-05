# Деплой Video Conference на сервер

## Быстрый старт (2 команды)

```bash
# 1. Клонируем проект
git clone <YOUR_REPO_URL> /opt/video_conference
cd /opt/video_conference

# 2. Создаём .env.prod (ЗАМЕНИТЕ IP на ваш!)
cat > .env.prod << 'EOF'
HOST_IP=91.210.106.151
LIVEKIT_FRONTEND_URL=ws://91.210.106.151:17880
POSTGRES_PASSWORD=videoconf_secret_pass_2024
JWT_ACCESS_SECRET=prod-access-secret-change-me-12345
JWT_REFRESH_SECRET=prod-refresh-secret-change-me-67890
LIVEKIT_API_KEY=prodkey
LIVEKIT_API_SECRET=prodsecret123
LIVEKIT_PORT=17880
EOF

# 3. Запускаем
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --build
```

## Используемые порты

| Порт | Протокол | Назначение |
|------|----------|------------|
| 18080 | TCP | Web интерфейс (HTTP) |
| 17880 | TCP | LiveKit WebSocket |
| 17881 | TCP | LiveKit TCP fallback |
| 17882 | UDP | LiveKit WebRTC |
| 50000-50050 | UDP | RTP Media |

## Откройте файрвол

```bash
# UFW (Ubuntu/Debian)
ufw allow 18080/tcp
ufw allow 17880/tcp
ufw allow 17881/tcp
ufw allow 17882/udp
ufw allow 50000:50050/udp

# или firewalld (CentOS/RHEL)
firewall-cmd --permanent --add-port=18080/tcp
firewall-cmd --permanent --add-port=17880/tcp
firewall-cmd --permanent --add-port=17881/tcp
firewall-cmd --permanent --add-port=17882/udp
firewall-cmd --permanent --add-port=50000-50050/udp
firewall-cmd --reload
```

## Проверка

Откройте в браузере: `http://YOUR_SERVER_IP:18080`

## Полезные команды

```bash
# Логи всех сервисов
docker compose -f docker-compose.prod.yml --env-file .env.prod logs -f

# Логи конкретного сервиса
docker compose -f docker-compose.prod.yml --env-file .env.prod logs -f backend
docker compose -f docker-compose.prod.yml --env-file .env.prod logs -f livekit

# Перезапуск
docker compose -f docker-compose.prod.yml --env-file .env.prod restart

# Остановка
docker compose -f docker-compose.prod.yml --env-file .env.prod down

# Пересборка
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --build

# Статус контейнеров
docker compose -f docker-compose.prod.yml --env-file .env.prod ps
```

## Troubleshooting

### Проверка health endpoints
```bash
curl http://localhost:18080/health
curl http://localhost:17880/
```

### Проверка портов
```bash
netstat -ntlp | grep -E "18080|17880|17881"
ss -tulpn | grep -E "18080|17880|17881"
```

### Просмотр логов при проблемах
```bash
# Все логи
docker compose -f docker-compose.prod.yml --env-file .env.prod logs

# Только ошибки backend
docker compose -f docker-compose.prod.yml --env-file .env.prod logs backend 2>&1 | grep -i error
```

### Полный сброс (если нужно начать заново)
```bash
docker compose -f docker-compose.prod.yml --env-file .env.prod down -v
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --build
```

## Обновление

```bash
cd /opt/video_conference
git pull
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --build
```

