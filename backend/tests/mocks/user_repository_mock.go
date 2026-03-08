package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/bivex/paywall-iap/internal/domain/entity"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

// NewMockUserRepository creates a new mock user repository
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{}
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByPlatformID(ctx context.Context, platformUserID string) (*entity.User, error) {
	args := m.Called(ctx, platformUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) ExistsByPlatformID(ctx context.Context, platformUserID string) (bool, error) {
	args := m.Called(ctx, platformUserID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) UpdatePurchaseChannel(ctx context.Context, id uuid.UUID, channel string) error {
	args := m.Called(ctx, id, channel)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateEmail(ctx context.Context, id uuid.UUID, email string) error {
	args := m.Called(ctx, id, email)
	return args.Error(0)
}

func (m *MockUserRepository) IncrementSessionCount(ctx context.Context, id uuid.UUID) (int, error) {
	args := m.Called(ctx, id)
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepository) UpdateHasViewedAds(ctx context.Context, id uuid.UUID, hasViewedAds bool) error {
	args := m.Called(ctx, id, hasViewedAds)
	return args.Error(0)
}
