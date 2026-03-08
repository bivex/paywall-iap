CREATE TABLE experiment_automation_decision_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
    source TEXT NOT NULL,
    decision_type TEXT NOT NULL CHECK (decision_type IN ('status_transition')),
    reason TEXT,
    from_status TEXT NOT NULL,
    to_status TEXT NOT NULL,
    idempotency_key TEXT,
    details JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_experiment_automation_decision_log_idempotency
    ON experiment_automation_decision_log(idempotency_key);

CREATE INDEX idx_experiment_automation_decision_log_experiment
    ON experiment_automation_decision_log(experiment_id, created_at DESC);

CREATE TABLE bandit_conversion_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
    arm_id UUID NOT NULL REFERENCES ab_test_arms(id) ON DELETE CASCADE,
    user_id UUID,
    pending_reward_id UUID REFERENCES bandit_pending_rewards(id) ON DELETE SET NULL,
    transaction_id UUID,
    event_type TEXT NOT NULL CHECK (event_type IN ('direct_reward', 'delayed_conversion', 'expired_pending_reward')),
    original_reward_value DOUBLE PRECISION NOT NULL DEFAULT 0,
    original_currency TEXT,
    normalized_reward_value DOUBLE PRECISION NOT NULL DEFAULT 0,
    normalized_currency TEXT,
    metadata JSONB,
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_bandit_conversion_events_experiment
    ON bandit_conversion_events(experiment_id, occurred_at DESC);

CREATE INDEX idx_bandit_conversion_events_user
    ON bandit_conversion_events(user_id, occurred_at DESC);

CREATE INDEX idx_bandit_conversion_events_arm
    ON bandit_conversion_events(arm_id, occurred_at DESC);

CREATE UNIQUE INDEX idx_bandit_conversion_events_pending_event
    ON bandit_conversion_events(pending_reward_id, event_type)
    WHERE pending_reward_id IS NOT NULL;

CREATE UNIQUE INDEX idx_bandit_conversion_events_transaction_delayed
    ON bandit_conversion_events(transaction_id)
    WHERE transaction_id IS NOT NULL AND event_type = 'delayed_conversion';

COMMENT ON TABLE experiment_automation_decision_log IS 'Append-only log of automation decisions applied by system reconciler flows';
COMMENT ON TABLE bandit_conversion_events IS 'Append-only reward and conversion event log for direct, delayed, and expired bandit outcomes';