-- Migração: Adicionar campo instructions (modo de preparo em Markdown) à tabela recipes
-- Data: 2025-12-29
-- Descrição: Adiciona coluna instructions do tipo TEXT para armazenar o modo de preparo
--            das receitas em formato Markdown. O campo é opcional (nullable).

-- Adicionar coluna instructions
ALTER TABLE recipes ADD COLUMN instructions TEXT;

-- Comentário explicativo (opcional, dependendo do banco de dados)
-- COMMENT ON COLUMN recipes.instructions IS 'Modo de preparo da receita em formato Markdown';

