# Frontend Audit: недобитый функционал admin dashboard

Дата: 2026-03-07

## Что проверено

- Sidebar-навигация и маршруты dashboard
- Наличие реальных страниц для пунктов меню
- Подключение к данным через `frontend/src/actions/*`
- Признаки незавершённости: статические массивы, placeholder UI, кнопки без обработки

## Рабочие / частично рабочие разделы

Эти страницы уже подключены к данным через `@/actions/*` и выглядят функциональными:

- `Dashboard` — `frontend/src/app/(main)/dashboard/default/page.tsx`
- `Analytics Reports` — `frontend/src/app/(main)/dashboard/analytics/page.tsx`
- `User List` — `frontend/src/app/(main)/dashboard/users/page.tsx`
- `User 360 Profile` — `frontend/src/app/(main)/dashboard/users/[id]/page.tsx`
- `Audit Log` — `frontend/src/app/(main)/dashboard/audit-log/page.tsx`
- `Revenue Ops / Overview` — `frontend/src/app/(main)/dashboard/revenue-ops/page.tsx`
- `Subscriptions` — `frontend/src/app/(main)/dashboard/subscriptions/page.tsx`
- `Transactions` — `frontend/src/app/(main)/dashboard/transactions/page.tsx`
- `Webhooks` — `frontend/src/app/(main)/dashboard/webhooks/page.tsx`

## Недобитые разделы

### 1. Matomo Analytics

Файл: `frontend/src/app/(main)/dashboard/matomo/page.tsx`

Признаки:
- нет подключения к `@/actions/*`
- KPI захардкожены
- iframe-секция помечена как placeholder
- кнопки `Open Matomo` и `Configure` не подключены

Статус: UI-макет, не рабочая интеграция.

### 2. Platform Settings

Файл: `frontend/src/app/(main)/dashboard/settings/page.tsx`

Признаки:
- нет `@/actions/*`
- формы состоят из `Input/Switch/Button`, но без submit/save логики
- нет сохранения настроек

Статус: статическая форма без backend wiring.

### 3. Dunning

Файл: `frontend/src/app/(main)/dashboard/dunning/page.tsx`

Признаки:
- локальный массив `retryRules`
- служебные подписи типа `dunning table`, `grace_periods table`
- кнопки `Save Config` и `Preview Email Template` без логики

Статус: отдельная страница не доведена, при том что dunning-данные уже есть в `Revenue Ops`.

### 4. Winback

Файл: `frontend/src/app/(main)/dashboard/winback/page.tsx`

Признаки:
- локальный массив `campaigns`
- форма редактирования статическая
- `New Campaign`, `Save Campaign`, `Launch Now` не подключены

Статус: UI-заглушка.

### 5. Pricing Tiers

Файл: `frontend/src/app/(main)/dashboard/pricing/page.tsx`

Признаки:
- локальный массив `plans`
- таблица и inline-форма работают только как макет
- `Edit`, `Activate`, `Deactivate`, `Save` без backend-логики

Статус: CRUD для тарифов не реализован.

### 6. A/B Tests

Файл: `frontend/src/app/(main)/dashboard/experiments/page.tsx`

Признаки:
- локальный массив `tests`
- completed tab — placeholder
- управляющие кнопки не подключены

Статус: mock UI.

### 7. Experiment Studio

Файл: `frontend/src/app/(main)/dashboard/experiments/studio/page.tsx`

Признаки:
- конфигуратор без data load/save
- `Save Draft` и `Launch Test` без логики

Статус: не подключён к бэку.

### 8. Bandit Model

Файл: `frontend/src/app/(main)/dashboard/experiments/bandit/page.tsx`

Признаки:
- локальный массив `arms`
- статические метрики и веса
- `Pause`, `Stop Bandit`, `Save Config` без логики

Статус: витринный экран.

### 9. Delayed Feedback

Файл: `frontend/src/app/(main)/dashboard/experiments/feedback/page.tsx`

Признаки:
- локальный массив `questions`
- нет сохранения формы
- reorder/delete кнопки визуальные, но нерабочие

Статус: mock UI.

### 10. Sliding Window

Файл: `frontend/src/app/(main)/dashboard/experiments/sliding-window/page.tsx`

Признаки:
- все показатели захардкожены
- `Save Config` / `Reset Window` не подключены

Статус: статический конфиг-экран.

### 11. Multi-Objective

Файл: `frontend/src/app/(main)/dashboard/experiments/multi-objective/page.tsx`

Признаки:
- локальный массив `objectives`
- статическая таблица Pareto
- нет расчётов/загрузки/сохранения

Статус: макет без функционала.

## Архитектурный сигнал незавершённости

В `frontend/src/actions` отсутствуют action-файлы для:

- `matomo`
- `settings`
- `dunning`
- `winback`
- `pricing`
- `experiments/*`

Это подтверждает, что перечисленные страницы пока не подключены к данным и операциям.

## UX-проблема

В sidebar уже есть поддержка `comingSoon`:

- `frontend/src/app/(main)/dashboard/_components/sidebar/nav-main.tsx`

Но в фактическом меню:

- `frontend/src/navigation/sidebar/use-sidebar-items.ts`

недоделанные разделы не помечены как `comingSoon` и отображаются как полноценные рабочие пункты.

## Частично недоделано, но уже работает

### Subscriptions
- страница живая
- сортировка сейчас делается на фронте
- в коде есть комментарий, что backend пока не отдаёт sort param

### Transactions
- страница живая
- сортировка сейчас делается на фронте
- backend sort API ещё не доведён

## Приоритет на добивку

Рекомендуемый порядок:

1. `Platform Settings`
2. `Pricing Tiers`
3. `Winback`
4. `Experiments` и дочерние разделы
5. `Matomo Analytics`
6. `Dunning` как отдельную страницу либо убрать, либо синхронизировать с `Revenue Ops`

## Быстрый вывод

На фронте недоделаны прежде всего административные конфигурационные и экспериментальные экраны. Основные операционные разделы (`users`, `transactions`, `subscriptions`, `audit log`, `webhooks`, `revenue ops`) уже подключены и выглядят рабочими, хотя местами ещё есть технический долг.