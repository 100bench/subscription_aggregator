# Subscription Aggregator

**Тестовое задание**  
REST API агрегатора подписок: управление подписками пользователей и расчет общей стоимости.

---

## Описание

REST API для управления подписками пользователей с возможностью:

* Создания, обновления, удаления подписок
* Получения списка подписок пользователя
* Расчет общей стоимости подписок с фильтрацией по сервису и периоду
* Swagger документация

---

## Технологии

* **Go 1.24**
* **PostgreSQL**
* **Swagger (OpenAPI 3.0)**
* **Docker, docker-compose**
* **Чистая архитектура**

---

## Запуск

1. **Клонируйте репозиторий:**
```bash
git clone https://github.com/100bench/subscription_aggregator.git
cd subscribtion_agregator
```

2. **Запустите сервисы:**
```bash
docker-compose up --build -d
```

3. **API будет доступен на:** http://localhost:8080
4. **Swagger UI:** http://localhost:8080/swagger/index.html

---

## API Endpoints

* **POST** `/api/v1/subscriptions` — создание подписки
* **GET** `/api/v1/subscriptions/{id}` — получение подписки по ID
* **GET** `/api/v1/subscriptions?user_id={userID}&service_name={serviceName}` — список подписок с фильтрами
* **PUT** `/api/v1/subscriptions/{id}` — обновление подписки
* **DELETE** `/api/v1/subscriptions/{id}` — удаление подписки
* **GET** `/api/v1/subscriptions/calculate?user_id={userID}&start_date={MM-YYYY}&end_date={MM-YYYY}&service_name={serviceName}` — общая стоимость подписок

**Подробная документация:** http://localhost:8080/swagger/index.html

---

## Структура проекта

```
subscribtion_agregator/
├── cmd/                    # Точка входа приложения
├── internal/
│   ├── entities/          # Бизнес-сущности
│   ├── cases/             # Use cases (бизнес-логика)
│   └── adapters/          # Адаптеры (PostgreSQL)
├── pkg/dto/               # Data Transfer Objects
├── ports/http/public/     # HTTP handlers
├── migrations/            # SQL миграции
├── Dockerfile
└── docker-compose.yml
```

---

## Примеры запросов

**Создание подписки:**
```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "service_name": "Netflix",
    "price": 1500,
    "start_date": "01-2024",
    "end_date": "12-2024"
  }'
```

**Получение подписки по ID:**
```bash
curl "http://localhost:8080/api/v1/subscriptions/1"
```

**Получение списка подписок пользователя:**
```bash
curl "http://localhost:8080/api/v1/subscriptions?user_id=123e4567-e89b-12d3-a456-426614174000&page=1&page_size=10"
```

**Получение списка с фильтром по сервису:**
```bash
curl "http://localhost:8080/api/v1/subscriptions?user_id=123e4567-e89b-12d3-a456-426614174000&service_name=Netflix"
```

**Обновление подписки:**
```bash
curl -X PUT http://localhost:8080/api/v1/subscriptions/1 \
  -H "Content-Type: application/json" \
  -d '{
    "price": 1800,
    "end_date": "12-2025"
  }'
```

**Удаление подписки:**
```bash
curl -X DELETE "http://localhost:8080/api/v1/subscriptions/1"
```

**Получение общей стоимости:**
```bash
curl "http://localhost:8080/api/v1/subscriptions/calculate?user_id=123e4567-e89b-12d3-a456-426614174000&service_name=Netflix&start_date=01-2024&end_date=12-2024"
```
