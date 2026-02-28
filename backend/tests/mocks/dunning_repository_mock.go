package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/bivex/paywall-iap/internal/domain/entity"
)

// MockDunningRepository is a mock implementation of DunningRepository
type MockDunningRepository struct {
	mock.Mock
}

// NewMockDunningRepository creates a new mock dunning repository
func NewMockDunningRepository() *MockDunningRepository {
	return &MockDunningRepository{}
}

func (m *MockDunningRepository) Create(ctx context.Context, dunning *entity.Dunning) error {
	args := m.Called(ctx, dunning)
	return args.Error(0)
}

func (m *MockDunningRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Dunning, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Dunning), args.Error(1)
}

func (m *MockDunningRepository) GetActiveBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) (*entity.Dunning, error) {
	args := m.Called(ctx, subscriptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Dunning), args.Error(1)
}

func (m *MockDunningRepository) Update(ctx context.Context, dunning *entity.Dunning) error {
	args := m.Called(ctx, dunning)
	return args.Error(0)
}

func (m *MockDunningRepository) GetPendingAttempts(ctx context.Context, limit int) ([]*entity.Dunning, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Dunning), args.Error(1)
}
