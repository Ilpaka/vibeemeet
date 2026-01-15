# 🚀 Запуск VibeMeet с новой системой авторизации

## ✅ Что было сделано

VibeMeet теперь интегрирован с централизованным Auth-сервисом:

- ✅ Создание комнат только для зарегистрированных пользователей
- ✅ Гостевой доступ по ссылке-приглашению (без регистрации)
- ✅ Страница авторизации (вход/регистрация)
- ✅ Dashboard для авторизованных пользователей
- ✅ JWT токены для безопасности
- ✅ Обновленный UI/UX

---

## 🎯 Быстрый запуск

### Шаг 1: Запустите Auth-сервис

```bash
cd /Users/ilpaka/Development/auth-service

# Запустить БД
docker-compose up -d

# Применить миграции (если еще не применены)
make migrate

# Запустить сервис
make run
```

**Проверка:** Auth-сервис доступен на http://localhost:8080

### Шаг 2: Запустите VibeMeet

```bash
cd /Users/ilpaka/Development/vibeemeet

# Запустить все сервисы (БД, LiveKit, Nginx, Backend)
docker-compose up -d

# ИЛИ запустить вручную:
# 1. БД и LiveKit
docker-compose up -d postgres livekit

# 2. Backend
go run cmd/server/main.go
```

**Проверка:**
- VibeMeet UI: http://localhost
- VibeMeet API: http://localhost:8081

### Шаг 3: Откройте браузер

```
http://localhost
```

---

## 🎨 Новый User Flow

### Для новых пользователей:

1. Откройте http://localhost
2. Нажмите **"Регистрация"**
3. Заполните форму:
   - Имя: `Ваше имя`
   - Email: `your@email.com`
   - Пароль: `минимум 8 символов`
4. Нажмите **"Зарегистрироваться"**
5. Вы попадете на Dashboard
6. Нажмите **"Создать комнату"**
7. Готово! 🎉

### Для существующих пользователей:

1. Откройте http://localhost
2. Нажмите **"Войти"**
3. Введите Email и Пароль
4. Нажмите **"Войти"**
5. Вы попадете на Dashboard
6. Нажмите **"Создать комнату"**
7. Готово! 🎉

### Для гостей (без регистрации):

1. Получите ссылку от хоста:
   ```
   http://localhost/room.html?room=<ID_КОМНАТЫ>
   ```
2. Откройте ссылку
3. Введите ваше имя
4. Нажмите **"Продолжить"**
5. Готово! Вы в комнате на 6 часов 🎉

---

## 🧪 Тестовые аккаунты

Для быстрого тестирования используйте:

```
Email: quickstart@test.com
Password: Quick123!@#
```

```
Email: user@test.com
Password: Test123!@#
```

```
Email: admin@test.com
Password: Admin123!@#
```

---

## 📊 Архитектура

```
┌──────────────────────────────────────────────────────┐
│                   VibeMeet Frontend                   │
│                   (localhost:80)                      │
│                                                       │
│  ┌──────────┐  ┌──────────┐  ┌──────────────┐      │
│  │ index.   │  │  Auth.   │  │  Dashboard.  │      │
│  │  html    │→ │  html    │→ │    html      │      │
│  └──────────┘  └──────────┘  └──────────────┘      │
└───────┬──────────────┬───────────────┬──────────────┘
        │              │               │
        │              ▼               │
        │     ┌─────────────────┐     │
        │     │  Auth-сервис    │     │
        │     │  (Port 8080)    │     │
        │     │                 │     │
        │     │  JWT Tokens     │     │
        │     │  User Sessions  │     │
        │     └────────┬────────┘     │
        │              │               │
        │              ▼               │
        │     ┌─────────────────┐     │
        │     │   Auth DB       │     │
        │     │  (Port 5433)    │     │
        │     │                 │     │
        │     │  - users        │     │
        │     │  - roles        │     │
        │     │  - permissions  │     │
        │     │  - guest_sess.  │     │
        │     └─────────────────┘     │
        │                             │
        ▼                             ▼
┌─────────────────┐         ┌─────────────────┐
│  VibeMeet API   │         │   LiveKit       │
│  (Port 8081)    │◄────────┤  (Port 7880)    │
│                 │         │                 │
│  - Rooms        │         │  - WebRTC       │
│  - Participants │         │  - Media        │
│  - Chat         │         │                 │
└────────┬────────┘         └─────────────────┘
         │
         ▼
┌─────────────────┐
│  VibeMeet DB    │
│  (Port 5432)    │
│                 │
│  - rooms        │
│  - participants │
│  - chat_msgs    │
└─────────────────┘
```

---

## 🔐 Безопасность

### JWT Токены

**Access Token:**
- Срок жизни: 15 минут
- Используется для всех API запросов
- Автоматически обновляется через Refresh Token

**Refresh Token:**
- Срок жизни: 7 дней
- Используется для получения нового Access Token
- Поддерживает rotation (старый токен инвалидируется)

**Guest Token:**
- Срок жизни: 6 часов
- Только для присоединения к комнатам
- Не дает права создавать комнаты

### Middleware

VibeMeet API проверяет JWT токены:

```go
// Только для авторизованных
POST /api/v1/rooms
Headers: {
  "Authorization": "Bearer <access_token>"
}

// Для всех (авторизованных и гостей)
GET /api/v1/rooms/:id
POST /api/v1/rooms/:id/join
```

---

## 📁 Новые файлы

### Frontend:

```
web/
├── Auth.html              # Страница входа/регистрации
├── Dashboard.html         # Личный кабинет
├── index.html            # Главная (обновлена)
├── js/
│   ├── Auth.js           # Логика авторизации
│   ├── Dashboard.js      # Логика Dashboard
│   └── MainMenu.js       # Обновлена
└── css/
    └── style.css         # Добавлены стили для Auth и Dashboard
```

### Backend:

```
internal/
├── middleware/
│   └── auth_service.go   # Middleware для проверки JWT
└── service/
    └── auth_service_client.go  # HTTP клиент для Auth-сервиса
```

### Документация:

```
docs/
└── USER_FLOW.md          # Подробный user flow

UPDATED_FLOW.md           # Краткое описание изменений
START_WITH_AUTH.md        # Этот файл
```

---

## 🛠️ Конфигурация

### VibeMeet `.env`

```env
# Auth Service
AUTH_SERVICE_URL=http://localhost:8080

# VibeMeet Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=vibeemeet_db
DB_USER=vibeemeet_user
DB_PASSWORD=your_password

# Server
VIBEEMEET_PORT=8081

# LiveKit
LIVEKIT_URL=ws://localhost:7880
LIVEKIT_API_KEY=your_api_key
LIVEKIT_API_SECRET=your_api_secret
```

### Auth-сервис `.env`

```env
# Database
DB_HOST=localhost
DB_PORT=5433
DB_NAME=auth_service
DB_USER=auth_user
DB_PASSWORD=your_password

# JWT
JWT_SECRET=your_jwt_secret
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d
JWT_GUEST_EXPIRY=6h

# Server
AUTH_SERVICE_PORT=8080

# CORS
ALLOWED_ORIGINS=http://localhost,http://localhost:8081
```

---

## 🐛 Troubleshooting

### Auth-сервис не запускается

```bash
# Проверьте, что БД запущена
docker ps | grep auth-db

# Проверьте логи
docker logs auth-db

# Перезапустите
cd /Users/ilpaka/Development/auth-service
docker-compose down
docker-compose up -d
make migrate
make run
```

### VibeMeet не запускается

```bash
# Проверьте, что все сервисы запущены
docker ps

# Проверьте логи
docker logs vibeemeet-postgres
docker logs vibeemeet-livekit
docker logs vibeemeet-nginx

# Перезапустите
cd /Users/ilpaka/Development/vibeemeet
docker-compose down
docker-compose up -d
```

### Не могу войти

1. Проверьте, что Auth-сервис запущен:
   ```bash
   curl http://localhost:8080/health
   ```

2. Проверьте консоль браузера (F12)

3. Попробуйте зарегистрироваться заново

### Не могу создать комнату

1. Убедитесь, что вы авторизованы (есть токен в localStorage)

2. Проверьте, что VibeMeet API запущен:
   ```bash
   curl http://localhost:8081/health
   ```

3. Проверьте консоль браузера (F12)

### CORS ошибки

Убедитесь, что в `.env` Auth-сервиса указаны правильные origins:

```env
ALLOWED_ORIGINS=http://localhost,http://localhost:8081
```

---

## 📚 Дополнительная документация

- **[USER_FLOW.md](docs/USER_FLOW.md)** - Подробный user flow с диаграммами
- **[AUTH_INTEGRATION.md](docs/AUTH_INTEGRATION.md)** - Интеграция с Auth-сервисом
- **[API_DOCUMENTATION.md](docs/API_DOCUMENTATION.md)** - API документация
- **[UPDATED_FLOW.md](UPDATED_FLOW.md)** - Краткое описание изменений

---

## 🎯 Следующие шаги

После успешного запуска:

1. ✅ Зарегистрируйтесь
2. ✅ Создайте комнату
3. ✅ Пригласите гостя (откройте ссылку в режиме инкогнито)
4. ✅ Протестируйте видео/аудио/чат
5. ✅ Проверьте выход из системы

---

## 💡 Полезные команды

### Остановить все сервисы

```bash
# Auth-сервис
cd /Users/ilpaka/Development/auth-service
docker-compose down

# VibeMeet
cd /Users/ilpaka/Development/vibeemeet
docker-compose down
```

### Очистить все данные

```bash
# Auth-сервис
cd /Users/ilpaka/Development/auth-service
docker-compose down -v

# VibeMeet
cd /Users/ilpaka/Development/vibeemeet
docker-compose down -v
```

### Посмотреть логи

```bash
# Auth-сервис
docker logs auth-db
docker logs -f auth-service  # если запущен в Docker

# VibeMeet
docker logs vibeemeet-postgres
docker logs vibeemeet-livekit
docker logs vibeemeet-nginx
```

---

## 🎉 Готово!

VibeMeet теперь работает с новой системой авторизации!

**Вопросы?** Смотрите документацию в `docs/USER_FLOW.md`

**Обновлено:** 2026-01-13
