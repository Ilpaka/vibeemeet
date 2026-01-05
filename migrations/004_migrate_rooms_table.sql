-- Миграция для преобразования таблицы rooms из старой схемы в новую
-- Преобразует: name -> livekit_room_name, owner_id -> host_user_id, 
-- is_private -> waiting_room_enabled, is_active -> status, добавляет недостающие поля

DO $$
DECLARE
    has_name BOOLEAN;
    has_livekit_name BOOLEAN;
BEGIN
    -- Проверяем наличие колонок
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'rooms' AND column_name = 'name'
    ) INTO has_name;
    
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'rooms' AND column_name = 'livekit_room_name'
    ) INTO has_livekit_name;

    -- 1. Переименовываем name в livekit_room_name если существует name
    IF has_name AND NOT has_livekit_name THEN
        ALTER TABLE rooms RENAME COLUMN name TO livekit_room_name;
        ALTER TABLE rooms ALTER COLUMN livekit_room_name SET NOT NULL;
        CREATE UNIQUE INDEX IF NOT EXISTS idx_rooms_livekit_name ON rooms(livekit_room_name);
        -- Обновляем переменную после переименования
        has_livekit_name := true;
        has_name := false;
    END IF;

    -- 2. Переименовываем owner_id в host_user_id если существует owner_id
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'rooms' AND column_name = 'owner_id'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'rooms' AND column_name = 'host_user_id'
    ) THEN
        DROP INDEX IF EXISTS idx_rooms_owner;
        ALTER TABLE rooms RENAME COLUMN owner_id TO host_user_id;
        ALTER TABLE rooms ALTER COLUMN host_user_id SET NOT NULL;
        CREATE INDEX IF NOT EXISTS idx_rooms_host ON rooms(host_user_id);
    END IF;

    -- 3. Добавляем title если его нет
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'rooms' AND column_name = 'title'
    ) THEN
        ALTER TABLE rooms ADD COLUMN title TEXT;
        
        -- Заполняем title из livekit_room_name (после переименования это будет актуально)
        IF has_livekit_name THEN
            EXECUTE 'UPDATE rooms SET title = livekit_room_name WHERE title IS NULL';
        ELSE
            EXECUTE 'UPDATE rooms SET title = ''Room '' || id::text WHERE title IS NULL';
        END IF;
        
        ALTER TABLE rooms ALTER COLUMN title SET NOT NULL;
    END IF;

    -- 4. Преобразуем is_private в waiting_room_enabled (инвертируем логику)
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'rooms' AND column_name = 'is_private'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'rooms' AND column_name = 'waiting_room_enabled'
    ) THEN
        ALTER TABLE rooms ADD COLUMN waiting_room_enabled BOOLEAN NOT NULL DEFAULT true;
        UPDATE rooms SET waiting_room_enabled = is_private WHERE waiting_room_enabled = true;
    END IF;

    -- 5. Преобразуем is_active в status
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'rooms' AND column_name = 'is_active'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'rooms' AND column_name = 'status'
    ) THEN
        ALTER TABLE rooms ADD COLUMN status TEXT NOT NULL DEFAULT 'active';
        UPDATE rooms SET status = CASE 
            WHEN is_active = true THEN 'active'
            ELSE 'ended'
        END;
        ALTER TABLE rooms ADD CONSTRAINT rooms_status_check 
            CHECK (status IN ('scheduled','active','ended','cancelled'));
        CREATE INDEX IF NOT EXISTS idx_rooms_status ON rooms(status);
    END IF;

    -- 6. Добавляем недостающие поля
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'rooms' AND column_name = 'scheduled_start_at'
    ) THEN
        ALTER TABLE rooms ADD COLUMN scheduled_start_at TIMESTAMPTZ;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'rooms' AND column_name = 'scheduled_end_at'
    ) THEN
        ALTER TABLE rooms ADD COLUMN scheduled_end_at TIMESTAMPTZ;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'rooms' AND column_name = 'actual_start_at'
    ) THEN
        ALTER TABLE rooms ADD COLUMN actual_start_at TIMESTAMPTZ;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'rooms' AND column_name = 'actual_end_at'
    ) THEN
        ALTER TABLE rooms ADD COLUMN actual_end_at TIMESTAMPTZ;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'rooms' AND column_name = 'is_locked'
    ) THEN
        ALTER TABLE rooms ADD COLUMN is_locked BOOLEAN NOT NULL DEFAULT false;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'rooms' AND column_name = 'settings'
    ) THEN
        ALTER TABLE rooms ADD COLUMN settings JSONB NOT NULL DEFAULT '{}'::jsonb;
    END IF;

    -- 7. Обновляем max_participants (добавляем CHECK constraint)
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'rooms' AND column_name = 'max_participants'
    ) THEN
        ALTER TABLE rooms DROP CONSTRAINT IF EXISTS rooms_max_participants_check;
        ALTER TABLE rooms ADD CONSTRAINT rooms_max_participants_check 
            CHECK (max_participants > 0 AND max_participants <= 500);
    END IF;

    -- 8. Создаем недостающие индексы
    CREATE INDEX IF NOT EXISTS idx_rooms_scheduled_start ON rooms(scheduled_start_at) 
        WHERE scheduled_start_at IS NOT NULL;

END $$;
