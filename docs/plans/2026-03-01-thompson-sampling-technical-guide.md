# Thompson Sampling для Пейволлов: Техническая Реализация

**Дата:** 2026-03-01
**Контекст:** Growth Layer - Multi-Armed Bandit A/B Framework
**Применение:** Оптимизация конверсии и выручки paywall-экранов

---

## Table of Contents

1. [Математическая Основа](#1-математическая-основа)
2. [АрхитектураBandit-Engine](#2-архитектура-bandit-engine)
3. [ТипыНаград](#3-типы-наград)
4. [Мульти-armedBanditдляРазныхЦен](#4-mysqli-armed-bandit-для-разных-цен)
5. [УчётВалют](#5-учёт-валют)
6. [КонтекстуальныйBandit](#6-контекстуальный-bandit)
7. [Production-Considerations](#7-production-considerations)

---

## 1. Математическая Основа

### 1.1 Beta-Распределение

Thompson Sampling использует **Beta-распределение** \( Beta(\alpha, \beta) \) для моделирования вероятности конверсии каждого варианта:

\[
Beta(\alpha, \beta) = \frac{x^{\alpha-1}(1-x)^{\beta-1}}{B(\alpha, \beta)}
\]

Где:
- \( \alpha \) = количество успехов + 1 (конверсии)
- \( \beta \) = количество неудач + 1 (отказы)
- \( B(\alpha, \beta) \) = бета-функция (нормировочная константа)

**Визуализация:**

```
В начале эксперимента (α=1, β=1):
        ▲
       /|\
      / | \
     /  |  \
    /   |   \
   /    |    \
  └─────┴─────┘
   0    0.5   1

После 100 конверсий, 900 отказов (α=101, β=901):
        ▲
       |
       |
      |
     |
    |
   └─────┴─────┘
   0   0.1   1
```

### 1.2 Алгоритм Thompson Sampling

```pseudocode
Псевдокод для каждого пользователя:

1. Для каждого пейволла (arm):
   a. Получить текущую статистику: α_i, β_i
   b. Сгенерировать случайное число из Beta(α_i, β_i)
   c. θ_i = сгенерированное значение

2. Выбрать пейволл с максимальным θ_i:
   winner = argmax(θ_i)

3. Показать winner пользователю

4. После действия пользователя:
   Если конверсия:
     α_winner += 1
   Если отказ:
     β_winner += 1
```

---

## 2. Архитектура Bandit Engine

### 2.1 Компоненты

```go
package service

import (
    "context"
    "fmt"
    "math/rand"
    "time"

    "github.com/bivex/paywall-iap/internal/domain/repository"
    "github.com/redis/go-redis/v9"
    "go.uber.org/zap"
)

// BanditEngine управляет multi-armed bandit экспериментами
type BanditEngine struct {
    repo      repository.ExperimentRepository
    redis      *redis.Client
    logger     *zap.Logger
    rng        *rand.Rand
}

// Arm представляет один вариант пейволла
type Arm struct {
    ID          string                 // "control", "variant_a", etc.
    Name        string                 // "Monthly $9.99"
    Config      PaywallConfig         // Конфигурация пейволла
    Alpha       float64                // α (успехи + 1)
    Beta        float64                // β (неудачи + 1)
    Samples     int64                  // Всего показов
    Conversions int64                  // Конверсии
    Revenue     float64                // Общая выручка
}

// PaywallConfig содержит параметры пейволла
type PaywallConfig struct {
    ProductID    string  `json:"product_id"`
    PriceUSD     float64 `json:"price_usd"`
    TrialDays    int     `json:"trial_days"`
    Currency     string  `json:"currency"`
    Highlight    string  `json:"highlight"`    // "best_value", "most_popular"
    Features     []string `json:"features"`     // ["cancel_anytime", "offline_mode"]
}

// SelectArm выбирает лучший пейволл для пользователя
func (b *BanditEngine) SelectArm(ctx context.Context, experimentID, userID string) (*Arm, error) {
    // 1. Проверяем кеш назначения (sticky assignment)
    cached, _ := b.getAssignment(ctx, experimentID, userID)
    if cached != nil {
        return cached, nil
    }

    // 2. Получаем все arms эксперимента
    arms, err := b.getArms(ctx, experimentID)
    if err != nil {
        return nil, err
    }

    // 3. Thompson Sampling: выбираем arm с максимальным sample
    var winner *Arm
    maxSample := -1.0

    for _, arm := range arms {
        // Генерируем sample из Beta(α, β)
        sample := b.sampleBeta(arm.Alpha, arm.Beta)

        if sample > maxSample {
            maxSample = sample
            winner = arm
        }
    }

    // 4. Сохраняем назначение в кеш (24h sticky)
    b.setAssignment(ctx, experimentID, userID, winner)

    // 5. Увеличиваем счётчик samples
    b.incrementSamples(ctx, experimentID, winner.ID)

    return winner, nil
}

// sampleBeta генерирует случайное число из Beta(α, β)
// Использует метод Marsaglia & Tsang (2000)
func (b *BanditEngine) sampleBeta(alpha, beta float64) float64 {
    // Особые случаи
    if alpha == 1.0 && beta == 1.0 {
        return b.rng.Float64() // Равномерное распределение
    }
    if alpha >= 1.0 && beta == 1.0 {
        return 1.0 - math.Pow(b.rng.Float64(), 1.0/alpha)
    }
    if alpha == 1.0 && beta >= 1.0 {
        return math.Pow(b.rng.Float64(), 1.0/beta)
    }

    // Общий случай: Marsaglia & Tsang
    // Генерируем две гамма-случайные величины
    u := b.rng.Float64()
    v := b.rng.Float64()

    x := math.Pow(u, 1.0/alpha)
    y := math.Pow(v, 1.0/beta)

    return x / (x + y)
}

// RecordReward записывает результат (конверсия или отказ)
func (b *BanditEngine) RecordReward(ctx context.Context, experimentID, armID, userID string, reward Reward) error {
    // Конверсия: увеличиваем α
    if reward.Converted {
        b.redis.Incr(ctx, fmt.Sprintf("bandit:%s:%s:alpha", experimentID, armID))
        b.redis.Incr(ctx, fmt.Sprintf("bandit:%s:%s:conversions", experimentID, armID))
    } else {
        // Отказ: увеличиваем β
        b.redis.Incr(ctx, fmt.Sprintf("bandit:%s:%s:beta", experimentID, armID))
    }

    // Добавляем revenue
    b.redis.IncrByFloat(ctx, fmt.Sprintf("bandit:%s:%s:revenue", experimentID, armID), reward.Value)

    return nil
}

// Reward представляет результат показа пейволла
type Reward struct {
    Converted bool    // true = покупка, false = отказ
    Value      float64 // сумма покупки в USD (или конвертированной валюте)
    Currency   string  // валюта транзакции
    Timestamp  time.Time
}
```

### 2.2 Sticky Assignment

```go
// getAssignment проверяет существующее назначение (24h sticky)
func (b *BanditEngine) getAssignment(ctx context.Context, experimentID, userID string) (*Arm, error) {
    key := fmt.Sprintf("bandit:assignment:%s:%s", experimentID, userID)

    data, err := b.redis.Get(ctx, key).Result()
    if err == redis.Nil {
        return nil, nil // Нет назначения
    }

    // Парсим сохранённый arm
    var assignment struct {
        ArmID    string    `json:"arm_id"`
        ExpiresAt time.Time `json:"expires_at"`
    }
    json.Unmarshal([]byte(data), &assignment)

    if time.Now().After(assignment.ExpiresAt) {
        // Истёк - удаляем
        b.redis.Del(ctx, key)
        return nil, nil
    }

    // Возвращаем arm
    arms, _ := b.getArms(ctx, experimentID)
    for _, arm := range arms {
        if arm.ID == assignment.ArmID {
            return arm, nil
        }
    }

    return nil, nil
}

// setAssignment сохраняет назначение на 24 часа
func (b *BanditEngine) setAssignment(ctx context.Context, experimentID, userID string, arm *Arm) error {
    key := fmt.Sprintf("bandit:assignment:%s:%s", experimentID, userID)

    assignment := map[string]interface{}{
        "arm_id":     arm.ID,
        "assigned_at": time.Now(),
        "expires_at": time.Now().Add(24 * time.Hour),
    }

    data, _ := json.Marshal(assignment)
    return b.redis.Set(ctx, key, data, 24*time.Hour).Err()
}
```

---

## 3. Типы Наград

Thompson Sampling может оптимизировать **разные типы целей**:

### 3.1 Конверсия (Conversion Rate)

```go
// Бинарная награда: 0 или 1
type ConversionReward struct {
    Converted bool
    Value     float64 // 1.0 или 0.0
}

// Обновление:
if reward.Converted {
    alpha += 1.0  // Успех
} else {
    beta += 1.0   // Неудача
}
```

### 3.2 Выручка (Revenue Optimization)

```go
// Непрерывная награда: любая сумма ≥ 0
type RevenueReward struct {
    Converted bool
    Value     float64 // Например: 9.99, 14.99, 0
}

// Обновление для revenue:
// α и β обновляются по конверсии
// Дополнительно отслеживаем среднюю награду

if reward.Converted {
    alpha += 1.0
    totalRevenue += reward.Value
    avgRevenue = totalRevenue / float64(alpha)
} else {
    beta += 1.0
}
```

**Проблема:** Оптимизация revenue без учёта цены может привести к плохому результату.

### 3.3 Нормализованная Выручка (Revenue-MAB)

```go
// Нормализуем награду на цену пейволла
type NormalizedRevenueReward struct {
    Converted bool
    Value     float64 // Фактическая выручка
    PriceUSD  float64 // Цена пейволла
}

// Обновление:
if reward.Converted {
    alpha += 1.0
    // Нормализуем: reward / price
    normalizedReward := reward.Value / reward.PriceUSD
    totalReward += normalizedReward
} else {
    beta += 1.0
}
```

**Почему это важно:**

```
Вариант A: $9.99, конверсия 10%, средний чек = $9.99
Вариант B: $99.99, конверсия 1%, средний чек = $99.99

Конверсия:
  A: 10% → Thompson покажет B чаще
  B: 1%

Revenue:
  A на юзера: 10% × $9.99 = $1.00
  B на юзера: 1% × $99.99 = $1.00

  → Thompson будет балансировать 50/50

Normalized Revenue:
  A: 10% × ($9.99 / $9.99) = 0.10
  B: 1% × ($99.99 / $99.99) = 0.01

  → Thompson покажет A в 10 раз чаще B
```

---

## 4. Multi-Armed Bandit для Разных Цен

### 4.1 Проблема: Разные Цены → Разные LTV

```go
// Три варианта пейволла
type PaywallExperiment struct {
    ID    string
    Arms  []Arm
}

arms := []Arm{
    {
        ID:   "price_a",
        Name: "Monthly $9.99",
        Config: PaywallConfig{PriceUSD: 9.99},
    },
    {
        ID:   "price_b",
        Name: "Monthly $14.99",
        Config: PaywallConfig{PriceUSD: 14.99},
    },
    {
        ID:   "price_c",
        Name: "Annual $79.99",
        Config: PaywallConfig{PriceUSD: 79.99, TrialDays: 14},
    },
}
```

### 4.2 Решение 1: Thompson Sampling по Конверсии

```go
// Оптимизируем ТОЛЬКО конверсию
type ConversionBandit struct{}

func (b *ConversionBandit) SelectArm(arms []Arm) *Arm {
    var winner *Arm
    maxSample := -1.0

    for _, arm := range arms {
        // Sample из Beta(α, β) на основе конверсии
        sample := b.sampleBeta(arm.Alpha, arm.Beta)

        if sample > maxSample {
            maxSample = sample
            winner = arm
        }
    }

    return winner
}

// Результат:
// $9.99   → 10% конверсия → показывается чаще
// $14.99  → 5% конверсия
// $79.99  → 3% конверсия → показывается реже
```

### 4.3 Решение 2: Thompson Sampling по LTV (Expected Value)

```go
// Оптимизируем ожидаемую ценность пользователя
type LTVBandit struct{}

func (b *LTVBandit) SelectArm(arms []Arm) *Arm {
    var winner *Arm
    maxSample := -1.0

    for _, arm := range arms {
        // Средний LTV для этого arm
        avgLTV := arm.TotalRevenue / float64(arm.Conversions)

        // Sample из Beta(α, β) с учётом конверсии
        conversionSample := b.sampleBeta(arm.Alpha, arm.Beta)

        // Ожидаемая ценность = P(конверсия) × LTV
        expectedValue := conversionSample * avgLTV

        if expectedValue > maxSample {
            maxSample = expectedValue
            winner = arm
        }
    }

    return winner
}

// Результат:
// $9.99   → LTV ~$20 → E[V] = 0.10 × $20 = $2.00
// $14.99  → LTV ~$30 → E[V] = 0.05 × $30 = $1.50
// $79.99  → LTV ~$120 → E[V] = 0.03 × $120 = $3.60
//
// → Annual $79.99 будет показываться чаще всего!
```

### 4.4 Решение 3: Gittins Index (оптимально для дисконтированных reward)

```go
// Gittins Index учитывает дисконтирование будущих reward
type GittinsIndexBandit struct {
    Gamma float64 // Дисконт-фактор (0.95-0.99)
}

func (b *GittinsIndexBandit) SelectArm(arms []Arm) *Arm {
    var winner *Arm
    maxIndex := -1.0

    for _, arm := range arms {
        // Gittins Index approximation
        // GI ≈ μ + sqrt(2 * ln(n) / n) для UCB, но с дисконтированием

        n := float64(arm.Samples)
        if n == 0 {
            n = 1 // Prior
        }

        // Оценка среднего значения
        mu := arm.TotalRevenue / float64(arm.Conversions)
        if arm.Conversions == 0 {
            mu = arm.Config.PriceUSD // Априорная оценка
        }

        // Байесовский credible interval
        // Упрощённый Gittins Index:
        explorationBonus := math.Sqrt(2 * math.Log(n) / n)
        gittinsIndex := mu + explorationBonus

        if gittinsIndex > maxIndex {
            maxIndex = gittinsIndex
            winner = arm
        }
    }

    return winner
}
```

---

## 5. Учёт Валют

### 5.1 Проблема Мультивалютности

```
Пользователь из США:
  Видит $9.99 → конвертирует → покупает за $9.99

Пользователь из России:
  Видит 499₽ → конвертирует → покупает за 499₽
```

### 5.2 Решение: Конвертация в Base Currency

```go
// CurrencyConverter конвертирует в базовую валюту (USD)
type CurrencyConverter struct {
    rates map[string]float64 // USD → XXX: 1.0, EUR → 0.92, RUB → 0.011
    mutex sync.RWMutex
}

func (c *CurrencyConverter) ToUSD(amount float64, currency string) float64 {
    c.mutex.RLock()
    defer c.mutex.RUnlock()

    rate, ok := c.rates[currency]
    if !ok {
        rate = 1.0 // Fallback
    }

    return amount * rate
}

// Использование при записи reward:
func (b *BanditEngine) RecordReward(ctx context.Context, reward Reward) error {
    // Конвертируем в USD
    valueUSD := b.converter.ToUSD(reward.Value, reward.Currency)

    // Записываем в bandit state (USD)
    b.redis.IncrByFloat(ctx,
        fmt.Sprintf("bandit:%s:%s:revenue_usd", experimentID, armID),
        valueUSD,
    )

    return nil
}
```

### 5.3 Умные Курсы Валют

```go
// ExchangeRateService автоматически обновляет курсы
type ExchangeRateService struct {
    httpClient *http.Client
    apiProvider string // " ECB", "fixer.io", etc.
    redis       *redis.Client
}

// UpdateRates обновляет курсы каждый час
func (s *ExchangeRateService) UpdateRates(ctx context.Context) error {
    // 1. Получаем курсы от ECB (бесплатно, надёжно)
    resp, err := s.httpClient.Get("https://api.exchangerate.host/latest")
    if err != nil {
        return err
    }

    var rates struct {
        Rates map[string]float64 `json:"rates"`
        Base  string              `json:"base"`
    }
    json.NewDecoder(resp.Body).Decode(&rates)

    // 2. Конвертируем в USD-базисный формат
    usdRates := make(map[string]float64)
    usdRates["USD"] = 1.0

    for currency, rate := range rates.Rates {
        if rates.Base == "USD" {
            usdRates[currency] = rate
        } else if currency == "USD" {
            // У нас rate EUR/USD, нам нужно USD/EUR
            usdRates["EUR"] = 1.0 / rate
        }
    }

    // 3. Сохраняем в Redis
    data, _ := json.Marshal(usdRates)
    s.redis.Set(ctx, "currency:rates", data, time.Hour)

    return nil
}
```

---

## 6. Контекстуальный Bandit

### 6.1 Идея: Разные пейволлы для разных сегментов

```go
// ContextualBandit учитывает контекст пользователя
type ContextualBandit struct {
    baseBandit *BanditEngine
    segments   []SegmentRule
}

type SegmentRule struct {
    Name      string
    Predicate func(user User) bool
    Weight    float64 // Модификатор вероятности
}

// Примеры сегментов
segments := []SegmentRule{
    {
        Name: "high_ltv_user",
        Predicate: func(u User) bool {
            return u.TotalSpent > 100
        },
        Weight: 0.2, // +20% к вероятности показать дорогой пейволл
    },
    {
        Name: "russia_region",
        Predicate: func(u User) bool {
            return u.Country == "RU"
        },
        Weight: 0.5, // Показывать бюджетный вариант чаще
    },
    {
        Name: "trial_user",
        Predicate: func(u User) bool {
            return u.DaysSinceInstall <= 7
        },
        Weight: 0.0, // Не показывать дорогие варианты
    },
}
```

### 6.2 LinUCB (Linear Upper Confidence Bound)

Для контекстуального бандита используем **LinUCB**:

```go
type LinUCBBandit struct {
    dimension int    // Размерность контекста
    arms      map[string]*ArmContext
}

type ArmContext struct {
    ArmID     string
    A         *mat64.Dense  // Матрица d×d
    b         []float64     // Вектор размерности d
    theta     []float64     // Параметры модели
}

func (l *LinUCBBandit) SelectArm(ctx []float64) string {
    var bestArm string
    bestUCB := -math.MaxFloat64

    for armID, arm := range l.arms {
        // UCB = θ^T × x + α × sqrt(x^T × A^(-1) × x)

        // 1. θ^T × x (предсказание)
        prediction := dotProduct(arm.theta, ctx)

        // 2. Дисперсия
        invA := matrixInverse(arm.A)
        variance := dotProduct(ctx, matrixVectorMult(invA, ctx))

        // 3. UCB
        ucb := prediction + 0.3 * math.Sqrt(variance)

        if ucb > bestUCB {
            bestUCB = ucb
            bestArm = armID
        }
    }

    return bestArm
}

// Обновление модели
func (l *LinUCBBandit) Update(armID string, ctx []float64, reward float64) {
    arm := l.arms[armID]

    // A += x × x^T
    outerProduct := outerProduct(ctx, ctx)
    arm.A.Add(outerProduct)

    // b += reward × x
    for i := range ctx {
        arm.b[i] += reward * ctx[i]
    }

    // θ = A^(-1) × b
    invA := matrixInverse(arm.A)
    arm.theta = matrixVectorMult(invA, arm.b)
}
```

### 6.3 Практический Пример Контекстов

```go
// Контекст пользователя для paywall
type PaywallContext struct {
    Country        string
    Device         string  // "ios" | "android"
    AppVersion     string
    DaysSinceInstall int
    TotalSpent     float64
    PreviousPurchases []string
}

// Преобразуем контекст в вектор признаков
func ContextToFeatures(ctx PaywallContext) []float64 {
    features := make([]float64, 20) // 20-мерный вектор

    // One-hot encoding страны
    switch ctx.Country {
    case "US":
        features[0] = 1.0
    case "RU":
        features[1] = 1.0
    case "DE":
        features[2] = 1.0
    // ... другие страны
    }

    // Устройство
    if ctx.Device == "ios" {
        features[10] = 1.0
    } else {
        features[11] = 1.0
    }

    // Нормализованные численные признаки
    features[18] = math.Min(float64(ctx.DaysSinceInstall) / 365.0, 1.0)
    features[19] = math.Min(ctx.TotalSpent / 1000.0, 1.0)

    return features
}
```

---

## 7. Production Considerations

### 7.1 Cold Start Problem

В начале эксперимента (α=β=1) все варианты равновероятны. Решения:

```go
// Решение 1: Warm-up период
type WarmupBandit struct {
    baseBandit BanditEngine
    minSamples int
}

func (w *WarmupBandit) SelectArm(ctx context.Context, arms []Arm) *Arm {
    // Проверяем: если у всех arm < minSamples
    for _, arm := range arms {
        if arm.Samples < w.minSamples {
            // Используем равномерное распределение
            return arms[w.rng.Intn(len(arms))]
        }
    }

    // После warm-up используем Thompson Sampling
    return w.baseBandit.SelectArm(ctx, arms)
}

// Решение 2: Априорные знания (Historical Data)
type InformedBandit struct {
    baseBandit BanditEngine
    priors      map[string]PriorInfo
}

type PriorInfo struct {
    Alpha float64 // Начальное значение (по умолчанию 1.0)
    Beta  float64 // Начальное значение (по умолчанию 1.0)
}

func (i *InformedBandit) LoadHistoricalData() {
    // Загружаем историческую конверсию для похожих продуктов
    // Например: Monthly $9.99 обычно даёт 8-12% конверсии

    i.priors["price_9.99"] = PriorInfo{
        Alpha: 12.0, // 11 конверсий (8-12 среднее ~10) + 1
        Beta:  90.0,  // 89 отказов
    }
}

func (i *InformedBandit) SelectArm(ctx context.Context, armID string) *Arm {
    // Используем prior info для инициализации
    if prior, ok := i.priors[armID]; ok {
        return i.sampleBeta(prior.Alpha, prior.Beta)
    }

    return i.baseBandit.SelectArm(ctx, armID)
}
```

### 7.2 Non-Stationarity (Изменение во времени)

Поведение пользователей меняется (сезонность,疲劳, конкуренты). Решение:

```go
// Sliding Window: учитываем только последние N показов
type SlidingWindowBandit struct {
    windowSize int64 // Например, последние 1000 показов
}

func (s *SlidingWindowBandit) GetStats(experimentID, armID string) (alpha, beta float64, err error) {
    // Получаем последние windowSize записей из Redis
    // Используем Redis Sorted Set с timestamp как score

    key := fmt.Sprintf("bandit:window:%s:%s", experimentID, armID)

    // Удаляем старые записи (out of window)
    cutoff := time.Now().Add(-24 * time.Hour)
    s.redis.ZRemRangeByScore(ctx, key, "0", fmt.Sprint(cutoff.Unix()))

    // Получаем количество записей
    nStr, _ := s.redis.ZCard(ctx, key).Result()
    n, _ := strconv.ParseInt(nStr, 10, 64)

    // Считаем successes и failures
    successes, _ := s.redis.ZCount(ctx, key, "-inf", "1").Result()
    failures := n - successes

    alpha = 1.0 + float64(successes)
    beta = 1.0 + float64(failures)

    return alpha, beta, nil
}

// Запись события с sliding window
func (s *SlidingWindowBandit) RecordEvent(ctx context.Context, experimentID, armID string, success bool) error {
    key := fmt.Sprintf("bandit:window:%s:%s", experimentID, armID)
    score := float64(time.Now().UnixNano()) / 1e9 // timestamp как score
    value := "0"
    if success {
        value = "1"
    }

    return s.redis.ZAdd(ctx, key, redis.Z{Score: score, Member: value}).Err()
}
```

### 7.3 Delayed Feedback (延迟反馈)

Проблема: пользователь может конвертировать через дни/недели после показа пейволла.

```go
// DelayedRewardBandit отслеживает pending rewards
type DelayedRewardBandit struct {
    redis *redis.Client
}

func (d *DelayedRewardBandit) RecordPendingReward(ctx context.Context, experimentID, armID, userID string) error {
    // Записываем pending reward
    key := fmt.Sprintf("bandit:pending:%s:%s:%s", experimentID, armID, userID)

    pending := map[string]interface{}{
        "arm_id":     armID,
        "user_id":    userID,
        "created_at": time.Now(),
        "expires_at": time.Now().Add(30 * 24 * time.Hour), // 30 дней
    }

    data, _ := json.Marshal(pending)
    return d.redis.Set(ctx, key, data, 30*24*time.Hour).Err()
}

// Worker проверяет pending rewards каждые 5 минут
func (d *DelayedRewardBandit) ProcessPendingRewards(ctx context.Context) error {
    // Получаем всех пользователей, кто сделал покупку
    // Ищем их pending rewards
    pattern := "bandit:pending:*"

    iter := d.redis.Scan(ctx, 0, pattern, 0).Iterator()
    for iter.Next(ctx) {
        key := iter.Val()

        var pending map[string]interface{}
        json.Unmarshal([]byte(key), &pending)

        // Проверяем: пользователь совершил покупку?
        converted, _ := d.checkConversion(ctx, pending["user_id"].(string))

        if converted {
            // Обновляем bandit статистику
            d.RecordReward(ctx, pending["experiment_id"].(string), pending["arm_id"].(string), Reward{
                Converted: true,
                Value:     converted.Amount,
                Currency:  converted.Currency,
            })

            // Удаляем pending reward
            d.redis.Del(ctx, key)
        }
    }

    return nil
}
```

---

## 8. Пример Конфигурации Эксперимента

```json
{
  "id": "paywall_price_test_2026_03_01",
  "name": "Paywall Price Optimization",
  "algorithm": "thompson_sampling",
  "objective": "ltv_optimization",
  "min_sample_size": 1000,
  "confidence_threshold": 0.95,
  "auto_rebalance": true,
  "warmup_samples": 100,

  "arms": [
    {
      "arm_id": "monthly_9_99",
      "name": "Monthly $9.99",
      "weight": 0.33,
      "config": {
        "product_id": "com.app.premium.monthly",
        "price_usd": 9.99,
        "trial_days": 7,
        "features": ["cancel_anytime", "offline_mode"],
        "highlight": "most_popular"
      },
      "prior": {
        "alpha": 10.0,
        "beta": 90.0
      }
    },
    {
      "arm_id": "monthly_14_99",
      "name": "Monthly $14.99",
      "weight": 0.33,
      "config": {
        "product_id": "com.app.premium.monthly_plus",
        "price_usd": 14.99,
        "trial_days": 14,
        "features": ["cancel_anytime", "offline_mode", "priority_support"],
        "highlight": "best_value"
      },
      "prior": {
        "alpha": 7.0,
        "beta": 93.0
      }
    },
    {
      "arm_id": "annual_79_99",
      "name": "Annual $79.99 (Save 33%)",
      "weight": 0.34,
      "config": {
        "product_id": "com.app.premium.annual",
        "price_usd": 79.99,
        "trial_days": 14,
        "features": ["cancel_anytime", "offline_mode", "priority_support"],
        "highlight": "best_value"
      },
      "prior": {
        "alpha": 4.0,
        "beta": 96.0
      }
    }
  ],

  "segments": [
    {
      "name": "vip_users",
      "predicate": "total_spent > 100",
      "arm_weights": {
        "monthly_9_99": 0.1,
        "monthly_14_99": 0.3,
        "annual_79_99": 0.6
      }
    },
    {
      "name": "new_users",
      "predicate": "days_since_install <= 7",
      "arm_weights": {
        "monthly_9_99": 0.6,
        "monthly_14_99": 0.2,
        "annual_79_99": 0.2
      }
    }
  ]
}
```

---

## 9. Мониторинг и Отладка

### 9.1 Метрики Bandit

```go
type BanditMetrics struct {
    // Потеря (Regret) - насколько хуже мы сделали по сравнению с лучшим arm
    Regret float64

    // Импульсивность (Exploration rate) - как часто переключаем arms
    ExplorationRate float64

    // Конвергенция - разница между лучшим и худшим arm
    ConvergenceGAP float64

    // Balance - насколько равномерно распределены показы
    BalanceIndex float64
}

func (b *BanditEngine) CalculateMetrics(ctx context.Context, experimentID string) (*BanditMetrics, error) {
    arms := b.getArms(ctx, experimentID)

    // Находим лучший arm (по среднему revenue)
    var bestArm *Arm
    bestAvg := -1.0
    totalSamples := int64(0)

    for _, arm := range arms {
        avg := arm.TotalRevenue / float64(arm.Conversions)
        if avg > bestAvg {
            bestAvg = avg
            bestArm = arm
        }
        totalSamples += arm.Samples
    }

    // Рассчитываем regret
    regret := 0.0
    for _, arm := range arms {
        expectedBest := bestAvg * float64(arm.Samples)
        actual := arm.TotalRevenue
        regret += expectedBest - actual
    }

    // Рассчитываем exploration rate
    uniqueArmsShown := len(arms)
    balanceIndex := float64(uniqueArmsShown) / math.Log(float64(totalSamples))

    return &BanditMetrics{
        Regret:          regret,
        ExplorationRate: balanceIndex,
        ConvergenceGAP:  bestAvg / b.getWorstAvg(arms),
        BalanceIndex:    balanceIndex,
    }, nil
}
```

### 9.2 Grafana Dashboard

```json
{
  "dashboard": {
    "title": "A/B Test Bandit Performance",
    "panels": [
      {
        "title": "Arm Distribution (Traffic %)",
        "targets": [
          {
            "expr": "sum(increase(bandit_samples_total{experiment_id=\"$exp\"}[5m])) by (arm)"
          }
        ]
      },
      {
        "title": "Conversion Rate by Arm",
        "targets": [
          {
            "expr": "sum(increase(bandit_conversions_total{experiment_id=\"$exp\"}[5m])) by (arm) / sum(increase(bandit_samples_total{experiment_id=\"$exp\"}[5m])) by (arm)"
          }
        ]
      },
      {
        "title": "Average Revenue per User (by Arm)",
        "targets": [
          {
            "expr": "sum(increase(bandit_revenue_usd_total{experiment_id=\"$exp\"}[5m])) by (arm) / sum(increase(bandit_conversions_total{experiment_id=\"$exp\"}[5m])) by (arm)"
          }
        ]
      },
      {
        "title": "Thompson Sampling α/β",
        "targets": [
          {
            "expr": "bandit_arm_alpha{experiment_id=\"$exp\", arm=\"control\"}"
          },
          {
            "expr": "bandit_arm_beta{experiment_id=\"$exp\", arm=\"control\"}"
          }
        ]
      },
      {
        "title": "Cumulative Regret",
        "targets": [
          {
            "expr": "bandit_cumulative_regret{experiment_id=\"$exp\"}"
          }
        ]
      }
    ]
  }
}
```

---

## 10. Production Checklist

- [ ] **Cold Start**: Использовать исторические данные или warm-up период
- [ ] **Delayed Feedback**: Worker обрабатывает pending rewards
- [ ] **Currency Conversion**: Автоматическое обновление курсов каждый час
- [ ] **Sticky Assignment**: 24-часовое закрепление за пользователем
- [ ] **Sliding Window**: Учитывать только последние 1000 показов для нестационарности
- [ ] **Monitoring**: Grafana dashboard с метриками regret, convergence
- [ ] **Fallback**: Если Redis недоступен → использовать PostgreSQL
- [ ] **Idempotency**: Запись reward идемпотентна (ключ: experiment_id + user_id)

---

**END OF DOCUMENT**

Для вопросов: architecture@bivex.com
