-- Criação da tabela de refresh tokens
-- Esta tabela armazena tokens de refresh para autenticação de longa duração

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    device_name VARCHAR(255),
    device_fingerprint VARCHAR(255),
    ip_address VARCHAR(45),
    expires_at TIMESTAMP NOT NULL,
    last_used_at TIMESTAMP,
    revoked_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Índices para otimizar queries
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_deleted_at ON refresh_tokens(deleted_at);

-- Comentários para documentação
COMMENT ON TABLE refresh_tokens IS 'Armazena tokens de refresh para autenticação de longa duração';
COMMENT ON COLUMN refresh_tokens.token_hash IS 'SHA256 hash do token - nunca armazenar token em texto puro';
COMMENT ON COLUMN refresh_tokens.device_name IS 'Nome amigável do dispositivo (ex: iPhone 13, iPad)';
COMMENT ON COLUMN refresh_tokens.device_fingerprint IS 'Hash do user-agent + IP para validação adicional';
COMMENT ON COLUMN refresh_tokens.ip_address IS 'Endereço IP que criou o token';
COMMENT ON COLUMN refresh_tokens.last_used_at IS 'Última vez que o token foi usado para refresh';
COMMENT ON COLUMN refresh_tokens.revoked_at IS 'Timestamp de revogação manual do token';

