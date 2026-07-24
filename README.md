# AutoMaster

**AutoMaster** — веб-приложение для учёта и организации работы автосервиса.  
Backend написан на Go, данные хранятся в PostgreSQL, клиентская часть реализована на HTML, CSS и JavaScript.

## Возможности

- регистрация, авторизация и редактирование профиля;
- управление клиентами, автомобилями, сотрудниками и услугами;
- учёт запчастей и расходных материалов;
- создание и изменение заказов;
- назначение сотрудников и используемых деталей на заказ.

## Технологии

- **Backend:** Go, `net/http`, REST API;
- **База данных:** PostgreSQL, `pgx`;
- **Frontend:** HTML, CSS, JavaScript;
- **Инфраструктура:** Docker, Docker Compose.

Пароли хранятся в виде bcrypt-хешей, авторизация реализована через подписанную HTTP-only cookie.

## Структура backend

Код разделён на предметные модули. Каждый основной модуль содержит:

- `handler` — обработку HTTP-запросов и ответов;
- `model` — структуры данных;
- `repository` — работу с PostgreSQL.

## Запуск

Для запуска необходимы Docker и Docker Compose.

```bash
git clone https://github.com/Mabuuka/autoservice.git
cd autoservice
cp .env.example .env
docker compose up --build
```

После запуска приложение будет доступно по адресу:

```text
http://localhost:8080
```

При первом запуске схема базы данных создаётся автоматически из файла `automaster_postgresql_schema.sql`.

Остановка контейнеров:

```bash
docker compose down
```
