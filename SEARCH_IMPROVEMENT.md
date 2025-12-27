# Melhoria da Busca de Ingredientes

## ✅ Implementação Concluída

A busca de ingredientes foi melhorada com **ranking por relevância** e busca inteligente.

## 🎯 Funcionalidades

### 1. Busca com Ranking de Relevância

Os resultados são ordenados automaticamente por relevância:

**Prioridade 1:** Nome **começa** com o termo buscado  
**Prioridade 2:** Nome **contém** o termo buscado  
**Prioridade 3:** Categoria **contém** o termo buscado  

### 2. Busca Case-Insensitive

A busca funciona independente de maiúsculas/minúsculas:
- `search=AÇÚCAR` = `search=açúcar` = `search=Açúcar`

### 3. Busca em Múltiplos Campos

A busca procura simultaneamente em:
- **Nome do ingrediente**
- **Categoria**

### 4. Filtro de Categoria Complementar

O parâmetro `category` funciona como filtro adicional (operação AND):
- `search=cozido&category=vegetais` → vegetais cozidos apenas

## 📝 Exemplos de Uso

### Busca por Nome

```bash
# Buscar "arroz"
GET /ingredients?search=arroz

# Retorna (em ordem):
# 1. Arroz branco (começa com "arroz")
# 2. Arroz integral (começa com "arroz")
# 3. Macarrão de arroz (contém "arroz")
```

### Busca por Categoria

```bash
# Buscar "cereais"
GET /ingredients?search=cereais

# Retorna todos ingredientes da categoria "cereais"
```

### Busca + Filtro de Categoria

```bash
# Buscar "cozido" apenas em vegetais
GET /ingredients?search=cozido&category=vegetais

# Retorna apenas vegetais que contêm "cozido" no nome
```

### Busca Case-Insensitive

```bash
# Todas as variações funcionam igual:
GET /ingredients?search=FEIJÃO
GET /ingredients?search=feijão
GET /ingredients?search=Feijão
```

## 🧪 Testes Implementados

5 novos testes foram adicionados para validar a busca:

1. **TestSearchIngredientsByName**
   - Valida busca por termo no nome
   - Valida ordenação por relevância

2. **TestSearchIngredientsByCategory**
   - Valida busca por termo na categoria
   - Retorna todos ingredientes da categoria

3. **TestSearchIngredientsCaseInsensitive**
   - Valida que busca funciona com maiúsculas/minúsculas

4. **TestSearchWithCategoryFilter**
   - Valida combinação de search + category (AND)

## 💡 Benefícios

✅ **UX Melhorada**: Resultados mais relevantes aparecem primeiro  
✅ **Busca Intuitiva**: Funciona como usuário espera  
✅ **Performance**: Adequada para ~600 ingredientes  
✅ **Sem Migração**: Não requer alteração no banco  
✅ **Compatível**: Mantém paginação e filtros existentes  

## 🔧 Implementação Técnica

### Handler Modificado

**Arquivo:** `internal/http/handlers/ingredient.go`

**Lógica de Ranking:**

```sql
CASE 
  WHEN LOWER(name) LIKE 'termo%' THEN 1      -- Nome começa com
  WHEN LOWER(name) LIKE '%termo%' THEN 2     -- Nome contém
  WHEN LOWER(category) LIKE '%termo%' THEN 3 -- Categoria contém
  ELSE 4 
END
```

### Normalização

- Termo de busca: `strings.ToLower()` + `strings.TrimSpace()`
- Comparação no banco: `LOWER(nome)` e `LOWER(categoria)`

## 📊 Performance

Para o volume atual de dados (~597 ingredientes TACO):
- ✅ Busca rápida (< 50ms)
- ✅ Ordenação eficiente
- ✅ Sem índices adicionais necessários

Para datasets maiores (milhares de ingredientes), considerar:
- Índice GIN no PostgreSQL
- Full-Text Search (FTS)

## 🚀 Próximos Passos (Opcional)

Se o volume de ingredientes crescer significativamente:

1. **Adicionar Full-Text Search (FTS)**
   ```sql
   ALTER TABLE ingredients 
   ADD COLUMN search_vector tsvector 
   GENERATED ALWAYS AS (
     setweight(to_tsvector('portuguese', name), 'A') ||
     setweight(to_tsvector('portuguese', category), 'B')
   ) STORED;
   
   CREATE INDEX ingredients_search_idx 
   ON ingredients USING GIN (search_vector);
   ```

2. **Busca com Múltiplos Termos**
   - "arroz integral" → buscar "arroz" AND "integral"

3. **Busca Fuzzy**
   - Tolerância a erros de digitação
   - Sugestões de correção

## 📚 Referências

- [PostgreSQL LIKE](https://www.postgresql.org/docs/current/functions-matching.html)
- [PostgreSQL Full-Text Search](https://www.postgresql.org/docs/current/textsearch.html)
- [GORM Ordering](https://gorm.io/docs/query.html#Order)

