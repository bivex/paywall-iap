package handlers

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	domainErrors "github.com/bivex/paywall-iap/internal/domain/errors"
	"github.com/bivex/paywall-iap/internal/domain/service"
)

// BanditAdvancedHandler handles advanced bandit feature HTTP endpoints
type BanditAdvancedHandler struct {
	engine          *service.AdvancedBanditEngine
	currencyService *service.CurrencyRateService
	logger          *zap.Logger
}

// NewBanditAdvancedHandler creates a new advanced bandit handler
func NewBanditAdvancedHandler(
	engine *service.AdvancedBanditEngine,
	currencyService *service.CurrencyRateService,
	logger *zap.Logger,
) *BanditAdvancedHandler {
	return &BanditAdvancedHandler{
		engine:          engine,
		currencyService: currencyService,
		logger:          logger,
	}
}

// RegisterRoutes registers all advanced bandit routes
func (h *BanditAdvancedHandler) RegisterRoutes(router *mux.Router) {
	// Currency management
	router.HandleFunc("/api/bandit/currency/rates", h.GetCurrencyRates).Methods("GET")
	router.HandleFunc("/api/bandit/currency/update", h.UpdateCurrencyRates).Methods("POST")
	router.HandleFunc("/api/bandit/currency/convert", h.ConvertCurrency).Methods("POST")

	// Objective management
	router.HandleFunc("/api/bandit/experiments/{id}/objectives", h.GetObjectiveScores).Methods("GET")
	router.HandleFunc("/api/bandit/experiments/{id}/objectives/config", h.GetObjectiveConfig).Methods("GET")
	router.HandleFunc("/api/bandit/experiments/{id}/objectives/config", h.SetObjectiveConfig).Methods("PUT")

	// Window management
	router.HandleFunc("/api/bandit/experiments/{id}/window/info", h.GetWindowInfo).Methods("GET")
	router.HandleFunc("/api/bandit/experiments/{id}/window/trim", h.TrimWindow).Methods("POST")
	router.HandleFunc("/api/bandit/experiments/{id}/window/events", h.ExportWindowEvents).Methods("GET")

	// Delayed feedback
	router.HandleFunc("/api/bandit/conversions", h.ProcessConversion).Methods("POST")
	router.HandleFunc("/api/bandit/pending/{id}", h.GetPendingReward).Methods("GET")
	router.HandleFunc("/api/bandit/users/{id}/pending", h.GetUserPendingRewards).Methods("GET")

	// Metrics and monitoring
	router.HandleFunc("/api/bandit/experiments/{id}/metrics", h.GetMetrics).Methods("GET")
	router.HandleFunc("/api/bandit/maintenance", h.RunMaintenance).Methods("POST")
}

// GetCurrencyRates returns current currency rates
func (h *BanditAdvancedHandler) GetCurrencyRates(w http.ResponseWriter, r *http.Request) {
	if h.currencyService == nil {
		http.Error(w, "Currency service not available", http.StatusServiceUnavailable)
		return
	}

	supported := h.currencyService.GetSupportedCurrencies()
	rates := make(map[string]float64)

	ctx := r.Context()
	for _, currency := range supported {
		if currency == "USD" {
			rates[currency] = 1.0
			continue
		}
		rate, err := h.currencyService.GetRate(ctx, currency)
		if err == nil {
			rates[currency] = rate
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"base":    "USD",
		"rates":   rates,
		"updated": time.Now(),
	})
}

// UpdateCurrencyRates triggers a currency rate update
func (h *BanditAdvancedHandler) UpdateCurrencyRates(w http.ResponseWriter, r *http.Request) {
	if h.currencyService == nil {
		http.Error(w, "Currency service not available", http.StatusServiceUnavailable)
		return
	}

	if err := h.currencyService.UpdateRates(r.Context()); err != nil {
		h.logger.Error("Failed to update currency rates", zap.Error(err))
		respondError(w, statusForServiceError(err, http.StatusInternalServerError), "Failed to update rates")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Currency rates updated successfully",
		"updated": time.Now(),
	})
}

// ConvertCurrency converts an amount between currencies
func (h *BanditAdvancedHandler) ConvertCurrency(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Amount   *float64 `json:"amount"`
		Currency string   `json:"currency"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Amount == nil || !isISO4217CurrencyCode(req.Currency) {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if h.currencyService == nil {
		http.Error(w, "Currency service not available", http.StatusServiceUnavailable)
		return
	}

	converted, err := h.currencyService.ConvertToUSD(r.Context(), *req.Amount, req.Currency)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !isFiniteJSONNumber(converted) {
		respondError(w, http.StatusBadRequest, "Amount is too large")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"original_amount":   *req.Amount,
		"original_currency": req.Currency,
		"converted_amount":  converted,
		"target_currency":   "USD",
	})
}

// GetObjectiveScores returns objective scores for all arms
func (h *BanditAdvancedHandler) GetObjectiveScores(w http.ResponseWriter, r *http.Request) {
	experimentID, err := parseUUIDPathParamAfter(r, "experiments")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid experiment ID")
		return
	}

	scores, err := h.engine.GetObjectiveScores(r.Context(), experimentID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, scores)
}

// GetObjectiveConfig returns the persisted objective configuration for an experiment.
func (h *BanditAdvancedHandler) GetObjectiveConfig(w http.ResponseWriter, r *http.Request) {
	experimentID, err := parseUUIDPathParamAfter(r, "experiments")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid experiment ID")
		return
	}

	config, err := h.engine.GetObjectiveConfig(r.Context(), experimentID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"experiment_id":  experimentID,
		"objective_type": config.ObjectiveType,
		"weights":        normalizeObjectiveWeights(config.ObjectiveWeights),
	})
}

// SetObjectiveConfig updates the objective configuration for an experiment
func (h *BanditAdvancedHandler) SetObjectiveConfig(w http.ResponseWriter, r *http.Request) {
	experimentID, err := parseUUIDPathParamAfter(r, "experiments")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid experiment ID")
		return
	}

	var req struct {
		ObjectiveType    service.ObjectiveType `json:"objective_type"`
		ObjectiveWeights map[string]float64    `json:"objective_weights"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	config, err := h.engine.SetObjectiveConfig(r.Context(), experimentID, req.ObjectiveType, req.ObjectiveWeights)
	if err != nil {
		respondError(w, statusForServiceError(err, http.StatusBadRequest), err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":        "Configuration updated",
		"experiment_id":  experimentID,
		"objective_type": config.ObjectiveType,
		"weights":        normalizeObjectiveWeights(config.ObjectiveWeights),
	})
}

func normalizeObjectiveWeights(weights map[string]float64) map[string]float64 {
	if weights == nil {
		return map[string]float64{}
	}

	return weights
}

// GetWindowInfo returns window information for an experiment
func (h *BanditAdvancedHandler) GetWindowInfo(w http.ResponseWriter, r *http.Request) {
	experimentID, err := parseUUIDPathParamAfter(r, "experiments")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid experiment ID")
		return
	}

	info, err := h.engine.GetWindowInfo(r.Context(), experimentID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"experiment_id": experimentID,
		"windows":       info,
	})
}

// TrimWindow trims the sliding window for an experiment
func (h *BanditAdvancedHandler) TrimWindow(w http.ResponseWriter, r *http.Request) {
	experimentID, err := parseUUIDPathParamAfter(r, "experiments")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid experiment ID")
		return
	}

	if err := h.engine.TrimWindow(r.Context(), experimentID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"experiment_id": experimentID,
		"message":       "Window trimmed successfully",
	})
}

// ExportWindowEvents exports events from the sliding window
func (h *BanditAdvancedHandler) ExportWindowEvents(w http.ResponseWriter, r *http.Request) {
	experimentID, err := parseUUIDPathParamAfter(r, "experiments")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid experiment ID")
		return
	}

	limit := int64(100)
	if queryValues, hasLimit := r.URL.Query()["limit"]; hasLimit {
		rawLimit := ""
		if len(queryValues) > 0 {
			rawLimit = strings.TrimSpace(queryValues[0])
		}
		if rawLimit == "" {
			respondError(w, http.StatusBadRequest, "Invalid limit")
			return
		}

		parsedLimit, parseErr := strconv.ParseInt(rawLimit, 10, 64)
		if parseErr != nil || parsedLimit <= 0 {
			respondError(w, http.StatusBadRequest, "Invalid limit")
			return
		}
		limit = parsedLimit
	}

	events, err := h.engine.ExportWindowEvents(r.Context(), experimentID, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"experiment_id": experimentID,
		"events":        events,
		"limit":         limit,
	})
}

// ProcessConversion processes a delayed conversion
func (h *BanditAdvancedHandler) ProcessConversion(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TransactionID   uuid.UUID `json:"transaction_id"`
		UserID          uuid.UUID `json:"user_id"`
		ConversionValue *float64  `json:"conversion_value"`
		Currency        string    `json:"currency"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.TransactionID == uuid.Nil || req.UserID == uuid.Nil || req.ConversionValue == nil || !isISO4217CurrencyCode(req.Currency) {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.engine.ProcessConversion(
		r.Context(),
		req.TransactionID,
		req.UserID,
		*req.ConversionValue,
		req.Currency,
	); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":        "Conversion processed successfully",
		"transaction_id": req.TransactionID,
	})
}

// GetPendingReward returns a pending reward by ID
func (h *BanditAdvancedHandler) GetPendingReward(w http.ResponseWriter, r *http.Request) {
	pendingID, err := parseUUIDPathParamAfter(r, "pending")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid pending reward ID")
		return
	}

	pendingReward, err := h.engine.GetPendingReward(r.Context(), pendingID)
	if err != nil {
		respondError(w, statusForServiceError(err, http.StatusInternalServerError), err.Error())
		return
	}

	respondJSON(w, http.StatusOK, pendingReward)
}

// GetUserPendingRewards returns all pending rewards for a user
func (h *BanditAdvancedHandler) GetUserPendingRewards(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUUIDPathParamAfter(r, "users")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	rewards, err := h.engine.GetUserPendingRewards(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rewards == nil {
		rewards = []*service.PendingReward{}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user_id": userID,
		"rewards": rewards,
	})
}

// GetMetrics returns production metrics for an experiment
func (h *BanditAdvancedHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	experimentID, err := parseUUIDPathParamAfter(r, "experiments")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid experiment ID")
		return
	}

	metrics, err := h.engine.GetMetrics(r.Context(), experimentID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, metrics)
}

// RunMaintenance triggers maintenance tasks
func (h *BanditAdvancedHandler) RunMaintenance(w http.ResponseWriter, r *http.Request) {
	if err := h.engine.RunMaintenance(r.Context()); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":   "Maintenance completed successfully",
		"timestamp": time.Now(),
	})
}

// Helper functions

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]interface{}{
		"error": message,
	})
}

func parseUUIDPathParamAfter(r *http.Request, segment string) (uuid.UUID, error) {
	if vars := mux.Vars(r); len(vars) > 0 {
		if raw, ok := vars["id"]; ok && raw != "" {
			return uuid.Parse(raw)
		}
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	for index := 0; index < len(parts)-1; index++ {
		if parts[index] == segment {
			return uuid.Parse(parts[index+1])
		}
	}

	return uuid.Parse("")
}

func statusForServiceError(err error, defaultStatus int) int {
	if err == nil {
		return defaultStatus
	}

	if strings.Contains(strings.ToLower(err.Error()), "not found") {
		return http.StatusNotFound
	}

	if errors.Is(err, domainErrors.ErrExternalServiceUnavailable) {
		return http.StatusServiceUnavailable
	}

	return defaultStatus
}

func isISO4217CurrencyCode(value string) bool {
	if len(value) != 3 {
		return false
	}
	for _, char := range value {
		if char < 'A' || char > 'Z' {
			return false
		}
	}
	return true
}

func isFiniteJSONNumber(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
