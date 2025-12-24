# Backend Go - API Receitas

Backend em Go desenvolvido com arquitetura limpa e escalável.

## 📋 Descrição

Este projeto estabelece a fundação para um serviço backend escrito em Go. A Fase 1 implementa a infraestrutura core com um servidor HTTP mínimo, endpoints básicos, testes unitários e comandos Cursor para automação de desenvolvimento.

## 🔧 Tecnologias

- **Go**: ≥ 1.22
- **Router**: [go-chi/chi](https://github.com/go-chi/chi) v5
- **Testes**: testing + httptest
- **Logging**: biblioteca padrão Go

## 📁 Estrutura do Projeto

```
receitas-app/
├── cmd/api/                    # Executáveis
│   └── main.go                 # Entrypoint da aplicação
├── internal/                   # Código interno da aplicação
│   ├── server/                 # Configuração do servidor
│   │   └── server.go
│   └── http/
│       ├── routes/             # Registro de rotas
│       │   └── routes.go
│       └── handlers/           # Handlers HTTP
│           └── test.go
├── pkg/                        # Utilitários reutilizáveis
│   └── response/
│       └── json.go             # Helpers para respostas JSON
├── test/                       # Testes unitários
│   └── test_handler_test.go
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
    "github.com/davidsonmarra/receitas-app/pkg/response"
)

func ExemploHandler(w http.ResponseWriter, r *http.Request) {
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

## 🎯 Roadmap

- [ ] Endpoints RESTful completos
- [ ] Camada de banco de dados
- [ ] Autenticação e autorização
- [ ] Migrations
- [ ] Observabilidade (logs estruturados, métricas, tracing)
- [ ] CI/CD
- [ ] Docker & Docker Compose
- [ ] Documentação da API (Swagger/OpenAPI)

## 📄 Licença

Este projeto está em desenvolvimento.

---

**Desenvolvido com Go** 🐹
