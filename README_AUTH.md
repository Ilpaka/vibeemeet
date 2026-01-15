# 🎉 VibeMeet 2.0 - С централизованной авторизацией

> **Важно:** Это обновленная версия VibeMeet с интеграцией Auth-сервиса!

---

## 🚀 Быстрый старт

### 1️⃣ Запустите систему

```bash
# Используйте автоматический скрипт
cd /Users/ilpaka/Development
./QUICK_START.sh
```

**ИЛИ** запустите вручную:

```bash
# 1. Auth-сервис
cd /Users/ilpaka/Development/auth-service
docker-compose up -d
make migrate
make run

# 2. VibeMeet
cd /Users/ilpaka/Development/vibeemeet
docker-compose up -d
go run cmd/server/main.go
```

### 2️⃣ Откройте браузер

```
http://localhost
```

### 3️⃣ Зарегистрируйтесь

- Нажмите **"Регистрация"**
- Заполните форму
- Готово! Вы на Dashboard

### 4️⃣ Создайте комнату

- Нажмите **"Создать комнату"**
- Пригласите друзей по ссылке!

---

## 📋 Что нового?

### ✨ Новые возможности

- ✅ **Регистрация и вход** - безопасная авторизация через Auth-сервис
- ✅ **Dashboard** - личный кабинет для управления комнатами
- ✅ **Гостевой доступ** - друзья могут присоединяться без регистрации
- ✅ **JWT токены** - современная система безопасности
- ✅ **Обновленный UI** - красивые формы и анимации

### 🔄 Изменения

- **Создание комнат** - только для зарегистрированных пользователей
- **Присоединение к комнатам** - для всех по ссылке-приглашению
- **Главная страница** - кнопки "Войти" и "Регистрация"

---

## 📁 Структура проекта

```
vibeemeet/
├── web/
│   ├── index.html          # Главная (обновлена)
│   ├── Auth.html           # Вход/Регистрация (новая)
│   ├── Dashboard.html      # Личный кабинет (новая)
│   ├── room.html           # Комната
│   ├── js/
│   │   ├── MainMenu.js     # Обновлена
│   │   ├── Auth.js         # Новая
│   │   ├── Dashboard.js    # Новая
│   │   └── room.js
│   └── css/
│       └── style.css       # Добавлены стили
├── internal/
│   ├── middleware/
│   │   └── auth_service.go # Новая
│   └── service/
│       └── auth_service_client.go # Новая
├── docs/
│   └── USER_FLOW.md        # Новая
├── START_WITH_AUTH.md      # Инструкции (новая)
├── UPDATED_FLOW.md         # Описание изменений (новая)
├── CHANGES_SUMMARY.md      # Сводка (новая)
└── README_AUTH.md          # Этот файл
```

---

## 🎯 User Flow

### Для новых пользователей:

```
index.html → Auth.html (Регистрация) → Dashboard.html → room.html
```

### Для существующих пользователей:

```
index.html → Auth.html (Вход) → Dashboard.html → room.html
```

### Для гостей:

```
Ссылка-приглашение → Ввод имени → room.html
```

---

## 🧪 Тестовые аккаунты

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
┌─────────────────┐
│  VibeMeet UI    │ (localhost)
└────────┬────────┘
         │
         ├──────────────────┐
         │                  │
         ▼                  ▼
┌─────────────────┐  ┌─────────────────┐
│  Auth-сервис    │  │  VibeMeet API   │
│  (Port 8080)    │  │  (Port 8081)    │
└────────┬────────┘  └────────┬────────┘
         │                    │
         ▼                    ▼
┌─────────────────┐  ┌─────────────────┐
│  Auth DB        │  │  VibeMeet DB    │
│  (Port 5433)    │  │  (Port 5432)    │
└─────────────────┘  └─────────────────┘
```

---

## 🔐 Безопасность

### JWT Токены

| Тип | Срок жизни | Назначение |
|-----|-----------|-----------|
| Access Token | 15 минут | API запросы |
| Refresh Token | 7 дней | Обновление Access Token |
| Guest Token | 6 часов | Гостевой доступ |

### Пароли

- Хеширование: **bcrypt**
- Минимальная длина: **8 символов**
- Требования: буквы и цифры

---

## 📚 Документация

### Основные файлы:

- **[START_WITH_AUTH.md](START_WITH_AUTH.md)** 📖
  - Подробные инструкции по запуску
  - Конфигурация
  - Troubleshooting

- **[UPDATED_FLOW.md](UPDATED_FLOW.md)** 🔄
  - Краткое описание изменений
  - Сравнение старого и нового flow

- **[docs/USER_FLOW.md](docs/USER_FLOW.md)** 🎯
  - Подробный user flow с диаграммами
  - Технические детали
  - Сценарии использования

- **[CHANGES_SUMMARY.md](CHANGES_SUMMARY.md)** 📋
  - Полная сводка изменений
  - Список измененных файлов
  - Changelog

### Дополнительные:

- **[docs/AUTH_INTEGRATION.md](docs/AUTH_INTEGRATION.md)** - Интеграция с Auth-сервисом
- **[docs/API_DOCUMENTATION.md](docs/API_DOCUMENTATION.md)** - API документация

---

## 🛠️ Конфигурация

### VibeMeet `.env`

```env
# Auth Service
AUTH_SERVICE_URL=http://localhost:8080

# Database
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

---

## 🐛 Troubleshooting

### Не могу войти

1. Проверьте, что Auth-сервис запущен:
   ```bash
   curl http://localhost:8080/health
   ```

2. Проверьте консоль браузера (F12)

3. Попробуйте зарегистрироваться заново

### Не могу создать комнату

1. Убедитесь, что вы авторизованы
2. Проверьте, что VibeMeet API запущен:
   ```bash
   curl http://localhost:8081/health
   ```

3. Проверьте консоль браузера (F12)

### Подробнее

Смотрите раздел Troubleshooting в [START_WITH_AUTH.md](START_WITH_AUTH.md)

---

## 🎯 Следующие шаги

После успешного запуска:

1. ✅ Зарегистрируйтесь
2. ✅ Создайте комнату
3. ✅ Пригласите гостя
4. ✅ Протестируйте видео/аудио/чат

---

## 💡 Полезные команды

### Остановить все

```bash
cd /Users/ilpaka/Development
./STOP_ALL.sh
```

### Посмотреть логи

```bash
# Auth-сервис
docker logs auth-db

# VibeMeet
docker logs vibeemeet-postgres
docker logs vibeemeet-livekit
```

### Очистить данные

```bash
# Auth-сервис
cd /Users/ilpaka/Development/auth-service
docker-compose down -v

# VibeMeet
cd /Users/ilpaka/Development/vibeemeet
docker-compose down -v
```

---

## 🎉 Готово!

VibeMeet теперь работает с новой системой авторизации!

**Вопросы?** Смотрите документацию:
- [START_WITH_AUTH.md](START_WITH_AUTH.md) - Инструкции по запуску
- [docs/USER_FLOW.md](docs/USER_FLOW.md) - Подробный user flow

---

**Версия:** 2.0  
**Дата:** 2026-01-13  
**Автор:** AI Assistant (Claude Sonnet 4.5)
