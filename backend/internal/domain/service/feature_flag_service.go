package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

var (
	ErrFeatureFlagNotFound = errors.New("feature flag not found")
)

// FeatureFlag represents an A/B test feature flag
type FeatureFlag struct {
	ID             string
	Name           string
	Enabled        bool
	RolloutPercent int      // 0-100
	UserIDs        []string // Specific users who should see the feature
}

// FeatureFlagService handles A/B testing feature flags
type FeatureFlagService struct {
	// In production, integrate with go-feature-flag or similar
	// For now, we'll use in-memory storage
	flags map[string]*FeatureFlag
}

// NewFeatureFlagService creates a new feature flag service
func NewFeatureFlagService() *FeatureFlagService {
	return &FeatureFlagService{
		flags: make(map[string]*FeatureFlag),
	}
}

// CreateFlag creates a new feature flag
func (s *FeatureFlagService) CreateFlag(id, name string, enabled bool, rolloutPercent int, userIDs []string) *FeatureFlag {
	flag := &FeatureFlag{
		ID:             id,
		Name:           name,
		Enabled:        enabled,
		RolloutPercent: rolloutPercent,
		UserIDs:        userIDs,
	}

	s.flags[id] = flag
	return flag
}

// IsFeatureEnabled checks if a feature is enabled for a specific user
func (s *FeatureFlagService) IsFeatureEnabled(ctx context.Context, flagID, userID string) (bool, error) {
	flag, exists := s.flags[flagID]
	if !exists {
		return false, ErrFeatureFlagNotFound
	}

	// If flag is disabled, return false
	if !flag.Enabled {
		return false, nil
	}

	// Check if user is in explicit user list
	for _, uid := range flag.UserIDs {
		if uid == userID {
			return true, nil
		}
	}

	// Use consistent hashing to determine if user is in rollout percentage
	return s.isUserInRollout(flagID, userID, flag.RolloutPercent), nil
}

// isUserInRollout uses consistent hashing to determine if user is in the rollout
func (s *FeatureFlagService) isUserInRollout(flagID, userID string, rolloutPercent int) bool {
	if rolloutPercent <= 0 {
		return false
	}
	if rolloutPercent >= 100 {
		return true
	}

	// Create consistent hash of flagID + userID
	hash := sha256.Sum256([]byte(flagID + ":" + userID))
	hashStr := hex.EncodeToString(hash[:])

	// Convert first 8 bytes of hash to number 0-100
	hashInt := hexToUint64(hashStr[:16])
	userBucket := hashInt % 100

	return userBucket < uint64(rolloutPercent)
}

func hexToUint64(s string) uint64 {
	var result uint64
	for i := 0; i < len(s); i++ {
		c := s[i]
		var val byte
		if c >= '0' && c <= '9' {
			val = c - '0'
		} else if c >= 'a' && c <= 'f' {
			val = c - 'a' + 10
		} else if c >= 'A' && c <= 'F' {
			val = c - 'A' + 10
		}
		result = result*16 + uint64(val)
	}
	return result
}

// GetFlag returns a feature flag by ID
func (s *FeatureFlagService) GetFlag(flagID string) (*FeatureFlag, error) {
	flag, exists := s.flags[flagID]
	if !exists {
		return nil, ErrFeatureFlagNotFound
	}
	return flag, nil
}

// UpdateFlag updates an existing feature flag
func (s *FeatureFlagService) UpdateFlag(flagID string, enabled *bool, rolloutPercent *int, userIDs []string) error {
	flag, exists := s.flags[flagID]
	if !exists {
		return ErrFeatureFlagNotFound
	}

	if enabled != nil {
		flag.Enabled = *enabled
	}
	if rolloutPercent != nil {
		flag.RolloutPercent = *rolloutPercent
	}
	if userIDs != nil {
		flag.UserIDs = userIDs
	}

	return nil
}

// DeleteFlag removes a feature flag
func (s *FeatureFlagService) DeleteFlag(flagID string) error {
	if _, exists := s.flags[flagID]; !exists {
		return ErrFeatureFlagNotFound
	}
	delete(s.flags, flagID)
	return nil
}

// GetAllFlags returns all feature flags
func (s *FeatureFlagService) GetAllFlags() []*FeatureFlag {
	flags := make([]*FeatureFlag, 0, len(s.flags))
	for _, flag := range s.flags {
		flags = append(flags, flag)
	}
	return flags
}

// EvaluatePaywallTest returns the paywall variant for a user
func (s *FeatureFlagService) EvaluatePaywallTest(ctx context.Context, userID string) (string, error) {
	enabled, err := s.IsFeatureEnabled(ctx, "paywall_variant_test", userID)
	if err != nil {
		if errors.Is(err, ErrFeatureFlagNotFound) {
			return "control", nil
		}
		return "", err
	}

	if enabled {
		return "variant_b", nil
	}
	return "control", nil
}
