# Указываем базовый образ Go (Alpine Linux) для этапа сборки backend
FROM golang:1.26-alpine AS backend-builder

# Устанавливаем git, необходимый для загрузки зависимостей из VCS
RUN apk add --no-cache git

# Задаём рабочую директорию внутри контейнера для сборки проекта
WORKDIR /build/code

# Копируем файлы модулей Go для кеширования зависимостей
COPY go.mod go.sum ./

# Скачиваем зависимости Go с использованием кеша модулей
RUN --mount=type=cache,target=/go/pkg/mod \
go mod download

# Устанавливаем утилиту goose для работы с миграциями БД
# RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Копируем весь исходный код проекта в контейнер
COPY . .

# Собираем Go-приложение с отключённым CGO для Linux amd64, используя кеш сборки
RUN --mount=type=cache,target=/root/.cache/go-build \
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /build/app .

# Используем минимальный образ Alpine для финального runtime-этапа
FROM alpine:3.22

# Устанавливаем рабочую директорию для запуска приложения
WORKDIR /app

# Копируем собранный бинарный файл приложения из этапа сборки
COPY --from=backend-builder /build/app /app/bin/app

# Копируем миграции базы данных в runtime-образ
# COPY --from=backend-builder build/code/db/migrations /app/db/migrations

# Копируем бинарник goose из builder-этапа в финальный образ
# COPY --from=backend-builder /go/bin/goose /usr/local/bin/goose

# Копируем скрипт запуска приложения
COPY bin/run.sh /app/bin/run.sh

# Делаем скрипт запуска исполняемым
RUN chmod +x /app/bin/run.sh

# Открываем порт 8080 для внешнего доступа к сервису
EXPOSE 8080

# Устанавливаем команду запуска контейнера
CMD ["/app/bin/run.sh"]
