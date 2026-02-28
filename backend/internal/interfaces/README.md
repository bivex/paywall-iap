# Interfaces Layer

HTTP handlers, webhooks, middleware.

## Structure

- `http/` - HTTP API (handlers, middleware, response writer)
- `webhook/` - Webhook handlers (Stripe, Apple, Google, Lago)

## Dependency Rule

Depends on application layer. Delegates business logic to use cases.
