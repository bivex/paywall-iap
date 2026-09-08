package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// banditStrategyGroup groups the pluggable reward / selection strategies together.
// Grouping reduces the fan-out of BanditRewardEngine below the coupling threshold.
type banditStrategyGroup struct {
	rewardStrategy    RewardStrategy
	selectionStrategy SelectionStrategy
	delayedStrategy   *DelayedRewardStrategy
}

// banditSubEngines groups references to companion sub-engines.
type banditSubEngines struct {
	windowEngine    *BanditWindowEngine
	objectiveEngine *BanditObjectiveEngine
}

func (g *banditStrategyGroup) getDelayedStrategy(repo BanditRepository, cache BanditCache, logger *zap.Logger, enabled bool) (*DelayedRewardStrategy, error) {
	if !enabled {
		return nil, fmt.Errorf("delayed feedback not enabled")
	}
	if g.delayedStrategy != nil {
		return g.delayedStrategy, nil
	}
	g.delayedStrategy = NewDelayedRewardStrategy(repo, cache, logger)
	return g.delayedStrategy, nil
}

// BanditSelectionEngine manages arm selection strategies
type BanditSelectionEngine struct {
	base          *ThompsonSamplingBandit
	repo          BanditRepository
	cache         BanditCache
	logger        *zap.Logger
	strategies    banditStrategyGroup
	enableDelayed bool
}

// BanditRewardExecutionEngine manages reward processing, normalization, and conversion
type BanditRewardExecutionEngine struct {
	base             *ThompsonSamplingBandit
	repo             BanditRepository
	currencyService  *CurrencyRateService
	logger           *zap.Logger
	strategies       banditStrategyGroup
	subEngines       banditSubEngines
	enableCurrency   bool
	enableContextual bool
	enableDelayed    bool
}

// BanditMetricsEngine manages metrics collection and stats querying
type BanditMetricsEngine struct {
	base          *ThompsonSamplingBandit
	repo          BanditRepository
	cache         BanditCache
	logger        *zap.Logger
	strategies    banditStrategyGroup
	subEngines    banditSubEngines
	enableDelayed bool
}

// BanditRewardEngine manages arm selection, reward recording, and metrics
type BanditRewardEngine struct {
	*BanditSelectionEngine
	*BanditRewardExecutionEngine
	*BanditMetricsEngine
}

// BanditWindowEngine manages sliding window operations for bandit experiments
type BanditWindowEngine struct {
	repo         BanditRepository
	redisClient  *redis.Client
	logger       *zap.Logger
	enableWindow bool
}

// BanditObjectiveEngine manages multi-objective optimization strategies and stats
type BanditObjectiveEngine struct {
	repo           BanditRepository
	cache          BanditCache
	logger         *zap.Logger
	base           *ThompsonSamplingBandit
	hybridStrategy *HybridObjectiveStrategy
	enableHybrid   bool
}

// BanditMaintenanceEngine performs periodic maintenance and retention cleanup
type BanditMaintenanceEngine struct {
	repo            BanditRepository
	logger          *zap.Logger
	currencyService *CurrencyRateService
	base            *ThompsonSamplingBandit
	rewardEngine    *BanditRewardEngine
	windowEngine    *BanditWindowEngine
	objectiveEngine *BanditObjectiveEngine
}

// AdvancedBanditEngine orchestrates all advanced bandit features
// It composes multiple specialized sub-engines and provides a unified interface
type AdvancedBanditEngine struct {
	*BanditRewardEngine
	*BanditWindowEngine
	*BanditObjectiveEngine
	*BanditMaintenanceEngine
}

const (
	defaultBanditMaintenanceBatchSize       = 100
	defaultBanditMaintenanceScanLimit       = 100
	defaultBanditContextRetentionWindow     = 90 * 24 * time.Hour
	defaultBanditExpiredAssignmentRetention = 24 * time.Hour
)

type banditMaintenanceRepository interface {
	ListWindowMaintenanceExperimentIDs(ctx context.Context, limit int) ([]uuid.UUID, error)
	ListObjectiveSyncExperimentIDs(ctx context.Context, limit int) ([]uuid.UUID, error)
	CleanupStaleUserContext(ctx context.Context, olderThan time.Duration) (int64, error)
	CleanupExpiredAssignments(ctx context.Context, olderThan time.Duration) (int64, error)
}

// BanditMaintenanceSummary reports what periodic maintenance actually did.
type BanditMaintenanceSummary struct {
	ExpiredPendingRewards       int   `json:"expired_pending_rewards"`
	CurrencyRatesUpdated        bool  `json:"currency_rates_updated"`
	WindowExperimentsScanned    int   `json:"window_experiments_scanned"`
	WindowsTrimmed              int   `json:"windows_trimmed"`
	ObjectiveExperimentsScanned int   `json:"objective_experiments_scanned"`
	ObjectiveStatsSynced        int   `json:"objective_stats_synced"`
	StaleContextsDeleted        int64 `json:"stale_contexts_deleted"`
	ExpiredAssignmentsDeleted   int64 `json:"expired_assignments_deleted"`
}

// EngineConfig configures the advanced bandit engine
type EngineConfig struct {
	ExperimentConfig *ExperimentConfig
	EnableCurrency   bool
	EnableContextual bool
	EnableDelayed    bool
	EnableWindow     bool
	EnableHybrid     bool
}

func fetchExperimentConfig(ctx context.Context, repo BanditRepository, experimentID uuid.UUID) (*ExperimentConfig, error) {
	config, err := repo.GetExperimentConfig(ctx, experimentID)
	if err != nil || config == nil {
		return &ExperimentConfig{ID: experimentID, ObjectiveType: ObjectiveConversion}, nil
	}
	if config.ID == uuid.Nil {
		config.ID = experimentID
	}
	if config.ObjectiveType == "" {
		config.ObjectiveType = ObjectiveConversion
	}
	return config, nil
}

// AdvancedBanditEngineParams encapsulates dependencies and configuration for AdvancedBanditEngine
type AdvancedBanditEngineParams struct {
	Base            *ThompsonSamplingBandit
	Repo            BanditRepository
	Cache           BanditCache
	RedisClient     *redis.Client
	CurrencyService *CurrencyRateService
	Logger          *zap.Logger
	Config          *EngineConfig
}

// NewAdvancedBanditEngine creates a new advanced bandit engine composed of specialized sub-engines
func NewAdvancedBanditEngine(params AdvancedBanditEngineParams) *AdvancedBanditEngine {
	base := params.Base
	repo := params.Repo
	cache := params.Cache
	redisClient := params.RedisClient
	currencyService := params.CurrencyService
	logger := params.Logger
	config := params.Config

	var enableCurrency, enableContextual, enableDelayed, enableWindow, enableHybrid bool
	if config != nil {
		enableCurrency = config.EnableCurrency
		enableContextual = config.EnableContextual
		enableDelayed = config.EnableDelayed
		enableWindow = config.EnableWindow
		enableHybrid = config.EnableHybrid
	}

	var hybridStrategy *HybridObjectiveStrategy
	var selectionStrategy SelectionStrategy
	var delayedStrategy *DelayedRewardStrategy
	var rewardStrategy RewardStrategy

	if config != nil && config.ExperimentConfig != nil {
		if config.EnableHybrid {
			hybridStrategy = NewHybridObjectiveStrategy(HybridObjectiveStrategyConfig{
				Repo:       repo,
				Cache:      cache,
				Logger:     logger,
				Config:     config.ExperimentConfig,
				BaseBandit: base,
			})
		}
		if config.EnableContextual && config.ExperimentConfig.EnableContextual {
			alpha := config.ExperimentConfig.ExplorationAlpha
			selectionStrategy = NewLinUCBSelectionStrategy(LinUCBSelectionStrategyConfig{
				Repo:      repo,
				Cache:     cache,
				Logger:    logger,
				Alpha:     alpha,
				Dimension: 20,
			})
		}
		if config.EnableDelayed && config.ExperimentConfig.EnableDelayed {
			delayedStrategy = NewDelayedRewardStrategy(repo, cache, logger)
		}
		if config.EnableCurrency && config.ExperimentConfig.EnableCurrency && currencyService != nil {
			rewardStrategy = NewCurrencyConversionRewardStrategy(nil, currencyService, logger)
		}
	}

	windowEngine := &BanditWindowEngine{
		repo:         repo,
		redisClient:  redisClient,
		logger:       logger,
		enableWindow: enableWindow,
	}

	objectiveEngine := &BanditObjectiveEngine{
		repo:           repo,
		cache:          cache,
		logger:         logger,
		base:           base,
		hybridStrategy: hybridStrategy,
		enableHybrid:   enableHybrid,
	}

	strategies := banditStrategyGroup{
		rewardStrategy:    rewardStrategy,
		selectionStrategy: selectionStrategy,
		delayedStrategy:   delayedStrategy,
	}
	subEngines := banditSubEngines{
		windowEngine:    windowEngine,
		objectiveEngine: objectiveEngine,
	}

	selectionEngine := &BanditSelectionEngine{
		base:          base,
		repo:          repo,
		cache:         cache,
		logger:        logger,
		strategies:    strategies,
		enableDelayed: enableDelayed,
	}

	rewardExecutionEngine := &BanditRewardExecutionEngine{
		base:             base,
		repo:             repo,
		currencyService:  currencyService,
		logger:           logger,
		strategies:       strategies,
		subEngines:       subEngines,
		enableCurrency:   enableCurrency,
		enableContextual: enableContextual,
		enableDelayed:    enableDelayed,
	}

	metricsEngine := &BanditMetricsEngine{
		base:          base,
		repo:          repo,
		cache:         cache,
		logger:        logger,
		strategies:    strategies,
		subEngines:    subEngines,
		enableDelayed: enableDelayed,
	}

	rewardEngine := &BanditRewardEngine{
		BanditSelectionEngine:       selectionEngine,
		BanditRewardExecutionEngine: rewardExecutionEngine,
		BanditMetricsEngine:         metricsEngine,
	}

	maintenanceEngine := &BanditMaintenanceEngine{
		repo:            repo,
		logger:          logger,
		currencyService: currencyService,
		base:            base,
		rewardEngine:    rewardEngine,
		windowEngine:    windowEngine,
		objectiveEngine: objectiveEngine,
	}

	return &AdvancedBanditEngine{
		BanditRewardEngine:      rewardEngine,
		BanditWindowEngine:      windowEngine,
		BanditObjectiveEngine:   objectiveEngine,
		BanditMaintenanceEngine: maintenanceEngine,
	}
}

func (e *BanditSelectionEngine) getExperimentConfig(
	ctx context.Context,
	experimentID uuid.UUID,
) (*ExperimentConfig, error) {
	return fetchExperimentConfig(ctx, e.repo, experimentID)
}

func (e *BanditSelectionEngine) getDelayedStrategy() (*DelayedRewardStrategy, error) {
	return e.strategies.getDelayedStrategy(e.repo, e.cache, e.logger, e.enableDelayed)
}

func (e *BanditObjectiveEngine) getHybridStrategy(
	ctx context.Context,
	experimentID uuid.UUID,
) (*HybridObjectiveStrategy, error) {
	if !e.enableHybrid {
		return nil, fmt.Errorf("hybrid objective not enabled")
	}

	config, err := fetchExperimentConfig(ctx, e.repo, experimentID)
	if err != nil {
		return nil, err
	}

	return NewHybridObjectiveStrategy(HybridObjectiveStrategyConfig{
		Repo:       e.repo,
		Cache:      e.cache,
		Logger:     e.logger,
		Config:     config,
		BaseBandit: e.base,
	}), nil
}

func (e *BanditWindowEngine) getWindowStrategy(
	ctx context.Context,
	experimentID uuid.UUID,
) (*SlidingWindowStrategy, error) {
	if !e.enableWindow {
		return nil, fmt.Errorf("sliding window not enabled")
	}
	if e.redisClient == nil {
		return nil, fmt.Errorf("sliding window requires redis")
	}

	config, err := fetchExperimentConfig(ctx, e.repo, experimentID)
	if err != nil {
		return nil, err
	}

	return NewSlidingWindowStrategy(SlidingWindowStrategyConfig{
		Repo:         e.repo,
		RedisClient:  e.redisClient,
		Logger:       e.logger,
		ExperimentID: experimentID,
		Config:       config.WindowConfig,
	}), nil
}

// SelectArm selects an arm using the configured strategies
func (e *BanditSelectionEngine) SelectArm(
	ctx context.Context,
	experimentID, userID uuid.UUID,
	userContext UserContext,
) (uuid.UUID, error) {
	arms, err := e.repo.GetArms(ctx, experimentID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get arms: %w", err)
	}

	selectedArm, err := e.resolveSelectedArm(ctx, resolveArmParams{
		experimentID: experimentID,
		userID:       userID,
		arms:         arms,
		userContext:  userContext,
	})
	if err != nil {
		return uuid.Nil, err
	}

	e.recordPendingRewardIfDelayed(ctx, experimentID, selectedArm.ID, userID)

	return selectedArm.ID, nil
}

type resolveArmParams struct {
	experimentID uuid.UUID
	userID       uuid.UUID
	arms         []Arm
	userContext  UserContext
}

func (e *BanditSelectionEngine) resolveSelectedArm(
	ctx context.Context,
	p resolveArmParams,
) (*Arm, error) {
	if e.strategies.selectionStrategy != nil {
		arm, err := e.strategies.selectionStrategy.SelectArm(ctx, p.arms, p.userContext)
		if err != nil {
			e.logger.Warn("Selection strategy failed, falling back to base", zap.Error(err))
		} else if arm != nil {
			return arm, nil
		}
	}

	armID, err := e.base.SelectArm(ctx, p.experimentID, p.userID)
	if err != nil {
		return nil, err
	}

	for _, arm := range p.arms {
		if arm.ID == armID {
			return &arm, nil
		}
	}

	return nil, fmt.Errorf("failed to select arm")
}

func (e *BanditSelectionEngine) recordPendingRewardIfDelayed(ctx context.Context, experimentID, armID, userID uuid.UUID) {
	delayedStrategy, err := e.getDelayedStrategy()
	if err != nil {
		return
	}
	if _, err := delayedStrategy.RecordPendingReward(ctx, experimentID, armID, userID); err != nil {
		e.logger.Warn("Failed to record pending reward", zap.Error(err))
	}
}

// RecordRewardParams encapsulates parameters for recording a reward
type RecordRewardParams struct {
	ExperimentID uuid.UUID
	ArmID        uuid.UUID
	UserID       uuid.UUID
	Reward       float64
	Currency     string
	UserContext  UserContext
}

type baseRewardEventParams struct {
	experimentID     uuid.UUID
	armID            uuid.UUID
	userID           uuid.UUID
	originalReward   float64
	originalCurrency string
	finalReward      float64
	finalCurrency    string
	recordedAt       time.Time
}

type windowRewardEventParams struct {
	experimentID uuid.UUID
	armID        uuid.UUID
	userID       uuid.UUID
	reward       float64
	currency     string
	recordedAt   time.Time
}

// ConversionRewardParams encapsulates parameters for processing delayed conversion
type ConversionRewardParams struct {
	TransactionID   uuid.UUID
	UserID          uuid.UUID
	ConversionValue float64
	Currency        string
}

// RecordReward records a reward with all applicable strategies
func (e *BanditRewardExecutionEngine) RecordReward(
	ctx context.Context,
	p RecordRewardParams,
) error {
	finalReward, finalCurrency := e.normalizeRewardCurrency(ctx, p.ExperimentID, p.Reward, p.Currency)
	recordedAt := time.Now().UTC()

	if err := e.recordBaseRewardEvent(ctx, baseRewardEventParams{
		experimentID:     p.ExperimentID,
		armID:            p.ArmID,
		userID:           p.UserID,
		originalReward:   p.Reward,
		originalCurrency: p.Currency,
		finalReward:      finalReward,
		finalCurrency:    finalCurrency,
		recordedAt:       recordedAt,
	}); err != nil {
		return err
	}

	e.updateLinUCBModelIfConfigured(ctx, p.ArmID, p.UserContext, finalReward)
	e.recordWindowEventIfConfigured(ctx, windowRewardEventParams{
		experimentID: p.ExperimentID,
		armID:        p.ArmID,
		userID:       p.UserID,
		reward:       finalReward,
		currency:     finalCurrency,
		recordedAt:   recordedAt,
	})
	e.recordObjectiveRewardIfConfigured(ctx, p.ExperimentID, p.ArmID, finalReward)

	return nil
}

func (e *BanditRewardExecutionEngine) normalizeRewardCurrency(ctx context.Context, experimentID uuid.UUID, reward float64, currency string) (float64, string) {
	if currency == "" || currency == "USD" || e.currencyService == nil || !e.enableCurrency {
		return reward, currency
	}
	config, err := fetchExperimentConfig(ctx, e.repo, experimentID)
	if err != nil || (config != nil && !config.EnableCurrency) {
		return reward, currency
	}
	converted, err := e.currencyService.ConvertToUSD(ctx, reward, currency)
	if err != nil {
		e.logger.Warn("Currency conversion failed", zap.Error(err))
		return reward, currency
	}
	return converted, "USD"
}

func (e *BanditRewardExecutionEngine) recordBaseRewardEvent(
	ctx context.Context,
	p baseRewardEventParams,
) error {
	if err := e.base.UpdateRewardWithEvent(ctx, RewardWithEventParams{
		ExperimentID: p.experimentID,
		ArmID:        p.armID,
		Reward:       p.finalReward,
		Event: &ConversionEvent{
			ExperimentID:          p.experimentID,
			ArmID:                 p.armID,
			UserID:                &p.userID,
			EventType:             ConversionEventTypeDirectReward,
			OriginalRewardValue:   p.originalReward,
			OriginalCurrency:      p.originalCurrency,
			NormalizedRewardValue: p.finalReward,
			NormalizedCurrency:    p.finalCurrency,
			Metadata: map[string]interface{}{
				"source": "advanced_bandit_engine",
			},
			OccurredAt: p.recordedAt,
		},
	}); err != nil {
		return fmt.Errorf("failed to update base reward: %w", err)
	}
	return nil
}

func (e *BanditRewardExecutionEngine) updateLinUCBModelIfConfigured(ctx context.Context, armID uuid.UUID, userContext UserContext, reward float64) {
	if linucbStrategy, ok := e.strategies.selectionStrategy.(*LinUCBSelectionStrategy); ok {
		if err := linucbStrategy.UpdateModel(ctx, armID, userContext, reward); err != nil {
			e.logger.Warn("Failed to update LinUCB model", zap.Error(err))
		}
	}
}

func (e *BanditRewardExecutionEngine) recordWindowEventIfConfigured(
	ctx context.Context,
	p windowRewardEventParams,
) {
	if e.subEngines.windowEngine == nil {
		return
	}
	windowStrategy, err := e.subEngines.windowEngine.getWindowStrategy(ctx, p.experimentID)
	if err != nil {
		return
	}
	event := RewardEvent{
		UserID:      p.userID,
		ArmID:       p.armID,
		RewardValue: p.reward,
		Currency:    p.currency,
		Timestamp:   p.recordedAt,
	}
	if err := windowStrategy.RecordEvent(ctx, p.armID, event); err != nil {
		e.logger.Warn("Failed to record window event", zap.Error(err))
	}
}

func (e *BanditRewardExecutionEngine) recordObjectiveRewardIfConfigured(ctx context.Context, experimentID, armID uuid.UUID, reward float64) {
	if e.subEngines.objectiveEngine == nil {
		return
	}
	hybridStrategy, err := e.subEngines.objectiveEngine.getHybridStrategy(ctx, experimentID)
	if err != nil {
		return
	}
	config := hybridStrategy.GetConfig()
	if config.ObjectiveType == ObjectiveHybrid {
		for objType := range config.ObjectiveWeights {
			if err := hybridStrategy.RecordObjectiveReward(ctx, ObjectiveRewardParams{
				ArmID:         armID,
				ObjectiveType: ObjectiveType(objType),
				Reward:        reward,
				LTV:           0,
			}); err != nil {
				e.logger.Warn("Failed to record objective reward", zap.String("objective", objType), zap.Error(err))
			}
		}
		return
	}
	if err := hybridStrategy.RecordObjectiveReward(ctx, ObjectiveRewardParams{
		ArmID:         armID,
		ObjectiveType: config.ObjectiveType,
		Reward:        reward,
		LTV:           0,
	}); err != nil {
		e.logger.Warn("Failed to record objective reward", zap.Error(err))
	}
}

// ProcessConversion processes a delayed conversion
func (e *BanditRewardExecutionEngine) ProcessConversion(
	ctx context.Context,
	p ConversionRewardParams,
) error {
	delayedStrategy, err := e.strategies.getDelayedStrategy(e.repo, nil, e.logger, e.enableDelayed)
	if err != nil {
		return err
	}

	// Process through delayed strategy
	if err := delayedStrategy.ProcessConversion(
		ctx,
		DelayedConversionParams{
			TransactionID:   p.TransactionID,
			UserID:          p.UserID,
			ConversionValue: p.ConversionValue,
			Currency:        p.Currency,
			BaseBandit:      e.base,
		},
	); err != nil {
		return err
	}

	return nil
}

// GetArmStatistics returns statistics for all arms in an experiment
func (e *BanditMetricsEngine) GetArmStatistics(
	ctx context.Context,
	experimentID uuid.UUID,
) (map[uuid.UUID]*ArmStats, error) {
	return e.base.GetArmStatistics(ctx, experimentID)
}

// GetObjectiveScores returns objective scores for all arms
func (e *BanditObjectiveEngine) GetObjectiveScores(
	ctx context.Context,
	experimentID uuid.UUID,
) (map[uuid.UUID]map[ObjectiveType]*ObjectiveScore, error) {
	hybridStrategy, err := e.getHybridStrategy(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	arms, err := e.repo.GetArms(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]map[ObjectiveType]*ObjectiveScore)
	for _, arm := range arms {
		scores, err := hybridStrategy.GetObjectiveScores(ctx, arm.ID)
		if err != nil {
			e.logger.Warn("Failed to get objective scores",
				zap.String("arm_id", arm.ID.String()),
				zap.Error(err),
			)
			continue
		}
		result[arm.ID] = scores
	}

	return result, nil
}

// SetObjectiveConfig persists optimization objective settings for an experiment.
func (e *BanditObjectiveEngine) SetObjectiveConfig(
	ctx context.Context,
	experimentID uuid.UUID,
	objectiveType ObjectiveType,
	objectiveWeights map[string]float64,
) (*ExperimentConfig, error) {
	switch objectiveType {
	case ObjectiveConversion, ObjectiveLTV, ObjectiveRevenue, ObjectiveHybrid:
	default:
		return nil, fmt.Errorf("invalid objective type: %s", objectiveType)
	}

	config := &ExperimentConfig{
		ID:               experimentID,
		ObjectiveType:    objectiveType,
		ObjectiveWeights: objectiveWeights,
	}

	if objectiveType == ObjectiveHybrid {
		hybridStrategy := NewHybridObjectiveStrategy(HybridObjectiveStrategyConfig{
			Repo:       e.repo,
			Cache:      e.cache,
			Logger:     e.logger,
			Config:     config,
			BaseBandit: e.base,
		})
		if err := hybridStrategy.ValidateWeights(config.ObjectiveWeights); err != nil {
			return nil, err
		}
		config.ObjectiveWeights = hybridStrategy.NormalizeWeights(config.ObjectiveWeights)
	} else {
		config.ObjectiveWeights = nil
	}

	if err := e.repo.UpdateObjectiveConfig(ctx, experimentID, config.ObjectiveType, config.ObjectiveWeights); err != nil {
		return nil, err
	}

	return config, nil
}

// GetMetrics returns production metrics for the engine
func (e *BanditMetricsEngine) GetMetrics(ctx context.Context, experimentID uuid.UUID) (*BanditMetrics, error) {
	stats, err := e.GetArmStatistics(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	metrics := &BanditMetrics{
		BalanceIndex: calculateBalanceIndex(stats),
	}

	e.populatePendingRewardsMetric(ctx, metrics)
	e.populateWindowUtilizationMetric(ctx, experimentID, metrics)

	return metrics, nil
}

func (e *BanditMetricsEngine) populatePendingRewardsMetric(ctx context.Context, metrics *BanditMetrics) {
	delayedStrategy, err := e.strategies.getDelayedStrategy(e.repo, e.cache, e.logger, e.enableDelayed)
	if err != nil {
		return
	}
	stats, err := delayedStrategy.GetStats(ctx)
	if err != nil {
		return
	}
	if expired, ok := stats["expired_unprocessed"].(int); ok {
		metrics.PendingRewards = int64(expired)
	}
}

func (e *BanditMetricsEngine) populateWindowUtilizationMetric(ctx context.Context, experimentID uuid.UUID, metrics *BanditMetrics) {
	if e.subEngines.windowEngine == nil {
		return
	}
	windowStrategy, err := e.subEngines.windowEngine.getWindowStrategy(ctx, experimentID)
	if err != nil {
		return
	}
	arms, err := e.repo.GetArms(ctx, experimentID)
	if err != nil || len(arms) == 0 {
		return
	}
	totalUtilization := 0.0
	count := 0.0
	for _, arm := range arms {
		utilization, utilErr := windowStrategy.GetUtilization(ctx, arm.ID)
		if utilErr != nil {
			continue
		}
		totalUtilization += utilization
		count++
	}
	if count > 0 {
		metrics.WindowUtilization = totalUtilization / count
	}
}

func (e *BanditWindowEngine) GetWindowInfo(
	ctx context.Context,
	experimentID uuid.UUID,
) (map[uuid.UUID]*WindowStats, error) {
	windowStrategy, err := e.getWindowStrategy(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	arms, err := e.repo.GetArms(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]*WindowStats, len(arms))
	for _, arm := range arms {
		info, infoErr := windowStrategy.GetWindowInfo(ctx, arm.ID)
		if infoErr != nil {
			return nil, infoErr
		}
		result[arm.ID] = info
	}

	return result, nil
}

func (e *BanditWindowEngine) TrimWindow(ctx context.Context, experimentID uuid.UUID) error {
	windowStrategy, err := e.getWindowStrategy(ctx, experimentID)
	if err != nil {
		return err
	}

	arms, err := e.repo.GetArms(ctx, experimentID)
	if err != nil {
		return err
	}

	for _, arm := range arms {
		if trimErr := windowStrategy.TrimWindow(ctx, arm.ID); trimErr != nil {
			return trimErr
		}
	}

	return nil
}

func (e *BanditMaintenanceEngine) ProcessExpiredPendingRewards(ctx context.Context, batchSize int) (int, error) {
	if batchSize <= 0 {
		batchSize = defaultBanditMaintenanceBatchSize
	}

	if e.rewardEngine == nil {
		return 0, nil
	}
	delayedStrategy, err := e.rewardEngine.getDelayedStrategy()
	if err != nil {
		return 0, nil
	}

	processed, err := delayedStrategy.ProcessExpiredRewards(ctx, e.base, batchSize)
	if err != nil {
		return 0, err
	}
	if processed > 0 {
		e.logger.Info("Processed expired pending rewards", zap.Int("count", processed))
	}
	return processed, nil
}

func (e *BanditWindowEngine) TrimConfiguredWindows(ctx context.Context, limit int) (int, error) {
	if !e.enableWindow {
		return 0, nil
	}
	if limit <= 0 {
		limit = defaultBanditMaintenanceScanLimit
	}

	maintenanceRepo, ok := e.repo.(banditMaintenanceRepository)
	if !ok {
		return 0, nil
	}

	experimentIDs, err := maintenanceRepo.ListWindowMaintenanceExperimentIDs(ctx, limit)
	if err != nil {
		return 0, fmt.Errorf("failed to list window maintenance experiments: %w", err)
	}

	return e.trimWindowsForExperiments(ctx, experimentIDs)
}

func (e *BanditWindowEngine) trimWindowsForExperiments(ctx context.Context, experimentIDs []uuid.UUID) (int, error) {
	trimmed := 0
	for _, experimentID := range experimentIDs {
		arms, err := e.repo.GetArms(ctx, experimentID)
		if err != nil {
			return trimmed, fmt.Errorf("failed to load arms for window maintenance: %w", err)
		}
		if len(arms) == 0 {
			continue
		}

		if err := e.TrimWindow(ctx, experimentID); err != nil {
			return trimmed, err
		}
		trimmed += len(arms)
	}
	return trimmed, nil
}

func (e *BanditObjectiveEngine) SyncObjectiveStats(ctx context.Context, limit int) (int, error) {
	if limit <= 0 {
		limit = defaultBanditMaintenanceScanLimit
	}

	maintenanceRepo, ok := e.repo.(banditMaintenanceRepository)
	if !ok {
		return 0, nil
	}
	objectiveRepo, ok := e.repo.(ObjectiveRepository)
	if !ok {
		return 0, nil
	}

	experimentIDs, err := maintenanceRepo.ListObjectiveSyncExperimentIDs(ctx, limit)
	if err != nil {
		return 0, fmt.Errorf("failed to list objective sync experiments: %w", err)
	}

	synced := 0
	for _, experimentID := range experimentIDs {
		n, err := e.syncExperimentObjectiveStats(ctx, experimentID, objectiveRepo)
		synced += n
		if err != nil {
			return synced, err
		}
	}

	return synced, nil
}

func (e *BanditObjectiveEngine) syncExperimentObjectiveStats(
	ctx context.Context,
	experimentID uuid.UUID,
	objectiveRepo ObjectiveRepository,
) (int, error) {
	config, err := fetchExperimentConfig(ctx, e.repo, experimentID)
	if err != nil {
		return 0, err
	}
	objectiveTypes := maintenanceObjectiveTypes(config)
	if len(objectiveTypes) == 0 {
		return 0, nil
	}

	arms, err := e.repo.GetArms(ctx, experimentID)
	if err != nil {
		return 0, fmt.Errorf("failed to load arms for objective sync: %w", err)
	}

	synced := 0
	for _, arm := range arms {
		armSynced, err := e.syncArmObjectiveStats(ctx, arm.ID, objectiveTypes, objectiveRepo)
		synced += armSynced
		if err != nil {
			return synced, err
		}
	}
	return synced, nil
}

func (e *BanditObjectiveEngine) syncArmObjectiveStats(
	ctx context.Context,
	armID uuid.UUID,
	objectiveTypes []ObjectiveType,
	objectiveRepo ObjectiveRepository,
) (int, error) {
	stats, err := e.repo.GetArmStats(ctx, armID)
	if err != nil {
		return 0, fmt.Errorf("failed to load arm stats for objective sync: %w", err)
	}

	synced := 0
	for _, objectiveType := range objectiveTypes {
		if err := objectiveRepo.UpdateObjectiveStats(ctx, &ArmObjectiveStats{
			ArmID:         armID,
			ObjectiveType: objectiveType,
			Alpha:         stats.Alpha,
			Beta:          stats.Beta,
			Samples:       stats.Samples,
			Conversions:   stats.Conversions,
			TotalRevenue:  stats.Revenue,
			AvgLTV:        stats.AvgReward,
		}); err != nil {
			return synced, fmt.Errorf("failed to sync objective stats: %w", err)
		}
		synced++
	}
	return synced, nil
}

func (e *BanditMaintenanceEngine) CleanupOldContextData(ctx context.Context, olderThan time.Duration) (int64, error) {
	if olderThan <= 0 {
		olderThan = defaultBanditContextRetentionWindow
	}

	maintenanceRepo, ok := e.repo.(banditMaintenanceRepository)
	if !ok {
		return 0, nil
	}

	return maintenanceRepo.CleanupStaleUserContext(ctx, olderThan)
}

func (e *BanditMaintenanceEngine) CleanupExpiredAssignments(ctx context.Context, olderThan time.Duration) (int64, error) {
	if olderThan <= 0 {
		olderThan = defaultBanditExpiredAssignmentRetention
	}

	maintenanceRepo, ok := e.repo.(banditMaintenanceRepository)
	if !ok {
		return 0, nil
	}

	return maintenanceRepo.CleanupExpiredAssignments(ctx, olderThan)
}

func (e *BanditWindowEngine) ExportWindowEvents(
	ctx context.Context,
	experimentID uuid.UUID,
	limit int64,
) (map[uuid.UUID][]RewardEvent, error) {
	windowStrategy, err := e.getWindowStrategy(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	arms, err := e.repo.GetArms(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID][]RewardEvent, len(arms))
	for _, arm := range arms {
		events, exportErr := windowStrategy.ExportEvents(ctx, arm.ID, limit)
		if exportErr != nil {
			return nil, exportErr
		}
		result[arm.ID] = events
	}

	return result, nil
}

func (e *BanditObjectiveEngine) GetObjectiveConfig(ctx context.Context, experimentID uuid.UUID) (*ExperimentConfig, error) {
	return fetchExperimentConfig(ctx, e.repo, experimentID)
}

func (e *BanditMetricsEngine) GetPendingReward(ctx context.Context, pendingID uuid.UUID) (*PendingReward, error) {
	delayedStrategy, err := e.strategies.getDelayedStrategy(e.repo, e.cache, e.logger, e.enableDelayed)
	if err != nil {
		return nil, err
	}

	return delayedStrategy.GetPendingReward(ctx, pendingID)
}

func (e *BanditMetricsEngine) GetUserPendingRewards(
	ctx context.Context,
	userID uuid.UUID,
) ([]*PendingReward, error) {
	delayedStrategy, err := e.strategies.getDelayedStrategy(e.repo, e.cache, e.logger, e.enableDelayed)
	if err != nil {
		return nil, err
	}

	return delayedStrategy.GetPendingRewardsByUser(ctx, userID, uuid.Nil)
}

// calculateBalanceIndex measures how evenly users are distributed
// Returns 1.0 for perfect balance, 0.0 for all users in one arm
func calculateBalanceIndex(stats map[uuid.UUID]*ArmStats) float64 {
	if len(stats) == 0 {
		return 0
	}

	// Calculate total samples
	totalSamples := 0
	for _, stat := range stats {
		totalSamples += stat.Samples
	}

	if totalSamples == 0 {
		return 1.0 // Perfect balance when no samples
	}

	// Calculate expected samples per arm
	expected := float64(totalSamples) / float64(len(stats))

	// Calculate deviation from expected
	totalDeviation := 0.0
	for _, stat := range stats {
		deviation := mathAbs(float64(stat.Samples) - expected)
		totalDeviation += deviation
	}

	// Normalize to [0, 1]
	// Maximum deviation is (totalSamples - expected) = totalSamples * (n-1) / n
	maxDeviation := float64(totalSamples) * float64(len(stats)-1) / float64(len(stats))
	balanceIndex := 1.0 - (totalDeviation / (maxDeviation * float64(len(stats))))

	return balanceIndex
}

// BanditMetrics represents production metrics for monitoring
type BanditMetrics struct {
	Regret            float64
	ExplorationRate   float64
	ConvergenceGap    float64
	BalanceIndex      float64
	WindowUtilization float64
	PendingRewards    int64
}

// RunMaintenanceDetailed performs periodic maintenance tasks and returns a structured summary.
func (e *BanditMaintenanceEngine) RunMaintenanceDetailed(ctx context.Context) (*BanditMaintenanceSummary, error) {
	summary := &BanditMaintenanceSummary{}

	processed, err := e.ProcessExpiredPendingRewards(ctx, defaultBanditMaintenanceBatchSize)
	if err != nil {
		return summary, err
	}
	summary.ExpiredPendingRewards = processed

	if err := e.runCurrencyMaintenance(ctx, summary); err != nil {
		return summary, err
	}

	if err := e.runWindowAndObjectiveInspection(ctx, summary); err != nil {
		return summary, err
	}

	if e.objectiveEngine != nil {
		synced, err := e.objectiveEngine.SyncObjectiveStats(ctx, defaultBanditMaintenanceScanLimit)
		if err != nil {
			return summary, err
		}
		summary.ObjectiveStatsSynced = synced
	}

	contextsDeleted, err := e.CleanupOldContextData(ctx, defaultBanditContextRetentionWindow)
	if err != nil {
		return summary, err
	}
	summary.StaleContextsDeleted = contextsDeleted

	assignmentsDeleted, err := e.CleanupExpiredAssignments(ctx, defaultBanditExpiredAssignmentRetention)
	if err != nil {
		return summary, err
	}
	summary.ExpiredAssignmentsDeleted = assignmentsDeleted

	return summary, nil
}

func (e *BanditMaintenanceEngine) runCurrencyMaintenance(ctx context.Context, summary *BanditMaintenanceSummary) error {
	if e.currencyService == nil {
		return nil
	}
	if err := e.currencyService.UpdateRates(ctx); err != nil {
		return err
	}
	summary.CurrencyRatesUpdated = true
	return nil
}

func (e *BanditMaintenanceEngine) runWindowAndObjectiveInspection(ctx context.Context, summary *BanditMaintenanceSummary) error {
	maintenanceRepo, ok := e.repo.(banditMaintenanceRepository)
	if !ok {
		return nil
	}

	windowExperimentIDs, err := maintenanceRepo.ListWindowMaintenanceExperimentIDs(ctx, defaultBanditMaintenanceScanLimit)
	if err != nil {
		return fmt.Errorf("failed to inspect window maintenance candidates: %w", err)
	}
	summary.WindowExperimentsScanned = len(windowExperimentIDs)

	if e.windowEngine != nil {
		trimmed, err := e.windowEngine.TrimConfiguredWindows(ctx, defaultBanditMaintenanceScanLimit)
		if err != nil {
			return err
		}
		summary.WindowsTrimmed = trimmed
	}

	objectiveExperimentIDs, err := maintenanceRepo.ListObjectiveSyncExperimentIDs(ctx, defaultBanditMaintenanceScanLimit)
	if err != nil {
		return fmt.Errorf("failed to inspect objective sync candidates: %w", err)
	}
	summary.ObjectiveExperimentsScanned = len(objectiveExperimentIDs)
	return nil
}

// RunMaintenance performs periodic maintenance tasks.
func (e *BanditMaintenanceEngine) RunMaintenance(ctx context.Context) error {
	_, err := e.RunMaintenanceDetailed(ctx)
	return err
}

func maintenanceObjectiveTypes(config *ExperimentConfig) []ObjectiveType {
	if config == nil {
		return nil
	}

	if config.ObjectiveType != ObjectiveHybrid {
		switch config.ObjectiveType {
		case ObjectiveConversion, ObjectiveLTV, ObjectiveRevenue:
			return []ObjectiveType{config.ObjectiveType}
		default:
			return nil
		}
	}

	valid := map[ObjectiveType]struct{}{
		ObjectiveConversion: {},
		ObjectiveLTV:        {},
		ObjectiveRevenue:    {},
	}
	objectiveTypes := make([]ObjectiveType, 0, len(valid))
	for objective := range config.ObjectiveWeights {
		objType := ObjectiveType(objective)
		if _, ok := valid[objType]; ok {
			objectiveTypes = append(objectiveTypes, objType)
		}
	}
	if len(objectiveTypes) == 0 {
		objectiveTypes = []ObjectiveType{ObjectiveConversion, ObjectiveLTV, ObjectiveRevenue}
	}
	sort.Slice(objectiveTypes, func(i, j int) bool { return objectiveTypes[i] < objectiveTypes[j] })
	return objectiveTypes
}

// Helper function for math
func mathAbs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
