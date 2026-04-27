CREATE EXTENSION IF NOT EXISTS "pgcrypto";

ALTER TABLE health_metrics 
ALTER COLUMN id SET DEFAULT gen_random_uuid();

ALTER TABLE outbox_events 
ALTER COLUMN id SET DEFAULT gen_random_uuid();

ALTER TABLE health_metrics ADD COLUMN IF NOT EXISTS external_id UUID;
ALTER TABLE health_metrics ADD COLUMN IF NOT EXISTS shard_id INTEGER;
ALTER TABLE health_metrics ADD COLUMN IF NOT EXISTS timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW();


CREATE UNIQUE INDEX IF NOT EXISTS idx_metrics_external_id ON health_metrics (external_id);

CREATE INDEX IF NOT EXISTS idx_metrics_pet_id ON health_metrics (pet_id);
