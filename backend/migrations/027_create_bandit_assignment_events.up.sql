CREATE TABLE bandit_assignment_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    assignment_id UUID NOT NULL REFERENCES ab_test_assignments(id) ON DELETE CASCADE,
    experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    arm_id UUID NOT NULL REFERENCES ab_test_arms(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL CHECK (event_type IN ('assigned')),
    metadata JSONB,
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_bandit_assignment_events_experiment
    ON bandit_assignment_events(experiment_id, occurred_at DESC);

CREATE INDEX idx_bandit_assignment_events_user
    ON bandit_assignment_events(user_id, occurred_at DESC);

CREATE INDEX idx_bandit_assignment_events_assignment
    ON bandit_assignment_events(assignment_id, occurred_at DESC);

COMMENT ON TABLE bandit_assignment_events IS 'Append-only assignment history for bandit arm selections';