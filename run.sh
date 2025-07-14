#!/bin/bash

set -e

if [ -z "$TG_KEY" ] || [ -z "$DB_URL" ] || [ -z "$DB_KEY" ]; then
  echo "❌ Error: You must provide TG_KEY, DB_URL and DB_KEY"
  echo "Usage: TG_KEY=... DB_URL=... DB_KEY=... [HOST_NAME=localhost] [PORT=8000] ./run.sh"
  exit 1
fi

HOST_NAME=${HOST_NAME:-localhost}
PORT=${PORT:-80} #for http - 80, for https - 443

CONTAINER_PORT=8080

cat > config.yaml <<EOF
host_name: "$HOST_NAME"
port: "$PORT"
tg_key: "$TG_KEY"
db_url: "$DB_URL"
db_key: "$DB_KEY"
EOF

echo "✅ Generated config.yaml"

docker build -t url-shortener-bot -f builds/Dockerfile .

docker run --rm -p "$PORT":"$CONTAINER_PORT" url-shortener-bot