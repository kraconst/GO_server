# TeleCorp Analytics

Менеджер корпоративных Telegram-аккаунтов для малого и среднего бизнеса: контроль переписок и активности сотрудников без ручного просмотра чатов.

## Модель домена

```
Company (компания-клиент)
 └── Account (Telegram-аккаунт)
      ├── закреплён за Employee (сотрудник)
      └── Chat (переписка с клиентом)
           └── Message (сообщение)
```

## Структура

```
cmd/app/main.go               точка входа: читает конфиг и запускает сервер
internal/config/              настройки из переменных окружения
internal/domain/              сущности (Account)
internal/handler/             HTTP-слой: роуты, разбор запросов, ответы
internal/usecase/             правила продукта (пока пусто)
internal/storage/             работа с базой (пока пусто)
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

| Переменная  | По умолчанию | Назначение              |
|-------------|--------------|-------------------------|
| `HTTP_ADDR` | `:8080`      | адрес, который слушает сервер |

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

## Как вносить изменения

1. Заберите свежий `main`: `git switch main && git pull`.
2. Создайте ветку под задачу: `git switch -c feature/<что-делаете>`.
3. Коммитьте и пушьте ветку: `git push -u origin feature/<что-делаете>`.
4. Откройте Pull Request в `main` и попросите кого-нибудь из команды посмотреть.

Перед коммитом прогоните `gofmt -w .` и `go vet ./...`.
