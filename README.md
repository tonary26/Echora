# Echora

Echora - локальная персональная аудиотека. Пользователь создаёт профиль, загружает аудиофайлы с необязательными обложками и слушает их прямо в браузере.

Проект состоит из Vue-интерфейса и Go API. Метаданные пользователей и треков хранятся в PostgreSQL, а сами аудиофайлы и обложки - в MinIO. Доступ к файлам выдаётся через временные подписанные ссылки, поэтому S3-бакет не требуется делать публичным.

## Возможности

- регистрация и вход по email и паролю;
- необязательный аватар пользователя по URL;
- JWT-авторизация;
- персональная библиотека для каждого пользователя;
- загрузка аудиофайлов с отображением прогресса;
- необязательная обложка трека;
- встроенный плеер с паузой, перемоткой, громкостью и переключением треков;
- временные подписанные ссылки MinIO для воспроизведения и обложек;
- адаптивный интерфейс для компьютеров и мобильных устройств;
- автоматический fallback в виде виниловой пластинки для треков без обложки.

## Стек

| Часть | Технологии |
|---|---|
| Frontend | Vue 3, Vue Router, Pinia, Axios, Vite |
| Backend | Go, стандартный `net/http`, pgx, JWT, bcrypt |
| База данных | PostgreSQL 16 |
| Файловое хранилище | MinIO, S3-совместимое API |
| Миграции | migrate/migrate |
| Локальная разработка | Docker Compose, Air, Vite HMR |

Redis присутствует в `docker-compose.yml` как заготовка для дальнейшего кэширования или фоновых задач, но текущая версия приложения его пока не использует.

## Быстрый запуск через Docker Compose

Это рекомендуемый способ локального запуска: не требуется отдельно устанавливать Go, Node.js, PostgreSQL или MinIO.

### 1. Требования

Установите:

- Docker Engine или Docker Desktop;
- Docker Compose v2 (`docker compose`);
- Git - если проект ещё нужно клонировать.

Проверка:

```bash
docker --version
docker compose version
```

### 2. Перейдите в каталог проекта

```bash
cd /path/to/echora
```

Все команды ниже выполняются из корня проекта, где находится `docker-compose.yml`.

### 3. Создайте `.env`

Создайте в корне проекта файл `.env`:

```dotenv
# PostgreSQL
DB_HOST=postgres
DB_PORT=5432
DB_USER=echora
DB_PASSWORD=echora_password
DB_NAME=echora
DB_URL=postgres://echora:echora_password@postgres:5432/echora?sslmode=disable

# Go API
APP_PORT=8080
JWT_SECRET=replace-with-a-long-random-secret

# Зарезервировано для будущего использования
REDIS_ADDR=redis:6379

# MinIO
MINIO_ROOT_USER=echora_minio
MINIO_ROOT_PASSWORD=echora_minio_password

# Адреса указываются без http://
S3_ENDPOINT=minio:9000
S3_PUBLIC_ENDPOINT=localhost:9000
S3_ACCESS_KEY=echora_minio
S3_SECRET_KEY=echora_minio_password
S3_BUCKET=echora

# Frontend
VITE_API_URL=http://localhost:8080
```

Для `JWT_SECRET` желательно использовать случайную строку:

```bash
openssl rand -hex 32
```

Обратите внимание:

- `S3_ENDPOINT=minio:9000` используется API внутри Docker-сети;
- `S3_PUBLIC_ENDPOINT=localhost:9000` попадает в подписанные ссылки, которые открывает браузер;
- `S3_ACCESS_KEY` должен совпадать с `MINIO_ROOT_USER`;
- `S3_SECRET_KEY` должен совпадать с `MINIO_ROOT_PASSWORD`;
- адреса MinIO в этой конфигурации указываются без `http://`.

Не публикуйте рабочий `.env` и не храните реальные секреты в репозитории.

### 4. Запустите приложение

```bash
docker compose up --build
```

Для запуска в фоне:

```bash
docker compose up --build -d
```

При первом запуске Docker Compose:

1. соберёт frontend и backend;
2. запустит PostgreSQL, MinIO и Redis;
3. применит SQL-миграции;
4. запустит Go API с Air;
5. запустит Vite dev server.

Контейнер `music_migrate` после успешного применения миграций завершится с кодом `0`. Это нормальное поведение, а не ошибка.

### 5. Проверьте контейнеры

```bash
docker compose ps
```

Основные адреса:

| Сервис | Адрес |
|---|---|
| Echora | <http://localhost:5173> |
| API | <http://localhost:8080> |
| MinIO API | <http://localhost:9000> |
| MinIO Console | <http://localhost:9001> |
| PostgreSQL | `localhost:5432` |
| Redis | `localhost:6379` |

Для входа в MinIO Console используйте значения `MINIO_ROOT_USER` и `MINIO_ROOT_PASSWORD` из `.env`.

## Первый пользовательский сценарий

1. Откройте <http://localhost:5173>.
2. Перейдите на страницу регистрации.
3. Введите имя, email и пароль.
4. При желании добавьте публичную ссылку на аватар.
5. После регистрации откроется личная библиотека.
6. Нажмите «Добавить трек».
7. Выберите аудиофайл и, при желании, изображение обложки.
8. После загрузки нажмите на трек или кнопку воспроизведения.

Frontend принимает следующие расширения аудио:

```text
MP3, WAV, OGG, M4A, AAC, FLAC, OPUS
```

Ограничения интерфейса:

- аудиофайл - до 100 МБ;
- обложка - до 10 МБ;
- обложка должна иметь MIME-тип `image/*`.

Если обложка не выбрана, интерфейс показывает стандартную чёрную пластинку с жёлтой центральной этикеткой.

## Управление приложением

Посмотреть логи всех сервисов:

```bash
docker compose logs -f
```

Логи только API и frontend:

```bash
docker compose logs -f api frontend
```

Перезапустить сервисы:

```bash
docker compose restart
```

Остановить приложение, сохранив данные:

```bash
docker compose down
```

Полностью удалить контейнеры и локальные данные PostgreSQL, MinIO и Redis:

```bash
docker compose down -v
```

Последняя команда необратимо удаляет зарегистрированных пользователей, треки, обложки и другие данные из Docker volumes.

## Запуск частей проекта отдельно

### Frontend

Требуется Node.js, совместимый с ограничением из `frontend/package.json`: Node.js `22.18+` или `24.12+`.

```bash
cd frontend
npm install
VITE_API_URL=http://localhost:8080 npm run dev
```

Frontend будет доступен на <http://localhost:5173>.

### Backend

Требуется Go `1.25+`, а также работающие PostgreSQL и MinIO.

Удобнее сначала поднять инфраструктуру Docker:

```bash
docker compose up -d postgres redis minio migrate
```

Затем экспортируйте переменные для запуска backend с хоста. В отличие от Docker-конфигурации, здесь PostgreSQL и MinIO доступны через `localhost`:

```bash
export DB_URL='postgres://echora:echora_password@localhost:5432/echora?sslmode=disable'
export APP_PORT='8080'
export JWT_SECRET='replace-with-a-long-random-secret'
export S3_ENDPOINT='localhost:9000'
export S3_PUBLIC_ENDPOINT='localhost:9000'
export S3_ACCESS_KEY='echora_minio'
export S3_SECRET_KEY='echora_minio_password'
export S3_BUCKET='echora'

cd backend
go run ./cmd/api
```

Значения должны соответствовать вашему корневому `.env`.

Для hot reload можно установить Air и запустить его из `backend/`:

```bash
go install github.com/air-verse/air@v1.61.1
air -c .air.toml
```

## API

Все защищённые маршруты ожидают заголовок:

```http
Authorization: Bearer <JWT_TOKEN>
```

| Метод | Маршрут | Авторизация | Назначение |
|---|---|---|---|
| `POST` | `/auth/register` | Нет | Регистрация пользователя |
| `POST` | `/auth/login` | Нет | Вход, получение JWT и пользователя |
| `GET` | `/auth/me` | Да | Получение текущего пользователя |
| `GET` | `/tracks/create` | Да | Получение треков текущего пользователя |
| `POST` | `/tracks/` | Да | Загрузка аудио и необязательной обложки |
| `GET` | `/tracks/{id}/stream-url` | Да | Получение временного URL аудиофайла |

Пример регистрации:

```bash
curl -X POST http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Alex",
    "email": "alex@example.com",
    "avatar_url": "https://example.com/avatar.jpg",
    "password": "change-me"
  }'
```

Пример входа:

```bash
curl -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "alex@example.com",
    "password": "change-me"
  }'
```

Пример загрузки трека:

```bash
curl -X POST http://localhost:8080/tracks/ \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -F 'file=@/absolute/path/to/song.mp3' \
  -F 'cover=@/absolute/path/to/cover.jpg'
```

Поле `cover` можно не передавать.

## Миграции

Миграции находятся в `backend/migrations/` и автоматически применяются сервисом `migrate` при запуске Docker Compose.

Применить миграции вручную через Compose:

```bash
docker compose run --rm migrate
```

Проверить таблицы PostgreSQL:

```bash
docker compose exec postgres \
  psql -U echora -d echora -c '\dt'
```

Если вы изменили `DB_USER` или `DB_NAME`, подставьте свои значения в команду.

## Проверка кода

Backend:

```bash
cd backend
go test ./...
```

Frontend production build:

```bash
cd frontend
npm install
npm run build
```

Собранные frontend-файлы появятся в `frontend/dist/`.

## Особенности текущей версии

- JWT действует 24 часа.
- Подписанные ссылки MinIO действуют один час.
- CORS настроен только для `http://localhost:5173`.
- Аватар задаётся публичным URL, загрузка аватара в MinIO пока не реализована.
- Удаление и редактирование треков пока не реализованы.
- Redis пока не подключён к прикладной логике.
- Конфигурация ориентирована на локальную разработку, а не на production-развёртывание.

## Безопасность

Для публичного развёртывания потребуется как минимум:

- заменить все демонстрационные пароли и `JWT_SECRET`;
- ограничить CORS реальным доменом;
- включить HTTPS;
- настроить production-доступ к PostgreSQL и MinIO;
- добавить строгую серверную проверку размеров и форматов файлов;
- добавить rate limiting для авторизации и загрузок;
- не использовать MinIO root-пользователя как прикладную учётную запись;
- собирать frontend статически и запускать backend без dev hot reload.
