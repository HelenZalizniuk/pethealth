CREATE TABLE IF NOT EXISTS health_metrics (
    id UUID PRIMARY KEY,
    pet_id BIGINT NOT NULL,
    type VARCHAR(50) NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS outbox_events (
    id UUID PRIMARY KEY,
    pet_id BIGINT NOT NULL,
    type VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    topic VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);


CREATE INDEX IF NOT EXISTS idx_outbox_status_pending ON outbox_events (created_at) 
WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_metrics_pet_id ON health_metrics (pet_id);