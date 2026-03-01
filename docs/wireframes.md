GitHub Copilot Chat Assistant

Ниже — компактное сопоставление ранее предложённых вайрфреймов админ‑панели с реальным API и кодом в репозитории bivex/paywall-iap. Я привожу: экран → существующие endpoint(ы) + DTOs + где реализовано (файлы), а также отмечаю отсутствующие или частично реализованные API, которые потребуется добавить для полного покрытия UX (templates, reconciliation UI и т.п.).

Важное: результаты поиска по коду могли быть неполными. Посмотреть больше результатов в GitHub UI:  
- Репозиторий: https://github.com/bivex/paywall-iap  
- Быстрый поиск по bandit: https://github.com/bivex/paywall-iap/search?q=bandit

1) Общая навигация / Auth / Health
- Health
  - Endpoint: GET /health
  - Где: backend/cmd/api/main.go (router.GET("/health"...))
- Auth
  - Регистрация: POST /v1/auth/register
  - Refresh: POST /v1/auth/refresh
  - DTOs: backend/internal/application/dto/dto.go (RegisterRequest, RefreshTokenRequest/Response)
  - Mobile: mobile/src/infrastructure/api/AuthService.ts

2) Dashboard (KPI, метрики)
- Endpoint (админ): GET /v1/admin/dashboard/metrics
- Где: backend/internal/interfaces/http/handlers/admin.go (GetDashboardMetrics регистрация в main.go)
- Доп. источники: analytics endpoints / API docs (docs/api/growth-layer.md — LTV endpoints)
- Замечание: детализированные графики (cohorts, funnels) берут данные из analytics service / cache — см. backend tests и docs.

3) Customers / Users экран
- Существующие endpoints:
  - Admin list users: GET /v1/admin/users
  - Grant subscription: POST /v1/admin/users/:id/grant
  - Revoke subscription: POST /v1/admin/users/:id/revoke
  - Где: backend/internal/interfaces/http/handlers/admin.go
  - DTOs: Grant request struct in admin.go; user-related DTOs/queries в backend/internal/application/dto/dto.go и sqlc queries (users.sql in docs)
- Mobile / backend user flow: registration & tokens (mobile AuthService)

4) Products / Plans
- Частично реализовано:
  - DTO PricingTier присутствует: backend/internal/application/dto/dto.go (PricingTier)
  - DB schema for products/plans — есть миграции для subscriptions/transactions; явных CRUD endpoints для products/plans в коде не найдены в быстром поиске.
- Рекомендация: добавить CRUD API для templates/pricing tiers (если нужен UI для управления продуктами).

5) Subscriptions экран (проверка доступа, отмена, детали)
- Endpoints:
  - GET /v1/subscription — получить подписку
  - GET /v1/subscription/access — check access (registered in main.go; protected)
  - DELETE /v1/subscription — cancel subscription
  - Verify IAP: POST /v1/verify/iap (used in tests and DTOs)
- Где: main.go регистрирует protected routes; DTOs VerifyIAPRequest/VerifyIAPResponse — backend/internal/application/dto/dto.go; tests coverage — backend tests/integration reference /v1/verify/iap.
- DB: migrations create subscriptions table (backend/migrations/002_create_subscriptions.up.sql)

6) Transactions / Purchases (таблица и детали)
- DB: transactions table exists in migrations (migrations/002...)
- API: нет явного публичного admin /transactions endpoints в найденных результатах — используются внутренние flows (receipt verification creates transactions).
- Рекомендация: добавить admin endpoints: GET /v1/admin/transactions, GET /v1/admin/transactions/:id, bulk actions (refunds, mark_reconciled).

7) Reconciliation
- Частичная реализация:
  - Worker tasks for reconciliation/grace period exist (backend/internal/worker/tasks/tasks.go)
  - Документы описывают reconciliation процессы (docs)
- Но API для запуска/просмотра reconciliation runs из админ UI не найден.
- Рекомендация: добавить endpoints: POST /v1/admin/reconciliation/run, GET /v1/admin/reconciliation/runs/:id, GET /v1/admin/reconciliation/issues

8) Revenue & Reports / Analytics
- Endpoints / docs:
  - Analytics LTV: GET /api/v1/analytics/ltv?user_id=... (docs/api/growth-layer.md) — в тестах используется similar router
  - Realtime metrics: GET /api/v1/analytics/realtime (tests & docs)
  - Cohort LTV: GET /api/v1/analytics/cohort/ltv (docs)
- Где: tests set up a simplified analytics router (backend/tests/integration/analytics_test.go). Реальные handler'ы в analytics service — смотреть backend/internal/service/analytics* (по репо).
- Рекомендация: в UI привязать KPI блоки к GET /v1/admin/dashboard/metrics и к аналитическим endpoints.

9) Integrations / Connectors / Webhooks
- Webhooks:
  - /webhook/stripe, /webhook/apple, /webhook/google — зарегистрированы в main.go (unauthenticated but signature-verified)
- Webhook monitoring:
  - Logs & retry endpoints не обнаружены как UI API — есть worker stubs & logging.
- Рекомендация: добавить admin endpoints для просмотра последних webhook deliveries и ручного retry.

10) Webhooks & Logs / Audit
- Audit service usage: Admin grant subscription writes audit log (auditService) — backend/internal/interfaces/http/handlers/admin.go
- General logs: logging middleware configured in main.go
- API to fetch audit logs not found — добавить GET /v1/admin/audit.

11) API keys & Roles / RBAC
- Admin middleware: middleware.AdminMiddleware(...) registered in main.go for /v1/admin
- Roles presence referenced in docs/tests, but explicit API for API key management not found.
- Рекомендация: реализовать endpoints для API keys CRUD and roles management (GET/POST /v1/admin/api-keys, /v1/admin/roles).

12) A/B Testing (Experiments / Variations / Assignments / Rewards)
- Реализация в репо:
  - Bandit endpoints (multi‑armed bandit):
    - POST /v1/bandit/assign  — Assign handler (backend/internal/interfaces/http/handlers/bandit.go)
    - POST /v1/bandit/reward  — Reward handler (records reward)
    - GET /v1/bandit/statistics, GET /v1/bandit/health
  - Routes registered in main.go under v1 group: bandit.POST("/assign", ...)
  - Docs also include planned AB endpoints under /api/v1/ab/* (docs/plans/2026-03-01-growth-layer-design.md) — similar to bandit.
  - Assignment cache design and experiment spec described in docs (assignment cache keys, algorithms such as thompson).
- DTOs: bandit.go contains AssignRequest/AssignResponse/RewardRequest/RewardResponse definitions.
- Mobile: SDK should call bandit assign/reward (docs mention /api/v1/ab endpoints and assignment cache).
- Mapping к вайрфрейму:
  - Экран Experiments → используется bandit service, но UI endpoints для CRUD Experiments (create/modify experiment, variations, template linking) в коде не обнаружены — только runtime assignment + reward + statistics.
  - Variation editor / template management — отсутствует серверная реализация template CRUD (docs mention templates but no handler found).
- Рекомендация: добавить:
  - CRUD endpoints for experiments & arms: POST/GET/PUT/DELETE /v1/admin/ab/experiments
  - Template management endpoints: /v1/admin/templates
  - Assignment inspection: GET /v1/admin/ab/assignments?experiment_id=...
  - Reports: GET /v1/admin/ab/:id/report (or reuse bandit/statistics)

13) Advanced Thompson Sampling (Currency, Contextual, Sliding Window, Delayed Feedback, Multi-Objective)
- Реализованные расширения bandit (новые endpoints):
  - Currency management:
    - GET /v1/bandit/currency/rates — текущие курсы валют (USD base)
    - POST /v1/bandit/currency/update — обновить курсы с ECB API
    - POST /v1/bandit/currency/convert — конвертировать сумму в USD
    - Где: backend/internal/interfaces/http/handlers/bandit_advanced.go
    - Worker: автоматическое обновление каждый час (currency_asynq.go)
  - Objective configuration (multi-objective optimization):
    - GET /v1/bandit/experiments/:id/objectives — счета целей (conversion/ltv/revenue/hybrid)
    - PUT /v1/bandit/experiments/:id/objectives/config — настроить веса целей
    - Supports: conversion rate optimization, LTV maximization, revenue optimization, hybrid (weighted combination)
  - Sliding window management:
    - GET /v1/bandit/experiments/:id/window/info — статистика окна
    - POST /v1/bandit/experiments/:id/window/trim — очистить старые события
    - GET /v1/bandit/experiments/:id/window/events — экспорт событий
    - Redis Sorted Sets для O(log N) операций
  - Delayed feedback (отложенные конверсии):
    - POST /v1/bandit/conversions — обработка конверсии (link к pending reward)
    - GET /v1/bandit/pending/:id — данные pending reward
    - GET /v1/bandit/users/:id/pending — список pending rewards пользователя
    - Worker: обработка истёкших pending каждые 15 минут
  - Metrics & maintenance:
    - GET /v1/bandit/experiments/:id/metrics — метрики эксперимента (balance index, pending rewards, etc.)
    - POST /v1/bandit/maintenance — запустить обслуживание вручную
- Contextual bandit (LinUCB):
  - Персонализация на основе: country, device, app_version, days_since_install, total_spent, last_purchase_at
  - Feature vector: 20 dimensions (one-hot country, device, numeric features, bias)
  - UCB = theta^T * x + alpha * sqrt(x^T * A^(-1) * x)
  - Модель обновляется с каждым reward (online learning)
- DB schema (migration 017):
  - currency_rates — курсы валют (base=USD)
  - bandit_user_context — кеш атрибутов пользователя
  - bandit_arm_context_model — LinUCB параметры (matrix_a, vector_b, theta)
  - bandit_window_events — события для sliding window
  - bandit_pending_rewards — отложенные конверсии
  - bandit_conversion_links — связь pending → transaction
  - bandit_arm_objective_stats — статистика по objectives
  - ab_tests — новые колонки: window_type, objective_type, enable_contextual, enable_delayed, enable_currency, exploration_alpha
- Services:
  - CurrencyRateService — ECB API integration + Redis cache + fallback rates
  - LinUCBSelectionStrategy — contextual arm selection
  - SlidingWindowStrategy — non-stationary behavior handling
  - DelayedRewardStrategy — delayed feedback processing
  - HybridObjectiveStrategy — multi-objective optimization
  - AdvancedBanditEngine — главный оркестратор (композиция всех стратегий)
- Mapping к UI:
  - Эксперименты → добавить настройки: objective type, weights, window config, contextual/delayed/currency toggles
  - Конверсии → автоматически link через pending rewards (transaction_id → pending_id)
  - Dashboard → новые метрики: currency conversion stats, objective scores, window utilization, pending rewards count
- Рекомендация: UI для управления:
  - GET /v1/admin/ab/experiments/:id/config — редактор конфигурации (objectives, window, features)
  - GET /v1/admin/bandit/currency/history — история курсов валют
  - GET /v1/admin/bandit/pending?status=expired — список истёкших pending для ручной обработки
  - GET /v1/admin/bandit/metrics/:id — детализированные метрики (regret, convergence, balance)

14) Events ingestion (paywall impressions, clicks, purchases)

14) DB / data model (что уже есть)
- Subscriptions, transactions tables exist (backend/migrations/002_create_subscriptions.up.sql)
- Tests and docs show expected tables and queries (sqlc queries in docs)
- For experiments: docs specify tables/redis keys (assignment cache), but SQL migrations for experiments/variations/assignments were not found in quick search.

15) Что отсутствует/неполно для полного покрытия вайрфреймов (кратко)
- Templates CRUD + Template versioning API (for Variation Editor)
- Experiments CRUD UI API (create/edit/targeting/scheduling)
- Admin endpoints for transactions listing, reconciliation runs and issue resolving
- Audit / Activity list API (exposed to UI)
- API Keys management endpoints
- Unified event ingestion endpoint for paywall impressions/clicks (instead of only bandit reward)

16) Предложение следующих шагов
- Хочешь, чтобы я:
  - A) сгенерировал аннотацию wireframes, где рядом с каждым элементом стоит конкретный endpoint и DTO (lo‑fi JSON / table), или
  - B) подготовил OpenAPI‑стабы для недостающих endpoints (templates, experiments CRUD, reconciliation), или
  - C) сделал список SQL миграций + sqlc queries для experiments/variations/assignments?
Выбери вариант — и я подготовлю артефакт.

Если нужно — могу также привести конкретные примеры запросов/ответов для каждого mapped endpoint (на основе DTOs в коде и docs).