package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/bivex/paywall-iap/internal/domain/entity"
)

// MockWinbackOfferRepository is a mock implementation of WinbackOfferRepository
type MockWinbackOfferRepository struct {
	mock.Mock
}

// NewMockWinbackOfferRepository creates a new mock winback offer repository
func NewMockWinbackOfferRepository() *MockWinbackOfferRepository {
	return &MockWinbackOfferRepository{}
}

func (m *MockWinbackOfferRepository) Create(ctx context.Context, offer *entity.WinbackOffer) error {
	args := m.Called(ctx, offer)
	return args.Error(0)
}

func (m *MockWinbackOfferRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.WinbackOffer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WinbackOffer), args.Error(1)
}

func (m *MockWinbackOfferRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.WinbackOffer, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.WinbackOffer), args.Error(1)
}

func (m *MockWinbackOfferRepository) GetActiveByUserAndCampaign(ctx context.Context, userID uuid.UUID, campaignID string) (*entity.WinbackOffer, error) {
	args := m.Called(ctx, userID, campaignID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WinbackOffer), args.Error(1)
}

func (m *MockWinbackOfferRepository) Update(ctx context.Context, offer *entity.WinbackOffer) error {
	args := m.Called(ctx, offer)
	return args.Error(0)
}

func (m *MockWinbackOfferRepository) GetExpiredOffers(ctx context.Context, limit int) ([]*entity.WinbackOffer, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.WinbackOffer), args.Error(1)
}
