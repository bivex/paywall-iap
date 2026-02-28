-- Create ab_test_assignments table for user assignment tracking
CREATE TABLE ab_test_assignments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id   UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    arm_id          UUID NOT NULL REFERENCES ab_test_arms(id) ON DELETE CASCADE,

    -- Assignment timestamps
    assigned_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at      TIMESTAMPTZ NOT NULL DEFAULT (now() + interval '24 hours'),

    -- Ensure one active assignment per user per experiment
    CONSTRAINT unique_active_assignment UNIQUE (experiment_id, user_id)
        WHERE expires_at > now()
);

-- Index for active assignment lookup
CREATE INDEX idx_ab_test_assignments_active ON ab_test_assignments(experiment_id, user_id, expires_at)
    WHERE expires_at > now();

-- Index for cleanup of expired assignments
CREATE INDEX idx_ab_test_assignments_expires ON ab_test_assignments(expires_at)
    WHERE expires_at <= now();

-- Index for user's assignment history
CREATE INDEX idx_ab_test_assignments_user_history ON ab_test_assignments(user_id, assigned_at DESC);

-- Function to check if assignment is still valid
CREATE OR REPLACE FUNCTION is_assignment_active(expires_at TIMESTAMPTZ) RETURNS BOOLEAN AS $$
BEGIN
    RETURN expires_at > now();
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Comment for documentation
COMMENT ON TABLE ab_test_assignments IS 'User assignments to experiment arms with sticky 24h TTL';
COMMENT ON COLUMN ab_test_assignments.expires_at IS 'Assignment expiration - after this, user can be reassigned';
COMMENT ON COLUMN ab_test_assignments.assigned_at IS 'When the user was first assigned to this arm';
COMMENT ON FUNCTION is_assignment_active IS 'Check if an assignment is still valid (not expired)';
