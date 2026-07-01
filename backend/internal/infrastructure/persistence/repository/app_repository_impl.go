package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	domainErrors "github.com/bivex/paywall-iap/internal/domain/errors"
	domainRepo "github.com/bivex/paywall-iap/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type appRepositoryImpl struct {
	pool       *pgxpool.Pool
	credEncKey []byte // 32-byte AES key from APP_CREDENTIALS_KEY env; empty = dev mode (no encryption)
}

// NewAppRepository creates a new AppRepository backed by a pgxpool.
// Reads APP_CREDENTIALS_KEY from env; if absent runs in dev mode (no encryption).
func NewAppRepository(pool *pgxpool.Pool) domainRepo.AppRepository {
	key := []byte(os.Getenv("APP_CREDENTIALS_KEY"))
	if len(key) != 0 && len(key) != 32 {
		panic("APP_CREDENTIALS_KEY must be exactly 32 bytes (AES-256)")
	}
	return &appRepositoryImpl{pool: pool, credEncKey: key}
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

// ── Settings ──────────────────────────────────────────────────────────────────

func (r *appRepositoryImpl) GetSettings(ctx context.Context, id uuid.UUID) (*entity.AppSettings, error) {
	var raw []byte
	err := r.pool.QueryRow(ctx, `SELECT settings FROM apps WHERE id = $1`, id).Scan(&raw)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("app not found: %w", domainErrors.ErrNotFound)
		}
		return nil, fmt.Errorf("get settings: %w", err)
	}
	var s entity.AppSettings
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("unmarshal settings: %w", err)
	}
	return &s, nil
}

func (r *appRepositoryImpl) UpdateSettings(ctx context.Context, id uuid.UUID, s *entity.AppSettings) error {
	raw, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	tag, err := r.pool.Exec(ctx,
		`UPDATE apps SET settings = $2, updated_at = now() WHERE id = $1`, id, raw)
	if err != nil {
		return fmt.Errorf("update settings: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("app not found: %w", domainErrors.ErrNotFound)
	}
	return nil
}

// ── Credentials ───────────────────────────────────────────────────────────────

func (r *appRepositoryImpl) GetCredentials(ctx context.Context, appID uuid.UUID) ([]*entity.AppCredentials, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, app_id, provider,
		       apple_shared_secret_enc, apple_team_id, apple_key_id,
		       apple_private_key_enc, apple_bundle_id, apple_environment,
		       google_package_name, google_service_account_enc,
		       stripe_publishable_key, stripe_secret_key_enc, stripe_webhook_secret_enc,
		       paddle_vendor_id, paddle_api_key_enc, paddle_webhook_secret_enc,
		       created_at, updated_at
		FROM app_credentials
		WHERE app_id = $1
		ORDER BY provider`, appID)
	if err != nil {
		return nil, fmt.Errorf("get credentials: %w", err)
	}
	defer rows.Close()

	var result []*entity.AppCredentials
	for rows.Next() {
		c, err := r.scanAndDecryptCredentials(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

func (r *appRepositoryImpl) UpsertCredentials(ctx context.Context, creds *entity.AppCredentials) error {
	k := r.credEncKey

	enc := func(s string) (string, error) { return encryptField(k, s) }

	appleSecretEnc, err := enc(creds.AppleSharedSecret)
	if err != nil {
		return fmt.Errorf("encrypt apple_shared_secret: %w", err)
	}
	appleKeyEnc, err := enc(creds.ApplePrivateKey)
	if err != nil {
		return fmt.Errorf("encrypt apple_private_key: %w", err)
	}
	googleSAEnc, err := enc(creds.GoogleServiceAccount)
	if err != nil {
		return fmt.Errorf("encrypt google_service_account: %w", err)
	}
	stripeSecretEnc, err := enc(creds.StripeSecretKey)
	if err != nil {
		return fmt.Errorf("encrypt stripe_secret_key: %w", err)
	}
	stripeWHEnc, err := enc(creds.StripeWebhookSecret)
	if err != nil {
		return fmt.Errorf("encrypt stripe_webhook_secret: %w", err)
	}
	paddleAPIEnc, err := enc(creds.PaddleAPIKey)
	if err != nil {
		return fmt.Errorf("encrypt paddle_api_key: %w", err)
	}
	paddleWHEnc, err := enc(creds.PaddleWebhookSecret)
	if err != nil {
		return fmt.Errorf("encrypt paddle_webhook_secret: %w", err)
	}

	appleEnv := creds.AppleEnvironment
	if appleEnv == "" {
		appleEnv = "production"
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO app_credentials (
			app_id, provider,
			apple_shared_secret_enc, apple_team_id, apple_key_id,
			apple_private_key_enc, apple_bundle_id, apple_environment,
			google_package_name, google_service_account_enc,
			stripe_publishable_key, stripe_secret_key_enc, stripe_webhook_secret_enc,
			paddle_vendor_id, paddle_api_key_enc, paddle_webhook_secret_enc,
			updated_at
		) VALUES (
			$1, $2,
			$3, $4, $5, $6, $7, $8,
			$9, $10,
			$11, $12, $13,
			$14, $15, $16,
			now()
		)
		ON CONFLICT (app_id, provider) DO UPDATE SET
			apple_shared_secret_enc    = EXCLUDED.apple_shared_secret_enc,
			apple_team_id              = EXCLUDED.apple_team_id,
			apple_key_id               = EXCLUDED.apple_key_id,
			apple_private_key_enc      = EXCLUDED.apple_private_key_enc,
			apple_bundle_id            = EXCLUDED.apple_bundle_id,
			apple_environment          = EXCLUDED.apple_environment,
			google_package_name        = EXCLUDED.google_package_name,
			google_service_account_enc = EXCLUDED.google_service_account_enc,
			stripe_publishable_key     = EXCLUDED.stripe_publishable_key,
			stripe_secret_key_enc      = EXCLUDED.stripe_secret_key_enc,
			stripe_webhook_secret_enc  = EXCLUDED.stripe_webhook_secret_enc,
			paddle_vendor_id           = EXCLUDED.paddle_vendor_id,
			paddle_api_key_enc         = EXCLUDED.paddle_api_key_enc,
			paddle_webhook_secret_enc  = EXCLUDED.paddle_webhook_secret_enc,
			updated_at                 = now()`,
		creds.AppID, creds.Provider,
		nullStr(appleSecretEnc), nullStr(creds.AppleTeamID), nullStr(creds.AppleKeyID),
		nullStr(appleKeyEnc), nullStr(creds.AppleBundleID), appleEnv,
		nullStr(creds.GooglePackageName), nullStr(googleSAEnc),
		nullStr(creds.StripePublishableKey), nullStr(stripeSecretEnc), nullStr(stripeWHEnc),
		nullStr(creds.PaddleVendorID), nullStr(paddleAPIEnc), nullStr(paddleWHEnc),
	)
	return err
}

func (r *appRepositoryImpl) DeleteCredentials(ctx context.Context, appID uuid.UUID, provider string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM app_credentials WHERE app_id = $1 AND provider = $2`, appID, provider)
	return err
}

func (r *appRepositoryImpl) GetCredentialsByProvider(ctx context.Context, appID uuid.UUID, provider string) (*entity.AppCredentials, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, app_id, provider,
			apple_shared_secret_enc, apple_team_id, apple_key_id,
			apple_private_key_enc, apple_bundle_id, apple_environment,
			google_package_name, google_service_account_enc,
			stripe_publishable_key, stripe_secret_key_enc, stripe_webhook_secret_enc,
			paddle_vendor_id, paddle_api_key_enc, paddle_webhook_secret_enc,
			created_at, updated_at
		FROM app_credentials
		WHERE app_id = $1 AND provider = $2
		LIMIT 1`, appID, provider)
	if err != nil {
		return nil, fmt.Errorf("query credentials by provider: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("credentials not found for provider %q", provider)
	}
	return r.scanAndDecryptCredentials(rows)
}

// ── helpers ───────────────────────────────────────────────────────────────────

// nullStr converts empty string to nil so pgx stores NULL instead of empty text.
func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func (r *appRepositoryImpl) scanAndDecryptCredentials(rows pgx.Rows) (*entity.AppCredentials, error) {
	var (
		c                                                   entity.AppCredentials
		appleSecretEnc, applePrivKeyEnc                     *string
		appleTeamID, appleKeyID, appleBundleID              *string
		googlePackageName                                   *string
		googleSAEnc                                         *string
		stripePublishableKey                                *string
		stripeSecretEnc, stripeWHEnc                        *string
		paddleVendorID                                      *string
		paddleAPIEnc, paddleWHEnc                           *string
	)
	err := rows.Scan(
		&c.ID, &c.AppID, &c.Provider,
		&appleSecretEnc, &appleTeamID, &appleKeyID,
		&applePrivKeyEnc, &appleBundleID, &c.AppleEnvironment,
		&googlePackageName, &googleSAEnc,
		&stripePublishableKey, &stripeSecretEnc, &stripeWHEnc,
		&paddleVendorID, &paddleAPIEnc, &paddleWHEnc,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan credentials: %w", err)
	}

	// Unpack nullable plain-text fields
	derefStr := func(p *string) string {
		if p == nil {
			return ""
		}
		return *p
	}
	c.AppleTeamID = derefStr(appleTeamID)
	c.AppleKeyID = derefStr(appleKeyID)
	c.AppleBundleID = derefStr(appleBundleID)
	c.GooglePackageName = derefStr(googlePackageName)
	c.StripePublishableKey = derefStr(stripePublishableKey)
	c.PaddleVendorID = derefStr(paddleVendorID)

	k := r.credEncKey
	dec := func(p *string) (string, error) {
		if p == nil {
			return "", nil
		}
		return decryptField(k, *p)
	}

	if c.AppleSharedSecret, err = dec(appleSecretEnc); err != nil {
		return nil, fmt.Errorf("decrypt apple_shared_secret: %w", err)
	}
	if c.ApplePrivateKey, err = dec(applePrivKeyEnc); err != nil {
		return nil, fmt.Errorf("decrypt apple_private_key: %w", err)
	}
	if c.GoogleServiceAccount, err = dec(googleSAEnc); err != nil {
		return nil, fmt.Errorf("decrypt google_service_account: %w", err)
	}
	if c.StripeSecretKey, err = dec(stripeSecretEnc); err != nil {
		return nil, fmt.Errorf("decrypt stripe_secret_key: %w", err)
	}
	if c.StripeWebhookSecret, err = dec(stripeWHEnc); err != nil {
		return nil, fmt.Errorf("decrypt stripe_webhook_secret: %w", err)
	}
	if c.PaddleAPIKey, err = dec(paddleAPIEnc); err != nil {
		return nil, fmt.Errorf("decrypt paddle_api_key: %w", err)
	}
	if c.PaddleWebhookSecret, err = dec(paddleWHEnc); err != nil {
		return nil, fmt.Errorf("decrypt paddle_webhook_secret: %w", err)
	}

	return &c, nil
}
