# Application Layer

Use cases, commands, queries, and DTOs.

## Structure

- `command/` - Command handlers (write operations)
- `query/` - Query handlers (read operations)
- `dto/` - Data transfer objects
- `middleware/` - Application middleware
- `validator/` - Request validators

## Dependency Rule

Depends on domain layer only.
