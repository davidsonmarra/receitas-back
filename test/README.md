# Testes - Receitas App

Este diretório contém todos os testes da aplicação, incluindo testes unitários e de integração.

## Estrutura

```
test/
├── testdb/
│   └── setup.go              # Infraestrutura de banco de dados para testes
├── *_test.go                 # Arquivos de teste
└── README.md                 # Este arquivo
```

## Tipos de Testes

### Testes Unitários

Testes que não dependem de banco de dados ou serviços externos:

- `auth_middleware_test.go` - Middleware de autenticação
- `cors_test.go` - Configuração CORS
- `jwt_test.go` - Geração e validação de tokens JWT
- `logger_test.go` - Sistema de logging
- `pagination_test.go` - Paginação de resultados
- `password_test.go` - Hash e verificação de senhas
- `ratelimit_test.go` - Rate limiting
- `security_test.go` - Headers de segurança
- `validation_test.go` - Validação de dados

### Testes de Integração

Testes que utilizam banco de dados SQLite in-memory:

- `admin_test.go` - Funcionalidades de administrador
- `ingredient_test.go` - CRUD de ingredientes
- `recipe_handler_test.go` - CRUD de receitas
- `recipe_authorization_test.go` - Autorização de receitas
- `recipe_image_test.go` - Upload e gerenciamento de imagens
- `user_handler_test.go` - Registro, login e logout de usuários
- `health_handler_test.go` - Endpoint de health check

### Testes que Requerem Serviços Externos

- `cloudinary_test.go` - Integração com Cloudinary (requer `CLOUDINARY_URL`)

## Como Rodar os Testes

### Todos os Testes

```bash
go test ./test/...
```

### Testes com Saída Detalhada

```bash
go test ./test/... -v
```

### Testes Específicos

```bash
# Rodar apenas testes de um arquivo
go test ./test/... -run TestListRecipes

# Rodar testes que correspondem a um padrão
go test ./test/... -run "Test.*Recipe"
```

### Com Cobertura

```bash
go test ./test/... -cover
go test ./test/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Infraestrutura de Testes

### SQLite In-Memory

Os testes de integração utilizam SQLite in-memory através do pacote `testdb`:

```go
func TestExample(t *testing.T) {
    testdb.SetupWithCleanup(t)
    
    // Seu teste aqui
    // O banco será automaticamente limpo após o teste
}
```

**Vantagens:**
- ✅ Não requer PostgreSQL instalado
- ✅ Muito rápido (tudo em memória)
- ✅ Isolamento perfeito entre testes
- ✅ Zero configuração

### Helpers Disponíveis

O pacote `testdb` fornece helpers úteis:

```go
// Criar usuário de teste
user := testdb.SeedUser(t, "Nome", "email@test.com", "hashedPassword", "user")

// Criar receita de teste
recipe := testdb.SeedRecipe(t, "Título", "Descrição", userID, false)

// Criar ingrediente de teste
ingredient := testdb.SeedIngredient(t, "Tomate", "vegetais", 15.0)

// Limpar tabela específica
testdb.CleanTable(t, "recipes")
```

## Variáveis de Ambiente para Testes

### Opcionais

- `CLOUDINARY_URL` - Para testes de integração com Cloudinary
  - Se não configurada, os testes relacionados serão pulados (SKIP)

### Não Necessárias

- `DATABASE_URL` - **Não é necessária!** Os testes usam SQLite in-memory
- `JWT_SECRET` - Usa valor padrão para testes
- `PORT` - Não usado em testes

## Boas Práticas

### 1. Isolamento de Testes

Cada teste deve ser independente e não afetar outros:

```go
func TestExample(t *testing.T) {
    testdb.SetupWithCleanup(t) // Limpa automaticamente
    
    // Criar dados necessários
    user := testdb.SeedUser(t, ...)
    
    // Executar teste
    // ...
}
```

### 2. Nomenclatura

- Use nomes descritivos: `TestCreateRecipe_WithValidData`
- Padrão: `Test<Função>_<Cenário>`

### 3. Asserções Claras

```go
if got != want {
    t.Errorf("esperado %v, obteve %v", want, got)
}
```

### 4. Subtestes

Para múltiplos cenários:

```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name string
        input string
        want bool
    }{
        {"valid email", "test@example.com", true},
        {"invalid email", "invalid", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := validate(tt.input)
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Debugging

### Ver Logs Durante Testes

```bash
go test ./test/... -v 2>&1 | grep -A 5 "FAIL"
```

### Rodar Teste Específico com Detalhes

```bash
go test ./test/... -run TestListRecipes -v
```

### Ver Apenas Resumo

```bash
go test ./test/... | grep -E "(PASS|FAIL|SKIP)"
```

## Estatísticas Atuais

Execute para ver estatísticas:

```bash
go test ./test/... -cover
```

## Troubleshooting

### Teste Falhando com "nil pointer"

Verifique se você está usando `testdb.SetupWithCleanup(t)` no início do teste.

### Teste Sendo Pulado (SKIP)

Alguns testes requerem variáveis de ambiente específicas (como `CLOUDINARY_URL`). Isso é esperado.

### Conflitos de Dados

Se testes estão interferindo uns com os outros, verifique se cada teste está usando `testdb.SetupWithCleanup(t)`.

## Contribuindo

Ao adicionar novos testes:

1. Use `testdb.SetupWithCleanup(t)` para testes que precisam de DB
2. Nomeie testes claramente
3. Documente cenários complexos
4. Mantenha testes rápidos e isolados
5. Adicione testes para bugs corrigidos

## Recursos Adicionais

- [Testing em Go](https://golang.org/pkg/testing/)
- [Table-Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [GORM Testing](https://gorm.io/docs/testing.html)

