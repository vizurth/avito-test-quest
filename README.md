# PR Reviewer Assignment Service

REST API сервис для управления командами, пользователями и автоматического распределения ревьюверов на Pull Request'ы.

## Особенности

- **Автоматическое назначение ревьюверов** — при создании PR автоматически назначаются до 2 активных ревьюверов из команды автора
- **Управление командами и пользователями** — создание команд, добавление участников, управление статусом активности
- **Управление PR** — создание, мерж PR, переназначение ревьюверов
- **Получение PR для ревьювера** — просмотр всех PR, где пользователь назначен ревьювером
- **Статистика** — получение статистики по количеству назначений ревьюверов и PR
- **Структурированное логирование** — используется `zap` для логирования операций

## Технологический стек

- **Язык**: Go 1.x
- **Web Framework**: [Gin](https://github.com/gin-gonic/gin)
- **БД**: PostgreSQL
- **Query Builder**: [Squirrel](https://github.com/Masterminds/squirrel)
- **Миграции**: [golang-migrate](https://github.com/golang-migrate/migrate)
- **Логирование**: [zap](https://github.com/uber-go/zap)
- **Драйвер БД**: [pgx](https://github.com/jackc/pgx)

## Архитектура

```
internal/
├── app/          # Инициализация приложения
├── config/       # Конфигурация
├── handler/      # HTTP обработчики (Gin)
├── logger/       # Логирование (zap)
├── models/       # Доменные модели (DTO)
├── postgres/     # Подключение к БД и миграции
├── repository/   # Уровень доступа к данным (CRUD операции)
└── service/      # Бизнес-логика

migrations/      # SQL миграции (с usar golang-migrate)
cmd/             # Точка входа приложения
configs/         # Файлы конфигурации
build/           # Docker файлы
```

## Установка и запуск

### 1. Клонирование репозитория

```bash
git clone https://github.com/vizurth/avito-test-quest.git
cd avito-test-quest
```

### 2. Запуск сервиса

Используйте Makefile команды для управления сервисом:

```bash
# Запустить сервис в фоне с пересборкой образа
make up

# Остановить сервис
make down

# Перезагрузить сервис (down + up)
make restart

# Пересобрать Docker образ
make build_and
```

Сервис запустится на `http://localhost:8080`.

Все миграции БД применяются автоматически при запуске приложения через `golang-migrate`.


## API Endpoints

### Teams

#### `POST /team/add` — Создать команду с участниками

Создает новую команду и одновременно создает или обновляет участников. Если команда с таким именем уже существует, возвращает ошибку `TEAM_EXISTS`.

**Request:**

```json
{
	"team_name": "backend",
	"members": [
		{ "user_id": "u1", "username": "Alice", "is_active": true },
		{ "user_id": "u2", "username": "Bob", "is_active": true }
	]
}
```

**Response:** 201 Created

```json
{
	"team": {
		"team_name": "backend",
		"members": [
			{ "user_id": "u1", "username": "Alice", "is_active": true },
			{ "user_id": "u2", "username": "Bob", "is_active": true }
		]
	}
}
```

#### `GET /team/get?team_name=<name>` — Получить команду по имени

Возвращает информацию о команде со списком всех её участников. Если команда не найдена, возвращает `NOT_FOUND`.

**Query Parameters:**

- `team_name` (обязательный) — название команды

**Response:** 200 OK

```json
{
	"team_name": "backend",
	"members": [
		{ "user_id": "u1", "username": "Alice", "is_active": true },
		{ "user_id": "u2", "username": "Bob", "is_active": false }
	]
}
```

### Users

#### `POST /users/setIsActive` — Установить статус активности пользователя

Изменяет статус активности пользователя (активен/неактивен). Только активные пользователи могут быть назначены ревьюверами на PR. Требует Admin токен.

**Request:**

```json
{
	"user_id": "u2",
	"is_active": false
}
```

**Response:** 200 OK

```json
{
	"user": {
		"user_id": "u2",
		"username": "Bob",
		"team_name": "backend",
		"is_active": false
	}
}
```

#### `GET /users/getReview?user_id=<id>` — Получить PR, где пользователь назначен ревьювером

Возвращает список всех Pull Request'ов, на которых пользователь назначен ревьювером. Если пользователь не найден, возвращает `NOT_FOUND`.

**Query Parameters:**

- `user_id` (обязательный) — идентификатор пользователя

**Response:** 200 OK

```json
{
	"user_id": "u2",
	"pull_requests": [
		{
			"pull_request_id": "pr-1001",
			"pull_request_name": "Add search functionality",
			"author_id": "u1",
			"status": "OPEN"
		},
		{
			"pull_request_id": "pr-1002",
			"pull_request_name": "Fix bug in login",
			"author_id": "u3",
			"status": "MERGED"
		}
	]
}
```

### Pull Requests

#### `POST /pullRequest/create` — Создать PR и автоматически назначить ревьюверов

Создает новый Pull Request и автоматически назначает до 2 активных ревьюверов из команды автора (исключая самого автора). Если PR с таким ID уже существует, возвращает `PR_EXISTS`. Если автор или его команда не найдены, возвращает `NOT_FOUND`. Требует Admin токен.

**Request:**

```json
{
	"pull_request_id": "pr-1001",
	"pull_request_name": "Add search functionality",
	"author_id": "u1"
}
```

**Response:** 201 Created

```json
{
	"pr": {
		"pull_request_id": "pr-1001",
		"pull_request_name": "Add search functionality",
		"author_id": "u1",
		"status": "OPEN",
		"assigned_reviewers": ["u2", "u3"],
		"createdAt": "2025-11-14T10:30:00Z",
		"mergedAt": null
	}
}
```

#### `POST /pullRequest/merge` — Пометить PR как merged

Изменяет статус PR на `MERGED` и устанавливает время мержа. Операция **идемпотентна** — если PR уже в статусе MERGED, просто возвращает текущее состояние. Если PR не найден, возвращает `NOT_FOUND`. Требует Admin токен.

**Request:**

```json
{
	"pull_request_id": "pr-1001"
}
```

**Response:** 200 OK

```json
{
	"pr": {
		"pull_request_id": "pr-1001",
		"pull_request_name": "Add search functionality",
		"author_id": "u1",
		"status": "MERGED",
		"assigned_reviewers": ["u2", "u3"],
		"createdAt": "2025-11-14T10:30:00Z",
		"mergedAt": "2025-11-14T10:45:00Z"
	}
}
```

#### `POST /pullRequest/reassign` — Переназначить ревьювера

Заменяет одного ревьювера на другого из той же команды. Выбор нового ревьювера происходит автоматически из активных членов команды (исключая автора PR и текущих ревьюверов).

Возможные ошибки:

- `NOT_FOUND` — PR или пользователь не найдены
- `PR_MERGED` — PR уже в статусе MERGED
- `NOT_ASSIGNED` — указанный ревьювер не назначен на этот PR
- `NO_CANDIDATE` — нет активных кандидатов для замены

Требует Admin токен.

**Request:**

```json
{
	"pull_request_id": "pr-1001",
	"old_user_id": "u2"
}
```

**Response:** 200 OK

```json
{
	"pr": {
		"pull_request_id": "pr-1001",
		"pull_request_name": "Add search functionality",
		"author_id": "u1",
		"status": "OPEN",
		"assigned_reviewers": ["u3", "u4"],
		"createdAt": "2025-11-14T10:30:00Z",
		"mergedAt": null
	},
	"replaced_by": "u4"
}
```

### Health

#### `GET /health` — Проверка состояния сервиса

Простая проверка здоровья сервиса. Возвращает статус "ok" если сервис работает.

**Response:** 200 OK

```json
{
	"status": "ok"
}
```

### Statistics

#### `GET /stats` — Получить статистику

Возвращает статистику по ревьюверам и Pull Request'ам. Статистика включает:

- **Reviewer Stats** — количество назначений (сколько раз каждый пользователь назначен ревьювером) на PR, отсортировано по убыванию
- **PR Stats** — количество назначенных ревьюверов на каждый PR, отсортировано по убыванию

**Response:** 200 OK

```json
{
	"reviewer_stats": [
		{
			"user_id": "u2",
			"username": "Bob",
			"assigned_count": 5
		},
		{
			"user_id": "u3",
			"username": "Charlie",
			"assigned_count": 3
		},
		{
			"user_id": "u1",
			"username": "Alice",
			"assigned_count": 0
		}
	],
	"pr_stats": [
		{
			"pull_request_id": "pr-1001",
			"pull_request_name": "Add search functionality",
			"author_id": "u1",
			"status": "OPEN",
			"reviewer_count": 2
		},
		{
			"pull_request_id": "pr-1002",
			"pull_request_name": "Fix bug in login",
			"author_id": "u3",
			"status": "MERGED",
			"reviewer_count": 2
		},
		{
			"pull_request_id": "pr-1003",
			"pull_request_name": "Update dependencies",
			"author_id": "u2",
			"status": "OPEN",
			"reviewer_count": 1
		}
	]
}
```

## Тестирование API

### Использование Postman

Проект включает готовую коллекцию запросов в файле `postman.json`. Для импорта:

1. Откройте Postman
2. Нажмите **Import** → **File** → выберите `postman.json`
3. Скопируется вся коллекция с предзаполненными запросами

**Переменные окружения:**

- `base_url` — адрес сервиса (по умолчанию: `http://localhost:8080`)
- `admin_token` — токен администратора для защищённых эндпоинтов

Коллекция содержит примеры:

- Создание команд (Backend, Frontend)
- Получение информации о командах
- Управление статусом пользователей
- Создание PR и автоматическое назначение ревьюверов
- Мерж PR (включая идемпотентные операции)
- Переназначение ревьюверов
- Тестирование ошибок (404, 409 и т.д.)

## Разработка

### Линтинг кода

Проект использует `golangci-lint` для проверки качества кода. Конфигурация находится в файле `.golangci.yml`.

**Установка golangci-lint:**

```bash
make install-golangci-lint
```

**Проверка кода:**

```bash
# Запустить линтер
make lint

# Исправить автоматически исправляемые ошибки
make lint-fix

# Отформатировать код
make format
```

**Тестирование кода**
```bash
# Запустить интеграционные-тесты
make test-integration
```

## База данных

### Таблицы

- **teams** — команды
- **users** — пользователи (связаны с командой)
- **pull_requests** — pull requests
- **pr_reviewers** — связь PR и ревьюверов

### Миграции

Все миграции находятся в папке `migrations/` и автоматически применяются при запуске приложения через `golang-migrate`. Никаких дополнительных действий не требуется.

## Коды ошибок

| Код            | Описание                                   |
| -------------- | ------------------------------------------ |
| `TEAM_EXISTS`  | Команда с таким названием уже существует   |
| `PR_EXISTS`    | PR с таким ID уже существует               |
| `PR_MERGED`    | Нельзя изменить PR, который уже merged     |
| `NOT_ASSIGNED` | Ревьювер не назначен на этот PR            |
| `NO_CANDIDATE` | Нет активных кандидатов для переназначения |
| `NOT_FOUND`    | Ресурс не найден                           |

## Логирование

Приложение использует `zap` для структурированного логирования. Логи выводятся в формате JSON в production режиме.

## Структура проекта

```
avito-test-quest/
├── cmd/
│   └── main.go              # точка входа
├── internal/
│   ├── app/
│   │   └── app.go           # инициализация приложения
│   ├── config/
│   │   └── config.go        # конфигурация
│   ├── handler/
│   │   ├── handler.go       # HTTP обработчики
│   │   └── interface.go     # интерфейсы
│   ├── logger/
│   │   └── logger.go        # логирование
│   ├── models/
│   │   └── models.go        # доменные модели
│   ├── postgres/
│   │   └── postgres.go      # подключение БД
│   ├── repository/
│   │   ├── repository.go    # CRUD операции
│   │   └── interface.go     # интерфейсы
│   └── service/
│       ├── service.go       # бизнес-логика
│       └── interface.go     # мнтерфейсы
├── migrations/              # SQL миграции
├── build/
│   └── docker/              # docker файлы
├── configs/
│   └── config.yaml          # конфигурация
├── Makefile                 # команды сборки
├── go.mod                   # модуль Go
├── go.sum                   # зависимости
└── README.md                # этот файл
```
