package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/password9090/paywall-iap/internal/domain/entity"
	"github.com/password9090/paywall-iap/internal/domain/repository"
	domainErrors "github.com/password9090/paywall-iap/internal/domain/errors"
	"github.com/password9090/paywall-iap/internal/infrastructure/persistence/sqlc/generated"
)

type subscriptionRepositoryImpl struct {
	queries *generated.Queries
}

// NewSubscriptionRepository creates a new subscription repository implementation
func NewSubscriptionRepository(queries *generated.Queries) repository.SubscriptionRepository {
	return &subscriptionRepositoryImpl{queries: queries}
}

func (r *subscriptionRepositoryImpl) Create(ctx context.Context, sub *entity.Subscription) error {
	params := generated.CreateSubscriptionParams{
		UserID:    sub.UserID,
		Status:    string(sub.Status),
		Source:    string(sub.Source),
		Platform:  sub.Platform,
		ProductID: sub.ProductID,
		PlanType:  string(sub.PlanType),
		ExpiresAt: sub.ExpiresAt,
		AutoRenew: sub.AutoRenew,
	}

	_, err := r.queries.CreateSubscription(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

func (r *subscriptionRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error) {
	row, err := r.queries.GetSubscriptionByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("subscription not found: %w", domainErrors.ErrSubscriptionNotFound)
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return r.mapToEntity(row), nil
}

func (r *subscriptionRepositoryImpl) GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*entity.Subscription, error) {
	row, err := r.queries.GetActiveSubscriptionByUserID(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("active subscription not found: %w", domainErrors.ErrSubscriptionNotActive)
		}
		return nil, fmt.Errorf("failed to get active subscription: %w", err)
	}

	return r.mapToEntity(row), nil
}

func (r *subscriptionRepositoryImpl) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Subscription, error) {
	// For now, return empty array if no subscriptions
	// In full implementation, add query to GetSubscriptionsByUserID
	return []*entity.Subscription{}, nil
}

func (r *subscriptionRepositoryImpl) Update(ctx context.Context, sub *entity.Subscription) error {
	params := generated.UpdateSubscriptionStatusParams{
		ID:     sub.ID,
		Status: string(sub.Status),
	}

	_, err := r.queries.UpdateSubscriptionStatus(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

func (r *subscriptionRepositoryImpl) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.SubscriptionStatus) error {
	params := generated.UpdateSubscriptionStatusParams{
		ID:     id,
		Status: string(status),
	}

	_, err := r.queries.UpdateSubscriptionStatus(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update subscription status: %w", err)
	}

	return nil
}

func (r *subscriptionRepositoryImpl) UpdateExpiry(ctx context.Context, id uuid.UUID, expiresAt interface{}) error {
	params := generated.UpdateSubscriptionExpiryParams{
		ID:        id,
		ExpiresAt: expiresAt.(time.Time),
	}

	_, err := r.queries.UpdateSubscriptionExpiry(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update subscription expiry: %w", err)
	}

	return nil
}

func (r *subscriptionRepositoryImpl) Cancel(ctx context.Context, id uuid.UUID) error {
	_, err := r.queries.CancelSubscription(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	return nil
}

func (r *subscriptionRepositoryImpl) CanAccess(ctx context.Context, userID uuid.UUID) (bool, error) {
	_, err := r.queries.GetAccessCheck(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check access: %w", err)
	}

	return true, nil
}

func (r *subscriptionRepositoryImpl) mapToEntity(row generated.Subscription) *entity.Subscription {
	return &entity.Subscription{
		ID:        row.ID,
		UserID:    row.UserID,
		Status:    entity.SubscriptionStatus(row.Status),
		Source:    entity.SubscriptionSource(row.Source),
		Platform:  row.Platform,
		ProductID: row.ProductID,
		PlanType:  entity.PlanType(row.PlanType),
		ExpiresAt: row.ExpiresAt,
		AutoRenew: row.AutoRenew,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		DeletedAt: row.DeletedAt,
	}
}
