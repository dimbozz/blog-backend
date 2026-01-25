-- =====================================================
-- Инициализация базы данных блога
-- Таблицы: users, posts, comments
-- =====================================================

-- 1. Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(30) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. Таблица постов
CREATE TABLE IF NOT EXISTS posts (
    id SERIAL PRIMARY KEY,
    author_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. Таблица комментариев
CREATE TABLE IF NOT EXISTS comments (
    id SERIAL PRIMARY KEY,
    post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    author_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для оптимизации поиска
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_posts_author_id ON posts(author_id);
CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);

-- Добавим комментарии к таблице для документации
COMMENT ON TABLE users IS 'Таблица пользователей системы';
COMMENT ON COLUMN users.email IS 'Email пользователя (уникальный)';
COMMENT ON COLUMN users.username IS 'Имя пользователя (уникальное)';
COMMENT ON COLUMN users.password_hash IS 'Хеш пароля (bcrypt)';
COMMENT ON COLUMN users.created_at IS 'Дата и время регистрации';

COMMENT ON TABLE posts IS 'Таблица постов блога';
COMMENT ON COLUMN posts.author_id IS 'ID автора поста (внешниий ключ → users)';
COMMENT ON COLUMN posts.title IS 'Заголовок поста';
COMMENT ON COLUMN posts.content IS 'Содержимое поста';

COMMENT ON TABLE comments IS 'Таблица комментариев к постам';
COMMENT ON COLUMN comments.post_id IS 'ID поста (внешниий ключ → posts)';
COMMENT ON COLUMN comments.author_id IS 'ID автора комментария (внешниий ключ → users)';
COMMENT ON COLUMN comments.content IS 'Текст комментария';

-- Проверка создания таблиц
DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users') THEN
        RAISE NOTICE '✅ Таблица users создана';
    END IF;
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'posts') THEN
        RAISE NOTICE '✅ Таблица posts создана';
    END IF;
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'comments') THEN
        RAISE NOTICE '✅ Таблица comments создана';
    END IF;
END $$;