-- Миграция для исправления проблемы с username
-- Делает username nullable или удаляет его, если display_name уже есть

-- Вариант 1: Делаем username nullable (если хотим сохранить для совместимости)
DO $$
BEGIN
    -- Проверяем, существует ли колонка username и является ли она NOT NULL
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' 
        AND column_name = 'username' 
        AND is_nullable = 'NO'
    ) THEN
        -- Делаем username nullable
        ALTER TABLE users ALTER COLUMN username DROP NOT NULL;
        
        -- Заполняем NULL значения username из display_name
        UPDATE users SET username = display_name WHERE username IS NULL;
        
        -- Если display_name длиннее 50 символов, обрезаем username
        UPDATE users SET username = LEFT(display_name, 50) 
        WHERE username IS NULL OR LENGTH(username) > 50;
    END IF;
END $$;

-- Вариант 2: Удаляем username полностью (раскомментируйте если нужно)
-- DO $$
-- BEGIN
--     IF EXISTS (
--         SELECT 1 FROM information_schema.columns 
--         WHERE table_name = 'users' AND column_name = 'username'
--     ) THEN
--         -- Удаляем индекс если есть
--         DROP INDEX IF EXISTS idx_users_username;
--         -- Удаляем constraint если есть
--         ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_key;
--         -- Удаляем колонку
--         ALTER TABLE users DROP COLUMN username;
--     END IF;
-- END $$;

