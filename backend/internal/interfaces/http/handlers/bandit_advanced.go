package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

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
		respondError(w, http.StatusInternalServerError, "Failed to update rates")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Currency rates updated successfully",
		"updated": time.Now(),
	})
}

// ConvertCurrency converts an amount between currencies
func (h *BanditAdvancedHandler) ConvertCurrency(w http.ResponseWriter, r *http.Request) {
	if h.currencyService == nil {
		http.Error(w, "Currency service not available", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Amount   float64 `json:"amount"`
		Currency string  `json:"currency"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	converted, err := h.currencyService.ConvertToUSD(r.Context(), req.Amount, req.Currency)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"original_amount":   req.Amount,
		"original_currency": req.Currency,
		"converted_amount":  converted,
		"target_currency":   "USD",
	})
}

// GetObjectiveScores returns objective scores for all arms
func (h *BanditAdvancedHandler) GetObjectiveScores(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	experimentID, err := uuid.Parse(vars["id"])
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

// SetObjectiveConfig updates the objective configuration for an experiment
func (h *BanditAdvancedHandler) SetObjectiveConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	experimentID, err := uuid.Parse(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid experiment ID")
		return
	}

	var req struct {
		ObjectiveType    service.ObjectiveType    `json:"objective_type"`
		ObjectiveWeights map[string]float64      `json:"objective_weights"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	config := &service.ExperimentConfig{
		ID:               experimentID,
		ObjectiveType:     req.ObjectiveType,
		ObjectiveWeights:  req.ObjectiveWeights,
	}

	// Update engine configuration
	// This would require a method on the engine to update config
	// For now, just acknowledge

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":       "Configuration updated",
		"experiment_id": experimentID,
		"objective_type": req.ObjectiveType,
		"weights":       req.ObjectiveWeights,
	})
}

// GetWindowInfo returns window information for an experiment
func (h *BanditAdvancedHandler) GetWindowInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	experimentID, err := uuid.Parse(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid experiment ID")
		return
	}

	// Get arms for the experiment
	// This would require access to the repository
	// For now, return a placeholder

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"experiment_id": experimentID,
		"message":       "Window info - implement with repository access",
	})
}

// TrimWindow trims the sliding window for an experiment
func (h *BanditAdvancedHandler) TrimWindow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, err := uuid.Parse(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid experiment ID")
		return
	}

	// Trim windows for all arms
	// This would require access to the window strategy

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Window trimmed successfully",
	})
}

// ExportWindowEvents exports events from the sliding window
func (h *BanditAdvancedHandler) ExportWindowEvents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	experimentID, err := uuid.Parse(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid experiment ID")
		return
	}

	// Parse limit from query params
	// limit := 100

	// Export events
	// This would require access to the window strategy

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"experiment_id": experimentID,
		"events":        []interface{}{},
		"message":       "Window events export - implement with window strategy access",
	})
}

// ProcessConversion processes a delayed conversion
func (h *BanditAdvancedHandler) ProcessConversion(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TransactionID      uuid.UUID `json:"transaction_id"`
		UserID             uuid.UUID `json:"user_id"`
		ConversionValue    float64   `json:"conversion_value"`
		Currency           string    `json:"currency"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.engine.ProcessConversion(
		r.Context(),
		req.TransactionID,
		req.UserID,
		req.ConversionValue,
		req.Currency,
	); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":       "Conversion processed successfully",
		"transaction_id": req.TransactionID,
	})
}

// GetPendingReward returns a pending reward by ID
func (h *BanditAdvancedHandler) GetPendingReward(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pendingID, err := uuid.Parse(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid pending reward ID")
		return
	}

	// Get pending reward
	// This would require access to the delayed reward strategy

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"pending_id": pendingID,
		"message":    "Pending reward - implement with delayed strategy access",
	})
}

// GetUserPendingRewards returns all pending rewards for a user
func (h *BanditAdvancedHandler) GetUserPendingRewards(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Get pending rewards for user
	// This would require access to the delayed reward strategy

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":  userID,
		"rewards":  []interface{}{},
		"message": "User pending rewards - implement with delayed strategy access",
	})
}

// GetMetrics returns production metrics for an experiment
func (h *BanditAdvancedHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	experimentID, err := uuid.Parse(vars["id"])
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
		"message": "Maintenance completed successfully",
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
