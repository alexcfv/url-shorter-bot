[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
![Go Version](https://img.shields.io/badge/Language-Go-blue)

#### 📝 Коменнтарий от автора

Это учебный проект, демонстрирующий работу с telegram bot api, supabase, cache, http и context.
В нем присутствуют интересные реализации такие как передача telegram_id в context через middleware и последующее взаимодействие с ним,
простой request limiter и самое главное реализация сервиса через паттерн Dependence Injection (DI)

# 🔗 URL Shortener Telegram Bot

Простой и быстрый **Telegram-бот для сокращения ссылок**, написанный на Go.  
Ты отправляешь ссылку — бот возвращает короткую. Всё просто.

---

## 📲 Как работает

1. Открой Telegram-бота.
2. Нажми **«Сократить ссылку»**.
3. Отправь любую ссылку (например: `https://example.com/some/very/long/url`).
4. Бот вернёт короткий вариант в таком формате "твой_протокол://твое_имя_хоста/хеш_ссылки"(например: `http://short.ly/128429213`).
5. Перейди по ней — и окажешься на оригинальном сайте.

---

## ⚙️ Преднастройка базы данных

Перейди в SQL Editor и вставь туда этот код

```bash
create or replace function table_exists(tbl text)
returns boolean
language plpgsql
as $$
begin
  return exists (
    select from pg_tables
    where tablename = tbl
  );
end;
$$;

grant execute on function table_exists(text) to service_role;
```

Он разрешает создание таблиц через RPC

---

## 🚀 Запуск

Бот может быть запущен двумя способами: через YAML вручную или через Docker.

### 🔧 Вариант 1: `config.yaml` (Вручную)

Создай файл `config.yaml` в корне проекта:

```yaml
host_name: "YOUR_HOST_NAME"         # По умолчанию: "localhost"
port: "YOUR_PORT"                   # Одно из двух: "80" (для HTTP) или "443" (для HTTPS)
tg_key: "YOUR_TELEGRAM_TOKEN"
db_url: "YOUR_SUPABASE_URL"
db_key: "YOUR_SUPABASE_API_KEY"
```
И запусти вручную

`/url-shorter-bot`
```bash
go run src/main
```

### 🐳 Вариант 2: запуск через Docker

Можно также запустить через скрипт run.sh, передав переменные окружения:

`/url-shorter-bot`
```bash
TG_KEY=your_telegram_token \
DB_URL=https://your-project.supabase.co \
DB_KEY=your_supabase_key \
./run.sh
```
Если не указаны host_name и port, используются значения по умолчанию:

host_name: localhost

port: 80 (протокол HTTP)

`/url-shorter-bot`
```bash
TG_KEY=your_telegram_token \
DB_URL=https://your-project.supabase.co \
DB_KEY=your_supabase_key \
HOST_NAME=your_host_name \
PORT=443 \
./run.sh
```

В этом случае используется порт 443 который по умолчанию открыт системой для HTTPS запросов -> исходя из этого программа решает что ты используешь HTTPS

---

### 🗄️ Таблицы базы данных

Таблицы создаются автоматически без RLC

| Таблица      | Назначение                                 |
| ------------ | ------------------------------------------ |
| `users`      | Список Telegram-пользователей              |
| `urls`       | Хранение оригинальных и сокращённых ссылок |
| `log_error`  | Журнал ошибок                              |
| `log_action` | Журнал действий пользователей              |

---

### 🧪 Тестирование

Написанны только юнит тесты, интеграционных и end-to-end тестов нет

Запуск тестов:

`/url-shorter-bot`
```bash
go test run ./...
```

---

### 📝 Лицензия

Проект распространяется под лицензией MIT. Ты можешь свободно использовать, изменять и распространять его.
