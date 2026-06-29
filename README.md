# Telegram-приложение для цифрового анализа характеристик растений и семян

**Автор:** Пархоменко Кирилл Александрович

---

## Описание проекта

Распределённая информационная система с доступом через **Telegram Mini Apps/Web App** и веб-интерфейс.

**Цель:** автоматизация цифрового фенотипирования семян сельскохозяйственных культур. Система заменяет ручную оценку (от 30 минут на пробу) экспресс-анализом по фотографии со смартфона (до 3 минут на типовую пробу из ~227 объектов), сокращая трудозатраты при сохранении необходимой точности.

**Основные возможности:**

- загрузка фото и постановка заявок на анализ;
- асинхронная обработка изображений и извлечение морфометрических признаков;
- классификация объектов по настраиваемым правилам;
- генерация PDF/CSV-отчётов.

> **Примечание.** Исходный код обработчиков изображений (Workers) в репозиторий не предоставляется — эти модули автором не разрабатывались. В репозитории представлена микросервисная платформа (оркестрация заявок, API, очереди, хранение, отчёты, клиентские приложения); интеграция с внешним CV-обработчиком выполняется по согласованному контракту обмена через RabbitMQ.

---

## Стек технологий

| Слой | Технологии |
|------|------------|
| Backend | Go, [Fiber v3](https://gofiber.io/) |
| Frontend | TypeScript, React, Vite, TanStack Router, Tailwind CSS |
| Базы данных | PostgreSQL, Redis |
| Очереди | RabbitMQ |
| Хранилище | MinIO (S3-совместимое) |
| Отчёты | chromedp (HTML → PDF) |
| Инфраструктура | Docker, Docker Compose |
| Наблюдаемость | OpenTelemetry, zerolog |

---

## Архитектура

Событийно-ориентированная микросервисная архитектура: интерактивные API отделены от тяжёлых вычислительных задач компьютерного зрения.

![Архитектура системы](docs/arch.png)

### Микросервисы

| Сервис | Назначение |
|--------|------------|
| **auth-service** | Аутентификация через `initData` Telegram Mini App, JWT (RS256), RBAC |
| **analysis-service** | Оркестрация заявок, жизненный цикл анализа, WebSocket-уведомления, Transactional Outbox |
| **classification-service** | Иерархические правила классификации семян |
| **correlation-service** | Расчёт корреляций объектов по атрибутам |
| **reports-service** | Сбор статистик, генерация PDF/CSV из HTML-шаблонов |
| **photobox** | Backend клиентского веб-приложения |
| **photobox (frontend)** | Веб-интерфейс и Telegram Mini App |
| **admin-panel** | Панель администрирования AuthService |

---

## Структура репозитория

```text
kalibr/
├── services/
│   ├── auth-service/
│   ├── analysis-service/
│   ├── classification-service/
│   ├── correlation-service/
│   ├── reports-service/
│   └── photobox/
├── frontends/
│   ├── photobox/
│   └── admin-panel/
└── README.md
```

Каждый сервис — самостоятельный модуль со своим `Makefile`, `compose.yml` и примером переменных окружения.

---

## Быстрый старт

### Требования

- Docker и Docker Compose
- Go 1.25+ (backend)
- Bun или Node.js 20+ (frontends)
- Python 3.12+ и [uv](https://docs.astral.sh/uv/) (correlation-service)

### Запуск backend-сервиса

```bash
cd services/auth-service
make dev
```

Аналогично для остальных Go-сервисов. В режиме разработки (`DEV_MODE=true`) создаётся пользователь `dev` / `dev`.

### Запуск correlation-service

```bash
cd services/correlation-service
cp env.example .env
make install && make run
```

### Запуск photobox (frontend)

```bash
cd frontends/photobox
make dev
```

Откройте <http://localhost:5173/login> (логин `dev` / `dev`).

Локальная эмуляция Mini App: <http://localhost:5173/?mock=messenger>

### Запуск admin-panel

Локальная разработка (Vite, порт 3000):

```bash
cd frontends/admin-panel
bun install   # или npm install
bun run dev   # или npm run dev
```

Docker Compose (сборка и запуск контейнера):

```bash
cd frontends/admin-panel
cp compose.env.example compose.env
docker compose --env-file compose.env -f compose.yml up --build
```

### Работа через мессенджер

Для входа пользователей через Telegram или MAX Mini App сначала зарегистрируйте бота в **admin-panel**:

1. Запустите `auth-service` и `admin-panel`, войдите под учётной записью администратора.
2. Откройте раздел **Боты** и нажмите «Добавить бота».
3. Укажите:
   - **Платформу** — `Telegram` или `MAX`;
   - **Имя бота** — уникальное имя (используется при аутентификации через `initData`);
   - **Токен** — API-ключ бота из BotFather (Telegram) или панели MAX.
4. Настройте Mini App в мессенджере так, чтобы он открывал `frontends/photobox` с тем же именем бота.

Без зарегистрированного бота `auth-service` не сможет проверить подпись `initData`, и вход из мессенджера работать не будет.

---

## Тестирование

[Протокол интеграционных испытаний](docs/integration-test-protocol.pdf)

```bash
cd services/<имя-сервиса> && go test ./...
make test-integration
cd frontends/photobox && bun test:run
cd frontends/admin-panel && bun test
cd services/correlation-service && make test
```
