package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/bivex/paywall-iap/internal/domain/entity"
)

// MockGracePeriodRepository is a mock implementation of GracePeriodRepository
type MockGracePeriodRepository struct {
	mock.Mock
}

// NewMockGracePeriodRepository creates a new mock grace period repository
func NewMockGracePeriodRepository() *MockGracePeriodRepository {
	return &MockGracePeriodRepository{}
}

func (m *MockGracePeriodRepository) Create(ctx context.Context, gracePeriod *entity.GracePeriod) error {
	args := m.Called(ctx, gracePeriod)
	return args.Error(0)
}

func (m *MockGracePeriodRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.GracePeriod, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.GracePeriod), args.Error(1)
}

func (m *MockGracePeriodRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*entity.GracePeriod, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.GracePeriod), args.Error(1)
}

func (m *MockGracePeriodRepository) GetActiveBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) (*entity.GracePeriod, error) {
	args := m.Called(ctx, subscriptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.GracePeriod), args.Error(1)
}

func (m *MockGracePeriodRepository) Update(ctx context.Context, gracePeriod *entity.GracePeriod) error {
	args := m.Called(ctx, gracePeriod)
	return args.Error(0)
}

func (m *MockGracePeriodRepository) GetExpiredGracePeriods(ctx context.Context, limit int) ([]*entity.GracePeriod, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.GracePeriod), args.Error(1)
}

func (m *MockGracePeriodRepository) GetExpiringSoon(ctx context.Context, withinHours int) ([]*entity.GracePeriod, error) {
	args := m.Called(ctx, withinHours)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.GracePeriod), args.Error(1)
}
