-- Миграция для исправления таблицы users
-- Преобразует существующую схему в нужную

-- 1. Переименовываем full_name в display_name если существует full_name
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'full_name'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'display_name'
    ) THEN
        ALTER TABLE users RENAME COLUMN full_name TO display_name;
        ALTER TABLE users ALTER COLUMN display_name SET NOT NULL;
    END IF;
END $$;

-- 2. Если есть username но нет display_name, используем username
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'username'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'display_name'
    ) THEN
        ALTER TABLE users ADD COLUMN display_name TEXT;
        UPDATE users SET display_name = username WHERE display_name IS NULL;
        ALTER TABLE users ALTER COLUMN display_name SET NOT NULL;
    END IF;
END $$;

-- 3. Добавляем display_name если его вообще нет
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'display_name'
    ) THEN
        ALTER TABLE users ADD COLUMN display_name TEXT NOT NULL DEFAULT '';
        UPDATE users SET display_name = split_part(email, '@', 1) WHERE display_name = '';
        ALTER TABLE users ALTER COLUMN display_name DROP DEFAULT;
    END IF;
END $$;

-- 4. Изменяем тип email на CITEXT если нужно
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'email' 
        AND data_type != 'USER-DEFINED'
    ) THEN
        -- Создаем временную колонку
        ALTER TABLE users ADD COLUMN email_citext CITEXT;
        UPDATE users SET email_citext = email;
        ALTER TABLE users DROP COLUMN email;
        ALTER TABLE users RENAME COLUMN email_citext TO email;
        ALTER TABLE users ALTER COLUMN email SET NOT NULL;
        -- Восстанавливаем уникальность
        CREATE UNIQUE INDEX IF NOT EXISTS users_email_key ON users(email);
    END IF;
END $$;

-- 5. Добавляем недостающие поля
DO $$
BEGIN
    -- global_role
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'global_role'
    ) THEN
        ALTER TABLE users ADD COLUMN global_role TEXT NOT NULL DEFAULT 'user';
        ALTER TABLE users ADD CONSTRAINT users_global_role_check CHECK (global_role IN ('user', 'technical_admin'));
    END IF;

    -- is_email_verified
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'is_email_verified'
    ) THEN
        ALTER TABLE users ADD COLUMN is_email_verified BOOLEAN NOT NULL DEFAULT false;
    END IF;

    -- last_login_at
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'last_login_at'
    ) THEN
        ALTER TABLE users ADD COLUMN last_login_at TIMESTAMPTZ;
    END IF;
END $$;

-- 6. Удаляем username если он больше не нужен (опционально, можно оставить)
-- DO $$
-- BEGIN
--     IF EXISTS (
--         SELECT 1 FROM information_schema.columns 
--         WHERE table_name = 'users' AND column_name = 'username'
--     ) THEN
--         DROP INDEX IF EXISTS idx_users_username;
--         ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_key;
--         ALTER TABLE users DROP COLUMN username;
--     END IF;
-- END $$;
