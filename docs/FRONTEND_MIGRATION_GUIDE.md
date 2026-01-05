# Руководство по миграции фронтенда на анонимную систему

## Основные изменения

### 1. Убрать аутентификацию

**Удалить:**
- Все запросы к `/api/v1/auth/*` (register, login, refresh)
- Хранение токенов в localStorage
- Компоненты регистрации и входа

**Было:**
```javascript
// Старый код с аутентификацией
const token = localStorage.getItem('accessToken');
const headers = {
  'Authorization': `Bearer ${token}`,
  'Content-Type': 'application/json'
};
```

**Стало:**
```javascript
// Новый код без аутентификации
function getParticipantID() {
  let participantID = localStorage.getItem('participant_id');
  if (!participantID) {
    participantID = crypto.randomUUID();
    localStorage.setItem('participant_id', participantID);
  }
  return participantID;
}

const headers = {
  'X-Participant-ID': getParticipantID(),
  'Content-Type': 'application/json'
};
```

### 2. Обновить API клиент

#### Создание комнаты

**Было:**
```javascript
async function createRoom(title, description) {
  const token = getAccessToken();
  const response = await fetch('/api/v1/rooms', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ title, description, max_participants: 10 })
  });
  return response.json();
}
```

**Стало:**
```javascript
async function createRoom(title, description, displayName) {
  const participantID = getParticipantID();
  const response = await fetch('/api/v1/rooms', {
    method: 'POST',
    headers: {
      'X-Participant-ID': participantID,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ 
      title, 
      description, 
      max_participants: 10,
      display_name: displayName 
    })
  });
  
  if (!response.ok) {
    throw new Error('Failed to create room');
  }
  
  const data = await response.json();
  // Сохраняем participant_id из ответа (если сервер его вернул)
  if (data.participant_id) {
    localStorage.setItem('participant_id', data.participant_id);
  }
  
  return data;
}
```

#### Присоединение к комнате

**Было:**
```javascript
async function joinRoom(roomId) {
  const token = getAccessToken();
  const response = await fetch(`/api/v1/rooms/${roomId}/join`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({})
  });
  return response.json();
}
```

**Стало:**
```javascript
async function joinRoom(roomId, displayName) {
  const participantID = getParticipantID();
  const response = await fetch(`/api/v1/rooms/${roomId}/join`, {
    method: 'POST',
    headers: {
      'X-Participant-ID': participantID,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ 
      display_name: displayName 
    })
  });
  
  if (!response.ok) {
    throw new Error('Failed to join room');
  }
  
  return response.json();
}
```

#### Отправка сообщений в чат

**Было:**
```javascript
async function sendMessage(roomId, content) {
  const token = getAccessToken();
  const response = await fetch(`/api/v1/rooms/${roomId}/chat/messages`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ content })
  });
  return response.json();
}
```

**Стало:**
```javascript
async function sendMessage(roomId, content, displayName) {
  const participantID = getParticipantID();
  const response = await fetch(`/api/v1/rooms/${roomId}/chat/messages`, {
    method: 'POST',
    headers: {
      'X-Participant-ID': participantID,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ 
      content,
      display_name: displayName 
    })
  });
  
  if (!response.ok) {
    throw new Error('Failed to send message');
  }
  
  return response.json();
}
```

### 3. Обновить UI компоненты

#### Главная страница (index.html)

**Убрать:**
- Кнопку "Войти" (если требуется аутентификация)
- Форму регистрации

**Добавить:**
- Поле для ввода имени (display_name) перед созданием комнаты
- Поле для ввода ID комнаты для присоединения

**Пример:**
```html
<div class="hero-content">
  <h1>Создавай комнаты, подключайся, делись экраном</h1>
  
  <!-- Создание комнаты -->
  <div class="create-room-section">
    <input type="text" id="display-name-input" placeholder="Ваше имя" />
    <input type="text" id="room-title-input" placeholder="Название комнаты" />
    <button id="btn-create-room">Создать комнату</button>
  </div>
  
  <!-- Присоединение к комнате -->
  <div class="join-room-section">
    <input type="text" id="room-id-input" placeholder="ID комнаты" />
    <input type="text" id="join-display-name-input" placeholder="Ваше имя" />
    <button id="btn-join-room">Войти в комнату</button>
  </div>
</div>
```

#### Комната (room.html)

**Обновить:**
- Убрать отображение профиля пользователя
- Показывать только display_name участника
- Обновить логику отправки сообщений

### 4. WebSocket подключение

**Было:**
```javascript
const ws = new WebSocket(`ws://localhost/ws/chat/${roomId}?token=${token}`);
```

**Стало:**
```javascript
const participantID = getParticipantID();
const ws = new WebSocket(`ws://localhost/ws/chat/${roomId}?participant_id=${participantID}`);
```

### 5. Утилиты для работы с participant_id

Создать файл `js/utils.js`:

```javascript
// Генерация и сохранение participant_id
export function getParticipantID() {
  let participantID = localStorage.getItem('participant_id');
  if (!participantID) {
    participantID = crypto.randomUUID();
    localStorage.setItem('participant_id', participantID);
  }
  return participantID;
}

// Получение заголовков для API запросов
export function getAPIHeaders() {
  return {
    'X-Participant-ID': getParticipantID(),
    'Content-Type': 'application/json'
  };
}

// Валидация UUID формата
export function isValidUUID(str) {
  const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
  return uuidRegex.test(str);
}
```

### 6. Обработка ошибок

**Добавить обработку случаев:**
- Комната не найдена (404)
- Участник не найден (403)
- Превышен лимит участников (429)

```javascript
async function handleAPIError(response) {
  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Unknown error' }));
    
    switch (response.status) {
      case 404:
        throw new Error('Комната не найдена');
      case 403:
        throw new Error('Доступ запрещен');
      case 429:
        throw new Error('Слишком много запросов. Попробуйте позже.');
      default:
        throw new Error(error.error || 'Произошла ошибка');
    }
  }
}
```

## Пример полного API клиента

```javascript
// api.js
const API_BASE = window.location.origin + '/api/v1';

function getParticipantID() {
  let id = localStorage.getItem('participant_id');
  if (!id) {
    id = crypto.randomUUID();
    localStorage.setItem('participant_id', id);
  }
  return id;
}

function getHeaders() {
  return {
    'X-Participant-ID': getParticipantID(),
    'Content-Type': 'application/json'
  };
}

export const api = {
  // Создать комнату
  async createRoom(title, description, displayName) {
    const response = await fetch(`${API_BASE}/rooms`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({ title, description, display_name: displayName, max_participants: 10 })
    });
    
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to create room');
    }
    
    return response.json();
  },
  
  // Получить информацию о комнате
  async getRoom(roomId) {
    const response = await fetch(`${API_BASE}/rooms/${roomId}`, {
      headers: getHeaders()
    });
    
    if (!response.ok) {
      throw new Error('Room not found');
    }
    
    return response.json();
  },
  
  // Присоединиться к комнате
  async joinRoom(roomId, displayName) {
    const response = await fetch(`${API_BASE}/rooms/${roomId}/join`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({ display_name: displayName })
    });
    
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to join room');
    }
    
    return response.json();
  },
  
  // Отправить сообщение
  async sendMessage(roomId, content, displayName) {
    const response = await fetch(`${API_BASE}/rooms/${roomId}/chat/messages`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({ content, display_name: displayName })
    });
    
    if (!response.ok) {
      throw new Error('Failed to send message');
    }
    
    return response.json();
  },
  
  // Получить сообщения
  async getMessages(roomId, limit = 50) {
    const response = await fetch(`${API_BASE}/rooms/${roomId}/chat/messages?limit=${limit}`, {
      headers: getHeaders()
    });
    
    if (!response.ok) {
      throw new Error('Failed to get messages');
    }
    
    return response.json();
  },
  
  // Получить LiveKit токен
  async getMediaToken(roomId) {
    const response = await fetch(`${API_BASE}/rooms/${roomId}/media/token`, {
      method: 'POST',
      headers: getHeaders()
    });
    
    if (!response.ok) {
      throw new Error('Failed to get media token');
    }
    
    return response.json();
  }
};
```

## Чеклист миграции

- [ ] Удалить все компоненты аутентификации
- [ ] Удалить хранение токенов
- [ ] Добавить генерацию participant_id
- [ ] Обновить все API запросы с заголовком X-Participant-ID
- [ ] Обновить формы создания/присоединения к комнате
- [ ] Обновить WebSocket подключения
- [ ] Обновить обработку ошибок
- [ ] Протестировать все сценарии использования

