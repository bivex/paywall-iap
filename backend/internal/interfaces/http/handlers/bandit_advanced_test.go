package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

type routerPathTestRepo struct {
	experimentID uuid.UUID
	config       *service.ExperimentConfig
	arms         []service.Arm
	statsByArmID map[uuid.UUID]*service.ArmStats
}

func (r *routerPathTestRepo) GetArms(_ context.Context, experimentID uuid.UUID) ([]service.Arm, error) {
	if experimentID != r.experimentID {
		return nil, nil
	}
	return r.arms, nil
}

func (r *routerPathTestRepo) GetArmStats(_ context.Context, armID uuid.UUID) (*service.ArmStats, error) {
	return r.statsByArmID[armID], nil
}

func (r *routerPathTestRepo) UpdateArmStats(_ context.Context, _ *service.ArmStats) error { return nil }

func (r *routerPathTestRepo) CreateAssignment(_ context.Context, _ *service.Assignment) error { return nil }

func (r *routerPathTestRepo) GetActiveAssignment(_ context.Context, _, _ uuid.UUID) (*service.Assignment, error) {
	return nil, nil
}

func (r *routerPathTestRepo) GetExperimentConfig(_ context.Context, experimentID uuid.UUID) (*service.ExperimentConfig, error) {
	if experimentID != r.experimentID {
		return nil, nil
	}
	return r.config, nil
}

func (r *routerPathTestRepo) UpdateObjectiveConfig(
	_ context.Context,
	_ uuid.UUID,
	objectiveType service.ObjectiveType,
	objectiveWeights map[string]float64,
) error {
	r.config.ObjectiveType = objectiveType
	r.config.ObjectiveWeights = objectiveWeights
	return nil
}

func (r *routerPathTestRepo) GetUserContext(_ context.Context, userID uuid.UUID) (*service.UserContext, error) {
	return &service.UserContext{UserID: userID}, nil
}

func (r *routerPathTestRepo) SetUserContext(_ context.Context, _ *service.UserContext) error { return nil }

type routerPathTestCache struct{}

func (c *routerPathTestCache) GetArmStats(_ context.Context, _ string) (*service.ArmStats, error) {
	return nil, nil
}

func (c *routerPathTestCache) SetArmStats(_ context.Context, _ string, _ *service.ArmStats, _ time.Duration) error {
	return nil
}

func (c *routerPathTestCache) GetAssignment(_ context.Context, _ string) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (c *routerPathTestCache) SetAssignment(_ context.Context, _ string, _ uuid.UUID, _ time.Duration) error {
	return nil
}

func TestParseUUIDPathParamAfter_ParsesGinStyleExperimentPath(t *testing.T) {
	experimentID := uuid.New()
	req := httptest.NewRequest("GET", "/v1/bandit/experiments/"+experimentID.String()+"/objectives", nil)

	parsed, err := parseUUIDPathParamAfter(req, "experiments")

	require.NoError(t, err)
	require.Equal(t, experimentID, parsed)
}

func TestParseUUIDPathParamAfter_ParsesNestedUserPath(t *testing.T) {
	userID := uuid.New()
	req := httptest.NewRequest("GET", "/v1/bandit/users/"+userID.String()+"/pending", nil)

	parsed, err := parseUUIDPathParamAfter(req, "users")

	require.NoError(t, err)
	require.Equal(t, userID, parsed)
}

func TestParseUUIDPathParamAfter_ReturnsErrorForMissingSegment(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/bandit/metrics", nil)

	_, err := parseUUIDPathParamAfter(req, "experiments")

	require.Error(t, err)
}

func TestStatusForServiceError_ReturnsNotFoundForNotFoundErrors(t *testing.T) {
	status := statusForServiceError(assertAnError("experiment not found"), http.StatusBadRequest)

	require.Equal(t, http.StatusNotFound, status)
}

func TestStatusForServiceError_PreservesDefaultForOtherErrors(t *testing.T) {
	status := statusForServiceError(assertAnError("validation failed"), http.StatusBadRequest)

	require.Equal(t, http.StatusBadRequest, status)
}

func TestGetObjectiveScores_GinWrappedRouteAcceptsValidExperimentID(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	experimentID := uuid.New()
	armID := uuid.New()

	repo := &routerPathTestRepo{
		experimentID: experimentID,
		config: &service.ExperimentConfig{
			ID:            experimentID,
			ObjectiveType: service.ObjectiveHybrid,
			ObjectiveWeights: map[string]float64{
				"conversion": 0.5,
				"ltv":        0.3,
				"revenue":    0.2,
			},
		},
		arms: []service.Arm{{
			ID:            armID,
			ExperimentID:  experimentID,
			Name:          "Control",
			IsControl:     true,
			TrafficWeight: 1,
		}},
		statsByArmID: map[uuid.UUID]*service.ArmStats{
			armID: {
				ArmID:       armID,
				Alpha:       12,
				Beta:        5,
				Samples:     17,
				Conversions: 11,
				Revenue:     123.45,
				AvgReward:   7.26,
				UpdatedAt:   time.Now().UTC(),
			},
		},
	}
	cache := &routerPathTestCache{}
	base := service.NewThompsonSamplingBandit(repo, cache, zap.NewNop())
	engine := service.NewAdvancedBanditEngine(base, repo, cache, nil, nil, zap.NewNop(), &service.EngineConfig{EnableHybrid: true})
	handler := NewBanditAdvancedHandler(engine, nil, zap.NewNop())

	router := gin.New()
	v1 := router.Group("/v1")
	bandit := v1.Group("/bandit")
	bandit.GET("/experiments/:id/objectives", gin.WrapF(handler.GetObjectiveScores))

	req := httptest.NewRequest(http.MethodGet, "/v1/bandit/experiments/"+experimentID.String()+"/objectives", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code, "body=%s", res.Body.String())

	var body map[string]map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body), "body=%s", res.Body.String())
	require.Contains(t, body, armID.String(), "expected arm scores in response body=%s", res.Body.String())
	require.Contains(t, body[armID.String()], string(service.ObjectiveHybrid), "body=%s", res.Body.String())
}

type assertAnError string

func (e assertAnError) Error() string { return string(e) }
