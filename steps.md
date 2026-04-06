# Пошаговая инструкция выполнения проектной работы "Кинобездна 2"

Эта инструкция предназначена для ИИ-агента для последовательного выполнения задач по рефакторингу монолита, контейнеризации, оркестрации и внедрению Service Mesh.

---

## Подготовка
1. Изучить структуру проекта и файлы `Project_template.md`, `api-specification.yaml`.
2. Работать в текущей ветке.
3. **ВАЖНО:** Агент НЕ должен делать коммиты. Все коммиты в ветку `cinema` пользователь выполнит самостоятельно.

---

## Шаг 1: Проектирование архитектуры (To-Be)
1. Спроектировать целевую архитектуру системы.
2. Разделить монолит на домены (Movies, Events, Auth/Billing и т.д.).
3. Определить API Gateway (Proxy) как единую точку входа.
4. Создать диаграмму контейнеров в нотации **C4**.
5. Сохранить диаграмму в проект и добавить ссылку в `Project_template.md` (Задание 1).

---

## Шаг 2: Реализация микросервисов

### 2.1 Прокси-сервис (Strangler Fig)
1. Реализовать сервис в `./src/microservices/proxy` (Go, Node.js или Python).
2. Реализовать логику проксирования:
   - Если `GRADUAL_MIGRATION == "true"`:
     - Использовать `MOVIES_MIGRATION_PERCENT` для распределения трафика между `monolith` и `movies-service` для эндпоинта `/api/movies`.
   - Проксировать `/api/events` на `events-service`.
   - Остальные запросы проксировать на `monolith`.
3. Создать `Dockerfile` для сервиса.
4. Проверить работу через `docker-compose up` и запуск тестов из `tests/postman` (`npm run test:local`).

### 2.2 Сервис событий (Kafka MVP)
1. Реализовать сервис в `./src/microservices/events`.
2. Реализовать Producer и Consumer для Kafka:
   - API эндпоинты для создания событий (User, Payment, Movie) согласно `api-specification.yaml`.
   - При получении запроса отправлять сообщение в топик Kafka.
   - Consumer должен читать сообщения из этого же топика и записывать их в лог.
3. Добавить сервис в `docker-compose.yml`.
4. Проверить работу через Postman и UI Kafka (http://localhost:8090). Сделать скриншоты для отчета.

---

## Шаг 3: CI/CD и Kubernetes

### 3.1 Настройка GitHub Actions
1. Отредактировать `.github/workflows/docker-build-push.yml`.
2. Добавить шаги сборки и пуша образов для `proxy` и `events` сервисов в GitHub Packages (GHCR).
3. Убедиться, что `api-tests.yml` успешно проходит при пуше.

### 3.2 Настройка Kubernetes (Manifests)
1. Доработать манифесты в `src/kubernetes/`:
   - `proxy-service.yaml`: Deployment и Service для прокси.
   - `events-service.yaml`: Deployment и Service для сервиса событий.
   - `ingress.yaml`: Настроить правила маршрутизации для внешнего доступа.
   - `configmap.yaml`: Настроить переменные окружения, включая `MOVIES_MIGRATION_PERCENT`.
2. Развернуть систему в локальном кластере (minikube/kind).
3. Проверить доступность по адресу `https://cinemaabyss.example.com/api/movies`.
4. Сделать скриншоты логов `events-service` и ответов API.

---

## Шаг 4: Реализация Helm-чартов
1. Создать/доработать Helm-чарт в `src/kubernetes/helm/`.
2. Вынести конфигурации (реплики, теги образов, лимиты, переменные окружения) в `values.yaml`.
3. Использовать шаблоны (templates) для всех компонентов (proxy, events, monolith, movies, postgres, kafka).
4. Проверить установку через `helm install`.
5. Убедиться в работоспособности системы после деплоя через Helm.

---

## Шаг 5: Istio и Circuit Breaker
1. Установить Istio в кластер.
2. Настроить `VirtualService` и `DestinationRule` для сервисов.
3. Реализовать паттерн **Circuit Breaker** для `monolith` и `movies-service` согласно `Project_template_adv.md`.
4. Провести нагрузочное тестирование с помощью `Fortio`.
5. Сделать скриншоты статистики и добавить в отчет.

---

## Завершение
1. Заполнить все разделы в `Project_template.md`.
2. Приложить все необходимые скриншоты.
3. Создать Pull Request из ветки `cinema` в `main`.
4. Проверить, что PR создан в свой репозиторий.
