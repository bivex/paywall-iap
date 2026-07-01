package appctx

import (
	"context"

	"github.com/google/uuid"
)

type contextKey struct{}

// WithAppID returns a new context carrying appID.
func WithAppID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, contextKey{}, id)
}

// AppIDFromCtx returns the app_id stored in ctx and whether it was present.
func AppIDFromCtx(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(contextKey{}).(uuid.UUID)
	return id, ok
}

// MustAppIDFromCtx returns the app_id from context, or uuid.Nil if not set.
func MustAppIDFromCtx(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(contextKey{}).(uuid.UUID)
	return id
}
