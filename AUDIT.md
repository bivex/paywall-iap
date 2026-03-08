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
| Platform Settings | `getPlatformSettings()` → `/v1/admin/settings` | `updatePlatformSettings()`, `changeAdminPasswordAction()` | ✅ Работает | save/password wiring есть и уже давно не placeholder |
| Dunning | `getRevenueOps(1)` + `DunningQueueCard` | page-local filters/search в query state, без backend mutation | 🟡 Частично | это живой экран очереди с операторскими фильтрами, но отдельного config/editor flow на странице нет |
| Winback | `getWinbackCampaigns()` | `launchWinbackCampaignAction()`, `deactivateWinbackCampaignAction()` | 🟡 Частично | список, запуск и деактивация кампании реальные; полноценного edit/update flow для существующих кампаний по-прежнему нет |
| Pricing Tiers | `getPricingTiers()` | `createPricingTierAction()`, `updatePricingTierAction()`, `activatePricingTierAction()`, `deactivatePricingTierAction()` | ✅ Работает | standalone CRUD wiring подключён; тот же live manager встроен в Studio, но linkage tier↔experiment arms всё ещё нет |
| A/B Tests | `getExperiments()` | `createExperimentAction()`, `pauseExperimentAction()`, `resumeExperimentAction()`, `completeExperimentAction()` | 🟡 Частично | список/создание/lifecycle реальные; всё ещё нет inline edit для существующих экспериментов, и остаётся hydration mismatch на timestamp render |
| Experiment Studio | `getStudioDashboardFromCookies()` + `/api/admin/studio/*` | `updateExperimentAction()`, `pauseExperimentAction()`, `resumeExperimentAction()`, `completeExperimentAction()`, embedded `PricingTierManager` actions | 🟡 Частично | metadata edit, lifecycle controls и live pricing manager уже есть; но полноценного arm editor / draft builder / явного tier-per-arm workflow пока нет |
| Bandit Model | `getBanditDashboardFromCookies()` + `/api/admin/bandit/*` | `updateExperimentAction()` (draft config), `pauseExperimentAction()`, `resumeExperimentAction()`, `completeExperimentAction()` | 🟡 Частично | live stats + lifecycle + draft config edit уже есть; runtime-specific bandit knobs/workbench beyond persisted experiment fields всё ещё отсутствуют |
| Delayed Feedback | `getDelayedFeedbackDashboardFromCookies()` + `/api/admin/delayed-feedback/*` | manual POST ingest через `/api/admin/delayed-feedback/conversions`, pending lookups via `/api/admin/delayed-feedback/pending/*` и `/api/admin/delayed-feedback/users/*` | 🟡 Частично | это уже не mock: есть probes, ручной conversion ingest и lookup pending rewards; editor/questionnaire management не реализован |
| Sliding Window | `getSlidingWindowDashboardFromCookies()` + `/api/admin/sliding-window/*` | trim через `/api/admin/sliding-window/trim`, events inspect/export via `/api/admin/sliding-window/events` | 🟡 Частично | live read-path, trim и export событий уже есть; полноценного config/reset management beyond trim нет |
| Multi-Objective | `getMultiObjectiveDashboardFromCookies()` + `/api/admin/multi-objective/*` | config read/save через `/api/admin/multi-objective/config` | 🟡 Частично | objective scores и persisted config wiring уже живые; отсутствует более широкий optimizer/workbench поверх этого экрана |

## Что в старом аудите уже неверно

| Старое утверждение | Что реально сейчас |
|---|---|
| Нет action-файлов для `settings`, `winback`, `pricing`, `experiments` | Уже есть `frontend/src/actions/platform-settings.ts`, `winback.ts`, `pricing.ts`, `experiments.ts` |
| `settings`, `pricing`, `winback`, `experiments` — статические mock pages | Это уже реальные страницы с data load и как минимум частью mutations |
| advanced experiment pages — только витринные экраны | Сейчас они используют server helpers и внутренние `/api/admin/*` routes; многие страницы уже умеют lifecycle/config/operator actions, хотя ещё не все workflow полные |
| `Experiment Studio` / `Bandit Model` остаются только read-only | Уже неверно: на обеих страницах появились lifecycle controls, а в Studio/Bandit также есть truthful draft/config editing в рамках реально существующих полей |

## Внутренние `/api/admin/*` routes реально присутствуют

| Область | Routes |
|---|---|
| Studio | `/api/admin/studio/dashboard`, `/api/admin/studio/snapshot` |
| Bandit | `/api/admin/bandit/dashboard`, `/api/admin/bandit/snapshot` |
| Delayed Feedback | `/api/admin/delayed-feedback/dashboard`, `/api/admin/delayed-feedback/snapshot`, `/api/admin/delayed-feedback/conversions`, `/api/admin/delayed-feedback/pending/[id]`, `/api/admin/delayed-feedback/users/[id]/pending` |
| Sliding Window | `/api/admin/sliding-window/dashboard`, `/api/admin/sliding-window/snapshot`, `/api/admin/sliding-window/trim`, `/api/admin/sliding-window/events` |
| Multi-Objective | `/api/admin/multi-objective/dashboard`, `/api/admin/multi-objective/snapshot`, `/api/admin/multi-objective/config` |

## Что добили после прошлого апдейта аудита

| Что добили | Что именно теперь есть |
|---|---|
| A/B Tests lifecycle | реальные `Start/Pause/Resume/Complete` actions на странице и backend lifecycle endpoints под них |
| Experiment Studio controls | lifecycle controls для выбранного эксперимента, draft metadata edit и встраивание live `PricingTierManager` |
| Bandit Model controls | lifecycle controls и truthful draft config edit для сохранённых полей эксперимента |
| Winback management | добавлена реальная деактивация существующих winback campaigns, а не только launch |
| Dunning usability | добавлены операторские filters/search вместо полностью статичного standalone экрана |
| Delayed Feedback tooling | lookup pending rewards по reward id и user id через реальные admin proxy routes |
| Sliding Window tooling | inspect/export live window events через реальный `/api/admin/sliding-window/events` |
| Multi-Objective config wiring | форма больше не стартует из hardcoded defaults — грузит и сохраняет persisted config |
| Studio / pricing truthfulness | pricing tiers редактируются из Studio тем же live CRUD, что и на standalone pricing page |
| Cold-start local data | есть `scripts/seed_all_test_data.sh`, интеграция в `run_dev.sh`, и детерминированные experiment/bandit fixtures для непустых admin pages |
| UI polish | в A/B Tests исправлено перекрытие long description text в таблице; Studio hydration mismatch по датам уже был добит ранее |

## Что реально осталось недоделанным

| Приоритет | Что добивать | Почему |
|---|---|---|
| P1 | True builder workflow для `Experiment Studio` | metadata edit и lifecycle уже есть, но Studio всё ещё не умеет полноценно собирать/редактировать arms и управлять draft как конструктором |
| P1 | Truthful linkage pricing tiers ↔ experiment arms | tiers уже live и доступны в Studio, но реальной persisted связи между tier catalogue и arms/domain экспериментом пока нет |
| P1 | `A/B Tests` polish: inline edit + hydration-safe timestamps | lifecycle уже добит, но редактирование существующих экспериментов и SSR/client timestamp mismatch ещё торчат |
| P2 | Runtime-specific config UX для `Bandit Model` | draft-only persisted config edit есть, но не хватает более глубокого runtime workbench поверх bandit metrics |
| P2 | Winback management beyond launch/deactivate | launch и deactivate работают, но нет полноценного update/edit flow кампаний |
| P2 | Sliding Window config/reset UX | trim и export уже есть, но экран не даёт полноценного runtime/config management |
| P2 | Multi-Objective optimizer workbench | persisted config и score inspection есть, но это ещё не полный decision cockpit |
| P3 | Dunning page product decision | либо делать отдельный config/editor экран, либо честно оставить её как operations mirror Revenue Ops |
| P3 | Matomo dedicated analytics UX | embed через saved settings работает, но page-level KPI/config experience минимален |

## Concrete implementation checklist by file/path (top 3)

### 1) Experiment Studio builder / arm editor

| File/path | Checklist |
|---|---|
| `frontend/src/app/(main)/dashboard/experiments/studio/studio-page-client.tsx` | добавить truthful arm editor для draft experiments: create/remove/relabel arms, editable weights/descriptions, clear save/apply UX |
| `frontend/src/lib/experiment-studio.ts` | расширить types/validation под editable arm payload, а не только metadata form state |
| `frontend/src/actions/experiments.ts` | если backend позволит, добавить server actions для arm create/update/delete без изобретения отдельной модели |
| `backend/internal/interfaces/http/handlers/admin_experiments.go` | поддержать persisted arm editing поверх существующих experiment endpoints или рядом с ними |
| `backend/tests/integration/admin_experiments_test.go` | покрыть draft arm-edit сценарии и защиту от некорректных weight/status комбинаций |

### 2) Pricing tier linkage to experiments

| File/path | Checklist |
|---|---|
| `backend/internal/domain/*` + `backend/internal/infrastructure/db/*` | определить, есть ли truthful место для persisted связи arm ↔ pricing tier; если нет — сначала ввести явную модель/таблицу вместо UI-притворства |
| `backend/internal/interfaces/http/handlers/admin_experiments.go` | отдать linkage в admin payload, если доменная связь будет добавлена |
| `frontend/src/lib/server/studio-admin.ts` | вернуть linkage в enriched Studio payload рядом с arms/pricing catalogue |
| `frontend/src/app/(main)/dashboard/experiments/studio/studio-page-client.tsx` | показывать/редактировать реальное сопоставление arm→pricing tier, а не просто два соседних независимых блока |
| `frontend/src/components/admin/pricing-tier-manager.tsx` | при необходимости поддержать режим pick/link, а не только standalone CRUD |

### 3) A/B Tests polish and remaining truthfulness gaps

| File/path | Checklist |
|---|---|
| `frontend/src/app/(main)/dashboard/experiments/experiments-page-client.tsx` | добавить truthful inline edit / navigation to edit for existing experiments; заодно добить hydration-safe timestamp render на overview table |
| `frontend/src/lib/experiments.ts` | при необходимости расширить UI types под richer edit state/result |
| `frontend/src/actions/experiments.ts` | переиспользовать существующий update action для overview-page editing, не разводя дублирующий mutation слой |
| `frontend/src/lib/time.ts` / shared date helpers | вынести единый hydration-safe форматтер дат, чтобы не ловить повторно SSR/client mismatch на experiment pages |

## Cross-cutting замечания

| Наблюдение | Статус |
|---|---|
| Механизм `comingSoon` существует | да |
| Частично готовые страницы явно помечены как `comingSoon` | нет — `comingSoonUrls` пустой, поэтому partial pages выглядят как fully ready |
| `Subscriptions` и `Transactions` всё ещё имеют технический долг по sort | да, сортировка местами остаётся на фронте |
| `Delayed Feedback` probe для pending-by-id больше не должен считаться broken из-за ожидаемого `404` | да, это уже исправлено |
| Локальный cold-start для demo/admin страниц теперь воспроизводим | да — через `scripts/seed_all_test_data.sh`, интегрированный в `run_dev.sh` |
| Playwright smoke по seeded experiment pages уже проходили | да — страницы грузятся с непустыми данными; отдельно остаётся hydration mismatch на `/dashboard/experiments` timestamps |

## Короткий вывод

Старый `AUDIT.md` был заметно устаревшим. На текущем коде уже **полностью рабочие** как минимум `Platform Settings` и `Pricing Tiers`, а `Winback`, `A/B Tests` и advanced experiment pages уже имеют **реальный wiring** и заметно больше управляющих действий, чем раньше. Главный остаток сейчас — не «подключить хоть что-то», а добить последние truthful workflow gaps: полноценный Studio builder, реальную связку pricing tiers с experiment arms и полировку overview/edit UX на `A/B Tests`.

