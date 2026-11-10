# TeleCorp Analytics

Менеджер корпоративных Telegram-аккаунтов для малого и среднего бизнеса: контроль переписок и активности сотрудников без ручного просмотра чатов.

## Структура

```
cmd/app/main.go               точка входа: читает конфиг, подключает БД и запускает сервер
docs/                         требования и техническое задание проекта
migrations/                   SQL-миграции схемы базы данных
internal/config/              настройки из переменных окружения
internal/domain/              сущности и ошибки предметной области
internal/handler/             HTTP-слой: роуты, разбор запросов, ответы
internal/usecase/             бизнес-логика и интерфейсы хранилищ
internal/storage/postgres/    реализация работы с базой данных (PostgreSQL + sqlx)
```

## Как запустить

Понадобятся [Go 1.26+](https://go.dev/dl/) и [Docker](https://www.docker.com/products/docker-desktop/) для запуска в контейнере.

### Локально

```bash
go run ./cmd/app
```

### В Docker

```bash
docker build -t go-server:1 .
docker run --rm -p 8080:8080 go-server:1
```

Готовый образ без сборки:

```bash
docker run --rm -p 8080:8080 ghcr.io/kraconst/go-server:1
```

### Настройки

| Переменная    | По умолчанию       | Назначение                    |
|---------------|--------------------|-------------------------------|
| `HTTP_ADDR`   | `:8080`            | адрес, который слушает сервер |
| `DB_HOST`     | `localhost`        | хост PostgreSQL               |
| `DB_PORT`     | `5432`             | порт PostgreSQL               |
| `DB_USER`     | `telecorp`         | пользователь базы данных      |
| `DB_PASSWORD` | `telecorp_secret`  | пароль базы данных            |
| `DB_NAME`     | `telecorp_db`      | имя базы данных               |
| `DB_SSLMODE`  | `disable`          | режим SSL для подключения к БД|

### База данных (PostgreSQL)

Запуск базы через Docker Compose:

```bash
docker compose up -d postgres
```

## API

| Метод | Путь             | Ответ                                         |
|-------|------------------|-----------------------------------------------|
| GET   | `/health`        | `200` — сервис жив                            |
| GET   | `/accounts`      | `200` — список аккаунтов                      |
| GET   | `/accounts/:id`  | `200` — аккаунт, `404` — нет такого, `400` — id не число |
| POST  | `/accounts`      | `201` — аккаунт создан, `400` — не хватает полей |

Пример запроса на создание:

```bash
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"phone":"+79990000000","username":"@new","employee_name":"Ира"}'
```

В PowerShell вместо `curl` пишите `curl.exe`, а для POST удобнее `Invoke-RestMethod`:

```powershell
Invoke-RestMethod -Method Post -Uri http://localhost:8080/accounts -ContentType "application/json" -Body '{"phone":"+79990000000","username":"@new","employee_name":"Ира"}'
```

## Тестовое подключение Telegram

Отдельная команда `cmd/telegram-dev` подключает существующий аккаунт через
MTProto (`gotd/td`). HTTP-сервер и база данных для неё не нужны.
Ключи приложения получают в [API development tools](https://my.telegram.org).

В интерактивном PowerShell задайте переменные окружения:

```powershell
$env:TELEGRAM_API_ID = "123456"
$env:TELEGRAM_API_HASH = "ваш_api_hash"
$env:TELEGRAM_PHONE = "+79991234567"
$env:TELEGRAM_SESSION_PATH = "./var/telegram/account-a.session"
go run ./cmd/telegram-dev -limit 20
```

Пример переменных есть в `.env.example`; команда не загружает `.env` автоматически.
`TELEGRAM_SESSION_PATH` необязательна: по умолчанию используется
`./var/telegram/<номер без плюса>.session`. Для каждого аккаунта нужен свой файл.
`-limit` принимает значения от 1 до 100 (по умолчанию 20).

## Как вносить изменения

Подробная инструкция по архитектуре, пошаговый чек-лист создания сущностей и распределение задач между участниками команды описаны в [CONTRIBUTING.md](CONTRIBUTING.md).

1. Заберите свежий `main`: `git switch main && git pull`.
2. Создайте ветку под задачу: `git switch -c feature/<что-делаете>`.
3. Коммитьте и пушьте ветку: `git push -u origin feature/<что-делаете>`.
4. Откройте Pull Request в `main` и попросите кого-нибудь из команды посмотреть.

Перед коммитом прогоните `gofmt -w .` и `go vet ./...`.
