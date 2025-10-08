CREATE TABLE IF NOT EXISTS links (
                       short_id VARCHAR(10) PRIMARY KEY,
                       original_url TEXT NOT NULL UNIQUE,
                       created_at TIMESTAMP DEFAULT NOW(),
                       click_count BIGINT DEFAULT 0
);
