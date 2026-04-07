# План выполнения проектной работы 2 для ИИ-агента

## Цель
Подготовить репозиторий к сдаче проектной работы: спроектировать целевую архитектуру, реализовать `proxy-service` и `events-service`, настроить локальную проверку, CI/CD, Kubernetes-манифесты, Helm-чарт и заполнить артефакты для `Project_template.md`.

## Общие правила работы
1. Ветку не переключать: изменения подготавливаются в текущей ветке, коммит и публикацию выполняет пользователь.
2. Не откладывать документацию на конец: по завершении каждого крупного шага сразу обновлять `Project_template.md`.
3. После каждого этапа запускать минимальную проверку, чтобы не накапливать ошибки.
4. Все новые сервисы делать совместимыми с `docker-compose.yml`, Postman-тестами и Kubernetes-конфигурацией.

## Этап 0. Подготовка и анализ
1. Проверить структуру репозитория, `README.md`, `api-specification.yaml`, `tests/postman/*`, `docker-compose.yml`.
2. Убедиться, что сейчас:
   - есть монолит `src/monolith`;
   - есть сервис фильмов `src/microservices/movies`;
   - отсутствует реализованный `src/microservices/events`;
   - не реализован `src/microservices/proxy`;
   - пустые `src/kubernetes/events-service.yaml` и `src/kubernetes/proxy-service.yaml`.
3. Зафиксировать список обязательных артефактов:
   - диаграмма C4;
   - код `proxy-service`;
   - код `events-service`;
   - CI workflow;
   - Kubernetes manifests;
   - Helm templates/values;
   - скриншоты тестов, Kafka UI, логов;
   - заполненный `Project_template.md`.

Критерий готовности:
понимание всех обязательных файлов, API и точек проверки.

## Этап 1. To-Be архитектура
1. Выделить домены:
   - `users`;
   - `payments`;
   - `subscriptions`;
   - `movies`;
   - `events/integration`.
2. Определить целевую схему взаимодействия:
   - клиентские приложения идут только через `proxy-service`;
   - `proxy-service` маршрутизирует запросы в монолит, `movies-service` и `events-service`;
   - `events-service` работает с Kafka;
   - общая БД пока допустима как переходное решение;
   - Kubernetes и Ingress являются слоем доставки;
   - CI/CD собирает и публикует образы в GHCR.
3. Построить контейнерную диаграмму C4.
4. Сохранить диаграмму в репозитории и добавить ссылку в `Project_template.md`.

Критерий готовности:
есть файл с C4-диаграммой и ссылка на него в [Project_template.md](/home/dev15/develop/architecture-pro-cinemaabyss/Project_template.md).

## Этап 2. Реализация `proxy-service`
1. Создать сервис в `src/microservices/proxy`.
2. Добавить обязательные файлы:
   - `main.*`;
   - `Dockerfile`;
   - файл зависимостей выбранного языка.
3. Реализовать эндпоинты:
   - `GET /health`;
   - проксирование `GET/POST /api/users` в монолит;
   - проксирование `GET/POST /api/payments` в монолит;
   - проксирование `GET/POST /api/subscriptions` в монолит;
   - проксирование `GET/POST /api/movies` либо в монолит, либо в `movies-service`;
   - проксирование `/api/events/*` в `events-service`.
4. Реализовать Strangler Fig для `movies`:
   - использовать `GRADUAL_MIGRATION`;
   - использовать `MOVIES_MIGRATION_PERCENT`;
   - при `0` весь трафик идёт в монолит;
   - при `100` весь трафик идёт в `movies-service`;
   - при промежуточном значении маршрутизация должна быть частичной и воспроизводимой.
5. Проксировать метод, query string, тело запроса и базовые заголовки без поломки контракта.
6. Добавить понятные логи маршрутизации, чтобы было видно, куда ушёл запрос.

Минимальная проверка:
1. `docker-compose build`
2. `docker-compose up -d`
3. `curl http://localhost:8000/health`
4. `curl http://localhost:8000/api/movies`
5. Проверка маршрутизации при `MOVIES_MIGRATION_PERCENT=0`, `50`, `100`

Критерий готовности:
локально через прокси проходят запросы к монолиту, `movies-service` и `events-service`.

## Этап 3. Реализация `events-service` c Kafka
1. Создать директорию `src/microservices/events`.
2. Добавить обязательные файлы:
   - `main.*`;
   - `Dockerfile`;
   - файл зависимостей.
3. Реализовать API по спецификации:
   - `GET /api/events/health`;
   - `POST /api/events/movie`;
   - `POST /api/events/user`;
   - `POST /api/events/payment`.
4. Для каждого типа события:
   - валидировать тело запроса;
   - публиковать сообщение в соответствующий Kafka topic;
   - читать сообщение consumer'ом внутри того же сервиса;
   - писать в лог факт обработки.
5. Использовать топики из `docker-compose.yml` и Kubernetes-конфигурации:
   - `movie-events`;
   - `user-events`;
   - `payment-events`.
6. Возвращать корректный HTTP-статус и JSON-ответ, ожидаемый тестами.

Минимальная проверка:
1. `docker-compose up -d kafka zookeeper`
2. `docker-compose up -d events-service`
3. Проверка `GET /api/events/health`
4. Проверка POST-запросов на три типа событий
5. Проверка логов контейнера `events-service`
6. Проверка Kafka UI на `http://localhost:8090`

Критерий готовности:
сервис публикует и сам читает события, а результат видно в логах и Kafka UI.

## Этап 4. Локальная интеграция и Postman
1. Поднять весь стек через `docker-compose`.
2. Запустить тесты из `tests/postman`.
3. Добиться результата:
   - все тесты зелёные для локального запуска;
   - если в описании этапа допускается исключение для `events`, явно проверить актуальное требование по факту реализованного сервиса.
4. Снять артефакты:
   - скриншот Postman/Newman-тестов;
   - скриншот Kafka UI.
5. Обновить `Project_template.md`.

Критерий готовности:
локальная среда воспроизводимо проходит проверку, артефакты подготовлены.

## Этап 5. CI/CD в GitHub Actions
1. Доработать [docker-build-push.yml](/home/dev15/develop/architecture-pro-cinemaabyss/.github/workflows/docker-build-push.yml).
2. Проверить три блока:
   - триггеры workflow;
   - сборка и публикация образов `events-service` и `proxy-service`;
   - запуск API-тестов после сборки.
3. Убедиться, что workflow:
   - логинится в GHCR;
   - собирает `monolith`, `movies-service`, `events-service`, `proxy-service`;
   - публикует образы;
   - запускает тесты в среде CI.
4. Проверить имена образов и теги, чтобы их можно было использовать в Kubernetes.

Критерий готовности:
после пуша в GitHub workflow зелёный, образы появляются в GHCR.

## Этап 6. Kubernetes manifests
1. Обновить пути до образов в:
   - `src/kubernetes/monolith.yaml`;
   - `src/kubernetes/movies-service.yaml`;
   - `src/kubernetes/events-service.yaml`;
   - `src/kubernetes/proxy-service.yaml`.
2. Заполнить `src/kubernetes/dockerconfigsecret.yaml` своим base64 от `~/.docker/config.json`.
3. Реализовать `src/kubernetes/events-service.yaml`:
   - `Deployment`;
   - `Service`;
   - переменные окружения для Kafka и порта.
4. Реализовать `src/kubernetes/proxy-service.yaml`:
   - `Deployment`;
   - `Service`;
   - переменные окружения для URL сервисов и feature flag.
5. Доработать `src/kubernetes/ingress.yaml`:
   - маршрут `/` или `/api` в `proxy-service`;
   - маршрут `/api/events` в `events-service`, если это требуется тестами.
6. Проверить совместимость с:
   - `src/kubernetes/configmap.yaml`;
   - `src/kubernetes/secret.yaml`;
   - `src/kubernetes/kafka/kafka.yaml`.

Порядок проверки:
1. namespace
2. config/secrets
3. postgres
4. kafka
5. monolith
6. movies-service
7. events-service
8. proxy-service
9. ingress

Критерий готовности:
в кластере доступны `https://cinemaabyss.example.com/api/movies` и создание событий через ingress.

## Этап 7. Проверка Kubernetes
1. Применить манифесты по шагам из `Project_template.md`.
2. Проверить:
   - `kubectl -n cinemaabyss get pods`;
   - `kubectl -n cinemaabyss get svc`;
   - `kubectl -n cinemaabyss get ingress`;
   - логи `events-service`;
   - ответ `https://cinemaabyss.example.com/api/movies`.
3. Изменить `MOVIES_MIGRATION_PERCENT` в [configmap.yaml](/home/dev15/develop/architecture-pro-cinemaabyss/src/kubernetes/configmap.yaml) и убедиться, что прокси переводит трафик на `movies-service`.
4. Запустить Postman-тесты против Kubernetes-окружения.
5. Подготовить скриншоты:
   - ответ `/api/movies`;
   - логи обработки событий.

Критерий готовности:
кластер развёрнут, маршрутизация и события работают через ingress.

## Этап 8. Helm
1. Проверить и доработать chart в `src/kubernetes/helm`.
2. Сверить шаблоны сервисов Helm с обычными манифестами Kubernetes.
3. Проверить:
   - values для образов;
   - secret/configmap values;
   - ingress;
   - templates для `proxy-service` и `events-service`.
4. Выполнить:
   - `helm lint`;
   - `helm template`;
   - `helm install` или `helm upgrade --install`.
5. Повторно проверить `https://cinemaabyss.example.com/api/movies`.

Критерий готовности:
chart разворачивает приложение без ручных правок шаблонов после установки.

## Этап 9. Istio и Circuit Breaker
1. Найти и изучить `Project_template_adv.md` или эквивалентную инструкцию, если файл отсутствует в репозитории.
2. Развернуть Istio.
3. Настроить Circuit Breaker для `monolith` и `movies-service`.
4. Проверить поведение через Fortio.
5. Сохранить скриншот статистики и ссылку/описание в `Project_template.md`.

Критерий готовности:
есть рабочая конфигурация Istio и подтверждение с теста Fortio.

## Этап 10. Финализация сдачи
1. Полностью заполнить [Project_template.md](/home/dev15/develop/architecture-pro-cinemaabyss/Project_template.md):
   - ссылки на диаграммы;
   - описание решений;
   - скриншоты;
   - команды проверки.
2. Проверить, что в репозитории есть все нужные файлы и нет временного мусора.
3. Ещё раз прогнать минимум:
   - локальные тесты;
   - валидацию Helm;
   - базовую проверку Kubernetes-конфигов.
4. Подготовить PR из `cinema` в `main`.

Критерий готовности:
репозиторий готов к ревью, все обязательные артефакты добавлены.

## Рекомендуемый порядок выполнения без распараллеливания
1. Этап 0
2. Этап 1
3. Этап 2
4. Этап 3
5. Этап 4
6. Этап 5
7. Этап 6
8. Этап 7
9. Этап 8
10. Этап 9
11. Этап 10

## Что контролировать на каждом шаге
1. Соответствие `api-specification.yaml` и Postman-тестам.
2. Совместимость локальной, CI и Kubernetes-конфигурации.
3. Корректность имен контейнеров, сервисов, портов и topic'ов.
4. Обновление `Project_template.md` сразу после завершения этапа.
