# Запуск Auth-сервиса для VibeMeet

## Проблема
Порт 5432 уже занят PostgreSQL из vibeemeet, поэтому NextUp-website postgres запускается на порту 5433.

## Быстрый запуск

### 1. Запустите PostgreSQL для auth-сервиса

```bash
cd /Users/ilpaka/Development/NextUp-website
docker compose up -d
```

PostgreSQL запустится на порту **5433** (чтобы не конфликтовать с vibeemeet на 5432).

### 2. Запустите auth-сервис

В том же терминале:

```bash
go run ./cmd/api
```

Или используйте Make:

```bash
make dev
```

Auth-сервис запустится на `http://localhost:8080`

### 3. Проверьте, что сервис работает

```bash
curl http://localhost:8080/health
```

Должен вернуться: `{"status":"ok"}`

## Проверка работы

После запуска auth-сервиса:

1. Перезапустите nginx в vibeemeet:
   ```bash
   cd /Users/ilpaka/Development/vibeemeet
   docker-compose restart nginx
   ```

2. Откройте браузер: `http://localhost`

3. Попробуйте зарегистрироваться или войти

## Остановка

```bash
cd /Users/ilpaka/Development/NextUp-website
docker compose down
```

## Примечания

- PostgreSQL для auth-сервиса работает на порту **5433**
- Auth-сервис работает на порту **8080**
- VibeMeet nginx проксирует `/auth-api/` на `host.docker.internal:8080/api/v1/auth/`
