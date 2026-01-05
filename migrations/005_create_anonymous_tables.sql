-- Миграция для анонимной платформы
-- Создание упрощенных таблиц без привязки к пользователям

-- Создание новой таблицы anonymous_rooms
CREATE TABLE IF NOT EXISTS anonymous_rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    livekit_room_name VARCHAR(255) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    max_participants INTEGER DEFAULT 10,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP
);

-- Индексы для anonymous_rooms
CREATE INDEX IF NOT EXISTS idx_anonymous_rooms_status ON anonymous_rooms(status);
CREATE INDEX IF NOT EXISTS idx_anonymous_rooms_expires_at ON anonymous_rooms(expires_at);
CREATE INDEX IF NOT EXISTS idx_anonymous_rooms_created_at ON anonymous_rooms(created_at);

-- Создание новой таблицы anonymous_participants
CREATE TABLE IF NOT EXISTS anonymous_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES anonymous_rooms(id) ON DELETE CASCADE,
    participant_id VARCHAR(255) NOT NULL, -- Временный ID участника (UUID строка с клиента)
    display_name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'participant', -- 'host' или 'participant'
    livekit_sid VARCHAR(255),
    joined_at TIMESTAMP NOT NULL DEFAULT NOW(),
    left_at TIMESTAMP,
    client_ip INET,
    user_agent TEXT,
    CONSTRAINT unique_room_participant UNIQUE (room_id, participant_id, left_at)
);

-- Индексы для anonymous_participants
CREATE INDEX IF NOT EXISTS idx_anonymous_participants_room ON anonymous_participants(room_id);
CREATE INDEX IF NOT EXISTS idx_anonymous_participants_active ON anonymous_participants(room_id, left_at);
CREATE INDEX IF NOT EXISTS idx_anonymous_participants_participant_id ON anonymous_participants(participant_id);

-- Комментарии к таблицам
COMMENT ON TABLE anonymous_rooms IS 'Анонимные комнаты видеоконференций без привязки к пользователям';
COMMENT ON TABLE anonymous_participants IS 'Анонимные участники комнат с временным participant_id';

-- Функция для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_anonymous_rooms_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Триггер для автоматического обновления updated_at
CREATE TRIGGER update_anonymous_rooms_updated_at_trigger
    BEFORE UPDATE ON anonymous_rooms
    FOR EACH ROW
    EXECUTE FUNCTION update_anonymous_rooms_updated_at();

