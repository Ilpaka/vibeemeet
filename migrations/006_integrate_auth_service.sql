-- ============================================
-- Интеграция с Auth-сервисом
-- ============================================

-- Изменяем таблицу rooms для поддержки external user_id
ALTER TABLE rooms
  ALTER COLUMN host_user_id TYPE TEXT;

-- Добавляем поля для хранения информации о host из Auth-сервиса
ALTER TABLE rooms
  ADD COLUMN IF NOT EXISTS host_email TEXT,
  ADD COLUMN IF NOT EXISTS host_display_name TEXT;

-- Временно делаем host_display_name nullable для миграции
ALTER TABLE rooms
  ALTER COLUMN host_display_name DROP NOT NULL;

-- Убираем FK constraint на локальную таблицу users
ALTER TABLE rooms
  DROP CONSTRAINT IF EXISTS rooms_host_user_id_fkey;

-- Комментарий
COMMENT ON COLUMN rooms.host_user_id IS 'External user ID from Auth-service (UUID as TEXT)';
COMMENT ON COLUMN rooms.host_email IS 'Host email from Auth-service';
COMMENT ON COLUMN rooms.host_display_name IS 'Host display name from Auth-service';

-- Обновляем существующие комнаты (если есть)
-- Мигрируем локальные user_id на external_id (нужно будет настроить вручную)
UPDATE rooms r
SET 
  host_email = u.email,
  host_display_name = u.display_name
FROM users u
WHERE r.host_user_id::UUID = u.id
  AND (r.host_email IS NULL OR r.host_display_name IS NULL);

-- После миграции можем сделать host_display_name обязательным
-- (раскомментируйте после миграции данных)
-- ALTER TABLE rooms
--   ALTER COLUMN host_display_name SET NOT NULL;

-- Индекс для поиска по host_user_id (внешний ID)
CREATE INDEX IF NOT EXISTS idx_rooms_host_user_id_text ON rooms(host_user_id);

-- ============================================
-- Комментарии к изменениям
-- ============================================

COMMENT ON TABLE rooms IS 'Комнаты видеоконференций (host теперь из Auth-сервиса)';
COMMENT ON TABLE anonymous_rooms IS 'Анонимные комнаты для гостей (без регистрации)';
COMMENT ON TABLE anonymous_participants IS 'Анонимные участники (гости по invite link)';
