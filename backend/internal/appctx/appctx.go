// Package appctx provides helpers for propagating app_id through context.Context.
// Repositories use AppIDFromCtx; Gin middleware injects via WithAppID.
package appctx

import (
	"context"

	"github.com/google/uuid"
)

type contextKey struct{}

// WithAppID returns a new context carrying appID.
func WithAppID(ctx context.Context, appID uuid.UUID) context.Context {
	return context.WithValue(ctx, contextKey{}, appID)
}

// AppIDFromCtx returns the app_id stored in ctx and whether it was present.
func AppIDFromCtx(ctx context.Context) (uuid.UUID, bool) {
	v, ok := ctx.Value(contextKey{}).(uuid.UUID)
	return v, ok
}
