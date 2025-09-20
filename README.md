# 🔗 URL Shortener Telegram Bot

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
![Go Version](https://img.shields.io/badge/Language-Go-blue)
[![Go Report Card](https://goreportcard.com/badge/github.com/alexcfv/go-pcaplite)](https://goreportcard.com/report/github.com/alexcfv/url-shorter-bot)
[![Supabase](https://img.shields.io/badge/Supabase-3ECF8E?logo=supabase&logoColor=white&style=for-the-badge)](https://supabase.com/)
[![Telegram Bot](https://img.shields.io/badge/Telegram-Bot-blue.svg)](https://core.telegram.org/bots)
![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macos%20%7C%20windows-blue)

A simple and fast **Telegram bot for shortening URLs**, written in Go.  
You send a link — the bot returns a short one. That’s it.

#### 📝 Author's Note

This is a learning project demonstrating interaction with the Telegram Bot API, Supabase, caching, HTTP, and context.  
Service architecture built using the Dependency Injection (DI) pattern. 

---

## 📲 How It Works

Before this you must run the service according to the instructions below

1. Open the Telegram bot.
2. Tap **“Shorten URL”**.
3. Send any link (e.g., `https://example.com/some/very/long/url`).
4. The bot will return a shortened version like "your_protocol://your_host_name/hash" (e.g., `http://short.ly/128429213`).
5. Follow the link — and you’ll be redirected to the original site.

---

## ⚙️ Database Pre-Setup


Create database in Supabase without any tables,
In table editor select schema public.
Go to the SQL Editor and paste the following code:

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

create or replace function execute_sql(sql text)
returns void
language plpgsql
as $$
begin
    execute sql;
end;
$$;

grant execute on function execute_sql(text) to service_role;
grant execute on function table_exists(text) to service_role;
grant create, usage on schema public to service_role;
```

It allows table creation via RPC: add functions for execute sql and check table exists.

## ⚙️ Project Pre-Setup

Go to the root folder and paste the following code:

`/url-shorter-bot`
```bash
go mod tidy
```

Download project dependency

---

## 🚀 Launch

The bot can be launched in two ways: manually via YAML or using Docker.

### 🔧 Option 1: `config.yaml` (Manual)

Create a `config.yaml` file in the root of the project:

```yaml
host_name: "YOUR_HOST_NAME"         # Default: "localhost"
port: "YOUR_PORT"                   # Default: "80" ("80" for HTTP or "443" for HTTPS)
tg_key: "YOUR_TELEGRAM_TOKEN"
db_url: "YOUR_SUPABASE_URL"
db_key: "YOUR_SUPABASE_SERVICE_ROLE_API_KEY"
```

You can get the service role key like this in Supabase project: Project settings -> Api Keys -> Api keys -> sevice_role

To get a telegram bot token, you must first create it.
You can do this [here](https://t.me/BotFather)

Then run it manually:

`/url-shorter-bot`
```bash
go run src/main.go
```

### 🐳 Option 2: Run via Docker

You can also run it using the run.sh script by passing variables:

`/url-shorter-bot`
```bash
TG_KEY=your_telegram_token \
DB_URL=https://your-project.supabase.co \
DB_KEY=your_supabase_service_role_key \
./run.sh
```
If host_name and port are not specified, default values are used:

host_name: localhost

port: 80 (HTTP protocol)

`/url-shorter-bot`
```bash
TG_KEY=your_telegram_token \
DB_URL=https://your-project.supabase.co \
DB_KEY=your_supabase_service_role_key \
HOST_NAME=your_host_name \
PORT=443 \
./run.sh
```

In this case, port 443 is used, which by default is open for HTTPS requests — so the program assumes you're using HTTPS.

### 📌 To work correctly, you need a real domain listed in the HostWhitelist and a DNS A record pointing to its IP address.

### 📌 You can also specify your custom port, but in response you will receive not http://your_domain/short_url, but http://your_domain:your_port/short_url

### 📌 However, it still works great with the default values in the YAML config using the HTTP protocol.

---

### 🗄️ Database Tables

Tables are created automatically with RLS.

| Table        | Purpose                             |
| ------------ | ----------------------------------- |
| `users`      | List of Telegram users              |
| `urls`       | Stores original and shortened links |
| `log_error`  | Error log                           |
| `log_action` | User action log                     |

---

### 🧪 Testing

Only unit tests are written. No integration or end-to-end tests available yet.

Run tests:

`/url-shorter-bot`
```bash
go test run ./...
```

---

### 📝 License

This project is licensed under the MIT License.

You are free to use, modify, and distribute it.
