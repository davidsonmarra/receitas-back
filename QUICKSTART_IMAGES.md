# 🚀 Quickstart: Upload de Imagens

## ⚡ Setup Rápido (5 minutos)

### 1. Criar Conta Cloudinary (2 min)

1. Acesse: https://cloudinary.com/users/register/free
2. Crie conta gratuita (email + senha)
3. Confirme email

### 2. Copiar Credenciais (1 min)

1. Faça login no [Dashboard Cloudinary](https://console.cloudinary.com/)
2. Na página inicial, procure por **"API Environment variable"**
3. Copie a URL completa (formato: `cloudinary://123:abc@name`)

### 3. Configurar no Railway (1 min)

1. Acesse seu projeto no [Railway](https://railway.app)
2. Clique no serviço da sua aplicação
3. Vá em **Variables**
4. Adicione nova variável:
   - **Name**: `CLOUDINARY_URL`
   - **Value**: Cole a URL copiada (ex: `cloudinary://123456:AbCdEf@mycloud`)
5. Salve (a aplicação reiniciará automaticamente)

### 4. Testar Localmente (1 min)

```bash
# Adicionar ao seu .env
echo 'CLOUDINARY_URL=cloudinary://123456:AbCdEf@mycloud' >> .env

# Reinstalar dependências
go mod tidy

# Rodar aplicação
go run ./cmd/api/main.go
```

## 📸 Testar Upload

### Com cURL:

```bash
# 1. Fazer login (obter token)
TOKEN=$(curl -X POST http://localhost:8080/users/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@admin.com","password":"admin123"}' \
  | jq -r '.token')

# 2. Criar uma receita
RECIPE_ID=$(curl -X POST http://localhost:8080/recipes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Bolo de Chocolate",
    "description": "Delicioso bolo",
    "prep_time": 45,
    "servings": 8,
    "difficulty": "média"
  }' | jq -r '.id')

# 3. Upload de imagem
curl -X POST "http://localhost:8080/recipes/$RECIPE_ID/image" \
  -H "Authorization: Bearer $TOKEN" \
  -F "image=@/path/to/sua-foto.jpg"
```

### Com Insomnia/Postman:

1. **Fazer Login**
   - POST `http://localhost:8080/users/login`
   - Body (JSON):
     ```json
     {"email": "admin@admin.com", "password": "admin123"}
     ```
   - Copie o `token` da resposta

2. **Criar Receita**
   - POST `http://localhost:8080/recipes`
   - Auth: Bearer Token (cole o token)
   - Body (JSON):
     ```json
     {
       "title": "Bolo de Chocolate",
       "description": "Delicioso bolo",
       "prep_time": 45,
       "servings": 8,
       "difficulty": "média"
     }
     ```
   - Copie o `id` da resposta

3. **Upload de Imagem**
   - POST `http://localhost:8080/recipes/1/image` (substitua 1 pelo id da receita)
   - Auth: Bearer Token
   - Body: **Multipart Form**
     - Adicione campo `image` do tipo **File**
     - Selecione uma imagem do seu computador
   - Enviar

4. **Ver Receita com Imagem**
   - GET `http://localhost:8080/recipes/1`
   - Resposta terá `image_url`

## ✅ Verificar se Funcionou

### Na API:

```bash
curl http://localhost:8080/recipes/1 | jq
```

Resposta deve incluir:
```json
{
  "id": 1,
  "title": "Bolo de Chocolate",
  "image_url": "https://res.cloudinary.com/.../recipes/recipe_1.jpg",
  "image_public_id": "recipes/recipe_1_1234567890",
  ...
}
```

### No Cloudinary Dashboard:

1. Acesse [Media Library](https://console.cloudinary.com/console/media_library)
2. Procure pela pasta **"recipes"**
3. Você verá sua imagem uploadada

## 🎨 Usar Imagens no Frontend

### URL Original:
```
https://res.cloudinary.com/seu-cloud/image/upload/v1234/recipes/recipe_1.jpg
```

### URL Otimizada (automática):
```javascript
// React/Vue/Angular
const imageUrl = `https://api.example.com/recipes/${id}/image/optimized?width=600&height=400`;

<img src={imageUrl} alt={recipe.title} />
```

### Diferentes Tamanhos:

```javascript
// Buscar todas as variantes
fetch(`/recipes/${id}/image/variants`)
  .then(res => res.json())
  .then(data => {
    console.log(data.thumbnail.url);  // 300x300
    console.log(data.medium.url);     // 600x600
    console.log(data.large.url);      // 1200x1200
    console.log(data.original.url);   // original
  });
```

## 🔄 Atualizar Imagem

```bash
# Simplesmente fazer novo upload (substitui automaticamente)
curl -X POST "http://localhost:8080/recipes/$RECIPE_ID/image" \
  -H "Authorization: Bearer $TOKEN" \
  -F "image=@nova-foto.jpg"
```

## 🗑️ Deletar Imagem

```bash
curl -X DELETE "http://localhost:8080/recipes/$RECIPE_ID/image" \
  -H "Authorization: Bearer $TOKEN"
```

## ❓ Troubleshooting Rápido

### ❌ "CLOUDINARY_URL não configurado"
**Solução:** Verifique se adicionou a variável no Railway/local

```bash
# Local
echo $CLOUDINARY_URL

# Railway
railway variables
```

### ❌ "formato de arquivo não suportado"
**Solução:** Apenas JPG, PNG, GIF, WEBP aceitos

### ❌ "imagem muito grande"
**Solução:** Reduza para menos de 5MB

### ❌ Upload funciona mas imagem não aparece
**Solução:** Verifique a URL no navegador diretamente

## 📚 Documentação Completa

Para mais detalhes, veja:
- [IMAGE_STORAGE_GUIDE.md](./IMAGE_STORAGE_GUIDE.md) - Guia completo
- [INSOMNIA_GUIDE.md](./INSOMNIA_GUIDE.md) - Exemplos de requisições

## 🎯 Próximos Passos

- [ ] Criar frontend para upload visual
- [ ] Adicionar cropping de imagens
- [ ] Suporte para múltiplas imagens por receita
- [ ] Adicionar imagens para ingredientes também

