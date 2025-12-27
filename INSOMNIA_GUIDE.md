# 📮 Guia do Insomnia Collection

Este guia explica como usar o Insomnia collection da API de Receitas de forma eficiente.

## 🚀 Importar a Collection

1. Abra o Insomnia
2. Clique em **Create** → **Import**
3. Selecione o arquivo `insomnia-collection.json`
4. A collection "Receitas API" será importada

## 🌍 Ambientes (Environments)

A collection possui **3 environments**:

### 1. Base Environment (Padrão)

Contém as variáveis compartilhadas entre todos os ambientes:

```json
{
  "auth_token": "", // Token JWT de usuário (auto-preenchido)
  "admin_token": "", // Token JWT de admin (auto-preenchido)
  "user_email": "user@example.com",
  "user_password": "senha123",
  "admin_email": "admin@receitas.com",
  "admin_password": "admin123"
}
```

### 2. Local

Para desenvolvimento local:

```json
{
  "base_url": "http://localhost:8080"
}
```

### 3. Production

Para o ambiente de produção:

```json
{
  "base_url": "https://receitas-back-production.up.railway.app"
}
```

## 🔄 Como Trocar de Ambiente

1. No canto superior esquerdo, clique no dropdown de environment
2. Selecione **Local** ou **Production**
3. Todas as requests usarão a `base_url` correspondente

## 🔐 Autenticação Automática

### Passo 1: Login User

1. Navegue até **0. Setup & Auth** → **Login User (auto-save token)**
2. Execute a request (Ctrl/Cmd + Enter)
3. ✅ O token será **salvo automaticamente** em `auth_token`
4. Todas as requests autenticadas usarão este token

**Como funciona:**

- A request usa as variáveis `user_email` e `user_password`
- Script post-response captura o token da resposta
- Token é salvo em `auth_token` automaticamente

### Passo 2: Login Admin

1. Navegue até **0. Setup & Auth** → **Login Admin (auto-save token)**
2. Execute a request
3. ✅ O token será **salvo automaticamente** em `admin_token`
4. Todas as requests admin usarão este token

## 📁 Organização das Pastas

### 0. Setup & Auth

**Comece aqui!** Execute os logins para configurar os tokens.

- ✅ Login User (auto-save token) - **Execute primeiro**
- ✅ Login Admin (auto-save token) - **Para requests admin**
- Register User
- Logout

### 1. Health & Test

Endpoints públicos de verificação:

- Health Check
- Test Endpoint

### 2. Recipes (Public)

Endpoints **sem autenticação**:

- List Recipes
- List Recipes (Paginated)
- List Recipes (Filtered)
- Get Recipe by ID

### 3. Recipes (Authenticated)

Requer `auth_token` (execute Login User primeiro):

- Create Recipe
- Update Recipe
- Delete Recipe

### 4. Admin - Recipes

Requer `admin_token` (execute Login Admin primeiro):

- Admin List All Recipes
- Admin Create General Recipe
- Admin Update Any Recipe
- Admin Delete Any Recipe

### 5. Ingredients (Public)

Endpoints **sem autenticação**:

- List Ingredients
- Search Ingredients
- Search with Category Filter
- Get Ingredient by ID
- Get Categories

### 6. Recipe Ingredients

Gerenciar ingredientes em receitas:

- List Recipe Ingredients (público)
- Add Ingredient to Recipe (requer auth)
- Update Recipe Ingredient (requer auth)
- Delete Recipe Ingredient (requer auth)
- Get Recipe Nutrition (público)

### 7. Admin - Ingredients

Requer `admin_token`:

- Create Ingredient
- Update Ingredient
- Delete Ingredient

### 8. Rate Limit Tests

Testar limites de requisições:

- Test Global Rate Limit (100 req/min)
- Test Read Rate Limit (60 req/min)
- Test Write Rate Limit (20 req/min)

## 💡 Workflows Comuns

### Criar Receita como Usuário

```
1. Login User (auto-save token)
2. Create Recipe
   → Usa automaticamente {{ _.auth_token }}
```

### Criar Receita Geral como Admin

```
1. Login Admin (auto-save token)
2. Admin Create General Recipe
   → Usa automaticamente {{ _.admin_token }}
```

### Buscar Ingredientes

```
1. Search Ingredients
   → Parâmetro: search=arroz
   → Busca com ranking de relevância
```

### Adicionar Ingredientes a uma Receita

```
1. Login User (auto-save token)
2. Create Recipe
3. Add Ingredient to Recipe
   → Use o ID da receita criada
4. Get Recipe Nutrition
   → Veja o cálculo nutricional automático
```

### Testar Admin em Produção

```
1. Selecione environment: Production
2. Login Admin (auto-save token)
3. Admin List All Recipes
   → Veja todas as receitas com info de usuário
```

## 🔧 Personalizar Credenciais

Para usar credenciais diferentes:

1. Abra **Manage Environments** (Ctrl/Cmd + E)
2. Selecione **Base Environment**
3. Edite as variáveis:

```json
{
  "user_email": "seu-email@example.com",
  "user_password": "sua-senha",
  "admin_email": "seu-admin@example.com",
  "admin_password": "senha-admin"
}
```

4. Salve (Ctrl/Cmd + S)
5. Execute Login User ou Login Admin novamente

## 🎯 Benefícios desta Collection

✅ **Tokens Automáticos** - Sem copiar/colar manualmente  
✅ **Multi-Ambiente** - Troque entre Local/Prod em 1 clique  
✅ **Variáveis Reutilizáveis** - Configure uma vez, use em todas requests  
✅ **Organização Clara** - Pastas separadas por funcionalidade  
✅ **Segurança** - Tokens em variáveis privadas  
✅ **Produtividade** - Scripts post-response automatizam tudo

## 🐛 Troubleshooting

### Token não está sendo salvo automaticamente

**Solução:**

1. Verifique se o login retornou status 200
2. Veja o **Console** no Insomnia (⌘/Ctrl + `)
3. Deve aparecer: `✅ Token de usuário salvo em auth_token`

### Request retorna 401 Unauthorized

**Solução:**

1. Execute **Login User** ou **Login Admin** novamente
2. Tokens expiram em 24 horas
3. Verifique se selecionou o environment correto

### Variável {{ _.auth_token }} aparece vazia

**Solução:**

1. Abra **Manage Environments** (⌘/Ctrl + E)
2. Verifique se `auth_token` tem valor
3. Execute Login User/Admin novamente

### Endpoint não encontrado (404)

**Solução:**

1. Verifique se selecionou o environment correto
2. Local → API deve estar rodando em `localhost:8080`
3. Production → Verifique se a URL está correta

## 📚 Recursos Adicionais

- [Documentação da API](README.md)
- [Autenticação JWT](AUTHENTICATION_IMPLEMENTATION.md)
- [Sistema de Admin](ADMIN_SYSTEM_IMPLEMENTATION.md)
- [Ingredientes](INGREDIENTS_IMPLEMENTATION.md)
- [Rate Limiting](RATE_LIMIT_IMPLEMENTATION.md)

## 🎓 Dicas Avançadas

### 1. Usar Request Runner

Para testar múltiplas requests em sequência:

1. Clique com botão direito em uma pasta
2. **Run**
3. Execute todos os endpoints automaticamente

### 2. Variables Scope

- **Base Environment** → Compartilhado entre Local e Production
- **Local/Production** → Sobrescreve Base quando selecionado

### 3. Scripts Customizados

Você pode adicionar scripts em qualquer request:

```javascript
// Pre-request Script
console.log("Enviando request...");

// Post-response Script
const response = await insomnia.response.json();
console.log("Resposta:", response);
```

### 4. Exportar Environment

Para compartilhar configurações:

1. **Manage Environments**
2. Selecione o environment
3. **Export** → JSON
4. Compartilhe com o time

## 🚀 Workflow Recomendado

**Desenvolvimento Local:**

```
1. Selecionar environment: Local
2. Garantir que API está rodando (go run ./cmd/api)
3. Login User → Testar endpoints
4. Login Admin → Testar admin endpoints
```

**Testes em Produção:**

```
1. Selecionar environment: Production
2. Login User/Admin
3. Testar funcionalidades
4. Verificar rate limits
```

---

**Dúvidas?** Consulte a [documentação completa](README.md) ou abra uma issue no projeto.
