# 🔐 Интеграция VibeMeet с Auth-сервисом

## Дата: 2026-01-12

---

## 📋 Обзор изменений

VibeMeet интегрирован с централизованным Auth-сервисом для реализации SSO аутентификации.

### Ключевые изменения:

1. **Создание комнат** - теперь доступно только зарегистрированным пользователям
2. **Гостевой доступ** - гости могут присоединяться по invite_link без регистрации
3. **JWT токены** - от Auth-сервиса для всех операций
4. **Middleware** - для проверки токенов от Auth-сервиса

---

## 🔧 Новые компоненты

### 1. Middleware для Auth токенов

Файл: `internal/middleware/auth_service.go`

```go
// AuthServiceMiddleware проверяет JWT токены от Auth-сервиса
type AuthServiceMiddleware struct {
    authServiceURL string
    jwtSecret      string
}

// Authenticate - проверяет токен локально (JWT)
// VerifyWithAuthService - проверяет токен на Auth-сервисе (опционально)
```

### 2. Обновленная структура `rooms`

```sql
ALTER TABLE rooms
  ALTER COLUMN host_user_id TYPE TEXT,  -- External user ID from Auth-service
  ADD COLUMN host_email TEXT,
  ADD COLUMN host_display_name TEXT NOT NULL;

-- Убираем FK constraint на локальную таблицу users
ALTER TABLE rooms
  DROP CONSTRAINT IF EXISTS rooms_host_user_id_fkey;
```

### 3. Логика invite links

Файл: `internal/handler/invite_handler.go`

- **POST /api/rooms/{room_id}/invites** - создание invite link (только host)
- **POST /api/join/{invite_token}** - присоединение гостя по ссылке

---

## 📊 Сценарии использования

### Сценарий A: Зарегистрированный пользователь

1. Пользователь логинится через Auth-сервис
   ```bash
   POST http://localhost:8080/auth/login
   {
     "email": "user@test.com",
     "password": "Test123!@#"
   }
   ```

2. Получает `access_token` и `refresh_token`

3. Создает комнату в VibeMeet
   ```bash
   POST http://localhost:8081/api/rooms
   Authorization: Bearer <access_token>
   {
     "title": "My Room",
     "description": "Test room",
     "max_participants": 10
   }
   ```

4. Комната создается с `host_user_id` из токена

### Сценарий B: Гость по ссылке

1. Host создает invite link
   ```bash
   POST http://localhost:8081/api/rooms/{room_id}/invites
   Authorization: Bearer <access_token>
   ```

2. Гость переходит по ссылке и вводит display_name
   ```bash
   POST http://localhost:8081/api/join/{invite_token}
   {
     "display_name": "Guest User"
   }
   ```

3. VibeMeet обращается к Auth-сервису для создания guest session
   ```bash
   POST http://localhost:8080/auth/guest
   {
     "display_name": "Guest User",
     "room_id": "test-room-001"
   }
   ```

4. Гость получает `guest_token` (TTL: 6 часов)

5. Гость присоединяется к комнате как `anonymous_participant`

---

## 🔑 Переменные окружения для VibeMeet

Добавить в `.env`:

```bash
# Auth Service Integration
AUTH_SERVICE_URL=http://localhost:8080
JWT_SECRET=<same_secret_as_auth_service>

# Или для production
AUTH_SERVICE_URL=https://auth.yourdomain.com
JWT_SECRET=<strong_secret>
```

---

## 📝 API Endpoints изменения

### Создание комнаты (теперь требует авторизации)

**Было:**
```bash
POST /api/rooms
# Без авторизации
```

**Стало:**
```bash
POST /api/rooms
Authorization: Bearer <access_token>
```

### Новые endpoints:

#### POST /api/rooms/{room_id}/invites
Создание invite link для комнаты.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "invite_link": "http://localhost:8081/join/abc123...",
  "invite_token": "abc123...",
  "expires_at": "2026-01-13T00:00:00Z"
}
```

#### POST /api/join/{invite_token}
Присоединение гостя по ссылке.

**Body:**
```json
{
  "display_name": "Guest User"
}
```

**Response:**
```json
{
  "guest_token": "eyJhbGciOiJIUzI1NiIs...",
  "guest_id": "550e8400-e29b-41d4-a716-446655440001",
  "room_id": "test-room-001",
  "livekit_token": "...",
  "expires_at": "2026-01-12T18:00:00Z"
}
```

---

## 🛠️ Реализация

### Шаг 1: Обновить схему БД

```bash
cd /Users/ilpaka/Development/vibeemeet
psql -U vibeemeet_user -d vibeemeet_db -f migrations/006_integrate_auth_service.sql
```

### Шаг 2: Добавить middleware

Создать файл `internal/middleware/auth_service.go` (см. ниже)

### Шаг 3: Обновить handlers

- `internal/handler/room.go` - добавить проверку авторизации
- `internal/handler/invite.go` - новый handler для invite links
- `internal/handler/guest_join.go` - новый handler для гостевого присоединения

### Шаг 4: Обновить service layer

- `internal/service/auth_service_client.go` - клиент для обращения к Auth-сервису
- `internal/service/room_service.go` - обновить логику создания комнат

---

## 🧪 Тестирование

### 1. Регистрация и вход

```bash
# Регистрация
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test123!@#",
    "display_name": "Test User"
  }'

# Вход
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test123!@#"
  }'
```

### 2. Создание комнаты

```bash
ACCESS_TOKEN="<токен из предыдущего шага>"

curl -X POST http://localhost:8081/api/rooms \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My Room",
    "description": "Test room",
    "max_participants": 10
  }'
```

### 3. Создание invite link

```bash
ROOM_ID="<id комнаты>"

curl -X POST http://localhost:8081/api/rooms/$ROOM_ID/invites \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### 4. Присоединение гостя

```bash
INVITE_TOKEN="<token из предыдущего шага>"

curl -X POST http://localhost:8081/api/join/$INVITE_TOKEN \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "Guest User"
  }'
```

---

## ⚠️ Важные замечания

1. **JWT Secret** должен быть одинаковым в Auth-сервисе и VibeMeet
2. **Guest sessions** истекают через 6 часов
3. **Background job** очищает expired guests каждые 30 минут
4. **Anonymous rooms** теперь используются только для гостей
5. **Host** всегда должен быть зарегистрированным пользователем

---

## 🔄 Миграция существующих данных

Для существующих комнат нужно:

1. Либо удалить их
2. Либо назначить валидный `host_user_id` из Auth-сервиса

```sql
-- Пример удаления старых комнат
DELETE FROM rooms WHERE host_user_id NOT IN (
  SELECT id::text FROM users
);

-- Или мигрировать на нового пользователя
UPDATE rooms
SET host_user_id = '<valid_auth_service_user_id>',
    host_email = 'migrated@system.local',
    host_display_name = 'Migrated User'
WHERE host_user_id IS NULL OR host_user_id = '';
```

---

## 📚 Дополнительные ресурсы

- [Auth Service README](/Users/ilpaka/Development/auth-service/README.md)
- [Auth Service API](/Users/ilpaka/Development/auth-service/docs/DB_ANALYSIS.md)
- [VibeMeet Anonymous Architecture](ANONYMOUS_ARCHITECTURE.md)

---

**Дата обновления:** 2026-01-12  
**Версия:** 1.0
