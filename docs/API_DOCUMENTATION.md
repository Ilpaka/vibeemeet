# Документация проекта Video Conference

## Содержание
1. [Обзор проекта](#обзор-проекта)
2. [Структура проекта](#структура-проекта)
3. [Точка входа](#точка-входа)
4. [Конфигурация](#конфигурация)
5. [Доменные модели](#доменные-модели)
6. [HTTP Handlers](#http-handlers)
7. [Сервисы](#сервисы)
8. [Репозитории](#репозитории)
9. [Middleware](#middleware)
10. [Утилиты](#утилиты)

---

## Обзор проекта

Проект представляет собой систему видеоконференций на Go с использованием LiveKit для WebRTC медиа-потоков. Архитектура следует принципам Clean Architecture с разделением на слои: handlers, services, repositories.

---

## Структура проекта

```
video_conference/
├── cmd/server/          # Точка входа приложения
├── internal/
│   ├── config/         # Конфигурация
│   ├── domain/         # Доменные модели
│   ├── handler/        # HTTP handlers
│   ├── middleware/     # Middleware
│   ├── repository/     # Репозитории для БД
│   └── service/        # Бизнес-логика
├── pkg/
│   ├── jwt/            # JWT утилиты
│   ├── logger/         # Логирование
│   └── errors/         # Обработка ошибок
└── web/                # Фронтенд
```

---

## Точка входа

### `cmd/server/main.go`

**Назначение:** Главный файл приложения, точка входа для запуска сервера.

**Функции:**

- **`main()`**
  - Загружает конфигурацию через `config.Load()`
  - Инициализирует логгер
  - Подключается к PostgreSQL и Redis
  - Инициализирует репозитории, сервисы и handlers
  - Настраивает роутер через `setupRouter()`
  - Запускает HTTP сервер с graceful shutdown

- **`setupRouter(handlers, authMiddleware, rateLimitMiddleware, cfg)`**
  - Настраивает Gin роутер
  - Подключает middleware (CORS, RequestLogger, ErrorHandler)
  - Регистрирует все API endpoints:
    - Health check: `GET /health`
    - Публичные: `/api/v1/auth/*`
    - Защищенные: `/api/v1/users/*`, `/api/v1/rooms/*`, `/api/v1/rooms/:id/chat/*`, `/api/v1/rooms/:id/media/*`, `/api/v1/rooms/:id/stats/*`
    - WebSocket: `GET /ws/chat/:id`

---

## Конфигурация

### `internal/config/config.go`

**Назначение:** Управление конфигурацией приложения из переменных окружения.

**Структуры:**

- **`Config`** - основная структура конфигурации
  - `Environment` - окружение (development/production)
  - `Server` - настройки сервера
  - `Database` - настройки БД
  - `Redis` - настройки Redis
  - `JWT` - настройки JWT токенов
  - `LiveKit` - настройки LiveKit
  - `Log` - настройки логирования

**Функции:**

- **`Load()`** - загружает конфигурацию из переменных окружения с дефолтными значениями
- **`validate()`** - валидирует обязательные поля конфигурации
- **`getEnv(key, defaultValue)`** - получает значение переменной окружения или возвращает дефолт
- **`getEnvAsInt(key, defaultValue)`** - получает int значение из переменной окружения
- **`getEnvAsDuration(key, defaultValue)`** - получает Duration значение из переменной окружения

---

## Доменные модели

### `internal/domain/user.go`

**Назначение:** Доменные модели для пользователей.

**Структуры:**

- **`User`** - пользователь системы
  - Поля: ID, Email, PasswordHash, DisplayName, AvatarURL, GlobalRole, IsActive, IsEmailVerified, LastLoginAt, CreatedAt, UpdatedAt

- **`UserSession`** - сессия пользователя
  - Поля: ID, UserID, RefreshTokenHash, CreatedAt, ExpiresAt, RevokedAt, RevokedReason, IPAddress, UserAgent

- **`UserSettings`** - настройки пользователя
  - Поля: UserID, DefaultCameraDeviceID, DefaultMicrophoneDeviceID, DefaultSpeakerDeviceID, PreferredVideoQuality, PreferredTheme, MuteMicOnJoin, DisableCameraOnJoin, LanguageCode, CreatedAt, UpdatedAt

- **`UserVideoProfile`** - профиль видео пользователя
  - Поля: ID, UserID, Name, BackgroundType, BackgroundImageURL, NoiseSuppressionLevel, IsDefault, CreatedAt

**Константы:**
- Роли: `GlobalRoleUser`, `GlobalRoleTechnicalAdmin`
- Качество видео: `VideoQuality1080p`, `VideoQuality720p`, `VideoQuality480p`, `VideoQuality360p`, `VideoQualityAuto`
- Темы: `ThemeLight`, `ThemeDark`, `ThemeSystem`
- Типы фона: `BackgroundTypeNone`, `BackgroundTypeBlur`, `BackgroundTypeImage`
- Уровни подавления шума: `NoiseSuppressionOff`, `NoiseSuppressionLow`, `NoiseSuppressionMedium`, `NoiseSuppressionHigh`

### `internal/domain/room.go`

**Назначение:** Доменные модели для комнат видеоконференций.

**Структуры:**

- **`Room`** - комната видеоконференции
  - Поля: ID, LiveKitRoomName, HostUserID, Title, Description, Status, ScheduledStartAt, ScheduledEndAt, ActualStartAt, ActualEndAt, MaxParticipants, WaitingRoomEnabled, IsLocked, PasswordHash, Settings, CreatedAt, UpdatedAt

- **`RoomInvite`** - приглашение в комнату
  - Поля: ID, RoomID, CreatedByUserID, LinkToken, Label, ExpiresAt, MaxUses, UsedCount, CreatedAt

- **`RoomParticipant`** - участник комнаты
  - Поля: ID, RoomID, UserID, Role, DisplayName, LiveKitSID, JoinedAt, LeftAt, LeaveReason, IsKicked, InitialMuted, ClientIP, UserAgent

- **`WaitingRoomEntry`** - запись в комнате ожидания
  - Поля: ID, RoomID, UserID, DisplayName, Status, RequestedAt, DecidedAt, DecidedByUserID, Reason

**Константы:**
- Статусы комнаты: `RoomStatusScheduled`, `RoomStatusActive`, `RoomStatusEnded`, `RoomStatusCancelled`
- Роли участников: `ParticipantRoleHost`, `ParticipantRoleCoHost`, `ParticipantRoleParticipant`
- Статусы waiting room: `WaitingRoomStatusPending`, `WaitingRoomStatusApproved`, `WaitingRoomStatusRejected`, `WaitingRoomStatusExpired`

### `internal/domain/chat.go`

**Назначение:** Доменные модели для чата.

**Структуры:**

- **`ChatMessage`** - сообщение в чате
  - Поля: ID, RoomID, SenderParticipantID, MessageType, Content, CreatedAt, EditedAt, DeletedAt, DeletedByParticipantID

**Константы:**
- Типы сообщений: `MessageTypeUser`, `MessageTypeSystem`

### `internal/domain/stats.go`

**Назначение:** Доменные модели для статистики.

**Структуры:**

- **`ParticipantStats`** - статистика участника
  - Поля: ID, RoomParticipantID, AvgRTTMs, MaxRTTMs, AvgJitterMs, PacketLossUpPercent, PacketLossDownPercent, AvgBitrateKbps, NetworkScore, CreatedAt

### `internal/domain/audit.go`

**Назначение:** Доменные модели для аудита.

**Структуры:**

- **`AuditLog`** - запись аудита
  - Поля: ID, EventTime, ActorUserID, ActorRole, RoomID, EventType, Payload

**Константы:**
- Роли акторов: `ActorRoleUser`, `ActorRoleHost`, `ActorRoleTechnicalAdmin`, `ActorRoleSystem`
- Типы событий: `EventTypeRoomCreated`, `EventTypeRoomUpdated`, `EventTypeRoomDeleted`, `EventTypeRoomJoined`, `EventTypeRoomLeft`, `EventTypeUserKicked`, `EventTypeRoomLocked`, `EventTypeRoomUnlocked`, `EventTypeWaitingRoomApproved`, `EventTypeWaitingRoomRejected`

### `internal/domain/rate_limit.go`

**Назначение:** Доменные модели для rate limiting.

**Структуры:**

- **`RateLimitRule`** - правило ограничения скорости
  - Поля: ID, Scope, Key, LimitPerMinute, LimitPerHour, LimitPerDay, Enabled, Description, CreatedAt, UpdatedAt

**Константы:**
- Области применения: `RateLimitScopeGlobal`, `RateLimitScopeUser`, `RateLimitScopeIP`, `RateLimitScopeRoom`

---

## HTTP Handlers

### `internal/handler/handlers.go`

**Назначение:** Контейнер для всех handlers.

**Структуры:**

- **`Handlers`** - содержит все handlers приложения
  - Поля: Health, Auth, User, Room, WaitingRoom, Chat, Media, Stats, WebSocket

**Функции:**

- **`NewHandlers(services, log)`** - создает новый экземпляр Handlers со всеми зависимостями

### `internal/handler/health.go`

**Назначение:** Health check endpoint.

**Структуры:**

- **`HealthHandler`** - handler для health check

**Функции:**

- **`NewHealthHandler()`** - создает новый HealthHandler
- **`Check(c)`** - возвращает статус сервиса (GET /health)

### `internal/handler/auth.go`

**Назначение:** Обработка аутентификации и авторизации.

**Структуры:**

- **`AuthHandler`** - handler для аутентификации
  - Поля: authService, log

- **`RegisterRequest`** - запрос на регистрацию
  - Поля: Email, Password, DisplayName

- **`LoginRequest`** - запрос на вход
  - Поля: Email, Password

- **`RefreshTokenRequest`** - запрос на обновление токена
  - Поля: RefreshToken

**Функции:**

- **`NewAuthHandler(authService, log)`** - создает новый AuthHandler
- **`Register(c)`** - регистрация нового пользователя (POST /api/v1/auth/register)
- **`Login(c)`** - вход пользователя (POST /api/v1/auth/login)
- **`RefreshToken(c)`** - обновление токена доступа (POST /api/v1/auth/refresh)

### `internal/handler/user.go`

**Назначение:** Обработка запросов пользователей.

**Структуры:**

- **`UserHandler`** - handler для пользователей
  - Поля: userService, log

- **`UpdateMeRequest`** - запрос на обновление профиля
  - Поля: DisplayName, AvatarURL

**Функции:**

- **`NewUserHandler(userService, log)`** - создает новый UserHandler
- **`GetMe(c)`** - получение информации о текущем пользователе (GET /api/v1/users/me)
- **`UpdateMe(c)`** - обновление профиля пользователя (PUT /api/v1/users/me)
- **`GetSettings(c)`** - получение настроек пользователя (GET /api/v1/users/me/settings)
- **`UpdateSettings(c)`** - обновление настроек пользователя (PUT /api/v1/users/me/settings)

### `internal/handler/room.go`

**Назначение:** Обработка запросов для комнат.

**Структуры:**

- **`RoomHandler`** - handler для комнат
  - Поля: roomService, log

- **`CreateRoomRequest`** - запрос на создание комнаты
  - Поля: Title, Description, MaxParticipants

- **`UpdateRoomRequest`** - запрос на обновление комнаты
  - Поля: Title, Description, MaxParticipants

- **`JoinRoomRequest`** - запрос на присоединение к комнате
  - Поля: DisplayName

**Функции:**

- **`NewRoomHandler(roomService, log)`** - создает новый RoomHandler
- **`Create(c)`** - создание новой комнаты (POST /api/v1/rooms)
- **`List(c)`** - получение списка комнат пользователя (GET /api/v1/rooms)
- **`GetByID(c)`** - получение комнаты по ID (GET /api/v1/rooms/:id)
- **`Update(c)`** - обновление комнаты (PUT /api/v1/rooms/:id)
- **`Delete(c)`** - удаление комнаты (DELETE /api/v1/rooms/:id)
- **`Join(c)`** - присоединение к комнате (POST /api/v1/rooms/:id/join)
- **`Leave(c)`** - выход из комнаты (POST /api/v1/rooms/:id/leave)
- **`CreateInvite(c)`** - создание приглашения в комнату (POST /api/v1/rooms/:id/invite)
- **`GetParticipants(c)`** - получение списка участников комнаты (GET /api/v1/rooms/:id/participants)

### `internal/handler/chat.go`

**Назначение:** Обработка запросов для чата.

**Структуры:**

- **`ChatHandler`** - handler для чата
  - Поля: chatService, log

- **`SendMessageRequest`** - запрос на отправку сообщения
  - Поля: Content

- **`EditMessageRequest`** - запрос на редактирование сообщения
  - Поля: Content

**Функции:**

- **`NewChatHandler(chatService, log)`** - создает новый ChatHandler
- **`GetMessages(c)`** - получение сообщений комнаты (GET /api/v1/rooms/:id/chat/messages)
- **`SendMessage(c)`** - отправка сообщения (POST /api/v1/rooms/:id/chat/messages)
- **`EditMessage(c)`** - редактирование сообщения (PUT /api/v1/rooms/:id/chat/messages/:messageId)
- **`DeleteMessage(c)`** - удаление сообщения (DELETE /api/v1/rooms/:id/chat/messages/:messageId)

### `internal/handler/media.go`

**Назначение:** Обработка запросов для медиа (LiveKit токены).

**Структуры:**

- **`MediaHandler`** - handler для медиа
  - Поля: mediaService, log

- **`GetTokenRequest`** - запрос на получение токена
  - Поля: DisplayName

**Функции:**

- **`NewMediaHandler(mediaService, log)`** - создает новый MediaHandler
- **`GetToken(c)`** - получение LiveKit токена для подключения к комнате (POST /api/v1/rooms/:id/media/token)

### `internal/handler/websocket.go`

**Назначение:** Обработка WebSocket соединений для чата.

**Структуры:**

- **`WebSocketHandler`** - handler для WebSocket
  - Поля: chatService, log

**Функции:**

- **`NewWebSocketHandler(chatService, log)`** - создает новый WebSocketHandler
- **`HandleChat(c)`** - обработка WebSocket соединения для чата (GET /ws/chat/:id)
  - Примечание: Текущая реализация только echo, требует доработки

### `internal/handler/waiting_room.go`

**Назначение:** Обработка запросов для комнаты ожидания.

**Структуры:**

- **`WaitingRoomHandler`** - handler для waiting room
  - Поля: roomService, log

**Функции:**

- **`NewWaitingRoomHandler(roomService, log)`** - создает новый WaitingRoomHandler
- **`List(c)`** - получение списка ожидающих (GET /api/v1/rooms/:id/waiting-room)
  - Примечание: Требует реализации
- **`Approve(c)`** - одобрение входа в комнату (POST /api/v1/rooms/:id/waiting-room/:entryId/approve)
  - Примечание: Требует реализации
- **`Reject(c)`** - отклонение входа в комнату (POST /api/v1/rooms/:id/waiting-room/:entryId/reject)
  - Примечание: Требует реализации

### `internal/handler/stats.go`

**Назначение:** Обработка запросов для статистики.

**Структуры:**

- **`StatsHandler`** - handler для статистики
  - Поля: statsService, log

**Функции:**

- **`NewStatsHandler(statsService, log)`** - создает новый StatsHandler
- **`GetRoomStats(c)`** - получение статистики комнаты (GET /api/v1/rooms/:id/stats)
- **`GetParticipantStats(c)`** - получение статистики участника (GET /api/v1/rooms/:id/stats/participants/:participantId)

---

## Сервисы

### `internal/service/services.go`

**Назначение:** Контейнер для всех сервисов.

**Структуры:**

- **`Services`** - содержит все сервисы приложения
  - Поля: Auth, User, Room, Chat, Media, Stats, RateLimit, Audit

**Функции:**

- **`NewServices(repos, cfg, log)`** - создает новый экземпляр Services со всеми зависимостями

### `internal/service/auth.go`

**Назначение:** Бизнес-логика аутентификации и авторизации.

**Интерфейсы:**

- **`AuthService`** - интерфейс сервиса аутентификации
  - Методы: Register, Login, RefreshToken, ValidateToken, Logout

**Структуры:**

- **`LoginResponse`** - ответ на вход
  - Поля: User, AccessToken, RefreshToken

- **`TokenResponse`** - ответ с токенами
  - Поля: AccessToken, RefreshToken

- **`authService`** - реализация AuthService
  - Поля: userRepo, jwtCfg, log

**Функции:**

- **`NewAuthService(userRepo, jwtCfg, log)`** - создает новый AuthService
- **`Register(ctx, email, password, displayName)`** - регистрация нового пользователя
  - Проверяет существование пользователя
  - Хеширует пароль с помощью bcrypt
  - Создает пользователя в БД
- **`Login(ctx, email, password)`** - вход пользователя
  - Проверяет email и пароль
  - Генерирует access и refresh токены
  - Создает сессию в БД
  - Обновляет время последнего входа
- **`RefreshToken(ctx, refreshToken)`** - обновление токена доступа
  - Валидирует refresh token
  - Проверяет сессию в БД
  - Генерирует новые токены
  - Отзывает старую сессию и создает новую
- **`ValidateToken(ctx, tokenString)`** - валидация access token
  - Проверяет токен и возвращает пользователя
- **`Logout(ctx, refreshToken)`** - выход пользователя
  - Отзывает сессию в БД
- **`hashToken(token)`** - хеширует токен для хранения в БД (SHA256)

### `internal/service/user.go`

**Назначение:** Бизнес-логика для пользователей.

**Интерфейсы:**

- **`UserService`** - интерфейс сервиса пользователей
  - Методы: GetMe, UpdateMe, GetSettings, UpdateSettings

**Структуры:**

- **`userService`** - реализация UserService
  - Поля: userRepo, auditRepo, log

**Функции:**

- **`NewUserService(userRepo, auditRepo, log)`** - создает новый UserService
- **`GetMe(ctx, userID)`** - получение информации о текущем пользователе
- **`UpdateMe(ctx, userID, displayName, avatarURL)`** - обновление профиля пользователя
- **`GetSettings(ctx, userID)`** - получение настроек пользователя
- **`UpdateSettings(ctx, userID, settings)`** - обновление настроек пользователя

### `internal/service/room.go`

**Назначение:** Бизнес-логика для комнат.

**Интерфейсы:**

- **`RoomService`** - интерфейс сервиса комнат
  - Методы: Create, GetByID, List, Update, Delete, Join, Leave, CreateInvite, GetParticipants

**Структуры:**

- **`roomService`** - реализация RoomService
  - Поля: roomRepo, auditRepo, cfg, log

**Функции:**

- **`NewRoomService(roomRepo, auditRepo, cfg, log)`** - создает новый RoomService
- **`Create(ctx, hostUserID, title, description, maxParticipants)`** - создание комнаты
  - Валидирует maxParticipants (1-500)
  - Создает комнату со статусом "scheduled"
  - Создает запись аудита
- **`GetByID(ctx, roomID)`** - получение комнаты по ID
- **`List(ctx, userID, limit, offset)`** - получение списка комнат пользователя
  - Валидирует limit (1-100)
- **`Update(ctx, roomID, userID, title, description, maxParticipants)`** - обновление комнаты
  - Проверяет права хоста
  - Валидирует maxParticipants
- **`Delete(ctx, roomID, userID)`** - удаление комнаты
  - Проверяет права хоста
- **`Join(ctx, roomID, userID, displayName)`** - присоединение к комнате
  - Проверяет статус комнаты
  - Если включен waiting room и пользователь не хост, создает запись в waiting room
  - Иначе создает участника
  - Обновляет статус комнаты на "active" при первом присоединении
- **`Leave(ctx, roomID, userID)`** - выход из комнаты
  - Обновляет запись участника
- **`CreateInvite(ctx, roomID, userID, label, expiresAt, maxUses)`** - создание приглашения
  - Проверяет права хоста
  - Генерирует уникальный токен
- **`GetParticipants(ctx, roomID)`** - получение списка участников комнаты

### `internal/service/chat.go`

**Назначение:** Бизнес-логика для чата.

**Интерфейсы:**

- **`ChatService`** - интерфейс сервиса чата
  - Методы: SendMessage, GetMessages, EditMessage, DeleteMessage

**Структуры:**

- **`chatService`** - реализация ChatService
  - Поля: chatRepo, roomRepo, auditRepo, log

**Функции:**

- **`NewChatService(chatRepo, roomRepo, auditRepo, log)`** - создает новый ChatService
- **`SendMessage(ctx, roomID, userID, content)`** - отправка сообщения
  - Проверяет существование комнаты
  - Получает или создает участника
  - Создает сообщение
- **`GetMessages(ctx, roomID, limit, offset)`** - получение сообщений
  - Валидирует limit (1-100)
- **`EditMessage(ctx, messageID, userID, content)`** - редактирование сообщения
  - Проверяет права отправителя
  - Обновляет сообщение
- **`DeleteMessage(ctx, messageID, userID)`** - удаление сообщения
  - Проверяет права отправителя
  - Помечает сообщение как удаленное

### `internal/service/media.go`

**Назначение:** Бизнес-логика для медиа (LiveKit токены).

**Интерфейсы:**

- **`MediaService`** - интерфейс сервиса медиа
  - Методы: GetToken

**Структуры:**

- **`mediaService`** - реализация MediaService
  - Поля: roomRepo, cfg, log

**Функции:**

- **`NewMediaService(roomRepo, cfg, log)`** - создает новый MediaService
- **`GetToken(ctx, roomID, userID, displayName)`** - генерация LiveKit токена
  - Проверяет существование комнаты
  - Создает access token с правами на публикацию и подписку
  - Устанавливает identity и имя пользователя
  - Токен действителен 1 час

### `internal/service/stats.go`

**Назначение:** Бизнес-логика для статистики.

**Интерфейсы:**

- **`StatsService`** - интерфейс сервиса статистики
  - Методы: GetParticipantStats, GetRoomStats, SaveParticipantStats

**Структуры:**

- **`statsService`** - реализация StatsService
  - Поля: statsRepo, log

**Функции:**

- **`NewStatsService(statsRepo, log)`** - создает новый StatsService
- **`GetParticipantStats(ctx, participantID)`** - получение статистики участника
- **`GetRoomStats(ctx, roomID)`** - получение статистики комнаты
- **`SaveParticipantStats(ctx, stats)`** - сохранение статистики участника

### `internal/service/audit.go`

**Назначение:** Бизнес-логика для аудита.

**Интерфейсы:**

- **`AuditService`** - интерфейс сервиса аудита
  - Методы: LogEvent

**Структуры:**

- **`auditService`** - реализация AuditService
  - Поля: auditRepo, log

**Функции:**

- **`NewAuditService(auditRepo, log)`** - создает новый AuditService
- **`LogEvent(ctx, actorUserID, actorRole, roomID, eventType, payload)`** - создание записи аудита
  - Создает запись с текущим временем и переданными данными

### `internal/service/rate_limit.go`

**Назначение:** Бизнес-логика для rate limiting.

**Интерфейсы:**

- **`RateLimitService`** - интерфейс сервиса rate limiting
  - Методы: CheckLimit, Increment

**Структуры:**

- **`rateLimitService`** - реализация RateLimitService
  - Поля: rateLimitRepo, log

**Функции:**

- **`NewRateLimitService(rateLimitRepo, log)`** - создает новый RateLimitService
- **`CheckLimit(ctx, key, limit, windowSeconds)`** - проверка лимита
  - Проверяет, не превышен ли лимит запросов
- **`Increment(ctx, key, windowSeconds)`** - увеличение счетчика запросов
  - Увеличивает счетчик в Redis

---

## Репозитории

### `internal/repository/repositories.go`

**Назначение:** Контейнер для всех репозиториев.

**Структуры:**

- **`Repositories`** - содержит все репозитории приложения
  - Поля: User, Room, Chat, Stats, Audit, RateLimit

**Функции:**

- **`NewRepositories(db, redis, log)`** - создает новый экземпляр Repositories со всеми зависимостями

### `internal/repository/user.go`

**Назначение:** Работа с данными пользователей в PostgreSQL.

**Интерфейсы:**

- **`UserRepository`** - интерфейс репозитория пользователей
  - Методы: Create, GetByID, GetByEmail, Update, CreateSession, GetSessionByTokenHash, RevokeSession, GetSettings, UpdateSettings, CreateVideoProfile, GetVideoProfiles

**Структуры:**

- **`userRepository`** - реализация UserRepository
  - Поля: db, log

**Функции:**

- **`NewUserRepository(db, log)`** - создает новый UserRepository
- **`Create(ctx, user)`** - создание пользователя в БД
- **`GetByID(ctx, id)`** - получение пользователя по ID
- **`GetByEmail(ctx, email)`** - получение пользователя по email
- **`Update(ctx, user)`** - обновление пользователя
- **`CreateSession(ctx, session)`** - создание сессии пользователя
- **`GetSessionByTokenHash(ctx, tokenHash)`** - получение сессии по хешу токена
  - Проверяет, что сессия не отозвана и не истекла
- **`RevokeSession(ctx, sessionID, reason)`** - отзыв сессии
- **`GetSettings(ctx, userID)`** - получение настроек пользователя
  - Если настроек нет, создает дефолтные
- **`createDefaultSettings(ctx, userID)`** - создание настроек по умолчанию
- **`UpdateSettings(ctx, settings)`** - обновление настроек пользователя
- **`CreateVideoProfile(ctx, profile)`** - создание видео профиля
- **`GetVideoProfiles(ctx, userID)`** - получение видео профилей пользователя

### `internal/repository/room.go`

**Назначение:** Работа с данными комнат в PostgreSQL.

**Интерфейсы:**

- **`RoomRepository`** - интерфейс репозитория комнат
  - Методы: Create, GetByID, GetByLiveKitRoomName, List, Update, Delete, CreateInvite, GetInviteByToken, IncrementInviteUsage, CreateParticipant, GetParticipant, GetParticipantByID, GetParticipantsByRoom, UpdateParticipant, CreateWaitingRoomEntry, GetWaitingRoomEntries, UpdateWaitingRoomEntry

**Структуры:**

- **`roomRepository`** - реализация RoomRepository
  - Поля: db, log

**Функции:**

- **`NewRoomRepository(db, log)`** - создает новый RoomRepository
- **`Create(ctx, room)`** - создание комнаты
- **`GetByID(ctx, id)`** - получение комнаты по ID
- **`GetByLiveKitRoomName(ctx, name)`** - получение комнаты по имени в LiveKit
- **`List(ctx, userID, limit, offset)`** - получение списка комнат пользователя
- **`Update(ctx, room)`** - обновление комнаты
- **`Delete(ctx, id)`** - удаление комнаты
- **`CreateInvite(ctx, invite)`** - создание приглашения
- **`GetInviteByToken(ctx, token)`** - получение приглашения по токену
- **`IncrementInviteUsage(ctx, inviteID)`** - увеличение счетчика использования приглашения
- **`CreateParticipant(ctx, participant)`** - создание участника
- **`GetParticipant(ctx, roomID, userID)`** - получение участника по комнате и пользователю
  - Возвращает только активных участников (left_at IS NULL)
- **`GetParticipantByID(ctx, participantID)`** - получение участника по ID
- **`GetParticipantsByRoom(ctx, roomID)`** - получение всех активных участников комнаты
- **`UpdateParticipant(ctx, participant)`** - обновление участника
- **`CreateWaitingRoomEntry(ctx, entry)`** - создание записи в waiting room
- **`GetWaitingRoomEntries(ctx, roomID, status)`** - получение записей waiting room по статусу
- **`UpdateWaitingRoomEntry(ctx, entry)`** - обновление записи waiting room

### `internal/repository/chat.go`

**Назначение:** Работа с данными чата в PostgreSQL.

**Интерфейсы:**

- **`ChatRepository`** - интерфейс репозитория чата
  - Методы: CreateMessage, GetMessages, GetMessageByID, UpdateMessage, DeleteMessage

**Структуры:**

- **`chatRepository`** - реализация ChatRepository
  - Поля: db, log

**Функции:**

- **`NewChatRepository(db, log)`** - создает новый ChatRepository
- **`CreateMessage(ctx, message)`** - создание сообщения
- **`GetMessages(ctx, roomID, limit, offset)`** - получение сообщений комнаты
  - Возвращает только неудаленные сообщения
  - Сортировка по дате создания (DESC)
- **`GetMessageByID(ctx, messageID)`** - получение сообщения по ID
- **`UpdateMessage(ctx, message)`** - обновление сообщения
  - Устанавливает edited_at
- **`DeleteMessage(ctx, messageID, deletedByParticipantID)`** - удаление сообщения
  - Помечает сообщение как удаленное (soft delete)

### `internal/repository/stats.go`

**Назначение:** Работа со статистикой в PostgreSQL.

**Интерфейсы:**

- **`StatsRepository`** - интерфейс репозитория статистики
  - Методы: CreateParticipantStats, GetParticipantStats, GetRoomStats

**Структуры:**

- **`statsRepository`** - реализация StatsRepository
  - Поля: db, log

**Функции:**

- **`NewStatsRepository(db, log)`** - создает новый StatsRepository
- **`CreateParticipantStats(ctx, stats)`** - создание статистики участника
- **`GetParticipantStats(ctx, participantID)`** - получение статистики участника
- **`GetRoomStats(ctx, roomID)`** - получение статистики всех участников комнаты
  - Использует JOIN с room_participants

### `internal/repository/audit.go`

**Назначение:** Работа с аудитом в PostgreSQL.

**Интерфейсы:**

- **`AuditRepository`** - интерфейс репозитория аудита
  - Методы: CreateLog

**Структуры:**

- **`auditRepository`** - реализация AuditRepository
  - Поля: db, log

**Функции:**

- **`NewAuditRepository(db, log)`** - создает новый AuditRepository
- **`CreateLog(ctx, auditLog)`** - создание записи аудита

### `internal/repository/rate_limit.go`

**Назначение:** Работа с rate limiting в Redis.

**Интерфейсы:**

- **`RateLimitRepository`** - интерфейс репозитория rate limiting
  - Методы: CheckLimit, Increment

**Структуры:**

- **`rateLimitRepository`** - реализация RateLimitRepository
  - Поля: redis, log

**Функции:**

- **`NewRateLimitRepository(redis, log)`** - создает новый RateLimitRepository
- **`CheckLimit(ctx, key, limit, window)`** - проверка лимита
  - Проверяет текущее значение счетчика в Redis
- **`Increment(ctx, key, window)`** - увеличение счетчика
  - Увеличивает счетчик в Redis
  - Устанавливает TTL при первом создании ключа

---

## Middleware

### `internal/middleware/auth.go`

**Назначение:** Middleware для аутентификации.

**Структуры:**

- **`AuthMiddleware`** - middleware для аутентификации
  - Поля: authService, log

**Функции:**

- **`NewAuthMiddleware(authService, log)`** - создает новый AuthMiddleware
- **`RequireAuth()`** - middleware функция для проверки аутентификации
  - Извлекает токен из заголовка Authorization
  - Валидирует токен через AuthService
  - Устанавливает user_id, user_email, user_role в контекст Gin

### `internal/middleware/rate_limit.go`

**Назначение:** Middleware для rate limiting.

**Структуры:**

- **`RateLimitMiddleware`** - middleware для rate limiting
  - Поля: rateLimitService, log

**Функции:**

- **`NewRateLimitMiddleware(rateLimitService, log)`** - создает новый RateLimitMiddleware
- **`Limit()`** - middleware функция для ограничения скорости запросов
  - Использует IP адрес клиента как ключ
  - Лимит: 100 запросов в 60 секунд
  - Устанавливает заголовки X-RateLimit-Limit и X-RateLimit-Remaining

### `internal/middleware/cors.go`

**Назначение:** Middleware для CORS.

**Функции:**

- **`CORS()`** - middleware функция для обработки CORS
  - Разрешает все источники (*)
  - Разрешает все методы (POST, OPTIONS, GET, PUT, DELETE, PATCH)
  - Обрабатывает preflight запросы (OPTIONS)

### `internal/middleware/error_handler.go`

**Назначение:** Middleware для обработки ошибок.

**Функции:**

- **`ErrorHandler()`** - middleware функция для обработки ошибок
  - Проверяет наличие ошибок в контексте Gin
  - Определяет HTTP статус код через errors.HTTPStatusFromError
  - Возвращает JSON с ошибкой

### `internal/middleware/request_logger.go`

**Назначение:** Middleware для логирования запросов.

**Функции:**

- **`RequestLogger()`** - middleware функция для логирования запросов
  - Логирует: IP адрес, метод, путь, статус код, время выполнения
  - Формат: `[IP] METHOD PATH STATUS LATENCY`

---

## Утилиты

### `pkg/jwt/jwt.go`

**Назначение:** Утилиты для работы с JWT токенами.

**Структуры:**

- **`Claims`** - структура claims для access token
  - Поля: UserID, Email, Role, RegisteredClaims

**Функции:**

- **`GenerateAccessToken(userID, email, role, secret, ttl)`** - генерация access token
  - Использует алгоритм HS256
  - Включает UserID, Email, Role в claims
- **`GenerateRefreshToken(userID, secret, ttl)`** - генерация refresh token
  - Использует алгоритм HS256
  - Содержит только стандартные claims
- **`ValidateToken(tokenString, secret)`** - валидация access token
  - Проверяет подпись и срок действия
  - Возвращает Claims
- **`ValidateRefreshToken(tokenString, secret)`** - валидация refresh token
  - Проверяет подпись и срок действия
  - Возвращает RegisteredClaims

### `pkg/logger/logger.go`

**Назначение:** Утилиты для логирования.

**Интерфейсы:**

- **`Logger`** - интерфейс логгера
  - Методы: Debug, Info, Warn, Error, Fatal

**Структуры:**

- **`logger`** - реализация Logger
  - Поля: *slog.Logger

**Функции:**

- **`New(level)`** - создает новый логгер
  - Поддерживает уровни: debug, info, warn, error
  - Использует JSON формат вывода
- **`Debug(msg, args...)`** - логирование на уровне debug
- **`Info(msg, args...)`** - логирование на уровне info
- **`Warn(msg, args...)`** - логирование на уровне warn
- **`Error(msg, args...)`** - логирование на уровне error
- **`Fatal(msg, args...)`** - логирование на уровне error и завершение программы

### `pkg/errors/errors.go`

**Назначение:** Утилиты для обработки ошибок.

**Переменные:**

- Предопределенные ошибки: `ErrNotFound`, `ErrUnauthorized`, `ErrForbidden`, `ErrBadRequest`, `ErrInternalServer`, `ErrInvalidCredentials`, `ErrUserAlreadyExists`, `ErrRoomNotFound`, `ErrParticipantNotFound`, `ErrInvalidToken`, `ErrTokenExpired`

**Структуры:**

- **`APIError`** - структура ошибки API
  - Поля: Message, Code

**Функции:**

- **`NewAPIError(message, code)`** - создание новой ошибки API
- **`HTTPStatusFromError(err)`** - определение HTTP статус кода по ошибке
  - Маппинг ошибок на соответствующие HTTP статусы

---

## Примечания

### Не реализовано полностью:

1. **Waiting Room** - handlers для waiting room требуют реализации бизнес-логики
2. **WebSocket чат** - текущая реализация только echo, требует полной реализации
3. **Запись видеоконференций** - не реализована

### Рекомендации по улучшению:

1. Добавить валидацию входных данных на уровне handlers
2. Реализовать полную функциональность waiting room
3. Добавить WebSocket для real-time чата
4. Добавить тесты для всех слоев
5. Добавить метрики и мониторинг
6. Реализовать rate limiting на уровне пользователя, а не только IP

---

*Документация создана автоматически на основе анализа кодовой базы проекта.*






