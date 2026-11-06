# ⚡ Distributed Task Scheduler & Worker Pool in Go

Образовательный проект асинхронной очереди и распределенного пула воркеров на чистом **Go (Golang)** и **PostgreSQL**. 

Проект разработан для глубокого изучения конкурентности в Go (горутины, каналы, контексты, синхронизация) и паттернов построения надежных фоновых очередей задач (Transactional Queue, Exponential Backoff, Graceful Shutdown).

---

## 🎯 Ключевые архитектурные решения и паттерны

### 1. Архитектура и слои (Clean Architecture)
- **`cmd/server`** — точка сборки приложения (Composition Root).
- **`internal/handler`** — HTTP-транспорт, DTO-контракты, валидация входящих запросов.
- **`internal/service`** — бизнес-логика задач, работа через интерфейсы.
- **`internal/repository`** — работа с PostgreSQL на чистом `database/sql` + `pgx`.
- **`internal/worker`** — ядро очереди: диспетчер, пул воркеров и реестр процессоров.
- **`internal/middleware`** — сквозные слои: `RequestID -> Logging -> Recover`.

### 2. Очередь на PostgreSQL (Без внешних брокеров)
- **Защита от Double Processing:** атомарный выбор и захват задач пачками через `WITH selected AS (...) UPDATE ... FOR UPDATE SKIP LOCKED`.
- **Частичный индекс (Partial Index):** `CREATE INDEX ... WHERE status = 'pending'` — индекс весит пару килобайт и ускоряет поиск по активным задачам в миллионной таблице.
- **Адаптивный опрос (Adaptive Polling):** использование `time.NewTimer` вместо жесткого тикера:
  - Пачка полная (10 из 10) $\to$ `timer.Reset(0)` (мгновенная выкачка без пауз).
  - Пачка пустая $\to$ `timer.Reset(pollInterval)` (экономный сон).
- **Восстановление зомби-задач (Startup Recovery):** при рестарте сервера зависшие задачи в статусе `running` автоматически возвращаются в `pending`.

### 3. Конкурентность и надежность воркеров
- **Backpressure (Защита ресурсов):** фиксированный пул горутин (`workerCount`), не позволяющий серверу уйти в OOM при наплыве задач.
- **Паттерн Стратегия (Processor Registry):** добавление новых типов задач (`email`, `image`, `report`) без изменения кода самого воркера.
- **Экспоненциальная задержка (Exponential Backoff):** побитовый расчет задержки повторов ($baseDelay \times 2^{attempts}$) через `1 << attempts`.
- **Честный Graceful Shutdown:** перехват сигналов ОС (`SIGTERM`, `SIGINT`), закрытие очередей и корректное ожидание завершения активных задач через `sync.WaitGroup`.

### 4. Комплексное тестирование (3 уровня)
- **Level 1 (Чистые функции):** Табличные тесты (`Table-Driven Tests`) на валидацию полезной нагрузки (`internal/model`).
- **Level 2 (Бизнес-логика):** Юнит-тесты сервисов с ручными моками интерфейсов (`internal/service`).
- **Level 3 (HTTP-слой):** Тестирование хендлеров в оперативной памяти без сети через `net/http/httptest` (`internal/handler`).

---

## 🛠 Стек технологий

- **Язык:** Go 1.22+
- **СУБД:** PostgreSQL 16
- **Драйвер БД:** `jackc/pgx/v5`
- **Миграции:** `golang-migrate`
- **Логирование:** структурированный `log/slog`
- **Контейнеризация:** Docker (Multi-stage build) + Docker Compose

---

## 🚀 Быстрый запуск (Docker Compose)

Проект полностью упакован и готов к запуску одной командой (база, автомиграции и сервис):

```bash
docker compose up -d --build
