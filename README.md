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
git clone github.com/100bench/subscription_aggregator.git
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

* **POST** `/subscriptions` — создание подписки
* **GET** `/subscriptions/{userID}` — все подписки пользователя
* **GET** `/subscriptions/{userID}/{serviceName}` — конкретная подписка
* **PUT** `/subscriptions/{userID}/{serviceName}` — обновление подписки
* **DELETE** `/subscriptions/{userID}/{serviceName}` — удаление подписки
* **GET** `/subscriptions/{userID}/total_cost` — общая стоимость подписок

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
curl -X POST http://localhost:8080/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "service_name": "Netflix",
    "price": 1500,
    "start_date": "01-2024",
    "end_date": "12-2024"
  }'
```

**Получение общей стоимости:**
```bash
curl "http://localhost:8080/subscriptions/123e4567-e89b-12d3-a456-426614174000/total_cost?service_name=Netflix&start_date=01-2024&end_date=12-2024"
```
