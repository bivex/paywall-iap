package iap

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/domain/repository"
)

const credentialTTL = 5 * time.Minute

type cachedCred struct {
	cred      *entity.AppCredentials
	expiresAt time.Time
}

// CredentialResolver loads app_credentials per (app_id, provider) with TTL cache.
type CredentialResolver struct {
	repo  repository.AppRepository
	mu    sync.RWMutex
	cache map[string]*cachedCred // key: "appID/provider"
}

func NewCredentialResolver(repo repository.AppRepository) *CredentialResolver {
	return &CredentialResolver{
		repo:  repo,
		cache: make(map[string]*cachedCred),
	}
}

func (r *CredentialResolver) key(appID uuid.UUID, provider string) string {
	return appID.String() + "/" + provider
}

// Resolve returns credentials for the given app+provider, using cache when fresh.
func (r *CredentialResolver) Resolve(ctx context.Context, appID uuid.UUID, provider string) (*entity.AppCredentials, error) {
	k := r.key(appID, provider)

	// fast path: read lock
	r.mu.RLock()
	if entry, ok := r.cache[k]; ok && time.Now().Before(entry.expiresAt) {
		r.mu.RUnlock()
		return entry.cred, nil
	}
	r.mu.RUnlock()

	// slow path: fetch from DB
	creds, err := r.repo.GetCredentialsByProvider(ctx, appID, provider)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	r.cache[k] = &cachedCred{cred: creds, expiresAt: time.Now().Add(credentialTTL)}
	r.mu.Unlock()

	return creds, nil
}

// Invalidate removes a single entry (call after PUT credentials).
func (r *CredentialResolver) Invalidate(appID uuid.UUID, provider string) {
	r.mu.Lock()
	delete(r.cache, r.key(appID, provider))
	r.mu.Unlock()
}
