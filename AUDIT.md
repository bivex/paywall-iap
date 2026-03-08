# Frontend Audit: актуальный статус admin dashboard

Дата: 2026-03-08

## Что проверено

- текущие `page.tsx` для ранее отмеченных разделов
- client components на предмет реального data load / mutation wiring
- `frontend/src/actions/*` и server helpers в `frontend/src/lib/server/*`
- наличие внутренних App Router route handlers в `frontend/src/app/api/admin/*`

## Легенда

| Статус | Значение |
|---|---|
| ✅ Работает | страница реально подключена и основные действия на месте |
| 🟡 Частично | есть реальный read-path и/или часть mutations, но не весь задуманный workflow |
| ⚪ Read-only | страница живая, но это в основном мониторинг/просмотр без полноценных управляющих действий |

## Таблица актуального статуса

| Раздел | Read path | Write path | Актуальный статус | Что ещё не добито |
|---|---|---|---|---|
| Matomo Analytics | `getPlatformSettings()` → `MatomoPageClient` | настройки меняются через `/dashboard/settings`, не на самой странице | 🟡 Частично | сама страница уже не заглушка: строит real Matomo URL и умеет открыть embed/new tab; но отдельного KPI fetch/config UI здесь нет |
| Platform Settings | `getPlatformSettings()` → `/v1/admin/settings` | `updatePlatformSettings()`, `changeAdminPasswordAction()` | ✅ Работает | старый пункт аудита устарел: save/password wiring уже есть |
| Dunning | `getRevenueOps(1)` + `DunningQueueCard` | нет page-local mutation path | 🟡 Частично | это уже живой read-only экран очереди, а не mock; но отдельного config/editor flow на странице нет |
| Winback | `getWinbackCampaigns()` | `launchWinbackCampaignAction()` | 🟡 Частично | список и запуск кампании реальные; полноценного edit/update/deactivate flow для существующих кампаний не видно |
| Pricing Tiers | `getPricingTiers()` | `createPricingTierAction()`, `updatePricingTierAction()`, `activatePricingTierAction()`, `deactivatePricingTierAction()` | ✅ Работает | старый пункт аудита устарел: CRUD wiring уже подключён |
| A/B Tests | `getExperiments()` | `createExperimentAction()` | 🟡 Частично | список и создание эксперимента реальные; нет edit/pause/resume/complete management |
| Experiment Studio | `getStudioDashboardFromCookies()` + `/api/admin/studio/*` | нет write actions | ⚪ Read-only | страница стала runtime dashboard с live snapshot/probes; `Save Draft` / `Launch Test` builder workflow по-прежнему отсутствует |
| Bandit Model | `getBanditDashboardFromCookies()` + `/api/admin/bandit/*` | нет bandit config mutations на этой странице | ⚪ Read-only | live stats/metrics уже есть; `Pause` / `Stop` / `Save Config` не реализованы |
| Delayed Feedback | `getDelayedFeedbackDashboardFromCookies()` + `/api/admin/delayed-feedback/*` | manual POST ingest через `/api/admin/delayed-feedback/conversions` | 🟡 Частично | это уже не mock: есть probes и ручной conversion ingest; но editor/questionnaire management не реализован |
| Sliding Window | `getSlidingWindowDashboardFromCookies()` + `/api/admin/sliding-window/*` | trim через `/api/admin/sliding-window/trim` | 🟡 Частично | live read-path есть, trim работает; полноценного config/reset management beyond trim нет |
| Multi-Objective | `getMultiObjectiveDashboardFromCookies()` + `/api/admin/multi-objective/*` | config save через `/api/admin/multi-objective/config` | 🟡 Частично | objective scores и config save уже живые; отсутствует более широкий lifecycle/workbench поверх этого экрана |

## Что в старом аудите уже неверно

| Старое утверждение | Что реально сейчас |
|---|---|
| Нет action-файлов для `settings`, `winback`, `pricing`, `experiments` | Уже есть `frontend/src/actions/platform-settings.ts`, `winback.ts`, `pricing.ts`, `experiments.ts` |
| `settings`, `pricing`, `winback`, `experiments` — статические mock pages | Это уже реальные страницы с data load и как минимум частью mutations |
| advanced experiment pages — только витринные экраны | Сейчас они используют server helpers и внутренние `/api/admin/*` routes; часть экранов read-only, но wiring реальный |

## Внутренние `/api/admin/*` routes реально присутствуют

| Область | Routes |
|---|---|
| Studio | `/api/admin/studio/dashboard`, `/api/admin/studio/snapshot` |
| Bandit | `/api/admin/bandit/dashboard`, `/api/admin/bandit/snapshot` |
| Delayed Feedback | `/api/admin/delayed-feedback/dashboard`, `/api/admin/delayed-feedback/snapshot`, `/api/admin/delayed-feedback/conversions` |
| Sliding Window | `/api/admin/sliding-window/dashboard`, `/api/admin/sliding-window/snapshot`, `/api/admin/sliding-window/trim` |
| Multi-Objective | `/api/admin/multi-objective/dashboard`, `/api/admin/multi-objective/snapshot`, `/api/admin/multi-objective/config` |

## Что реально осталось not done

| Приоритет | Что добивать | Почему |
|---|---|---|
| P1 | Lifecycle actions для `A/B Tests` | список/создание есть, но нет полноценного управления статусами и редактирования |
| P1 | Write workflow для `Experiment Studio` | страница стала полезным read-only dashboard, но не studio в смысле draft/launch builder |
| P1 | Управляющие actions для `Bandit Model` | мониторинг есть, но нет pause/stop/save-config controls |
| P2 | Dunning page decision | либо сделать отдельный config/editor экран, либо честно оставить как read-only mirror Revenue Ops |
| P2 | Winback management beyond launch | launch работает, но нет полноценного управления существующими кампаниями |
| P2 | Sliding Window config UX | trim уже есть, но экран не даёт полноценного runtime/config management |
| P2 | Multi-Objective workbench | config save и score inspection есть, но это ещё не полный decision cockpit |
| P3 | Matomo dedicated analytics UX | embed через saved settings работает, но page-level KPI/config experience минимален |

## Concrete implementation checklist by file/path (top 3)

### 1) A/B Tests lifecycle actions

| File/path | Checklist |
|---|---|
| `backend/cmd/api/main.go` | добавить admin routes для update/status lifecycle эксперимента (минимум: update + pause/resume/complete) |
| `backend/internal/interfaces/http/handlers/admin_experiments.go` | реализовать handlers поверх существующих list/create: загрузка по `id`, update полей, смена `status`, валидация допустимых переходов |
| `backend/tests/integration/admin_experiments_test.go` | покрыть `PUT/PATCH/POST` lifecycle сценарии: draft→running, running→paused, paused→running, running→completed, invalid transition |
| `frontend/src/actions/experiments.ts` | добавить server actions: `updateExperimentAction`, `pauseExperimentAction`, `resumeExperimentAction`, `completeExperimentAction` |
| `frontend/src/app/(main)/dashboard/experiments/experiments-page-client.tsx` | добавить row actions, optimistic/local state update, disable/loading states, toast feedback |
| `frontend/src/lib/experiments.ts` | при необходимости расширить input/result types под update/status mutations |

### 2) Experiment Studio write workflow

| File/path | Checklist |
|---|---|
| `frontend/src/app/(main)/dashboard/experiments/studio/studio-page-client.tsx` | превратить read-only dashboard в builder: editable draft fields, arm editor, draft/save/launch CTA |
| `frontend/src/lib/experiment-studio.ts` | добавить form/view-model types для draft payload, launch payload, validation/result state |
| `frontend/src/app/api/admin/studio/dashboard/route.ts` | оставить read route как есть |
| `frontend/src/app/api/admin/studio/snapshot/route.ts` | оставить refresh/snapshot route как есть |
| `frontend/src/app/api/admin/studio/save/route.ts` **(new)** | добавить proxy route для save draft/update experiment |
| `frontend/src/app/api/admin/studio/launch/route.ts` **(new)** | добавить proxy route для launch/status transition draft→running |
| `frontend/src/lib/server/studio-admin.ts` | при необходимости вернуть enriched draft payload для studio editor, не только runtime snapshot |
| `frontend/src/server/server-actions.ts` | использовать cookie helpers, если выбранный draft/studio state нужно закреплять между refresh/navigation |
| `backend/cmd/api/main.go` + `backend/internal/interfaces/http/handlers/admin_experiments.go` | переиспользовать/добавить backend update + launch lifecycle endpoints, чтобы studio не inventил отдельную domain модель |

### 3) Bandit Model control actions

| File/path | Checklist |
|---|---|
| `backend/cmd/api/main.go` | добавить admin/bandit control routes или admin experiment lifecycle routes, которыми сможет пользоваться bandit page |
| `backend/internal/interfaces/http/handlers/admin_experiments.go` | если pause/resume/complete делаются через experiment status, реализовать это здесь |
| `backend/internal/interfaces/http/handlers/bandit_advanced.go` | если нужен отдельный bandit config save path, добавить handler/wiring рядом с objective/window endpoints |
| `backend/internal/interfaces/http/handlers/bandit_advanced_test.go` | добавить regression tests для новых control/config routes |
| `frontend/src/app/api/admin/bandit/dashboard/route.ts` | оставить текущий read route |
| `frontend/src/app/api/admin/bandit/snapshot/route.ts` | оставить текущий snapshot route |
| `frontend/src/app/api/admin/bandit/pause/route.ts` **(new)** | proxy route для pause action |
| `frontend/src/app/api/admin/bandit/resume/route.ts` **(new)** | proxy route для resume/start action |
| `frontend/src/app/api/admin/bandit/config/route.ts` **(new)** | proxy route для config save, если управление останется на bandit page |
| `frontend/src/app/(main)/dashboard/experiments/bandit/bandit-page-client.tsx` | добавить `Pause`, `Resume/Start`, `Save Config` controls, loading states и refresh snapshot после mutation |
| `frontend/src/lib/server/bandit-admin.ts` | при необходимости расширить snapshot/dashboard payload конфигом, который можно редактировать и сохранять |

## Cross-cutting замечания

| Наблюдение | Статус |
|---|---|
| Механизм `comingSoon` существует | да |
| Частично готовые страницы явно помечены как `comingSoon` | нет — `comingSoonUrls` пустой, поэтому partial pages выглядят как fully ready |
| `Subscriptions` и `Transactions` всё ещё имеют технический долг по sort | да, сортировка местами остаётся на фронте |
| `Delayed Feedback` probe для pending-by-id больше не должен считаться broken из-за ожидаемого `404` | да, это уже исправлено |

## Короткий вывод

Старый `AUDIT.md` был заметно устаревшим. На текущем коде уже **полностью рабочие** как минимум `Platform Settings` и `Pricing Tiers`, а `Winback`, `A/B Tests` и advanced experiment pages уже имеют **реальный wiring**, но часть из них пока остаётся **partial / read-only**, а не полным admin workflow.

