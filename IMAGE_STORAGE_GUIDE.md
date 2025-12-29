# 📸 Guia de Armazenamento de Imagens

## 🎯 Visão Geral

Este guia explica como o sistema de upload e armazenamento de imagens funciona na aplicação de receitas, utilizando **Cloudinary** como serviço de hospedagem de imagens.

## 🏗️ Arquitetura

### Por que Cloudinary?

Para aplicações hospedadas em plataformas como Railway, **não é recomendado** armazenar arquivos no filesystem local, pois:

- ❌ Containers podem ser reiniciados e perder dados
- ❌ Não há escalabilidade horizontal (múltiplas instâncias)
- ❌ Sem CDN para entrega rápida
- ❌ Sem otimização automática de imagens

### Vantagens do Cloudinary

- ✅ **Tier Grátis Generoso**: 25 GB storage, 25 GB bandwidth/mês
- ✅ **CDN Global**: Entrega rápida em qualquer lugar do mundo
- ✅ **Otimização Automática**: Compressão, conversão de formato, WebP automático
- ✅ **Transformações On-the-fly**: Resize, crop, filtros sem reprocessar
- ✅ **SDK Go Oficial**: Integração simples e robusta
- ✅ **URLs Amigáveis**: Fácil de usar e cachear

## 🔧 Configuração

### 1. Criar Conta no Cloudinary

1. Acesse [cloudinary.com](https://cloudinary.com) e crie uma conta gratuita
2. Após login, vá em **Dashboard**
3. Copie a **API Environment variable** (formato: `cloudinary://API_KEY:API_SECRET@CLOUD_NAME`)

### 2. Configurar Variável de Ambiente

#### No Railway:

1. Acesse seu projeto no Railway
2. Vá em **Variables**
3. Adicione:
   ```
   CLOUDINARY_URL=cloudinary://123456789012345:AbCdEfGhIjKlMnOpQrStUvWx@your-cloud-name
   ```

#### Localmente (.env):

```bash
CLOUDINARY_URL=cloudinary://123456789012345:AbCdEfGhIjKlMnOpQrStUvWx@your-cloud-name
```

### 3. Instalar Dependências

```bash
go get github.com/cloudinary/cloudinary-go/v2
go get github.com/cloudinary/cloudinary-go/v2/api/uploader
```

## 📋 Modelo de Dados

### Campos Adicionados ao Recipe

```go
type Recipe struct {
    // ... outros campos ...
    ImageURL      string `gorm:"size:500" json:"image_url,omitempty"`        // URL da imagem
    ImagePublicID string `gorm:"size:200" json:"image_public_id,omitempty"` // ID para deletar
}
```

### Migração do Banco de Dados

Ao iniciar a aplicação, o GORM automaticamente adiciona as novas colunas à tabela `recipes`:

```sql
ALTER TABLE recipes ADD COLUMN image_url VARCHAR(500);
ALTER TABLE recipes ADD COLUMN image_public_id VARCHAR(200);
```

## 🚀 Endpoints da API

### 1. Upload de Imagem

**POST** `/recipes/{id}/image`

Faz upload de uma imagem para uma receita.

**Headers:**
```
Authorization: Bearer {token}
Content-Type: multipart/form-data
```

**Body (form-data):**
```
image: [arquivo da imagem]
```

**Restrições:**
- ✅ Formatos aceitos: JPG, JPEG, PNG, GIF, WEBP, BMP
- ✅ Tamanho máximo: 5MB
- ✅ Dimensões máximas: 2000x2000px
- ✅ Apenas o dono da receita ou admin pode fazer upload

**Exemplo (cURL):**
```bash
curl -X POST "http://localhost:8080/recipes/1/image" \
  -H "Authorization: Bearer seu-token-aqui" \
  -F "image=@/path/to/foto-receita.jpg"
```

**Exemplo (Insomnia/Postman):**
1. Método: POST
2. URL: `http://localhost:8080/recipes/1/image`
3. Auth: Bearer Token
4. Body: Form (multipart)
   - Campo: `image`
   - Tipo: File
   - Arquivo: selecionar imagem

**Response (200 OK):**
```json
{
  "message": "Imagem enviada com sucesso",
  "image_url": "https://res.cloudinary.com/seu-cloud/image/upload/v1234567890/recipes/recipe_1_1234567890.jpg",
  "image_public_id": "recipes/recipe_1_1234567890",
  "width": 1920,
  "height": 1080,
  "format": "jpg",
  "size_bytes": 245678
}
```

**Errors:**
- `400`: Imagem inválida ou muito grande
- `401`: Token inválido ou ausente
- `403`: Sem permissão para modificar esta receita
- `404`: Receita não encontrada
- `500`: Erro no upload

### 2. Deletar Imagem

**DELETE** `/recipes/{id}/image`

Remove a imagem de uma receita.

**Headers:**
```
Authorization: Bearer {token}
```

**Exemplo:**
```bash
curl -X DELETE "http://localhost:8080/recipes/1/image" \
  -H "Authorization: Bearer seu-token-aqui"
```

**Response (200 OK):**
```json
{
  "message": "Imagem removida com sucesso"
}
```

**Errors:**
- `401`: Token inválido
- `403`: Sem permissão
- `404`: Receita não encontrada ou sem imagem

### 3. Obter Variantes da Imagem

**GET** `/recipes/{id}/image/variants`

Retorna URLs otimizadas da imagem em diferentes tamanhos (thumbnail, medium, large, original).

**Exemplo:**
```bash
curl "http://localhost:8080/recipes/1/image/variants"
```

**Response (200 OK):**
```json
{
  "thumbnail": {
    "url": "https://res.cloudinary.com/.../w_300,h_300,c_fill,q_auto,f_auto/recipes/recipe_1.jpg",
    "width": 300,
    "height": 300
  },
  "medium": {
    "url": "https://res.cloudinary.com/.../w_600,h_600,c_fill,q_auto,f_auto/recipes/recipe_1.jpg",
    "width": 600,
    "height": 600
  },
  "large": {
    "url": "https://res.cloudinary.com/.../w_1200,h_1200,c_fill,q_auto,f_auto/recipes/recipe_1.jpg",
    "width": 1200,
    "height": 1200
  },
  "original": {
    "url": "https://res.cloudinary.com/.../recipes/recipe_1.jpg"
  }
}
```

### 4. Obter URL Otimizada Customizada

**GET** `/recipes/{id}/image/optimized?width=800&height=600&quality=auto`

Retorna URL otimizada com tamanho e qualidade customizados.

**Query Parameters:**
- `width`: Largura desejada (1-2000, padrão: 800)
- `height`: Altura desejada (1-2000, padrão: 800)
- `quality`: Qualidade (`auto`, `best`, `good`, `eco`, `low`, padrão: `auto`)

**Exemplo:**
```bash
curl "http://localhost:8080/recipes/1/image/optimized?width=400&height=400&quality=good"
```

**Response (200 OK):**
```json
{
  "url": "https://res.cloudinary.com/.../w_400,h_400,c_fill,q_good,f_auto/recipes/recipe_1.jpg",
  "width": 400,
  "height": 400,
  "quality": "good"
}
```

## 💡 Casos de Uso

### Frontend Web

#### 1. Listagem de Receitas (Cards)
```typescript
// Usar thumbnail ou medium para cards
const imageUrl = recipe.image_url 
  ? `${API_URL}/recipes/${recipe.id}/image/optimized?width=400&height=300`
  : '/placeholder.jpg';

<img src={imageUrl} alt={recipe.title} />
```

#### 2. Página de Detalhes da Receita
```typescript
// Usar large para visualização completa
const imageUrl = recipe.image_url
  ? `${API_URL}/recipes/${recipe.id}/image/optimized?width=1200&height=800`
  : '/placeholder.jpg';

<img src={imageUrl} alt={recipe.title} loading="lazy" />
```

#### 3. Upload de Imagem
```typescript
async function uploadRecipeImage(recipeId: number, file: File) {
  const formData = new FormData();
  formData.append('image', file);

  const response = await fetch(`${API_URL}/recipes/${recipeId}/image`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`
    },
    body: formData
  });

  return response.json();
}
```

### Mobile (React Native / Flutter)

#### Thumbnail para Lista
```dart
// Flutter
Image.network(
  'https://api.example.com/recipes/${recipe.id}/image/optimized?width=300&height=300',
  fit: BoxFit.cover,
  loadingBuilder: (context, child, progress) {
    return progress == null ? child : CircularProgressIndicator();
  },
)
```

### Otimização Automática

O Cloudinary automaticamente:
- 📦 Comprime imagens sem perda visível de qualidade (`q_auto`)
- 🎨 Converte para WebP em navegadores que suportam (`f_auto`)
- 🚀 Serve via CDN global (baixa latência)
- 💾 Cacheia transformações (segunda request é instantânea)

## 🔒 Segurança e Autorização

### Permissões de Upload/Delete

- ✅ **Dono da receita**: Pode fazer upload e deletar
- ✅ **Admin**: Pode fazer upload e deletar de qualquer receita
- ❌ **Outros usuários**: Não podem modificar imagens

### Validações Implementadas

1. **Tipo de arquivo**: Apenas imagens permitidas
2. **Tamanho**: Máximo 5MB
3. **Dimensões**: Redimensiona para máximo 2000x2000
4. **Rate Limiting**: Protege contra abuse (20 req/min)
5. **Autenticação**: Token JWT obrigatório

## 📊 Limites do Tier Grátis Cloudinary

| Recurso | Limite Grátis |
|---------|---------------|
| Storage | 25 GB |
| Bandwidth | 25 GB/mês |
| Transformações | 25 créditos/mês |
| Imagens | Ilimitadas |
| API Requests | Ilimitadas |

**Estimativa:**
- ~5.000 imagens de 5MB (storage)
- ~50.000 visualizações/mês (bandwidth)
- Suficiente para MVPs e aplicações pequenas/médias

## 🧪 Testando

### 1. Testar Upload

```bash
# Criar uma receita primeiro
curl -X POST "http://localhost:8080/recipes" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Bolo de Chocolate",
    "description": "Delicioso bolo",
    "prep_time": 45,
    "servings": 8,
    "difficulty": "média"
  }'

# Upload de imagem
curl -X POST "http://localhost:8080/recipes/1/image" \
  -H "Authorization: Bearer $TOKEN" \
  -F "image=@foto-bolo.jpg"
```

### 2. Verificar Receita com Imagem

```bash
curl "http://localhost:8080/recipes/1"
```

Response incluirá:
```json
{
  "id": 1,
  "title": "Bolo de Chocolate",
  "image_url": "https://res.cloudinary.com/.../recipes/recipe_1.jpg",
  "image_public_id": "recipes/recipe_1_1234567890",
  // ... outros campos
}
```

### 3. Testar Variantes

```bash
curl "http://localhost:8080/recipes/1/image/variants"
```

### 4. Deletar Imagem

```bash
curl -X DELETE "http://localhost:8080/recipes/1/image" \
  -H "Authorization: Bearer $TOKEN"
```

## 🛠️ Troubleshooting

### Erro: "CLOUDINARY_URL não configurado"

**Causa:** Variável de ambiente não definida

**Solução:**
```bash
# Verificar no Railway
railway variables

# Verificar localmente
echo $CLOUDINARY_URL
```

### Erro: "formato de arquivo não suportado"

**Causa:** Tentou fazer upload de arquivo não-imagem

**Solução:** Apenas JPG, PNG, GIF, WEBP, BMP são aceitos

### Erro: "imagem muito grande"

**Causa:** Imagem maior que 5MB

**Solução:** Comprimir imagem antes de fazer upload ou aumentar `maxImageSizeMB`

### Imagens não aparecem no Cloudinary Dashboard

**Causa:** Pasta pode estar diferente

**Solução:** No dashboard, verificar pasta "recipes" ou pesquisar por "recipe_"

## 🚀 Melhorias Futuras

- [ ] Suporte para múltiplas imagens por receita
- [ ] Upload de vídeos de preparo
- [ ] Marcas d'água automáticas
- [ ] Reconhecimento de imagem (AI) para sugerir ingredientes
- [ ] Moderação automática de conteúdo
- [ ] Thumbnail animado (GIF) a partir de vídeo
- [ ] Upload direto do frontend para Cloudinary (signed upload)

## 📚 Referências

- [Documentação Cloudinary Go SDK](https://cloudinary.com/documentation/go_integration)
- [Transformações de Imagem](https://cloudinary.com/documentation/image_transformations)
- [Otimização Automática](https://cloudinary.com/documentation/image_optimization)
- [Railway Deployment](https://docs.railway.app/)

## 💬 Suporte

Se encontrar problemas ou tiver dúvidas:
1. Verifique se `CLOUDINARY_URL` está configurado corretamente
2. Teste com imagens pequenas primeiro (< 1MB)
3. Verifique logs da aplicação para erros detalhados
4. Consulte o dashboard do Cloudinary para ver se uploads estão chegando

