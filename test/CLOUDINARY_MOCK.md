# Mock do Cloudinary para Testes

Este documento explica como os testes do Cloudinary foram mockados para permitir execução sem dependências externas.

## Problema Original

Antes da implementação do mock, os testes relacionados ao Cloudinary tinham os seguintes problemas:

- ❌ Requeriam `CLOUDINARY_URL` configurada
- ❌ Faziam chamadas HTTP reais para a API do Cloudinary
- ❌ Eram lentos devido às chamadas de rede
- ❌ Eram pulados (SKIP) quando executados localmente sem credenciais
- ❌ Não eram determinísticos (dependiam de resposta externa)

## Solução: Mock Service

Foi criado um `MockCloudinaryService` que implementa a mesma interface do serviço real, mas sem fazer chamadas externas.

### Localização

```
test/testdb/cloudinary_mock.go
```

### Funcionalidades Mockadas

#### 1. Upload de Imagem

```go
service := testdb.NewMockCloudinaryService()
result, err := service.UploadImage(ctx, params)
```

**Validações implementadas:**
- ✅ Arquivo não pode ser nulo
- ✅ Arquivo não pode estar vazio
- ✅ Nome do arquivo é obrigatório
- ✅ Extensão do arquivo deve ser válida (.jpg, .jpeg, .png, .gif, .webp)

**Retorna:**
- PublicID simulado: `mock/test-image-123`
- URLs simuladas do Cloudinary
- Dimensões e formato da imagem
- Tamanho em bytes

#### 2. Exclusão de Imagem

```go
err := service.DeleteImage(ctx, publicID)
```

**Validações implementadas:**
- ✅ PublicID não pode ser vazio

#### 3. URL Otimizada

```go
url, err := service.GetOptimizedURL(publicID, width, height, quality)
```

**Validações implementadas:**
- ✅ PublicID não pode ser vazio

**Retorna:**
- URL otimizada com parâmetros de transformação
- Formato: `https://res.cloudinary.com/mock/image/upload/w_600,h_400,q_auto/publicID`

#### 4. Variantes de Imagem

```go
variants := service.GetImageVariants(publicID)
```

**Retorna:**
- `thumbnail`: 150x150
- `small`: 400x400
- `medium`: 800x800
- `large`: 1200x1200
- `original`: sem transformações

## Testes Atualizados

Os seguintes testes agora usam o mock e **não são mais pulados**:

### ✅ `cloudinary_test.go`

- `TestUploadImage_EmptyFile` - Valida erro ao fazer upload de arquivo vazio
- `TestDeleteImage_EmptyPublicID` - Valida erro ao deletar sem publicID
- `TestGetOptimizedURL_EmptyPublicID` - Valida erro ao gerar URL sem publicID
- `TestGetOptimizedURL_ValidParams` - Valida geração de URL com parâmetros

### ✅ `recipe_image_test.go`

- `TestDeleteRecipeImage_RecipeNotFound` - Testa exclusão de imagem de receita inexistente
- `TestGetRecipeImageVariants_RecipeNotFound` - Testa obtenção de variantes de receita inexistente
- `TestGetRecipeImageVariants_NoImage` - Testa obtenção de variantes quando receita não tem imagem

### ✅ Todos os Testes Executando

Todos os testes relacionados ao Cloudinary agora executam com sucesso usando o mock, incluindo:

- `TestGetOptimizedRecipeImage_WithQueryParams` - Agora usa injeção de dependência via `storage.ServiceFactory`

## Como Usar o Mock

### Método 1: Uso Direto (para testes unitários)

```go
func TestMyFeature(t *testing.T) {
    // Criar mock
    service := testdb.NewMockCloudinaryService()
    
    // Usar como serviço normal
    result, err := service.UploadImage(ctx, storage.UploadImageParams{
        File:     mockFile,
        FileName: "test.jpg",
        Folder:   "recipes",
    })
    
    if err != nil {
        t.Fatalf("erro inesperado: %v", err)
    }
    
    // Validar resultado
    if result.PublicID == "" {
        t.Error("publicID não deve ser vazio")
    }
}
```

### Método 2: Injeção via ServiceFactory (para testes de handlers)

```go
func TestMyHandler(t *testing.T) {
    testdb.SetupWithCleanup(t)
    
    // Substituir o ServiceFactory por um mock
    originalFactory := storage.ServiceFactory
    defer func() { storage.ServiceFactory = originalFactory }()
    
    mockService := testdb.NewMockCloudinaryService()
    storage.ServiceFactory = func() (storage.ImageService, error) {
        return mockService, nil
    }
    
    // Agora os handlers usarão o mock automaticamente
    // ...
}
```

### Simular Erros

```go
// Simular falha no upload
service := testdb.NewMockCloudinaryService()
service.ShouldFailUpload = true

_, err := service.UploadImage(ctx, params)
// err será "erro simulado de upload"

// Simular falha na exclusão
service.ShouldFailDelete = true
err = service.DeleteImage(ctx, "some-id")
// err será "erro simulado ao deletar imagem"
```

## Benefícios

### ✅ Velocidade
Testes rodam instantaneamente sem chamadas HTTP

### ✅ Confiabilidade
Não dependem de conectividade ou disponibilidade da API externa

### ✅ Determinismo
Sempre retornam os mesmos resultados para as mesmas entradas

### ✅ Cobertura
Permitem testar cenários de erro que seriam difíceis de reproduzir com a API real

### ✅ Desenvolvimento Offline
Desenvolvedores podem rodar testes sem credenciais do Cloudinary

## Estatísticas

### Antes do Mock
- **Testes pulados**: 8
- **Testes executáveis**: 87
- **Taxa de execução**: 91.6%

### Depois do Mock (com ServiceFactory)
- **Testes pulados**: 0 ✅
- **Testes executáveis**: 95
- **Taxa de execução**: 100% 🎉

## Solução: ServiceFactory Pattern

Para permitir que handlers usem o mock sem refatoração massiva, foi implementado o padrão **ServiceFactory**:

### Implementação

```go
// pkg/storage/cloudinary.go

// Interface que define os métodos do serviço de imagens
type ImageService interface {
    UploadImage(ctx context.Context, params UploadImageParams) (*UploadResult, error)
    DeleteImage(ctx context.Context, publicID string) error
    GetOptimizedURL(publicID string, width, height int, quality string) (string, error)
    GetImageVariants(publicID string) map[string]string
}

// Factory que pode ser substituída nos testes
var ServiceFactory func() (ImageService, error) = func() (ImageService, error) {
    return NewCloudinaryService()
}
```

### Uso nos Handlers

```go
func GetOptimizedRecipeImage(w http.ResponseWriter, r *http.Request) {
    // Usa o ServiceFactory ao invés de instanciar diretamente
    imageService, err := storage.ServiceFactory()
    if err != nil {
        // ...
    }
    
    // Usar o serviço normalmente
    url, err := imageService.GetOptimizedURL(publicID, width, height, quality)
    // ...
}
```

### Benefícios

- ✅ **Zero impacto no código de produção**: Handlers continuam funcionando normalmente
- ✅ **Fácil de testar**: Basta substituir o `ServiceFactory` nos testes
- ✅ **Type-safe**: Interface garante compatibilidade em tempo de compilação
- ✅ **Flexível**: Permite trocar implementações facilmente

## Arquitetura da Solução

```
┌─────────────────────────────────────────────────────────────┐
│                    Código de Produção                        │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  Handler                    ServiceFactory                   │
│    │                             │                           │
│    └──> storage.ServiceFactory() ├──> NewCloudinaryService() │
│                                   │    (Cloudinary real)     │
│                                                               │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                       Em Testes                              │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  Handler                    ServiceFactory                   │
│    │                             │                           │
│    └──> storage.ServiceFactory() ├──> MockCloudinaryService  │
│         (substituído no teste)   │    (Mock em memória)     │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

## Possíveis Melhorias Futuras

1. **Testes de Integração E2E**: Adicionar testes que rodam apenas em CI com Cloudinary real
2. **Mock Mais Realista**: Adicionar validações de tamanho de arquivo, tipos MIME, etc.
3. **Métricas de Upload**: Simular tempos de upload e taxas de erro
4. **Cache de Imagens**: Adicionar simulação de cache CDN

## Conclusão

Com a implementação do **ServiceFactory Pattern** e do **MockCloudinaryService**, agora **100% dos testes** são executados localmente sem dependências externas! 🎉

### Resultados Finais

- ✅ **95 testes passando**
- ✅ **0 testes pulados**
- ✅ **0 testes falhando**
- ✅ **100% de taxa de execução**
- ⚡ **Testes instantâneos** (sem chamadas HTTP)
- 🎯 **Testes determinísticos e confiáveis**

A solução mantém a qualidade e cobertura dos testes enquanto melhora significativamente a velocidade, confiabilidade e experiência de desenvolvimento da suite de testes.

