# Backend Go - API Receitas

Backend em Go desenvolvido com arquitetura limpa e escalável.

## 📋 Descrição

Este projeto estabelece a fundação para um serviço backend escrito em Go. A Fase 1 implementa a infraestrutura core com um servidor HTTP mínimo, endpoints básicos, testes unitários e comandos Cursor para automação de desenvolvimento.

## 🔧 Tecnologias

- **Go**: ≥ 1.22
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
│   ├── server/                 # Configuração do servidor
│   │   └── server.go
│   └── http/
│       ├── middleware/         # Middlewares HTTP
│       │   └── requestid.go    # Middleware de Request ID
│       ├── routes/             # Registro de rotas
│       │   └── routes.go
│       └── handlers/           # Handlers HTTP
│           └── test.go
├── pkg/                        # Utilitários reutilizáveis
│   ├── log/                    # Sistema de logging
│   │   ├── logger.go           # API de logging (estilo Android)
│   │   └── config.go           # Configuração do logger
│   └── response/
│       └── json.go             # Helpers para respostas JSON
├── test/                       # Testes unitários
│   ├── test_handler_test.go
│   └── logger_test.go
├── .cursor/commands/           # Comandos Cursor
│   ├── create-route.md
│   └── create-test.md
├── go.mod                      # Dependências
└── README.md
```

## 🚀 Como Executar

### Pré-requisitos

- Go 1.22 ou superior instalado

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

| Nível | Variável | O que mostra |
|-------|----------|--------------|
| **debug** | `LOG_LEVEL=debug` | Tudo (debug, info, warn, error) |
| **info** | `LOG_LEVEL=info` | info, warn, error (padrão produção) |
| **warn** | `LOG_LEVEL=warn` | warn, error |
| **error** | `LOG_LEVEL=error` | Somente erros |

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

2. **Configurar Variáveis de Ambiente**
   ```
   ENV=production
   LOG_LEVEL=info
   ```

3. **Deploy Automático**
   - Cada push para a branch main fará deploy automático
   - Railway define a variável `PORT` automaticamente
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

| Variável | Obrigatória | Padrão | Descrição |
|----------|-------------|--------|-----------|
| `ENV` | Não | `development` | Ambiente: `development`, `staging`, `production` |
| `LOG_LEVEL` | Não | `info` | Nível de log: `debug`, `info`, `warn`, `error` |
| `PORT` | Não | `8080` | Porta do servidor (auto-definida em clouds) |

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
- [ ] Endpoints RESTful completos
- [ ] Camada de banco de dados
- [ ] Autenticação e autorização
- [ ] Migrations
- [ ] Observabilidade (métricas, tracing)
- [ ] CI/CD
- [ ] Documentação da API (Swagger/OpenAPI)

## 📄 Licença

Este projeto está em desenvolvimento.

---

**Desenvolvido com Go** 🐹
