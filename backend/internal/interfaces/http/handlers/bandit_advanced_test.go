package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

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
