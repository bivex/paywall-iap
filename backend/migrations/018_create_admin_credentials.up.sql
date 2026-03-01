-- Admin credentials: password-based login for admin/superadmin users
CREATE TABLE IF NOT EXISTS admin_credentials (
    user_id     UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    password_hash TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
