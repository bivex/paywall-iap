package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserProfileService handles user profile data aggregation
type UserProfileService struct {
	dbPool *pgxpool.Pool
}

// NewUserProfileService creates a new user profile service
func NewUserProfileService(dbPool *pgxpool.Pool) *UserProfileService {
	return &UserProfileService{dbPool: dbPool}
}

// UserInfo represents user identity information
type UserInfo struct {
	ID             string  `json:"id"`
	PlatformUserID string  `json:"platform_user_id"`
	DeviceID       *string `json:"device_id"`
	Platform       string  `json:"platform"`
	AppVersion     string  `json:"app_version"`
	Email          string  `json:"email"`
	Role           string  `json:"role"`
	LTV            float64 `json:"ltv"`
	CreatedAt      string  `json:"created_at"`
}

// SubscriptionRow represents a user's subscription
type SubscriptionRow struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Source    string `json:"source"`
	Platform  string `json:"platform"`
	ProductID string `json:"product_id"`
	PlanType  string `json:"plan_type"`
	ExpiresAt string `json:"expires_at"`
	AutoRenew bool   `json:"auto_renew"`
	CreatedAt string `json:"created_at"`
}

// TransactionRow represents a user's transaction
type TransactionRow struct {
	ID       string  `json:"id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Status   string  `json:"status"`
	TxID     *string `json:"provider_tx_id"`
	Date     string  `json:"date"`
}

// AuditRow represents an audit log entry
type AuditRow struct {
	Action     string `json:"action"`
	AdminEmail string `json:"admin_email"`
	Detail     string `json:"detail"`
	Date       string `json:"date"`
}

// DunningRow represents a dunning entry
type DunningRow struct {
	Status       string  `json:"status"`
	AttemptCount int     `json:"attempt_count"`
	MaxAttempts  int     `json:"max_attempts"`
	NextAttempt  *string `json:"next_attempt_at"`
	CreatedAt    string  `json:"created_at"`
}

// Profile represents the complete user profile
type Profile struct {
	User          UserInfo           `json:"user"`
	Subscriptions []SubscriptionRow `json:"subscriptions"`
	Transactions  []TransactionRow  `json:"transactions"`
	AuditLog      []AuditRow        `json:"audit_log"`
	Dunning       []DunningRow      `json:"dunning"`
}

// GetProfile fetches the complete user profile
func (s *UserProfileService) GetProfile(ctx context.Context, userID uuid.UUID) (*Profile, error) {
	user, err := s.fetchUserInfo(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetch user info: %w", err)
	}

	subs, err := s.fetchSubscriptions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetch subscriptions: %w", err)
	}

	txs, err := s.fetchTransactions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetch transactions: %w", err)
	}

	audits, err := s.fetchAuditLog(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetch audit log: %w", err)
	}

	dunnings, err := s.fetchDunning(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetch dunning: %w", err)
	}

	return &Profile{
		User:          *user,
		Subscriptions: subs,
		Transactions:  txs,
		AuditLog:      audits,
		Dunning:       dunnings,
	}, nil
}

// fetchUserInfo retrieves user identity information
func (s *UserProfileService) fetchUserInfo(ctx context.Context, userID uuid.UUID) (*UserInfo, error) {
	var user UserInfo
	var createdAt time.Time

	err := s.dbPool.QueryRow(ctx,
		`SELECT id, platform_user_id, device_id, platform, app_version, email, role, ltv, created_at
		 FROM users WHERE id = $1`, userID,
	).Scan(&userID, &user.PlatformUserID, &user.DeviceID, &user.Platform, &user.AppVersion,
		&user.Email, &user.Role, &user.LTV, &createdAt)
	if err != nil {
		return nil, err
	}

	user.ID = userID.String()
	user.CreatedAt = createdAt.Format(time.RFC3339)
	return &user, nil
}

// fetchSubscriptions retrieves user subscriptions
func (s *UserProfileService) fetchSubscriptions(ctx context.Context, userID uuid.UUID) ([]SubscriptionRow, error) {
	rows, err := s.dbPool.Query(ctx,
		`SELECT id, status, source, platform, product_id, plan_type, expires_at, auto_renew, created_at
		 FROM subscriptions WHERE user_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 10`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []SubscriptionRow
	for rows.Next() {
		var s SubscriptionRow
		var sid uuid.UUID
		var exp, cat time.Time
		if err := rows.Scan(&sid, &s.Status, &s.Source, &s.Platform, &s.ProductID, &s.PlanType, &exp, &s.AutoRenew, &cat); err != nil {
			continue
		}
		s.ID = sid.String()
		s.ExpiresAt = exp.Format(time.RFC3339)
		s.CreatedAt = cat.Format(time.RFC3339)
		subs = append(subs, s)
	}

	return subs, nil
}

// fetchTransactions retrieves user transactions
func (s *UserProfileService) fetchTransactions(ctx context.Context, userID uuid.UUID) ([]TransactionRow, error) {
	rows, err := s.dbPool.Query(ctx,
		`SELECT id, amount, currency, status, provider_tx_id, created_at
		 FROM transactions WHERE user_id = $1 ORDER BY created_at DESC LIMIT 20`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []TransactionRow
	for rows.Next() {
		var t TransactionRow
		var tid uuid.UUID
		var tdate time.Time
		var amountStr string
		if err := rows.Scan(&tid, &amountStr, &t.Currency, &t.Status, &t.TxID, &tdate); err != nil {
			continue
		}
		t.ID = tid.String()
		fmt.Sscanf(amountStr, "%f", &t.Amount)
		t.Date = tdate.Format(time.RFC3339)
		txs = append(txs, t)
	}

	return txs, nil
}

// fetchAuditLog retrieves audit log entries for the user
func (s *UserProfileService) fetchAuditLog(ctx context.Context, userID uuid.UUID) ([]AuditRow, error) {
	rows, err := s.dbPool.Query(ctx,
		`SELECT a.action, COALESCE(u2.email, a.admin_id::text), COALESCE(a.details::text,'{}'), a.created_at
		 FROM admin_audit_log a LEFT JOIN users u2 ON u2.id = a.admin_id
		 WHERE a.target_user_id = $1 ORDER BY a.created_at DESC LIMIT 10`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var audits []AuditRow
	for rows.Next() {
		var a AuditRow
		var adate time.Time
		if err := rows.Scan(&a.Action, &a.AdminEmail, &a.Detail, &adate); err != nil {
			continue
		}
		a.Date = adate.Format(time.RFC3339)
		audits = append(audits, a)
	}

	return audits, nil
}

// fetchDunning retrieves dunning entries for the user
func (s *UserProfileService) fetchDunning(ctx context.Context, userID uuid.UUID) ([]DunningRow, error) {
	rows, err := s.dbPool.Query(ctx,
		`SELECT status, attempt_count, max_attempts, next_attempt_at, created_at
		 FROM dunning WHERE user_id = $1 ORDER BY created_at DESC LIMIT 5`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dunnings []DunningRow
	for rows.Next() {
		var d DunningRow
		var ddate time.Time
		var nextAt *time.Time
		if err := rows.Scan(&d.Status, &d.AttemptCount, &d.MaxAttempts, &nextAt, &ddate); err != nil {
			continue
		}
		if nextAt != nil {
			s := nextAt.Format(time.RFC3339)
			d.NextAttempt = &s
		}
		d.CreatedAt = ddate.Format(time.RFC3339)
		dunnings = append(dunnings, d)
	}

	return dunnings, nil
}
