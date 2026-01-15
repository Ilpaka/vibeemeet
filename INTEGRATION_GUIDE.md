# 🚀 VibeMeet + Auth Service - Integration Guide

## Архитектура

```
┌─────────────────────────────────────────────────────────┐
│  Browser (localhost:80)                                 │
│  ├── /auth-api/* ──► nginx ──► NextUp Auth Service     │
│  └── /api/v1/*   ──► nginx ──► VibeMeet Backend        │
└─────────────────────────────────────────────────────────┘
```

### Микросервисы

1. **NextUp Auth Service** (порт 8080)
   - Централизованная аутентификация
   - PostgreSQL БД для пользователей
   - JWT токены с `user_id`, `email`, `display_name`

2. **VibeMeet** (порт 80)
   - Видеоконференции через LiveKit
   - PostgreSQL БД для комнат
   - Auto-provisioning пользователей

## Запуск

### 1. Auth Service
```bash
cd ~/Development/NextUp-website
docker compose up --build
```

Дождитесь: `server started addr=:8080`

### 2. VibeMeet
```bash
cd ~/Development/vibeemeet
docker compose up --build
```

Дождитесь: `Starting server port=8080`

## Тестирование

1. **Открыть:** http://localhost
2. **Зарегистрироваться** (вкладка Register)
3. **Создать комнату** (кнопка "Создать комнату")
4. **Подключиться** - увидите себя в видео

## Ключевые изменения

### ✅ Auto-Provisioning
При первом запросе пользователь из Auth-сервиса автоматически создается в БД VibeMeet.

### ✅ JWT Integration
- **JWT Secret:** `change_me_in_production_super_secret_key_123` (синхронизирован)
- **Claims:** `user_id`, `email`, `display_name`
- **Валидация:** `ExternalAuthMiddleware` в VibeMeet

### ✅ API Protection
Все endpoints `/api/v1/rooms/*` требуют JWT токен в заголовке:
```http
Authorization: Bearer <access_token>
```

## Файлы изменений

### NextUp Auth Service
- `Dockerfile` - контейнеризация
- `docker-compose.yml` - postgres + auth-api
- `internal/auth/tokens.go` - добавлены email, display_name в JWT
- `internal/auth/service.go` - обновлена генерация токенов

### VibeMeet
- `internal/middleware/external_auth.go` - валидация JWT + auto-provisioning
- `cmd/server/main.go` - подключен ExternalAuthMiddleware
- `web/js/room.js` - добавлен Authorization header
- `web/js/Dashboard.js` - исправлено `data.room_id` → `data.id`
- `docker-compose.yml` - синхронизирован JWT секрет
- `nginx.conf` - проксирование `/auth-api/`

## Troubleshooting

### 502 Bad Gateway на /auth-api/*
**Причина:** Auth-сервис не запущен
**Решение:** Запустите NextUp auth service

### 401 Unauthorized при создании комнаты
**Причина:** JWT токен не передается или невалиден
**Решение:** Проверьте:
- JWT секреты совпадают
- `accessToken` есть в localStorage
- Токен не истек (15 минут)

### "failed to create room"
**Причина:** User не существует в БД VibeMeet
**Решение:** Auto-provisioning должен создать пользователя автоматически
- Проверьте логи VibeMeet: `docker logs video_conference_backend`

### Не вижу себя в комнате
**Причина:** Не получен LiveKit токен
**Решение:** Проверьте:
- Endpoint `/api/v1/rooms/{id}/media/token` доступен
- Authorization header передается в запросе

## Production Checklist

- [ ] Изменить JWT секрет на случайный 32+ символов
- [ ] Включить HTTPS (SSL сертификаты)
- [ ] Настроить CORS для production домена
- [ ] Включить `COOKIE_SECURE=true`
- [ ] Настроить firewall для портов
- [ ] Backup стратегия для обеих БД
- [ ] Мониторинг и логирование
- [ ] Rate limiting на Auth endpoints

## Environment Variables

### Auth Service (.env)
```env
JWT_SECRET=your_production_secret_here
POSTGRES_PASSWORD=strong_password_here
COOKIE_SECURE=true
CORS_ORIGINS=https://yourdomain.com
```

### VibeMeet (.env)
```env
JWT_ACCESS_SECRET=same_as_auth_service_jwt_secret
POSTGRES_PASSWORD=another_strong_password
LIVEKIT_API_KEY=your_livekit_key
LIVEKIT_API_SECRET=your_livekit_secret
```

## Architecture Benefits

✅ **Separation of Concerns** - каждый сервис имеет свою БД
✅ **Scalability** - сервисы масштабируются независимо
✅ **Security** - централизованная аутентификация
✅ **Flexibility** - легко добавить новые сервисы
✅ **Maintainability** - изменения в одном сервисе не ломают другие

---

**Статус:** ✅ Production Ready (после production checklist)
