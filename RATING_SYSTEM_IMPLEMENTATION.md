# Sistema de Avaliações - Documentação de Implementação

## Visão Geral

Sistema completo de avaliações para receitas, permitindo usuários autenticados avaliar receitas com notas de 1 a 5 estrelas e comentários opcionais. Implementado com controle de unicidade (um usuário pode ter apenas uma avaliação por receita), suporte a edição/exclusão, cálculo automático de estatísticas e moderação administrativa.

## Estrutura de Dados

### Modelo Rating

**Arquivo**: `internal/models/rating.go`

```go
type Rating struct {
    ID        uint           // ID único da avaliação
    RecipeID  uint           // ID da receita (FK)
    UserID    uint           // ID do usuário (FK)
    Score     int            // Nota de 1 a 5
    Comment   string         // Comentário opcional (max 1000 chars)
    Recipe    *Recipe        // Relacionamento com receita
    User      *User          // Relacionamento com usuário
    CreatedAt time.Time      // Data de criação
    UpdatedAt time.Time      // Data de atualização
    DeletedAt gorm.DeletedAt // Soft delete
}
```

**Constraints**:
- Índice único composto (recipe_id, user_id) - garante uma avaliação por usuário/receita
- Score validado entre 1 e 5 (validação no modelo e constraint no DB)
- Comment limitado a 1000 caracteres

### Campos Adicionais no Recipe

O modelo `Recipe` foi atualizado com campos calculados:

```go
AverageRating float64 `gorm:"-" json:"average_rating,omitempty"` // Média das avaliações
RatingCount   int64   `gorm:"-" json:"rating_count,omitempty"`   // Total de avaliações
```

Estes campos são calculados em tempo real e não são salvos no banco de dados (tag `gorm:"-"`).

## Migration

**Arquivo**: `migrations/002_create_ratings_table.sql`

Cria a tabela `ratings` com:
- Constraint de unicidade (recipe_id, user_id)
- Constraint CHECK para score (1-5)
- Índices para performance em queries
- Cascade delete quando receita ou usuário são deletados

## Endpoints da API

### 1. Criar ou Atualizar Avaliação (Upsert)

```http
POST /recipes/{id}/ratings
Authorization: Bearer {token}
Content-Type: application/json

{
  "score": 5,
  "comment": "Receita maravilhosa!" // opcional
}
```

**Comportamento**:
- Se o usuário já avaliou: atualiza a avaliação existente (200 OK)
- Se é a primeira avaliação: cria nova (201 Created)
- Valida score (1-5) e tamanho do comentário (max 1000 chars)

**Resposta**:
```json
{
  "id": 1,
  "recipe_id": 5,
  "user_id": 10,
  "user": {
    "id": 10,
    "name": "João Silva"
  },
  "score": 5,
  "comment": "Receita maravilhosa!",
  "created_at": "2026-01-04T10:00:00Z",
  "updated_at": "2026-01-04T10:00:00Z"
}
```

### 2. Obter Minha Avaliação

```http
GET /recipes/{id}/ratings/me
Authorization: Bearer {token}
```

Retorna a avaliação do usuário logado para a receita especificada.
- 200 OK com a avaliação
- 404 Not Found se o usuário ainda não avaliou

### 3. Deletar Minha Avaliação

```http
DELETE /recipes/{id}/ratings/me
Authorization: Bearer {token}
```

Deleta a avaliação do usuário logado (soft delete).
- 200 OK com mensagem de sucesso
- 404 Not Found se a avaliação não existe

### 4. Listar Avaliações de uma Receita

```http
GET /recipes/{id}/ratings?page=1&limit=10&sort=newest
```

**Endpoint público** (não requer autenticação).

**Query Parameters**:
- `page`: número da página (padrão: 1)
- `limit`: itens por página (padrão: 10)
- `sort`: ordenação
  - `newest`: mais recentes primeiro (padrão)
  - `oldest`: mais antigas primeiro
  - `highest`: maior nota primeiro
  - `lowest`: menor nota primeiro

**Resposta**:
```json
{
  "data": [
    {
      "id": 1,
      "recipe_id": 5,
      "user_id": 10,
      "user": {
        "id": 10,
        "name": "João Silva"
      },
      "score": 5,
      "comment": "Receita maravilhosa!",
      "created_at": "2026-01-04T10:00:00Z",
      "updated_at": "2026-01-04T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 50,
    "total_pages": 5
  }
}
```

### 5. Obter Estatísticas de Avaliação

```http
GET /recipes/{id}/ratings/stats
```

**Endpoint público** (não requer autenticação).

**Resposta**:
```json
{
  "average_rating": 4.5,
  "total_ratings": 120,
  "distribution": {
    "1": 5,
    "2": 10,
    "3": 15,
    "4": 30,
    "5": 60
  }
}
```

### 6. Deletar Avaliação (Admin)

```http
DELETE /admin/ratings/{rating_id}
Authorization: Bearer {admin_token}
```

Permite que administradores deletem qualquer avaliação (moderação).
- Requer role "admin"
- Soft delete

## Modificações em Receitas

### GetRecipe

Agora inclui automaticamente os campos `average_rating` e `rating_count`:

```json
{
  "id": 5,
  "title": "Bolo de Chocolate",
  "average_rating": 4.5,
  "rating_count": 120,
  ...
}
```

### ListRecipes

Suporte a ordenação por rating:

```http
GET /recipes?sort_by=rating
```

Ordena receitas pela melhor avaliação média (em caso de empate, usa quantidade de avaliações e data).

## Segurança e Validações

### Autenticação
- Criar, editar e deletar avaliações: **requer autenticação**
- Listar avaliações e ver estatísticas: **público**

### Validações
- **Score**: obrigatório, entre 1 e 5
- **Comment**: opcional, máximo 1000 caracteres
- **Unicidade**: um usuário só pode ter uma avaliação por receita (enforced no DB)

### Rate Limiting
- Endpoints de leitura: limite de leitura configurado
- Endpoints de escrita: limite de escrita configurado
- Seguem a configuração global do sistema

### Autorização
- Usuários só podem deletar/editar suas próprias avaliações
- Admins podem deletar qualquer avaliação (moderação)

### Integridade Referencial
- Deletar receita: cascade delete em todas as avaliações
- Deletar usuário: cascade delete em todas as avaliações do usuário

## Casos de Uso

### 1. Usuário Avalia Receita pela Primeira Vez
1. Usuário faz login
2. Navega para uma receita
3. Envia POST com score e comentário opcional
4. Sistema cria nova avaliação
5. Receita atualiza média e contagem automaticamente

### 2. Usuário Edita sua Avaliação
1. Usuário já tem avaliação na receita
2. Envia POST novamente com novos valores
3. Sistema detecta avaliação existente
4. Atualiza score e/ou comentário
5. Estatísticas são recalculadas

### 3. Admin Modera Avaliação Imprópria
1. Admin identifica avaliação inadequada
2. Faz DELETE em `/admin/ratings/{rating_id}`
3. Sistema faz soft delete da avaliação
4. Estatísticas são recalculadas automaticamente

### 4. Buscar Receitas Mais Bem Avaliadas
1. Cliente solicita `/recipes?sort_by=rating`
2. Sistema ordena por média de avaliação
3. Receitas sem avaliação aparecem por último
4. Em caso de empate, usa quantidade de avaliações

## Performance

### Índices Criados
- `idx_ratings_recipe_id`: queries por receita
- `idx_ratings_user_id`: queries por usuário
- `idx_ratings_deleted_at`: soft delete
- `idx_ratings_score`: filtros e ordenação por score
- `idx_ratings_recipe_created`: listagem ordenada por data

### Otimizações
- Cálculo de estatísticas usa queries agregadas (AVG, COUNT)
- Campos calculados no Recipe não são salvos no DB
- Soft delete preserva histórico sem comprometer queries
- Índice composto para constraint de unicidade

## Testes

**Arquivo**: `test/rating_test.go`

### Testes Implementados:
1. ✅ Criação de avaliação
2. ✅ Atualização de avaliação (upsert)
3. ✅ Validações (score inválido, comentário muito longo)
4. ✅ Obter avaliação do usuário
5. ✅ Deletar avaliação
6. ✅ Listagem com paginação
7. ✅ Estatísticas e distribuição
8. ✅ Receitas com ratings calculados
9. ✅ Moderação admin
10. ✅ Ordenação de avaliações
11. ✅ Ordenação de receitas por rating

### Executar Testes

```bash
# Todos os testes de rating
go test ./test -run TestRating

# Teste específico
go test ./test -run TestRatingCreation -v

# Com coverage
go test ./test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Próximas Melhorias (Futuro)

- [ ] Sistema de notificações (quando alguém avalia sua receita)
- [ ] "Foi útil" em avaliações (helpful votes)
- [ ] Filtrar avaliações por score na listagem
- [ ] Reportar avaliações abusivas
- [ ] Resposta do autor da receita aos comentários
- [ ] Verificar se usuário fez a receita (verified reviewer)
- [ ] Fotos nas avaliações
- [ ] Timeline de avaliações no perfil do usuário

## Exemplos de Integração

### Frontend - Criar Avaliação

```typescript
async function rateRecipe(recipeId: number, score: number, comment?: string) {
  const response = await api.post(`/recipes/${recipeId}/ratings`, {
    score,
    comment
  });
  return response.data;
}
```

### Frontend - Obter Estatísticas

```typescript
async function getRecipeStats(recipeId: number) {
  const response = await api.get(`/recipes/${recipeId}/ratings/stats`);
  return response.data; // { average_rating, total_ratings, distribution }
}
```

### Frontend - Listar Avaliações

```typescript
async function getRecipeRatings(recipeId: number, page = 1, sort = 'newest') {
  const response = await api.get(`/recipes/${recipeId}/ratings`, {
    params: { page, limit: 10, sort }
  });
  return response.data;
}
```

## Troubleshooting

### Erro: "Você já avaliou esta receita"
- **Causa**: Constraint de unicidade
- **Solução**: Use POST novamente para atualizar (upsert)

### Erro: Score inválido
- **Causa**: Score fora do range 1-5
- **Solução**: Garantir validação no frontend

### Média não atualiza
- **Causa**: Soft delete não considerado na query
- **Solução**: Sempre usar `WHERE deleted_at IS NULL`

### Performance lenta ao listar receitas por rating
- **Causa**: Muitas receitas, subquery complexa
- **Solução**: Implementar cache de estatísticas (futuro)

## Arquivos Criados/Modificados

### Novos Arquivos
- ✅ `internal/models/rating.go`
- ✅ `internal/http/handlers/rating.go`
- ✅ `migrations/002_create_ratings_table.sql`
- ✅ `test/rating_test.go`
- ✅ `RATING_SYSTEM_IMPLEMENTATION.md`

### Arquivos Modificados
- ✅ `internal/models/recipe.go` (campos AverageRating, RatingCount)
- ✅ `internal/http/handlers/recipe.go` (cálculo de stats, ordenação)
- ✅ `internal/http/routes/routes.go` (novas rotas)

## Conclusão

O sistema de avaliações está completamente implementado e testado, pronto para uso em produção. Oferece uma experiência completa para usuários avaliarem receitas, com todas as funcionalidades modernas esperadas em um sistema de reviews.

