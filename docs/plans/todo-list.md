# Paywall & D2C Todo List

> Обновлено: 2026-03-17 (Приоритет 1 выполнен)
> На основе анализа кода + инсайты из HyperHub/Two and a Half Gamers стримов

## Приоритет 1 — Критично для запуска D2C

- [x] **PaywallTriggerService** — когда показывать paywall
  - [x] Счётчик сессий пользователя (session_count)
  - [x] Флаг "просмотрел рекламу" (has_viewed_ads)
  - [x] Флаг "покупал через IAP" (purchased_via_iap)
  - [ ] Триггер: показывать paywall после проигрыша, не на старте
  - [x] Backend: `GET /v1/user/trigger-status`
  - [ ] Mobile: интеграция с session manager

- [x] **D2CButtonModule** — кнопка "купи на сайте"
  - [x] Проверка purchase_channel перед показом
  - [x] Не показывать если purchase_channel == "iap" (правило Google)
  - [ ] Скрывать кнопку после D2C покупки
  - [ ] Логирование показов для аналитики

- [x] **Email capture на paywall**
  - [x] Добавить email поле в paywall schema
  - [ ] Email как первый шаг перед показом цен
  - [x] Валидация email на клиенте
  - [x] Сохранение email в User entity до покупки

- [x] **PurchaseChannelRepository** — откуда покупал
  - [x] Поле purchase_channel в User (iap, stripe, null)
  - [x] Обновление при первой IAP покупке
  - [ ] Обновление при Stripe webhook
  - [x] Migration: ALTER TABLE users ADD COLUMN purchase_channel TEXT

## Приоритет 2 — Монетизация по данным стримов

- [ ] **AdSegmentationService** — логика показа рекламы
  - [ ] Новые юзеры: 0 реклам первые 2-3 сессии
  - [ ] Купил что-либо: реклама выключена полностью
  - [ ] Rewarded видео: всегда доступно, но агрессивнее не платившим
  - [ ] Интерстишл: между сессиями, с cooldown
  - [ ] Remote config для настройки порогов

- [ ] **AdARPDAUTracker** — метрика рекламного дохода
  - [ ] Трекинг impressions per user per day
  - [ ] Средний eCPM по юзеру
  - [ ] Fill rate (цель ≥98%)
  - [ ] Backend: daily aggregation
  - [ ] Админка: отчёт ARPDAU vs IAP ARPDAU

- [ ] **OfferManager** — контекстные офферы
  - [ ] Оффер в момент проигрыша (strongest trigger)
  - [ ] Progressive offer: цена падает по мере действий
  - [ ] Delayed reward: оффер на N дней для D1 retention
  - [ ] A/B тестирование разных офферов

## Приоритет 3 — UI/UX улучшения

- [ ] **Paywall вариации по поведению**
  - [ ] Для новых юзеров: акцент на "попробуй бесплатно"
  - [ ] Для платящих: "апгрейд节省"
  - [ ] Для уходящих: winback оффер со скидкой
  - [ ] Динамическая загрузка конфига с бэкенда

- [ ] **Social proof на paywall**
  - [ ] "N человек купили сегодня"
  - [ ] "Популярный выбор" бейдж на среднем тарифе
  - [ ] Реальные данные из статистики

## Приоритет 4 — Аналитика и отчёты

- [ ] **Cohort анализ по каналу покупки**
  - [ ] IAP vs D2C retention
  - [ ] IAP vs D2C LTV
  - [ ] IAP vs D2C refund rate

- [ ] **Steering эффективность**
  - [ ] % юзеров перешедших на D2C
  - [ ] Потерянная комиссия Apple/Google
  - [ ] Email capture rate

## Приоритет 5 — Мобильная часть

- [ ] **React Native Paywall Component**
  - [ ] Переиспользуемый компонент paywall
  - [ ] Динамическая загрузка схемы с бэкенда
  - [ ] Анимации переходов
  - [ ] Интеграция с IAP store

- [ ] **Session tracking**
  - [ ] Трекинг сессий на мобильном
  - [ ] Отправка на бэкенд для триггеров
  - [ ] Первый запуск vs возвратившийся

## Технический долг

- [ ] **Stripe webhook обработчик**
  - [ ] invoice.payment_succeeded → обновление subscription
  - [ ] customer.subscription.deleted → отмена
  - [ ] Не забыл: в коде есть только приём, нет обработки

- [ ] **Email рассылки**
  - [ ] Welcome email после регистрации
  - [ ] Через неделю: "как тебе без рекламы?"
  - [ ] За день до списания: напоминание
  - [ ] Winback для неактивных

## Справка — источники решений

**HyperHub стримы:**
- Контекстный оффер в момент проигрыша = самый сильный триггер
- После первой покупки реклама выключается (Candy Crush pattern)
- Новые юзеры сначала без рекламы

**Two and a Half Gamers:**
- Felix: главный показатель = ARPDAU, не только eCPM
- Fill rate ≥98% или теряешь деньги
- 30-40% UA бюджета уходит в rewarded UA

**Chip Thirsten (D2C эксперт):**
- После D2C покупки убери все steering кнопки из интерфейса
- Иначе Google засчитает как попытку увести комиссию

**Finch кейс:**
- Email capture = главная ценность D2C
- 10M DAU через подписку с петомцем
- Re-engagement через email = ключевая метрика
