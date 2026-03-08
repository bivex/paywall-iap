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
| Pricing Tiers | `getPricingTiers()` | `createPricingTierAction()`, `updatePricingTierAction()`, `activatePricingTierAction()`, `deactivatePricingTierAction()` | ✅ Работает | standalone CRUD wiring подключён; тот же live manager встроен в Studio, а draft arm builder теперь использует реальный tier↔experiment arm linkage |
| A/B Tests | `getExperiments()` | `createExperimentAction()`, `updateExperimentAction()`, `pauseExperimentAction()`, `resumeExperimentAction()`, `completeExperimentAction()` | 🟡 Частично | список/создание/lifecycle реальные; existing draft experiments теперь truthfully редактируются inline по metadata, а overview timestamps больше не ломают hydration; full arm/tier editing намеренно остаётся в Studio |
| Experiment Studio | `getStudioDashboardFromCookies()` + `/api/admin/studio/*` | `updateExperimentAction()`, `pauseExperimentAction()`, `resumeExperimentAction()`, `completeExperimentAction()`, embedded `PricingTierManager` actions | 🟡 Частично | truthful draft builder уже есть: metadata edit, full arm-set editing, tier-per-arm linkage, lifecycle controls, recommendation history/guards и live pricing manager; дальше остаётся в основном deeper runtime/operator polish |
| Bandit Model | `getBanditDashboardFromCookies()` + `/api/admin/bandit/*` | `updateExperimentAction()` (draft config), `pauseExperimentAction()`, `resumeExperimentAction()`, `completeExperimentAction()` | 🟡 Частично | live stats + lifecycle + draft config edit уже есть; current winner recommendation, guard checks, operator-safe next actions и audit history уже surfaced, но runtime-specific bandit knobs/workbench beyond persisted experiment fields всё ещё отсутствуют |
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
| A/B Tests overview polish | existing draft experiments теперь truthfully редактируются inline по metadata, а timestamps на overview table/lifecycle snippets больше не дают SSR/client hydration mismatch |
| Experiment Studio controls | lifecycle controls для выбранного эксперимента, truthful draft arm builder/edit/save flow, recommendation history/guards и встраивание live `PricingTierManager` |
| Bandit Model controls | lifecycle controls, truthful draft config edit и recommendation history/guards/operator-safe next actions для сохранённых экспериментов |
| Winback management | добавлена реальная деактивация существующих winback campaigns, а не только launch |
| Dunning usability | добавлены операторские filters/search вместо полностью статичного standalone экрана |
| Delayed Feedback tooling | lookup pending rewards по reward id и user id через реальные admin proxy routes |
| Sliding Window tooling | inspect/export live window events через реальный `/api/admin/sliding-window/events` |
| Multi-Objective config wiring | форма больше не стартует из hardcoded defaults — грузит и сохраняет persisted config |
| Studio / pricing truthfulness | pricing tiers редактируются из Studio тем же live CRUD, а draft arm builder умеет реальный arm→tier linkage |
| Recommendation surfacing | `Bandit Model` и `Experiment Studio` теперь показывают persisted winner recommendation layer, guard checks, operator-safe next-step guidance и newest-first recommendation audit history |
| Cold-start local data | есть `scripts/seed_all_test_data.sh`, интеграция в `run_dev.sh`, и детерминированные experiment/bandit fixtures для непустых admin pages |
| UI polish | в A/B Tests исправлено перекрытие long description text в таблице, а experiment overview timestamps теперь hydration-safe; Studio даты и related admin experiment surfaces не откатываются к mismatch |
| Experiment automation foundation | backend теперь имеет единый lifecycle service layer, persisted `automation_policy`, scheduled reconciler, transactional lifecycle audit/idempotency, `latest_lifecycle_audit` и full lifecycle history surface в admin payloads/UI |
| Experiment lifecycle audit history surface | появился реальный `GET /v1/admin/experiments/:id/lifecycle-audit`, Studio snapshot подгружает полную историю, а UI показывает newest-first lifecycle timeline без mock-данных |

## Что реально осталось недоделанным

| Приоритет | Что добивать | Почему |
|---|---|---|
| P1 | Runtime-specific config UX для `Bandit Model` | draft config edit и recommendation surfacing уже есть, но не хватает более глубокого runtime workbench поверх metrics/probabilities/current history surfaces |
| P2 | Safe auto-rollout controls | recommendation layer теперь surfaced end-to-end, но guarded auto-promote / auto-reweight / explicit rollout controls всё ещё не начаты |
| P2 | Winback management beyond launch/deactivate | launch и deactivate работают, но нет полноценного update/edit flow кампаний |
| P2 | Sliding Window config/reset UX | trim и export уже есть, но экран не даёт полноценного runtime/config management |
| P2 | Multi-Objective optimizer workbench | persisted config и score inspection есть, но это ещё не полный decision cockpit |
| P3 | Dunning page product decision | либо делать отдельный config/editor экран, либо честно оставить её как operations mirror Revenue Ops |
| P3 | Matomo dedicated analytics UX | embed через saved settings работает, но page-level KPI/config experience минимален |

### Top-3 ближайших ticket'а (рекомендуемый порядок)

| Порядок | Ticket | Почему сейчас |
|---|---|---|
| 1 | Runtime-specific config UX для `Bandit Model` | после закрытия overview edit и recommendation surfacing это остаётся самым заметным gap на experiment admin surfaces |
| 2 | Safe auto-rollout controls | recommendation/guard layer уже surfaced, так что следующий логичный шаг — не новый read-path, а аккуратные action/policy controls поверх него |
| 3 | Winback management beyond launch/deactivate | winback уже не mock, но без truthful update/edit flow страница всё ещё выглядит заметно менее зрелой, чем experiment surfaces |

## Concrete implementation checklist by file/path (top 3)

### 1) Bandit Model runtime workbench

| File/path | Checklist |
|---|---|
| `frontend/src/app/(main)/dashboard/experiments/bandit/bandit-page-client.tsx` | добавить richer runtime workbench поверх уже имеющихся metrics/probabilities/config/recommendation views |
| `frontend/src/lib/bandit.ts` | при необходимости расширить UI types под richer runtime/recommendation payload |
| `frontend/src/lib/server/bandit-admin.ts` | вернуть дополнительные truthful runtime fields, если они уже существуют в backend contract |
| `backend/internal/interfaces/http/handlers/admin_bandit.go` | безопасно расширять только реально существующие runtime/config surfaces, не выдумывая hidden knobs |

### 2) Safe auto-rollout controls

| File/path | Checklist |
|---|---|
| `frontend/src/app/(main)/dashboard/experiments/bandit/bandit-page-client.tsx` | если добавлять rollout actions, показывать их только как guarded/manual operator controls поверх уже выведенных recommendation guards |
| `frontend/src/app/(main)/dashboard/experiments/studio/studio-page-client.tsx` | держать те же safe next-step semantics рядом с lifecycle/lock controls, не превращая Studio в silent auto-rollout surface |
| `backend/internal/interfaces/http/handlers/admin_bandit.go` + related service layer | вводить rollout/action endpoints только при явных policy guards, audit log и уважении `manual_override` / `locked_until` |
| `backend/tests/integration/*bandit*` | покрыть guard semantics, auditability и explicit operator confirmation paths для rollout-related actions |

### 3) Winback management beyond launch/deactivate

| File/path | Checklist |
|---|---|
| `frontend/src/app/(main)/dashboard/winback/winback-page-client.tsx` | добавить truthful edit/update UX для существующих winback campaigns вместо страницы только для launch/deactivate |
| `frontend/src/actions/winback.ts` | переиспользовать/расширить существующий action layer без дублирующих mutation flows |
| `backend/internal/interfaces/http/handlers/admin_winback.go` + related domain/service layer | отдать/update existing campaign fields только если они уже реально поддерживаются persisted model, без UI-притворства |
| `backend/tests/integration/*winback*` | покрыть update/edit сценарии для существующих кампаний и защиту от невалидных transitions |

## Какие backend-механизмы потребуются для автоматики

Ниже — не «идеальный wish-list», а то, что реально понадобится, если доводить experiment/bandit/admin workflows до автоматического режима. Важно: часть базы уже есть — в проекте уже поднят `worker` на `asynq`, есть `Scheduler`, persisted execution log/idempotency contract для scheduled automation jobs и lifecycle audit/history surface. Но часть bandit maintenance логики всё ещё placeholder-уровня и не даёт построить полноценную автоматику поверх UI.

| Механизм | Что уже есть | Что ещё потребуется добить |
|---|---|---|
| Background worker + scheduler | уже есть `backend/cmd/worker/main.go`, `backend/internal/worker/tasks/tasks.go`, scheduled jobs на `asynq`; experiment automation и maintenance уже реально идут через этот execution layer | дальше добивать не сам контур запуска, а observability, richer repair flows и production-grade maintenance internals |
| Experiment lifecycle reconciler | ручные transitions уже вынесены в общий service/use-case слой, а периодический reconciler уже сканирует `ab_tests` и запускает auto-start / auto-complete | при необходимости следующим этапом добавлять safety-pause rules и более богатые guardrails, а не базовый reconciler с нуля |
| Единый command/service layer для lifecycle | lifecycle transitions уже вынесены в общий domain/service use-case слой и используются как ручными admin actions, так и scheduler automation | дальше держать новые auto-action paths на том же service contract, а не возвращаться к ad hoc логике в handlers |
| Persisted automation policy | у эксперимента уже есть явная persisted `automation_policy` с `manual_override`, `locked_until`, `locked_by`, `lock_reason` и threshold-driven semantics | если развивать auto-rollout дальше, то только расширяя эту policy-модель, а не заводя параллельные hidden flags |
| Очередь/джобы для bandit maintenance | scheduled maintenance в `backend/internal/worker/tasks/currency_asynq.go` проходит через persisted idempotency contract, а `RunMaintenance`/targeted jobs теперь реально делают `process_expired_rewards`, `trim_windows`, `cleanup_old_context_data`, `sync_objective_stats` и expired assignment cleanup через repository-backed paths; targeted operator scopes для cleanup тоже уже есть | дальше развивать maintenance уже как richer decision/runtime layer (например, recommendation/audit/event surfaces), а не возвращаться к placeholder-логике |
| Immutable event / conversion log | теперь есть append-only `experiment_automation_decision_log`, `bandit_conversion_events`, `bandit_assignment_events`, `bandit_impression_events` и `experiment_winner_recommendation_log`; direct reward / delayed conversion / expired pending reward, новые arm assignments, explicit impression calls и admin-facing winner recommendation evaluations уже пишутся в immutable history | следующий backend шаг уже не про event trail, а про guarded action/policy layer поверх уже имеющейся recommendation history |
| Idempotent job execution | scheduler-backed automation и maintenance jobs теперь используют persisted execution log с window-based idempotency key, claim/skip semantics и retry-after-failure | дальше развивать это как единый contract для новых scheduled paths, а не возвращаться к best-effort execution |
| Audit trail для auto-actions | для experiment lifecycle automation уже есть отдельный audit layer: source/reason/transition/time, latest audit в summary payloads и full history endpoint/UI | при расширении автоматики сохранять тот же уровень прозрачности для новых decision paths, а не откатываться к «silent background changes» |
| Reconciliation / repair jobs | есть explicit admin repair path и scheduled background repair reconciler на `asynq` с window-idempotent execution log; explicit repair теперь делает assignment snapshot, создаёт missing `ab_test_arm_stats`, синхронизирует per-experiment `objective stats`, пересчитывает `winner_confidence` и обрабатывает expired pending rewards, а maintenance layer отдельно чистит stale context/expired assignments и даёт targeted operator scopes для этих cleanup paths | coverage автоматики всё ещё не полная: следующий gap уже больше про richer recommendation/decision/event surfaces, чем про базовый cleanup plumbing |
| Experiment arm editing backend | draft experiment update теперь умеет persist full arm-set changes через существующий `PUT /v1/admin/experiments/:id`: add/remove/relabel/reweight arms и обновлять `pricing_tier_id` вместе с metadata | Studio frontend теперь тоже использует этот contract как реальный draft builder, так что следующий шаг уже про polish, а не про wiring |
| Pricing tier linkage model | live pricing tiers уже truthfully связаны с arms через `ab_test_arms.pricing_tier_id`, create/update/read paths уже есть | linkage уже surfaced и в Studio draft workflow; дальше только richer operator surfacing там, где это реально полезно |
| Automation-safe selection policy | bandit runtime уже выбирает arm и кэширует sticky assignment, а admin read-path теперь отдаёт safe winner recommendation по win probability / sample-size / confidence guards | если вводить auto-promotion / auto-winner / auto-rollout, потребуется отдельная policy-логика: когда система только рекомендует winner, а когда реально меняет allocation/status автоматически |
| Manual override + lock semantics | persisted `automation_policy` уже поддерживает `manual_override`, `locked_until`, `locked_by`, `lock_reason`, а admin API/UI уже умеют `lock/unlock` experiment automation | дальше держать это как единый contract: все scheduler-driven automation paths должны уважать как explicit manual override, так и time-bound lock window |
| Observability для автоматики | worker/logging уже присутствуют | нужны метрики и алерты: сколько auto transitions прошло, сколько jobs упало, сколько stale experiments, сколько pending rewards не обработано, сколько window trims skipped |

### Минимальный truthful backend slice для следующего этапа автоматики

Если делать не «всё сразу», а минимальный полезный следующий backend-срез, то приоритет теперь выглядит так:

1. **Расширить runtime workbench** на `Bandit Model` поверх уже существующих truthful метрик и recommendation surfaces.
2. **Ввести safe auto-rollout controls** только поверх уже существующего recommendation layer.
3. **Добить observability** для automation worker path, чтобы rollout/maintenance не оставались "чёрным ящиком".
4. **Полировать оставшиеся partial admin surfaces** (`Winback`, `Sliding Window`, `Multi-Objective`) без отката к mock-поведению.
5. **Держать pricing/tier linkage visible only where it adds real operator value**, а не плодить дублирующие псевдо-editor flows.

Без этих пяти вещей автоматика останется либо UI-имитацией, либо набором хрупких cron-скриптов поверх уже существующих ручных endpoints.

## Backend roadmap по этапам

### Stage 1 — safe automation foundation

Цель: сначала построить безопасную backend-основу, чтобы любая будущая автоматика не жила в `HTTP handler`-ах и не ломала данные при повторных job runs.

| Направление | Что сделать |
|---|---|
| Lifecycle service layer | вынести status transitions и общую валидацию из `admin_experiments.go` в domain/service use-case слой |
| Automation policy model | добавить persisted policy/config для auto-start / auto-stop / confidence/sample thresholds / manual override flags |
| Job idempotency | ввести idempotent execution contract для auto-actions и maintenance jobs |
| Audit trail | писать system-generated lifecycle changes отдельно от user-triggered actions |
| Observability | добавить метрики/логи/alerts для scheduler, stale experiments, failed jobs, skipped runs |

**Результат Stage 1:** backend готов принимать automation rules без дублирования логики и без риска, что cron/job path будет жить отдельно от ручного admin path.

### Stage 2 — experiment auto-lifecycle

Цель: автоматизировать базовый жизненный цикл экспериментов без «магии», только на понятных и проверяемых правилах.

| Направление | Что сделать |
|---|---|
| Experiment reconciler job | добавить scheduled worker job, который сканирует `ab_tests` и применяет automation policy |
| Auto-start / auto-complete | автоматически запускать draft experiments по времени и завершать running experiments по end date / sample / confidence rules |
| Manual override semantics | если эксперимент вручную paused/locked оператором, job не должен его трогать |
| Repair / reconciliation | добавить safe backfill path для пересчёта derived state после сбоев или пропущенных runs |
| Admin visibility | вернуть в admin payload/system notes причину auto-transition: какая rule сработала и когда |

**Результат Stage 2:** experiments умеют жить не только через ручные кнопки в UI, но и через предсказуемый scheduler-driven lifecycle.

### Stage 3 — bandit auto-ops / auto-winner

Цель: довести bandit automation до реального production-friendly режима, где backend не просто хранит метрики, а умеет безопасно сопровождать runtime decisions.

| Направление | Что сделать |
|---|---|
| Real maintenance jobs | ✅ уже repository-backed: `process_expired_rewards`, `trim_windows`, `cleanup_old_context_data`, `sync_objective_stats`, expired assignment cleanup, structured maintenance summary + idempotent scheduler execution |
| Immutable decisions/conversions log | ✅ append-only история automation decisions, reward-resolution событий, arm assignments, impressions и winner recommendation evaluations уже есть; дальше новые slices уже не про базовый event trail |
| Winner recommendation policy | 🟡 уже есть admin-facing recommendation layer, append-only recommendation audit trail/history и UI surfacing в `Bandit Model`/`Experiment Studio`; дальше расширять policy/action layer, не смешивая это с auto-rollout |
| Safe rollout controls | только после накопления audit log и policy guards — вводить auto-promote / auto-reweight / auto-stop loser flows |
| Pricing/arm linkage | ✅ persisted arm↔tier linkage model и Studio draft workflow уже есть; дальше использовать этот metadata surface только там, где он реально нужен automation/operator layers |

**Результат Stage 3:** появляется не просто «bandit dashboard», а реальная backend-автоматика для эксплуатации и принятия решений с контролируемым риском.

### Рекомендуемый порядок реализации

1. **Stage 1 полностью**
2. Из Stage 2: `reconciler + auto-start/auto-complete + manual override`
3. Из Stage 3: `real maintenance jobs + immutable event log + recommendation layer`
4. Только потом — `auto-winner / auto-rollout / deeper optimization`

Иначе есть высокий риск построить внешне красивую автоматику, которая будет опираться на неполный event trail, placeholder jobs и расходящиеся manual/system code paths.

## Ticket-ready backend backlog со статусами

Легенда статусов:

- ✅ Done — уже реализовано достаточно, как отдельный ticket закрывать не нужно
- 🟡 Partial — база есть, но как отдельный этап ещё не добито
- ⚪ Not started — явной реализации под задачу пока нет

| Stage | Ticket | Приоритет | Статус | Что должно получиться |
|---|---|---|---|---|
| Stage 1 | Вынести experiment lifecycle в service/use-case слой | P1 | ✅ Done | manual actions и scheduler automation используют единый transition engine без дублирования логики в HTTP handlers |
| Stage 1 | Persisted automation policy для `ab_tests` | P1 | ✅ Done | experiment domain хранит persisted правила auto-start / auto-complete / override и связанные thresholds |
| Stage 1 | Idempotent contract для automation jobs | P1 | ✅ Done | scheduler-backed automation/maintenance jobs используют persisted execution log с unique idempotency key, window-based claim/skip semantics и retry-after-failure |
| Stage 1 | Audit trail для system-triggered lifecycle changes | P1 | ✅ Done | lifecycle changes прозрачно фиксируют source, reason, transition и timestamp для system-triggered paths |
| Stage 1 | Observability для automation worker path | P2 | 🟡 Partial | structured worker logs и persisted job-run log уже есть, но отдельных метрик, dashboards и alerts по failed/skipped/stale automation runs пока нет |
| Stage 2 | Scheduled experiment reconciler job | P1 | ✅ Done | периодический worker scan-ит `ab_tests` и применяет automation policy через общий lifecycle service |
| Stage 2 | Auto-start / auto-complete rules | P1 | ✅ Done | experiments автоматически стартуют и завершаются по времени, sample-size и confidence-driven rules |
| Stage 2 | Manual override / lock semantics | P1 | ✅ Done | persisted `automation_policy` теперь поддерживает `manual_override`, `locked_until`, `locked_by`, `lock_reason`, отдельные admin `lock/unlock` flows и reconciler уважает как explicit manual lock, так и time-bound lock window |
| Stage 2 | Reconciliation / repair job для derived experiment state | P2 | ✅ Done | explicit admin `repair` path и scheduled/background repair reconciler на `asynq` с window-idempotent execution теперь последовательно покрывают assignment snapshot, missing arm-stats backfill, expired pending reward processing, per-experiment objective stats sync и `winner_confidence` recalculation без stale objective-state после repair |
| Stage 2 | Admin-visible reason codes для auto-transitions | P2 | ✅ Done | admin payload/UI показывает, каким rule и по какой причине система перевела experiment в новый status |
| Stage 2 | Full lifecycle audit history UI/API surface | P2 | ✅ Done | admin API и Studio UI отдают полный newest-first lifecycle audit trail по experiment без mock-данных |
| Stage 3 | Repository-backed bandit maintenance jobs | P1 | ✅ Done | scheduler wiring и idempotent execution уже есть, а maintenance layer теперь реально закрывает expired rewards, currency refresh, `trim_windows`, context cleanup, objective stats sync и expired assignment cleanup через repository-backed paths |
| Stage 3 | Immutable conversions / decisions log | P1 | ✅ Done | append-only `experiment_automation_decision_log`, `bandit_conversion_events`, `bandit_assignment_events`, `bandit_impression_events` и `experiment_winner_recommendation_log` уже покрывают lifecycle/runtime/recommendation history |
| Stage 3 | Winner recommendation policy | P2 | 🟡 Partial | recommendation layer уже считает candidate winner, пишет append-only audit trail/history и truthfully surfaced в `Bandit Model`/`Experiment Studio`; дальше нужны guarded action/policy controls поверх этого слоя |
| Stage 3 | Safe auto-rollout controls | P2 | ⚪ Not started | после recommendation layer и UI surfacing нужны guarded auto-promote / auto-reweight flows с явными safety controls, auditability и respect for `manual_override` / `locked_until` |
| Stage 3 | Persisted pricing tier ↔ arm linkage model | P1 | ✅ Done | `ab_test_arms` truthfully хранят `pricing_tier_id`, admin payload/read path возвращает linkage, draft update contract его сохраняет, а Studio builder уже использует это end-to-end |

### Что уже можно считать опорой, а не отдельными backlog-задачами

| Механизм | Статус | Комментарий |
|---|---|---|
| `asynq` worker + scheduler | ✅ Done | базовый execution layer уже есть в проекте |
| Ручные experiment lifecycle endpoints | ✅ Done | `pause/resume/complete` уже работают и пригодятся как референс для service extraction |
| Bandit maintenance scheduling hooks | ✅ Done | scheduler wiring, persisted idempotency и repository-backed maintenance handlers уже есть; дальше развитие идёт в recommendation/rollout layer, а не в базовые hooks |
| Admin experiment metadata update | ✅ Done | draft metadata и related overview/studio editing уже можно сохранять через реальный update contract, это полезная база для будущего automation policy UI |

## Cross-cutting замечания

| Наблюдение | Статус |
|---|---|
| Механизм `comingSoon` существует | да |
| Частично готовые страницы явно помечены как `comingSoon` | нет — `comingSoonUrls` пустой, поэтому partial pages выглядят как fully ready |
| `Subscriptions` и `Transactions` всё ещё имеют технический долг по sort | да, сортировка местами остаётся на фронте |
| `Delayed Feedback` probe для pending-by-id больше не должен считаться broken из-за ожидаемого `404` | да, это уже исправлено |
| Локальный cold-start для demo/admin страниц теперь воспроизводим | да — через `scripts/seed_all_test_data.sh`, интегрированный в `run_dev.sh` |
| Playwright smoke по seeded experiment pages уже проходили | да — страницы грузятся с непустыми данными; hydration mismatch на `/dashboard/experiments` timestamps больше не должен считаться актуальным открытым gap |

## Короткий вывод

Старый `AUDIT.md` был заметно устаревшим. На текущем коде уже **полностью рабочие** как минимум `Platform Settings` и `Pricing Tiers`, а `Winback`, `A/B Tests` и advanced experiment pages уже имеют **реальный wiring**, scheduler-backed experiment automation и lifecycle audit/history surface. После добивки truthful draft builder в `Experiment Studio`, overview/edit polish в `A/B Tests` и guarded recommendation surfacing в `Bandit Model`/`Studio` главный остаток сейчас — не «подключить хоть что-то», а добивать более глубокий runtime/operator слой: `Bandit Model` workbench, guarded rollout controls и оставшиеся partial admin surfaces вроде `Winback`.

