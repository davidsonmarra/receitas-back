# Backend Go - API Receitas

Backend em Go desenvolvido com arquitetura limpa e escalável.

## 📋 Descrição

Este projeto estabelece a fundação para um serviço backend escrito em Go. A Fase 1 implementa a infraestrutura core com um servidor HTTP mínimo, endpoints básicos, testes unitários e comandos Cursor para automação de desenvolvimento.

## 🔧 Tecnologias

- **Go**: ≥ 1.24
- **Database**: [PostgreSQL](https://www.postgresql.org/) - Database relacional
- **ORM**: [GORM](https://gorm.io/) v1.31+ - ORM completo para Go
- **Router**: [go-chi/chi](https://github.com/go-chi/chi) v5
- **CORS**: [go-chi/cors](https://github.com/go-chi/cors) - Cross-Origin Resource Sharing
- **Validator**: [go-playground/validator](https://github.com/go-playground/validator) v10 - Validação de structs
- **Logger**: [uber-go/zap](https://github.com/uber-go/zap) - Alta performance
- **UUID**: [google/uuid](https://github.com/google/uuid) - Geração de Request IDs
- **Testes**: testing + httptest

## 📁 Estrutura do Projeto

```
receitas-app/
├── cmd/api/                    # Executáveis
│   └── main.go                 # Entrypoint da aplicação
├── internal/                   # Código interno da aplicação
│   ├── models/                 # Modelos de dados
│   │   └── recipe.go           # Modelo Recipe (GORM)
│   ├── server/                 # Configuração do servidor
│   │   └── server.go
│   └── http/
│       ├── middleware/         # Middlewares HTTP
│       │   ├── requestid.go    # Middleware de Request ID
│       │   ├── requestsize.go  # Limite de tamanho de request
│       │   └── cors.go         # Configuração de CORS
│       ├── routes/             # Registro de rotas
│       │   └── routes.go
│       └── handlers/           # Handlers HTTP
│           ├── health.go       # Health check
│           ├── test.go         # Handler de teste
│           └── recipe.go       # CRUD de receitas
├── pkg/                        # Utilitários reutilizáveis
│   ├── database/               # Conexão com database
│   │   └── connection.go       # PostgreSQL + GORM
│   ├── validation/             # Sistema de validação
│   │   └── validator.go        # Validação com traduções PT-BR
│   ├── log/                    # Sistema de logging
│   │   ├── logger.go           # API de logging (estilo Android)
│   │   └── config.go           # Configuração do logger
│   └── response/
│       └── json.go             # Helpers para respostas JSON
├── test/                       # Testes unitários
│   ├── test_handler_test.go
│   ├── health_handler_test.go
│   ├── recipe_handler_test.go
│   └── logger_test.go
├── .cursor/commands/           # Comandos Cursor
│   ├── create-route.md
│   └── create-test.md
├── Dockerfile                  # Multi-stage build
├── .dockerignore
├── railway.toml               # Configuração Railway
├── .env.example               # Variáveis de ambiente
├── go.mod                     # Dependências
└── README.md
```

## 🚀 Como Executar

### Pré-requisitos

- Go 1.24 ou superior instalado

### Executar o servidor

```bash
go run ./cmd/api
```

O servidor será iniciado na porta **8080**.

Acesse: http://localhost:8080/test

### Resposta esperada

```json
{
  "message": "hello world"
}
```

### Configurar Variáveis de Ambiente

```bash
# Opcional: Definir nível de log (debug, info, warn, error)
export LOG_LEVEL=debug

# Opcional: Definir ambiente (development ou production)
export ENV=development

# Obrigatório para autenticação: Secret do JWT
export JWT_SECRET="sua-chave-secreta-muito-longa-e-aleatoria"

# Obrigatório: String de conexão PostgreSQL
export DATABASE_URL="postgres://usuario:senha@localhost:5432/receitas_db?sslmode=disable"

# Executar servidor
go run ./cmd/api
```

## 📊 Sistema de Logging

O projeto utiliza um sistema de logging profissional baseado em **zap** (Uber) com API estilo Android.

### API de Logging

```go
import "github.com/davidsonmarra/receitas-app/pkg/log"

// Logs básicos
log.Debug("debug message", "key", "value")
log.Info("info message", "key", "value")
log.Warn("warning message", "key", "value")
log.Error("error message", "error", err)

// Logs com contexto (inclui Request ID automaticamente)
log.DebugCtx(ctx, "processing request", "user_id", 123)
log.InfoCtx(ctx, "request completed", "duration_ms", 45)
log.WarnCtx(ctx, "slow query detected")
log.ErrorCtx(ctx, "operation failed", "error", err)
```

### Níveis de Log

Configure o nível através da variável `LOG_LEVEL`:

| Nível     | Variável          | O que mostra                        |
| --------- | ----------------- | ----------------------------------- |
| **debug** | `LOG_LEVEL=debug` | Tudo (debug, info, warn, error)     |
| **info**  | `LOG_LEVEL=info`  | info, warn, error (padrão produção) |
| **warn**  | `LOG_LEVEL=warn`  | warn, error                         |
| **error** | `LOG_LEVEL=error` | Somente erros                       |

### Formato de Saída

#### Desenvolvimento (ENV != production)

Logs formatados e coloridos para leitura humana:

```
2025-12-24T10:30:45.123Z    DEBUG   handling test request   {"request_id": "abc-123", "method": "GET", "path": "/test"}
2025-12-24T10:30:45.124Z    INFO    server starting         {"port": 8080, "address": ":8080"}
```

#### Produção (ENV = production)

JSON estruturado para agregadores de log:

```json
{"level":"info","timestamp":"2025-12-24T10:30:45.001Z","msg":"server starting","port":8080,"address":":8080"}
{"level":"info","timestamp":"2025-12-24T10:30:45.123Z","msg":"request completed","request_id":"abc-123","duration_ms":45}
```

### Request ID

Cada requisição HTTP recebe um **UUID único** automaticamente:

- Adicionado ao header de resposta: `X-Request-ID`
- Incluído automaticamente em logs com `*Ctx()` functions
- Útil para rastreamento distribuído e debugging

**Exemplo de resposta:**

```http
HTTP/1.1 200 OK
Content-Type: application/json
X-Request-ID: 550e8400-e29b-41d4-a716-446655440000

{"message":"hello world"}
```

### Vantagens

✅ **Performance**: zap é extremamente rápido (zero alocações)  
✅ **Estruturado**: JSON facilita parsing e agregação  
✅ **Rastreável**: Request ID em cada log  
✅ **Configurável**: Níveis de log por ambiente  
✅ **Familiar**: API estilo Android (`log.Debug`, `log.Info`, etc)

## 🧪 Como Testar

### Executar todos os testes

```bash
go test ./...
```

### Executar testes com verbose

```bash
go test -v ./...
```

### Executar testes de um pacote específico

```bash
go test ./test
```

## 🛠 Comandos Cursor

Este projeto inclui comandos Cursor para automatizar tarefas comuns:

### Create Route

Cria uma nova rota HTTP seguindo o padrão do projeto.

**Localização**: `.cursor/commands/create-route.md`

**Uso**: Execute o comando Cursor "Create Route" e forneça:

- Caminho da rota (ex: `/users`)
- Nome do handler (ex: `UsersHandler`)

### Create Test

Cria testes unitários para handlers HTTP.

**Localização**: `.cursor/commands/create-test.md`

**Uso**: Execute o comando Cursor "Create Test" e especifique o handler a ser testado.

## 📐 Princípios Arquiteturais

- `/cmd` → executáveis da aplicação
- `/internal` → lógica core da aplicação (não exportável)
- `/pkg` → utilitários reutilizáveis (exportáveis)
- Handlers são stateless e mínimos
- Separação clara de responsabilidades
- Código idiomático Go
- Sem estado global mutável

## 🌐 CORS (Cross-Origin Resource Sharing)

A API implementa CORS para permitir que aplicações web de diferentes domínios consumam a API.

### Configuração

O CORS é configurado automaticamente baseado no ambiente:

#### Development (`ENV != production`)
```
Permite origens:
- http://localhost:* (qualquer porta)
- http://127.0.0.1:*
- http://[::1]:*
```

#### Production (`ENV == production`)
```
Permite origens baseado em:
1. Variável CORS_ORIGINS (recomendado)
   Exemplo: CORS_ORIGINS="https://app.com,https://admin.app.com"

2. Padrão: https://* (qualquer origem HTTPS)
```

### Headers Configurados

| Header | Valor | Descrição |
|--------|-------|-----------|
| `Access-Control-Allow-Origin` | Baseado em config | Origem permitida |
| `Access-Control-Allow-Methods` | GET, POST, PUT, DELETE, OPTIONS | Métodos HTTP permitidos |
| `Access-Control-Allow-Headers` | Accept, Authorization, Content-Type, X-Request-ID | Headers aceitos |
| `Access-Control-Expose-Headers` | X-Request-ID | Headers expostos ao client |
| `Access-Control-Allow-Credentials` | false | Cookies não permitidos |
| `Access-Control-Max-Age` | 300 | Cache de preflight (5 min) |

### React Native

**Importante**: Apps React Native **nativos** (iOS/Android) **não precisam de CORS** pois não rodam em navegador. CORS só se aplica a:
- React Native Web
- Expo Web
- Aplicações web que consumem a API

### Testar CORS

#### Com curl (simular preflight):
```bash
curl -H "Origin: http://localhost:3000" \
     -H "Access-Control-Request-Method: POST" \
     -H "Access-Control-Request-Headers: Content-Type" \
     -X OPTIONS \
     https://receitas-back-production.up.railway.app/recipes -v
```

#### Resposta esperada:
```
< HTTP/2 200
< access-control-allow-origin: http://localhost:3000
< access-control-allow-methods: POST
< access-control-allow-headers: Content-Type
< access-control-max-age: 300
```

#### No navegador:
```javascript
fetch('https://receitas-back-production.up.railway.app/recipes')
  .then(res => res.json())
  .then(data => console.log('✅ CORS funcionando!', data))
  .catch(err => console.error('❌ Erro:', err))
```

### Configurar para Produção

Adicione a variável de ambiente no Railway:

```bash
CORS_ORIGINS=https://seu-frontend.vercel.app,https://seu-dominio.com
```

**Atenção**: Nunca use `*` em produção com `AllowCredentials: true`.

## 📄 Paginação

A API implementa paginação reutilizável em todos os endpoints que retornam listas, otimizada para performance em apps móveis.

### Como Usar

Adicione os parâmetros `page` e `limit` na query string:

```bash
GET /recipes?page=1&limit=20
```

### Parâmetros

| Parâmetro | Tipo | Padrão | Mín | Máx | Descrição |
|-----------|------|--------|-----|-----|-----------|
| `page` | int | 1 | 1 | ∞ | Número da página |
| `limit` | int | 20 | 1 | 100 | Itens por página |

### Validação Automática

A API valida e corrige automaticamente parâmetros inválidos:

| Entrada | Corrigido para | Motivo |
|---------|----------------|--------|
| `?page=0` | `page=1` | Mínimo é 1 |
| `?page=-5` | `page=1` | Mínimo é 1 |
| `?limit=0` | `limit=20` | Mínimo é 1 |
| `?limit=500` | `limit=100` | Máximo é 100 |
| `?page=abc` | `page=1` | Inválido, usa padrão |

### Formato de Resposta

Todas as respostas paginadas seguem o mesmo formato:

```json
{
  "data": [
    {
      "id": 1,
      "title": "Bolo de Chocolate",
      "description": "Delicioso bolo",
      "prep_time": 60,
      "servings": 8,
      "difficulty": "média",
      "created_at": "2025-12-24T10:30:45Z",
      "updated_at": "2025-12-24T10:30:45Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8,
    "has_next": true,
    "has_prev": false
  }
}
```

### Metadata de Paginação

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `page` | int | Página atual |
| `limit` | int | Itens por página |
| `total` | int64 | Total de registros |
| `total_pages` | int | Total de páginas |
| `has_next` | bool | Tem próxima página? |
| `has_prev` | bool | Tem página anterior? |

### Exemplos

#### Primeira página (padrão)
```bash
GET /recipes
# Retorna 20 primeiros itens
```

#### Segunda página
```bash
GET /recipes?page=2&limit=10
# Retorna itens 11-20 (10 por página)
```

#### Limite customizado
```bash
GET /recipes?page=1&limit=50
# Retorna 50 primeiros itens
```

### Uso no React Native

#### Scroll Infinito

```javascript
const [recipes, setRecipes] = useState([]);
const [page, setPage] = useState(1);
const [hasNext, setHasNext] = useState(true);
const [loading, setLoading] = useState(false);

const loadMore = async () => {
  if (!hasNext || loading) return;
  
  setLoading(true);
  try {
    const response = await fetch(
      `${API_URL}/recipes?page=${page}&limit=20`
    );
    const data = await response.json();
    
    setRecipes([...recipes, ...data.data]);
    setHasNext(data.pagination.has_next);
    setPage(page + 1);
  } catch (error) {
    console.error('Erro ao carregar receitas:', error);
  } finally {
    setLoading(false);
  }
};

// No FlatList
<FlatList
  data={recipes}
  onEndReached={loadMore}
  onEndReachedThreshold={0.5}
/>
```

#### Pull to Refresh

```javascript
const [refreshing, setRefreshing] = useState(false);

const onRefresh = async () => {
  setRefreshing(true);
  try {
    const response = await fetch(`${API_URL}/recipes?page=1&limit=20`);
    const data = await response.json();
    
    setRecipes(data.data);
    setPage(1);
    setHasNext(data.pagination.has_next);
  } catch (error) {
    console.error('Erro ao atualizar:', error);
  } finally {
    setRefreshing(false);
  }
};

<FlatList
  data={recipes}
  refreshing={refreshing}
  onRefresh={onRefresh}
/>
```

### Performance

#### Otimizações Implementadas

1. **Queries Separadas**
   - Count query otimizada (sem SELECT *)
   - Data query com LIMIT/OFFSET
   
2. **Índice em created_at**
   - Ordenação rápida (< 10ms)
   - Funciona mesmo com milhares de registros

3. **Limit máximo de 100**
   - Previne requests gigantes
   - Protege memória e bandwidth

4. **Default baixo (20 itens)**
   - Ideal para scroll infinito
   - Menos dados transferidos

#### Benchmark Esperado

| Cenário | Tempo Estimado |
|---------|----------------|
| 100 receitas, page 1 | < 50ms |
| 10.000 receitas, page 1 | < 100ms |
| 10.000 receitas, page 500 | < 150ms |

### Reutilização

Para adicionar paginação em qualquer endpoint futuro:

```go
func ListUsers(w http.ResponseWriter, r *http.Request) {
    // 1. Extrair parâmetros
    params := pagination.ExtractParams(r)
    
    // 2. Count total
    var total int64
    database.DB.Model(&models.User{}).Count(&total)
    
    // 3. Buscar dados paginados
    var users []models.User
    offset := pagination.CalculateOffset(params)
    database.DB.Limit(params.Limit).Offset(offset).Find(&users)
    
    // 4. Retornar resposta paginada
    response.Paginated(w, http.StatusOK, users, params, total)
}
```

**Apenas 3 linhas de código!** ✅

## ✅ Validação de Inputs

A API implementa validação robusta de dados de entrada usando `validator/v10` com mensagens amigáveis em português, projetadas para serem exibidas diretamente no frontend.

### Formato de Erro

Todos os erros de validação retornam a seguinte estrutura (apenas o **primeiro erro** encontrado):

```json
{
  "error": {
    "title": "Ops, algo deu errado!",
    "message": "O título é obrigatório."
  }
}
```

**Status**: 400 Bad Request  
**Content-Type**: application/json

> **Nota**: Se múltiplos campos forem inválidos, apenas o primeiro erro será retornado. Corrija-o e envie novamente para ver o próximo erro, se houver.

### Regras de Validação

#### Recipe (Receita)

| Campo | Obrigatório | Regras | Descrição |
|-------|-------------|--------|-----------|
| `title` | ✅ Sim | 3-200 caracteres | Título da receita |
| `description` | ❌ Não | Texto livre | Descrição detalhada |
| `prep_time` | ✅ Sim | Mínimo: 1 minuto | Tempo de preparo |
| `servings` | ✅ Sim | Mínimo: 1 porção | Número de porções |
| `difficulty` | ❌ Não | `fácil`, `média`, `difícil` | Nível de dificuldade |

### Proteção de Campos

No **UPDATE** (`PUT /recipes/{id}`), os seguintes campos são **protegidos** e não podem ser modificados:

- `id` - ID da receita
- `created_at` - Data de criação
- `updated_at` - Data de atualização (gerenciada automaticamente)
- `deleted_at` - Data de exclusão (soft delete)

### Limite de Requisição

- **Tamanho máximo do body**: 1MB
- **Timeout**: 15 segundos

Se o body exceder 1MB, a API retorna:

```json
{
  "error": {
    "title": "Ops, algo deu errado!",
    "message": "A requisição é muito grande. Limite: 1MB."
  }
}
```

### Exemplos de Erros

#### Campo obrigatório ausente

```bash
POST /recipes
{
  "prep_time": 30,
  "servings": 4
}
```

**Resposta**:

```json
{
  "error": {
    "title": "Ops, algo deu errado!",
    "message": "O título é obrigatório."
  }
}
```

#### Múltiplos campos inválidos

```bash
POST /recipes
{
  "title": "AB",
  "prep_time": 0,
  "servings": -1
}
```

**Resposta** (retorna apenas o primeiro erro):

```json
{
  "error": {
    "title": "Ops, algo deu errado!",
    "message": "O título deve ter no mínimo 3 caracteres."
  }
}
```

Após corrigir o título e enviar novamente, o próximo erro será exibido (prep_time).

#### Valor inválido

```bash
POST /recipes
{
  "title": "Bolo de Chocolate",
  "prep_time": 30,
  "servings": 4,
  "difficulty": "impossível"
}
```

**Resposta**:

```json
{
  "error": {
    "title": "Ops, algo deu errado!",
    "message": "A dificuldade deve ser uma das opções: fácil, média, difícil."
  }
}
```

### Implementação

A validação é realizada em três camadas:

1. **Middleware de Request Size** - Limita tamanho do body antes de processar
2. **Validação Estrutural** - Verifica tipos e formatos JSON
3. **Validação de Negócio** - Aplica regras de negócio (mínimos, máximos, opções)

Pacote: [`pkg/validation`](pkg/validation/validator.go)

## 🔐 Autenticação JWT

A API utiliza **JSON Web Tokens (JWT)** para autenticação de usuários. Tokens expiram em 24 horas e podem ser invalidados através do logout.

### Configuração

Defina a variável de ambiente `JWT_SECRET` com uma string longa e aleatória:

```bash
# Desenvolvimento
export JWT_SECRET="desenvolvimento-secret-nao-usar-em-producao-12345"

# Produção (Railway)
railway variables set JWT_SECRET="$(openssl rand -base64 32)"
```

**⚠️ IMPORTANTE**: Use um secret forte e único em produção. Nunca compartilhe ou commite o JWT_SECRET!

### Endpoints de Autenticação

#### POST /users/register

Cadastra um novo usuário e retorna token automaticamente.

**Request**:
```json
{
  "name": "João Silva",
  "email": "joao@example.com",
  "password": "senha123"
}
```

**Response** (201 Created):
```json
{
  "user": {
    "id": 1,
    "name": "João Silva",
    "email": "joao@example.com",
    "created_at": "2025-12-26T10:00:00Z",
    "updated_at": "2025-12-26T10:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Validações**:
- Nome: mínimo 3 caracteres, máximo 100
- E-mail: formato válido, único no sistema
- Senha: mínimo 6 caracteres

#### POST /users/login

Autentica um usuário e retorna token.

**Request**:
```json
{
  "email": "joao@example.com",
  "password": "senha123"
}
```

**Response** (200 OK):
```json
{
  "user": {
    "id": 1,
    "name": "João Silva",
    "email": "joao@example.com",
    "created_at": "2025-12-26T10:00:00Z",
    "updated_at": "2025-12-26T10:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Erro** (401 Unauthorized):
```json
{
  "error": {
    "title": "Ops, algo deu errado!",
    "message": "E-mail ou senha inválidos"
  }
}
```

#### POST /users/logout

Invalida o token atual (requer autenticação).

**Request Headers**:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response** (200 OK):
```json
{
  "message": "Logout realizado com sucesso"
}
```

### Usando Tokens

Para acessar endpoints protegidos, inclua o token no header Authorization:

```bash
curl -H "Authorization: Bearer SEU_TOKEN_AQUI" \
  http://localhost:8080/users/logout
```

### Segurança

✅ **Senhas**:
- Hash com bcrypt (cost 12)
- Nunca retornadas nas respostas
- Validação de força mínima

✅ **Tokens JWT**:
- Expiração de 24 horas
- Assinados com HS256 (HMAC-SHA256)
- Blacklist para logout efetivo
- Claims incluem: user_id, email, exp, iat, nbf

✅ **E-mails**:
- Índice único no banco
- Validação de formato
- Case-sensitive

### Receitas e Usuários

A API suporta dois tipos de receitas:

1. **Receitas Gerais**: Sem `user_id` (apenas admins podem editar/deletar)
2. **Receitas Personalizadas**: Com `user_id` (criadas por usuários autenticados)

#### Autorização de Receitas

✅ **Criar receitas**: Requer autenticação (sempre terá `user_id`)  
✅ **Listar/Visualizar**: Público (sem autenticação)  
✅ **Editar/Deletar receitas próprias**: Apenas o criador  
✅ **Editar/Deletar receitas gerais**: Apenas admins (preparado para futuro)

**Exemplos:**

```bash
# Criar receita (requer token)
curl -X POST http://localhost:8080/recipes \
  -H "Authorization: Bearer SEU_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Minha Receita","prep_time":30,"servings":4}'

# Listar receitas (público)
curl http://localhost:8080/recipes

# Editar receita própria (requer token)
curl -X PUT http://localhost:8080/recipes/1 \
  -H "Authorization: Bearer SEU_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Receita Atualizada"}'
```

**Respostas de Autorização:**

- **401 Unauthorized**: Token ausente ou inválido
- **403 Forbidden**: Tentativa de editar/deletar receita de outro usuário

## 🔌 Endpoints

### GET /health

Health check endpoint para monitoramento e plataformas cloud.

**Response**:

```json
{
  "status": "healthy",
  "timestamp": 1703433600
}
```

**Status**: 200 OK  
**Content-Type**: application/json

### GET /test

Endpoint de teste que retorna uma mensagem "hello world".

**Response**:

```json
{
  "message": "hello world"
}
```

**Status**: 200 OK  
**Content-Type**: application/json

### GET /recipes

Lista todas as receitas cadastradas (gerais e personalizadas).

**Response**:

```json
{
  "data": [
    {
      "id": 1,
      "title": "Bolo de Chocolate",
      "description": "Delicioso bolo de chocolate",
      "prep_time": 45,
      "servings": 8,
      "difficulty": "média",
      "user_id": null,
      "created_at": "2025-12-24T10:30:45Z",
      "updated_at": "2025-12-24T10:30:45Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 1,
    "total_pages": 1,
    "has_next": false,
    "has_prev": false
  }
}
```

### POST /recipes

Cria uma nova receita.

**Request Body**:

```json
{
  "title": "Bolo de Chocolate",
  "description": "Delicioso bolo de chocolate",
  "prep_time": 45,
  "servings": 8,
  "difficulty": "média"
}
```

**Response**: 201 Created

### GET /recipes/{id}

Busca uma receita específica por ID.

**Response**: 200 OK

### PUT /recipes/{id}

Atualiza uma receita existente.

**Response**: 200 OK

### DELETE /recipes/{id}

Remove uma receita (soft delete).

**Response**: 200 OK

## 🗄️ Database PostgreSQL

O projeto utiliza **PostgreSQL** com **GORM** para persistência de dados.

### Modelo de Dados

#### Receita (Recipe)

| Campo         | Tipo      | Descrição                          |
| ------------- | --------- | ---------------------------------- |
| `id`          | uint      | ID único da receita                |
| `title`       | string    | Título (max 200 caracteres)        |
| `description` | text      | Descrição detalhada                |
| `prep_time`   | int       | Tempo de preparo em minutos        |
| `servings`    | int       | Número de porções                  |
| `difficulty`  | string    | Dificuldade: fácil, média, difícil |
| `created_at`  | timestamp | Data de criação                    |
| `updated_at`  | timestamp | Data de atualização                |
| `deleted_at`  | timestamp | Data de exclusão (soft delete)     |

### Configuração Local

Para desenvolvimento local com PostgreSQL:

```bash
# 1. Instalar PostgreSQL
# macOS: brew install postgresql
# Ubuntu: sudo apt install postgresql

# 2. Criar database
createdb receitas_db

# 3. Configurar variável de ambiente
export DATABASE_URL="postgres://usuario:senha@localhost:5432/receitas_db?sslmode=disable"

# 4. Executar aplicação (migrations automáticas)
go run ./cmd/api
```

### Railway - PostgreSQL

No Railway, adicionar PostgreSQL é simples:

1. **Dashboard Railway** → **New** → **Database** → **Add PostgreSQL**
2. Railway cria automaticamente a variável `DATABASE_URL`
3. Aplicação conecta automaticamente ao database
4. Migrations executam no startup

### GORM Features

- ✅ **AutoMigrate**: Cria/atualiza tabelas automaticamente
- ✅ **Soft Delete**: Registros deletados ficam recuperáveis
- ✅ **Connection Pool**: Performance otimizada
- ✅ **Timestamps**: `created_at` e `updated_at` automáticos
- ✅ **Query Logging**: Queries logadas em desenvolvimento

### Exemplos de Uso

**Criar Receita:**

```bash
curl -X POST http://localhost:8080/recipes \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Bolo de Chocolate",
    "description": "Delicioso bolo de chocolate com cobertura",
    "prep_time": 45,
    "servings": 8,
    "difficulty": "média"
  }'
```

**Listar Receitas:**

```bash
curl http://localhost:8080/recipes
```

**Buscar Receita:**

```bash
curl http://localhost:8080/recipes/1
```

**Atualizar Receita:**

```bash
curl -X PUT http://localhost:8080/recipes/1 \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Bolo de Chocolate Especial",
    "prep_time": 50
  }'
```

**Deletar Receita:**

```bash
curl -X DELETE http://localhost:8080/recipes/1
```

## 📝 Desenvolvimento

### Adicionar nova rota

1. Criar handler em `/internal/http/handlers/{nome}.go`
2. Implementar a função do handler
3. Registrar a rota em `/internal/http/routes/routes.go`
4. Criar testes em `/test/{nome}_handler_test.go`

**Exemplo**:

```go
// internal/http/handlers/exemplo.go
package handlers

import (
    "net/http"
    "github.com/davidsonmarra/receitas-app/pkg/log"
    "github.com/davidsonmarra/receitas-app/pkg/response"
)

func ExemploHandler(w http.ResponseWriter, r *http.Request) {
    // Log com contexto (inclui request_id automaticamente)
    log.InfoCtx(r.Context(), "processing example request")

    response.JSON(w, http.StatusOK, map[string]string{
        "message": "exemplo",
    })
}
```

### Formato de código

O projeto segue as convenções padrão de Go. Para formatar o código:

```bash
go fmt ./...
```

## 🚀 Deploy em Produção

O projeto está pronto para deploy em diversas plataformas cloud.

### 🚂 Railway

1. **Conectar Repositório**

   - Acesse [Railway](https://railway.app)
   - Conecte seu repositório GitHub
   - Railway detectará automaticamente o Dockerfile

2. **Adicionar PostgreSQL**

   - No dashboard → **New** → **Database** → **Add PostgreSQL**
   - Railway cria automaticamente `DATABASE_URL`
   - Database gratuito até 500MB

3. **Configurar Variáveis de Ambiente**

   ```
   ENV=production
   LOG_LEVEL=info
   ```

   (DATABASE_URL é criado automaticamente pelo Railway)

4. **Deploy Automático**
   - Cada push para a branch main fará deploy automático
   - Railway define a variável `PORT` automaticamente
   - Migrations executam no startup
   - Health check configurado em `/health`

### 🟣 Heroku

```bash
# Login no Heroku
heroku login

# Criar aplicação
heroku create minha-api-receitas

# Configurar variáveis
heroku config:set ENV=production
heroku config:set LOG_LEVEL=info

# Deploy
git push heroku main

# Verificar logs
heroku logs --tail
```

### 🐳 Docker Local

```bash
# Build da imagem
docker build -t receitas-app .

# Executar container
docker run -p 8080:8080 \
  -e ENV=production \
  -e LOG_LEVEL=info \
  receitas-app

# Verificar saúde
curl http://localhost:8080/health
```

### ☁️ Google Cloud Run

```bash
# Fazer deploy direto do código
gcloud run deploy receitas-app \
  --source . \
  --set-env-vars ENV=production,LOG_LEVEL=info \
  --allow-unauthenticated \
  --region us-central1

# Ou usando Docker
gcloud builds submit --tag gcr.io/PROJECT_ID/receitas-app
gcloud run deploy receitas-app \
  --image gcr.io/PROJECT_ID/receitas-app \
  --set-env-vars ENV=production,LOG_LEVEL=info
```

### 📋 Variáveis de Ambiente Necessárias

| Variável       | Obrigatória | Padrão        | Descrição                                        |
| -------------- | ----------- | ------------- | ------------------------------------------------ |
| `ENV`          | Não         | `development` | Ambiente: `development`, `staging`, `production` |
| `LOG_LEVEL`    | Não         | `info`        | Nível de log: `debug`, `info`, `warn`, `error`   |
| `PORT`         | Não         | `8080`        | Porta do servidor (auto-definida em clouds)      |
| `DATABASE_URL` | Sim         | -             | PostgreSQL connection string (auto no Railway)   |
| `JWT_SECRET`   | Sim         | -             | Secret para assinar tokens JWT (min 32 chars)    |

### ✅ Checklist Pré-Deploy

- [ ] Testes passando: `go test ./...`
- [ ] Build funcional: `go build ./cmd/api`
- [ ] Docker build: `docker build -t receitas-app .`
- [ ] Health check: `curl http://localhost:8080/health`
- [ ] Variáveis de ambiente configuradas
- [ ] Logs estruturados testados

### 🔍 Monitoramento Pós-Deploy

**Health Check Endpoint:**

```bash
curl https://sua-app.railway.app/health
```

**Resposta esperada:**

```json
{
  "status": "healthy",
  "timestamp": 1703433600
}
```

**Logs em Produção:**

```bash
# Railway
railway logs

# Heroku
heroku logs --tail

# Google Cloud Run
gcloud run services logs read receitas-app --limit=50
```

## 🛡️ Rate Limiting

A API implementa **rate limiting** para proteger contra abuso e garantir qualidade de serviço. O sistema limita o número de requisições por IP em janelas de tempo de 1 minuto.

### Estratégia de Limites

A API utiliza **dois níveis de rate limiting**:

1. **Global**: Limite máximo para qualquer endpoint
2. **Por Endpoint**: Limites específicos baseados no tipo de operação

| Endpoint | Método | Limite | Tipo |
|----------|--------|--------|------|
| `/health` | GET | 100/min | Global |
| `/test` | GET | 100/min | Global |
| `/recipes` | GET | 60/min | Leitura |
| `/recipes` | POST | 20/min | Escrita |
| `/recipes/{id}` | GET | 60/min | Leitura |
| `/recipes/{id}` | PUT | 20/min | Escrita |
| `/recipes/{id}` | DELETE | 20/min | Escrita |

### Configuração

Configure os limites através de variáveis de ambiente:

```bash
# Habilitar/desabilitar rate limiting (padrão: true)
RATE_LIMIT_ENABLED=true

# Limite global para todos os endpoints (padrão: 100 req/min)
RATE_LIMIT_GLOBAL=100

# Limite para endpoints de leitura (padrão: 60 req/min)
RATE_LIMIT_READ=60

# Limite para endpoints de escrita (padrão: 20 req/min)
RATE_LIMIT_WRITE=20
```

### Resposta 429 (Too Many Requests)

Quando o limite é excedido, a API retorna:

**Status**: `429 Too Many Requests`

**Headers**:
```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1735215720
Retry-After: 42
Content-Type: application/json
```

**Body**:
```json
{
  "error": {
    "title": "Ops, muitas requisições!",
    "message": "Você excedeu o limite de requisições. Tente novamente em alguns segundos."
  }
}
```

### Identificação do Cliente

O rate limiting identifica clientes pelo **endereço IP**, considerando proxies e load balancers:

1. **X-Forwarded-For**: Primeiro IP da lista (cliente original)
2. **X-Real-IP**: IP real do cliente (nginx, etc)
3. **RemoteAddr**: Fallback para IP direto

Isso garante que o rate limiting funcione corretamente em ambientes de produção com proxies reversos (Railway, Heroku, etc).

### Desabilitar em Desenvolvimento

Para desabilitar o rate limiting durante o desenvolvimento:

```bash
export RATE_LIMIT_ENABLED=false
go run ./cmd/api
```

### Testar Rate Limiting

#### Teste Manual com curl

```bash
# Fazer múltiplas requisições rapidamente
for i in {1..65}; do
  curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/recipes
done

# Primeiras 60 devem retornar 200
# Demais devem retornar 429
```

#### Verificar Headers

```bash
curl -I http://localhost:8080/recipes

# Headers de rate limit:
# X-RateLimit-Limit: 60
# X-RateLimit-Remaining: 59
# X-RateLimit-Reset: 1735215720
```

### Escalabilidade

**Implementação Atual**: Memória local (in-memory)
- ✅ Simples e performático
- ✅ Sem dependências externas
- ✅ Ideal para instância única (padrão Railway)
- ⚠️ Não compartilha estado entre múltiplas instâncias

**Migração Futura para Redis** (se necessário):

Se você escalar para múltiplas instâncias no Railway, a arquitetura está preparada para trocar o storage de memória local por Redis, permitindo rate limiting compartilhado entre todas as instâncias.

### Vantagens

✅ **Proteção contra abuso**: Previne ataques de força bruta e DDoS  
✅ **Qualidade de serviço**: Garante recursos para todos os usuários  
✅ **Flexível**: Limites diferentes por tipo de operação  
✅ **Configurável**: Ajuste via variáveis de ambiente  
✅ **Informativo**: Headers seguem padrões RFC 6585  
✅ **Transparente**: Logs de rate limit com IP do cliente

## 🔒 Security Headers

A API implementa headers de segurança seguindo as recomendações da OWASP para proteger contra vulnerabilidades comuns.

### Headers Implementados

| Header | Valor | Proteção |
|--------|-------|----------|
| `X-Frame-Options` | DENY | Previne clickjacking |
| `X-Content-Type-Options` | nosniff | Previne MIME type sniffing |
| `X-XSS-Protection` | 1; mode=block | Proteção XSS (browsers antigos) |
| `Strict-Transport-Security` | max-age=31536000 | Force HTTPS por 1 ano |
| `Content-Security-Policy` | default-src 'none' | Previne XSS e injection |
| `Referrer-Policy` | strict-origin-when-cross-origin | Controla referrer |
| `Permissions-Policy` | Desabilita APIs desnecessárias | Limita acesso a features |

### Verificar Headers

```bash
# Verificar headers em produção
curl -I https://sua-api.railway.app/health

# Ou com ferramenta de análise
curl -I https://sua-api.railway.app/health | grep -E "(X-|Strict-|Content-Security|Referrer|Permissions)"
```

### Score de Segurança

Teste sua API em:
- [SecurityHeaders.com](https://securityheaders.com)
- [Mozilla Observatory](https://observatory.mozilla.org)

Resultado esperado: **Nota A** ✅

### Detalhes dos Headers

#### X-Frame-Options: DENY
Previne que a página seja carregada em um iframe, protegendo contra ataques de clickjacking.

#### X-Content-Type-Options: nosniff
Impede que o browser tente "adivinhar" o tipo MIME de arquivos, prevenindo ataques baseados em MIME confusion.

#### X-XSS-Protection: 1; mode=block
Ativa a proteção XSS em browsers antigos que ainda suportam este header (browsers modernos usam CSP).

#### Strict-Transport-Security (HSTS)
Force o uso de HTTPS por 1 ano, incluindo subdomínios. Só é enviado em conexões HTTPS.

#### Content-Security-Policy
Define que nenhum recurso externo pode ser carregado, protegendo contra XSS e injection attacks.

#### Referrer-Policy
Controla quais informações de referrer são enviadas, protegendo privacidade dos usuários.

#### Permissions-Policy
Desabilita APIs do browser que não são necessárias para uma API REST (geolocation, camera, microphone, etc).

### Compliance

✅ **OWASP Top 10** - Proteção contra vulnerabilidades mais comuns  
✅ **PCI DSS** - Requisitos de segurança para dados de cartão  
✅ **GDPR** - Proteção de dados e privacidade  
✅ **LGPD** - Lei Geral de Proteção de Dados (Brasil)

## 🎯 Roadmap

- [x] Logs estruturados com zap
- [x] Request ID tracking
- [x] Graceful shutdown
- [x] Health check endpoint
- [x] Docker & Dockerfile multi-stage
- [x] Production-ready (Railway, Heroku, Cloud Run)
- [x] PostgreSQL + GORM
- [x] CRUD completo de Receitas
- [x] Migrations automáticas (GORM AutoMigrate)
- [x] Soft Delete
- [x] Validação de dados (go-playground/validator)
- [x] Paginação e filtros
- [x] Rate Limiting (proteção contra abuso)
- [x] Security Headers (OWASP compliance)
- [x] Autenticação JWT (login, logout, registro)
- [x] Sistema de usuários
- [x] Receitas gerais e personalizadas
- [ ] Relacionamentos (Ingredientes, Categorias)
- [ ] Busca full-text
- [ ] Autorização de receitas por usuário
- [ ] Upload de imagens
- [ ] Observabilidade (métricas, tracing)
- [ ] CI/CD
- [ ] Documentação da API (Swagger/OpenAPI)

## 📄 Licença

Este projeto está em desenvolvimento.

---

**Desenvolvido com Go** 🐹
