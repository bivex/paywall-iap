package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuditService handles logging admin actions
type AuditService struct {
	pool *pgxpool.Pool
}

// NewAuditService creates a new audit service
func NewAuditService(pool *pgxpool.Pool) *AuditService {
	return &AuditService{
		pool: pool,
	}
}

// AuditActionParams encapsulates parameters for logging an admin audit action
type AuditActionParams struct {
	AdminID      uuid.UUID
	Action       string
	TargetType   string
	TargetUserID *uuid.UUID
	Details      map[string]interface{}
}

// LogAction logs an admin action to the audit log
func (s *AuditService) LogAction(ctx context.Context, p AuditActionParams) error {
	query := `
		INSERT INTO admin_audit_log (
			admin_id, action, target_type, target_user_id, details
		) VALUES ($1, $2, $3, $4, $5)
	`
	_, err := s.pool.Exec(ctx, query, p.AdminID, p.Action, p.TargetType, p.TargetUserID, p.Details)
	return err
}
