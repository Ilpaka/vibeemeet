# Архитектура анонимной платформы видеоконференций

## Обзор

Платформа полностью анонимная, без регистрации и аутентификации. Пользователи могут только создавать комнаты и входить в них по ID.

## Принципы проектирования

1. **Анонимность**: Нет пользователей, нет аутентификации
2. **Временные данные**: Чат хранится в Redis на 6 часов
3. **Долговременные данные**: Комнаты и участники в PostgreSQL
4. **Масштабируемость**: Минимальные точки отказа, горизонтальное масштабирование

---

## Структура хранения данных

### PostgreSQL (долговременные данные)

#### Таблица: `rooms`
```sql
CREATE TABLE rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    livekit_room_name VARCHAR(255) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    max_participants INTEGER DEFAULT 10,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP, -- Автоматическое удаление неактивных комнат
    INDEX idx_rooms_status (status),
    INDEX idx_rooms_expires_at (expires_at)
);
```

#### Таблица: `room_participants`
```sql
CREATE TABLE room_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    participant_id VARCHAR(255) NOT NULL, -- Временный ID участника (генерируется на клиенте)
    display_name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'participant', -- 'host' или 'participant'
    livekit_sid VARCHAR(255),
    joined_at TIMESTAMP NOT NULL DEFAULT NOW(),
    left_at TIMESTAMP,
    client_ip INET,
    user_agent TEXT,
    INDEX idx_participants_room (room_id),
    INDEX idx_participants_active (room_id, left_at)
);
```

**Изменения:**
- Убрано поле `user_id` (нет пользователей)
- Добавлено поле `participant_id` (временный UUID, генерируется на клиенте)
- `role` определяет, кто создал комнату (host) или присоединился (participant)

### Redis (временные данные - 6 часов TTL)

#### Структура ключей:

1. **Чат сообщения**: `chat:room:{room_id}:messages`
   - Тип: Sorted Set (ZSET)
   - Score: timestamp (миллисекунды)
   - Value: JSON сообщения
   - TTL: 6 часов

2. **Активные участники комнаты**: `room:{room_id}:participants`
   - Тип: Hash
   - Key: participant_id
   - Value: JSON участника
   - TTL: 6 часов

3. **Метаданные комнаты (кэш)**: `room:{room_id}:meta`
   - Тип: Hash
   - TTL: 6 часов

#### Формат сообщения в Redis:
```json
{
  "id": "msg-uuid",
  "room_id": "room-uuid",
  "participant_id": "participant-uuid",
  "display_name": "Имя участника",
  "message_type": "user",
  "content": "Текст сообщения",
  "created_at": "2025-12-21T20:00:00Z"
}
```

---

## Модели данных (Domain)

### Room (упрощенная)
```go
type Room struct {
    ID              uuid.UUID `json:"id"`
    LiveKitRoomName string    `json:"livekit_room_name"`
    Title           string    `json:"title"`
    Description     *string   `json:"description,omitempty"`
    Status          string    `json:"status"` // active, ended
    MaxParticipants int       `json:"max_participants"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
    ExpiresAt       *time.Time `json:"expires_at,omitempty"`
}
```

### RoomParticipant (анонимный)
```go
type RoomParticipant struct {
    ID            uuid.UUID  `json:"id"`
    RoomID        uuid.UUID  `json:"room_id"`
    ParticipantID string     `json:"participant_id"` // Временный ID с клиента
    DisplayName   string     `json:"display_name"`
    Role          string     `json:"role"` // host, participant
    LiveKitSID    *string   `json:"livekit_sid,omitempty"`
    JoinedAt      time.Time `json:"joined_at"`
    LeftAt        *time.Time `json:"left_at,omitempty"`
    ClientIP      *string   `json:"client_ip,omitempty"`
    UserAgent     *string   `json:"user_agent,omitempty"`
}
```

### ChatMessage (для Redis)
```go
type ChatMessage struct {
    ID            string    `json:"id"`
    RoomID        uuid.UUID `json:"room_id"`
    ParticipantID string    `json:"participant_id"`
    DisplayName   string    `json:"display_name"`
    MessageType   string    `json:"message_type"` // user, system
    Content       string    `json:"content"`
    CreatedAt     time.Time `json:"created_at"`
}
```

---

## Backend API Архитектура

### Публичные endpoints (без аутентификации)

```
POST   /api/v1/rooms              - Создать комнату
GET    /api/v1/rooms/:id          - Получить информацию о комнате
POST   /api/v1/rooms/:id/join     - Присоединиться к комнате
POST   /api/v1/rooms/:id/leave    - Покинуть комнату
GET    /api/v1/rooms/:id/participants - Список участников

GET    /api/v1/rooms/:id/chat/messages - Получить сообщения чата
POST   /api/v1/rooms/:id/chat/messages - Отправить сообщение
PUT    /api/v1/rooms/:id/chat/messages/:messageId - Редактировать сообщение
DELETE /api/v1/rooms/:id/chat/messages/:messageId - Удалить сообщение

POST   /api/v1/rooms/:id/media/token - Получить LiveKit токен

WS     /ws/chat/:room_id          - WebSocket для чата
```

### Middleware изменения

1. **Убрать AuthMiddleware** - все endpoints публичные
2. **Добавить RateLimitMiddleware** - защита от злоупотреблений
3. **Добавить ParticipantMiddleware** - валидация participant_id из заголовка или body

### Новый ParticipantMiddleware

```go
// Проверяет наличие participant_id в заголовке X-Participant-ID
// Если нет - генерирует новый UUID и добавляет в контекст
func ParticipantMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        participantID := c.GetHeader("X-Participant-ID")
        if participantID == "" {
            participantID = uuid.New().String()
        }
        c.Set("participant_id", participantID)
        c.Next()
    }
}
```

---

## Repository Layer

### ChatRepository (Redis)

```go
type ChatRepository interface {
    // Сохранить сообщение в Redis (TTL 6 часов)
    SaveMessage(ctx context.Context, roomID uuid.UUID, message *domain.ChatMessage) error
    
    // Получить сообщения из Redis (последние N)
    GetMessages(ctx context.Context, roomID uuid.UUID, limit int) ([]*domain.ChatMessage, error)
    
    // Получить сообщения после определенного времени
    GetMessagesAfter(ctx context.Context, roomID uuid.UUID, after time.Time, limit int) ([]*domain.ChatMessage, error)
    
    // Удалить сообщение
    DeleteMessage(ctx context.Context, roomID uuid.UUID, messageID string) error
    
    // Обновить сообщение
    UpdateMessage(ctx context.Context, roomID uuid.UUID, message *domain.ChatMessage) error
}
```

### RoomRepository (PostgreSQL)

Изменения:
- Убрать методы связанные с `user_id`
- Добавить методы для работы с `participant_id`
- Добавить автоочистку неактивных комнат

---

## Service Layer

### RoomService

```go
type RoomService interface {
    // Создать комнату (без user_id)
    Create(ctx context.Context, title string, description *string, maxParticipants int, participantID string) (*domain.Room, *domain.RoomParticipant, error)
    
    // Получить комнату по ID
    GetByID(ctx context.Context, roomID uuid.UUID) (*domain.Room, error)
    
    // Присоединиться к комнате
    Join(ctx context.Context, roomID uuid.UUID, participantID string, displayName string) (*domain.RoomParticipant, error)
    
    // Покинуть комнату
    Leave(ctx context.Context, roomID uuid.UUID, participantID string) error
    
    // Получить участников комнаты
    GetParticipants(ctx context.Context, roomID uuid.UUID) ([]*domain.RoomParticipant, error)
    
    // Автоочистка неактивных комнат (cron job)
    CleanupInactiveRooms(ctx context.Context, inactiveDuration time.Duration) error
}
```

### ChatService

```go
type ChatService interface {
    // Отправить сообщение (сохранить в Redis)
    SendMessage(ctx context.Context, roomID uuid.UUID, participantID string, displayName string, content string) (*domain.ChatMessage, error)
    
    // Получить сообщения
    GetMessages(ctx context.Context, roomID uuid.UUID, limit int) ([]*domain.ChatMessage, error)
    
    // Редактировать сообщение
    EditMessage(ctx context.Context, roomID uuid.UUID, messageID string, participantID string, newContent string) error
    
    // Удалить сообщение
    DeleteMessage(ctx context.Context, roomID uuid.UUID, messageID string, participantID string) error
    
    // WebSocket: подписка на новые сообщения
    SubscribeToMessages(ctx context.Context, roomID uuid.UUID) (<-chan *domain.ChatMessage, error)
}
```

---

## Frontend изменения

### Убрать:
1. Регистрацию и логин
2. Хранение токенов в localStorage
3. Запросы к `/api/v1/auth/*`

### Добавить:
1. Генерация `participant_id` на клиенте при первом посещении
2. Сохранение `participant_id` в localStorage
3. Отправка `X-Participant-ID` заголовка во всех запросах
4. Упрощенный UI без профиля пользователя

### Изменения в API клиенте:

```javascript
// Генерация и сохранение participant_id
function getParticipantID() {
  let participantID = localStorage.getItem('participant_id');
  if (!participantID) {
    participantID = crypto.randomUUID();
    localStorage.setItem('participant_id', participantID);
  }
  return participantID;
}

// Все запросы с заголовком
const headers = {
  'Content-Type': 'application/json',
  'X-Participant-ID': getParticipantID()
};

// Создание комнаты
async function createRoom(title, description) {
  const response = await fetch('/api/v1/rooms', {
    method: 'POST',
    headers,
    body: JSON.stringify({ title, description, max_participants: 10 })
  });
  return response.json();
}

// Присоединение к комнате
async function joinRoom(roomId, displayName) {
  const response = await fetch(`/api/v1/rooms/${roomId}/join`, {
    method: 'POST',
    headers,
    body: JSON.stringify({ display_name: displayName })
  });
  return response.json();
}
```

---

## Масштабируемость и отказоустойчивость

### Redis Cluster
- Шардирование по `room_id` для распределения нагрузки
- Репликация для отказоустойчивости

### PostgreSQL
- Connection pooling (pgxpool)
- Индексы на часто используемых полях
- Партиционирование таблицы `room_participants` по `room_id` (при необходимости)

### Rate Limiting
- По IP адресу для создания комнат
- По `participant_id` для отправки сообщений
- По `room_id` для запросов к комнате

### Мониторинг
- Метрики Redis (memory usage, operations/sec)
- Метрики PostgreSQL (connection pool, query time)
- Метрики API (latency, error rate)

---

## Миграция данных

### Шаги миграции:

1. Создать новые таблицы с упрощенной схемой
2. Мигрировать существующие комнаты (если нужно)
3. Удалить таблицы пользователей и аутентификации
4. Обновить код для работы с новой схемой
5. Настроить Redis для чата

---

## Безопасность

1. **Rate Limiting**: Защита от злоупотреблений
2. **Input Validation**: Валидация всех входных данных
3. **CORS**: Правильная настройка CORS
4. **IP Tracking**: Логирование IP для модерации
5. **Content Moderation**: Фильтрация контента (опционально)

---

## Производительность

1. **Redis для чата**: Быстрый доступ к временным данным
2. **Кэширование**: Кэш метаданных комнат в Redis
3. **Connection Pooling**: Эффективное использование соединений с БД
4. **WebSocket**: Real-time обновления без polling

