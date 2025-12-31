# Testes de Recipe Ingredient Handlers

## 📋 Visão Geral

Este documento descreve os testes criados para validar a refatoração dos handlers de ingredientes de receitas, que agora utilizam DTOs (Data Transfer Objects) em vez de validar diretamente os modelos do banco de dados.

## 🎯 Objetivo dos Testes

Garantir que:

1. ✅ A validação ocorre apenas nos campos da API (não em relações)
2. ✅ DTOs são corretamente mapeados para modelos
3. ✅ Validações de negócio funcionam (ownership, ingrediente existente, etc.)
4. ✅ Atualizações parciais preservam campos não enviados

## 📝 Testes Implementados

### `TestAddRecipeIngredient_Success`

**Objetivo**: Validar criação bem-sucedida de ingrediente em receita

**Cenário**:

- Usuário autenticado cria receita
- Adiciona ingrediente válido com todos os campos
- Verifica que os dados foram salvos corretamente

**Validações**:

- Status 201 Created
- Campos retornados correspondem aos enviados
- Relacionamento Recipe-Ingredient criado

---

### `TestAddRecipeIngredient_ValidationErrors`

**Objetivo**: Validar erros de validação do DTO

**Cenários testados**:

1. **Quantidade negativa** → Status 400
2. **Quantidade zero** → Status 400
3. **Sem ingredient_id** → Status 400
4. **Sem unit** → Status 400

**Validações**:

- Todas retornam Status 400 Bad Request
- Mensagens de erro apropriadas

---

### `TestAddRecipeIngredient_IngredientNotFound`

**Objetivo**: Validar erro quando ingrediente não existe

**Cenário**:

- Tenta adicionar ingrediente com ID inexistente (999)

**Validações**:

- Status 400 Bad Request
- Mensagem "Ingrediente não encontrado"

---

### `TestAddRecipeIngredient_Unauthorized`

**Objetivo**: Validar controle de acesso (ownership)

**Cenário**:

- Usuário A cria receita
- Usuário B tenta adicionar ingrediente

**Validações**:

- Status 403 Forbidden
- Mensagem "You don't have permission to modify this recipe"

---

### `TestUpdateRecipeIngredient_Success`

**Objetivo**: Validar atualização completa de ingrediente

**Cenário**:

- Atualiza quantity, unit e notes simultaneamente

**Validações**:

- Status 200 OK
- Todos os campos atualizados corretamente

---

### `TestUpdateRecipeIngredient_PartialUpdate`

**Objetivo**: Validar atualização parcial (apenas alguns campos)

**Cenário**:

- Atualiza apenas `quantity`
- Outros campos (unit, notes) devem permanecer inalterados

**Validações**:

- Status 200 OK
- Apenas campo enviado foi atualizado
- Campos não enviados mantêm valores originais

---

### `TestUpdateRecipeIngredient_InvalidQuantity`

**Objetivo**: Validar erro de validação em atualização

**Cenário**:

- Tenta atualizar com quantidade negativa

**Validações**:

- Status 400 Bad Request
- Validação do DTO funciona em updates

---

## 🔧 Estrutura dos Testes

### Setup

```go
testdb.SetupWithCleanup(t) // Banco in-memory + cleanup automático
```

### Helpers Utilizados

- `testdb.SeedUser()` - Cria usuário de teste
- `testdb.SeedRecipe()` - Cria receita de teste
- `testdb.SeedIngredient()` - Cria ingrediente de teste
- `testdb.AddChiURLParam()` - Adiciona parâmetros de URL (Chi router)

### Padrão de Contexto

```go
// 1. Adicionar UserID ao contexto
ctx := context.WithValue(req.Context(), middleware.UserIDKey, user.ID)
req = req.WithContext(ctx)

// 2. Adicionar parâmetros de URL do Chi
ctx = testdb.AddChiURLParam(req, "id", fmt.Sprint(recipe.ID))
req = req.WithContext(ctx)
```

⚠️ **Importante**: A ordem importa! O contexto deve ser atualizado após cada adição.

## 📊 Cobertura

### Handlers Testados

- ✅ `AddRecipeIngredient` (5 testes)
- ✅ `UpdateRecipeIngredient` (3 testes)

### Cenários Cobertos

- ✅ Sucesso (happy path)
- ✅ Validações de entrada (DTO)
- ✅ Validações de negócio (ingrediente existe)
- ✅ Controle de acesso (ownership)
- ✅ Atualizações parciais

### Não Cobertos (handlers existentes)

- ⏭️ `ListRecipeIngredients` (já testado indiretamente)
- ⏭️ `DeleteRecipeIngredient` (pode ser adicionado)
- ⏭️ `GetRecipeNutrition` (pode ser adicionado)

## 🚀 Executando os Testes

### Todos os testes de recipe ingredient

```bash
cd /Users/davidsonmarra/receitas-back
go test -v ./test -run "TestAddRecipeIngredient|TestUpdateRecipeIngredient" -count=1
```

### Teste específico

```bash
go test -v ./test -run TestAddRecipeIngredient_Success -count=1
```

### Todos os testes do projeto

```bash
go test -v ./test -count=1
```

## ✅ Resultado

```
=== RUN   TestAddRecipeIngredient_Success
--- PASS: TestAddRecipeIngredient_Success (0.26s)
=== RUN   TestAddRecipeIngredient_ValidationErrors
--- PASS: TestAddRecipeIngredient_ValidationErrors (0.26s)
=== RUN   TestAddRecipeIngredient_IngredientNotFound
--- PASS: TestAddRecipeIngredient_IngredientNotFound (0.26s)
=== RUN   TestAddRecipeIngredient_Unauthorized
--- PASS: TestAddRecipeIngredient_Unauthorized (0.26s)
=== RUN   TestUpdateRecipeIngredient_Success
--- PASS: TestUpdateRecipeIngredient_Success (0.26s)
=== RUN   TestUpdateRecipeIngredient_PartialUpdate
--- PASS: TestUpdateRecipeIngredient_PartialUpdate (0.26s)
=== RUN   TestUpdateRecipeIngredient_InvalidQuantity
--- PASS: TestUpdateRecipeIngredient_InvalidQuantity (0.26s)
PASS
ok  	github.com/davidsonmarra/receitas-app/test	2.478s
```

**8 testes, 100% de sucesso** ✅

## 📚 Aprendizados

### 1. DTOs vs Modelos

- **Antes**: Validação no modelo causava erros em relações vazias
- **Depois**: DTOs validam apenas campos da API

### 2. Contexto do Chi Router

- Contexto deve ser atualizado após cada modificação
- `AddChiURLParam` preserva contexto anterior

### 3. Atualizações Parciais

- DTOs com ponteiros (`*float64`, `*string`) permitem distinguir "não enviado" de "zero"
- Aplicar apenas campos não-nil

## 🔗 Arquivos Relacionados

- **Handler**: `/Users/davidsonmarra/receitas-back/internal/http/handlers/recipe_ingredient.go`
- **Modelo**: `/Users/davidsonmarra/receitas-back/internal/models/recipe_ingredient.go`
- **Testes**: `/Users/davidsonmarra/receitas-back/test/recipe_ingredient_handler_test.go`
- **Test Helpers**: `/Users/davidsonmarra/receitas-back/test/testdb/setup.go`
