# 🚀 Быстрый старт с Auth-сервисом

## Запуск полной системы с SSO аутентификацией

---

## Шаг 1: Запустить Auth-сервис

```bash
# 1. Перейти в директорию Auth-сервиса
cd /Users/ilpaka/Development/auth-service

# 2. Создать .env файл
cat > .env << 'EOF'
# Database
DB_HOST=auth-db
DB_PORT=5432
DB_NAME=auth_service
DB_USER=auth_user
DB_PASSWORD=auth_strong_password_2026
DB_SSLMODE=disable

# JWT
JWT_SECRET=super_secret_jwt_key_change_in_production_2026
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h
JWT_GUEST_EXPIRY=6h

# Server
AUTH_SERVICE_PORT=8080
AUTH_SERVICE_HOST=0.0.0.0
ENVIRONMENT=development
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8081

# Security
BCRYPT_COST=12
LOG_LEVEL=debug
LOG_FORMAT=json
EOF

# 3. Запустить Docker Compose (БД)
docker-compose up -d

# 4. Подождать готовности БД
sleep 10

# 5. Применить миграции
make migrate-up

# 6. Запустить Auth-сервис
go run cmd/server/main.go
```

**Auth-сервис запущен на:** `http://localhost:8080`

---

## Шаг 2: Настроить VibeMeet

```bash
# Открыть новый терминал
cd /Users/ilpaka/Development/vibeemeet

# Применить миграцию для интеграции с Auth
psql -h localhost -p 5432 -U vibeemeet_user -d vibeemeet_db -f migrations/006_integrate_auth_service.sql

# Обновить .env (добавить Auth-сервис настройки)
echo "" >> .env
echo "# Auth Service Integration" >> .env
echo "AUTH_SERVICE_URL=http://localhost:8080" >> .env
echo "JWT_SECRET=super_secret_jwt_key_change_in_production_2026" >> .env
```

---

## Шаг 3: Тестирование

### 3.1. Регистрация нового пользователя

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@vibeemeet.com",
    "password": "TestPass123!@#",
    "display_name": "Test User"
  }'
```

**Ожидаемый ответ:**
```json
{
  "id": "...",
  "email": "test@vibeemeet.com",
  "display_name": "Test User",
  "is_active": true,
  "roles": [{"code": "user", "name": "User"}],
  ...
}
```

### 3.2. Вход в систему

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@vibeemeet.com",
    "password": "TestPass123!@#"
  }' | jq .
```

**Ожидаемый ответ:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "550e8400-e29b-41d4-a716...",
  "user": {
    "id": "...",
    "email": "test@vibeemeet.com",
    "display_name": "Test User",
    "roles": ["user"],
    "permissions": ["room.create", "room.join", "invite.create"]
  }
}
```

**Сохраните `access_token` для дальнейших запросов!**

### 3.3. Создание комнаты (требует авторизации)

```bash
# Замените <ACCESS_TOKEN> на токен из предыдущего шага
ACCESS_TOKEN="eyJhbGciOiJIUzI1NiIs..."

curl -X POST http://localhost:8081/api/rooms \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My First Auth Room",
    "description": "Room created with Auth-service authentication",
    "max_participants": 10,
    "waiting_room_enabled": false
  }' | jq .
```

**Ожидаемый ответ:**
```json
{
  "id": "...",
  "livekit_room_name": "...",
  "title": "My First Auth Room",
  "host_user_id": "...",    // External user ID from Auth-service
  "host_email": "test@vibeemeet.com",
  "host_display_name": "Test User",
  "status": "scheduled",
  ...
}
```

### 3.4. Создание invite link для гостей

```bash
ROOM_ID="<ID комнаты из предыдущего шага>"

curl -X POST "http://localhost:8081/api/rooms/$ROOM_ID/invites" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Ожидаемый ответ:**
```json
{
  "invite_link": "http://localhost:8081/join/abc123def456...",
  "invite_token": "abc123def456...",
  "expires_at": "2026-01-13T00:00:00Z"
}
```

### 3.5. Присоединение гостя по ссылке (без регистрации)

```bash
INVITE_TOKEN="<token из предыдущего шага>"

curl -X POST "http://localhost:8081/api/join/$INVITE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "Guest User"
  }' | jq .
```

**Ожидаемый ответ:**
```json
{
  "guest_id": "...",
  "guest_token": "eyJhbGciOiJIUzI1NiIs...",  // JWT токен для гостя
  "room_id": "...",
  "livekit_token": "...",  // Токен для подключения к LiveKit
  "expires_at": "2026-01-12T18:00:00Z"  // +6 часов
}
```

### 3.6. Использование тестовых пользователей

Auth-сервис создает тестовых пользователей при инициализации:

#### Обычный пользователь:
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@test.com",
    "password": "Test123!@#"
  }' | jq .
```

#### Администратор:
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@test.com",
    "password": "Admin123!@#"
  }' | jq .
```

---

## Шаг 4: Проверка интеграции

### Health checks

```bash
# Auth-сервис
curl http://localhost:8080/health
curl http://localhost:8080/ready

# VibeMeet
curl http://localhost:8081/api/health
```

### Верификация токена

```bash
curl -X POST http://localhost:8080/auth/verify \
  -H "Content-Type: application/json" \
  -d "{\"token\": \"$ACCESS_TOKEN\"}" | jq .
```

**Ожидаемый ответ:**
```json
{
  "valid": true,
  "user_id": "...",
  "email": "test@vibeemeet.com",
  "display_name": "Test User",
  "roles": ["user"],
  "permissions": ["room.create", "room.join", "invite.create"],
  "is_guest": false,
  "expires_at": 1705075200
}
```

---

## 📊 Архитектура взаимодействия

```
┌─────────────────────────────────────┐
│   Client (Browser / Mobile App)    │
└──────────────┬──────────────────────┘
               │
               │ 1. Login
               ▼
┌─────────────────────────────────────┐
│   Auth Service (localhost:8080)     │
│   - JWT tokens                      │
│   - Refresh tokens                  │
│   - Guest sessions                  │
└──────────────┬──────────────────────┘
               │
               │ 2. access_token
               ▼
┌─────────────────────────────────────┐
│   VibeMeet (localhost:8081)         │
│   - Create room (with token)        │
│   - Generate invite link            │
└──────────────┬──────────────────────┘
               │
               │ 3. Guest join
               ▼
┌─────────────────────────────────────┐
│   Auth Service                      │
│   - Create guest_token (6h TTL)     │
└──────────────┬──────────────────────┘
               │
               │ 4. guest_token + livekit_token
               ▼
┌─────────────────────────────────────┐
│   LiveKit                           │
│   - Video/Audio streaming           │
└─────────────────────────────────────┘
```

---

## 🔧 Troubleshooting

### Проблема: "Database is not available"

```bash
# Проверить статус Docker контейнеров
docker-compose ps

# Проверить логи БД
docker-compose logs auth-db

# Пересоздать БД
docker-compose down -v
docker-compose up -d
sleep 10
make migrate-up
```

### Проблема: "Invalid token"

Убедитесь что `JWT_SECRET` одинаковый в:
- `/Users/ilpaka/Development/auth-service/.env`
- `/Users/ilpaka/Development/vibeemeet/.env`

### Проблема: "Failed to create room"

Проверьте что:
1. Auth-сервис запущен и доступен
2. Токен валидный и не истек
3. Пользователь имеет право `room.create`

```bash
# Проверить токен
curl -X POST http://localhost:8080/auth/verify \
  -H "Content-Type: application/json" \
  -d "{\"token\": \"$ACCESS_TOKEN\"}" | jq .valid
```

---

## 📚 Дополнительная информация

- [Auth Service README](/Users/ilpaka/Development/auth-service/README.md)
- [VibeMeet Auth Integration](/Users/ilpaka/Development/vibeemeet/docs/AUTH_INTEGRATION.md)
- [Database Analysis](/Users/ilpaka/Development/auth-service/docs/DB_ANALYSIS.md)

---

## ✅ Критерии успеха

После выполнения всех шагов вы должны иметь:

- ✅ Auth-сервис работает на `localhost:8080`
- ✅ VibeMeet интегрирован с Auth-сервисом
- ✅ Можно зарегистрировать нового пользователя
- ✅ Можно войти и получить токены
- ✅ Можно создать комнату с авторизацией
- ✅ Можно создать invite link
- ✅ Гость может присоединиться без регистрации
- ✅ Guest token истекает через 6 часов

---

**Дата:** 2026-01-12  
**Версия:** 1.0
