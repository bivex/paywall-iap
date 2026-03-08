package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	domainErrors "github.com/bivex/paywall-iap/internal/domain/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCurrencyRateService_UpdateRates_ReturnsExternalServiceUnavailableOnNon200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	svc := NewCurrencyRateService(nil, zap.NewNop())
	svc.ecbAPIURL = server.URL
	svc.httpClient = server.Client()

	err := svc.UpdateRates(context.Background())

	require.Error(t, err)
	require.ErrorIs(t, err, domainErrors.ErrExternalServiceUnavailable)
}

func TestCurrencyRateService_UpdateRates_ReturnsExternalServiceUnavailableOnTransportFailure(t *testing.T) {
	svc := NewCurrencyRateService(nil, zap.NewNop())
	svc.ecbAPIURL = "http://127.0.0.1:1"
	svc.httpClient = &http.Client{}

	err := svc.UpdateRates(context.Background())

	require.Error(t, err)
	require.True(t, errors.Is(err, domainErrors.ErrExternalServiceUnavailable))
}
