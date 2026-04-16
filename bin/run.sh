#!/usr/bin/env sh

# Завершаем выполнение скрипта при любой ошибке, использовании неинициализированных переменных или ошибках в пайпах
set -euo pipefail

echo "[run.sh] Starting service"
echo "[run.sh] Running DB migrations"

# Запускаем миграции с помощью goose:
# - используем директорию с миграциями ./db/migrations
# - подключаемся к PostgreSQL через переменную окружения DATABASE_URL
# - применяем все доступные миграции вверх (up)
# goose -dir ./db/migrations postgres "${DATABASE_URL}" up

echo "[run.sh] Starting Go app"

# Передаём управление процессу Go-приложения (exec заменяет shell на процесс приложения)
# это важно для корректной обработки сигналов (SIGTERM, SIGINT) в контейнере
exec /app/bin/app
