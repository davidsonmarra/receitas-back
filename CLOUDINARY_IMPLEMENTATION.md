# Implementação do Cloudinary - Sistema de Imagens

## 📋 Visão Geral

Este documento descreve a implementação completa do sistema de upload e gerenciamento de imagens usando Cloudinary no projeto Receitas App.

## 🏗️ Arquitetura

### Componentes Principais

1. **Storage Service** (`pkg/storage/cloudinary.go`)
   - Gerencia toda a comunicação com a API do Cloudinary
   - Implementa upload, deleção e geração de URLs otimizadas
   - Validação de arquivos e tratamento de erros

2. **HTTP Handlers** (`internal/http/handlers/recipe_image.go`)
   - Endpoints REST para operações com imagens
   - Validação de autenticação e autorização
   - Integração com o banco de dados

3. **Database Models** (`internal/models/recipe.go`)
   - Campos `ImageURL` e `ImagePublicID` no modelo Recipe
   - Armazenamento de referências às imagens

## 🔧 Configuração

### Variáveis de Ambiente

```bash
CLOUDINARY_URL=cloudinary://API_KEY:API_SECRET@CLOUD_NAME
```

**Formato obrigatório:**
- Prefixo: `cloudinary://`
- API Key e API Secret separados por `:`
- Cloud Name após `@`

### Limites e Constantes

```go
const (
    maxImageSizeMB = 5         // 5MB máximo por imagem
    maxImageWidth  = 2000      // pixels
    maxImageHeight = 2000      // pixels
    imageFolder    = "recipes" // pasta no Cloudinary
)
```

## 📡 Endpoints da API

### 1. Upload de Imagem
```
POST /api/v1/recipes/{id}/image
Content-Type: multipart/form-data
Authorization: Bearer {token}
```

**Body:**
- `image`: arquivo de imagem (jpg, jpeg, png, gif, webp, bmp)

**Resposta (200 OK):**
```json
{
  "message": "Imagem enviada com sucesso",
  "image_url": "https://res.cloudinary.com/...",
  "image_public_id": "recipes/recipe_123",
  "width": 1920,
  "height": 1080,
  "format": "jpg",
  "size_bytes": 245678
}
```

### 2. Deletar Imagem
```
DELETE /api/v1/recipes/{id}/image
Authorization: Bearer {token}
```

**Resposta (200 OK):**
```json
{
  "message": "Imagem removida com sucesso"
}
```

### 3. Obter Variantes da Imagem
```
GET /api/v1/recipes/{id}/image/variants
```

**Resposta (200 OK):**
```json
{
  "thumbnail": {
    "url": "https://res.cloudinary.com/.../w_300,h_300,c_fill,q_auto,f_auto/...",
    "width": 300,
    "height": 300
  },
  "medium": {
    "url": "https://res.cloudinary.com/.../w_600,h_600,c_fill,q_auto,f_auto/...",
    "width": 600,
    "height": 600
  },
  "large": {
    "url": "https://res.cloudinary.com/.../w_1200,h_1200,c_fill,q_auto,f_auto/...",
    "width": 1200,
    "height": 1200
  },
  "original": {
    "url": "https://res.cloudinary.com/..."
  }
}
```

### 4. Obter URL Otimizada Customizada
```
GET /api/v1/recipes/{id}/image/optimized?width=500&height=500&quality=80
```

**Query Parameters:**
- `width`: largura desejada (1-2000 pixels, padrão: 800)
- `height`: altura desejada (1-2000 pixels, padrão: 800)
- `quality`: qualidade da imagem (padrão: "auto")

**Resposta (200 OK):**
```json
{
  "url": "https://res.cloudinary.com/.../w_500,h_500,c_fill,q_80,f_auto/...",
  "width": 500,
  "height": 500,
  "quality": "80"
}
```

## 🔒 Segurança e Autorização

### Regras de Acesso

1. **Upload e Deleção:**
   - Requer autenticação (JWT token)
   - Apenas o dono da receita ou admin pode modificar
   - Validação via middleware `RequireAuth`

2. **Visualização:**
   - Endpoints de leitura são públicos
   - URLs do Cloudinary são públicas mas ofuscadas

### Validações

- **Tipo de arquivo:** apenas imagens (jpg, jpeg, png, gif, webp, bmp)
- **Tamanho:** máximo 5MB
- **Dimensões:** redimensionamento automático para max 2000x2000
- **Formato:** conversão automática para formato otimizado

## 🧪 Testes

### Testes Unitários (`test/cloudinary_test.go`)

```bash
go test ./test/cloudinary_test.go -v
```

**Cobertura:**
- ✅ Validação de CLOUDINARY_URL ausente
- ✅ Validação de URL inválida
- ✅ Validação de extensões de arquivo válidas
- ✅ Validação de extensões de arquivo inválidas
- ✅ Upload de arquivo vazio
- ✅ Deleção com publicID vazio
- ✅ Geração de URL otimizada com publicID vazio
- ✅ Geração de URL otimizada com parâmetros válidos

### Testes de Integração (`test/recipe_image_test.go`)

```bash
go test ./test/recipe_image_test.go -v
```

**Cobertura:**
- ✅ Upload sem autenticação
- ✅ Upload sem arquivo
- ✅ Upload com receita inexistente
- ✅ Deleção sem autenticação
- ✅ Deleção com receita inexistente
- ✅ Obter variantes de receita inexistente
- ✅ Obter variantes de receita sem imagem
- ✅ Obter URL otimizada com query params

## 🎨 Transformações Cloudinary

### Transformações Automáticas

O serviço aplica automaticamente:
- **q_auto**: qualidade otimizada automaticamente
- **f_auto**: formato otimizado (WebP para navegadores compatíveis)
- **c_limit**: redimensionar mantendo proporção, sem ultrapassar limites

### Exemplo de URL Transformada

```
https://res.cloudinary.com/{cloud_name}/image/upload/w_600,h_600,c_fill,q_auto,f_auto/recipes/recipe_123
```

**Parâmetros:**
- `w_600`: largura 600px
- `h_600`: altura 600px
- `c_fill`: preencher dimensões (crop inteligente)
- `q_auto`: qualidade automática
- `f_auto`: formato automático

## 🔄 Fluxo de Upload (Frontend)

### 1. Preparar FormData

```javascript
const formData = new FormData();
formData.append('image', fileInput.files[0]);
```

### 2. Fazer Request

```javascript
const response = await fetch(`/api/v1/recipes/${recipeId}/image`, {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`
  },
  body: formData
});
```

### 3. Processar Resposta

```javascript
const result = await response.json();
console.log('Image URL:', result.image_url);
console.log('Dimensions:', result.width, 'x', result.height);
```

## 📊 Estrutura do Banco de Dados

### Modelo Recipe (atualizado)

```go
type Recipe struct {
    ID            uint      `gorm:"primarykey" json:"id"`
    Title         string    `gorm:"not null;size:200" json:"title"`
    Description   string    `gorm:"type:text" json:"description"`
    // ... outros campos ...
    ImageURL      string    `gorm:"size:500" json:"image_url,omitempty"`
    ImagePublicID string    `gorm:"size:200" json:"image_public_id,omitempty"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}
```

## 🚀 Deploy

### Checklist de Deploy

- [ ] Configurar `CLOUDINARY_URL` no Railway
- [ ] Verificar limites do plano Cloudinary
- [ ] Testar upload em produção
- [ ] Monitorar uso de banda e armazenamento
- [ ] Configurar backup (opcional)

### Monitoramento

**Cloudinary Dashboard:**
- Uso de armazenamento
- Bandwidth consumido
- Transformações realizadas
- Créditos restantes

**Logs da Aplicação:**
```bash
# Filtrar logs de upload
railway logs | grep "cloudinary upload"

# Filtrar erros
railway logs | grep "failed to upload"
```

## 💰 Custos e Limites

### Plano Free do Cloudinary

- ✅ 25 GB de armazenamento
- ✅ 25 GB de bandwidth/mês
- ✅ 25.000 transformações/mês
- ✅ Imagens ilimitadas

### Comparação com Railway Buckets

| Recurso | Cloudinary Free | Railway Buckets |
|---------|----------------|-----------------|
| Armazenamento | 25 GB grátis | $0.10/GB/mês |
| Bandwidth | 25 GB/mês grátis | $0.10/GB |
| Transformações | 25k/mês grátis | Não disponível |
| CDN Global | ✅ Incluído | ❌ Não |
| Otimização automática | ✅ Sim | ❌ Não |

**Recomendação:** Cloudinary é mais econômico até ~250 GB de uso mensal.

## 🐛 Troubleshooting

### Erro: "CLOUDINARY_URL não configurado"

**Solução:**
```bash
# Verificar variável no Railway
railway variables

# Adicionar se não existir
railway variables set CLOUDINARY_URL=cloudinary://...
```

### Erro: "invalid file parameter of unsupported type"

**Causa:** Arquivo não está sendo lido corretamente.

**Solução:** O código já implementa:
1. Reset do cursor do arquivo (`Seek(0, 0)`)
2. Leitura completa para bytes
3. Conversão para `io.Reader`

### Erro: "cloudinary retornou dados vazios"

**Causas possíveis:**
1. CLOUDINARY_URL inválida
2. Credenciais incorretas
3. Cloud Name errado

**Solução:**
```bash
# Verificar formato
echo $CLOUDINARY_URL
# Deve ser: cloudinary://KEY:SECRET@CLOUDNAME

# Testar no Cloudinary Dashboard
# https://cloudinary.com/console
```

### Imagem não aparece no frontend

**Checklist:**
1. ✅ URL retornada no response?
2. ✅ CORS configurado corretamente?
3. ✅ URL é HTTPS (SecureURL)?
4. ✅ Imagem existe no Cloudinary Dashboard?

## 📚 Referências

- [Cloudinary Go SDK Documentation](https://cloudinary.com/documentation/go_integration)
- [Cloudinary Go Quick Start](https://cloudinary.com/documentation/go_quick_start)
- [Cloudinary Image Transformations](https://cloudinary.com/documentation/image_transformations)
- [Railway Environment Variables](https://docs.railway.app/develop/variables)

## ✅ Checklist de Implementação

- [x] Criar serviço Cloudinary (`pkg/storage/cloudinary.go`)
- [x] Adicionar campos de imagem no modelo Recipe
- [x] Implementar handlers de upload, deleção e otimização
- [x] Adicionar rotas no router
- [x] Criar testes unitários
- [x] Criar testes de integração
- [x] Atualizar documentação (README, INSOMNIA_GUIDE)
- [x] Atualizar collection do Insomnia
- [x] Remover logs de debug
- [x] Melhorar tratamento de erros
- [x] Validar segurança e autorização

## 🎯 Próximos Passos (Opcional)

1. **Adicionar suporte a múltiplas imagens por receita**
   - Galeria de fotos
   - Imagens do passo a passo

2. **Implementar upload direto do frontend**
   - Signed upload URLs
   - Upload widget do Cloudinary

3. **Adicionar watermark automático**
   - Proteção de imagens
   - Branding

4. **Implementar lazy loading**
   - Placeholder blur
   - Progressive JPEG

5. **Analytics de imagens**
   - Imagens mais visualizadas
   - Performance de carregamento

---

**Última atualização:** 29/12/2025
**Versão:** 1.0.0
**Status:** ✅ Implementação completa e testada

