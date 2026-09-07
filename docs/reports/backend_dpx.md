# 🐹 DPX-Go: Software Design Pattern & Architecture Report

- **Target Path:** `/Volumes/External/Code/paywall-iap/backend`
- **Files Scanned:** `149`
- **Total Patterns & Findings:** `675`
- **Analysis Elapsed Time:** `0.157s`

## 📊 Breakdown by Category

| Category | Count |
|---|:---:|
| **CREATIONAL** | 106 |
| **STRUCTURAL** | 5 |
| **BEHAVIORAL** | 15 |
| **IDIOM** | 412 |
| **PRINCIPLE** | 137 |

## 📋 Detailed Pattern Findings

### #1 FACTORY_METHOD on `NewRateLimiter`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/middleware/rate_limiter.go:31:1`
- **Summary:** Factory constructor function 'NewRateLimiter()' encapsulates instantiation of '*RateLimiter'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewRateLimiter()' encapsulates instantiation of '*RateLimiter' -> `/Volumes/External/Code/paywall-iap/backend/internal/application/middleware/rate_limiter.go:31:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/application/middleware/rate_limiter.go:31:1`

### #2 FACTORY_METHOD on `NewJWTMiddleware`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/middleware/jwt.go:39:1`
- **Summary:** Factory constructor function 'NewJWTMiddleware()' encapsulates instantiation of '*JWTMiddleware'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewJWTMiddleware()' encapsulates instantiation of '*JWTMiddleware' -> `/Volumes/External/Code/paywall-iap/backend/internal/application/middleware/jwt.go:39:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 3 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/application/middleware/jwt.go:39:1`

### #3 FACTORY_METHOD on `NewRegisterCommand`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/register.go:23:1`
- **Summary:** Factory constructor function 'NewRegisterCommand()' encapsulates instantiation of '*RegisterCommand'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewRegisterCommand()' encapsulates instantiation of '*RegisterCommand' -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/register.go:23:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/register.go:23:1`

### #4 FACTORY_METHOD on `NewResolveGracePeriodCommand`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/resolve_grace_period.go:29:1`
- **Summary:** Factory constructor function 'NewResolveGracePeriodCommand()' encapsulates instantiation of '*ResolveGracePeriodCommand'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewResolveGracePeriodCommand()' encapsulates instantiation of '*ResolveGracePeriodCommand' -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/resolve_grace_period.go:29:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/resolve_grace_period.go:29:1`

### #5 FACTORY_METHOD on `NewTrackSessionCommand`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/track_session.go:17:1`
- **Summary:** Factory constructor function 'NewTrackSessionCommand()' encapsulates instantiation of '*TrackSessionCommand'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewTrackSessionCommand()' encapsulates instantiation of '*TrackSessionCommand' -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/track_session.go:17:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/track_session.go:17:1`

### #6 FACTORY_METHOD on `NewCaptureEmailCommand`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/capture_email.go:18:1`
- **Summary:** Factory constructor function 'NewCaptureEmailCommand()' encapsulates instantiation of '*CaptureEmailCommand'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewCaptureEmailCommand()' encapsulates instantiation of '*CaptureEmailCommand' -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/capture_email.go:18:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/capture_email.go:18:1`

### #7 FACTORY_METHOD on `NewAdminLoginCommand`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/admin_login.go:23:1`
- **Summary:** Factory constructor function 'NewAdminLoginCommand()' encapsulates instantiation of '*AdminLoginCommand'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewAdminLoginCommand()' encapsulates instantiation of '*AdminLoginCommand' -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/admin_login.go:23:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 3 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/admin_login.go:23:1`

### #8 FACTORY_METHOD on `NewCancelSubscriptionCommand`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/cancel_subscription.go:18:1`
- **Summary:** Factory constructor function 'NewCancelSubscriptionCommand()' encapsulates instantiation of '*CancelSubscriptionCommand'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewCancelSubscriptionCommand()' encapsulates instantiation of '*CancelSubscriptionCommand' -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/cancel_subscription.go:18:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/cancel_subscription.go:18:1`

### #9 FACTORY_METHOD on `NewAcceptWinbackOfferCommand`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/accept_winback_offer.go:33:1`
- **Summary:** Factory constructor function 'NewAcceptWinbackOfferCommand()' encapsulates instantiation of '*AcceptWinbackOfferCommand'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewAcceptWinbackOfferCommand()' encapsulates instantiation of '*AcceptWinbackOfferCommand' -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/accept_winback_offer.go:33:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/accept_winback_offer.go:33:1`

### #10 FACTORY_METHOD on `NewVerifyIAPCommand`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/verify_iap.go:58:1`
- **Summary:** Factory constructor function 'NewVerifyIAPCommand()' encapsulates instantiation of '*VerifyIAPCommand'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewVerifyIAPCommand()' encapsulates instantiation of '*VerifyIAPCommand' -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/verify_iap.go:58:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 5 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/verify_iap.go:58:1`

### #11 FACTORY_METHOD on `NewVerifyIAPCommandLegacy`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/verify_iap.go:75:1`
- **Summary:** Factory constructor function 'NewVerifyIAPCommandLegacy()' encapsulates instantiation of '*VerifyIAPCommand'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewVerifyIAPCommandLegacy()' encapsulates instantiation of '*VerifyIAPCommand' -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/verify_iap.go:75:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 5 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/verify_iap.go:75:1`

### #12 FACTORY_METHOD on `NewCreateGracePeriodCommand`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/create_grace_period.go:35:1`
- **Summary:** Factory constructor function 'NewCreateGracePeriodCommand()' encapsulates instantiation of '*CreateGracePeriodCommand'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewCreateGracePeriodCommand()' encapsulates instantiation of '*CreateGracePeriodCommand' -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/create_grace_period.go:35:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/create_grace_period.go:35:1`

### #13 FACTORY_METHOD on `NewGetSubscriptionQuery`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/query/subscription.go:20:1`
- **Summary:** Factory constructor function 'NewGetSubscriptionQuery()' encapsulates instantiation of '*GetSubscriptionQuery'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewGetSubscriptionQuery()' encapsulates instantiation of '*GetSubscriptionQuery' -> `/Volumes/External/Code/paywall-iap/backend/internal/application/query/subscription.go:20:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/application/query/subscription.go:20:1`

### #14 FACTORY_METHOD on `NewCheckAccessQuery`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/query/subscription.go:62:1`
- **Summary:** Factory constructor function 'NewCheckAccessQuery()' encapsulates instantiation of '*CheckAccessQuery'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewCheckAccessQuery()' encapsulates instantiation of '*CheckAccessQuery' -> `/Volumes/External/Code/paywall-iap/backend/internal/application/query/subscription.go:62:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/application/query/subscription.go:62:1`

### #15 FACTORY_METHOD on `NewGetTriggerStatusQuery`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/query/get_trigger_status.go:18:1`
- **Summary:** Factory constructor function 'NewGetTriggerStatusQuery()' encapsulates instantiation of '*GetTriggerStatusQuery'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewGetTriggerStatusQuery()' encapsulates instantiation of '*GetTriggerStatusQuery' -> `/Volumes/External/Code/paywall-iap/backend/internal/application/query/get_trigger_status.go:18:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/application/query/get_trigger_status.go:18:1`

### #16 FACTORY_METHOD on `NewCohortWorker`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/cohort_jobs.go:48:1`
- **Summary:** Factory constructor function 'NewCohortWorker()' encapsulates instantiation of '*CohortWorker'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewCohortWorker()' encapsulates instantiation of '*CohortWorker' -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/cohort_jobs.go:48:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/cohort_jobs.go:48:1`

### #17 FACTORY_METHOD on `NewCohortAggregationTask`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/cohort_jobs.go:67:1`
- **Summary:** Factory constructor function 'NewCohortAggregationTask()' encapsulates instantiation of '(*asynq.Task, error)'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewCohortAggregationTask()' encapsulates instantiation of '(*asynq.Task, error)' -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/cohort_jobs.go:67:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 3 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/cohort_jobs.go:67:1`

### #18 FACTORY_METHOD on `NewBanditMaintenanceJobs`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/bandit_maintenance_jobs.go:24:1`
- **Summary:** Factory constructor function 'NewBanditMaintenanceJobs()' encapsulates instantiation of '*BanditMaintenanceJobs'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewBanditMaintenanceJobs()' encapsulates instantiation of '*BanditMaintenanceJobs' -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/bandit_maintenance_jobs.go:24:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/bandit_maintenance_jobs.go:24:1`

### #19 FACTORY_METHOD on `NewTaskHandlers`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/tasks.go:44:1`
- **Summary:** Factory constructor function 'NewTaskHandlers()' encapsulates instantiation of '*TaskHandlers'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewTaskHandlers()' encapsulates instantiation of '*TaskHandlers' -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/tasks.go:44:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/tasks.go:44:1`

### #20 FACTORY_METHOD on `NewDunningJobHandler`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/dunning_jobs.go:34:1`
- **Summary:** Factory constructor function 'NewDunningJobHandler()' encapsulates instantiation of '*DunningJobHandler'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewDunningJobHandler()' encapsulates instantiation of '*DunningJobHandler' -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/dunning_jobs.go:34:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/dunning_jobs.go:34:1`

### #21 FACTORY_METHOD on `NewABTestJobHandler`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/ab_test_jobs.go:28:1`
- **Summary:** Factory constructor function 'NewABTestJobHandler()' encapsulates instantiation of '*ABTestJobHandler'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewABTestJobHandler()' encapsulates instantiation of '*ABTestJobHandler' -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/ab_test_jobs.go:28:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/ab_test_jobs.go:28:1`

### #22 FACTORY_METHOD on `NewWinbackJobHandler`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/winback_jobs.go:41:1`
- **Summary:** Factory constructor function 'NewWinbackJobHandler()' encapsulates instantiation of '*WinbackJobHandler'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewWinbackJobHandler()' encapsulates instantiation of '*WinbackJobHandler' -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/winback_jobs.go:41:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/winback_jobs.go:41:1`

### #23 FACTORY_METHOD on `NewGracePeriodJobHandler`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/grace_period_jobs.go:35:1`
- **Summary:** Factory constructor function 'NewGracePeriodJobHandler()' encapsulates instantiation of '*GracePeriodJobHandler'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewGracePeriodJobHandler()' encapsulates instantiation of '*GracePeriodJobHandler' -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/grace_period_jobs.go:35:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/grace_period_jobs.go:35:1`

### #24 FACTORY_METHOD on `NewAnalyticsJobHandler`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/analytics_jobs.go:23:1`
- **Summary:** Factory constructor function 'NewAnalyticsJobHandler()' encapsulates instantiation of '*AnalyticsJobHandler'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewAnalyticsJobHandler()' encapsulates instantiation of '*AnalyticsJobHandler' -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/analytics_jobs.go:23:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/analytics_jobs.go:23:1`

### #25 FACTORY_METHOD on `NewCurrencyJobs`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/currency_jobs.go:18:1`
- **Summary:** Factory constructor function 'NewCurrencyJobs()' encapsulates instantiation of '*CurrencyJobs'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewCurrencyJobs()' encapsulates instantiation of '*CurrencyJobs' -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/currency_jobs.go:18:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/currency_jobs.go:18:1`

### #26 FACTORY_METHOD on `NewMatomoWorker`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/matomo_jobs.go:30:1`
- **Summary:** Factory constructor function 'NewMatomoWorker()' encapsulates instantiation of '*MatomoWorker'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewMatomoWorker()' encapsulates instantiation of '*MatomoWorker' -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/matomo_jobs.go:30:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/matomo_jobs.go:30:1`

### #27 FACTORY_METHOD on `NewProcessBatchTask`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/matomo_jobs.go:43:1`
- **Summary:** Factory constructor function 'NewProcessBatchTask()' encapsulates instantiation of '(*asynq.Task, error)'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewProcessBatchTask()' encapsulates instantiation of '(*asynq.Task, error)' -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/matomo_jobs.go:43:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/matomo_jobs.go:43:1`

### #28 FACTORY_METHOD on `NewSendEventTask`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/matomo_jobs.go:120:1`
- **Summary:** Factory constructor function 'NewSendEventTask()' encapsulates instantiation of '(*asynq.Task, error)'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewSendEventTask()' encapsulates instantiation of '(*asynq.Task, error)' -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/matomo_jobs.go:120:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/matomo_jobs.go:120:1`

### #29 FACTORY_METHOD on `NewSendEcommerceTask`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/matomo_jobs.go:176:1`
- **Summary:** Factory constructor function 'NewSendEcommerceTask()' encapsulates instantiation of '(*asynq.Task, error)'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewSendEcommerceTask()' encapsulates instantiation of '(*asynq.Task, error)' -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/matomo_jobs.go:176:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/matomo_jobs.go:176:1`

### #30 FACTORY_METHOD on `NewRedisBanditCache`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:27:1`
- **Summary:** Factory constructor function 'NewRedisBanditCache()' encapsulates instantiation of '*RedisBanditCache'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewRedisBanditCache()' encapsulates instantiation of '*RedisBanditCache' -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:27:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:27:1`

### #31 FACTORY_METHOD on `NewAnalyticsCache`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:21:1`
- **Summary:** Factory constructor function 'NewAnalyticsCache()' encapsulates instantiation of '*AnalyticsCache'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewAnalyticsCache()' encapsulates instantiation of '*AnalyticsCache' -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:21:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:21:1`

### #32 FACTORY_METHOD on `NewPostgresMatomoEventRepository`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/matomo_fallback.go:24:1`
- **Summary:** Factory constructor function 'NewPostgresMatomoEventRepository()' encapsulates instantiation of '*PostgresMatomoEventRepository'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewPostgresMatomoEventRepository()' encapsulates instantiation of '*PostgresMatomoEventRepository' -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/matomo_fallback.go:24:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/matomo_fallback.go:24:1`

### #33 FACTORY_METHOD on `NewPostgresBanditRepository`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:33:1`
- **Summary:** Factory constructor function 'NewPostgresBanditRepository()' encapsulates instantiation of '*PostgresBanditRepository'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewPostgresBanditRepository()' encapsulates instantiation of '*PostgresBanditRepository' -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:33:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:33:1`

### #34 FACTORY_METHOD on `NewExperimentAdminRepository`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:23:1`
- **Summary:** Factory constructor function 'NewExperimentAdminRepository()' encapsulates instantiation of '*ExperimentAdminRepository'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewExperimentAdminRepository()' encapsulates instantiation of '*ExperimentAdminRepository' -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:23:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:23:1`

### #35 FACTORY_METHOD on `NewAutomationJobRunRepository`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/automation_job_run_repository.go:18:1`
- **Summary:** Factory constructor function 'NewAutomationJobRunRepository()' encapsulates instantiation of '*AutomationJobRunRepository'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewAutomationJobRunRepository()' encapsulates instantiation of '*AutomationJobRunRepository' -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/automation_job_run_repository.go:18:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/automation_job_run_repository.go:18:1`

### #36 FACTORY_METHOD on `NewDunningRepository`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/dunning_repository_impl.go:21:1`
- **Summary:** Factory constructor function 'NewDunningRepository()' encapsulates instantiation of '*DunningRepositoryImpl'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewDunningRepository()' encapsulates instantiation of '*DunningRepositoryImpl' -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/dunning_repository_impl.go:21:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/dunning_repository_impl.go:21:1`

### #37 FACTORY_METHOD on `New`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/sqlc/generated/db.go:20:1`
- **Summary:** Factory constructor function 'New()' encapsulates instantiation of '*Queries'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'New()' encapsulates instantiation of '*Queries' -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/sqlc/generated/db.go:20:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/sqlc/generated/db.go:20:1`

### #38 FACTORY_METHOD on `NewPool`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/pool/pgxpool.go:12:1`
- **Summary:** Factory constructor function 'NewPool()' encapsulates instantiation of '(*pgxpool.Pool, error)'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewPool()' encapsulates instantiation of '(*pgxpool.Pool, error)' -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/pool/pgxpool.go:12:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/pool/pgxpool.go:12:1`

### #39 FACTORY_METHOD on `NewCredentialResolver`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/credential_resolver.go:28:1`
- **Summary:** Factory constructor function 'NewCredentialResolver()' encapsulates instantiation of '*CredentialResolver'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewCredentialResolver()' encapsulates instantiation of '*CredentialResolver' -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/credential_resolver.go:28:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/credential_resolver.go:28:1`

### #40 FACTORY_METHOD on `NewGoogleVerifier`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/google_verifier.go:24:1`
- **Summary:** Factory constructor function 'NewGoogleVerifier()' encapsulates instantiation of '*GoogleVerifier'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewGoogleVerifier()' encapsulates instantiation of '*GoogleVerifier' -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/google_verifier.go:24:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 3 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/google_verifier.go:24:1`

### #41 FACTORY_METHOD on `NewIAPAdapter`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/adapter.go:18:1`
- **Summary:** Factory constructor function 'NewIAPAdapter()' encapsulates instantiation of '*IAPAdapter'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewIAPAdapter()' encapsulates instantiation of '*IAPAdapter' -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/adapter.go:18:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/adapter.go:18:1`

### #42 FACTORY_METHOD on `NewAppleVerifierAdapter`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/adapter.go:82:1`
- **Summary:** Factory constructor function 'NewAppleVerifierAdapter()' encapsulates instantiation of '*AppleVerifierAdapter'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewAppleVerifierAdapter()' encapsulates instantiation of '*AppleVerifierAdapter' -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/adapter.go:82:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/adapter.go:82:1`

### #43 FACTORY_METHOD on `NewAndroidVerifierAdapter`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/adapter.go:95:1`
- **Summary:** Factory constructor function 'NewAndroidVerifierAdapter()' encapsulates instantiation of '*AndroidVerifierAdapter'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewAndroidVerifierAdapter()' encapsulates instantiation of '*AndroidVerifierAdapter' -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/adapter.go:95:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/adapter.go:95:1`

### #44 FACTORY_METHOD on `NewDynamicAppleVerifier`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/dynamic_verifiers.go:18:1`
- **Summary:** Factory constructor function 'NewDynamicAppleVerifier()' encapsulates instantiation of '*DynamicAppleVerifier'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewDynamicAppleVerifier()' encapsulates instantiation of '*DynamicAppleVerifier' -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/dynamic_verifiers.go:18:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/dynamic_verifiers.go:18:1`

### #45 FACTORY_METHOD on `NewDynamicGoogleVerifier`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/dynamic_verifiers.go:50:1`
- **Summary:** Factory constructor function 'NewDynamicGoogleVerifier()' encapsulates instantiation of '*DynamicGoogleVerifier'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewDynamicGoogleVerifier()' encapsulates instantiation of '*DynamicGoogleVerifier' -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/dynamic_verifiers.go:50:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/dynamic_verifiers.go:50:1`

### #46 FACTORY_METHOD on `NewAppleVerifier`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/verifier.go:76:1`
- **Summary:** Factory constructor function 'NewAppleVerifier()' encapsulates instantiation of '*AppleVerifier'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewAppleVerifier()' encapsulates instantiation of '*AppleVerifier' -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/verifier.go:76:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 3 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/verifier.go:76:1`

### #47 FACTORY_METHOD on `NewClient`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:42:1`
- **Summary:** Factory constructor function 'NewClient()' encapsulates instantiation of '*Client'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewClient()' encapsulates instantiation of '*Client' -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:42:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:42:1`

### #48 FACTORY_METHOD on `NewMoney`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/valueobject/money.go:20:1`
- **Summary:** Factory constructor function 'NewMoney()' encapsulates instantiation of '(*Money, error)'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewMoney()' encapsulates instantiation of '(*Money, error)' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/valueobject/money.go:20:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/valueobject/money.go:20:1`

### #49 FACTORY_METHOD on `NewEmail`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/valueobject/email.go:19:1`
- **Summary:** Factory constructor function 'NewEmail()' encapsulates instantiation of '(*Email, error)'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewEmail()' encapsulates instantiation of '(*Email, error)' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/valueobject/email.go:19:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/valueobject/email.go:19:1`

### #50 FACTORY_METHOD on `NewSubscriptionStatus`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/valueobject/subscription_status.go:21:1`
- **Summary:** Factory constructor function 'NewSubscriptionStatus()' encapsulates instantiation of '(SubscriptionStatus, error)'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewSubscriptionStatus()' encapsulates instantiation of '(SubscriptionStatus, error)' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/valueobject/subscription_status.go:21:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/valueobject/subscription_status.go:21:1`

### #51 FACTORY_METHOD on `NewPlanType`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/valueobject/plan_type.go:20:1`
- **Summary:** Factory constructor function 'NewPlanType()' encapsulates instantiation of '(PlanType, error)'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewPlanType()' encapsulates instantiation of '(PlanType, error)' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/valueobject/plan_type.go:20:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/valueobject/plan_type.go:20:1`

### #52 FACTORY_METHOD on `NewSubscription`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/subscription.go:50:1`
- **Summary:** Factory constructor function 'NewSubscription()' encapsulates instantiation of '*Subscription'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewSubscription()' encapsulates instantiation of '*Subscription' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/subscription.go:50:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 6 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/subscription.go:50:1`

### #53 FACTORY_METHOD on `NewUser`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/user.go:47:1`
- **Summary:** Factory constructor function 'NewUser()' encapsulates instantiation of '*User'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewUser()' encapsulates instantiation of '*User' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/user.go:47:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 6 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/user.go:47:1`

### #54 FACTORY_METHOD on `NewDunning`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/dunning.go:36:1`
- **Summary:** Factory constructor function 'NewDunning()' encapsulates instantiation of '*Dunning'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewDunning()' encapsulates instantiation of '*Dunning' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/dunning.go:36:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 3 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/dunning.go:36:1`

### #55 FACTORY_METHOD on `NewTransaction`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/transaction.go:31:1`
- **Summary:** Factory constructor function 'NewTransaction()' encapsulates instantiation of '*Transaction'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewTransaction()' encapsulates instantiation of '*Transaction' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/transaction.go:31:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 5 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/transaction.go:31:1`

### #56 FACTORY_METHOD on `NewGracePeriod`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/grace_period.go:32:1`
- **Summary:** Factory constructor function 'NewGracePeriod()' encapsulates instantiation of '*GracePeriod'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewGracePeriod()' encapsulates instantiation of '*GracePeriod' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/grace_period.go:32:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 3 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/grace_period.go:32:1`

### #57 FACTORY_METHOD on `NewWinbackOffer`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/winback_offer.go:43:1`
- **Summary:** Factory constructor function 'NewWinbackOffer()' encapsulates instantiation of '*WinbackOffer'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewWinbackOffer()' encapsulates instantiation of '*WinbackOffer' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/winback_offer.go:43:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 5 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/winback_offer.go:43:1`

### #58 FACTORY_METHOD on `NewValidationError`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/errors/validation_errors.go:45:1`
- **Summary:** Factory constructor function 'NewValidationError()' encapsulates instantiation of '*ValidationError'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewValidationError()' encapsulates instantiation of '*ValidationError' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/errors/validation_errors.go:45:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/errors/validation_errors.go:45:1`

### #59 FACTORY_METHOD on `NewGracePeriodService`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/grace_period_service.go:29:1`
- **Summary:** Factory constructor function 'NewGracePeriodService()' encapsulates instantiation of '*GracePeriodService'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewGracePeriodService()' encapsulates instantiation of '*GracePeriodService' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/grace_period_service.go:29:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 3 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/grace_period_service.go:29:1`

### #60 FACTORY_METHOD on `NewAdvancedBanditEngine`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:72:1`
- **Summary:** Factory constructor function 'NewAdvancedBanditEngine()' encapsulates instantiation of '*AdvancedBanditEngine'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewAdvancedBanditEngine()' encapsulates instantiation of '*AdvancedBanditEngine' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:72:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 7 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:72:1`

### #61 FACTORY_METHOD on `NewExperimentAdminService`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:165:1`
- **Summary:** Factory constructor function 'NewExperimentAdminService()' encapsulates instantiation of '*ExperimentAdminService'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewExperimentAdminService()' encapsulates instantiation of '*ExperimentAdminService' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:165:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:165:1`

### #62 FACTORY_METHOD on `NewFeatureFlagService`
- **Category:** `creational`
- **Confidence:** **65%** [MEDIUM]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/feature_flag_service.go:31:1`
- **Summary:** Factory constructor function 'NewFeatureFlagService()' encapsulates instantiation of '*FeatureFlagService'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewFeatureFlagService()' encapsulates instantiation of '*FeatureFlagService' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/feature_flag_service.go:31:1`

### #63 FACTORY_METHOD on `NewLTVSubscriptionAdapter`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_subscription_adapter.go:18:1`
- **Summary:** Factory constructor function 'NewLTVSubscriptionAdapter()' encapsulates instantiation of 'SubscriptionRepository'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewLTVSubscriptionAdapter()' encapsulates instantiation of 'SubscriptionRepository' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_subscription_adapter.go:18:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_subscription_adapter.go:18:1`

### #64 FACTORY_METHOD on `NewAnalyticsReportService`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_report_service.go:18:1`
- **Summary:** Factory constructor function 'NewAnalyticsReportService()' encapsulates instantiation of '*AnalyticsReportService'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewAnalyticsReportService()' encapsulates instantiation of '*AnalyticsReportService' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_report_service.go:18:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_report_service.go:18:1`

### #65 FACTORY_METHOD on `NewRevenueOpsService`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/revenue_ops_service.go:17:1`
- **Summary:** Factory constructor function 'NewRevenueOpsService()' encapsulates instantiation of '*RevenueOpsService'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewRevenueOpsService()' encapsulates instantiation of '*RevenueOpsService' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/revenue_ops_service.go:17:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/revenue_ops_service.go:17:1`

### #66 FACTORY_METHOD on `NewUserProfileService`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/user_profile_service.go:18:1`
- **Summary:** Factory constructor function 'NewUserProfileService()' encapsulates instantiation of '*UserProfileService'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewUserProfileService()' encapsulates instantiation of '*UserProfileService' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/user_profile_service.go:18:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/user_profile_service.go:18:1`

### #67 FACTORY_METHOD on `NewMatomoForwarder`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/matomo_forwarder.go:52:1`
- **Summary:** Factory constructor function 'NewMatomoForwarder()' encapsulates instantiation of '*MatomoForwarder'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewMatomoForwarder()' encapsulates instantiation of '*MatomoForwarder' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/matomo_forwarder.go:52:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 3 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/matomo_forwarder.go:52:1`

### #68 FACTORY_METHOD on `NewPaywallTriggerService`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/paywall_trigger_service.go:26:1`
- **Summary:** Factory constructor function 'NewPaywallTriggerService()' encapsulates instantiation of '*PaywallTriggerService'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewPaywallTriggerService()' encapsulates instantiation of '*PaywallTriggerService' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/paywall_trigger_service.go:26:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/paywall_trigger_service.go:26:1`

### #69 FACTORY_METHOD on `NewThompsonSamplingBandit`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:229:1`
- **Summary:** Factory constructor function 'NewThompsonSamplingBandit()' encapsulates instantiation of '*ThompsonSamplingBandit'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewThompsonSamplingBandit()' encapsulates instantiation of '*ThompsonSamplingBandit' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:229:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 3 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:229:1`

### #70 FACTORY_METHOD on `NewHybridObjectiveStrategy`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:54:1`
- **Summary:** Factory constructor function 'NewHybridObjectiveStrategy()' encapsulates instantiation of '*HybridObjectiveStrategy'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewHybridObjectiveStrategy()' encapsulates instantiation of '*HybridObjectiveStrategy' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:54:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 5 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:54:1`

### #71 FACTORY_METHOD on `NewAutomationJobExecutionService`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/automation_job_execution_service.go:51:1`
- **Summary:** Factory constructor function 'NewAutomationJobExecutionService()' encapsulates instantiation of '*AutomationJobExecutionService'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewAutomationJobExecutionService()' encapsulates instantiation of '*AutomationJobExecutionService' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/automation_job_execution_service.go:51:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/automation_job_execution_service.go:51:1`

### #72 FACTORY_METHOD on `NewDelayedRewardStrategy`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:65:1`
- **Summary:** Factory constructor function 'NewDelayedRewardStrategy()' encapsulates instantiation of '*DelayedRewardStrategy'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewDelayedRewardStrategy()' encapsulates instantiation of '*DelayedRewardStrategy' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:65:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 3 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:65:1`

### #73 FACTORY_METHOD on `NewWinbackService`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/winback_service.go:29:1`
- **Summary:** Factory constructor function 'NewWinbackService()' encapsulates instantiation of '*WinbackService'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewWinbackService()' encapsulates instantiation of '*WinbackService' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/winback_service.go:29:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 3 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/winback_service.go:29:1`

### #74 FACTORY_METHOD on `NewCurrencyConversionRewardStrategy`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/currency_conversion_strategy.go:28:1`
- **Summary:** Factory constructor function 'NewCurrencyConversionRewardStrategy()' encapsulates instantiation of '*CurrencyConversionRewardStrategy'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewCurrencyConversionRewardStrategy()' encapsulates instantiation of '*CurrencyConversionRewardStrategy' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/currency_conversion_strategy.go:28:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 3 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/currency_conversion_strategy.go:28:1`

### #75 FACTORY_METHOD on `NewLTVService`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:55:1`
- **Summary:** Factory constructor function 'NewLTVService()' encapsulates instantiation of '*LTVService'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewLTVService()' encapsulates instantiation of '*LTVService' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:55:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 5 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:55:1`

### #76 FACTORY_METHOD on `NewNotificationService`
- **Category:** `creational`
- **Confidence:** **65%** [MEDIUM]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/notification_service.go:26:1`
- **Summary:** Factory constructor function 'NewNotificationService()' encapsulates instantiation of '*NotificationService'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewNotificationService()' encapsulates instantiation of '*NotificationService' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/notification_service.go:26:1`

### #77 FACTORY_METHOD on `NewLinUCBSelectionStrategy`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/linucb_strategy.go:33:1`
- **Summary:** Factory constructor function 'NewLinUCBSelectionStrategy()' encapsulates instantiation of '*LinUCBSelectionStrategy'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewLinUCBSelectionStrategy()' encapsulates instantiation of '*LinUCBSelectionStrategy' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/linucb_strategy.go:33:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 5 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/linucb_strategy.go:33:1`

### #78 FACTORY_METHOD on `NewAuditService`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/audit_service.go:16:1`
- **Summary:** Factory constructor function 'NewAuditService()' encapsulates instantiation of '*AuditService'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewAuditService()' encapsulates instantiation of '*AuditService' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/audit_service.go:16:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/audit_service.go:16:1`

### #79 FACTORY_METHOD on `NewExperimentWinnerRecommendationService`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:107:1`
- **Summary:** Factory constructor function 'NewExperimentWinnerRecommendationService()' encapsulates instantiation of '*ExperimentWinnerRecommendationService'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewExperimentWinnerRecommendationService()' encapsulates instantiation of '*ExperimentWinnerRecommendationService' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:107:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:107:1`

### #80 FACTORY_METHOD on `NewExperimentWinnerRecommendationServiceWithCalculator`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:116:1`
- **Summary:** Factory constructor function 'NewExperimentWinnerRecommendationServiceWithCalculator()' encapsulates instantiation of '*ExperimentWinnerRecommendationService'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewExperimentWinnerRecommendationServiceWithCalculator()' encapsulates instantiation of '*ExperimentWinnerRecommendationService' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:116:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:116:1`

### #81 FACTORY_METHOD on `NewExperimentWinnerRecommendationServiceWithCalculatorAndAppender`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:120:1`
- **Summary:** Factory constructor function 'NewExperimentWinnerRecommendationServiceWithCalculatorAndAppender()' encapsulates instantiation of '*ExperimentWinnerRecommendationService'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewExperimentWinnerRecommendationServiceWithCalculatorAndAppender()' encapsulates instantiation of '*ExperimentWinnerRecommendationService' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:120:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:120:1`

### #82 FACTORY_METHOD on `NewAnalyticsService`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_service.go:17:1`
- **Summary:** Factory constructor function 'NewAnalyticsService()' encapsulates instantiation of '*AnalyticsService'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewAnalyticsService()' encapsulates instantiation of '*AnalyticsService' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_service.go:17:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_service.go:17:1`

### #83 FACTORY_METHOD on `NewDunningService`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/dunning_service.go:24:1`
- **Summary:** Factory constructor function 'NewDunningService()' encapsulates instantiation of '*DunningService'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewDunningService()' encapsulates instantiation of '*DunningService' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/dunning_service.go:24:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 4 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/dunning_service.go:24:1`

### #84 FACTORY_METHOD on `NewCurrencyRateService`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/currency_service.go:61:1`
- **Summary:** Factory constructor function 'NewCurrencyRateService()' encapsulates instantiation of '*CurrencyRateService'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewCurrencyRateService()' encapsulates instantiation of '*CurrencyRateService' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/currency_service.go:61:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/currency_service.go:61:1`

### #85 FACTORY_METHOD on `NewExperimentRepairReconciler`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_reconciler.go:34:1`
- **Summary:** Factory constructor function 'NewExperimentRepairReconciler()' encapsulates instantiation of '*ExperimentRepairReconciler'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewExperimentRepairReconciler()' encapsulates instantiation of '*ExperimentRepairReconciler' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_reconciler.go:34:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_reconciler.go:34:1`

### #86 FACTORY_METHOD on `NewExperimentAutomationReconciler`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_automation_service.go:43:1`
- **Summary:** Factory constructor function 'NewExperimentAutomationReconciler()' encapsulates instantiation of '*ExperimentAutomationReconciler'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewExperimentAutomationReconciler()' encapsulates instantiation of '*ExperimentAutomationReconciler' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_automation_service.go:43:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_automation_service.go:43:1`

### #87 FACTORY_METHOD on `NewABAnalyticsService`
- **Category:** `creational`
- **Confidence:** **65%** [MEDIUM]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ab_analytics_service.go:29:1`
- **Summary:** Factory constructor function 'NewABAnalyticsService()' encapsulates instantiation of '*ABAnalyticsService'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewABAnalyticsService()' encapsulates instantiation of '*ABAnalyticsService' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ab_analytics_service.go:29:1`

### #88 FACTORY_METHOD on `NewSlidingWindowStrategy`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:38:1`
- **Summary:** Factory constructor function 'NewSlidingWindowStrategy()' encapsulates instantiation of '*SlidingWindowStrategy'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewSlidingWindowStrategy()' encapsulates instantiation of '*SlidingWindowStrategy' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:38:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 5 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:38:1`

### #89 FACTORY_METHOD on `NewExperimentRepairService`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:69:1`
- **Summary:** Factory constructor function 'NewExperimentRepairService()' encapsulates instantiation of '*ExperimentRepairService'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewExperimentRepairService()' encapsulates instantiation of '*ExperimentRepairService' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:69:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:69:1`

### #90 FACTORY_METHOD on `NewABTestMiddleware`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/middleware/ab_test_middleware.go:15:1`
- **Summary:** Factory constructor function 'NewABTestMiddleware()' encapsulates instantiation of '*ABTestMiddleware'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewABTestMiddleware()' encapsulates instantiation of '*ABTestMiddleware' -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/middleware/ab_test_middleware.go:15:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/middleware/ab_test_middleware.go:15:1`

### #91 FACTORY_METHOD on `NewWebhookHandler`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/webhook.go:34:1`
- **Summary:** Factory constructor function 'NewWebhookHandler()' encapsulates instantiation of '*WebhookHandler'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewWebhookHandler()' encapsulates instantiation of '*WebhookHandler' -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/webhook.go:34:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 5 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/webhook.go:34:1`

### #92 FACTORY_METHOD on `NewPaywallHandler`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/paywall.go:21:1`
- **Summary:** Factory constructor function 'NewPaywallHandler()' encapsulates instantiation of '*PaywallHandler'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewPaywallHandler()' encapsulates instantiation of '*PaywallHandler' -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/paywall.go:21:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 4 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/paywall.go:21:1`

### #93 FACTORY_METHOD on `NewSubscriptionHandler`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/subscription.go:23:1`
- **Summary:** Factory constructor function 'NewSubscriptionHandler()' encapsulates instantiation of '*SubscriptionHandler'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewSubscriptionHandler()' encapsulates instantiation of '*SubscriptionHandler' -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/subscription.go:23:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 4 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/subscription.go:23:1`

### #94 FACTORY_METHOD on `NewABTestHandler`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/ab.go:17:1`
- **Summary:** Factory constructor function 'NewABTestHandler()' encapsulates instantiation of '*ABTestHandler'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewABTestHandler()' encapsulates instantiation of '*ABTestHandler' -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/ab.go:17:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/ab.go:17:1`

### #95 FACTORY_METHOD on `NewWinbackHandler`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/winback.go:21:1`
- **Summary:** Factory constructor function 'NewWinbackHandler()' encapsulates instantiation of '*WinbackHandler'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewWinbackHandler()' encapsulates instantiation of '*WinbackHandler' -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/winback.go:21:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 3 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/winback.go:21:1`

### #96 FACTORY_METHOD on `NewIAPHandler`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/iap.go:26:1`
- **Summary:** Factory constructor function 'NewIAPHandler()' encapsulates instantiation of '*IAPHandler'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewIAPHandler()' encapsulates instantiation of '*IAPHandler' -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/iap.go:26:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 3 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/iap.go:26:1`

### #97 FACTORY_METHOD on `NewAuthHandler`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/auth.go:27:1`
- **Summary:** Factory constructor function 'NewAuthHandler()' encapsulates instantiation of '*AuthHandler'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewAuthHandler()' encapsulates instantiation of '*AuthHandler' -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/auth.go:27:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 3 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/auth.go:27:1`

### #98 FACTORY_METHOD on `NewAppSettingsHandler`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_app_settings.go:23:1`
- **Summary:** Factory constructor function 'NewAppSettingsHandler()' encapsulates instantiation of '*AppSettingsHandler'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewAppSettingsHandler()' encapsulates instantiation of '*AppSettingsHandler' -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_app_settings.go:23:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 2 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_app_settings.go:23:1`

### #99 FACTORY_METHOD on `NewAppsHandler`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_apps.go:20:1`
- **Summary:** Factory constructor function 'NewAppsHandler()' encapsulates instantiation of '*AppsHandler'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewAppsHandler()' encapsulates instantiation of '*AppsHandler' -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_apps.go:20:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_apps.go:20:1`

### #100 FACTORY_METHOD on `NewAnalyticsHandlersExtended`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/analytics_extended.go:23:1`
- **Summary:** Factory constructor function 'NewAnalyticsHandlersExtended()' encapsulates instantiation of '*AnalyticsHandlersExtended'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewAnalyticsHandlersExtended()' encapsulates instantiation of '*AnalyticsHandlersExtended' -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/analytics_extended.go:23:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 3 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/analytics_extended.go:23:1`

### #101 FACTORY_METHOD on `NewAdminHandler`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin.go:50:1`
- **Summary:** Factory constructor function 'NewAdminHandler()' encapsulates instantiation of '*AdminHandler'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewAdminHandler()' encapsulates instantiation of '*AdminHandler' -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin.go:50:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 12 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin.go:50:1`

### #102 FACTORY_METHOD on `NewBanditHandler`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/bandit.go:34:1`
- **Summary:** Factory constructor function 'NewBanditHandler()' encapsulates instantiation of '*BanditHandler'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewBanditHandler()' encapsulates instantiation of '*BanditHandler' -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/bandit.go:34:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/bandit.go:34:1`

### #103 FACTORY_METHOD on `NewAdminPaywallsHandler`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_paywalls.go:40:1`
- **Summary:** Factory constructor function 'NewAdminPaywallsHandler()' encapsulates instantiation of '*AdminPaywallsHandler'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewAdminPaywallsHandler()' encapsulates instantiation of '*AdminPaywallsHandler' -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_paywalls.go:40:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 1 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_paywalls.go:40:1`

### #104 FACTORY_METHOD on `NewBanditAdvancedHandler`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/bandit_advanced.go:37:1`
- **Summary:** Factory constructor function 'NewBanditAdvancedHandler()' encapsulates instantiation of '*BanditAdvancedHandler'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewBanditAdvancedHandler()' encapsulates instantiation of '*BanditAdvancedHandler' -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/bandit_advanced.go:37:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 3 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/bandit_advanced.go:37:1`

### #105 FACTORY_METHOD on `NewAnalyticsHandler`
- **Category:** `creational`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/analytics.go:25:1`
- **Summary:** Factory constructor function 'NewAnalyticsHandler()' encapsulates instantiation of '*AnalyticsHandler'

#### Evidence Trail:
- `+65%` **[FACTORY_METHOD_CONSTRUCTOR]** Factory constructor function 'NewAnalyticsHandler()' encapsulates instantiation of '*AnalyticsHandler' -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/analytics.go:25:1`
- `+30%` **[FACTORY_METHOD_PARAMETERIZED]** Encapsulates parameterized construction across 4 input parameter(s) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/analytics.go:25:1`

### #106 ABSTRACT_FACTORY on `Querier`
- **Category:** `creational`
- **Confidence:** **55%** [MEDIUM]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/sqlc/generated/querier.go:13:1`
- **Summary:** Declares family of 3 product creation method(s) (CreateSubscription, CreateTransaction, CreateUser)

#### Evidence Trail:
- `+55%` **[ABSTRACT_FACTORY_METHODS]** Declares family of 3 product creation method(s) (CreateSubscription, CreateTransaction, CreateUser) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/sqlc/generated/querier.go:13:1`

### #107 DECORATOR on `numericStringPatcher`
- **Category:** `structural`
- **Confidence:** **50%** [MEDIUM]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/verifier.go:18:1`
- **Summary:** Wraps inner component 'wrapped: http.RoundTripper' to decorate behavior

#### Evidence Trail:
- `+50%` **[DECORATOR_WRAPS_COMPONENT]** Wraps inner component 'wrapped: http.RoundTripper' to decorate behavior -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/verifier.go:18:1`

### #108 FACADE on `Client`
- **Category:** `structural`
- **Confidence:** **67%** [MEDIUM]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:35:1`
- **Summary:** Struct 'Client' follows Facade / High-Level Client naming convention

#### Evidence Trail:
- `+40%` **[FACADE_NAMING]** Struct 'Client' follows Facade / High-Level Client naming convention -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:35:1`
- `+45%` **[FACADE_AGGREGATES_SUBSYSTEMS]** Aggregates 3 subsystem services (config, httpClient, logger) behind unified API -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:35:1`

### #109 FACADE on `AdvancedBanditEngine`
- **Category:** `structural`
- **Confidence:** **67%** [MEDIUM]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:16:1`
- **Summary:** Struct 'AdvancedBanditEngine' follows Facade / High-Level Client naming convention

#### Evidence Trail:
- `+40%` **[FACADE_NAMING]** Struct 'AdvancedBanditEngine' follows Facade / High-Level Client naming convention -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:16:1`
- `+45%` **[FACADE_AGGREGATES_SUBSYSTEMS]** Aggregates 11 subsystem services (base, rewardStrategy, selectionStrategy) behind unified API -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:16:1`

### #110 FACADE on `LTVService`
- **Category:** `structural`
- **Confidence:** **67%** [MEDIUM]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:29:1`
- **Summary:** Struct 'LTVService' follows Facade / High-Level Client naming convention

#### Evidence Trail:
- `+40%` **[FACADE_NAMING]** Struct 'LTVService' follows Facade / High-Level Client naming convention -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:29:1`
- `+45%` **[FACADE_AGGREGATES_SUBSYSTEMS]** Aggregates 4 subsystem services (matomoClient, cohortWorker, subscriptionRepo) behind unified API -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:29:1`

### #111 FACADE on `CurrencyRateService`
- **Category:** `structural`
- **Confidence:** **67%** [MEDIUM]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/currency_service.go:22:1`
- **Summary:** Struct 'CurrencyRateService' follows Facade / High-Level Client naming convention

#### Evidence Trail:
- `+40%` **[FACADE_NAMING]** Struct 'CurrencyRateService' follows Facade / High-Level Client naming convention -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/currency_service.go:22:1`
- `+45%` **[FACADE_AGGREGATES_SUBSYSTEMS]** Aggregates 3 subsystem services (redisClient, logger, httpClient) behind unified API -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/currency_service.go:22:1`

### #112 STRATEGY on `RewardStrategy`
- **Category:** `behavioral`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:101:1`
- **Summary:** Interface 'RewardStrategy' defines polymorphic Strategy algorithm interface

#### Evidence Trail:
- `+75%` **[STRATEGY_INTERFACE_NAMING]** Interface 'RewardStrategy' defines polymorphic Strategy algorithm interface -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:101:1`

### #113 STRATEGY on `SelectionStrategy`
- **Category:** `behavioral`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:107:1`
- **Summary:** Interface 'SelectionStrategy' defines polymorphic Strategy algorithm interface

#### Evidence Trail:
- `+75%` **[STRATEGY_INTERFACE_NAMING]** Interface 'SelectionStrategy' defines polymorphic Strategy algorithm interface -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:107:1`

### #114 STRATEGY on `WindowStrategy`
- **Category:** `behavioral`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:113:1`
- **Summary:** Interface 'WindowStrategy' defines polymorphic Strategy algorithm interface

#### Evidence Trail:
- `+75%` **[STRATEGY_INTERFACE_NAMING]** Interface 'WindowStrategy' defines polymorphic Strategy algorithm interface -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:113:1`

### #115 COMMAND on `RegisterCommand`
- **Category:** `behavioral`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/register.go:17:1`
- **Summary:** Struct 'RegisterCommand' encapsulates executable command operation with 'Execute()'

#### Evidence Trail:
- `+70%` **[COMMAND_STRUCT]** Struct 'RegisterCommand' encapsulates executable command operation with 'Execute()' -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/register.go:17:1`

### #116 COMMAND on `ResolveGracePeriodCommand`
- **Category:** `behavioral`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/resolve_grace_period.go:12:1`
- **Summary:** Struct 'ResolveGracePeriodCommand' encapsulates executable command operation with 'Execute()'

#### Evidence Trail:
- `+70%` **[COMMAND_STRUCT]** Struct 'ResolveGracePeriodCommand' encapsulates executable command operation with 'Execute()' -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/resolve_grace_period.go:12:1`

### #117 COMMAND on `TrackSessionCommand`
- **Category:** `behavioral`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/track_session.go:13:1`
- **Summary:** Struct 'TrackSessionCommand' encapsulates executable command operation with 'Execute()'

#### Evidence Trail:
- `+70%` **[COMMAND_STRUCT]** Struct 'TrackSessionCommand' encapsulates executable command operation with 'Execute()' -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/track_session.go:13:1`

### #118 COMMAND on `CaptureEmailCommand`
- **Category:** `behavioral`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/capture_email.go:14:1`
- **Summary:** Struct 'CaptureEmailCommand' encapsulates executable command operation with 'Execute()'

#### Evidence Trail:
- `+70%` **[COMMAND_STRUCT]** Struct 'CaptureEmailCommand' encapsulates executable command operation with 'Execute()' -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/capture_email.go:14:1`

### #119 COMMAND on `AdminLoginCommand`
- **Category:** `behavioral`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/admin_login.go:16:1`
- **Summary:** Struct 'AdminLoginCommand' encapsulates executable command operation with 'Execute()'

#### Evidence Trail:
- `+70%` **[COMMAND_STRUCT]** Struct 'AdminLoginCommand' encapsulates executable command operation with 'Execute()' -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/admin_login.go:16:1`

### #120 COMMAND on `CancelSubscriptionCommand`
- **Category:** `behavioral`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/cancel_subscription.go:13:1`
- **Summary:** Struct 'CancelSubscriptionCommand' encapsulates executable command operation with 'Execute()'

#### Evidence Trail:
- `+70%` **[COMMAND_STRUCT]** Struct 'CancelSubscriptionCommand' encapsulates executable command operation with 'Execute()' -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/cancel_subscription.go:13:1`

### #121 COMMAND on `AcceptWinbackOfferCommand`
- **Category:** `behavioral`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/accept_winback_offer.go:12:1`
- **Summary:** Struct 'AcceptWinbackOfferCommand' encapsulates executable command operation with 'Execute()'

#### Evidence Trail:
- `+70%` **[COMMAND_STRUCT]** Struct 'AcceptWinbackOfferCommand' encapsulates executable command operation with 'Execute()' -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/accept_winback_offer.go:12:1`

### #122 COMMAND on `VerifyIAPCommand`
- **Category:** `behavioral`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/verify_iap.go:49:1`
- **Summary:** Struct 'VerifyIAPCommand' encapsulates executable command operation with 'Execute()'

#### Evidence Trail:
- `+70%` **[COMMAND_STRUCT]** Struct 'VerifyIAPCommand' encapsulates executable command operation with 'Execute()' -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/verify_iap.go:49:1`

### #123 COMMAND on `CreateGracePeriodCommand`
- **Category:** `behavioral`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/create_grace_period.go:13:1`
- **Summary:** Struct 'CreateGracePeriodCommand' encapsulates executable command operation with 'Execute()'

#### Evidence Trail:
- `+70%` **[COMMAND_STRUCT]** Struct 'CreateGracePeriodCommand' encapsulates executable command operation with 'Execute()' -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/create_grace_period.go:13:1`

### #124 MEDIATOR on `Experiment`
- **Category:** `behavioral`
- **Confidence:** **55%** [MEDIUM]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:912:1`
- **Summary:** Maintains decoupled participant registry 'ObjectiveWeights: *map[string]float64'

#### Evidence Trail:
- `+55%` **[MEDIATOR_PARTICIPANTS_MAP]** Maintains decoupled participant registry 'ObjectiveWeights: *map[string]float64' -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:912:1`

### #125 MEDIATOR on `CredentialResolver`
- **Category:** `behavioral`
- **Confidence:** **55%** [MEDIUM]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/credential_resolver.go:22:1`
- **Summary:** Maintains decoupled participant registry 'cache: map[string]*cachedCred'

#### Evidence Trail:
- `+55%` **[MEDIATOR_PARTICIPANTS_MAP]** Maintains decoupled participant registry 'cache: map[string]*cachedCred' -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/credential_resolver.go:22:1`

### #126 MEDIATOR on `FeatureFlagService`
- **Category:** `behavioral`
- **Confidence:** **55%** [MEDIUM]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/feature_flag_service.go:24:1`
- **Summary:** Maintains decoupled participant registry 'flags: map[string]*FeatureFlag'

#### Evidence Trail:
- `+55%` **[MEDIATOR_PARTICIPANTS_MAP]** Maintains decoupled participant registry 'flags: map[string]*FeatureFlag' -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/feature_flag_service.go:24:1`

### #127 CONTEXT_PROPAGATION on `mustInitDB`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/cmd/api/main.go:159:1`
- **Summary:** Function 'mustInitDB' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'mustInitDB' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/cmd/api/main.go:159:1`

### #128 CONTEXT_PROPAGATION on `mustInitRedis`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/cmd/api/main.go:174:1`
- **Summary:** Function 'mustInitRedis' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'mustInitRedis' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/cmd/api/main.go:174:1`

### #129 CONTEXT_PROPAGATION on `WithAppID`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/appctx/appctx.go:12:1`
- **Summary:** Function 'WithAppID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'WithAppID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/appctx/appctx.go:12:1`

### #130 CONTEXT_PROPAGATION on `AppIDFromCtx`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/appctx/appctx.go:17:1`
- **Summary:** Function 'AppIDFromCtx' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'AppIDFromCtx' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/appctx/appctx.go:17:1`

### #131 CONTEXT_PROPAGATION on `MustAppIDFromCtx`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/appctx/appctx.go:23:1`
- **Summary:** Function 'MustAppIDFromCtx' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'MustAppIDFromCtx' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/appctx/appctx.go:23:1`

### #132 CONTEXT_PROPAGATION on `RevokeToken`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/middleware/jwt.go:267:1`
- **Summary:** Function 'RevokeToken' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'RevokeToken' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/application/middleware/jwt.go:267:1`

### #133 CONTEXT_PROPAGATION on `IsRevoked`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/middleware/jwt.go:272:1`
- **Summary:** Function 'IsRevoked' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'IsRevoked' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/application/middleware/jwt.go:272:1`

### #134 CONTEXT_PROPAGATION on `Execute`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/register.go:31:1`
- **Summary:** Function 'Execute' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Execute' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/register.go:31:1`

### #135 CONTEXT_PROPAGATION on `Execute`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/resolve_grace_period.go:36:1`
- **Summary:** Function 'Execute' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Execute' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/resolve_grace_period.go:36:1`

### #136 CONTEXT_PROPAGATION on `Execute`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/track_session.go:22:1`
- **Summary:** Function 'Execute' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Execute' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/track_session.go:22:1`

### #137 CONTEXT_PROPAGATION on `Execute`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/capture_email.go:22:1`
- **Summary:** Function 'Execute' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Execute' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/capture_email.go:22:1`

### #138 CONTEXT_PROPAGATION on `Execute`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/admin_login.go:36:1`
- **Summary:** Function 'Execute' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Execute' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/admin_login.go:36:1`

### #139 CONTEXT_PROPAGATION on `Execute`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/cancel_subscription.go:25:1`
- **Summary:** Function 'Execute' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Execute' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/cancel_subscription.go:25:1`

### #140 CONTEXT_PROPAGATION on `Execute`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/accept_winback_offer.go:40:1`
- **Summary:** Function 'Execute' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Execute' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/accept_winback_offer.go:40:1`

### #141 CONTEXT_PROPAGATION on `VerifyReceipt`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/verify_iap.go:44:1`
- **Summary:** Function 'VerifyReceipt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'VerifyReceipt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/verify_iap.go:44:1`

### #142 CONTEXT_PROPAGATION on `Execute`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/verify_iap.go:91:1`
- **Summary:** Function 'Execute' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Execute' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/verify_iap.go:91:1`

### #143 CONTEXT_PROPAGATION on `Execute`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/create_grace_period.go:42:1`
- **Summary:** Function 'Execute' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Execute' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/create_grace_period.go:42:1`

### #144 CONTEXT_PROPAGATION on `Execute`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/query/subscription.go:27:1`
- **Summary:** Function 'Execute' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Execute' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/application/query/subscription.go:27:1`

### #145 CONTEXT_PROPAGATION on `Execute`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/query/subscription.go:69:1`
- **Summary:** Function 'Execute' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Execute' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/application/query/subscription.go:69:1`

### #146 CONTEXT_PROPAGATION on `Execute`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/query/get_trigger_status.go:22:1`
- **Summary:** Function 'Execute' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Execute' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/application/query/get_trigger_status.go:22:1`

### #147 CONTEXT_PROPAGATION on `HandleCohortAggregation`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/cohort_jobs.go:83:1`
- **Summary:** Function 'HandleCohortAggregation' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'HandleCohortAggregation' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/cohort_jobs.go:83:1`

### #148 CONTEXT_PROPAGATION on `CalculateLTVFromCohorts`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/cohort_jobs.go:203:1`
- **Summary:** Function 'CalculateLTVFromCohorts' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'CalculateLTVFromCohorts' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/cohort_jobs.go:203:1`

### #149 CONTEXT_PROPAGATION on `GetCohortMetrics`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/cohort_jobs.go:236:1`
- **Summary:** Function 'GetCohortMetrics' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetCohortMetrics' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/cohort_jobs.go:236:1`

### #150 CONTEXT_PROPAGATION on `ProcessExpiredPendingRewards`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/bandit_maintenance_jobs.go:47:1`
- **Summary:** Function 'ProcessExpiredPendingRewards' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ProcessExpiredPendingRewards' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/bandit_maintenance_jobs.go:47:1`

### #151 CONTEXT_PROPAGATION on `TrimSlidingWindows`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/bandit_maintenance_jobs.go:83:1`
- **Summary:** Function 'TrimSlidingWindows' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'TrimSlidingWindows' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/bandit_maintenance_jobs.go:83:1`

### #152 CONTEXT_PROPAGATION on `CleanupOldContextData`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/bandit_maintenance_jobs.go:126:1`
- **Summary:** Function 'CleanupOldContextData' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'CleanupOldContextData' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/bandit_maintenance_jobs.go:126:1`

### #153 CONTEXT_PROPAGATION on `CalculateWinProbabilities`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/bandit_maintenance_jobs.go:163:1`
- **Summary:** Function 'CalculateWinProbabilities' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'CalculateWinProbabilities' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/bandit_maintenance_jobs.go:163:1`

### #154 CONTEXT_PROPAGATION on `RunFullMaintenance`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/bandit_maintenance_jobs.go:242:1`
- **Summary:** Function 'RunFullMaintenance' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'RunFullMaintenance' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/bandit_maintenance_jobs.go:242:1`

### #155 CONTEXT_PROPAGATION on `SyncObjectiveStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/bandit_maintenance_jobs.go:290:1`
- **Summary:** Function 'SyncObjectiveStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SyncObjectiveStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/bandit_maintenance_jobs.go:290:1`

### #156 CONTEXT_PROPAGATION on `HandleUpdateLTV`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/tasks.go:111:1`
- **Summary:** Function 'HandleUpdateLTV' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'HandleUpdateLTV' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/tasks.go:111:1`

### #157 CONTEXT_PROPAGATION on `HandleComputeAnalytics`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/tasks.go:152:1`
- **Summary:** Function 'HandleComputeAnalytics' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'HandleComputeAnalytics' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/tasks.go:152:1`

### #158 CONTEXT_PROPAGATION on `HandleProcessWebhook`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/tasks.go:221:1`
- **Summary:** Function 'HandleProcessWebhook' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'HandleProcessWebhook' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/tasks.go:221:1`

### #159 CONTEXT_PROPAGATION on `handleStripeEvent`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/tasks.go:273:1`
- **Summary:** Function 'handleStripeEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'handleStripeEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/tasks.go:273:1`

### #160 CONTEXT_PROPAGATION on `HandleSendNotification`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/tasks.go:328:1`
- **Summary:** Function 'HandleSendNotification' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'HandleSendNotification' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/tasks.go:328:1`

### #161 CONTEXT_PROPAGATION on `HandleProcessDunningAttempt`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/dunning_jobs.go:44:1`
- **Summary:** Function 'HandleProcessDunningAttempt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'HandleProcessDunningAttempt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/dunning_jobs.go:44:1`

### #162 CONTEXT_PROPAGATION on `HandleCheckPendingDunning`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/dunning_jobs.go:58:1`
- **Summary:** Function 'HandleCheckPendingDunning' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'HandleCheckPendingDunning' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/dunning_jobs.go:58:1`

### #163 CONTEXT_PROPAGATION on `HandleCalculateABTestStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/ab_test_jobs.go:35:1`
- **Summary:** Function 'HandleCalculateABTestStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'HandleCalculateABTestStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/ab_test_jobs.go:35:1`

### #164 CONTEXT_PROPAGATION on `HandleProcessExpiredWinbackOffers`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/winback_jobs.go:52:1`
- **Summary:** Function 'HandleProcessExpiredWinbackOffers' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'HandleProcessExpiredWinbackOffers' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/winback_jobs.go:52:1`

### #165 CONTEXT_PROPAGATION on `HandleCreateWinbackCampaign`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/winback_jobs.go:75:1`
- **Summary:** Function 'HandleCreateWinbackCampaign' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'HandleCreateWinbackCampaign' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/winback_jobs.go:75:1`

### #166 CONTEXT_PROPAGATION on `HandleProcessExpiredGracePeriods`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/grace_period_jobs.go:46:1`
- **Summary:** Function 'HandleProcessExpiredGracePeriods' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'HandleProcessExpiredGracePeriods' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/grace_period_jobs.go:46:1`

### #167 CONTEXT_PROPAGATION on `HandleNotifyExpiringGracePeriods`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/grace_period_jobs.go:67:1`
- **Summary:** Function 'HandleNotifyExpiringGracePeriods' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'HandleNotifyExpiringGracePeriods' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/grace_period_jobs.go:67:1`

### #168 CONTEXT_PROPAGATION on `HandleAggregateDailyMetrics`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/analytics_jobs.go:30:1`
- **Summary:** Function 'HandleAggregateDailyMetrics' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'HandleAggregateDailyMetrics' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/analytics_jobs.go:30:1`

### #169 CONTEXT_PROPAGATION on `UpdateExchangeRates`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/currency_jobs.go:36:1`
- **Summary:** Function 'UpdateExchangeRates' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UpdateExchangeRates' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/currency_jobs.go:36:1`

### #170 CONTEXT_PROPAGATION on `GetSupportedCurrencies`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/currency_jobs.go:61:1`
- **Summary:** Function 'GetSupportedCurrencies' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetSupportedCurrencies' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/currency_jobs.go:61:1`

### #171 CONTEXT_PROPAGATION on `HandleProcessBatch`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/matomo_jobs.go:55:1`
- **Summary:** Function 'HandleProcessBatch' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'HandleProcessBatch' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/matomo_jobs.go:55:1`

### #172 CONTEXT_PROPAGATION on `HandleSendEvent`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/matomo_jobs.go:130:1`
- **Summary:** Function 'HandleSendEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'HandleSendEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/matomo_jobs.go:130:1`

### #173 CONTEXT_PROPAGATION on `HandleSendEcommerce`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/matomo_jobs.go:186:1`
- **Summary:** Function 'HandleSendEcommerce' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'HandleSendEcommerce' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/matomo_jobs.go:186:1`

### #174 CONTEXT_PROPAGATION on `GetArmStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:47:1`
- **Summary:** Function 'GetArmStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetArmStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:47:1`

### #175 CONTEXT_PROPAGATION on `SetArmStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:74:1`
- **Summary:** Function 'SetArmStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SetArmStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:74:1`

### #176 CONTEXT_PROPAGATION on `GetAssignment`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:105:1`
- **Summary:** Function 'GetAssignment' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetAssignment' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:105:1`

### #177 CONTEXT_PROPAGATION on `SetAssignment`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:124:1`
- **Summary:** Function 'SetAssignment' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SetAssignment' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:124:1`

### #178 CONTEXT_PROPAGATION on `InvalidateArmStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:139:1`
- **Summary:** Function 'InvalidateArmStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'InvalidateArmStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:139:1`

### #179 CONTEXT_PROPAGATION on `SetBytes`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:150:1`
- **Summary:** Function 'SetBytes' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SetBytes' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:150:1`

### #180 CONTEXT_PROPAGATION on `GetBytes`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:158:1`
- **Summary:** Function 'GetBytes' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetBytes' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:158:1`

### #181 CONTEXT_PROPAGATION on `DeleteKey`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:170:1`
- **Summary:** Function 'DeleteKey' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'DeleteKey' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:170:1`

### #182 CONTEXT_PROPAGATION on `InvalidateAssignment`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:178:1`
- **Summary:** Function 'InvalidateAssignment' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'InvalidateAssignment' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:178:1`

### #183 CONTEXT_PROPAGATION on `BulkInvalidateAssignments`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:192:1`
- **Summary:** Function 'BulkInvalidateAssignments' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'BulkInvalidateAssignments' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:192:1`

### #184 CONTEXT_PROPAGATION on `GetArmStatsBatch`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:218:1`
- **Summary:** Function 'GetArmStatsBatch' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetArmStatsBatch' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:218:1`

### #185 CONTEXT_PROPAGATION on `SetArmStatsBatch`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:272:1`
- **Summary:** Function 'SetArmStatsBatch' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SetArmStatsBatch' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:272:1`

### #186 CONTEXT_PROPAGATION on `Ping`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:311:1`
- **Summary:** Function 'Ping' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Ping' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/bandit_cache.go:311:1`

### #187 CONTEXT_PROPAGATION on `SetRealtimeMetric`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:55:1`
- **Summary:** Function 'SetRealtimeMetric' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SetRealtimeMetric' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:55:1`

### #188 CONTEXT_PROPAGATION on `GetRealtimeMetric`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:76:1`
- **Summary:** Function 'GetRealtimeMetric' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetRealtimeMetric' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:76:1`

### #189 CONTEXT_PROPAGATION on `IncrementRealtimeMetric`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:96:1`
- **Summary:** Function 'IncrementRealtimeMetric' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'IncrementRealtimeMetric' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:96:1`

### #190 CONTEXT_PROPAGATION on `SetRealtimeMetrics`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:114:1`
- **Summary:** Function 'SetRealtimeMetrics' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SetRealtimeMetrics' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:114:1`

### #191 CONTEXT_PROPAGATION on `GetRealtimeMetrics`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:140:1`
- **Summary:** Function 'GetRealtimeMetrics' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetRealtimeMetrics' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:140:1`

### #192 CONTEXT_PROPAGATION on `SetCohortData`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:193:1`
- **Summary:** Function 'SetCohortData' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SetCohortData' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:193:1`

### #193 CONTEXT_PROPAGATION on `GetCohortData`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:215:1`
- **Summary:** Function 'GetCohortData' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetCohortData' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:215:1`

### #194 CONTEXT_PROPAGATION on `InvalidateCohort`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:235:1`
- **Summary:** Function 'InvalidateCohort' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'InvalidateCohort' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:235:1`

### #195 CONTEXT_PROPAGATION on `SetFunnelData`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:281:1`
- **Summary:** Function 'SetFunnelData' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SetFunnelData' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:281:1`

### #196 CONTEXT_PROPAGATION on `GetFunnelData`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:299:1`
- **Summary:** Function 'GetFunnelData' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetFunnelData' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:299:1`

### #197 CONTEXT_PROPAGATION on `SetLTV`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:331:1`
- **Summary:** Function 'SetLTV' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SetLTV' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:331:1`

### #198 CONTEXT_PROPAGATION on `GetLTV`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:348:1`
- **Summary:** Function 'GetLTV' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetLTV' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:348:1`

### #199 CONTEXT_PROPAGATION on `InvalidateLTV`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:368:1`
- **Summary:** Function 'InvalidateLTV' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'InvalidateLTV' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:368:1`

### #200 CONTEXT_PROPAGATION on `GetCacheStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:380:1`
- **Summary:** Function 'GetCacheStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetCacheStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:380:1`

### #201 CONTEXT_PROPAGATION on `FlushPattern`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:423:1`
- **Summary:** Function 'FlushPattern' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'FlushPattern' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:423:1`

### #202 CONTEXT_PROPAGATION on `EnqueueEvent`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/matomo_fallback.go:32:1`
- **Summary:** Function 'EnqueueEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'EnqueueEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/matomo_fallback.go:32:1`

### #203 CONTEXT_PROPAGATION on `GetPendingEvents`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/matomo_fallback.go:72:1`
- **Summary:** Function 'GetPendingEvents' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetPendingEvents' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/matomo_fallback.go:72:1`

### #204 CONTEXT_PROPAGATION on `UpdateEventStatus`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/matomo_fallback.go:132:1`
- **Summary:** Function 'UpdateEventStatus' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UpdateEventStatus' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/matomo_fallback.go:132:1`

### #205 CONTEXT_PROPAGATION on `GetFailedEvents`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/matomo_fallback.go:220:1`
- **Summary:** Function 'GetFailedEvents' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetFailedEvents' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/matomo_fallback.go:220:1`

### #206 CONTEXT_PROPAGATION on `DeleteEvent`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/matomo_fallback.go:275:1`
- **Summary:** Function 'DeleteEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'DeleteEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/matomo_fallback.go:275:1`

### #207 CONTEXT_PROPAGATION on `RetryFailedEvent`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/matomo_fallback.go:285:1`
- **Summary:** Function 'RetryFailedEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'RetryFailedEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/matomo_fallback.go:285:1`

### #208 CONTEXT_PROPAGATION on `CleanupOldSentEvents`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/matomo_fallback.go:304:1`
- **Summary:** Function 'CleanupOldSentEvents' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'CleanupOldSentEvents' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/matomo_fallback.go:304:1`

### #209 CONTEXT_PROPAGATION on `GetEventStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/matomo_fallback.go:322:1`
- **Summary:** Function 'GetEventStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetEventStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/matomo_fallback.go:322:1`

### #210 CONTEXT_PROPAGATION on `GetEventByID`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/matomo_fallback.go:359:1`
- **Summary:** Function 'GetEventByID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetEventByID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/matomo_fallback.go:359:1`

### #211 CONTEXT_PROPAGATION on `Create`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/subscription_repository_impl.go:26:1`
- **Summary:** Function 'Create' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Create' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/subscription_repository_impl.go:26:1`

### #212 CONTEXT_PROPAGATION on `GetByID`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/subscription_repository_impl.go:49:1`
- **Summary:** Function 'GetByID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetByID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/subscription_repository_impl.go:49:1`

### #213 CONTEXT_PROPAGATION on `GetActiveByUserID`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/subscription_repository_impl.go:61:1`
- **Summary:** Function 'GetActiveByUserID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetActiveByUserID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/subscription_repository_impl.go:61:1`

### #214 CONTEXT_PROPAGATION on `GetByUserID`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/subscription_repository_impl.go:77:1`
- **Summary:** Function 'GetByUserID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetByUserID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/subscription_repository_impl.go:77:1`

### #215 CONTEXT_PROPAGATION on `Update`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/subscription_repository_impl.go:93:1`
- **Summary:** Function 'Update' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Update' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/subscription_repository_impl.go:93:1`

### #216 CONTEXT_PROPAGATION on `UpdateStatus`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/subscription_repository_impl.go:113:1`
- **Summary:** Function 'UpdateStatus' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UpdateStatus' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/subscription_repository_impl.go:113:1`

### #217 CONTEXT_PROPAGATION on `UpdateExpiry`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/subscription_repository_impl.go:127:1`
- **Summary:** Function 'UpdateExpiry' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UpdateExpiry' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/subscription_repository_impl.go:127:1`

### #218 CONTEXT_PROPAGATION on `Cancel`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/subscription_repository_impl.go:141:1`
- **Summary:** Function 'Cancel' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Cancel' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/subscription_repository_impl.go:141:1`

### #219 CONTEXT_PROPAGATION on `CanAccess`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/subscription_repository_impl.go:150:1`
- **Summary:** Function 'CanAccess' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'CanAccess' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/subscription_repository_impl.go:150:1`

### #220 CONTEXT_PROPAGATION on `GetUsersWithCancelledSubscriptions`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/subscription_repository_impl.go:166:1`
- **Summary:** Function 'GetUsersWithCancelledSubscriptions' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetUsersWithCancelledSubscriptions' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/subscription_repository_impl.go:166:1`

### #221 CONTEXT_PROPAGATION on `GetTotalRevenue`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/subscription_repository_impl.go:174:1`
- **Summary:** Function 'GetTotalRevenue' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetTotalRevenue' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/subscription_repository_impl.go:174:1`

### #222 CONTEXT_PROPAGATION on `Create`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/transaction_repository_impl.go:24:1`
- **Summary:** Function 'Create' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Create' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/transaction_repository_impl.go:24:1`

### #223 CONTEXT_PROPAGATION on `GetByID`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/transaction_repository_impl.go:44:1`
- **Summary:** Function 'GetByID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetByID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/transaction_repository_impl.go:44:1`

### #224 CONTEXT_PROPAGATION on `GetByUserID`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/transaction_repository_impl.go:56:1`
- **Summary:** Function 'GetByUserID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetByUserID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/transaction_repository_impl.go:56:1`

### #225 CONTEXT_PROPAGATION on `GetBySubscriptionID`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/transaction_repository_impl.go:77:1`
- **Summary:** Function 'GetBySubscriptionID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetBySubscriptionID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/transaction_repository_impl.go:77:1`

### #226 CONTEXT_PROPAGATION on `GetSegmentedLTV`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/transaction_repository_impl.go:91:1`
- **Summary:** Function 'GetSegmentedLTV' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetSegmentedLTV' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/transaction_repository_impl.go:91:1`

### #227 CONTEXT_PROPAGATION on `CheckDuplicateReceipt`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/transaction_repository_impl.go:116:1`
- **Summary:** Function 'CheckDuplicateReceipt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'CheckDuplicateReceipt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/transaction_repository_impl.go:116:1`

### #228 CONTEXT_PROPAGATION on `GetArms`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:82:1`
- **Summary:** Function 'GetArms' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetArms' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:82:1`

### #229 CONTEXT_PROPAGATION on `GetArmStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:120:1`
- **Summary:** Function 'GetArmStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetArmStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:120:1`

### #230 CONTEXT_PROPAGATION on `UpdateArmStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:171:1`
- **Summary:** Function 'UpdateArmStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UpdateArmStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:171:1`

### #231 CONTEXT_PROPAGATION on `CreateAssignment`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:211:1`
- **Summary:** Function 'CreateAssignment' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'CreateAssignment' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:211:1`

### #232 CONTEXT_PROPAGATION on `GetActiveAssignment`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:287:1`
- **Summary:** Function 'GetActiveAssignment' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetActiveAssignment' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:287:1`

### #233 CONTEXT_PROPAGATION on `SaveConversion`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:320:1`
- **Summary:** Function 'SaveConversion' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SaveConversion' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:320:1`

### #234 CONTEXT_PROPAGATION on `AppendConversionEvent`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:332:1`
- **Summary:** Function 'AppendConversionEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'AppendConversionEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:332:1`

### #235 CONTEXT_PROPAGATION on `AppendImpressionEvent`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:379:1`
- **Summary:** Function 'AppendImpressionEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'AppendImpressionEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:379:1`

### #236 CONTEXT_PROPAGATION on `AppendWinnerRecommendationEvent`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:414:1`
- **Summary:** Function 'AppendWinnerRecommendationEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'AppendWinnerRecommendationEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:414:1`

### #237 CONTEXT_PROPAGATION on `ProcessPendingConversion`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:466:1`
- **Summary:** Function 'ProcessPendingConversion' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ProcessPendingConversion' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:466:1`

### #238 CONTEXT_PROPAGATION on `ProcessExpiredPendingReward`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:558:1`
- **Summary:** Function 'ProcessExpiredPendingReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ProcessExpiredPendingReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:558:1`

### #239 CONTEXT_PROPAGATION on `GetAssignmentHistory`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:627:1`
- **Summary:** Function 'GetAssignmentHistory' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetAssignmentHistory' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:627:1`

### #240 CONTEXT_PROPAGATION on `CleanupExpiredAssignments`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:662:1`
- **Summary:** Function 'CleanupExpiredAssignments' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'CleanupExpiredAssignments' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:662:1`

### #241 CONTEXT_PROPAGATION on `ListWindowMaintenanceExperimentIDs`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:679:1`
- **Summary:** Function 'ListWindowMaintenanceExperimentIDs' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ListWindowMaintenanceExperimentIDs' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:679:1`

### #242 CONTEXT_PROPAGATION on `ListObjectiveSyncExperimentIDs`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:721:1`
- **Summary:** Function 'ListObjectiveSyncExperimentIDs' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ListObjectiveSyncExperimentIDs' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:721:1`

### #243 CONTEXT_PROPAGATION on `CleanupStaleUserContext`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:762:1`
- **Summary:** Function 'CleanupStaleUserContext' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'CleanupStaleUserContext' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:762:1`

### #244 CONTEXT_PROPAGATION on `GetAllArmStatsForExperiment`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:779:1`
- **Summary:** Function 'GetAllArmStatsForExperiment' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetAllArmStatsForExperiment' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:779:1`

### #245 CONTEXT_PROPAGATION on `CreateExperiment`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:815:1`
- **Summary:** Function 'CreateExperiment' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'CreateExperiment' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:815:1`

### #246 CONTEXT_PROPAGATION on `GetExperiment`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:842:1`
- **Summary:** Function 'GetExperiment' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetExperiment' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:842:1`

### #247 CONTEXT_PROPAGATION on `CreateArm`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:889:1`
- **Summary:** Function 'CreateArm' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'CreateArm' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:889:1`

### #248 CONTEXT_PROPAGATION on `GetExperimentConfig`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:944:1`
- **Summary:** Function 'GetExperimentConfig' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetExperimentConfig' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:944:1`

### #249 CONTEXT_PROPAGATION on `UpdateObjectiveConfig`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1003:1`
- **Summary:** Function 'UpdateObjectiveConfig' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UpdateObjectiveConfig' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1003:1`

### #250 CONTEXT_PROPAGATION on `GetUserContext`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1038:1`
- **Summary:** Function 'GetUserContext' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetUserContext' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1038:1`

### #251 CONTEXT_PROPAGATION on `SetUserContext`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1071:1`
- **Summary:** Function 'SetUserContext' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SetUserContext' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1071:1`

### #252 CONTEXT_PROPAGATION on `GetObjectiveStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1108:1`
- **Summary:** Function 'GetObjectiveStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetObjectiveStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1108:1`

### #253 CONTEXT_PROPAGATION on `UpdateObjectiveStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1149:1`
- **Summary:** Function 'UpdateObjectiveStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UpdateObjectiveStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1149:1`

### #254 CONTEXT_PROPAGATION on `GetAllObjectiveStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1183:1`
- **Summary:** Function 'GetAllObjectiveStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetAllObjectiveStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1183:1`

### #255 CONTEXT_PROPAGATION on `CreatePendingReward`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1222:1`
- **Summary:** Function 'CreatePendingReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'CreatePendingReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1222:1`

### #256 CONTEXT_PROPAGATION on `GetPendingReward`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1245:1`
- **Summary:** Function 'GetPendingReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetPendingReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1245:1`

### #257 CONTEXT_PROPAGATION on `GetPendingRewardsByUser`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1268:1`
- **Summary:** Function 'GetPendingRewardsByUser' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetPendingRewardsByUser' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1268:1`

### #258 CONTEXT_PROPAGATION on `GetExpiredPendingRewards`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1301:1`
- **Summary:** Function 'GetExpiredPendingRewards' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetExpiredPendingRewards' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1301:1`

### #259 CONTEXT_PROPAGATION on `UpdatePendingReward`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1330:1`
- **Summary:** Function 'UpdatePendingReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UpdatePendingReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1330:1`

### #260 CONTEXT_PROPAGATION on `LinkConversion`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1358:1`
- **Summary:** Function 'LinkConversion' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'LinkConversion' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1358:1`

### #261 CONTEXT_PROPAGATION on `applyRewardToArmTx`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1373:1`
- **Summary:** Function 'applyRewardToArmTx' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'applyRewardToArmTx' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1373:1`

### #262 CONTEXT_PROPAGATION on `loadArmStatsTx`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1411:1`
- **Summary:** Function 'loadArmStatsTx' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'loadArmStatsTx' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1411:1`

### #263 CONTEXT_PROPAGATION on `insertConversionEventTx`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1437:1`
- **Summary:** Function 'insertConversionEventTx' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'insertConversionEventTx' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1437:1`

### #264 CONTEXT_PROPAGATION on `GetByTransactionID`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1499:1`
- **Summary:** Function 'GetByTransactionID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetByTransactionID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:1499:1`

### #265 CONTEXT_PROPAGATION on `GetRevenueBetween`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/analytics_repository_impl.go:26:1`
- **Summary:** Function 'GetRevenueBetween' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetRevenueBetween' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/analytics_repository_impl.go:26:1`

### #266 CONTEXT_PROPAGATION on `GetMRR`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/analytics_repository_impl.go:42:1`
- **Summary:** Function 'GetMRR' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetMRR' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/analytics_repository_impl.go:42:1`

### #267 CONTEXT_PROPAGATION on `GetActiveSubscriptionCountAt`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/analytics_repository_impl.go:71:1`
- **Summary:** Function 'GetActiveSubscriptionCountAt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetActiveSubscriptionCountAt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/analytics_repository_impl.go:71:1`

### #268 CONTEXT_PROPAGATION on `GetChurnedCountBetween`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/analytics_repository_impl.go:87:1`
- **Summary:** Function 'GetChurnedCountBetween' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetChurnedCountBetween' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/analytics_repository_impl.go:87:1`

### #269 CONTEXT_PROPAGATION on `GetMRRTrend`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/analytics_repository_impl.go:104:1`
- **Summary:** Function 'GetMRRTrend' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetMRRTrend' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/analytics_repository_impl.go:104:1`

### #270 CONTEXT_PROPAGATION on `GetSubscriptionStatusCounts`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/analytics_repository_impl.go:160:1`
- **Summary:** Function 'GetSubscriptionStatusCounts' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetSubscriptionStatusCounts' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/analytics_repository_impl.go:160:1`

### #271 CONTEXT_PROPAGATION on `GetChurnRiskCount`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/analytics_repository_impl.go:183:1`
- **Summary:** Function 'GetChurnRiskCount' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetChurnRiskCount' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/analytics_repository_impl.go:183:1`

### #272 CONTEXT_PROPAGATION on `GetWebhookHealthByProvider`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/analytics_repository_impl.go:199:1`
- **Summary:** Function 'GetWebhookHealthByProvider' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetWebhookHealthByProvider' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/analytics_repository_impl.go:199:1`

### #273 CONTEXT_PROPAGATION on `GetRecentAuditLog`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/analytics_repository_impl.go:227:1`
- **Summary:** Function 'GetRecentAuditLog' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetRecentAuditLog' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/analytics_repository_impl.go:227:1`

### #274 CONTEXT_PROPAGATION on `GetAuditLogPaginated`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/analytics_repository_impl.go:268:1`
- **Summary:** Function 'GetAuditLogPaginated' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetAuditLogPaginated' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/analytics_repository_impl.go:268:1`

### #275 CONTEXT_PROPAGATION on `ensureDraftExperimentPricingTiersExist`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:193:1`
- **Summary:** Function 'ensureDraftExperimentPricingTiersExist' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ensureDraftExperimentPricingTiersExist' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:193:1`

### #276 CONTEXT_PROPAGATION on `insertExperimentLifecycleAudit`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:241:1`
- **Summary:** Function 'insertExperimentLifecycleAudit' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'insertExperimentLifecycleAudit' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:241:1`

### #277 CONTEXT_PROPAGATION on `GetExperimentMutationState`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:27:1`
- **Summary:** Function 'GetExperimentMutationState' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetExperimentMutationState' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:27:1`

### #278 CONTEXT_PROPAGATION on `UpdateExperimentDraft`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:57:1`
- **Summary:** Function 'UpdateExperimentDraft' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UpdateExperimentDraft' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:57:1`

### #279 CONTEXT_PROPAGATION on `syncDraftExperimentArms`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:112:1`
- **Summary:** Function 'syncDraftExperimentArms' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'syncDraftExperimentArms' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:112:1`

### #280 CONTEXT_PROPAGATION on `UpdateExperimentStatus`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:216:1`
- **Summary:** Function 'UpdateExperimentStatus' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UpdateExperimentStatus' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:216:1`

### #281 CONTEXT_PROPAGATION on `UpdateExperimentAutomationPolicy`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:299:1`
- **Summary:** Function 'UpdateExperimentAutomationPolicy' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UpdateExperimentAutomationPolicy' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:299:1`

### #282 CONTEXT_PROPAGATION on `UpdateExperimentStatusWithAudit`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:319:1`
- **Summary:** Function 'UpdateExperimentStatusWithAudit' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UpdateExperimentStatusWithAudit' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:319:1`

### #283 CONTEXT_PROPAGATION on `UpdateExperimentStatusAndAutomationPolicyWithAudit`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:351:1`
- **Summary:** Function 'UpdateExperimentStatusAndAutomationPolicyWithAudit' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UpdateExperimentStatusAndAutomationPolicyWithAudit' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:351:1`

### #284 CONTEXT_PROPAGATION on `ListExperimentAutomationStates`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:394:1`
- **Summary:** Function 'ListExperimentAutomationStates' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ListExperimentAutomationStates' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:394:1`

### #285 CONTEXT_PROPAGATION on `CountExperimentAssignments`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:463:1`
- **Summary:** Function 'CountExperimentAssignments' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'CountExperimentAssignments' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:463:1`

### #286 CONTEXT_PROPAGATION on `EnsureExperimentArmStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:474:1`
- **Summary:** Function 'EnsureExperimentArmStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'EnsureExperimentArmStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:474:1`

### #287 CONTEXT_PROPAGATION on `CountExperimentPendingRewards`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:488:1`
- **Summary:** Function 'CountExperimentPendingRewards' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'CountExperimentPendingRewards' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:488:1`

### #288 CONTEXT_PROPAGATION on `ProcessExpiredPendingRewards`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:501:1`
- **Summary:** Function 'ProcessExpiredPendingRewards' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ProcessExpiredPendingRewards' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:501:1`

### #289 CONTEXT_PROPAGATION on `UpdateExperimentWinnerConfidence`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:565:1`
- **Summary:** Function 'UpdateExperimentWinnerConfidence' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UpdateExperimentWinnerConfidence' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:565:1`

### #290 CONTEXT_PROPAGATION on `GetExperimentObjectiveConfig`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:580:1`
- **Summary:** Function 'GetExperimentObjectiveConfig' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetExperimentObjectiveConfig' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:580:1`

### #291 CONTEXT_PROPAGATION on `ListExperimentRepairCandidateIDs`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:607:1`
- **Summary:** Function 'ListExperimentRepairCandidateIDs' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ListExperimentRepairCandidateIDs' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:607:1`

### #292 CONTEXT_PROPAGATION on `ClaimAutomationJobRun`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/automation_job_run_repository.go:22:1`
- **Summary:** Function 'ClaimAutomationJobRun' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ClaimAutomationJobRun' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/automation_job_run_repository.go:22:1`

### #293 CONTEXT_PROPAGATION on `FinishAutomationJobRun`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/automation_job_run_repository.go:72:1`
- **Summary:** Function 'FinishAutomationJobRun' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'FinishAutomationJobRun' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/automation_job_run_repository.go:72:1`

### #294 CONTEXT_PROPAGATION on `SetPassword`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/admin_credential_repository_impl.go:22:1`
- **Summary:** Function 'SetPassword' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SetPassword' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/admin_credential_repository_impl.go:22:1`

### #295 CONTEXT_PROPAGATION on `GetPasswordHash`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/admin_credential_repository_impl.go:33:1`
- **Summary:** Function 'GetPasswordHash' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetPasswordHash' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/admin_credential_repository_impl.go:33:1`

### #296 CONTEXT_PROPAGATION on `Create`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/winback_offer_repository_impl.go:24:1`
- **Summary:** Function 'Create' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Create' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/winback_offer_repository_impl.go:24:1`

### #297 CONTEXT_PROPAGATION on `GetByID`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/winback_offer_repository_impl.go:46:1`
- **Summary:** Function 'GetByID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetByID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/winback_offer_repository_impl.go:46:1`

### #298 CONTEXT_PROPAGATION on `GetActiveByUserID`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/winback_offer_repository_impl.go:75:1`
- **Summary:** Function 'GetActiveByUserID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetActiveByUserID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/winback_offer_repository_impl.go:75:1`

### #299 CONTEXT_PROPAGATION on `GetActiveByUserAndCampaign`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/winback_offer_repository_impl.go:114:1`
- **Summary:** Function 'GetActiveByUserAndCampaign' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetActiveByUserAndCampaign' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/winback_offer_repository_impl.go:114:1`

### #300 CONTEXT_PROPAGATION on `GetActiveByCampaignID`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/winback_offer_repository_impl.go:144:1`
- **Summary:** Function 'GetActiveByCampaignID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetActiveByCampaignID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/winback_offer_repository_impl.go:144:1`

### #301 CONTEXT_PROPAGATION on `Update`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/winback_offer_repository_impl.go:183:1`
- **Summary:** Function 'Update' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Update' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/winback_offer_repository_impl.go:183:1`

### #302 CONTEXT_PROPAGATION on `GetExpiredOffers`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/winback_offer_repository_impl.go:200:1`
- **Summary:** Function 'GetExpiredOffers' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetExpiredOffers' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/winback_offer_repository_impl.go:200:1`

### #303 CONTEXT_PROPAGATION on `Create`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/dunning_repository_impl.go:28:1`
- **Summary:** Function 'Create' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Create' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/dunning_repository_impl.go:28:1`

### #304 CONTEXT_PROPAGATION on `GetByID`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/dunning_repository_impl.go:44:1`
- **Summary:** Function 'GetByID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetByID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/dunning_repository_impl.go:44:1`

### #305 CONTEXT_PROPAGATION on `GetActiveBySubscriptionID`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/dunning_repository_impl.go:64:1`
- **Summary:** Function 'GetActiveBySubscriptionID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetActiveBySubscriptionID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/dunning_repository_impl.go:64:1`

### #306 CONTEXT_PROPAGATION on `Update`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/dunning_repository_impl.go:85:1`
- **Summary:** Function 'Update' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Update' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/dunning_repository_impl.go:85:1`

### #307 CONTEXT_PROPAGATION on `GetPendingAttempts`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/dunning_repository_impl.go:101:1`
- **Summary:** Function 'GetPendingAttempts' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetPendingAttempts' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/dunning_repository_impl.go:101:1`

### #308 CONTEXT_PROPAGATION on `Create`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/grace_period_repository_impl.go:24:1`
- **Summary:** Function 'Create' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Create' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/grace_period_repository_impl.go:24:1`

### #309 CONTEXT_PROPAGATION on `GetByID`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/grace_period_repository_impl.go:44:1`
- **Summary:** Function 'GetByID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetByID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/grace_period_repository_impl.go:44:1`

### #310 CONTEXT_PROPAGATION on `GetActiveByUserID`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/grace_period_repository_impl.go:71:1`
- **Summary:** Function 'GetActiveByUserID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetActiveByUserID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/grace_period_repository_impl.go:71:1`

### #311 CONTEXT_PROPAGATION on `GetActiveBySubscriptionID`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/grace_period_repository_impl.go:100:1`
- **Summary:** Function 'GetActiveBySubscriptionID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetActiveBySubscriptionID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/grace_period_repository_impl.go:100:1`

### #312 CONTEXT_PROPAGATION on `Update`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/grace_period_repository_impl.go:129:1`
- **Summary:** Function 'Update' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Update' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/grace_period_repository_impl.go:129:1`

### #313 CONTEXT_PROPAGATION on `GetExpiredGracePeriods`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/grace_period_repository_impl.go:147:1`
- **Summary:** Function 'GetExpiredGracePeriods' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetExpiredGracePeriods' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/grace_period_repository_impl.go:147:1`

### #314 CONTEXT_PROPAGATION on `GetExpiringSoon`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/grace_period_repository_impl.go:185:1`
- **Summary:** Function 'GetExpiringSoon' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetExpiringSoon' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/grace_period_repository_impl.go:185:1`

### #315 CONTEXT_PROPAGATION on `Create`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:26:1`
- **Summary:** Function 'Create' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Create' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:26:1`

### #316 CONTEXT_PROPAGATION on `GetByID`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:46:1`
- **Summary:** Function 'GetByID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetByID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:46:1`

### #317 CONTEXT_PROPAGATION on `GetByPlatformID`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:58:1`
- **Summary:** Function 'GetByPlatformID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetByPlatformID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:58:1`

### #318 CONTEXT_PROPAGATION on `GetByEmail`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:74:1`
- **Summary:** Function 'GetByEmail' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetByEmail' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:74:1`

### #319 CONTEXT_PROPAGATION on `Update`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:86:1`
- **Summary:** Function 'Update' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Update' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:86:1`

### #320 CONTEXT_PROPAGATION on `SoftDelete`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:108:1`
- **Summary:** Function 'SoftDelete' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SoftDelete' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:108:1`

### #321 CONTEXT_PROPAGATION on `ExistsByPlatformID`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:117:1`
- **Summary:** Function 'ExistsByPlatformID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ExistsByPlatformID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:117:1`

### #322 CONTEXT_PROPAGATION on `ExistsByPlatformIDAndApp`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:133:1`
- **Summary:** Function 'ExistsByPlatformIDAndApp' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ExistsByPlatformIDAndApp' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:133:1`

### #323 CONTEXT_PROPAGATION on `UpdatePurchaseChannel`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:147:1`
- **Summary:** Function 'UpdatePurchaseChannel' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UpdatePurchaseChannel' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:147:1`

### #324 CONTEXT_PROPAGATION on `UpdateEmail`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:158:1`
- **Summary:** Function 'UpdateEmail' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UpdateEmail' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:158:1`

### #325 CONTEXT_PROPAGATION on `IncrementLTV`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:169:1`
- **Summary:** Function 'IncrementLTV' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'IncrementLTV' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:169:1`

### #326 CONTEXT_PROPAGATION on `IncrementSessionCount`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:186:1`
- **Summary:** Function 'IncrementSessionCount' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'IncrementSessionCount' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:186:1`

### #327 CONTEXT_PROPAGATION on `UpdateHasViewedAds`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:194:1`
- **Summary:** Function 'UpdateHasViewedAds' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UpdateHasViewedAds' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/user_repository_impl.go:194:1`

### #328 CONTEXT_PROPAGATION on `GetByID`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:36:1`
- **Summary:** Function 'GetByID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetByID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:36:1`

### #329 CONTEXT_PROPAGATION on `GetByBundleID`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:42:1`
- **Summary:** Function 'GetByBundleID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetByBundleID' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:42:1`

### #330 CONTEXT_PROPAGATION on `List`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:48:1`
- **Summary:** Function 'List' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'List' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:48:1`

### #331 CONTEXT_PROPAGATION on `Create`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:90:1`
- **Summary:** Function 'Create' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Create' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:90:1`

### #332 CONTEXT_PROPAGATION on `Update`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:99:1`
- **Summary:** Function 'Update' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Update' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:99:1`

### #333 CONTEXT_PROPAGATION on `Delete`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:111:1`
- **Summary:** Function 'Delete' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Delete' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:111:1`

### #334 CONTEXT_PROPAGATION on `GetSettings`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:122:1`
- **Summary:** Function 'GetSettings' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetSettings' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:122:1`

### #335 CONTEXT_PROPAGATION on `UpdateSettings`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:138:1`
- **Summary:** Function 'UpdateSettings' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UpdateSettings' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:138:1`

### #336 CONTEXT_PROPAGATION on `GetCredentials`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:156:1`
- **Summary:** Function 'GetCredentials' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetCredentials' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:156:1`

### #337 CONTEXT_PROPAGATION on `UpsertCredentials`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:184:1`
- **Summary:** Function 'UpsertCredentials' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UpsertCredentials' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:184:1`

### #338 CONTEXT_PROPAGATION on `DeleteCredentials`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:266:1`
- **Summary:** Function 'DeleteCredentials' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'DeleteCredentials' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:266:1`

### #339 CONTEXT_PROPAGATION on `GetCredentialsByProvider`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:272:1`
- **Summary:** Function 'GetCredentialsByProvider' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetCredentialsByProvider' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:272:1`

### #340 CONTEXT_PROPAGATION on `NewPool`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/pool/pgxpool.go:12:1`
- **Summary:** Function 'NewPool' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'NewPool' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/pool/pgxpool.go:12:1`

### #341 CONTEXT_PROPAGATION on `Ping`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/pool/pgxpool.go:36:1`
- **Summary:** Function 'Ping' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Ping' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/pool/pgxpool.go:36:1`

### #342 CONTEXT_PROPAGATION on `Resolve`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/credential_resolver.go:40:1`
- **Summary:** Function 'Resolve' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Resolve' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/credential_resolver.go:40:1`

### #343 CONTEXT_PROPAGATION on `VerifyReceipt`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/google_verifier.go:33:1`
- **Summary:** Function 'VerifyReceipt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'VerifyReceipt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/google_verifier.go:33:1`

### #344 CONTEXT_PROPAGATION on `VerifyReceipt`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/adapter.go:37:1`
- **Summary:** Function 'VerifyReceipt' adheres to idiomatic Go context propagation passing '_: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'VerifyReceipt' adheres to idiomatic Go context propagation passing '_: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/adapter.go:37:1`

### #345 CONTEXT_PROPAGATION on `VerifyAppleReceipt`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/adapter.go:45:1`
- **Summary:** Function 'VerifyAppleReceipt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'VerifyAppleReceipt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/adapter.go:45:1`

### #346 CONTEXT_PROPAGATION on `VerifyGoogleReceipt`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/adapter.go:62:1`
- **Summary:** Function 'VerifyGoogleReceipt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'VerifyGoogleReceipt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/adapter.go:62:1`

### #347 CONTEXT_PROPAGATION on `VerifyReceipt`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/adapter.go:87:1`
- **Summary:** Function 'VerifyReceipt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'VerifyReceipt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/adapter.go:87:1`

### #348 CONTEXT_PROPAGATION on `VerifyReceipt`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/adapter.go:100:1`
- **Summary:** Function 'VerifyReceipt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'VerifyReceipt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/adapter.go:100:1`

### #349 CONTEXT_PROPAGATION on `VerifyReceipt`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/dynamic_verifiers.go:22:1`
- **Summary:** Function 'VerifyReceipt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'VerifyReceipt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/dynamic_verifiers.go:22:1`

### #350 CONTEXT_PROPAGATION on `VerifyReceipt`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/dynamic_verifiers.go:54:1`
- **Summary:** Function 'VerifyReceipt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'VerifyReceipt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/dynamic_verifiers.go:54:1`

### #351 CONTEXT_PROPAGATION on `VerifyReceipt`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/verifier.go:100:1`
- **Summary:** Function 'VerifyReceipt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'VerifyReceipt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/verifier.go:100:1`

### #352 CONTEXT_PROPAGATION on `TrackEvent`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:71:1`
- **Summary:** Function 'TrackEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'TrackEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:71:1`

### #353 CONTEXT_PROPAGATION on `TrackEcommerce`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:129:1`
- **Summary:** Function 'TrackEcommerce' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'TrackEcommerce' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:129:1`

### #354 CONTEXT_PROPAGATION on `GetCohorts`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:198:1`
- **Summary:** Function 'GetCohorts' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetCohorts' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:198:1`

### #355 CONTEXT_PROPAGATION on `GetFunnels`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:247:1`
- **Summary:** Function 'GetFunnels' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetFunnels' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:247:1`

### #356 CONTEXT_PROPAGATION on `GetRealtimeVisitors`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:279:1`
- **Summary:** Function 'GetRealtimeVisitors' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetRealtimeVisitors' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:279:1`

### #357 CONTEXT_PROPAGATION on `doRequest`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:302:1`
- **Summary:** Function 'doRequest' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'doRequest' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:302:1`

### #358 CONTEXT_PROPAGATION on `doJSONRequest`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:352:1`
- **Summary:** Function 'doJSONRequest' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'doJSONRequest' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:352:1`

### #359 CONTEXT_PROPAGATION on `HealthCheck`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:411:1`
- **Summary:** Function 'HealthCheck' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'HealthCheck' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:411:1`

### #360 CONTEXT_PROPAGATION on `CreateGracePeriod`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/grace_period_service.go:42:1`
- **Summary:** Function 'CreateGracePeriod' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'CreateGracePeriod' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/grace_period_service.go:42:1`

### #361 CONTEXT_PROPAGATION on `ResolveGracePeriod`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/grace_period_service.go:79:1`
- **Summary:** Function 'ResolveGracePeriod' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ResolveGracePeriod' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/grace_period_service.go:79:1`

### #362 CONTEXT_PROPAGATION on `ExpireGracePeriod`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/grace_period_service.go:112:1`
- **Summary:** Function 'ExpireGracePeriod' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ExpireGracePeriod' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/grace_period_service.go:112:1`

### #363 CONTEXT_PROPAGATION on `ProcessExpiredGracePeriods`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/grace_period_service.go:145:1`
- **Summary:** Function 'ProcessExpiredGracePeriods' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ProcessExpiredGracePeriods' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/grace_period_service.go:145:1`

### #364 CONTEXT_PROPAGATION on `GetGracePeriodStatus`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/grace_period_service.go:165:1`
- **Summary:** Function 'GetGracePeriodStatus' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetGracePeriodStatus' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/grace_period_service.go:165:1`

### #365 CONTEXT_PROPAGATION on `NotifyExpiringSoon`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/grace_period_service.go:175:1`
- **Summary:** Function 'NotifyExpiringSoon' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'NotifyExpiringSoon' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/grace_period_service.go:175:1`

### #366 CONTEXT_PROPAGATION on `getExperimentConfig`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:142:1`
- **Summary:** Function 'getExperimentConfig' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'getExperimentConfig' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:142:1`

### #367 CONTEXT_PROPAGATION on `getHybridStrategy`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:173:1`
- **Summary:** Function 'getHybridStrategy' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'getHybridStrategy' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:173:1`

### #368 CONTEXT_PROPAGATION on `getWindowStrategy`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:189:1`
- **Summary:** Function 'getWindowStrategy' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'getWindowStrategy' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:189:1`

### #369 CONTEXT_PROPAGATION on `SelectArm`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:209:1`
- **Summary:** Function 'SelectArm' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SelectArm' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:209:1`

### #370 CONTEXT_PROPAGATION on `RecordReward`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:271:1`
- **Summary:** Function 'RecordReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'RecordReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:271:1`

### #371 CONTEXT_PROPAGATION on `ProcessConversion`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:365:1`
- **Summary:** Function 'ProcessConversion' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ProcessConversion' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:365:1`

### #372 CONTEXT_PROPAGATION on `GetArmStatistics`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:389:1`
- **Summary:** Function 'GetArmStatistics' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetArmStatistics' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:389:1`

### #373 CONTEXT_PROPAGATION on `GetObjectiveScores`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:397:1`
- **Summary:** Function 'GetObjectiveScores' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetObjectiveScores' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:397:1`

### #374 CONTEXT_PROPAGATION on `SetObjectiveConfig`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:428:1`
- **Summary:** Function 'SetObjectiveConfig' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SetObjectiveConfig' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:428:1`

### #375 CONTEXT_PROPAGATION on `GetMetrics`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:464:1`
- **Summary:** Function 'GetMetrics' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetMetrics' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:464:1`

### #376 CONTEXT_PROPAGATION on `GetWindowInfo`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:505:1`
- **Summary:** Function 'GetWindowInfo' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetWindowInfo' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:505:1`

### #377 CONTEXT_PROPAGATION on `TrimWindow`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:531:1`
- **Summary:** Function 'TrimWindow' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'TrimWindow' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:531:1`

### #378 CONTEXT_PROPAGATION on `ProcessExpiredPendingRewards`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:551:1`
- **Summary:** Function 'ProcessExpiredPendingRewards' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ProcessExpiredPendingRewards' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:551:1`

### #379 CONTEXT_PROPAGATION on `TrimConfiguredWindows`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:571:1`
- **Summary:** Function 'TrimConfiguredWindows' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'TrimConfiguredWindows' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:571:1`

### #380 CONTEXT_PROPAGATION on `SyncObjectiveStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:608:1`
- **Summary:** Function 'SyncObjectiveStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SyncObjectiveStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:608:1`

### #381 CONTEXT_PROPAGATION on `CleanupOldContextData`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:670:1`
- **Summary:** Function 'CleanupOldContextData' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'CleanupOldContextData' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:670:1`

### #382 CONTEXT_PROPAGATION on `CleanupExpiredAssignments`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:683:1`
- **Summary:** Function 'CleanupExpiredAssignments' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'CleanupExpiredAssignments' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:683:1`

### #383 CONTEXT_PROPAGATION on `ExportWindowEvents`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:696:1`
- **Summary:** Function 'ExportWindowEvents' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ExportWindowEvents' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:696:1`

### #384 CONTEXT_PROPAGATION on `GetObjectiveConfig`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:723:1`
- **Summary:** Function 'GetObjectiveConfig' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetObjectiveConfig' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:723:1`

### #385 CONTEXT_PROPAGATION on `GetPendingReward`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:727:1`
- **Summary:** Function 'GetPendingReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetPendingReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:727:1`

### #386 CONTEXT_PROPAGATION on `GetUserPendingRewards`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:736:1`
- **Summary:** Function 'GetUserPendingRewards' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetUserPendingRewards' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:736:1`

### #387 CONTEXT_PROPAGATION on `RunMaintenanceDetailed`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:794:1`
- **Summary:** Function 'RunMaintenanceDetailed' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'RunMaintenanceDetailed' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:794:1`

### #388 CONTEXT_PROPAGATION on `RunMaintenance`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:852:1`
- **Summary:** Function 'RunMaintenance' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'RunMaintenance' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:852:1`

### #389 CONTEXT_PROPAGATION on `UpdateDraftExperiment`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:172:1`
- **Summary:** Function 'UpdateDraftExperiment' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UpdateDraftExperiment' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:172:1`

### #390 CONTEXT_PROPAGATION on `UpdateExperimentAutomationPolicy`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:183:1`
- **Summary:** Function 'UpdateExperimentAutomationPolicy' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UpdateExperimentAutomationPolicy' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:183:1`

### #391 CONTEXT_PROPAGATION on `TransitionExperimentStatus`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:204:1`
- **Summary:** Function 'TransitionExperimentStatus' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'TransitionExperimentStatus' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:204:1`

### #392 CONTEXT_PROPAGATION on `TransitionExperimentStatusWithAudit`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:208:1`
- **Summary:** Function 'TransitionExperimentStatusWithAudit' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'TransitionExperimentStatusWithAudit' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:208:1`

### #393 CONTEXT_PROPAGATION on `LockExperimentAutomation`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:212:1`
- **Summary:** Function 'LockExperimentAutomation' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'LockExperimentAutomation' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:212:1`

### #394 CONTEXT_PROPAGATION on `UnlockExperimentAutomation`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:238:1`
- **Summary:** Function 'UnlockExperimentAutomation' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UnlockExperimentAutomation' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:238:1`

### #395 CONTEXT_PROPAGATION on `HoldExperimentForReview`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:253:1`
- **Summary:** Function 'HoldExperimentForReview' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'HoldExperimentForReview' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:253:1`

### #396 CONTEXT_PROPAGATION on `transitionExperimentStatus`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:305:1`
- **Summary:** Function 'transitionExperimentStatus' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'transitionExperimentStatus' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:305:1`

### #397 CONTEXT_PROPAGATION on `IsFeatureEnabled`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/feature_flag_service.go:52:1`
- **Summary:** Function 'IsFeatureEnabled' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'IsFeatureEnabled' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/feature_flag_service.go:52:1`

### #398 CONTEXT_PROPAGATION on `EvaluatePaywallTest`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/feature_flag_service.go:159:1`
- **Summary:** Function 'EvaluatePaywallTest' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'EvaluatePaywallTest' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/feature_flag_service.go:159:1`

### #399 CONTEXT_PROPAGATION on `GetUserSubscriptions`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_subscription_adapter.go:22:1`
- **Summary:** Function 'GetUserSubscriptions' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetUserSubscriptions' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_subscription_adapter.go:22:1`

### #400 CONTEXT_PROPAGATION on `GetTotalRevenue`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_subscription_adapter.go:44:1`
- **Summary:** Function 'GetTotalRevenue' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetTotalRevenue' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_subscription_adapter.go:44:1`

### #401 CONTEXT_PROPAGATION on `GetReport`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_report_service.go:67:1`
- **Summary:** Function 'GetReport' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetReport' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_report_service.go:67:1`

### #402 CONTEXT_PROPAGATION on `fetchMRR`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_report_service.go:123:1`
- **Summary:** Function 'fetchMRR' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'fetchMRR' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_report_service.go:123:1`

### #403 CONTEXT_PROPAGATION on `fetchLTV`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_report_service.go:138:1`
- **Summary:** Function 'fetchLTV' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'fetchLTV' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_report_service.go:138:1`

### #404 CONTEXT_PROPAGATION on `fetchNewSubsMonth`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_report_service.go:147:1`
- **Summary:** Function 'fetchNewSubsMonth' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'fetchNewSubsMonth' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_report_service.go:147:1`

### #405 CONTEXT_PROPAGATION on `fetchChurnRate`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_report_service.go:158:1`
- **Summary:** Function 'fetchChurnRate' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'fetchChurnRate' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_report_service.go:158:1`

### #406 CONTEXT_PROPAGATION on `fetchTrend`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_report_service.go:178:1`
- **Summary:** Function 'fetchTrend' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'fetchTrend' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_report_service.go:178:1`

### #407 CONTEXT_PROPAGATION on `fetchByPlatform`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_report_service.go:227:1`
- **Summary:** Function 'fetchByPlatform' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'fetchByPlatform' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_report_service.go:227:1`

### #408 CONTEXT_PROPAGATION on `fetchByPlan`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_report_service.go:253:1`
- **Summary:** Function 'fetchByPlan' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'fetchByPlan' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_report_service.go:253:1`

### #409 CONTEXT_PROPAGATION on `fetchStatusCounts`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_report_service.go:279:1`
- **Summary:** Function 'fetchStatusCounts' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'fetchStatusCounts' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_report_service.go:279:1`

### #410 CONTEXT_PROPAGATION on `GetReport`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/revenue_ops_service.go:98:1`
- **Summary:** Function 'GetReport' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetReport' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/revenue_ops_service.go:98:1`

### #411 CONTEXT_PROPAGATION on `fetchDunningData`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/revenue_ops_service.go:123:1`
- **Summary:** Function 'fetchDunningData' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'fetchDunningData' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/revenue_ops_service.go:123:1`

### #412 CONTEXT_PROPAGATION on `fetchDunningQueue`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/revenue_ops_service.go:138:1`
- **Summary:** Function 'fetchDunningQueue' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'fetchDunningQueue' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/revenue_ops_service.go:138:1`

### #413 CONTEXT_PROPAGATION on `fetchDunningStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/revenue_ops_service.go:191:1`
- **Summary:** Function 'fetchDunningStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'fetchDunningStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/revenue_ops_service.go:191:1`

### #414 CONTEXT_PROPAGATION on `fetchWebhookData`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/revenue_ops_service.go:223:1`
- **Summary:** Function 'fetchWebhookData' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'fetchWebhookData' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/revenue_ops_service.go:223:1`

### #415 CONTEXT_PROPAGATION on `fetchWebhookEvents`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/revenue_ops_service.go:261:1`
- **Summary:** Function 'fetchWebhookEvents' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'fetchWebhookEvents' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/revenue_ops_service.go:261:1`

### #416 CONTEXT_PROPAGATION on `fetchPendingWebhooks`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/revenue_ops_service.go:304:1`
- **Summary:** Function 'fetchPendingWebhooks' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'fetchPendingWebhooks' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/revenue_ops_service.go:304:1`

### #417 CONTEXT_PROPAGATION on `fetchWebhookCounts`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/revenue_ops_service.go:336:1`
- **Summary:** Function 'fetchWebhookCounts' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'fetchWebhookCounts' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/revenue_ops_service.go:336:1`

### #418 CONTEXT_PROPAGATION on `fetchWebhookProviderStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/revenue_ops_service.go:343:1`
- **Summary:** Function 'fetchWebhookProviderStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'fetchWebhookProviderStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/revenue_ops_service.go:343:1`

### #419 CONTEXT_PROPAGATION on `fetchMatomoStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/revenue_ops_service.go:369:1`
- **Summary:** Function 'fetchMatomoStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'fetchMatomoStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/revenue_ops_service.go:369:1`

### #420 CONTEXT_PROPAGATION on `GetProfile`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/user_profile_service.go:85:1`
- **Summary:** Function 'GetProfile' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetProfile' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/user_profile_service.go:85:1`

### #421 CONTEXT_PROPAGATION on `fetchUserInfo`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/user_profile_service.go:121:1`
- **Summary:** Function 'fetchUserInfo' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'fetchUserInfo' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/user_profile_service.go:121:1`

### #422 CONTEXT_PROPAGATION on `fetchSubscriptions`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/user_profile_service.go:140:1`
- **Summary:** Function 'fetchSubscriptions' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'fetchSubscriptions' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/user_profile_service.go:140:1`

### #423 CONTEXT_PROPAGATION on `fetchTransactions`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/user_profile_service.go:167:1`
- **Summary:** Function 'fetchTransactions' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'fetchTransactions' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/user_profile_service.go:167:1`

### #424 CONTEXT_PROPAGATION on `fetchAuditLog`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/user_profile_service.go:195:1`
- **Summary:** Function 'fetchAuditLog' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'fetchAuditLog' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/user_profile_service.go:195:1`

### #425 CONTEXT_PROPAGATION on `fetchDunning`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/user_profile_service.go:220:1`
- **Summary:** Function 'fetchDunning' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'fetchDunning' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/user_profile_service.go:220:1`

### #426 CONTEXT_PROPAGATION on `TrackEvent`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/matomo_forwarder.go:66:1`
- **Summary:** Function 'TrackEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'TrackEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/matomo_forwarder.go:66:1`

### #427 CONTEXT_PROPAGATION on `TrackPurchase`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/matomo_forwarder.go:100:1`
- **Summary:** Function 'TrackPurchase' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'TrackPurchase' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/matomo_forwarder.go:100:1`

### #428 CONTEXT_PROPAGATION on `ProcessBatch`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/matomo_forwarder.go:134:1`
- **Summary:** Function 'ProcessBatch' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ProcessBatch' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/matomo_forwarder.go:134:1`

### #429 CONTEXT_PROPAGATION on `processEvent`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/matomo_forwarder.go:171:1`
- **Summary:** Function 'processEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'processEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/matomo_forwarder.go:171:1`

### #430 CONTEXT_PROPAGATION on `sendEvent`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/matomo_forwarder.go:196:1`
- **Summary:** Function 'sendEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'sendEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/matomo_forwarder.go:196:1`

### #431 CONTEXT_PROPAGATION on `sendEcommerce`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/matomo_forwarder.go:240:1`
- **Summary:** Function 'sendEcommerce' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'sendEcommerce' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/matomo_forwarder.go:240:1`

### #432 CONTEXT_PROPAGATION on `HandleError`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/matomo_forwarder.go:298:1`
- **Summary:** Function 'HandleError' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'HandleError' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/matomo_forwarder.go:298:1`

### #433 CONTEXT_PROPAGATION on `GetQueueSize`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/matomo_forwarder.go:309:1`
- **Summary:** Function 'GetQueueSize' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetQueueSize' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/matomo_forwarder.go:309:1`

### #434 CONTEXT_PROPAGATION on `GetFailedEventsCount`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/matomo_forwarder.go:320:1`
- **Summary:** Function 'GetFailedEventsCount' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetFailedEventsCount' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/matomo_forwarder.go:320:1`

### #435 CONTEXT_PROPAGATION on `Evaluate`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/paywall_trigger_service.go:31:1`
- **Summary:** Function 'Evaluate' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Evaluate' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/paywall_trigger_service.go:31:1`

### #436 CONTEXT_PROPAGATION on `SelectArm`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:245:1`
- **Summary:** Function 'SelectArm' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SelectArm' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:245:1`

### #437 CONTEXT_PROPAGATION on `SelectArmWithMeta`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:360:1`
- **Summary:** Function 'SelectArmWithMeta' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SelectArmWithMeta' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:360:1`

### #438 CONTEXT_PROPAGATION on `UpdateReward`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:372:1`
- **Summary:** Function 'UpdateReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UpdateReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:372:1`

### #439 CONTEXT_PROPAGATION on `TrackImpression`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:376:1`
- **Summary:** Function 'TrackImpression' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'TrackImpression' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:376:1`

### #440 CONTEXT_PROPAGATION on `UpdateRewardWithEvent`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:438:1`
- **Summary:** Function 'UpdateRewardWithEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UpdateRewardWithEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:438:1`

### #441 CONTEXT_PROPAGATION on `GetArmStatistics`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:647:1`
- **Summary:** Function 'GetArmStatistics' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetArmStatistics' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:647:1`

### #442 CONTEXT_PROPAGATION on `CalculateWinProbability`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:668:1`
- **Summary:** Function 'CalculateWinProbability' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'CalculateWinProbability' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:668:1`

### #443 CONTEXT_PROPAGATION on `CalculateScore`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:86:1`
- **Summary:** Function 'CalculateScore' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'CalculateScore' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:86:1`

### #444 CONTEXT_PROPAGATION on `calculateConversionScore`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:105:1`
- **Summary:** Function 'calculateConversionScore' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'calculateConversionScore' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:105:1`

### #445 CONTEXT_PROPAGATION on `calculateLVTScore`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:116:1`
- **Summary:** Function 'calculateLVTScore' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'calculateLVTScore' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:116:1`

### #446 CONTEXT_PROPAGATION on `calculateRevenueScore`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:149:1`
- **Summary:** Function 'calculateRevenueScore' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'calculateRevenueScore' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:149:1`

### #447 CONTEXT_PROPAGATION on `calculateHybridScore`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:166:1`
- **Summary:** Function 'calculateHybridScore' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'calculateHybridScore' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:166:1`

### #448 CONTEXT_PROPAGATION on `RecordObjectiveReward`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:258:1`
- **Summary:** Function 'RecordObjectiveReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'RecordObjectiveReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:258:1`

### #449 CONTEXT_PROPAGATION on `GetObjectiveScores`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:315:1`
- **Summary:** Function 'GetObjectiveScores' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetObjectiveScores' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:315:1`

### #450 CONTEXT_PROPAGATION on `ExecuteScheduled`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/automation_job_execution_service.go:58:1`
- **Summary:** Function 'ExecuteScheduled' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ExecuteScheduled' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/automation_job_execution_service.go:58:1`

### #451 CONTEXT_PROPAGATION on `RecordPendingReward`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:80:1`
- **Summary:** Function 'RecordPendingReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'RecordPendingReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:80:1`

### #452 CONTEXT_PROPAGATION on `ProcessConversion`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:124:1`
- **Summary:** Function 'ProcessConversion' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ProcessConversion' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:124:1`

### #453 CONTEXT_PROPAGATION on `ProcessExpiredRewards`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:253:1`
- **Summary:** Function 'ProcessExpiredRewards' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ProcessExpiredRewards' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:253:1`

### #454 CONTEXT_PROPAGATION on `GetPendingReward`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:346:1`
- **Summary:** Function 'GetPendingReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetPendingReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:346:1`

### #455 CONTEXT_PROPAGATION on `GetPendingRewardsByUser`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:372:1`
- **Summary:** Function 'GetPendingRewardsByUser' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetPendingRewardsByUser' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:372:1`

### #456 CONTEXT_PROPAGATION on `GetStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:385:1`
- **Summary:** Function 'GetStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:385:1`

### #457 CONTEXT_PROPAGATION on `cachePendingReward`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:429:1`
- **Summary:** Function 'cachePendingReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'cachePendingReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:429:1`

### #458 CONTEXT_PROPAGATION on `getCachedPendingReward`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:437:1`
- **Summary:** Function 'getCachedPendingReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'getCachedPendingReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:437:1`

### #459 CONTEXT_PROPAGATION on `invalidatePendingCache`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:449:1`
- **Summary:** Function 'invalidatePendingCache' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'invalidatePendingCache' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:449:1`

### #460 CONTEXT_PROPAGATION on `GetConversionLinks`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:454:1`
- **Summary:** Function 'GetConversionLinks' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetConversionLinks' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:454:1`

### #461 CONTEXT_PROPAGATION on `CreateWinbackOffer`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/winback_service.go:42:1`
- **Summary:** Function 'CreateWinbackOffer' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'CreateWinbackOffer' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/winback_service.go:42:1`

### #462 CONTEXT_PROPAGATION on `AcceptWinbackOffer`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/winback_service.go:76:1`
- **Summary:** Function 'AcceptWinbackOffer' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'AcceptWinbackOffer' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/winback_service.go:76:1`

### #463 CONTEXT_PROPAGATION on `GetActiveWinbackOffers`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/winback_service.go:104:1`
- **Summary:** Function 'GetActiveWinbackOffers' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetActiveWinbackOffers' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/winback_service.go:104:1`

### #464 CONTEXT_PROPAGATION on `ProcessExpiredWinbackOffers`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/winback_service.go:109:1`
- **Summary:** Function 'ProcessExpiredWinbackOffers' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ProcessExpiredWinbackOffers' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/winback_service.go:109:1`

### #465 CONTEXT_PROPAGATION on `CreateWinbackCampaignForChurnedUsers`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/winback_service.go:133:1`
- **Summary:** Function 'CreateWinbackCampaignForChurnedUsers' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'CreateWinbackCampaignForChurnedUsers' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/winback_service.go:133:1`

### #466 CONTEXT_PROPAGATION on `DeactivateCampaign`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/winback_service.go:161:1`
- **Summary:** Function 'DeactivateCampaign' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'DeactivateCampaign' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/winback_service.go:161:1`

### #467 CONTEXT_PROPAGATION on `CalculateReward`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/currency_conversion_strategy.go:41:1`
- **Summary:** Function 'CalculateReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'CalculateReward' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/currency_conversion_strategy.go:41:1`

### #468 CONTEXT_PROPAGATION on `RecordRewardWithCurrency`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/currency_conversion_strategy.go:157:1`
- **Summary:** Function 'RecordRewardWithCurrency' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'RecordRewardWithCurrency' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/currency_conversion_strategy.go:157:1`

### #469 CONTEXT_PROPAGATION on `GetConversionRate`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/currency_conversion_strategy.go:205:1`
- **Summary:** Function 'GetConversionRate' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetConversionRate' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/currency_conversion_strategy.go:205:1`

### #470 CONTEXT_PROPAGATION on `EstimateRevenueUSD`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/currency_conversion_strategy.go:218:1`
- **Summary:** Function 'EstimateRevenueUSD' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'EstimateRevenueUSD' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/currency_conversion_strategy.go:218:1`

### #471 CONTEXT_PROPAGATION on `CalculateLTV`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:91:1`
- **Summary:** Function 'CalculateLTV' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'CalculateLTV' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:91:1`

### #472 CONTEXT_PROPAGATION on `predictLTVFromCohorts`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:178:1`
- **Summary:** Function 'predictLTVFromCohorts' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'predictLTVFromCohorts' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:178:1`

### #473 CONTEXT_PROPAGATION on `GetCohortLTV`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:313:1`
- **Summary:** Function 'GetCohortLTV' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetCohortLTV' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:313:1`

### #474 CONTEXT_PROPAGATION on `UpdateUserLTV`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:370:1`
- **Summary:** Function 'UpdateUserLTV' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UpdateUserLTV' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:370:1`

### #475 CONTEXT_PROPAGATION on `GetSegmentedLTV`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:388:1`
- **Summary:** Function 'GetSegmentedLTV' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetSegmentedLTV' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:388:1`

### #476 CONTEXT_PROPAGATION on `PredictChurnRisk`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:405:1`
- **Summary:** Function 'PredictChurnRisk' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'PredictChurnRisk' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:405:1`

### #477 CONTEXT_PROPAGATION on `sendEmail`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/notification_service.go:44:1`
- **Summary:** Function 'sendEmail' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'sendEmail' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/notification_service.go:44:1`

### #478 CONTEXT_PROPAGATION on `SendGracePeriodExpiringNotification`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/notification_service.go:121:1`
- **Summary:** Function 'SendGracePeriodExpiringNotification' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SendGracePeriodExpiringNotification' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/notification_service.go:121:1`

### #479 CONTEXT_PROPAGATION on `SendWinbackOfferNotification`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/notification_service.go:137:1`
- **Summary:** Function 'SendWinbackOfferNotification' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SendWinbackOfferNotification' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/notification_service.go:137:1`

### #480 CONTEXT_PROPAGATION on `SendSubscriptionExpiredNotification`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/notification_service.go:149:1`
- **Summary:** Function 'SendSubscriptionExpiredNotification' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SendSubscriptionExpiredNotification' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/notification_service.go:149:1`

### #481 CONTEXT_PROPAGATION on `SendPaymentRetryNotification`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/notification_service.go:160:1`
- **Summary:** Function 'SendPaymentRetryNotification' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SendPaymentRetryNotification' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/notification_service.go:160:1`

### #482 CONTEXT_PROPAGATION on `SendPaymentSuccessNotification`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/notification_service.go:171:1`
- **Summary:** Function 'SendPaymentSuccessNotification' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SendPaymentSuccessNotification' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/notification_service.go:171:1`

### #483 CONTEXT_PROPAGATION on `SendAllRetriesFailedNotification`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/notification_service.go:181:1`
- **Summary:** Function 'SendAllRetriesFailedNotification' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SendAllRetriesFailedNotification' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/notification_service.go:181:1`

### #484 CONTEXT_PROPAGATION on `SendPaymentFinalFailureNotification`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/notification_service.go:191:1`
- **Summary:** Function 'SendPaymentFinalFailureNotification' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SendPaymentFinalFailureNotification' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/notification_service.go:191:1`

### #485 CONTEXT_PROPAGATION on `SelectArm`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/linucb_strategy.go:58:1`
- **Summary:** Function 'SelectArm' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SelectArm' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/linucb_strategy.go:58:1`

### #486 CONTEXT_PROPAGATION on `UpdateModel`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/linucb_strategy.go:117:1`
- **Summary:** Function 'UpdateModel' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'UpdateModel' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/linucb_strategy.go:117:1`

### #487 CONTEXT_PROPAGATION on `getOrCreateModel`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/linucb_strategy.go:188:1`
- **Summary:** Function 'getOrCreateModel' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'getOrCreateModel' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/linucb_strategy.go:188:1`

### #488 CONTEXT_PROPAGATION on `saveModel`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/linucb_strategy.go:217:1`
- **Summary:** Function 'saveModel' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'saveModel' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/linucb_strategy.go:217:1`

### #489 CONTEXT_PROPAGATION on `GetModelStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/linucb_strategy.go:351:1`
- **Summary:** Function 'GetModelStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetModelStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/linucb_strategy.go:351:1`

### #490 CONTEXT_PROPAGATION on `LogAction`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/audit_service.go:23:1`
- **Summary:** Function 'LogAction' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'LogAction' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/audit_service.go:23:1`

### #491 CONTEXT_PROPAGATION on `Recommend`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:124:1`
- **Summary:** Function 'Recommend' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Recommend' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:124:1`

### #492 CONTEXT_PROPAGATION on `finalizeRecommendation`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:192:1`
- **Summary:** Function 'finalizeRecommendation' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'finalizeRecommendation' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:192:1`

### #493 CONTEXT_PROPAGATION on `GetArmStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:83:1`
- **Summary:** Function 'GetArmStats' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetArmStats' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:83:1`

### #494 CONTEXT_PROPAGATION on `SetArmStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:87:1`
- **Summary:** Function 'SetArmStats' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SetArmStats' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:87:1`

### #495 CONTEXT_PROPAGATION on `GetAssignment`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:91:1`
- **Summary:** Function 'GetAssignment' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetAssignment' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:91:1`

### #496 CONTEXT_PROPAGATION on `SetAssignment`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:95:1`
- **Summary:** Function 'SetAssignment' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SetAssignment' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:95:1`

### #497 CONTEXT_PROPAGATION on `SetBytes`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:99:1`
- **Summary:** Function 'SetBytes' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SetBytes' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:99:1`

### #498 CONTEXT_PROPAGATION on `GetBytes`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:102:1`
- **Summary:** Function 'GetBytes' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetBytes' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:102:1`

### #499 CONTEXT_PROPAGATION on `DeleteKey`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:105:1`
- **Summary:** Function 'DeleteKey' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'DeleteKey' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:105:1`

### #500 CONTEXT_PROPAGATION on `CalculateRevenueMetrics`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_service.go:42:1`
- **Summary:** Function 'CalculateRevenueMetrics' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'CalculateRevenueMetrics' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_service.go:42:1`

### #501 CONTEXT_PROPAGATION on `CalculateChurnMetrics`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_service.go:63:1`
- **Summary:** Function 'CalculateChurnMetrics' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'CalculateChurnMetrics' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_service.go:63:1`

### #502 CONTEXT_PROPAGATION on `GetMRRTrend`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_service.go:90:1`
- **Summary:** Function 'GetMRRTrend' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetMRRTrend' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_service.go:90:1`

### #503 CONTEXT_PROPAGATION on `GetSubscriptionStatusCounts`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_service.go:95:1`
- **Summary:** Function 'GetSubscriptionStatusCounts' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetSubscriptionStatusCounts' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_service.go:95:1`

### #504 CONTEXT_PROPAGATION on `GetChurnRiskCount`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_service.go:100:1`
- **Summary:** Function 'GetChurnRiskCount' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetChurnRiskCount' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_service.go:100:1`

### #505 CONTEXT_PROPAGATION on `GetWebhookHealthByProvider`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_service.go:105:1`
- **Summary:** Function 'GetWebhookHealthByProvider' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetWebhookHealthByProvider' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_service.go:105:1`

### #506 CONTEXT_PROPAGATION on `GetRecentAuditLog`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_service.go:110:1`
- **Summary:** Function 'GetRecentAuditLog' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetRecentAuditLog' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_service.go:110:1`

### #507 CONTEXT_PROPAGATION on `GetAuditLogPaginated`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_service.go:116:1`
- **Summary:** Function 'GetAuditLogPaginated' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetAuditLogPaginated' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_service.go:116:1`

### #508 CONTEXT_PROPAGATION on `StartDunning`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/dunning_service.go:39:1`
- **Summary:** Function 'StartDunning' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'StartDunning' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/dunning_service.go:39:1`

### #509 CONTEXT_PROPAGATION on `ProcessDunningAttempt`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/dunning_service.go:63:1`
- **Summary:** Function 'ProcessDunningAttempt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ProcessDunningAttempt' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/dunning_service.go:63:1`

### #510 CONTEXT_PROPAGATION on `GetPendingDunningAttempts`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/dunning_service.go:129:1`
- **Summary:** Function 'GetPendingDunningAttempts' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetPendingDunningAttempts' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/dunning_service.go:129:1`

### #511 CONTEXT_PROPAGATION on `Reconcile`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_reconciler.go:38:1`
- **Summary:** Function 'Reconcile' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Reconcile' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_reconciler.go:38:1`

### #512 CONTEXT_PROPAGATION on `Reconcile`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_automation_service.go:51:1`
- **Summary:** Function 'Reconcile' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'Reconcile' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_automation_service.go:51:1`

### #513 CONTEXT_PROPAGATION on `TrackExposure`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ab_analytics_service.go:36:1`
- **Summary:** Function 'TrackExposure' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'TrackExposure' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ab_analytics_service.go:36:1`

### #514 CONTEXT_PROPAGATION on `TrackConversion`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ab_analytics_service.go:51:1`
- **Summary:** Function 'TrackConversion' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'TrackConversion' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ab_analytics_service.go:51:1`

### #515 CONTEXT_PROPAGATION on `TrackRevenue`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ab_analytics_service.go:66:1`
- **Summary:** Function 'TrackRevenue' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'TrackRevenue' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ab_analytics_service.go:66:1`

### #516 CONTEXT_PROPAGATION on `GetArmStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:63:1`
- **Summary:** Function 'GetArmStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetArmStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:63:1`

### #517 CONTEXT_PROPAGATION on `RecordEvent`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:93:1`
- **Summary:** Function 'RecordEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'RecordEvent' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:93:1`

### #518 CONTEXT_PROPAGATION on `calculateWindowStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:142:1`
- **Summary:** Function 'calculateWindowStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'calculateWindowStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:142:1`

### #519 CONTEXT_PROPAGATION on `cacheStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:221:1`
- **Summary:** Function 'cacheStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'cacheStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:221:1`

### #520 CONTEXT_PROPAGATION on `GetWindowInfo`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:260:1`
- **Summary:** Function 'GetWindowInfo' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetWindowInfo' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:260:1`

### #521 CONTEXT_PROPAGATION on `TrimWindow`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:314:1`
- **Summary:** Function 'TrimWindow' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'TrimWindow' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:314:1`

### #522 CONTEXT_PROPAGATION on `ClearWindow`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:331:1`
- **Summary:** Function 'ClearWindow' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ClearWindow' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:331:1`

### #523 CONTEXT_PROPAGATION on `HasEnoughSamples`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:348:1`
- **Summary:** Function 'HasEnoughSamples' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'HasEnoughSamples' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:348:1`

### #524 CONTEXT_PROPAGATION on `GetUtilization`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:364:1`
- **Summary:** Function 'GetUtilization' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetUtilization' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:364:1`

### #525 CONTEXT_PROPAGATION on `ExportEvents`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:384:1`
- **Summary:** Function 'ExportEvents' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'ExportEvents' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:384:1`

### #526 CONTEXT_PROPAGATION on `GetArmStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:45:1`
- **Summary:** Function 'GetArmStats' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetArmStats' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:45:1`

### #527 CONTEXT_PROPAGATION on `SetArmStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:48:1`
- **Summary:** Function 'SetArmStats' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SetArmStats' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:48:1`

### #528 CONTEXT_PROPAGATION on `GetAssignment`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:51:1`
- **Summary:** Function 'GetAssignment' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetAssignment' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:51:1`

### #529 CONTEXT_PROPAGATION on `SetAssignment`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:54:1`
- **Summary:** Function 'SetAssignment' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SetAssignment' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:54:1`

### #530 CONTEXT_PROPAGATION on `SetBytes`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:57:1`
- **Summary:** Function 'SetBytes' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'SetBytes' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:57:1`

### #531 CONTEXT_PROPAGATION on `GetBytes`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:58:1`
- **Summary:** Function 'GetBytes' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'GetBytes' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:58:1`

### #532 CONTEXT_PROPAGATION on `DeleteKey`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:61:1`
- **Summary:** Function 'DeleteKey' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'DeleteKey' adheres to idiomatic Go context propagation passing ': context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:61:1`

### #533 CONTEXT_PROPAGATION on `RepairExperiment`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:73:1`
- **Summary:** Function 'RepairExperiment' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'RepairExperiment' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:73:1`

### #534 CONTEXT_PROPAGATION on `syncObjectiveStats`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:124:1`
- **Summary:** Function 'syncObjectiveStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'syncObjectiveStats' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:124:1`

### #535 CONTEXT_PROPAGATION on `recalculateWinnerConfidence`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:166:1`
- **Summary:** Function 'recalculateWinnerConfidence' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'recalculateWinnerConfidence' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:166:1`

### #536 CONTEXT_PROPAGATION on `validatePricingTiersExist`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_experiments.go:989:1`
- **Summary:** Function 'validatePricingTiersExist' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'validatePricingTiersExist' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_experiments.go:989:1`

### #537 CONTEXT_PROPAGATION on `applyExperimentArmPricingTierUpdates`
- **Category:** `idiom`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_experiments.go:1004:1`
- **Summary:** Function 'applyExperimentArmPricingTierUpdates' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter

#### Evidence Trail:
- `+70%` **[CONTEXT_FIRST_PARAM]** Function 'applyExperimentArmPricingTierUpdates' adheres to idiomatic Go context propagation passing 'ctx: context.Context' as first parameter -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_experiments.go:1004:1`

### #538 STRUCT_EMBEDDING on `JWTClaims`
- **Category:** `idiom`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/middleware/jwt.go:21:1`
- **Summary:** Struct 'JWTClaims' implements idiomatic Go Composition Over Inheritance via anonymous embedding of (jwt.RegisteredClaims, jwt.RegisteredClaims)

#### Evidence Trail:
- `+75%` **[STRUCT_EMBEDDING_COMPOSITION]** Struct 'JWTClaims' implements idiomatic Go Composition Over Inheritance via anonymous embedding of (jwt.RegisteredClaims, jwt.RegisteredClaims) -> `/Volumes/External/Code/paywall-iap/backend/internal/application/middleware/jwt.go:21:1`

### #539 SINGLE_RESPONSIBILITY on `AnalyticsCache`
- **Category:** `principle`
- **Confidence:** **85%** [VERY_HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:15:1`
- **Summary:** SRP Violation (God Struct): Struct 'AnalyticsCache' has 15 methods and 2 fields, indicating mixed domain responsibilities

#### Evidence Trail:
- `+85%` **[SRP_GOD_STRUCT]** SRP Violation (God Struct): Struct 'AnalyticsCache' has 15 methods and 2 fields, indicating mixed domain responsibilities -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:15:1`

### #540 SINGLE_RESPONSIBILITY on `PostgresBanditRepository`
- **Category:** `principle`
- **Confidence:** **85%** [VERY_HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:23:1`
- **Summary:** SRP Violation (God Struct): Struct 'PostgresBanditRepository' has 37 methods and 2 fields, indicating mixed domain responsibilities

#### Evidence Trail:
- `+85%` **[SRP_GOD_STRUCT]** SRP Violation (God Struct): Struct 'PostgresBanditRepository' has 37 methods and 2 fields, indicating mixed domain responsibilities -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:23:1`

### #541 SINGLE_RESPONSIBILITY on `ExperimentAdminRepository`
- **Category:** `principle`
- **Confidence:** **85%** [VERY_HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:19:1`
- **Summary:** SRP Violation (God Struct): Struct 'ExperimentAdminRepository' has 15 methods and 1 fields, indicating mixed domain responsibilities

#### Evidence Trail:
- `+85%` **[SRP_GOD_STRUCT]** SRP Violation (God Struct): Struct 'ExperimentAdminRepository' has 15 methods and 1 fields, indicating mixed domain responsibilities -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:19:1`

### #542 SINGLE_RESPONSIBILITY on `AdvancedBanditEngine`
- **Category:** `principle`
- **Confidence:** **85%** [VERY_HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:16:1`
- **Summary:** SRP Violation (God Struct): Struct 'AdvancedBanditEngine' has 25 methods and 16 fields, indicating mixed domain responsibilities

#### Evidence Trail:
- `+85%` **[SRP_GOD_STRUCT]** SRP Violation (God Struct): Struct 'AdvancedBanditEngine' has 25 methods and 16 fields, indicating mixed domain responsibilities -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:16:1`

### #543 SINGLE_RESPONSIBILITY on `SlidingWindowStrategy`
- **Category:** `principle`
- **Confidence:** **85%** [VERY_HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:16:1`
- **Summary:** SRP Violation (God Struct): Struct 'SlidingWindowStrategy' has 17 methods and 5 fields, indicating mixed domain responsibilities

#### Evidence Trail:
- `+85%` **[SRP_GOD_STRUCT]** SRP Violation (God Struct): Struct 'SlidingWindowStrategy' has 17 methods and 5 fields, indicating mixed domain responsibilities -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:16:1`

### #544 SINGLE_RESPONSIBILITY on `AdminHandler`
- **Category:** `principle`
- **Confidence:** **85%** [VERY_HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin.go:31:1`
- **Summary:** SRP Violation (God Struct): Struct 'AdminHandler' has 19 methods and 15 fields, indicating mixed domain responsibilities

#### Evidence Trail:
- `+85%` **[SRP_GOD_STRUCT]** SRP Violation (God Struct): Struct 'AdminHandler' has 19 methods and 15 fields, indicating mixed domain responsibilities -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin.go:31:1`

### #545 SINGLE_RESPONSIBILITY on `BanditAdvancedHandler`
- **Category:** `principle`
- **Confidence:** **85%** [VERY_HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/bandit_advanced.go:23:1`
- **Summary:** SRP Violation (God Struct): Struct 'BanditAdvancedHandler' has 15 methods and 3 fields, indicating mixed domain responsibilities

#### Evidence Trail:
- `+85%` **[SRP_GOD_STRUCT]** SRP Violation (God Struct): Struct 'BanditAdvancedHandler' has 15 methods and 3 fields, indicating mixed domain responsibilities -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/bandit_advanced.go:23:1`

### #546 INTERFACE_POLLUTION on `Querier`
- **Category:** `principle`
- **Confidence:** **80%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/sqlc/generated/querier.go:13:1`
- **Summary:** Interface Pollution (ISP Violation): Interface 'Querier' declares 42 methods; idiomatic Go encourages small, single-purpose interfaces (1-2 methods)

#### Evidence Trail:
- `+80%` **[INTERFACE_POLLUTION_FAT_INTERFACE]** Interface Pollution (ISP Violation): Interface 'Querier' declares 42 methods; idiomatic Go encourages small, single-purpose interfaces (1-2 methods) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/sqlc/generated/querier.go:13:1`

### #547 INTERFACE_POLLUTION on `AppRepository`
- **Category:** `principle`
- **Confidence:** **80%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/repository/app_repository.go:11:1`
- **Summary:** Interface Pollution (ISP Violation): Interface 'AppRepository' declares 12 methods; idiomatic Go encourages small, single-purpose interfaces (1-2 methods)

#### Evidence Trail:
- `+80%` **[INTERFACE_POLLUTION_FAT_INTERFACE]** Interface Pollution (ISP Violation): Interface 'AppRepository' declares 12 methods; idiomatic Go encourages small, single-purpose interfaces (1-2 methods) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/repository/app_repository.go:11:1`

### #548 INTERFACE_POLLUTION on `UserRepository`
- **Category:** `principle`
- **Confidence:** **80%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/repository/user_repository.go:11:1`
- **Summary:** Interface Pollution (ISP Violation): Interface 'UserRepository' declares 13 methods; idiomatic Go encourages small, single-purpose interfaces (1-2 methods)

#### Evidence Trail:
- `+80%` **[INTERFACE_POLLUTION_FAT_INTERFACE]** Interface Pollution (ISP Violation): Interface 'UserRepository' declares 13 methods; idiomatic Go encourages small, single-purpose interfaces (1-2 methods) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/repository/user_repository.go:11:1`

### #549 INTERFACE_POLLUTION on `SubscriptionRepository`
- **Category:** `principle`
- **Confidence:** **80%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/repository/subscription_repository.go:11:1`
- **Summary:** Interface Pollution (ISP Violation): Interface 'SubscriptionRepository' declares 11 methods; idiomatic Go encourages small, single-purpose interfaces (1-2 methods)

#### Evidence Trail:
- `+80%` **[INTERFACE_POLLUTION_FAT_INTERFACE]** Interface Pollution (ISP Violation): Interface 'SubscriptionRepository' declares 11 methods; idiomatic Go encourages small, single-purpose interfaces (1-2 methods) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/repository/subscription_repository.go:11:1`

### #550 INTERFACE_POLLUTION on `AnalyticsRepository`
- **Category:** `principle`
- **Confidence:** **80%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/repository/analytics_repository.go:54:1`
- **Summary:** Interface Pollution (ISP Violation): Interface 'AnalyticsRepository' declares 10 methods; idiomatic Go encourages small, single-purpose interfaces (1-2 methods)

#### Evidence Trail:
- `+80%` **[INTERFACE_POLLUTION_FAT_INTERFACE]** Interface Pollution (ISP Violation): Interface 'AnalyticsRepository' declares 10 methods; idiomatic Go encourages small, single-purpose interfaces (1-2 methods) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/repository/analytics_repository.go:54:1`

### #551 INTERFACE_POLLUTION on `BanditRepository`
- **Category:** `principle`
- **Confidence:** **80%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:25:1`
- **Summary:** Interface Pollution (ISP Violation): Interface 'BanditRepository' declares 9 methods; idiomatic Go encourages small, single-purpose interfaces (1-2 methods)

#### Evidence Trail:
- `+80%` **[INTERFACE_POLLUTION_FAT_INTERFACE]** Interface Pollution (ISP Violation): Interface 'BanditRepository' declares 9 methods; idiomatic Go encourages small, single-purpose interfaces (1-2 methods) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:25:1`

### #552 OPEN_CLOSED on `HandleSendNotification`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/tasks.go:328:1`
- **Summary:** OCP Violation: Function 'HandleSendNotification' uses type switch with 18 branches; consider interface method polymorphism for open extensibility

#### Evidence Trail:
- `+75%` **[OCP_TYPE_SWITCH_CASCADE]** OCP Violation: Function 'HandleSendNotification' uses type switch with 18 branches; consider interface method polymorphism for open extensibility -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/tasks.go:328:1`

### #553 DEPENDENCY_INVERSION on `NewVerifyIAPCommand`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/verify_iap.go:58:1`
- **Summary:** DIP Adherence: Function 'NewVerifyIAPCommand' depends on interface abstraction(s) (iosVerifier DynamicIAPVerifier, androidVerifier DynamicIAPVerifier) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'NewVerifyIAPCommand' depends on interface abstraction(s) (iosVerifier DynamicIAPVerifier, androidVerifier DynamicIAPVerifier) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/verify_iap.go:58:1`

### #554 DEPENDENCY_INVERSION on `NewVerifyIAPCommandLegacy`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/verify_iap.go:75:1`
- **Summary:** DIP Adherence: Function 'NewVerifyIAPCommandLegacy' depends on interface abstraction(s) (iosVerifier IAPVerifier, androidVerifier IAPVerifier) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'NewVerifyIAPCommandLegacy' depends on interface abstraction(s) (iosVerifier IAPVerifier, androidVerifier IAPVerifier) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/verify_iap.go:75:1`

### #555 DEPENDENCY_INVERSION on `TrackEvent`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:71:1`
- **Summary:** DIP Adherence: Function 'TrackEvent' depends on interface abstraction(s) (ctx context.Context, req TrackEventRequest) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'TrackEvent' depends on interface abstraction(s) (ctx context.Context, req TrackEventRequest) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:71:1`

### #556 DEPENDENCY_INVERSION on `TrackEcommerce`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:129:1`
- **Summary:** DIP Adherence: Function 'TrackEcommerce' depends on interface abstraction(s) (ctx context.Context, req TrackEcommerceRequest) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'TrackEcommerce' depends on interface abstraction(s) (ctx context.Context, req TrackEcommerceRequest) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:129:1`

### #557 DEPENDENCY_INVERSION on `GetCohorts`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:198:1`
- **Summary:** DIP Adherence: Function 'GetCohorts' depends on interface abstraction(s) (ctx context.Context, req CohortRequest) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'GetCohorts' depends on interface abstraction(s) (ctx context.Context, req CohortRequest) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:198:1`

### #558 DEPENDENCY_INVERSION on `GetFunnels`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:247:1`
- **Summary:** DIP Adherence: Function 'GetFunnels' depends on interface abstraction(s) (ctx context.Context, req FunnelRequest) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'GetFunnels' depends on interface abstraction(s) (ctx context.Context, req FunnelRequest) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/matomo/client.go:247:1`

### #559 DEPENDENCY_INVERSION on `NewSubscription`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/subscription.go:50:1`
- **Summary:** DIP Adherence: Function 'NewSubscription' depends on interface abstraction(s) (source SubscriptionSource, planType PlanType) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'NewSubscription' depends on interface abstraction(s) (source SubscriptionSource, planType PlanType) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/subscription.go:50:1`

### #560 DEPENDENCY_INVERSION on `NewAdvancedBanditEngine`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:72:1`
- **Summary:** DIP Adherence: Function 'NewAdvancedBanditEngine' depends on interface abstraction(s) (repo BanditRepository, cache BanditCache) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'NewAdvancedBanditEngine' depends on interface abstraction(s) (repo BanditRepository, cache BanditCache) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:72:1`

### #561 DEPENDENCY_INVERSION on `SelectArm`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:209:1`
- **Summary:** DIP Adherence: Function 'SelectArm' depends on interface abstraction(s) (ctx context.Context, userContext UserContext) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'SelectArm' depends on interface abstraction(s) (ctx context.Context, userContext UserContext) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:209:1`

### #562 DEPENDENCY_INVERSION on `RecordReward`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:271:1`
- **Summary:** DIP Adherence: Function 'RecordReward' depends on interface abstraction(s) (ctx context.Context, userContext UserContext) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'RecordReward' depends on interface abstraction(s) (ctx context.Context, userContext UserContext) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:271:1`

### #563 DEPENDENCY_INVERSION on `SetObjectiveConfig`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:428:1`
- **Summary:** DIP Adherence: Function 'SetObjectiveConfig' depends on interface abstraction(s) (ctx context.Context, objectiveType ObjectiveType) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'SetObjectiveConfig' depends on interface abstraction(s) (ctx context.Context, objectiveType ObjectiveType) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:428:1`

### #564 DEPENDENCY_INVERSION on `UpdateDraftExperiment`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:172:1`
- **Summary:** DIP Adherence: Function 'UpdateDraftExperiment' depends on interface abstraction(s) (ctx context.Context, input UpdateExperimentInput) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'UpdateDraftExperiment' depends on interface abstraction(s) (ctx context.Context, input UpdateExperimentInput) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:172:1`

### #565 DEPENDENCY_INVERSION on `UpdateExperimentAutomationPolicy`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:183:1`
- **Summary:** DIP Adherence: Function 'UpdateExperimentAutomationPolicy' depends on interface abstraction(s) (ctx context.Context, input UpdateExperimentAutomationPolicyInput) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'UpdateExperimentAutomationPolicy' depends on interface abstraction(s) (ctx context.Context, input UpdateExperimentAutomationPolicyInput) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:183:1`

### #566 DEPENDENCY_INVERSION on `LockExperimentAutomation`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:212:1`
- **Summary:** DIP Adherence: Function 'LockExperimentAutomation' depends on interface abstraction(s) (ctx context.Context, input ExperimentLockInput) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'LockExperimentAutomation' depends on interface abstraction(s) (ctx context.Context, input ExperimentLockInput) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:212:1`

### #567 DEPENDENCY_INVERSION on `HoldExperimentForReview`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:253:1`
- **Summary:** DIP Adherence: Function 'HoldExperimentForReview' depends on interface abstraction(s) (ctx context.Context, input ExperimentLockInput) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'HoldExperimentForReview' depends on interface abstraction(s) (ctx context.Context, input ExperimentLockInput) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_admin_service.go:253:1`

### #568 DEPENDENCY_INVERSION on `NewThompsonSamplingBandit`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:229:1`
- **Summary:** DIP Adherence: Function 'NewThompsonSamplingBandit' depends on interface abstraction(s) (repo BanditRepository, cache BanditCache) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'NewThompsonSamplingBandit' depends on interface abstraction(s) (repo BanditRepository, cache BanditCache) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:229:1`

### #569 DEPENDENCY_INVERSION on `NewHybridObjectiveStrategy`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:54:1`
- **Summary:** DIP Adherence: Function 'NewHybridObjectiveStrategy' depends on interface abstraction(s) (repo BanditRepository, cache BanditCache) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'NewHybridObjectiveStrategy' depends on interface abstraction(s) (repo BanditRepository, cache BanditCache) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:54:1`

### #570 DEPENDENCY_INVERSION on `RecordObjectiveReward`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:258:1`
- **Summary:** DIP Adherence: Function 'RecordObjectiveReward' depends on interface abstraction(s) (ctx context.Context, objectiveType ObjectiveType) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'RecordObjectiveReward' depends on interface abstraction(s) (ctx context.Context, objectiveType ObjectiveType) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:258:1`

### #571 DEPENDENCY_INVERSION on `ExecuteScheduled`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/automation_job_execution_service.go:58:1`
- **Summary:** DIP Adherence: Function 'ExecuteScheduled' depends on interface abstraction(s) (ctx context.Context, spec ScheduledAutomationJobSpec) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'ExecuteScheduled' depends on interface abstraction(s) (ctx context.Context, spec ScheduledAutomationJobSpec) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/automation_job_execution_service.go:58:1`

### #572 DEPENDENCY_INVERSION on `NewDelayedRewardStrategy`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:65:1`
- **Summary:** DIP Adherence: Function 'NewDelayedRewardStrategy' depends on interface abstraction(s) (repo BanditRepository, cache BanditCache) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'NewDelayedRewardStrategy' depends on interface abstraction(s) (repo BanditRepository, cache BanditCache) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:65:1`

### #573 DEPENDENCY_INVERSION on `CalculateReward`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/currency_conversion_strategy.go:41:1`
- **Summary:** DIP Adherence: Function 'CalculateReward' depends on interface abstraction(s) (ctx context.Context, arm Arm) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'CalculateReward' depends on interface abstraction(s) (ctx context.Context, arm Arm) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/currency_conversion_strategy.go:41:1`

### #574 DEPENDENCY_INVERSION on `NewLTVService`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:55:1`
- **Summary:** DIP Adherence: Function 'NewLTVService' depends on interface abstraction(s) (cohortWorker CohortWorker, subscriptionRepo SubscriptionRepository) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'NewLTVService' depends on interface abstraction(s) (cohortWorker CohortWorker, subscriptionRepo SubscriptionRepository) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:55:1`

### #575 DEPENDENCY_INVERSION on `NewLinUCBSelectionStrategy`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/linucb_strategy.go:33:1`
- **Summary:** DIP Adherence: Function 'NewLinUCBSelectionStrategy' depends on interface abstraction(s) (repo BanditRepository, cache BanditCache) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'NewLinUCBSelectionStrategy' depends on interface abstraction(s) (repo BanditRepository, cache BanditCache) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/linucb_strategy.go:33:1`

### #576 DEPENDENCY_INVERSION on `SelectArm`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/linucb_strategy.go:58:1`
- **Summary:** DIP Adherence: Function 'SelectArm' depends on interface abstraction(s) (ctx context.Context, userContext UserContext) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'SelectArm' depends on interface abstraction(s) (ctx context.Context, userContext UserContext) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/linucb_strategy.go:58:1`

### #577 DEPENDENCY_INVERSION on `UpdateModel`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/linucb_strategy.go:117:1`
- **Summary:** DIP Adherence: Function 'UpdateModel' depends on interface abstraction(s) (ctx context.Context, userContext UserContext) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'UpdateModel' depends on interface abstraction(s) (ctx context.Context, userContext UserContext) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/linucb_strategy.go:117:1`

### #578 DEPENDENCY_INVERSION on `Recommend`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:124:1`
- **Summary:** DIP Adherence: Function 'Recommend' depends on interface abstraction(s) (ctx context.Context, input ExperimentWinnerRecommendationInput) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'Recommend' depends on interface abstraction(s) (ctx context.Context, input ExperimentWinnerRecommendationInput) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:124:1`

### #579 DEPENDENCY_INVERSION on `finalizeRecommendation`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:192:1`
- **Summary:** DIP Adherence: Function 'finalizeRecommendation' depends on interface abstraction(s) (ctx context.Context, input ExperimentWinnerRecommendationInput) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'finalizeRecommendation' depends on interface abstraction(s) (ctx context.Context, input ExperimentWinnerRecommendationInput) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:192:1`

### #580 DEPENDENCY_INVERSION on `NewExperimentRepairReconciler`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_reconciler.go:34:1`
- **Summary:** DIP Adherence: Function 'NewExperimentRepairReconciler' depends on interface abstraction(s) (candidates ExperimentRepairCandidateRepository, repairer ExperimentRepairExecutor) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'NewExperimentRepairReconciler' depends on interface abstraction(s) (candidates ExperimentRepairCandidateRepository, repairer ExperimentRepairExecutor) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_reconciler.go:34:1`

### #581 DEPENDENCY_INVERSION on `NewExperimentAutomationReconciler`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_automation_service.go:43:1`
- **Summary:** DIP Adherence: Function 'NewExperimentAutomationReconciler' depends on interface abstraction(s) (repo ExperimentAutomationRepository, transitions ExperimentStatusTransitioner) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'NewExperimentAutomationReconciler' depends on interface abstraction(s) (repo ExperimentAutomationRepository, transitions ExperimentStatusTransitioner) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_automation_service.go:43:1`

### #582 DEPENDENCY_INVERSION on `RecordEvent`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:93:1`
- **Summary:** DIP Adherence: Function 'RecordEvent' depends on interface abstraction(s) (ctx context.Context, event RewardEvent) rather than concrete struct pointers

#### Evidence Trail:
- `+75%` **[DIP_INTERFACE_PARAMETER]** DIP Adherence: Function 'RecordEvent' depends on interface abstraction(s) (ctx context.Context, event RewardEvent) rather than concrete struct pointers -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:93:1`

### #583 HIGH_COHESION_LOW_COUPLING on `dependencies`
- **Category:** `principle`
- **Confidence:** **80%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/cmd/api/main.go:193:1`
- **Summary:** High Coupling (High Fan-Out): Struct 'dependencies' couples directly with 33 distinct types (app_handler.AdminHandler, app_handler.AdminPaywallsHandler, app_handler.AnalyticsHandlersExtended, app_handler.AppSettingsHandler...)

#### Evidence Trail:
- `+80%` **[COUPLING_HIGH_FAN_OUT]** High Coupling (High Fan-Out): Struct 'dependencies' couples directly with 33 distinct types (app_handler.AdminHandler, app_handler.AdminPaywallsHandler, app_handler.AnalyticsHandlersExtended, app_handler.AppSettingsHandler...) -> `/Volumes/External/Code/paywall-iap/backend/cmd/api/main.go:193:1`

### #584 HIGH_COHESION_LOW_COUPLING on `Config`
- **Category:** `principle`
- **Confidence:** **80%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/config/config.go:12:1`
- **Summary:** High Coupling (High Fan-Out): Struct 'Config' couples directly with 8 distinct types (DatabaseConfig, IAPConfig, JWTConfig, LagoConfig...)

#### Evidence Trail:
- `+80%` **[COUPLING_HIGH_FAN_OUT]** High Coupling (High Fan-Out): Struct 'Config' couples directly with 8 distinct types (DatabaseConfig, IAPConfig, JWTConfig, LagoConfig...) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/config/config.go:12:1`

### #585 HIGH_COHESION_LOW_COUPLING on `AdvancedBanditEngine`
- **Category:** `principle`
- **Confidence:** **80%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:16:1`
- **Summary:** High Coupling (High Fan-Out): Struct 'AdvancedBanditEngine' couples directly with 11 distinct types (BanditCache, BanditRepository, CurrencyRateService, DelayedRewardStrategy...)

#### Evidence Trail:
- `+80%` **[COUPLING_HIGH_FAN_OUT]** High Coupling (High Fan-Out): Struct 'AdvancedBanditEngine' couples directly with 11 distinct types (BanditCache, BanditRepository, CurrencyRateService, DelayedRewardStrategy...) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:16:1`

### #586 HIGH_COHESION_LOW_COUPLING on `AdminHandler`
- **Category:** `principle`
- **Confidence:** **80%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin.go:31:1`
- **Summary:** High Coupling (High Fan-Out): Struct 'AdminHandler' couples directly with 15 distinct types (asynq.Client, domainRepo.SubscriptionRepository, domainRepo.UserRepository, generated.Queries...)

#### Evidence Trail:
- `+80%` **[COUPLING_HIGH_FAN_OUT]** High Coupling (High Fan-Out): Struct 'AdminHandler' couples directly with 15 distinct types (asynq.Client, domainRepo.SubscriptionRepository, domainRepo.UserRepository, generated.Queries...) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin.go:31:1`

### #587 KISS on `main`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/cmd/migrator/main.go:15:1`
- **Summary:** KISS Violation (High Complexity): Function 'main' has cyclomatic complexity of 13 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'main' has cyclomatic complexity of 13 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/cmd/migrator/main.go:15:1`

### #588 KISS on `main`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/cmd/seed/main.go:26:1`
- **Summary:** KISS Violation (High Complexity): Function 'main' has cyclomatic complexity of 14 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'main' has cyclomatic complexity of 14 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/cmd/seed/main.go:26:1`

### #589 KISS on `doJSON`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/cmd/loadgen/main.go:107:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'doJSON' takes 5 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'doJSON' takes 5 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/cmd/loadgen/main.go:107:1`

### #590 KISS on `runUser`
- **Category:** `principle`
- **Confidence:** **92%** [VERY_HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/cmd/loadgen/main.go:145:1`
- **Summary:** KISS Violation (High Complexity): Function 'runUser' has cyclomatic complexity of 20 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'runUser' has cyclomatic complexity of 20 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/cmd/loadgen/main.go:145:1`
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'runUser' takes 7 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/cmd/loadgen/main.go:145:1`

### #591 KISS on `AdminMiddleware`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/middleware/admin_auth.go:16:1`
- **Summary:** KISS Violation (High Complexity): Function 'AdminMiddleware' has cyclomatic complexity of 10 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'AdminMiddleware' has cyclomatic complexity of 10 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/application/middleware/admin_auth.go:16:1`

### #592 KISS on `Authenticate`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/middleware/jwt.go:50:1`
- **Summary:** KISS Violation (High Complexity): Function 'Authenticate' has cyclomatic complexity of 10 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'Authenticate' has cyclomatic complexity of 10 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/application/middleware/jwt.go:50:1`

### #593 KISS on `Execute`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/register.go:31:1`
- **Summary:** KISS Violation (High Complexity): Function 'Execute' has cyclomatic complexity of 14 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'Execute' has cyclomatic complexity of 14 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/register.go:31:1`

### #594 KISS on `NewVerifyIAPCommand`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/verify_iap.go:58:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'NewVerifyIAPCommand' takes 5 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'NewVerifyIAPCommand' takes 5 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/verify_iap.go:58:1`

### #595 KISS on `NewVerifyIAPCommandLegacy`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/verify_iap.go:75:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'NewVerifyIAPCommandLegacy' takes 5 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'NewVerifyIAPCommandLegacy' takes 5 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/verify_iap.go:75:1`

### #596 KISS on `Execute`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/application/command/verify_iap.go:91:1`
- **Summary:** KISS Violation (High Complexity): Function 'Execute' has cyclomatic complexity of 14 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'Execute' has cyclomatic complexity of 14 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/application/command/verify_iap.go:91:1`

### #597 KISS on `CalculateWinProbabilities`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/bandit_maintenance_jobs.go:163:1`
- **Summary:** KISS Violation (High Complexity): Function 'CalculateWinProbabilities' has cyclomatic complexity of 13 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'CalculateWinProbabilities' has cyclomatic complexity of 13 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/bandit_maintenance_jobs.go:163:1`

### #598 KISS on `HandleProcessWebhook`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/tasks.go:221:1`
- **Summary:** KISS Violation (High Complexity): Function 'HandleProcessWebhook' has cyclomatic complexity of 10 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'HandleProcessWebhook' has cyclomatic complexity of 10 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/tasks.go:221:1`

### #599 KISS on `HandleSendNotification`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/tasks.go:328:1`
- **Summary:** KISS Violation (High Complexity): Function 'HandleSendNotification' has cyclomatic complexity of 66 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'HandleSendNotification' has cyclomatic complexity of 66 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/tasks.go:328:1`

### #600 KISS on `RegisterBanditMaintenanceTasks`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/currency_asynq.go:74:1`
- **Summary:** KISS Violation (High Complexity): Function 'RegisterBanditMaintenanceTasks' has cyclomatic complexity of 14 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'RegisterBanditMaintenanceTasks' has cyclomatic complexity of 14 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/worker/tasks/currency_asynq.go:74:1`

### #601 KISS on `SetFunnelData`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:281:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'SetFunnelData' takes 5 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'SetFunnelData' takes 5 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/cache/analytics_cache.go:281:1`

### #602 KISS on `UpdateEventStatus`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/matomo_fallback.go:132:1`
- **Summary:** KISS Violation (High Complexity): Function 'UpdateEventStatus' has cyclomatic complexity of 10 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'UpdateEventStatus' has cyclomatic complexity of 10 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/matomo_fallback.go:132:1`

### #603 KISS on `SaveConversion`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:320:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'SaveConversion' takes 5 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'SaveConversion' takes 5 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:320:1`

### #604 KISS on `ProcessPendingConversion`
- **Category:** `principle`
- **Confidence:** **92%** [VERY_HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:466:1`
- **Summary:** KISS Violation (High Complexity): Function 'ProcessPendingConversion' has cyclomatic complexity of 12 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'ProcessPendingConversion' has cyclomatic complexity of 12 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:466:1`
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'ProcessPendingConversion' takes 6 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:466:1`

### #605 KISS on `ProcessExpiredPendingReward`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:558:1`
- **Summary:** KISS Violation (High Complexity): Function 'ProcessExpiredPendingReward' has cyclomatic complexity of 10 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'ProcessExpiredPendingReward' has cyclomatic complexity of 10 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/bandit_repository.go:558:1`

### #606 KISS on `GetAuditLogPaginated`
- **Category:** `principle`
- **Confidence:** **92%** [VERY_HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/analytics_repository_impl.go:268:1`
- **Summary:** KISS Violation (High Complexity): Function 'GetAuditLogPaginated' has cyclomatic complexity of 14 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'GetAuditLogPaginated' has cyclomatic complexity of 14 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/analytics_repository_impl.go:268:1`
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'GetAuditLogPaginated' takes 7 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/analytics_repository_impl.go:268:1`

### #607 KISS on `insertExperimentLifecycleAudit`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:241:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'insertExperimentLifecycleAudit' takes 6 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'insertExperimentLifecycleAudit' takes 6 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:241:1`

### #608 KISS on `syncDraftExperimentArms`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:112:1`
- **Summary:** KISS Violation (High Complexity): Function 'syncDraftExperimentArms' has cyclomatic complexity of 19 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'syncDraftExperimentArms' has cyclomatic complexity of 19 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:112:1`

### #609 KISS on `UpdateExperimentStatus`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:216:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'UpdateExperimentStatus' takes 5 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'UpdateExperimentStatus' takes 5 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:216:1`

### #610 KISS on `UpdateExperimentStatusWithAudit`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:319:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'UpdateExperimentStatusWithAudit' takes 7 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'UpdateExperimentStatusWithAudit' takes 7 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:319:1`

### #611 KISS on `UpdateExperimentStatusAndAutomationPolicyWithAudit`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:351:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'UpdateExperimentStatusAndAutomationPolicyWithAudit' takes 8 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'UpdateExperimentStatusAndAutomationPolicyWithAudit' takes 8 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/experiment_admin_repository.go:351:1`

### #612 KISS on `scanAndDecryptCredentials`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:305:1`
- **Summary:** KISS Violation (High Complexity): Function 'scanAndDecryptCredentials' has cyclomatic complexity of 11 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'scanAndDecryptCredentials' has cyclomatic complexity of 11 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/persistence/repository/app_repository_impl.go:305:1`

### #613 KISS on `VerifyReceipt`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/verifier.go:100:1`
- **Summary:** KISS Violation (High Complexity): Function 'VerifyReceipt' has cyclomatic complexity of 10 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'VerifyReceipt' has cyclomatic complexity of 10 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/infrastructure/external/iap/verifier.go:100:1`

### #614 KISS on `NewSubscription`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/subscription.go:50:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'NewSubscription' takes 6 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'NewSubscription' takes 6 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/subscription.go:50:1`

### #615 KISS on `NewUser`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/user.go:47:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'NewUser' takes 6 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'NewUser' takes 6 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/user.go:47:1`

### #616 KISS on `NewTransaction`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/transaction.go:31:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'NewTransaction' takes 5 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'NewTransaction' takes 5 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/transaction.go:31:1`

### #617 KISS on `NewWinbackOffer`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/winback_offer.go:43:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'NewWinbackOffer' takes 5 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'NewWinbackOffer' takes 5 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/entity/winback_offer.go:43:1`

### #618 KISS on `NewAdvancedBanditEngine`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:72:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'NewAdvancedBanditEngine' takes 7 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'NewAdvancedBanditEngine' takes 7 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:72:1`

### #619 KISS on `SelectArm`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:209:1`
- **Summary:** KISS Violation (High Complexity): Function 'SelectArm' has cyclomatic complexity of 13 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'SelectArm' has cyclomatic complexity of 13 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:209:1`

### #620 KISS on `RecordReward`
- **Category:** `principle`
- **Confidence:** **92%** [VERY_HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:271:1`
- **Summary:** KISS Violation (High Complexity): Function 'RecordReward' has cyclomatic complexity of 14 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'RecordReward' has cyclomatic complexity of 14 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:271:1`
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'RecordReward' takes 7 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:271:1`

### #621 KISS on `ProcessConversion`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:365:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'ProcessConversion' takes 5 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'ProcessConversion' takes 5 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:365:1`

### #622 KISS on `GetMetrics`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:464:1`
- **Summary:** KISS Violation (High Complexity): Function 'GetMetrics' has cyclomatic complexity of 10 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'GetMetrics' has cyclomatic complexity of 10 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:464:1`

### #623 KISS on `TrimConfiguredWindows`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:571:1`
- **Summary:** KISS Violation (High Complexity): Function 'TrimConfiguredWindows' has cyclomatic complexity of 10 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'TrimConfiguredWindows' has cyclomatic complexity of 10 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:571:1`

### #624 KISS on `SyncObjectiveStats`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:608:1`
- **Summary:** KISS Violation (High Complexity): Function 'SyncObjectiveStats' has cyclomatic complexity of 15 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'SyncObjectiveStats' has cyclomatic complexity of 15 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:608:1`

### #625 KISS on `RunMaintenanceDetailed`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:794:1`
- **Summary:** KISS Violation (High Complexity): Function 'RunMaintenanceDetailed' has cyclomatic complexity of 11 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'RunMaintenanceDetailed' has cyclomatic complexity of 11 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/advanced_bandit_engine.go:794:1`

### #626 KISS on `CreateFlag`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/feature_flag_service.go:38:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'CreateFlag' takes 5 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'CreateFlag' takes 5 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/feature_flag_service.go:38:1`

### #627 KISS on `TrackEvent`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/matomo_forwarder.go:66:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'TrackEvent' takes 7 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'TrackEvent' takes 7 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/matomo_forwarder.go:66:1`

### #628 KISS on `TrackPurchase`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/matomo_forwarder.go:100:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'TrackPurchase' takes 6 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'TrackPurchase' takes 6 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/matomo_forwarder.go:100:1`

### #629 KISS on `SelectArm`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:245:1`
- **Summary:** KISS Violation (High Complexity): Function 'SelectArm' has cyclomatic complexity of 11 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'SelectArm' has cyclomatic complexity of 11 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:245:1`

### #630 KISS on `TrackImpression`
- **Category:** `principle`
- **Confidence:** **92%** [VERY_HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:376:1`
- **Summary:** KISS Violation (High Complexity): Function 'TrackImpression' has cyclomatic complexity of 14 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'TrackImpression' has cyclomatic complexity of 14 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:376:1`
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'TrackImpression' takes 5 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:376:1`

### #631 KISS on `UpdateRewardWithEvent`
- **Category:** `principle`
- **Confidence:** **92%** [VERY_HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:438:1`
- **Summary:** KISS Violation (High Complexity): Function 'UpdateRewardWithEvent' has cyclomatic complexity of 16 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'UpdateRewardWithEvent' has cyclomatic complexity of 16 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:438:1`
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'UpdateRewardWithEvent' takes 5 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:438:1`

### #632 KISS on `CalculateWinProbability`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:668:1`
- **Summary:** KISS Violation (High Complexity): Function 'CalculateWinProbability' has cyclomatic complexity of 10 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'CalculateWinProbability' has cyclomatic complexity of 10 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/bandit_service.go:668:1`

### #633 KISS on `NewHybridObjectiveStrategy`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:54:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'NewHybridObjectiveStrategy' takes 5 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'NewHybridObjectiveStrategy' takes 5 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:54:1`

### #634 KISS on `calculateHybridScore`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:166:1`
- **Summary:** KISS Violation (High Complexity): Function 'calculateHybridScore' has cyclomatic complexity of 10 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'calculateHybridScore' has cyclomatic complexity of 10 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:166:1`

### #635 KISS on `RecordObjectiveReward`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:258:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'RecordObjectiveReward' takes 5 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'RecordObjectiveReward' takes 5 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:258:1`

### #636 KISS on `GetObjectiveScores`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:315:1`
- **Summary:** KISS Violation (High Complexity): Function 'GetObjectiveScores' has cyclomatic complexity of 13 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'GetObjectiveScores' has cyclomatic complexity of 13 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/hybrid_objective_strategy.go:315:1`

### #637 KISS on `ExecuteScheduled`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/automation_job_execution_service.go:58:1`
- **Summary:** KISS Violation (High Complexity): Function 'ExecuteScheduled' has cyclomatic complexity of 10 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'ExecuteScheduled' has cyclomatic complexity of 10 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/automation_job_execution_service.go:58:1`

### #638 KISS on `ProcessConversion`
- **Category:** `principle`
- **Confidence:** **92%** [VERY_HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:124:1`
- **Summary:** KISS Violation (High Complexity): Function 'ProcessConversion' has cyclomatic complexity of 17 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'ProcessConversion' has cyclomatic complexity of 17 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:124:1`
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'ProcessConversion' takes 6 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:124:1`

### #639 KISS on `ProcessExpiredRewards`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:253:1`
- **Summary:** KISS Violation (High Complexity): Function 'ProcessExpiredRewards' has cyclomatic complexity of 13 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'ProcessExpiredRewards' has cyclomatic complexity of 13 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/delayed_reward_strategy.go:253:1`

### #640 KISS on `CreateWinbackOffer`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/winback_service.go:42:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'CreateWinbackOffer' takes 6 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'CreateWinbackOffer' takes 6 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/winback_service.go:42:1`

### #641 KISS on `CreateWinbackCampaignForChurnedUsers`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/winback_service.go:133:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'CreateWinbackCampaignForChurnedUsers' takes 6 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'CreateWinbackCampaignForChurnedUsers' takes 6 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/winback_service.go:133:1`

### #642 KISS on `RecordRewardWithCurrency`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/currency_conversion_strategy.go:157:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'RecordRewardWithCurrency' takes 7 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'RecordRewardWithCurrency' takes 7 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/currency_conversion_strategy.go:157:1`

### #643 KISS on `NewLTVService`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:55:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'NewLTVService' takes 5 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'NewLTVService' takes 5 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:55:1`

### #644 KISS on `CalculateLTV`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:91:1`
- **Summary:** KISS Violation (High Complexity): Function 'CalculateLTV' has cyclomatic complexity of 13 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'CalculateLTV' has cyclomatic complexity of 13 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:91:1`

### #645 KISS on `GetCohortLTV`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:313:1`
- **Summary:** KISS Violation (High Complexity): Function 'GetCohortLTV' has cyclomatic complexity of 10 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'GetCohortLTV' has cyclomatic complexity of 10 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ltv_service.go:313:1`

### #646 KISS on `NewLinUCBSelectionStrategy`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/linucb_strategy.go:33:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'NewLinUCBSelectionStrategy' takes 5 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'NewLinUCBSelectionStrategy' takes 5 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/linucb_strategy.go:33:1`

### #647 KISS on `LogAction`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/audit_service.go:23:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'LogAction' takes 6 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'LogAction' takes 6 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/audit_service.go:23:1`

### #648 KISS on `Recommend`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:124:1`
- **Summary:** KISS Violation (High Complexity): Function 'Recommend' has cyclomatic complexity of 12 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'Recommend' has cyclomatic complexity of 12 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_winner_recommendation_service.go:124:1`

### #649 KISS on `GetAuditLogPaginated`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_service.go:116:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'GetAuditLogPaginated' takes 7 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'GetAuditLogPaginated' takes 7 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/analytics_service.go:116:1`

### #650 KISS on `ProcessDunningAttempt`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/dunning_service.go:63:1`
- **Summary:** KISS Violation (High Complexity): Function 'ProcessDunningAttempt' has cyclomatic complexity of 10 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'ProcessDunningAttempt' has cyclomatic complexity of 10 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/dunning_service.go:63:1`

### #651 KISS on `NewCurrencyRateService`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/currency_service.go:61:1`
- **Summary:** KISS Violation (High Complexity): Function 'NewCurrencyRateService' has cyclomatic complexity of 29 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'NewCurrencyRateService' has cyclomatic complexity of 29 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/currency_service.go:61:1`

### #652 KISS on `TrackRevenue`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ab_analytics_service.go:66:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'TrackRevenue' takes 5 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'TrackRevenue' takes 5 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/ab_analytics_service.go:66:1`

### #653 KISS on `NewSlidingWindowStrategy`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:38:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'NewSlidingWindowStrategy' takes 5 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'NewSlidingWindowStrategy' takes 5 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/sliding_window_strategy.go:38:1`

### #654 KISS on `RepairExperiment`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:73:1`
- **Summary:** KISS Violation (High Complexity): Function 'RepairExperiment' has cyclomatic complexity of 10 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'RepairExperiment' has cyclomatic complexity of 10 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:73:1`

### #655 KISS on `syncObjectiveStats`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:124:1`
- **Summary:** KISS Violation (High Complexity): Function 'syncObjectiveStats' has cyclomatic complexity of 10 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'syncObjectiveStats' has cyclomatic complexity of 10 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/domain/service/experiment_repair_service.go:124:1`

### #656 KISS on `NewWebhookHandler`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/webhook.go:34:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'NewWebhookHandler' takes 5 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'NewWebhookHandler' takes 5 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/webhook.go:34:1`

### #657 KISS on `StripeWebhook`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/webhook.go:51:1`
- **Summary:** KISS Violation (High Complexity): Function 'StripeWebhook' has cyclomatic complexity of 11 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'StripeWebhook' has cyclomatic complexity of 11 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/webhook.go:51:1`

### #658 KISS on `AppleWebhook`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/webhook.go:123:1`
- **Summary:** KISS Violation (High Complexity): Function 'AppleWebhook' has cyclomatic complexity of 10 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'AppleWebhook' has cyclomatic complexity of 10 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/webhook.go:123:1`

### #659 KISS on `GoogleWebhook`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/webhook.go:204:1`
- **Summary:** KISS Violation (High Complexity): Function 'GoogleWebhook' has cyclomatic complexity of 10 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'GoogleWebhook' has cyclomatic complexity of 10 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/webhook.go:204:1`

### #660 KISS on `experimentAutomationPolicyChangedFields`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_experiments.go:424:1`
- **Summary:** KISS Violation (High Complexity): Function 'experimentAutomationPolicyChangedFields' has cyclomatic complexity of 17 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'experimentAutomationPolicyChangedFields' has cyclomatic complexity of 17 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_experiments.go:424:1`

### #661 KISS on `validateCreateAdminExperimentRequest`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_experiments.go:629:1`
- **Summary:** KISS Violation (High Complexity): Function 'validateCreateAdminExperimentRequest' has cyclomatic complexity of 25 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'validateCreateAdminExperimentRequest' has cyclomatic complexity of 25 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_experiments.go:629:1`

### #662 KISS on `validateUpdateAdminExperimentRequest`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_experiments.go:709:1`
- **Summary:** KISS Violation (High Complexity): Function 'validateUpdateAdminExperimentRequest' has cyclomatic complexity of 27 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'validateUpdateAdminExperimentRequest' has cyclomatic complexity of 27 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_experiments.go:709:1`

### #663 KISS on `scanAdminExperiment`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_experiments.go:828:1`
- **Summary:** KISS Violation (High Complexity): Function 'scanAdminExperiment' has cyclomatic complexity of 13 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'scanAdminExperiment' has cyclomatic complexity of 13 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_experiments.go:828:1`

### #664 KISS on `AdminLogout`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/auth.go:179:1`
- **Summary:** KISS Violation (High Complexity): Function 'AdminLogout' has cyclomatic complexity of 18 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'AdminLogout' has cyclomatic complexity of 18 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/auth.go:179:1`

### #665 KISS on `PutAppSettings`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_app_settings.go:61:1`
- **Summary:** KISS Violation (High Complexity): Function 'PutAppSettings' has cyclomatic complexity of 19 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'PutAppSettings' has cyclomatic complexity of 19 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_app_settings.go:61:1`

### #666 KISS on `UpdateApp`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_apps.go:131:1`
- **Summary:** KISS Violation (High Complexity): Function 'UpdateApp' has cyclomatic complexity of 10 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'UpdateApp' has cyclomatic complexity of 10 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_apps.go:131:1`

### #667 KISS on `NewAdminHandler`
- **Category:** `principle`
- **Confidence:** **70%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin.go:50:1`
- **Summary:** KISS Violation (Long Parameter List): Function 'NewAdminHandler' takes 12 parameters; consider a Config struct or Functional Options

#### Evidence Trail:
- `+70%` **[KISS_LONG_PARAMETER_LIST]** KISS Violation (Long Parameter List): Function 'NewAdminHandler' takes 12 parameters; consider a Config struct or Functional Options -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin.go:50:1`

### #668 KISS on `SearchUsers`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin.go:415:1`
- **Summary:** KISS Violation (High Complexity): Function 'SearchUsers' has cyclomatic complexity of 10 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'SearchUsers' has cyclomatic complexity of 10 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin.go:415:1`

### #669 KISS on `ForceRenew`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin.go:563:1`
- **Summary:** KISS Violation (High Complexity): Function 'ForceRenew' has cyclomatic complexity of 10 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'ForceRenew' has cyclomatic complexity of 10 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin.go:563:1`

### #670 KISS on `ListWebhooks`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin.go:746:1`
- **Summary:** KISS Violation (High Complexity): Function 'ListWebhooks' has cyclomatic complexity of 15 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'ListWebhooks' has cyclomatic complexity of 15 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin.go:746:1`

### #671 KISS on `ListSubscriptions`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin.go:1008:1`
- **Summary:** KISS Violation (High Complexity): Function 'ListSubscriptions' has cyclomatic complexity of 14 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'ListSubscriptions' has cyclomatic complexity of 14 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin.go:1008:1`

### #672 KISS on `ListTransactions`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin.go:1235:1`
- **Summary:** KISS Violation (High Complexity): Function 'ListTransactions' has cyclomatic complexity of 13 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'ListTransactions' has cyclomatic complexity of 13 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin.go:1235:1`

### #673 KISS on `Statistics`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/bandit.go:310:1`
- **Summary:** KISS Violation (High Complexity): Function 'Statistics' has cyclomatic complexity of 16 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'Statistics' has cyclomatic complexity of 16 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/bandit.go:310:1`

### #674 KISS on `UpdatePaywall`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_paywalls.go:159:1`
- **Summary:** KISS Violation (High Complexity): Function 'UpdatePaywall' has cyclomatic complexity of 10 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'UpdatePaywall' has cyclomatic complexity of 10 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/admin_paywalls.go:159:1`

### #675 KISS on `RunMaintenance`
- **Category:** `principle`
- **Confidence:** **75%** [HIGH]
- **Primary Location:** `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/bandit_advanced.go:423:1`
- **Summary:** KISS Violation (High Complexity): Function 'RunMaintenance' has cyclomatic complexity of 10 (control branch points)

#### Evidence Trail:
- `+75%` **[KISS_HIGH_CYCLOMATIC_COMPLEXITY]** KISS Violation (High Complexity): Function 'RunMaintenance' has cyclomatic complexity of 10 (control branch points) -> `/Volumes/External/Code/paywall-iap/backend/internal/interfaces/http/handlers/bandit_advanced.go:423:1`
