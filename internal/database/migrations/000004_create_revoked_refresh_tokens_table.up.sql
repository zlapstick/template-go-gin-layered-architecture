CREATE TABLE revoked_refresh_tokens (
    token_hash VARCHAR(255) PRIMARY KEY,
    revoked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);