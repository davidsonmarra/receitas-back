# Migrações de Banco de Dados

Este diretório contém os scripts SQL para migrações do banco de dados.

## Como Aplicar Migrações

### Usando PostgreSQL diretamente

```bash
# Conectar ao banco de dados e executar o script
psql -h localhost -U seu_usuario -d nome_do_banco -f migrations/001_add_instructions_to_recipes.sql
```

### Usando Docker (se o banco estiver em container)

```bash
# Copiar o arquivo para dentro do container e executar
docker cp migrations/001_add_instructions_to_recipes.sql nome_do_container:/tmp/
docker exec -it nome_do_container psql -U seu_usuario -d nome_do_banco -f /tmp/001_add_instructions_to_recipes.sql
```

### Verificar se a migração foi aplicada

```sql
-- Verificar estrutura da tabela recipes
\d recipes

-- Ou usar SQL padrão
SELECT column_name, data_type, is_nullable 
FROM information_schema.columns 
WHERE table_name = 'recipes' AND column_name = 'instructions';
```

## Lista de Migrações

### 001_add_instructions_to_recipes.sql
- **Data:** 2025-12-29
- **Descrição:** Adiciona coluna `instructions` (TEXT, nullable) à tabela `recipes` para armazenar o modo de preparo em formato Markdown
- **Reversão:** `ALTER TABLE recipes DROP COLUMN instructions;`

### 002_create_ratings_table.sql
- **Data:** 2026-01-04
- **Descrição:** Cria tabela `ratings` para sistema de avaliações de receitas com scores (1-5), comentários opcionais, constraint de unicidade por usuário/receita e índices para performance
- **Reversão:** `DROP TABLE ratings;`

## Notas Importantes

- As migrações devem ser aplicadas na ordem numérica
- Sempre faça backup do banco de dados antes de aplicar migrações em produção
- Teste as migrações em ambiente de desenvolvimento primeiro
- O GORM pode criar automaticamente a coluna quando o servidor iniciar (auto-migrate), mas é recomendado aplicar a migração manualmente para ter mais controle

