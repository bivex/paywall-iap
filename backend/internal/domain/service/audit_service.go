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

// LogAction logs an admin action to the audit log
func (s *AuditService) LogAction(ctx context.Context, adminID uuid.UUID, action, targetType string, targetUserID *uuid.UUID, details map[string]interface{}) error {
	query := `
		INSERT INTO admin_audit_log (
			admin_id, action, target_type, target_user_id, details
		) VALUES ($1, $2, $3, $4, $5)
	`
	_, err := s.pool.Exec(ctx, query, adminID, action, targetType, targetUserID, details)
	return err
}
