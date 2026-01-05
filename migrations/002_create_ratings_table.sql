-- Migration: Create ratings table
-- Description: Cria tabela de avaliações de receitas com constraint de unicidade por usuário

CREATE TABLE IF NOT EXISTS ratings (
    id SERIAL PRIMARY KEY,
    recipe_id INTEGER NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    score INTEGER NOT NULL CHECK (score >= 1 AND score <= 5),
    comment TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT unique_user_recipe_rating UNIQUE (recipe_id, user_id)
);

-- Índices para melhorar performance de queries
CREATE INDEX idx_ratings_recipe_id ON ratings(recipe_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_ratings_user_id ON ratings(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_ratings_deleted_at ON ratings(deleted_at);
CREATE INDEX idx_ratings_score ON ratings(score) WHERE deleted_at IS NULL;

-- Índice composto para queries de listagem de ratings por receita
CREATE INDEX idx_ratings_recipe_created ON ratings(recipe_id, created_at DESC) WHERE deleted_at IS NULL;

COMMENT ON TABLE ratings IS 'Avaliações de receitas feitas por usuários autenticados';
COMMENT ON COLUMN ratings.score IS 'Nota de 1 a 5 estrelas';
COMMENT ON COLUMN ratings.comment IS 'Comentário opcional (max 1000 caracteres)';
COMMENT ON CONSTRAINT unique_user_recipe_rating ON ratings IS 'Garante que cada usuário pode ter apenas uma avaliação por receita';

