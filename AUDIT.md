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
| Experiment automation foundation | backend теперь имеет единый lifecycle service layer, persisted `automation_policy`, scheduled reconciler, transactional lifecycle audit/idempotency, `latest_lifecycle_audit` и full lifecycle history surface в admin payloads/UI |
| Experiment lifecycle audit history surface | появился реальный `GET /v1/admin/experiments/:id/lifecycle-audit`, Studio snapshot подгружает полную историю, а UI показывает newest-first lifecycle timeline без mock-данных |

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

### Top-3 ближайших ticket'а (рекомендуемый порядок)

| Порядок | Ticket | Почему сейчас |
|---|---|---|
| 1 | True builder workflow для `Experiment Studio` | это самый заметный truthful UX gap: metadata и lifecycle уже есть, но Studio всё ещё не собирает experiment как полноценный draft/builder workflow |
| 2 | Persisted pricing tier ↔ arm linkage model | без этого tiers и arms остаются рядом, но не становятся одной доменной моделью; это блокирует честный experiment builder и rollout story |
| 3 | Remaining repair/self-heal coverage | immutable log и recommendation layer уже появились как backend foundation; следующий backend шаг — добить оставшиеся derived surfaces и reconciliation gaps |

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

## Какие backend-механизмы потребуются для автоматики

Ниже — не «идеальный wish-list», а то, что реально понадобится, если доводить experiment/bandit/admin workflows до автоматического режима. Важно: часть базы уже есть — в проекте уже поднят `worker` на `asynq`, есть `Scheduler`, persisted execution log/idempotency contract для scheduled automation jobs и lifecycle audit/history surface. Но часть bandit maintenance логики всё ещё placeholder-уровня и не даёт построить полноценную автоматику поверх UI.

| Механизм | Что уже есть | Что ещё потребуется добить |
|---|---|---|
| Background worker + scheduler | уже есть `backend/cmd/worker/main.go`, `backend/internal/worker/tasks/tasks.go`, scheduled jobs на `asynq`; experiment automation и maintenance уже реально идут через этот execution layer | дальше добивать не сам контур запуска, а observability, richer repair flows и production-grade maintenance internals |
| Experiment lifecycle reconciler | ручные transitions уже вынесены в общий service/use-case слой, а периодический reconciler уже сканирует `ab_tests` и запускает auto-start / auto-complete | при необходимости следующим этапом добавлять safety-pause rules и более богатые guardrails, а не базовый reconciler с нуля |
| Единый command/service layer для lifecycle | сейчас lifecycle в основном живёт в HTTP handler-логике | вынести transitions в domain/service слой, чтобы и manual UI actions, и cron/worker automation использовали один и тот же код с одинаковой валидацией переходов |
| Persisted automation policy | сейчас у эксперимента есть status/algorithm/sample/confidence, но нет явной модели automation rules | нужна явная конфигурация: auto-start, auto-stop by end date, auto-complete by sample size/confidence, safety thresholds, manual override flags |
| Очередь/джобы для bandit maintenance | scheduled maintenance в `backend/internal/worker/tasks/currency_asynq.go` проходит через persisted idempotency contract, а `RunMaintenance`/targeted jobs теперь реально делают `process_expired_rewards`, `trim_windows`, `cleanup_old_context_data`, `sync_objective_stats` и expired assignment cleanup через repository-backed paths | дальше развивать maintenance уже как richer decision/runtime layer (например, win-probability / recommendation jobs), а не возвращаться к placeholder-логике |
| Immutable event / conversion log | теперь есть append-only `experiment_automation_decision_log` и `bandit_conversion_events`, direct reward / delayed conversion / expired pending reward уже пишутся в immutable history, а delayed conversion path стал реально обновлять arm stats | дальше расширять event trail до assignments / impressions и richer recommendation events, чтобы автоматика опиралась не только на агрегаты |
| Idempotent job execution | scheduler-backed automation и maintenance jobs теперь используют persisted execution log с window-based idempotency key, claim/skip semantics и retry-after-failure | дальше развивать это как единый contract для новых scheduled paths, а не возвращаться к best-effort execution |
| Audit trail для auto-actions | для experiment lifecycle automation уже есть отдельный audit layer: source/reason/transition/time, latest audit в summary payloads и full history endpoint/UI | при расширении автоматики сохранять тот же уровень прозрачности для новых decision paths, а не откатываться к «silent background changes» |
| Reconciliation / repair jobs | есть explicit admin repair path и scheduled background repair reconciler на `asynq` с window-idempotent execution log; вместе с maintenance layer они уже делают assignment snapshot, создают missing `ab_test_arm_stats`, пересчитывают `winner_confidence`, синхронизируют `objective stats`, обрабатывают expired pending rewards и чистят stale context/expired assignments | coverage derived state всё ещё не полная: за пределами текущего repair/maintenance scope остаются другие derived surfaces и richer recommendation/decision paths |
| Experiment arm editing backend | create experiment с arms уже есть, edit существующего draft пока ограничен metadata-only update | для настоящего Studio builder понадобится persisted arm CRUD + server-side validation суммарных weight/control arm invariants |
| Pricing tier linkage model | live pricing tiers уже есть, но truthful linkage tier ↔ arm пока отсутствует | если Studio должен автоматизировать pricing-experiment workflows, нужна отдельная persisted linkage model/table, а не просто соседние UI-блоки |
| Automation-safe selection policy | bandit runtime уже выбирает arm и кэширует sticky assignment, а admin read-path теперь отдаёт safe winner recommendation по win probability / sample-size / confidence guards | если вводить auto-promotion / auto-winner / auto-rollout, потребуется отдельная policy-логика: когда система только рекомендует winner, а когда реально меняет allocation/status автоматически |
| Manual override + lock semantics | persisted `automation_policy` уже поддерживает `manual_override`, `locked_until`, `locked_by`, `lock_reason`, а admin API/UI уже умеют `lock/unlock` experiment automation | дальше держать это как единый contract: все scheduler-driven automation paths должны уважать как explicit manual override, так и time-bound lock window |
| Observability для автоматики | worker/logging уже присутствуют | нужны метрики и алерты: сколько auto transitions прошло, сколько jobs упало, сколько stale experiments, сколько pending rewards не обработано, сколько window trims skipped |

### Минимальный truthful backend slice для следующего этапа автоматики

Если делать не «всё сразу», а минимальный полезный следующий backend-срез, то приоритет теперь выглядит так:

1. **Расширить repair/self-heal coverage** на оставшиеся derived surfaces вне текущего repair/maintenance scope.
2. **Добить persisted arm CRUD + validation** для truthful experiment builder workflow.
3. **Ввести persisted pricing tier ↔ arm linkage model** для реального pricing-experiment orchestration.
4. **Расширить immutable event trail** до assignments / impressions / richer decision events.
5. **Ввести safe auto-rollout controls** только поверх уже существующего recommendation layer.

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
| Immutable decisions/conversions log | 🟡 уже есть append-only история automation decisions и reward-resolution событий; дальше расширять до assignment/impression trail |
| Winner recommendation policy | 🟡 уже есть admin-facing recommendation layer: backend считает winning arm по win probabilities и применяет sample-size / confidence guards; дальше расширять surfacing/logging, не смешивая это с auto-rollout |
| Safe rollout controls | только после накопления audit log и policy guards — вводить auto-promote / auto-reweight / auto-stop loser flows |
| Pricing/arm linkage | если bandit/Studio должны автоматически работать с pricing tiers, сначала ввести persisted arm↔tier linkage model |

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
| Stage 2 | Reconciliation / repair job для derived experiment state | P2 | 🟡 Partial | есть explicit admin `repair` path и scheduled/background repair reconciler на `asynq` с window-idempotent execution: они делают assignment snapshot, создают missing arm-stats rows, пересчитывают `winner_confidence` и обрабатывают expired pending rewards из persisted state, но coverage всего derived surface пока неполная |
| Stage 2 | Admin-visible reason codes для auto-transitions | P2 | ✅ Done | admin payload/UI показывает, каким rule и по какой причине система перевела experiment в новый status |
| Stage 2 | Full lifecycle audit history UI/API surface | P2 | ✅ Done | admin API и Studio UI отдают полный newest-first lifecycle audit trail по experiment без mock-данных |
| Stage 3 | Repository-backed bandit maintenance jobs | P1 | ✅ Done | scheduler wiring и idempotent execution уже есть, а maintenance layer теперь реально закрывает expired rewards, currency refresh, `trim_windows`, context cleanup, objective stats sync и expired assignment cleanup через repository-backed paths |
| Stage 3 | Immutable conversions / decisions log | P1 | 🟡 Partial | появились append-only `experiment_automation_decision_log` и `bandit_conversion_events`, direct reward / delayed conversion / expired pending reward теперь пишутся в immutable history, а delayed conversion path стал truthful; но полного assignment/impression trail и richer recommendation events пока нет |
| Stage 3 | Winner recommendation policy | P2 | 🟡 Partial | появился read-only recommendation layer в admin payload: backend считает candidate winner по win probabilities и отдаёт recommendation с sample-size / confidence guards; но отдельного recommendation audit trail, richer UI surfacing и auto-rollout controls пока нет |
| Stage 3 | Safe auto-rollout controls | P2 | ⚪ Not started | после recommendation layer нужны guarded auto-promote / auto-reweight flows с явными safety controls |
| Stage 3 | Persisted pricing tier ↔ arm linkage model | P1 | ⚪ Not started | pricing tiers должны стать реальной частью experiment/bandit domain модели, а не только соседним CRUD UI |

### Что уже можно считать опорой, а не отдельными backlog-задачами

| Механизм | Статус | Комментарий |
|---|---|---|
| `asynq` worker + scheduler | ✅ Done | базовый execution layer уже есть в проекте |
| Ручные experiment lifecycle endpoints | ✅ Done | `pause/resume/complete` уже работают и пригодятся как референс для service extraction |
| Bandit maintenance scheduling hooks | ✅ Done | scheduler wiring, persisted idempotency и repository-backed maintenance handlers уже есть; дальше развитие идёт в recommendation/rollout layer, а не в базовые hooks |
| Admin experiment metadata update | ✅ Done | draft metadata уже можно сохранять, это полезная база для будущего automation policy UI |

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

Старый `AUDIT.md` был заметно устаревшим. На текущем коде уже **полностью рабочие** как минимум `Platform Settings` и `Pricing Tiers`, а `Winback`, `A/B Tests` и advanced experiment pages уже имеют **реальный wiring**, scheduler-backed experiment automation и lifecycle audit/history surface. Главный остаток сейчас — не «подключить хоть что-то», а добить последние truthful workflow gaps: полноценный Studio builder, реальную связку pricing tiers с experiment arms и полировку overview/edit UX на `A/B Tests`.

