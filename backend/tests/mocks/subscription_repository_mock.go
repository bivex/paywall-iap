package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/bivex/paywall-iap/internal/domain/entity"
)

// MockSubscriptionRepository is a mock implementation of SubscriptionRepository
type MockSubscriptionRepository struct {
	mock.Mock
}

// NewMockSubscriptionRepository creates a new mock subscription repository
func NewMockSubscriptionRepository() *MockSubscriptionRepository {
	return &MockSubscriptionRepository{}
}

func (m *MockSubscriptionRepository) Create(ctx context.Context, subscription *entity.Subscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*entity.Subscription, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Subscription, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) Update(ctx context.Context, subscription *entity.Subscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.SubscriptionStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) UpdateExpiry(ctx context.Context, id uuid.UUID, expiresAt interface{}) error {
	args := m.Called(ctx, id, expiresAt)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) Cancel(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) CanAccess(ctx context.Context, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}
