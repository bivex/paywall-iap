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

// AdvancedBanditEngine orchestrates all advanced bandit features
// It composes multiple strategies and provides a unified interface
type AdvancedBanditEngine struct {
	base              *ThompsonSamplingBandit
	rewardStrategy    RewardStrategy
	selectionStrategy SelectionStrategy
	windowStrategy    WindowStrategy
	delayedStrategy   *DelayedRewardStrategy
	hybridStrategy    *HybridObjectiveStrategy
	currencyService   *CurrencyRateService
	repo              BanditRepository
	cache             BanditCache
	redisClient       *redis.Client
	logger            *zap.Logger
	enableCurrency    bool
	enableContextual  bool
	enableDelayed     bool
	enableWindow      bool
	enableHybrid      bool
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

// NewAdvancedBanditEngine creates a new advanced bandit engine
func NewAdvancedBanditEngine(
	base *ThompsonSamplingBandit,
	repo BanditRepository,
	cache BanditCache,
	redisClient *redis.Client,
	currencyService *CurrencyRateService,
	logger *zap.Logger,
	config *EngineConfig,
) *AdvancedBanditEngine {
	engine := &AdvancedBanditEngine{
		base:            base,
		repo:            repo,
		cache:           cache,
		redisClient:     redisClient,
		currencyService: currencyService,
		logger:          logger,
	}

	if config != nil {
		engine.enableCurrency = config.EnableCurrency
		engine.enableContextual = config.EnableContextual
		engine.enableDelayed = config.EnableDelayed
		engine.enableWindow = config.EnableWindow
		engine.enableHybrid = config.EnableHybrid
	}

	// Configure strategies based on config
	if config != nil && config.ExperimentConfig != nil {
		// Hybrid objective strategy
		if config.EnableHybrid {
			engine.hybridStrategy = NewHybridObjectiveStrategy(
				repo, cache, logger, config.ExperimentConfig, base,
			)
		}

		// Contextual bandit (LinUCB)
		if config.EnableContextual && config.ExperimentConfig.EnableContextual {
			alpha := config.ExperimentConfig.ExplorationAlpha
			engine.selectionStrategy = NewLinUCBSelectionStrategy(
				repo, cache, logger, alpha, 20, // 20 dimension features
			)
		}

		// Sliding window
		if config.EnableWindow && config.ExperimentConfig.WindowConfig != nil {
			// Note: Need Redis client for window strategy
			// engine.windowStrategy = NewSlidingWindowStrategy(...)
		}

		// Delayed feedback
		if config.EnableDelayed && config.ExperimentConfig.EnableDelayed {
			engine.delayedStrategy = NewDelayedRewardStrategy(repo, cache, logger)
		}

		// Currency conversion
		if config.EnableCurrency && config.ExperimentConfig.EnableCurrency {
			if currencyService != nil {
				// Wrap base reward strategy with currency conversion
				engine.rewardStrategy = NewCurrencyConversionRewardStrategy(
					nil, // No base strategy needed for conversion-only
					currencyService,
					logger,
				)
			}
		}
	}

	return engine
}

func (e *AdvancedBanditEngine) getExperimentConfig(
	ctx context.Context,
	experimentID uuid.UUID,
) (*ExperimentConfig, error) {
	config, err := e.repo.GetExperimentConfig(ctx, experimentID)
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

func (e *AdvancedBanditEngine) getDelayedStrategy() (*DelayedRewardStrategy, error) {
	if !e.enableDelayed {
		return nil, fmt.Errorf("delayed feedback not enabled")
	}
	if e.delayedStrategy != nil {
		return e.delayedStrategy, nil
	}

	e.delayedStrategy = NewDelayedRewardStrategy(e.repo, e.cache, e.logger)
	return e.delayedStrategy, nil
}

func (e *AdvancedBanditEngine) getHybridStrategy(
	ctx context.Context,
	experimentID uuid.UUID,
) (*HybridObjectiveStrategy, error) {
	if !e.enableHybrid {
		return nil, fmt.Errorf("hybrid objective not enabled")
	}

	config, err := e.getExperimentConfig(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	return NewHybridObjectiveStrategy(e.repo, e.cache, e.logger, config, e.base), nil
}

func (e *AdvancedBanditEngine) getWindowStrategy(
	ctx context.Context,
	experimentID uuid.UUID,
) (*SlidingWindowStrategy, error) {
	if !e.enableWindow {
		return nil, fmt.Errorf("sliding window not enabled")
	}
	if e.redisClient == nil {
		return nil, fmt.Errorf("sliding window requires redis")
	}

	config, err := e.getExperimentConfig(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	return NewSlidingWindowStrategy(e.repo, e.redisClient, e.logger, experimentID, config.WindowConfig), nil
}

// SelectArm selects an arm using the configured strategies
func (e *AdvancedBanditEngine) SelectArm(
	ctx context.Context,
	experimentID, userID uuid.UUID,
	userContext UserContext,
) (uuid.UUID, error) {
	// Get experiment arms
	arms, err := e.repo.GetArms(ctx, experimentID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get arms: %w", err)
	}

	var selectedArm *Arm

	// Use selection strategy if configured
	if e.selectionStrategy != nil {
		arm, err := e.selectionStrategy.SelectArm(ctx, arms, userContext)
		if err != nil {
			e.logger.Warn("Selection strategy failed, falling back to base", zap.Error(err))
		} else {
			selectedArm = arm
		}
	}

	// Fall back to base Thompson Sampling
	if selectedArm == nil {
		// Use base SelectArm but we need to handle user context
		armID, err := e.base.SelectArm(ctx, experimentID, userID)
		if err != nil {
			return uuid.Nil, err
		}

		// Find the arm object
		for _, arm := range arms {
			if arm.ID == armID {
				selectedArm = &arm
				break
			}
		}
	}

	if selectedArm == nil {
		return uuid.Nil, fmt.Errorf("failed to select arm")
	}

	// Record pending reward if delayed feedback is enabled
	if delayedStrategy, err := e.getDelayedStrategy(); err == nil {
		_, err := delayedStrategy.RecordPendingReward(ctx, experimentID, selectedArm.ID, userID)
		if err != nil {
			e.logger.Warn("Failed to record pending reward", zap.Error(err))
		}
	}

	// Update LinUCB model if contextual is enabled
	if linucbStrategy, ok := e.selectionStrategy.(*LinUCBSelectionStrategy); ok {
		// Model will be updated when reward is recorded
		_ = linucbStrategy
	}

	return selectedArm.ID, nil
}

// RecordReward records a reward with all applicable strategies
func (e *AdvancedBanditEngine) RecordReward(
	ctx context.Context,
	experimentID, armID, userID uuid.UUID,
	reward float64,
	currency string,
	userContext UserContext,
) error {
	// Convert currency if enabled
	finalReward := reward
	finalCurrency := currency

	if currency != "" && currency != "USD" && e.currencyService != nil && e.enableCurrency {
		config, err := e.getExperimentConfig(ctx, experimentID)
		if err == nil && (config == nil || config.EnableCurrency) {
			converted, err := e.currencyService.ConvertToUSD(ctx, reward, currency)
			if err != nil {
				e.logger.Warn("Currency conversion failed", zap.Error(err))
			} else {
				finalReward = converted
				finalCurrency = "USD"
			}
		}
	}

	// Record with base bandit
	if err := e.base.UpdateReward(ctx, experimentID, armID, finalReward); err != nil {
		return fmt.Errorf("failed to update base reward: %w", err)
	}

	// Update LinUCB model if contextual is enabled
	if linucbStrategy, ok := e.selectionStrategy.(*LinUCBSelectionStrategy); ok {
		if err := linucbStrategy.UpdateModel(ctx, armID, userContext, finalReward); err != nil {
			e.logger.Warn("Failed to update LinUCB model", zap.Error(err))
		}
	}

	// Update window strategy if enabled
	if windowStrategy, err := e.getWindowStrategy(ctx, experimentID); err == nil {
		event := RewardEvent{
			UserID:      userID,
			ArmID:       armID,
			RewardValue: finalReward,
			Currency:    finalCurrency,
			Timestamp:   time.Now(),
		}
		if err := windowStrategy.RecordEvent(ctx, armID, event); err != nil {
			e.logger.Warn("Failed to record window event", zap.Error(err))
		}
	}

	// Update hybrid objective stats if enabled
	if hybridStrategy, err := e.getHybridStrategy(ctx, experimentID); err == nil {
		// Determine which objectives to update
		objectiveType := hybridStrategy.GetConfig().ObjectiveType

		if objectiveType == ObjectiveHybrid {
			// Update all objectives
			for objType := range hybridStrategy.GetConfig().ObjectiveWeights {
				if err := hybridStrategy.RecordObjectiveReward(
					ctx, armID, ObjectiveType(objType), finalReward, 0,
				); err != nil {
					e.logger.Warn("Failed to record objective reward",
						zap.String("objective", objType),
						zap.Error(err),
					)
				}
			}
		} else {
			if err := hybridStrategy.RecordObjectiveReward(
				ctx, armID, objectiveType, finalReward, 0,
			); err != nil {
				e.logger.Warn("Failed to record objective reward", zap.Error(err))
			}
		}
	}

	return nil
}

// ProcessConversion processes a delayed conversion
func (e *AdvancedBanditEngine) ProcessConversion(
	ctx context.Context,
	transactionID uuid.UUID,
	userID uuid.UUID,
	conversionValue float64,
	currency string,
) error {
	delayedStrategy, err := e.getDelayedStrategy()
	if err != nil {
		return err
	}

	// Process through delayed strategy
	if err := delayedStrategy.ProcessConversion(
		ctx, transactionID, userID, conversionValue, currency,
	); err != nil {
		return err
	}

	return nil
}

// GetArmStatistics returns statistics for all arms in an experiment
func (e *AdvancedBanditEngine) GetArmStatistics(
	ctx context.Context,
	experimentID uuid.UUID,
) (map[uuid.UUID]*ArmStats, error) {
	return e.base.GetArmStatistics(ctx, experimentID)
}

// GetObjectiveScores returns objective scores for all arms
func (e *AdvancedBanditEngine) GetObjectiveScores(
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
func (e *AdvancedBanditEngine) SetObjectiveConfig(
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
		hybridStrategy := NewHybridObjectiveStrategy(e.repo, e.cache, e.logger, config, e.base)
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
func (e *AdvancedBanditEngine) GetMetrics(ctx context.Context, experimentID uuid.UUID) (*BanditMetrics, error) {
	stats, err := e.GetArmStatistics(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	metrics := &BanditMetrics{
		BalanceIndex: e.calculateBalanceIndex(stats),
	}

	// Get additional metrics if strategies are enabled
	if delayedStrategy, err := e.getDelayedStrategy(); err == nil {
		if stats, err := delayedStrategy.GetStats(ctx); err == nil {
			if expired, ok := stats["expired_unprocessed"].(int); ok {
				metrics.PendingRewards = int64(expired)
			}
		}
	}

	if windowStrategy, err := e.getWindowStrategy(ctx, experimentID); err == nil {
		arms, armsErr := e.repo.GetArms(ctx, experimentID)
		if armsErr == nil && len(arms) > 0 {
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
	}

	return metrics, nil
}

func (e *AdvancedBanditEngine) GetWindowInfo(
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

func (e *AdvancedBanditEngine) TrimWindow(ctx context.Context, experimentID uuid.UUID) error {
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

func (e *AdvancedBanditEngine) ProcessExpiredPendingRewards(ctx context.Context, batchSize int) (int, error) {
	if batchSize <= 0 {
		batchSize = defaultBanditMaintenanceBatchSize
	}

	delayedStrategy, err := e.getDelayedStrategy()
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

func (e *AdvancedBanditEngine) TrimConfiguredWindows(ctx context.Context, limit int) (int, error) {
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

func (e *AdvancedBanditEngine) SyncObjectiveStats(ctx context.Context, limit int) (int, error) {
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
		config, err := e.getExperimentConfig(ctx, experimentID)
		if err != nil {
			return synced, err
		}
		objectiveTypes := maintenanceObjectiveTypes(config)
		if len(objectiveTypes) == 0 {
			continue
		}

		arms, err := e.repo.GetArms(ctx, experimentID)
		if err != nil {
			return synced, fmt.Errorf("failed to load arms for objective sync: %w", err)
		}

		for _, arm := range arms {
			stats, err := e.repo.GetArmStats(ctx, arm.ID)
			if err != nil {
				return synced, fmt.Errorf("failed to load arm stats for objective sync: %w", err)
			}

			for _, objectiveType := range objectiveTypes {
				if err := objectiveRepo.UpdateObjectiveStats(ctx, &ArmObjectiveStats{
					ArmID:         arm.ID,
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
		}
	}

	return synced, nil
}

func (e *AdvancedBanditEngine) CleanupOldContextData(ctx context.Context, olderThan time.Duration) (int64, error) {
	if olderThan <= 0 {
		olderThan = defaultBanditContextRetentionWindow
	}

	maintenanceRepo, ok := e.repo.(banditMaintenanceRepository)
	if !ok {
		return 0, nil
	}

	return maintenanceRepo.CleanupStaleUserContext(ctx, olderThan)
}

func (e *AdvancedBanditEngine) CleanupExpiredAssignments(ctx context.Context, olderThan time.Duration) (int64, error) {
	if olderThan <= 0 {
		olderThan = defaultBanditExpiredAssignmentRetention
	}

	maintenanceRepo, ok := e.repo.(banditMaintenanceRepository)
	if !ok {
		return 0, nil
	}

	return maintenanceRepo.CleanupExpiredAssignments(ctx, olderThan)
}

func (e *AdvancedBanditEngine) ExportWindowEvents(
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

func (e *AdvancedBanditEngine) GetObjectiveConfig(ctx context.Context, experimentID uuid.UUID) (*ExperimentConfig, error) {
	return e.getExperimentConfig(ctx, experimentID)
}

func (e *AdvancedBanditEngine) GetPendingReward(ctx context.Context, pendingID uuid.UUID) (*PendingReward, error) {
	delayedStrategy, err := e.getDelayedStrategy()
	if err != nil {
		return nil, err
	}

	return delayedStrategy.GetPendingReward(ctx, pendingID)
}

func (e *AdvancedBanditEngine) GetUserPendingRewards(
	ctx context.Context,
	userID uuid.UUID,
) ([]*PendingReward, error) {
	delayedStrategy, err := e.getDelayedStrategy()
	if err != nil {
		return nil, err
	}

	return delayedStrategy.GetPendingRewardsByUser(ctx, userID, uuid.Nil)
}

// calculateBalanceIndex measures how evenly users are distributed
// Returns 1.0 for perfect balance, 0.0 for all users in one arm
func (e *AdvancedBanditEngine) calculateBalanceIndex(stats map[uuid.UUID]*ArmStats) float64 {
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
func (e *AdvancedBanditEngine) RunMaintenanceDetailed(ctx context.Context) (*BanditMaintenanceSummary, error) {
	summary := &BanditMaintenanceSummary{}

	processed, err := e.ProcessExpiredPendingRewards(ctx, defaultBanditMaintenanceBatchSize)
	if err != nil {
		return summary, err
	}
	summary.ExpiredPendingRewards = processed

	if e.currencyService != nil {
		if err := e.currencyService.UpdateRates(ctx); err != nil {
			return summary, err
		}
		summary.CurrencyRatesUpdated = true
	}

	if maintenanceRepo, ok := e.repo.(banditMaintenanceRepository); ok {
		windowExperimentIDs, err := maintenanceRepo.ListWindowMaintenanceExperimentIDs(ctx, defaultBanditMaintenanceScanLimit)
		if err != nil {
			return summary, fmt.Errorf("failed to inspect window maintenance candidates: %w", err)
		}
		summary.WindowExperimentsScanned = len(windowExperimentIDs)

		trimmed, err := e.TrimConfiguredWindows(ctx, defaultBanditMaintenanceScanLimit)
		if err != nil {
			return summary, err
		}
		summary.WindowsTrimmed = trimmed

		objectiveExperimentIDs, err := maintenanceRepo.ListObjectiveSyncExperimentIDs(ctx, defaultBanditMaintenanceScanLimit)
		if err != nil {
			return summary, fmt.Errorf("failed to inspect objective sync candidates: %w", err)
		}
		summary.ObjectiveExperimentsScanned = len(objectiveExperimentIDs)
	}

	synced, err := e.SyncObjectiveStats(ctx, defaultBanditMaintenanceScanLimit)
	if err != nil {
		return summary, err
	}
	summary.ObjectiveStatsSynced = synced

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

// RunMaintenance performs periodic maintenance tasks.
func (e *AdvancedBanditEngine) RunMaintenance(ctx context.Context) error {
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
