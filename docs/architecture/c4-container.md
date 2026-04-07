# C4 Container Diagram

Контейнерная диаграмма фиксирует целевую to-be архитектуру CinemaAbyss на уровне контейнеров, но без низкоуровневых технических деталей. Она подходит для обсуждения со смешанной аудиторией: разработкой, аналитиками, менеджерами и эксплуатацией.

Основной исходник диаграммы в формате PlantUML:
[c4-container.puml](/home/dev15/develop/architecture-pro-cinemaabyss/docs/architecture/c4-container.puml)

Дополнительная контекстная диаграмма для ещё более общего уровня:
[c4-context.puml](/home/dev15/develop/architecture-pro-cinemaabyss/docs/architecture/c4-context.puml)

## Домены
- `users` и `payments` остаются внутри `monolith` на текущем этапе миграции.
- `subscriptions` остаются внутри `monolith`.
- `movies` выделены в `movies-service` и доступны через постепенное переключение трафика.
- `events/integration` вынесены в `events-service` и Kafka.

## Ключевые решения
- Все клиентские вызовы идут только через `proxy-service`.
- `movies-service` и `monolith` временно делят одну PostgreSQL-базу, что упрощает миграцию без двойной записи.
- `events-service` отделён от основного синхронного пользовательского потока и отвечает за событийную интеграцию.
- Детали API, маршрутизации, инфраструктуры и поставки вынесены в отдельные технические диаграммы.
