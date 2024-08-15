CREATE TABLE IF NOT EXISTS logs (
                             "id" SERIAL NOT NULL,
                             time        TIMESTAMPTZ(3)     NOT NULL,
                             "hash" TEXT NOT NULL,
                             log jsonb NOT  NULL
);

CREATE UNIQUE INDEX "log_hash_key" ON "logs"("hash", "time");

SELECT create_hypertable('logs', 'time', if_not_exists => TRUE, create_default_indexes => TRUE);
SELECT add_retention_policy('logs', drop_after => INTERVAL '1 week', if_not_exists => TRUE);
