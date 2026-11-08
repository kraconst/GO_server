# Руководство разработчика: добавление сущностей

Этот документ описывает, как устроена архитектура проекта, как распределены задачи между участниками команды и по каким шагам реализовывать свою сущность.

---

## 1. Архитектура проекта

В проекте используется трёхслойная чистая архитектура (Clean Architecture):

```text
HTTP Handler  ──►  Usecase (Бизнес-логика)  ──►  Storage (PostgreSQL)
(Gin роуты)        (Валидация, правила)          (SQL-запросы sqlx)
       ▲                    ▲                             ▲
       └────────────────────┴─────────────────────────────┘
                                  │
                                domain
                    (структуры и ошибки без зависимостей)
```

1. **`internal/domain/`**: Чистые структуры Go (`Account`, `Employee`, `Chat` и т.д.) и ошибки (`internal/domain/errors.go`). Слой ни от кого не зависит.
2. **`internal/usecase/interfaces.go`**: Контракты интерфейсов репозиториев (хранилищ).
3. **`internal/storage/postgres/`**: Реализация интерфейсов хранилищ через библиотеку `sqlx` и чистый SQL.
4. **`internal/usecase/`**: Бизнес-правила (проверка входных данных, привязка к компании).
5. **`internal/handler/`**: Разбор HTTP-запросов (Gin), вызов usecase и возврат JSON.
6. **`cmd/app/main.go`**: Единая точка сборки: читает конфиг, подключает БД, связывает `storage -> usecase -> handler` и стартует сервер.

> **Эталоны в коде**: в качестве шпаргалки используйте уже готовые сущности:
> * [`Account`](internal/storage/postgres/account.go) — сложная сущность с внешними связями и историей.
> * [`Employee`](internal/storage/postgres/employee.go) — классический справочник/CRUD.

---

## 2. Пошаговый чек-лист: как реализовать свою сущность (5 шагов)

Когда вы берёте сущность (например, `Contact` или `Message`), выполните следующие шаги:

### Шаг 1. Ветка в Git
```bash
git switch main
git pull
git switch -c feature/<имя-сущности>
```

### Шаг 2. Миграция базы данных (DDL)
В папке `migrations/` создайте следующий порядковый номер миграции:
* `migrations/00000X_create_<имя>.up.sql` (создание таблицы и индексов)
* `migrations/00000X_create_<имя>.down.sql` (удаление таблицы)

Пример:
```sql
CREATE TABLE IF NOT EXISTS contacts (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    telegram_user_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL DEFAULT '',
    UNIQUE (company_id, telegram_user_id)
);
CREATE INDEX IF NOT EXISTS idx_contacts_company_id ON contacts(company_id);
```

### Шаг 3. Добавление `db` тегов в Domain
Откройте файл в `internal/domain/<имя>.go` и добавьте struct-теги `db:"..."`, совпадающие с колонками в таблице:
```go
type Contact struct {
    ID             int64  `json:"id" db:"id"`
    CompanyID      int64  `json:"company_id" db:"company_id"`
    TelegramUserID int64  `json:"telegram_user_id" db:"telegram_user_id"`
    Name           string `json:"name" db:"name"`
    Username       string `json:"username" db:"username"`
}
```

### Шаг 4. Интерфейс в Usecase
В файле `internal/usecase/interfaces.go` добавьте контракт:
```go
type ContactRepository interface {
    Create(ctx context.Context, c domain.Contact) (int64, error)
    GetByID(ctx context.Context, id int64) (*domain.Contact, error)
    GetByTelegramID(ctx context.Context, companyID, tgUserID int64) (*domain.Contact, error)
    ListByCompany(ctx context.Context, companyID int64) ([]domain.Contact, error)
}
```

### Шаг 5. Реализация Storage и Usecase
1. Создайте `internal/storage/postgres/<имя>.go`: реализуйте методы интерфейса, используя `s.db.QueryRowContext`, `s.db.GetContext` или `s.db.SelectContext` (как в [employee.go](internal/storage/postgres/employee.go)).
2. Создайте `internal/usecase/<имя>.go`: реализуйте структуру `*<Name>Usecase`, добавьте валидацию (например, имя не должно быть пустым, ID > 0).

### Шаг 6. Роуты в Handler и подключение в Main
1. Добавьте методы обработки HTTP-запросов в `internal/handler/`.
2. В конструктор `handler.New(...)` передайте новый usecase.
3. В `cmd/app/main.go` проинициализируйте репозиторий и usecase и передайте в handler.
4. Проверьте форматирование и сборку:
   ```bash
   gofmt -w .
   go vet ./...
   go build ./...
   ```

---

## 3. Распределение задач между участниками

Ниже рекомендуемое разбиение оставшихся сущностей на независимые задачи:

### Задача 1. Клиенты и Диалоги (`Contact` + `Chat`)
* **Файлы домена**: `internal/domain/contact.go`, `internal/domain/chat.go`
* **Что сделать**:
  * Миграция создания таблиц `contacts` и `chats`.
  * Репозитории `ContactStorage` и `ChatStorage`:
    - Поиск/создание контакта по Telegram User ID;
    - Список диалогов компании с фильтрами по статусу и ответственному;
    - Ручная смена статуса диалога (`status_source = manual`).
  * Роуты: `GET /chats`, `GET /chats/:id`, `POST /chats/:id/status`.

### Задача 2. Сообщения переписки (`Message`)
* **Файлы домена**: `internal/domain/message.go`
* **Что сделать**:
  * Миграция создания таблицы `messages` с составным уникальным индексом `(chat_id, telegram_message_id)`.
  * Репозиторий `MessageStorage`:
    - Вставка сообщения с защитой от дубликатов (`ON CONFLICT (chat_id, telegram_message_id) DO NOTHING`);
    - Получение истории сообщений диалога с пагинацией (LIMIT / OFFSET) и сортировкой по `sent_at ASC`.
  * Роуты: `GET /chats/:id/messages`.

### Задача 3. Сигналы и Уведомления (`Alert`)
* **Файлы домена**: `internal/domain/alert.go`
* **Что сделать**:
  * Миграция создания таблицы `alerts` (типы: `no_reply`, `complaint`, `refusal`, `negative`, `urgent`).
  * Репозиторий `AlertStorage`:
    - Создание алерта;
    - Список активных алертов компании (`resolved_at IS NULL`);
    - Закрытие алерта (простановка `resolved_at = NOW()`).
  * Роуты: `GET /alerts`, `POST /alerts/:id/resolve`.

### Задача 4. Дневная аналитика (`AccountDailyStats`)
* **Файлы домена**: `internal/domain/account_daily_stats.go`
* **Что сделать**:
  * Миграция таблицы `account_daily_stats` (составной первичный ключ `account_id + day`).
  * Репозиторий `AccountDailyStatsStorage`:
    - Инкремент счётчиков (`messages_in`, `messages_out`, `new_contacts`);
    - Выборка метрик за диапазон дат для построения дашборда.
  * Роуты: `GET /analytics/dashboard`, `GET /analytics/accounts`.

### Задача 5. AI-Саммари диалога (`ChatSummary`)
* **Файлы домена**: `internal/domain/chat_summary.go`
* **Что сделать**:
  * Миграция таблицы `chat_summaries` (связь 1:1 к `chat_id`).
  * Репозиторий `ChatSummaryStorage`:
    - Сохранение/перезапись саммари (`UPSERT`);
    - Чтение текущего саммари для карточки чата.
  * Включение объекта саммари в ответ `GET /chats/:id`.

---

## 4. Локальный запуск БД для разработки

База данных запускается одной командой:
```bash
docker compose up -d postgres
```
Все таблицы первой миграции создаются автоматически при первом старте контейнера.
Параметры подключения по умолчанию: порт `5432`, пользователь `telecorp`, пароль `telecorp_secret`, база `telecorp_db`.
