# API contract testing

## 1. Поднять локальное окружение

Из корня репозитория:

```bash
docker compose -f infra/docker-compose/docker-compose.local.yml up -d --build api apple-iap-mock google-billing-mock
```

Если БД / Redis ещё не подняты, можно поднять весь стек:

```bash
docker compose -f infra/docker-compose/docker-compose.local.yml up -d --build
```

## 2. Проверить health API

```bash
curl -fsS http://localhost:8081/health
```

Если health-check не проходит, контрактный прогон дальше не имеет смысла.

## 3. Основной контрактный прогон

Базовый запуск из корня репозитория:

```bash
bash ./scripts/test_api_contract_schemathesis.sh
```

Быстрый прогон с ранней остановкой на первой проблеме:

```bash
bash ./scripts/test_api_contract_schemathesis.sh --max-failures 1
```

## 4. Что script делает автоматически

По умолчанию `scripts/test_api_contract_schemathesis.sh`:

1. проверяет `GET /health`
2. берёт admin token через `/v1/admin/auth/login`
3. регистрирует fresh user через `/v1/auth/register`
4. создаёт valid iOS receipt в local Apple mock
5. подтверждает подписку через `POST /v1/verify/iap`
6. гоняет Schemathesis не одним run, а по группам:
   - `public endpoints`
   - `admin endpoints`
   - `admin auth endpoints`
   - `user-protected endpoints`

Это сделано специально, чтобы:
- убрать ложные admin auth warnings
- тестировать subscription endpoints на живых seeded данных

## 5. Когда нужен кастомный запуск

Если нужно тестировать только часть API, можно передать обычные аргументы Schemathesis:

```bash
bash ./scripts/test_api_contract_schemathesis.sh --include-tag admin --max-failures 1
```

```bash
bash ./scripts/test_api_contract_schemathesis.sh --include-path-regex '^/v1/admin/experiments$'
```

Важно: если переданы свои `--include-*`, `--exclude-*`, кастомные header'ы или token'ы, script переключается в single-run режим и **не** делает дефолтное split-by-suite поведение.

## 6. Полезные env vars

Можно переопределять:

- `API_BASE` — базовый URL API, по умолчанию `http://localhost:8081`
- `SCHEMA_URL` — URL OpenAPI схемы
- `SCHEMA_PATH` — путь до локальной OpenAPI схемы
- `SCHEMATHESIS_PHASES` — фазы, по умолчанию `examples,coverage,fuzzing`
- `SCHEMATHESIS_AUTH_TOKEN` — свой admin bearer token
- `SCHEMATHESIS_USER_AUTH_TOKEN` — свой user bearer token
- `SCHEMATHESIS_HEADER` — дополнительный header
- `SKIP_HEALTHCHECK=1` — пропустить health-check
- `ADMIN_EMAIL` / `ADMIN_PASSWORD` — admin credentials для auto-login
- `APPLE_MOCK_BASE` — base URL Apple IAP mock, по умолчанию `http://localhost:9090`
- `IAP_PRODUCT_ID` — product id для seed'а подписки

Пример:

```bash
API_BASE=http://localhost:8081 SCHEMATHESIS_PHASES=coverage,fuzzing bash ./scripts/test_api_contract_schemathesis.sh
```

## 7. Как мы читаем результат

### Success

- `return code = 0`
- нет блока `FAILURES`
- warnings допустимы, но их нужно разбирать отдельно

### Failure

Сначала смотрим:

1. какой suite упал: `public`, `admin`, `admin auth`, `user-protected`
2. тип проблемы:
   - `Server error`
   - `API rejected schema-compliant request`
   - `API accepted schema-violating request`
   - `Missing test data`
   - `Schema validation mismatch`

## 8. Если schema менялась

После изменения OpenAPI или handler'ов полезно пересобрать API:

```bash
docker compose -f infra/docker-compose/docker-compose.local.yml up -d --build api
```

И затем повторить прогон:

```bash
bash ./scripts/test_api_contract_schemathesis.sh
```

## 9. Минимальный рабочий цикл

Обычно тестим так:

```bash
docker compose -f infra/docker-compose/docker-compose.local.yml up -d --build api
curl -fsS http://localhost:8081/health
bash ./scripts/test_api_contract_schemathesis.sh --max-failures 1
```

Если нашли mismatch / 500:

1. чиним code или OpenAPI
2. при необходимости пересобираем `api`
3. повторяем targeted run
4. потом повторяем полный run