-- Полная миграция для создания всех недостающих таблиц
-- Проверяет существование и создает только отсутствующие таблицы

-- Убеждаемся что расширения установлены
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "citext";

-- ============================================
-- 1. user_sessions (refresh tokens) - КРИТИЧНО!
-- ============================================
CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    revoked_reason TEXT,
    ip_address INET,
    user_agent TEXT
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_token_hash ON user_sessions(refresh_token_hash);

-- ============================================
-- 2. room_invites
-- ============================================
CREATE TABLE IF NOT EXISTS room_invites (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    created_by_user_id UUID NOT NULL REFERENCES users(id),
    link_token TEXT NOT NULL UNIQUE,
    label TEXT,
    expires_at TIMESTAMPTZ,
    max_uses INTEGER,
    used_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_room_invites_room_id ON room_invites(room_id);
CREATE INDEX IF NOT EXISTS idx_room_invites_token ON room_invites(link_token);

-- ============================================
-- 3. waiting_room_entries
-- ============================================
CREATE TABLE IF NOT EXISTS waiting_room_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    display_name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','approved','rejected','expired')),
    requested_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    decided_at TIMESTAMPTZ,
    decided_by_user_id UUID REFERENCES users(id),
    reason TEXT
);

CREATE INDEX IF NOT EXISTS idx_wre_room_status ON waiting_room_entries(room_id, status);
CREATE INDEX IF NOT EXISTS idx_wre_user ON waiting_room_entries(user_id);

-- ============================================
-- 4. chat_messages
-- ============================================
CREATE TABLE IF NOT EXISTS chat_messages (
    id BIGSERIAL PRIMARY KEY,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    sender_participant_id UUID REFERENCES room_participants(id),
    message_type TEXT NOT NULL DEFAULT 'user' CHECK (message_type IN ('user','system')),
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    edited_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    deleted_by_participant_id UUID REFERENCES room_participants(id)
);

CREATE INDEX IF NOT EXISTS idx_chat_room_created_at ON chat_messages(room_id, created_at);
CREATE INDEX IF NOT EXISTS idx_chat_sender ON chat_messages(sender_participant_id, created_at);

-- ============================================
-- 5. participant_stats
-- ============================================
CREATE TABLE IF NOT EXISTS participant_stats (
    id BIGSERIAL PRIMARY KEY,
    room_participant_id UUID NOT NULL UNIQUE REFERENCES room_participants(id) ON DELETE CASCADE,
    avg_rtt_ms NUMERIC(10,2),
    max_rtt_ms NUMERIC(10,2),
    avg_jitter_ms NUMERIC(10,2),
    packet_loss_up_pct NUMERIC(5,2),
    packet_loss_down_pct NUMERIC(5,2),
    avg_bitrate_kbps NUMERIC(10,2),
    network_score SMALLINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_ps_network_score ON participant_stats(network_score);

-- ============================================
-- 6. user_settings
-- ============================================
CREATE TABLE IF NOT EXISTS user_settings (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    default_camera_device_id TEXT,
    default_microphone_device_id TEXT,
    default_speaker_device_id TEXT,
    preferred_video_quality TEXT NOT NULL DEFAULT 'auto' CHECK (preferred_video_quality IN ('1080p','720p','480p','360p','auto')),
    preferred_theme TEXT NOT NULL DEFAULT 'system' CHECK (preferred_theme IN ('light','dark','system')),
    mute_mic_on_join BOOLEAN NOT NULL DEFAULT false,
    disable_camera_on_join BOOLEAN NOT NULL DEFAULT false,
    language_code VARCHAR(8) DEFAULT 'en',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================
-- 7. user_video_profiles
-- ============================================
CREATE TABLE IF NOT EXISTS user_video_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    background_type TEXT NOT NULL CHECK (background_type IN ('none','blur','image')),
    background_image_url TEXT,
    noise_suppression_level TEXT NOT NULL DEFAULT 'medium' CHECK (noise_suppression_level IN ('off','low','medium','high')),
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_uvp_user ON user_video_profiles(user_id);

-- ============================================
-- 8. audit_log
-- ============================================
CREATE TABLE IF NOT EXISTS audit_log (
    id BIGSERIAL PRIMARY KEY,
    event_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    actor_user_id UUID REFERENCES users(id),
    actor_role TEXT NOT NULL CHECK (actor_role IN ('user','host','technical_admin','system')),
    room_id UUID REFERENCES rooms(id),
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_audit_room_time ON audit_log(room_id, event_time);
CREATE INDEX IF NOT EXISTS idx_audit_actor_time ON audit_log(actor_user_id, event_time);
CREATE INDEX IF NOT EXISTS idx_audit_event_type ON audit_log(event_type);

-- ============================================
-- 9. Функция для автоматического обновления updated_at
-- ============================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- ============================================
-- 10. Триггеры для updated_at
-- ============================================
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_rooms_updated_at ON rooms;
CREATE TRIGGER update_rooms_updated_at BEFORE UPDATE ON rooms
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_user_settings_updated_at ON user_settings;
CREATE TRIGGER update_user_settings_updated_at BEFORE UPDATE ON user_settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

