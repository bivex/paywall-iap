package tasks_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/bivex/paywall-iap/internal/worker/tasks"
)

type mockAnalyticsRepo struct {
	mock.Mock
}

func (m *mockAnalyticsRepo) StoreCohortData(ctx context.Context, aggregate *tasks.CohortAggregate) error {
	args := m.Called(ctx, aggregate)
	return args.Error(0)
}

func (m *mockAnalyticsRepo) GetCohortData(ctx context.Context, metricName string, date time.Time) (*tasks.CohortAggregate, error) {
	args := m.Called(ctx, metricName, date)
	if res := args.Get(0); res != nil {
		return res.(*tasks.CohortAggregate), args.Error(1)
	}
	return nil, args.Error(1)
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-4
}

func TestCalculateLTVFromCohorts_Modeled(t *testing.T) {
	repo := new(mockAnalyticsRepo)
	worker := tasks.NewCohortWorker(nil, repo)

	aggregate := &tasks.CohortAggregate{
		CohortSize: 100,
		Revenue: map[string]float64{
			"day30": 1000.0, // ARPU30 = 10.0
		},
	}

	repo.On("GetCohortData", mock.Anything, mock.Anything, mock.Anything).Return(aggregate, nil)

	ltv, err := worker.CalculateLTVFromCohorts(context.Background(), uuid.New())
	assert.NoError(t, err)
	assert.NotNil(t, ltv)

	assert.True(t, almostEqual(ltv["ltv30"], 10.0))
	assert.True(t, almostEqual(ltv["ltv90"], 16.5))  // 10.0 * 1.65
	assert.True(t, almostEqual(ltv["ltv365"], 28.5)) // 10.0 * 2.85
}

func TestCalculateLTVFromCohorts_ActualRecorded(t *testing.T) {
	repo := new(mockAnalyticsRepo)
	worker := tasks.NewCohortWorker(nil, repo)

	aggregate := &tasks.CohortAggregate{
		CohortSize: 50,
		Revenue: map[string]float64{
			"day30":  500.0,  // ARPU30 = 10.0
			"day90":  750.0,  // ARPU90 = 15.0
			"day365": 1200.0, // ARPU365 = 24.0
		},
	}

	repo.On("GetCohortData", mock.Anything, mock.Anything, mock.Anything).Return(aggregate, nil)

	ltv, err := worker.CalculateLTVFromCohorts(context.Background(), uuid.New())
	assert.NoError(t, err)
	assert.NotNil(t, ltv)

	assert.True(t, almostEqual(ltv["ltv30"], 10.0))
	assert.True(t, almostEqual(ltv["ltv90"], 15.0))
	assert.True(t, almostEqual(ltv["ltv365"], 24.0))
}
