# ✅ Checklist de Deploy - Sistema de Imagens

## 🚀 Railway Deployment

### Antes do Deploy

- [ ] **Criar conta Cloudinary** (grátis)
  - Acesse: https://cloudinary.com/users/register/free
  - Confirme email

- [ ] **Copiar credenciais Cloudinary**
  - Login no dashboard
  - Copiar "API Environment variable"
  - Formato: `cloudinary://KEY:SECRET@CLOUD_NAME`

- [ ] **Código atualizado no Git**
  ```bash
  git add .
  git commit -m "feat: adicionar sistema de upload de imagens"
  git push origin main
  ```

### Durante o Deploy

- [ ] **Configurar variável no Railway**
  1. Acessar projeto no Railway
  2. Clicar no serviço
  3. **Variables** → **New Variable**
  4. Nome: `CLOUDINARY_URL`
  5. Valor: `cloudinary://123:abc@cloud` (sua URL)
  6. Salvar

- [ ] **Aguardar deploy automático**
  - Railway detecta push no Git
  - Build e deploy automáticos
  - Ver logs em tempo real

- [ ] **Verificar logs**
  - Procurar por erros
  - Confirmar que aplicação iniciou

### Após o Deploy

- [ ] **Testar health check**
  ```bash
  curl https://seu-app.railway.app/health
  ```

- [ ] **Fazer login**
  ```bash
  curl -X POST https://seu-app.railway.app/users/login \
    -H "Content-Type: application/json" \
    -d '{"email":"admin@admin.com","password":"admin123"}'
  ```

- [ ] **Criar receita de teste**
  ```bash
  curl -X POST https://seu-app.railway.app/recipes \
    -H "Authorization: Bearer SEU_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
      "title": "Teste Deploy",
      "prep_time": 10,
      "servings": 1,
      "difficulty": "fácil"
    }'
  ```

- [ ] **Testar upload de imagem**
  ```bash
  curl -X POST https://seu-app.railway.app/recipes/1/image \
    -H "Authorization: Bearer SEU_TOKEN" \
    -F "image=@teste.jpg"
  ```

- [ ] **Verificar imagem no Cloudinary**
  - Acessar Media Library
  - Procurar pasta "recipes"
  - Confirmar imagem está lá

- [ ] **Testar URLs otimizadas**
  ```bash
  curl https://seu-app.railway.app/recipes/1/image/variants
  ```

## 🔧 Troubleshooting

### Erro: "CLOUDINARY_URL não configurado"

**Causa**: Variável não foi adicionada ou está incorreta

**Solução**:
1. Railway → Variables
2. Verificar se `CLOUDINARY_URL` existe
3. Formato correto: `cloudinary://API_KEY:API_SECRET@CLOUD_NAME`
4. Reiniciar serviço manualmente se necessário

### Erro: Build failed

**Causa**: Dependências não instaladas

**Solução**:
```bash
# Localmente
go mod tidy
git add go.mod go.sum
git commit -m "fix: atualizar dependências"
git push
```

### Erro: Upload funciona mas retorna 500

**Causa**: Credenciais Cloudinary inválidas

**Solução**:
1. Verificar se CLOUDINARY_URL está correta
2. Copiar novamente do dashboard
3. Testar localmente primeiro

### Imagem não aparece

**Causa**: URL pode estar bloqueada

**Solução**:
1. Testar URL da imagem diretamente no navegador
2. Verificar configurações de CORS no Cloudinary
3. Verificar se imagem foi realmente enviada (Media Library)

## 📊 Monitoramento

### Uso do Cloudinary

- [ ] **Verificar uso mensal**
  - Dashboard → Reports
  - Storage usado
  - Bandwidth usado
  - Transformações

### Limites do Tier Grátis

| Recurso | Limite |
|---------|--------|
| Storage | 25 GB |
| Bandwidth | 25 GB/mês |
| Transformações | 25 créditos/mês |

**Dica**: Configurar alertas quando atingir 80% do limite

### Logs Railway

- [ ] **Monitorar logs regularmente**
  - Erros de upload
  - Tentativas de acesso não autorizado
  - Performance

## 🔒 Segurança

### Produção

- [ ] **JWT_SECRET forte**
  ```bash
  # Gerar secret seguro
  openssl rand -base64 32
  ```

- [ ] **Rate limiting ativo**
  - Verificar `RATE_LIMIT_ENABLED=true`

- [ ] **HTTPS ativo**
  - Railway fornece automaticamente

- [ ] **Cloudinary signed URLs** (opcional, para proteção extra)
  - Configurar no código se necessário

## 📝 Documentação

- [ ] **Atualizar README.md** ✅
- [ ] **Criar IMAGE_STORAGE_GUIDE.md** ✅
- [ ] **Criar QUICKSTART_IMAGES.md** ✅
- [ ] **Atualizar INSOMNIA_GUIDE.md** ✅
- [ ] **Compartilhar com equipe**

## 🎯 Testes em Produção

### Fluxo Completo

1. [ ] Registrar novo usuário
2. [ ] Fazer login
3. [ ] Criar receita
4. [ ] Upload de imagem
5. [ ] Ver receita com imagem
6. [ ] Obter variantes
7. [ ] Deletar imagem
8. [ ] Upload nova imagem
9. [ ] Deletar receita (deve deletar imagem também)

### Performance

- [ ] **Testar upload de diferentes tamanhos**
  - 100KB
  - 1MB
  - 5MB (máximo)

- [ ] **Testar diferentes formatos**
  - JPG
  - PNG
  - WEBP
  - GIF

- [ ] **Verificar tempo de resposta**
  - Upload < 3s (depende da imagem e conexão)
  - GET variantes < 500ms
  - GET otimizada < 500ms

## 🔄 Rollback

### Se algo der errado:

1. **Reverter deploy**
   ```bash
   git revert HEAD
   git push origin main
   ```

2. **Ou voltar para commit anterior**
   - Railway → Deployments
   - Selecionar deploy anterior
   - Redeploy

3. **Remover variável Cloudinary** (temporário)
   - Railway → Variables
   - Deletar `CLOUDINARY_URL`

## 🎉 Deploy Completo!

Após completar todos os itens:

- ✅ Sistema de imagens funcionando
- ✅ Testes passando
- ✅ Monitoramento ativo
- ✅ Documentação atualizada

**Próximos passos:**
- Implementar frontend para upload visual
- Adicionar múltiplas imagens por receita
- Implementar cropping de imagens
- Adicionar imagens para ingredientes

