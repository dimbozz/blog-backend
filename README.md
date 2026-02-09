# Расширенная система управления блогом

### API эндпоинты:
| Метод  |            Путь             |           Описание                | Требует токен |
|--------|-----------------------------|-----------------------------------|---------------|
|  POST  | `/register`                 | Регистрация пользователя          |      Нет      |
|  POST  | `/login`                    | Вход в систему                    |      Нет      |
|  GET   | `/health`                   | Проверка состояния                |      Нет      |
|  GET   | `/api/posts`                | Получить все посты                |      Нет      |
|  POST  | `/api/posts`                | Создать пост                      |      Да       |
|  GET   | `/api/posts/1`              | Получить один пост                |      Нет      |
|  PUT   | `/api/posts/1`              | Обновить пост                     |      Да       |
| DELETE | `/api/posts/1`              | Удалить пост                      |      Да       |
|  GET   | `/api/posts/1/comments`     | Получить комментарии к посту 1    |      Нет      |
|  POST  | `/api/posts/1/comments`     | Создать комментарий к посту 1     |      Да       |

## 🏗️ Структура проекта

```
blog-backend/
├── README.md
├── cmd
│   └── api
│       └── main.go
├── docker-compose.yml
├── go.mod
├── go.sum
├── init.sql
├── internal
│   ├── config
│   │   └── env.go
│   ├── handlers
│   │   ├── auth_handler_test.go
│   │   ├── comment.go
│   │   ├── health.go
│   │   ├── middleware
│   │   │   ├── auth.go
│   │   │   ├── error.go
│   │   │   ├── panic_recovery.go
│   │   │   └── request_logger.go
│   │   ├── post.go
│   │   ├── post_handler_test.go
│   │   └── user.go
│   ├── model
│   │   └── model.go
│   └── repository
│       ├── interfaces.go
│       └── postgres
│           ├── comment_repositoy.go
│           ├── db.go
│           ├── post_repository.go
│           └── user_repository.go
├── pkg
│   ├── auth
│   │   └── context.go
│   └── jwt
│       └── jwt.go
├── service
│   ├── comment_service.go
│   ├── post_service.go
│   ├── service_test
│   │   ├── memory_post_test.go
│   │   ├── mock_user_repo_test.go
│   │   └── post_publish_test.go
│   └── user_service.go
└── task.md                     # Задание на разработку проекта
```

## 🚀 Быстрый старт

### 1. Настройка окружения

```bash
# Создайте .env файл из примера
cp .env.example .env

# ВАЖНО: Измените JWT_SECRET в .env на свой ключ (минимум 32 символа)
nano .env
```

### 2. Запуск базы данных

```bash
# Запустите PostgreSQL в Docker
docker-compose up -d

# Проверьте, что БД запустилась
docker-compose ps
```

### 3. Установка зависимостей

```bash
# Скачайте Go модули
go mod download
```

### 4. Запуск и тестирование

```bash
# Запустите сервер
go run *.go

# В другом терминале тестируйте API
curl -X POST http://localhost:8088/api/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user-1@example.com","username":"testuser-1","password":"SecurePass123"}'
```

## 🧪 Тестирование API

### Проверка здоровья сервиса
```bash
curl http://localhost:8088/api/health
```

### Регистрация пользователя
```bash
curl -X POST http://localhost:8088/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user-1@example.com",
    "username": "testuser-1",
    "password": "SecurePass123"
  }'
```

### Вход в систему
```bash
curl -X POST http://localhost:8088/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user-1@example.com",
    "password": "SecurePass123"
  }'
```

### Получить все посты
```bash
curl http://localhost:8088/api/posts
```

### Создать пост (требуется JWT токен, полученнный при входе в систему)
```bash
curl -X POST http://localhost:8088/api/posts \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"title":"Пост номер 1","content":"Текст поста номер 1"}'
```

### Создать пост с отложенной публикацией (требуется JWT токен)
```bash
curl -X POST http://localhost:8088/api/posts \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
    -d '{"title":"Проверка публикации",
       "content":"Этот пост опубликуется 2026-02-01 в 09:25",
       "publish_at": "2026-02-01T09:25:00Z"
       }'
```

### Получить один пост c id=1 (без токена)
```bash
curl http://localhost:8088/api/posts/1
```

### Обновить пост id=1 (требуется JWT токен)
```bash
curl -X PUT http://localhost:8088/api/posts/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Обновленный пост",
       "content":"Обновленный пост опубликуется 2026-02-01 в 09:25",
       "publish_at": "2026-02-01T09:25:00Z"
       }'
```

### Удалить пост
```bash
curl -X DELETE http://localhost:8088/api/posts/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Получить все посты с пагинацией
```bash
curl "http://localhost:8088/api/posts?limit=2&offset=1"
```

### Создать комментарий к посту id=1 (требуется JWT токен, полученнный при входе в систему)
```bash
curl -X POST http://localhost:8088/api/posts/1/comments \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"content": "Отличный пост!"}'
```

### Получить все комментарии к посту 1
```bash
curl "curl http://localhost:8088/api/posts/6/comments"
```

## 📊 Автотесты

### Перейти в корень проекта
```bash
cd blog-backend
```

### Тесты handlers
```bash
go test ./internal/handlers -v
```

### Покрытие кода тестами
```bash
go test ./internal/handlers -cover
```

### Тест отложенных публикаций
```bash
go test ./service/service_test -v
```

## 🆘 Получение помощи

### Если что-то не работает:

1. **БД не запускается**
   ```bash
   docker-compose down
   docker-compose up -d
   docker-compose logs postgres
   ```

2. **Ошибки компиляции**
   ```bash
   go mod tidy
   go mod download
   ```

3. **Сервер не запускается**
   - Проверьте .env файл
   - Убедитесь, что JWT_SECRET длиннее 32 символов
   - Проверьте, что PostgreSQL запущен

4. **Тесты API не проходят**
   - Проверьте логи сервера
   - Убедитесь, что все TODO функции реализованы
   - Проверьте правильность JSON в curl запросах