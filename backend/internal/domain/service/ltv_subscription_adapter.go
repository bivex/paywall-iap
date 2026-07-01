package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	domainRepo "github.com/bivex/paywall-iap/internal/domain/repository"
)

// ltvSubscriptionAdapter adapts domain SubscriptionRepository to LTVService's internal interface
type ltvSubscriptionAdapter struct {
	repo domainRepo.SubscriptionRepository
}

// NewLTVSubscriptionAdapter wraps a domain SubscriptionRepository for use with LTVService
func NewLTVSubscriptionAdapter(repo domainRepo.SubscriptionRepository) SubscriptionRepository {
	return &ltvSubscriptionAdapter{repo: repo}
}

func (a *ltvSubscriptionAdapter) GetUserSubscriptions(ctx context.Context, userID uuid.UUID) ([]Subscription, error) {
	subs, err := a.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]Subscription, 0, len(subs))
	for _, s := range subs {
		sub := Subscription{
			ID:        s.ID,
			UserID:    s.UserID,
			Status:    string(s.Status),
			CreatedAt: s.CreatedAt,
		}
		if !s.ExpiresAt.IsZero() {
			t := s.ExpiresAt
			sub.EndDate = &t
		}
		result = append(result, sub)
	}
	return result, nil
}

func (a *ltvSubscriptionAdapter) GetTotalRevenue(ctx context.Context, userID uuid.UUID) (float64, error) {
	return a.repo.GetTotalRevenue(ctx, userID)
}

// ensure time import is used
var _ = time.Time{}
