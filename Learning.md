# 📚 Learning.md — Как строился этот проект: шаг за шагом

> Документ описывает весь путь от идеи до 508 виджетов в StarUML — инструмент, подход, ошибки и инсайты.

---

## 📑 Содержание

1. [Старт: что такое MCP и StarUML API](#1-старт)
2. [Первый прототип: 5 экранов из wireframes.md](#2-первый-прототип)
3. [Рефакторинг: новый spec-документ](#3-рефакторинг-spec)
4. [SQL merge script: единая схема](#4-sql-merge)
5. [Строим 20 экранов: приоритизация](#5-строим-20-экранов)
6. [Priority 1 (Screens 5–8): бизнес-ядро](#6-priority-1)
7. [Priority 2 (Screens 9–12): рост и операции](#7-priority-2)
8. [Priority 3 (Screens 13–16): отладка и advanced](#8-priority-3)
9. [Priority 4 (Screens 17–20): настройки и инфра](#9-priority-4)
10. [Навигация: grouped sidebar + breadcrumbs](#10-навигация)
11. [MD Parser: конвертируем .mdj в Markdown](#11-md-parser)
12. [Ключевые технические инсайты](#12-технические-инсайты)
13. [Архитектура финального результата](#13-архитектура)

---

## 1. Старт

### Что такое MCP?

MCP (Model Context Protocol) — протокол, позволяющий AI-агенту вызывать внешние инструменты напрямую из чата. В этом проекте использовались два MCP-сервера:

- **`calc-mcp`** — калькулятор, первый тест: `2+3 = 5` ✅
- **`staruml-wireframe`** — управление StarUML через REST API на `localhost:12345`

### Как работает StarUML MCP API

```
Агент → HTTP запрос → StarUML Controller (localhost:12345) → StarUML App
```

Ключевые эндпоинты:
```
POST /api/wireframe/diagrams          → создать диаграмму
POST /api/wireframe/frames            → создать фрейм (экран)
POST /api/wireframe/widgets           → добавить виджет
GET  /api/diagrams/{id}/views         → получить view IDs
POST /api/project/save                → сохранить .mdj
```

---

## 2. Первый прототип

### Исходный файл

Начали с готового `docs/wireframes.md` (824 строки) — документ описывал 5 разделов:
1. Authentication & System Health
2. Dashboard & KPI Metrics
3. Users & Customers Management
4. Subscriptions Management
5. Transactions & Purchases

### Что сделали

```
project_new → wireframe_create_diagram → 5× wireframe_create_frame → N× wireframe_create_widget
```

Создали **5 WFDesktopFrame** в сетке 3×2:

```
[Frame 1: Auth]       [Frame 2: Dashboard]  [Frame 3: Users]
(x=50, y=50)          (x=950, y=50)         (x=1950, y=50)

[Frame 4: Subscriptions]  [Frame 5: Transactions]
(x=50, y=850)             (x=950, y=850)
```

**118 виджетов**, сохранили в `docs/paywall-iap-wireframes.mdj`.

### Первый критический баг 🐛

> **Проблема**: виджеты создавались вне фрейма — координаты были относительными, а должны быть абсолютными.

**Урок**: координаты в StarUML MCP — **всегда абсолютные** на холсте, не относительные к фрейму.

---

## 3. Рефакторинг: новый spec-документ

### Проблема первого прототипа

Первый `wireframes.md` был написан без привязки к реальной БД — таблицы, колонки, индексы не совпадали со схемой.

### Решение

Создали `docs/Wireframes_Rethink.md` — полностью переработанный spec с:
- ASCII-макетами каждого экрана
- Маппингом каждого UI-элемента на реальную колонку БД
- Описанием взаимодействий и edge cases
- Таблицей компонентов для переиспользования

### Новый проект StarUML

```
project_new → diagram "Wireframes_Rethink" → 4 фрейма (2×2 сетка)
```

Фреймы Row 1 (y=50):
- Frame 1: Admin Dashboard (x=50)
- Frame 2: User 360° Profile (x=1150)

Фреймы Row 2 (y=1000):
- Frame 3: Experiment Studio (x=50)
- Frame 4: Revenue Ops Center (x=1150)

**94 виджета**, сохранили в `docs/Wireframes_Rethink.mdj`.

---

## 4. SQL Merge Script

### Зачем

17 отдельных migration файлов (`001_create_users.up.sql` ... `017_bandit_advanced.up.sql`) сложно читать. Нужна единая схема как reference.

### Скрипт

```bash
# scripts/merge_migrations.sh
find backend/migrations -name "*.up.sql" | sort -V | while read f; do
    echo "-- === $f ==="
    cat "$f"
    echo ""
done > backend/migrations/schema_merged.sql
```

**Результат**: `schema_merged.sql` — 797 строк, 17 миграций, ~22 таблицы.

### Таблицы в схеме

| Группа | Таблицы |
|--------|---------|
| Core | `users`, `subscriptions`, `transactions` |
| Billing | `pricing_tiers`, `grace_periods`, `dunning` |
| Growth | `winback_offers`, `ab_tests`, `ab_test_arms` |
| Analytics | `analytics_aggregates`, `matomo_staged_events` |
| Admin | `admin_audit_log`, `webhook_events` |
| Bandit | `bandit_arm_context_model`, `bandit_user_context`, `bandit_pending_rewards`, `bandit_window_events`, `bandit_arm_objective_stats` |

---

## 5. Строим 20 экранов: приоритизация

### Методология приоритизации

После 4 базовых экранов составили бэклог из 16 экранов с разбивкой по ценности:

```
🔥 Priority 1 — Core Business Flows (Must Have)
   5. Pricing Tiers Manager      → таблица pricing_tiers
   6. User List + Filters        → users + subscriptions
   7. Subscription Management    → subscriptions + dunning
   8. Transaction Reconciliation → transactions

🚀 Priority 2 — Growth & Ops (Should Have)
   9.  Winback Campaign Builder  → winback_offers + ab_tests
   10. Dunning Campaign Config   → dunning + grace_periods
   11. Analytics Reports         → analytics_aggregates
   12. A/B Test Discovery        → ab_tests + ab_test_arms

🔧 Priority 3 — Debug & Advanced (Could Have)
   13. Webhook Event Inspector   → webhook_events
   14. Matomo Queue Monitor      → matomo_staged_events
   15. Bandit Context Inspector  → bandit_arm_context_model
   16. Admin Audit Log           → admin_audit_log

⚙️ Priority 4 — Settings & Infrastructure
   17. Platform Settings         → currency_rates + ab_tests config
   18. Delayed Feedback Monitor  → bandit_pending_rewards
   19. Sliding Window Analytics  → bandit_window_events
   20. Multi-Objective Config    → bandit_arm_objective_stats
```

### Layout-сетка для 20 фреймов

```
Row  1 (y=50):   Frame  1 (x=50)   Frame  2 (x=1150)
Row  2 (y=1000): Frame  3 (x=50)   Frame  4 (x=1150)
Row  3 (y=1950): Frame  5 (x=50)   Frame  6 (x=1150)
Row  4 (y=2900): Frame  7 (x=50)   Frame  8 (x=1150)
Row  5 (y=3950): Frame  9 (x=50)   Frame 10 (x=1150)
Row  6 (y=4900): Frame 11 (x=50)   Frame 12 (x=1150)
Row  7 (y=5950): Frame 13 (x=50)   Frame 14 (x=1150)
Row  8 (y=6900): Frame 15 (x=50)   Frame 16 (x=1150)
Row  9 (y=7950): Frame 17 (x=50)   Frame 18 (x=1150)
Row 10 (y=8900): Frame 19 (x=50)   Frame 20 (x=1150)
```

Каждый фрейм: **1000×850px**, gap между колонками **100px**, между рядами **50px**.

---

## 6. Priority 1 (Screens 5–8)

### Screen 5: Pricing Tiers Manager

**Таблица**: `pricing_tiers`

Компоненты:
- 3 карточки тарифов (Basic / Pro / Enterprise) с ценой/фичами
- Inline-редактор при клике Edit: Name, Description, Monthly/Annual Price, Currency, Active toggle, Features JSONB
- Platform Preview tabs: iOS / Android / Web (разный вид paywall)
- Soft Delete (не физическое удаление)

```
[Basic Card]  [Pro Card]  [Enterprise Card]
[Edit] [Del]  [Edit] [Del]  [Edit] [Del]

EDIT TIER FORM:
Name: ___________  Description: ___________
Monthly: ___  Annual: ___  Currency: [USD▼]  Active: [●]
Features JSONB: { "max_devices": 3, "offline": true }
[iOS] [Android] [Web]      [Cancel] [Save Tier]
```

### Screen 6: User List + Advanced Filters

**Таблицы**: `users`, `subscriptions`

Компоненты:
- Search по email/user_id + фильтры: Platform, Role, LTV range, Sub Status
- Bulk actions: Select All, Cancel Selected, Grant Access
- Таблица с badge-колонками: Platform (🍎/🤖), Role, Status
- Export CSV кнопка
- Pagination

### Screen 7: Subscription Management List

**Таблицы**: `subscriptions`, `grace_periods`, `dunning`

Ключевая фича — колонка **"Grace/Dunning"** с индикаторами:
- `🔶 Grace (2d left)` — из `grace_periods.expires_at`
- `⚠️ Dunning (3/5)` — из `dunning.attempt_count / max_attempts`

### Screen 8: Transaction Reconciliation

**Таблицы**: `transactions`, `users`

KPI-бар сверху: Total | Success | Failed | Refunded

Двойной поиск:
- `receipt_hash` — для верификации IAP
- `provider_tx_id` — для сверки со Stripe/Apple/Google

---

## 7. Priority 2 (Screens 9–12)

### Screen 9: Winback Campaign Builder

Split-layout: левая панель (список кампаний) + правая (редактор).

Редактор кампании:
```
Campaign Name: ___________
Discount Type: [%▼]  Value: [20]
Targeting: Churned > [30] days, Platform: [All▼]
Expires At: [2026-04-01]
A/B Test: [●ON]  Arm A: [Control]  Arm B: [Discount20]
Confidence Threshold: [0.95]

PREVIEW:
"Come back! 20% off Pro Annual"
[Claim Offer]    Eligible users: 1,247
[Cancel] [Save Draft] [Launch Campaign]
```

### Screen 10: Dunning Campaign Config

Конфигурация retry-логики через таблицу:

```
Attempt | Interval | Action         | Notification Template
   1    |  3 days  | Retry charge   | email_dunning_1
   2    |  7 days  | Retry + notify | email_dunning_2
   3    | 14 days  | Final warning  | email_final_warning
   4    | 21 days  | Cancel sub     | email_cancelled
```

ASCII Flow Preview:
```
Failed Charge → 3d retry → 7d retry → 14d retry → Grace Period → Cancel
```

### Screen 11: Analytics Reports Dashboard

4 KPI-карточки + 3 графика (текстовые, ASCII-art):
- MRR Trend (бар-чарт по месяцам)
- Churn by Platform (горизонтальные бары)
- Revenue by Plan (breakdown)

Drill-down таблица с `dimensions JSONB`.

### Screen 12: A/B Test Discovery List

Карточки экспериментов (не таблица!) — каждая показывает:
- Тип (Pricing / Winback / Onboarding)
- Статус (Running 🟢 / Draft / Completed)
- Ключевые метрики: samples, winner_confidence
- Inline actions: Stop Test / View Details / Edit Draft

---

## 8. Priority 3 (Screens 13–16)

### Screen 13: Webhook Event Inspector

Split-layout с двумя панелями:

**Левая**: список событий с фильтрами Status + Provider

**Правая**: детали события
```
Event ID:    evt_abc123
Type:        customer.subscription.updated
Provider:    Stripe
Status:      ✅ processed
Retry:       0/3
Idempotency: stripe_evt_abc123

PAYLOAD:
{
  "event_type": "subscription.updated",
  "subscription_id": "sub_xyz",
  "status": "active",
  ...
}

[Manual Replay] [Mark Processed] [Check Idempotency]
```

### Screen 15: Bandit Context Model Inspector

Самый технически сложный экран:

```
User ID: [usr_550e8400____]  [Load Context]

USER CONTEXT (bandit_user_context):
platform:        "ios"
ltv_bucket:      "high"
sub_status:      "churned"
days_since_churn: 45
country:         "US"

FEATURE IMPORTANCE (theta vector):
ltv_bucket:       ████████  0.82
sub_status:       ██████    0.61
days_since_churn: ████      0.43
platform:         ███       0.31

A MATRIX: 5×5 diagonal, b VECTOR: [0.82, 0.61, 0.43, 0.31, 0.12]
θ (theta): [0.74, 0.58, 0.39, 0.27, 0.11]

UCB Score: 0.847  (θᵀx + α√(xᵀA⁻¹x))
```

---

## 9. Priority 4 (Screens 17–20)

### Screen 19: Sliding Window Analytics

Объяснение концепции скользящего окна для bandit:
```
Window Type: [tumbling▼]  Size: [7] days
Events: assign → reward → expire (timeline)
[Reset Window Stats] [Export Events CSV]
```

### Screen 20: Multi-Objective Config

Слайдеры весов с live-валидацией:
```
Revenue Weight:    ████████░░  0.50
Retention Weight:  ████░░░░░░  0.30
Engagement Weight: ██░░░░░░░░  0.20
                   ─────────────────
Total:             1.00 ✅

Hybrid Score = 0.50 × revenue + 0.30 × retention + 0.20 × engagement

[A/B Test These Weights ●] [Save Config] [Reset to Defaults]
```

---

## 10. Навигация: Grouped Sidebar + Breadcrumbs

### Проблема исходного дизайна

Плоский сайдбар из 7 пунктов не масштабировался на 20 экранов:
```
❌ До:
NAVIGATION
▶ Dashboard
   Users
   Subscriptions
   Experiments (Bandit)
   Revenue Ops
   Pricing
   Settings
```

### Решение: 5 доменных групп

```
✅ После:
📊 MONITORING
  ▶ 🏠 Dashboard
  📈 Analytics Reports
  🔔 Alert Center
─────────────────
👥 USER MANAGEMENT
  👥 User List
  🔍 User 360°
  📋 Audit Log
─────────────────
💰 REVENUE OPS
  💳 Subscriptions
  💸 Transactions
  ⚠️ Dunning 🔴3
  🎁 Winback
─────────────────
🧪 EXPERIMENTS
  🧪 A/B Tests 🟢2
  ⚙️ Exp Studio
  🧠 Bandit
─────────────────
⚙️ CONFIG
  💰 Pricing Tiers
  🌐 Platform
  📡 Webhook ⚠️12
  🚩 Feature Flags
```

### Notification Badges (привязка к БД)

| Badge | SQL source |
|-------|-----------|
| `🔴3` на Dunning | `SELECT COUNT(*) FROM dunning WHERE status IN ('pending','in_progress')` |
| `🟢2` на A/B Tests | `SELECT COUNT(*) FROM ab_tests WHERE winner_confidence >= confidence_threshold` |
| `⚠️12` на Webhook | `SELECT COUNT(*) FROM webhook_events WHERE processed_at IS NULL` |

### Breadcrumbs

Каждый экран получил `WFText` (серый цвет `#6C757D`) в header:
```
🏠 Home  ›  💰 Revenue Ops  ›  Dunning Campaign Config
```

### Как добавили на 20 фреймов

Для каждого фрейма:
1. Создали `WFPanel` (185×390px, `fillColor: #F8F9FA`) — grouped nav
2. Создали `WFText` в header — breadcrumb
3. `▶` маркер активного пункта уникален для каждого экрана

**Итого 40 новых виджетов** (20 panels + 20 breadcrumbs).

---

## 11. MD Parser

### Что такое `uml_wireframe_parser.py`

Python CLI-скрипт, конвертирующий `.mdj` → `.md`.

```bash
python3 scripts/uml_wireframe_parser.py \
  docs/Wireframes_Rethink.mdj \
  docs/Wireframes_Rethink_parsed.md
```

### Как работает

1. Парсит `.mdj` как JSON (это просто JSON файл!)
2. Находит элементы `_type == "WFWireframeDiagram"`
3. Итерирует `ownedViews` — все виджеты на диаграмме
4. Извлекает `_type`, позицию (`left`, `top`, `width`, `height`), label из `subViews[LabelView].text` или `name`
5. Генерирует Markdown-таблицу

### Структура .mdj файла

```json
{
  "_type": "Project",
  "ownedElements": [
    {
      "_type": "UMLPackage",
      "ownedElements": [
        {
          "_type": "WFWireframeDiagram",
          "ownedViews": [
            {
              "_type": "WFDesktopFrameView",
              "_id": "AAAAAAGc...",
              "left": 50, "top": 50,
              "width": 1000, "height": 850
            },
            ...
          ]
        }
      ]
    }
  ]
}
```

### Результат

```
Found 1 wireframe diagram(s)
Wireframes_Rethink: 508 widgets
→ docs/Wireframes_Rethink_parsed.md (1449 строк)
```

---

## 12. Ключевые технические инсайты

### 🔑 Инсайт 1: абсолютные координаты

```
❌ Неправильно: widget.x = 10 (относительно фрейма)
✅ Правильно:   widget.x = frame.left + 10 (абсолютно на холсте)

Пример: фрейм в (x=1150, y=2900), кнопка внутри:
widget.x1 = 1150 + 10 = 1160
widget.y1 = 2900 + 60 = 2960
```

### 🔑 Инсайт 2: view ID vs model ID

```
wireframe_create_frame() → возвращает MODEL id (для операций с данными)
diagram_list_views()     → возвращает VIEW id (для tailViewId, позиционирования)

tailViewId ВСЕГДА должен быть VIEW id, не model id!
```

### 🔑 Инсайт 3: таймаут при batch-запросах

StarUML зависает при >25-30 параллельных widget calls одновременно.

```
❌ 40 параллельных вызовов → timeout
✅ Батчи по 10-15 → стабильно

Recovery после таймаута:
1. get_status() — убедиться что сервер отвечает
2. Повторить только упавшие запросы
```

### 🔑 Инсайт 4: WFPanel как таблица

Нет нативного "Table" виджета — эмулируем через:
```
WFPanel → заголовок ("Column1 | Column2 | Column3")
WFText  → строки данных ("value1  | value2  | value3")
```

### 🔑 Инсайт 5: WFSlider как прогресс-бар

```python
# Визуализация данных через слайдер:
WFSlider(name="Active (85%)     ██████████░░░", width=450)
WFSlider(name="Grace Period (5%)  ███", width=450)
```

### 🔑 Инсайт 6: ASCII в WFPanel/WFText

StarUML поддерживает Unicode и ASCII art в именах виджетов:
```
"A MATRIX: 5×5\nb VECTOR: [0.82, 0.61, 0.43]\nθ (theta): [0.74, 0.58]"
```

### 🔑 Инсайт 7: SQL todo tracking

```sql
-- Хорошо работает для отслеживания прогресса в сессии:
INSERT INTO todos (id, title, status) VALUES ('frame-5', 'Pricing Tiers', 'pending');
UPDATE todos SET status = 'in_progress' WHERE id = 'frame-5';
UPDATE todos SET status = 'done' WHERE id = 'frame-5';
```

---

## 13. Архитектура финального результата

### Файлы

```
paywall-iap/
├── docs/
│   ├── Wireframes_Rethink.mdj        ← StarUML проект (508 виджетов, 20 экранов)
│   ├── Wireframes_Rethink_parsed.md  ← Auto-generated из .mdj (1449 строк)
│   ├── Wireframes_Rethink.md         ← Spec-документ (ASCII wireframes, reference)
│   └── paywall-iap-wireframes.mdj    ← Первый прототип (118 виджетов, 5 экранов)
├── scripts/
│   ├── merge_migrations.sh           ← Bash скрипт для слияния SQL
│   └── uml_wireframe_parser.py       ← MDJ → MD конвертер
├── backend/
│   └── migrations/
│       └── schema_merged.sql         ← Единая схема БД (797 строк, 22 таблицы)
└── Learning.md                       ← Этот файл
```

### Статистика проекта

| Метрика | Значение |
|---------|---------|
| Экранов | 20 |
| Виджетов | 508 |
| Строк в parsed.md | 1449 |
| Таблиц БД покрыто | 22/22 (100%) |
| Приоритетных итераций | 4 |
| SQL-миграций | 17 |
| Строк схемы | 797 |

### Workflow (воспроизводимый)

```
1. Изучить схему БД
   → cat backend/migrations/schema_merged.sql

2. Написать spec (ASCII wireframes)
   → docs/Wireframes_Rethink.md

3. Открыть StarUML, создать проект
   → staruml: project_new

4. Создать диаграмму и фреймы
   → wireframe_create_diagram
   → 20× wireframe_create_frame (layout grid)

5. Получить view IDs
   → element_list_views (per frame, faster)

6. Добавить виджеты батчами по 10-15
   → wireframe_create_widget (с tailViewId)

7. Сохранить проект
   → save_project → .mdj

8. Конвертировать в MD
   → python3 scripts/uml_wireframe_parser.py input.mdj output.md
```

---

## 🎯 Итог

За несколько итераций мы прошли путь от:
- 📄 Простого `wireframes.md` без привязки к схеме
- ➡️ Переработанного spec с маппингом на реальные таблицы
- ➡️ 4 базовых экранов (94 виджета)
- ➡️ 20 полных экранов (508 виджетов)
- ➡️ Grouped navigation + breadcrumbs + notification badges
- ➡️ Автоматически генерируемой MD-документации

**Главный вывод**: хорошие wireframes начинаются со схемы данных, а не с UI-фантазий. Каждый элемент должен иметь источник в реальной таблице БД.
