Отличная задача! У вас **508 виджетов** на 20 экранах. Чтобы не превратить ответ в бесконечный список, я сгруппировал компоненты по **функциональным паттернам** и привязал их к конкретным экранам из вашей схемы.

Это готовая спецификация для frontend-разработчика на стеке **React + Tailwind + shadcn/ui**.

---

## 🎨 Глобальная карта компонентов (Mapping Table)

| Wireframe Type | Элемент в StarUML | shadcn/ui Компонент | Примечание |
|:---|:---|:---|:---|
| **Container** | `Desktop Frame` | `Layout` (Sidebar + Header) | Использовать `Sidebar` component из examples |
| **Card/Panel** | `Panel` (KPI, Info) | `Card`, `CardHeader`, `CardTitle`, `CardContent` | Для блоков статистики и форм |
| **Data Grid** | `Panel` (Tables) | `Table` + `TanStack Table` | Для списков пользователей, транзакций |
| **Navigation** | `Tab Bar` | `Tabs`, `TabsList`, `TabsTrigger` | Для переключения вкладок (Profile, Settings) |
| **Action** | `Button` | `Button` (variants: default, outline, destructive, ghost) | Основные действия |
| **Input** | `Input` (Text) | `Input` | Текстовые поля |
| **Input** | `Input` (Number) | `Input` (type="number") | Цены, количества |
| **Input** | `Input` (JSONB) | `Textarea` + `Syntax Highlighter` | Для редактирования features/payload |
| **Select** | `Dropdown` | `Select` (простой) / `Combobox` (поиск) | Фильтры, статусы |
| **Toggle** | `Switch` | `Switch` | Boolean флаги (Active, Enable) |
| **Check** | `Checkbox` | `Checkbox` | Массовые действия, права |
| **Range** | `Slider` | `Slider` | Веса бандитов, проценты |
| **Status** | `Text` (Status) | `Badge` (variant: outline/default/destructive) | Active, Failed, Dunning |
| **User** | `Avatar` | `Avatar`, `AvatarImage`, `AvatarFallback` | Профиль админа, пользователи |
| **Divider** | `Separator` | `Separator` | Разделители секций |
| **Link** | `Link` | `Link` (или `Button` variant="link") | Навигация |
| **Text** | `Text` | `Typography` (h1-h4, p, muted, small) | Заголовки, описания |

---

## 📱 Детальная разбивка по экранам (20 Screens)

### 1. Monitoring (Screens 1, 11)
**Экраны:** Admin Dashboard, Analytics Reports
*   **KPI Cards:** `Card` (4 шт в ряд). Внутри: `div` для значения, `p class="text-muted"` для лейбла.
*   **Charts:** `Chart` (shadcn chart component based on `recharts`). LineChart для MRR, BarChart для Churn.
*   **Audit Log:** `Table` с пагинацией (`Pagination`).
*   **Webhook Health:** `Alert` (variant: default/destructive) + `Badge` для статусов провайдеров.
*   **Filters:** `Popover` + `Command` (Combobox pattern) для выбора Date Range и Dimensions.

### 2. User Management (Screens 2, 6, 16)
**Экраны:** User 360°, User List, Admin Audit Log
*   **Identity Card:** `Card` с сеткой (`grid grid-cols-2`).
*   **Profile Tabs:** `Tabs` (Subscription, Billing, Experiments...).
*   **User List:** `DataTable` (TanStack Table) с колонками: Checkbox, Email, Platform, LTV, Status (`Badge`), Actions (`DropdownMenu`).
*   **Actions (Force Cancel, etc.):** `Dialog` (Confirmation modal) перед деструктивными действиями.
*   **Audit Log Diff:** `ScrollArea` + `pre/code` блок для отображения JSONB diff (зеленый/красный фон для строк).
*   **Search:** `Input` с иконкой поиска слева (`Search` icon from `lucide-react`).

### 3. Revenue Ops (Screens 4, 7, 8, 9, 10)
**Экраны:** Revenue Ops Center, Subscriptions, Transactions, Winback, Dunning
*   **Dunning Queue:** `Table` с колонкой "Attempt Count" (можно использовать `Progress` компонент для визуализации 3/5).
*   **Row Actions:** `DropdownMenu` (Retry, Grace, Cancel) вместо кучи кнопок в строке.
*   **Transaction Reconciliation:** `Table` с выравниванием чисел по правому краю (`text-right`). Итоговая строка (`TableFooter`) для Total Amount.
*   **Winback Builder:** `Form` (react-hook-form + zod).
    *   Discount Value: `Input` + `Select` (%, fixed).
    *   Targeting: `Slider` (для дней churn) или `Input` range.
    *   Preview: `Card` с имитацией мобильного экрана.
*   **Dunning Config:** `Accordion` для правил retry (Attempt 1, Attempt 2...), чтобы не занимать место.

### 4. Experiments & Bandit (Screens 3, 12, 15, 19, 20)
**Экраны:** Experiment Studio, A/B List, Context Model, Window Analytics, Multi-Objective
*   **Arm Performance:** `Table` с колонкой Confidence. Использовать `Progress` для визуализации % (0.96 = 96% заполнен).
*   **Winner Badge:** `Badge` (variant: secondary) с иконкой 🏆.
*   **Context Weights:** `Card` + `div` с flex-барами (кастомный визуал для theta vector).
*   **Sliders (Weights):** `Slider` (multi-handle если нужно суммирование) + `Input` рядом для точного значения. Валидация суммы = 1.0 через `Form`.
*   **Model Matrices:** `ScrollArea` + `pre` блок для отображения матриц A, b, θ (моноширинный шрифт).
*   **Event Log:** `ScrollArea` (auto-scroll to bottom) для live-лога событий.

### 5. Configuration & Debug (Screens 5, 13, 14, 17, 18)
**Экраны:** Pricing, Webhook Inspector, Matomo, Settings, Delayed Feedback
*   **Pricing Tiers:** `Card` (3 колонки). Кнопки `Button` внутри карточки.
*   **JSONB Editor:** `Textarea` + кнопка "Validate JSON". Лучше использовать библиотеку `react-json-view` внутри `Card`.
*   **Webhook Payload:** `Dialog` (полноэкранный) для просмотра полного JSON. Кнопки "Copy", "Replay".
*   **Feature Flags:** `Switch` в списке (`List` component).
*   **Currency Rates:** `Table` с inline-редактированием (или `Dialog` для редактирования).
*   **Matomo Queue:** `Table` с цветовой кодировкой статусов (`Badge` variant: outline для pending, destructive для failed).

---

## 🧩 Специфические паттерны (Complex Patterns)

Для вашей схемы есть несколько сложных мест, где стандартных компонентов недостаточно:

### 1. JSONB Editor (pricing_tiers.features, webhook payload)
*   **Компоненты:** `Textarea` + `Button` (Format/Validate).
*   **Рекомендация:** Используйте `react-json-view` или `monaco-editor` (легковесный) внутри `Card` для красивого отображения структуры.
*   **Валидация:** `zod` схема на фронтенде перед отправкой.

### 2. Data Tables с фильтрами (User List, Transactions)
*   **Компоненты:** `Table` + `Pagination` + `DropdownMenu` (для колонок) + `Popover` (для фильтров).
*   **Паттерн:** shadcn/ui `DataTable` example.
*   **Фильтры:** Используйте `Command` (Combobox) для полей типа "Platform", "Status", чтобы можно было искать внутри выпадающего списка.

### 3. Bandit Weights & Sliders
*   **Компоненты:** `Slider` + `Input` (синхронизированные).
*   **Логика:** При изменении одного слайдера, остальные должны пересчитываться (чтобы сумма была 1.0), либо показывать ошибку `FormMessage`.
*   **Визуал:** Добавьте `Tooltip` на ползунок, показывающий точное числовое значение.

### 4. Status Badges (Schema Driven)
Создайте утилиту для маппинга статусов из БД в цвета:
```typescript
// lib/status-config.ts
export const statusConfig = {
  subscription: {
    active: { label: "Active", variant: "default" },
    expired: { label: "Expired", variant: "secondary" },
    grace: { label: "Grace", variant: "warning" }, // кастомный variant
    dunning: { label: "Dunning", variant: "destructive" },
  },
  webhook: {
    pending: { label: "Pending", variant: "outline" },
    processed: { label: "Processed", variant: "default" },
    failed: { label: "Failed", variant: "destructive" },
  }
}
// Использование: <Badge variant={statusConfig.subscription[status].variant}>
```

### 5. Navigation (Sidebar)
*   **Компонент:** `Sidebar` (из новых examples shadcn).
*   **Структура:** Группировка по доменам (Monitoring, Users, Revenue, Experiments, Config).
*   **Бейджи:** `Badge` внутри пункта меню для счетчиков (🔴3 Dunning).
*   **Breadcrumbs:** `Breadcrumb` компонент в хедере каждой страницы.

---

## 🛠 Необходимые дополнительные библиотеки

shadcn/ui — это база, но для вашей схемы понадобятся:

1.  **`@tanstack/react-table`** — для всех таблиц (Users, Transactions, Logs).
2.  **`recharts`** — для графиков (Dashboard, Analytics).
3.  **`date-fns` + `react-day-picker`** — для всех полей даты (Expires at, Created at). shadcn использует `Popover` + `Calendar`.
4.  **`zod` + `react-hook-form`** — для валидации форм (Campaign Builder, Settings).
5.  **`nuqs`** или `zustand` — для управления состоянием фильтров в URL (чтобы при обновлении страницы фильтры таблиц не сбрасывались).
6.  **`react-json-view`** — для отображения JSONB полей.
7.  **`lucide-react`** — иконки (уже входит в shadcn).

---

## 📋 Чеклист для реализации (Priority)

1.  **Setup:** Инициализировать shadcn/ui, настроить `tailwind.config` (colors, radius).
2.  **Layout:** Сверстать `Sidebar` + `Header` + `Breadcrumb` (используется на всех 20 экранах).
3.  **Core Components:** Создать обертки для `StatusBadge`, `JsonViewer`, `DataTable`.
4.  **Screen 1 (Dashboard):** Реализовать KPI Cards и Charts (самый видимый экран).
5.  **Screen 6 (User List):** Реализовать таблицу с фильтрами (самый сложный паттерн данных).
6.  **Screen 2 (User 360):** Реализовать табы и профиль.
7.  **Screen 3 (Experiment):** Реализовать слайдеры и визуализацию бандитов.
8.  **Forms:** Настроить `react-hook-form` для Campaign Builder и Settings.

Эта подборка закрывает 100% ваших 508 виджетов, используя стандартные и расширенные возможности экосистемы shadcn/ui.