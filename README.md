# Backend Go - API Receitas

Backend em Go desenvolvido com arquitetura limpa e escalável.

## 📋 Descrição

Este projeto estabelece a fundação para um serviço backend escrito em Go. A Fase 1 implementa a infraestrutura core com um servidor HTTP mínimo, endpoints básicos, testes unitários e comandos Cursor para automação de desenvolvimento.

## 🔧 Tecnologias

- **Go**: ≥ 1.23
- **Database**: [PostgreSQL](https://www.postgresql.org/) - Database relacional
- **ORM**: [GORM](https://gorm.io/) v1.31+ - ORM completo para Go
- **Router**: [go-chi/chi](https://github.com/go-chi/chi) v5
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
│       │   └── requestid.go    # Middleware de Request ID
│       ├── routes/             # Registro de rotas
│       │   └── routes.go
│       └── handlers/           # Handlers HTTP
│           ├── health.go       # Health check
│           ├── test.go         # Handler de teste
│           └── recipe.go       # CRUD de receitas
├── pkg/                        # Utilitários reutilizáveis
│   ├── database/               # Conexão com database
│   │   └── connection.go       # PostgreSQL + GORM
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

Lista todas as receitas cadastradas.

**Response**:

```json
[
  {
    "id": 1,
    "title": "Bolo de Chocolate",
    "description": "Delicioso bolo de chocolate",
    "prep_time": 45,
    "servings": 8,
    "difficulty": "média",
    "created_at": "2025-12-24T10:30:45Z",
    "updated_at": "2025-12-24T10:30:45Z"
  }
]
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
- [ ] Relacionamentos (Ingredientes, Categorias, Usuários)
- [ ] Validação de dados (go-playground/validator)
- [ ] Paginação e filtros
- [ ] Busca full-text
- [ ] Autenticação e autorização (JWT)
- [ ] Upload de imagens
- [ ] Observabilidade (métricas, tracing)
- [ ] CI/CD
- [ ] Documentação da API (Swagger/OpenAPI)

## 📄 Licença

Este projeto está em desenvolvimento.

---

**Desenvolvido com Go** 🐹
