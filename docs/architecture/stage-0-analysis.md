# Этап 0. Подготовка и анализ

## Проверенные файлы и каталоги
- `README.md`
- `api-specification.yaml`
- `docker-compose.yml`
- `tests/postman/README.md`
- `tests/postman/CinemaAbyss.postman_collection.json`
- `src/monolith`
- `src/microservices/movies`
- `src/microservices/proxy`
- `src/kubernetes/events-service.yaml`
- `src/kubernetes/proxy-service.yaml`

## Текущее состояние репозитория
- Монолит реализован в `src/monolith` и обслуживает домены `users`, `movies`, `payments`, `subscriptions`.
- Микросервис фильмов реализован в `src/microservices/movies` и использует ту же PostgreSQL-базу.
- `src/microservices/events` отсутствует.
- `src/microservices/proxy` пока не реализован: в каталоге есть только `.gitkeep`.
- `src/kubernetes/events-service.yaml` пустой.
- `src/kubernetes/proxy-service.yaml` пустой.

## Контракты и точки интеграции
- `docker-compose.yml` уже ожидает контейнеры `events-service` и `proxy-service`.
- В OpenAPI-спецификации описаны:
  - `GET /health` для `proxy-service`;
  - CRUD-маршруты `users`, `movies`, `payments`, `subscriptions`;
  - `GET /api/movies/health` для `movies-service`;
  - `GET /api/events/health`, `POST /api/events/movie`, `POST /api/events/user`, `POST /api/events/payment` для `events-service`.
- Postman-коллекция проверяет отдельно `movies-service`, `events-service` и `proxy-service`.
- Kafka-топики из `docker-compose.yml`: `movie-events`, `user-events`, `payment-events`.

## Обязательные артефакты по плану
- контейнерная диаграмма C4;
- код `proxy-service`;
- код `events-service`;
- CI workflow;
- Kubernetes manifests;
- Helm templates и values;
- артефакты проверок: скриншоты тестов, Kafka UI и логов;
- заполненный `Project_template.md`.

## Наблюдения
- `README.md` описывает `events-service` и `proxy-service` как существующие, но фактическая структура репозитория пока этому не соответствует.
- Переходное решение с общей БД уже фактически используется монолитом и `movies-service`, что подходит для целевой схемы миграции.
