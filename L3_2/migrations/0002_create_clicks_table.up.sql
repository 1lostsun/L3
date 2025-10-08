CREATE TABLE IF NOT EXISTS clicks (
    id BIGSERIAL PRIMARY KEY,
    short_id TEXT NOT NULL REFERENCES links(short_id) ON DELETE CASCADE,
    user_agent TEXT,
    ip TEXT,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW()
    );
