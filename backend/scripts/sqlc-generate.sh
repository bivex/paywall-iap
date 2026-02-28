#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/../internal/infrastructure/persistence/sqlc"

echo "Generating sqlc code..."
sqlc generate

echo "sqlc code generated successfully"
