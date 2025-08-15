CREATE TABLE IF NOT EXISTS blacklisted_tokens (
    id SERIAL PRIMARY KEY,
    token TEXT NOT NULL UNIQUE,
    user_id INTEGER NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_blacklisted_tokens_token ON blacklisted_tokens(token);
CREATE INDEX idx_blacklisted_tokens_expires_at ON blacklisted_tokens(expires_at);
CREATE INDEX idx_blacklisted_tokens_user_id ON blacklisted_tokens(user_id); 