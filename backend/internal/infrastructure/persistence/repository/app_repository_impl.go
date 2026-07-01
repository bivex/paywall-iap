package repository

import (
	"context"
	"fmt"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	domainErrors "github.com/bivex/paywall-iap/internal/domain/errors"
	domainRepo "github.com/bivex/paywall-iap/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type appRepositoryImpl struct {
	pool *pgxpool.Pool
}

// NewAppRepository creates a new AppRepository backed by a pgxpool.
func NewAppRepository(pool *pgxpool.Pool) domainRepo.AppRepository {
	return &appRepositoryImpl{pool: pool}
}

const appSelectColumns = `
	id, name, display_name, platform, bundle_id, is_active, created_at, updated_at
`

func (r *appRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.App, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT`+appSelectColumns+`FROM apps WHERE id = $1 AND is_active = true`, id)
	return scanApp(row)
}

func (r *appRepositoryImpl) GetByBundleID(ctx context.Context, bundleID string) (*entity.App, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT`+appSelectColumns+`FROM apps WHERE bundle_id = $1 AND is_active = true`, bundleID)
	return scanApp(row)
}

func (r *appRepositoryImpl) List(ctx context.Context) ([]*entity.App, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT`+appSelectColumns+`FROM apps ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("failed to list apps: %w", err)
	}
	defer rows.Close()

	var apps []*entity.App
	for rows.Next() {
		app, err := scanAppRow(rows)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}
	return apps, rows.Err()
}

func scanApp(row pgx.Row) (*entity.App, error) {
	var a entity.App
	err := row.Scan(&a.ID, &a.Name, &a.DisplayName, &a.Platform, &a.BundleID,
		&a.IsActive, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("app not found: %w", domainErrors.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to scan app: %w", err)
	}
	return &a, nil
}

func scanAppRow(rows pgx.Rows) (*entity.App, error) {
	var a entity.App
	err := rows.Scan(&a.ID, &a.Name, &a.DisplayName, &a.Platform, &a.BundleID,
		&a.IsActive, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan app row: %w", err)
	}
	return &a, nil
}

func (r *appRepositoryImpl) Create(ctx context.Context, name, bundleID, platform string) (*entity.App, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO apps (name, display_name, bundle_id, platform, is_active)
		VALUES ($1, $1, $2, $3, true)
		RETURNING`+appSelectColumns,
		name, bundleID, platform)
	return scanApp(row)
}

func (r *appRepositoryImpl) Update(ctx context.Context, app *entity.App) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE apps
		SET name = $2, display_name = $3, bundle_id = $4, platform = $5, is_active = $6, updated_at = now()
		WHERE id = $1`,
		app.ID, app.Name, app.DisplayName, app.BundleID, app.Platform, app.IsActive)
	if err != nil {
		return fmt.Errorf("failed to update app: %w", err)
	}
	return nil
}

func (r *appRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE apps SET is_active = false, updated_at = now() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete app: %w", err)
	}
	return nil
}
