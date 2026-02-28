# Domain Layer

Core business logic and enterprise rules.

## Structure

- `entity/` - Business entities (User, Subscription, Transaction, etc.)
- `valueobject/` - Value objects (Money, Email, PlanType, etc.)
- `repository/` - Repository interfaces (implemented by infrastructure)
- `service/` - Domain services (pricing, eligibility, churn risk)
- `event/` - Domain events
- `errors/` - Domain-specific errors

## Dependency Rule

The domain layer has NO dependencies on other layers.
