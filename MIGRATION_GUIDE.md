# Guia de Migração do Banco de Dados Railway

Este guia explica como fazer backup completo do seu banco de dados PostgreSQL no Railway e restaurá-lo em uma nova conta gratuita.

## 📋 Índice

1. [Pré-requisitos](#pré-requisitos)
2. [Método 1: Backup via pg_dump (Recomendado)](#método-1-backup-via-pg_dump-recomendado)
3. [Método 2: Backup via JSON](#método-2-backup-via-json)
4. [Validação dos Dados](#validação-dos-dados)
5. [Configuração da Nova Aplicação](#configuração-da-nova-aplicação)
6. [Solução de Problemas](#solução-de-problemas)

---

## Pré-requisitos

### Ferramentas Necessárias

1. **PostgreSQL Client Tools** instalado localmente
   ```bash
   # macOS
   brew install postgresql

   # Ubuntu/Debian
   sudo apt-get install postgresql-client

   # Windows
   # Baixe e instale do site oficial: https://www.postgresql.org/download/windows/
   ```

2. **Go** (para método JSON)
   - Já deve estar instalado se você está desenvolvendo o backend

3. **Permissões de execução** para scripts bash
   ```bash
   chmod +x scripts/*.sh
   ```

### Informações Necessárias

Você precisará de:
- `DATABASE_URL` do banco atual (Railway)
- `NEW_DATABASE_URL` do novo banco (nova conta Railway)

Para obter a `DATABASE_URL` no Railway:
1. Acesse seu projeto no Railway
2. Clique no serviço PostgreSQL
3. Vá na aba "Connect"
4. Copie a "Postgres Connection URL"

---

## Método 1: Backup via pg_dump (Recomendado)

Este é o método **mais confiável e rápido** para migração completa.

### Passo 1: Fazer Backup do Banco Atual

```bash
# Definir a DATABASE_URL do banco atual
export DATABASE_URL="postgresql://usuario:senha@host:porta/database"

# Executar o backup
./scripts/backup-railway.sh
```

Isso criará um arquivo em `backups/backup_receitas_TIMESTAMP.sql`.

**Exemplo de saída:**
```
🔄 Iniciando backup do banco de dados...
📦 Criando dump do banco...
✅ Backup criado com sucesso!
📄 Arquivo: backups/backup_receitas_20260116_143022.sql
📊 Tamanho: 2.4M
```

### Passo 2: Criar Nova Conta e Projeto no Railway

1. **Criar nova conta Railway**
   - Use um email diferente ou crie conta via GitHub/Google
   - Aproveite os $5 de crédito gratuito do trial

2. **Criar novo projeto**
   - Clique em "New Project"
   - Selecione "Provision PostgreSQL"
   - Aguarde a criação do banco

3. **Copiar nova DATABASE_URL**
   - Clique no serviço PostgreSQL
   - Copie a "Postgres Connection URL"

### Passo 3: Restaurar no Novo Banco

```bash
# Definir a DATABASE_URL do novo banco
export NEW_DATABASE_URL="postgresql://novo_usuario:nova_senha@novo_host:porta/database"

# Executar a restauração
./scripts/restore-railway.sh backups/backup_receitas_20260116_143022.sql
```

O script pedirá confirmação antes de sobrescrever os dados.

**Exemplo de saída:**
```
🔄 Iniciando restauração do banco de dados...
📄 Arquivo de backup: backups/backup_receitas_20260116_143022.sql
📊 Tamanho: 2.4M

⚠️  Isso irá SOBRESCREVER todos os dados no novo banco. Continuar? (s/N): s
📦 Restaurando dump no novo banco...
✅ Restauração concluída com sucesso!
```

### Passo 4: Validar a Migração

```bash
# Validar os dados restaurados
./scripts/validate-backup.sh
```

Este script verificará:
- ✅ Contagem de registros em cada tabela
- ✅ Integridade das foreign keys
- ✅ Sequences (auto-increment) configuradas corretamente
- ✅ Índices criados

---

## Método 2: Backup via JSON

Este método é útil se você quiser **inspecionar visualmente** os dados ou se tiver problemas com pg_dump.

### Passo 1: Fazer Backup em JSON

```bash
# Definir a DATABASE_URL do banco atual
export DATABASE_URL="postgresql://usuario:senha@host:porta/database"

# Executar o backup JSON
./scripts/backup-json.sh
```

Isso criará uma pasta em `backups/json/backup_TIMESTAMP/` com arquivos:
- `users.json` - Todos os usuários
- `ingredients.json` - Todos os ingredientes
- `recipes.json` - Todas as receitas
- `recipe_ingredients.json` - Relacionamentos receita-ingrediente
- `ratings.json` - Todas as avaliações

### Passo 2: Restaurar do JSON

```bash
# Definir a DATABASE_URL do novo banco
export NEW_DATABASE_URL="postgresql://novo_usuario:nova_senha@novo_host:porta/database"

# Executar a restauração
./scripts/restore-json.sh backups/json/backup_20260116_143500
```

⚠️ **IMPORTANTE:** Este método irá:
1. Dropar todas as tabelas existentes
2. Recriar o schema
3. Importar todos os dados
4. Atualizar as sequences

---

## Validação dos Dados

Após qualquer método de restauração, **sempre execute a validação:**

```bash
export NEW_DATABASE_URL="sua_nova_connection_string"
./scripts/validate-backup.sh
```

### O que o Script de Validação Verifica

1. **Contagem de Registros**
   ```
   📊 Contagem de registros por tabela:
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
     users:               25 registros
     ingredients:         1543 registros
     recipes:             87 registros
     recipe_ingredients:  342 registros
     ratings:             156 registros
   ```

2. **Integridade Referencial**
   - Verifica se todos os `user_id` em recipes existem na tabela users
   - Verifica se todos os `recipe_id` e `ingredient_id` em recipe_ingredients são válidos
   - Verifica se todos os `recipe_id` e `user_id` em ratings são válidos

3. **Sequences (Auto-increment)**
   - Confirma que os sequences estão configurados para o próximo ID correto
   - Evita erros de "duplicate key" ao criar novos registros

4. **Índices**
   - Lista os índices criados para verificar otimização

### Validação Manual Adicional

Após a validação automatizada, teste manualmente:

```bash
# Conectar ao novo banco
psql "$NEW_DATABASE_URL"

# Verificar alguns registros
SELECT COUNT(*) FROM users;
SELECT COUNT(*) FROM recipes;
SELECT * FROM users LIMIT 5;
SELECT * FROM recipes LIMIT 5;
```

---

## Configuração da Nova Aplicação

### 1. Atualizar Variáveis de Ambiente no Railway

Na nova conta Railway, configure o serviço da API:

1. **Criar novo serviço para a API**
   - "New Service" → "GitHub Repo"
   - Selecione seu repositório
   - Configure a branch

2. **Adicionar variáveis de ambiente**
   ```
   DATABASE_URL=postgresql://... (automático do PostgreSQL)
   JWT_SECRET=seu_jwt_secret_aqui
   CLOUDINARY_CLOUD_NAME=seu_cloud_name
   CLOUDINARY_API_KEY=sua_api_key
   CLOUDINARY_API_SECRET=sua_api_secret
   GEMINI_API_KEY=sua_gemini_key (se usar)
   ENV=production
   PORT=8080
   ```

   ⚠️ **IMPORTANTE sobre JWT_SECRET:**
   - Se você usar o **mesmo JWT_SECRET**, os tokens antigos continuarão válidos
   - Se você usar um **novo JWT_SECRET**, todos os usuários precisarão fazer login novamente

3. **Configurar domínio público**
   - Settings → "Generate Domain"
   - Copie o domínio público (ex: `sua-api.up.railway.app`)

### 2. Atualizar App Mobile

Atualize o arquivo `config.json` no app:

```json
{
  "API_BASE_URL": "https://sua-nova-api.up.railway.app",
  "ENABLE_LOGS": false
}
```

### 3. Testar Funcionalidades Principais

#### a) Testar Autenticação

```bash
# Login de usuário existente
curl -X POST https://sua-nova-api.up.railway.app/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "usuario@exemplo.com",
    "password": "senha123"
  }'
```

#### b) Testar Receitas

```bash
# Listar receitas
curl https://sua-nova-api.up.railway.app/api/v1/recipes
```

#### c) Testar Imagens

- As imagens no Cloudinary **não precisam ser migradas**
- Os URLs já estão salvos na tabela `recipes`
- Apenas verifique se estão carregando corretamente no app

---

## Solução de Problemas

### Erro: "pg_dump: command not found"

**Solução:** Instale o PostgreSQL client:
```bash
# macOS
brew install postgresql

# Ubuntu
sudo apt-get install postgresql-client
```

### Erro: "connection refused" ou "timeout"

**Possíveis causas:**
1. DATABASE_URL incorreta
2. Firewall bloqueando conexão
3. Banco não está rodando

**Solução:**
```bash
# Testar conectividade
psql "$DATABASE_URL" -c "SELECT 1;"
```

### Erro: "permission denied" ao executar scripts

**Solução:**
```bash
chmod +x scripts/*.sh
```

### Erro: "duplicate key" após restauração JSON

**Causa:** Sequences não foram atualizadas corretamente.

**Solução:**
```bash
# Conectar ao banco
psql "$NEW_DATABASE_URL"

# Atualizar sequences manualmente
SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));
SELECT setval('ingredients_id_seq', (SELECT MAX(id) FROM ingredients));
SELECT setval('recipes_id_seq', (SELECT MAX(id) FROM recipes));
SELECT setval('recipe_ingredients_id_seq', (SELECT MAX(id) FROM recipe_ingredients));
SELECT setval('ratings_id_seq', (SELECT MAX(id) FROM ratings));
```

### Erro: "out of memory" durante pg_dump

**Solução:** Use streaming para arquivos grandes:
```bash
pg_dump "$DATABASE_URL" | gzip > backup.sql.gz
gunzip -c backup.sql.gz | psql "$NEW_DATABASE_URL"
```

### Diferença na contagem de registros

**Verificar:**
1. O backup incluiu registros deletados (soft delete)?
2. Houve inserções/deleções durante o backup?

**Solução:** Pause a aplicação durante o backup:
```bash
# No Railway, escale para 0 réplicas temporariamente
# Faça o backup
# Restaure o serviço
```

---

## Checklist de Migração Completa

Use este checklist para garantir uma migração bem-sucedida:

### Antes do Backup
- [ ] PostgreSQL client instalado localmente
- [ ] Scripts com permissão de execução (`chmod +x`)
- [ ] DATABASE_URL do banco atual copiada
- [ ] Aplicação pausada (opcional, para consistência máxima)

### Durante o Backup
- [ ] Backup executado com sucesso
- [ ] Arquivo de backup criado e verificado
- [ ] Tamanho do backup razoável (< 500MB para Railway free)

### Nova Conta Railway
- [ ] Nova conta criada com email diferente
- [ ] Novo projeto criado
- [ ] PostgreSQL provisionado
- [ ] NEW_DATABASE_URL copiada

### Restauração
- [ ] Backup restaurado no novo banco
- [ ] Script de validação executado
- [ ] Contagem de registros conferida
- [ ] Integridade referencial verificada

### Configuração da API
- [ ] Serviço da API criado no Railway
- [ ] Todas as variáveis de ambiente configuradas
- [ ] DATABASE_URL apontando para o novo PostgreSQL
- [ ] Domínio público gerado
- [ ] API rodando e acessível

### Configuração do App
- [ ] config.json atualizado com novo API_BASE_URL
- [ ] App testado em desenvolvimento
- [ ] Login testado
- [ ] Listagem de receitas testada
- [ ] Imagens carregando corretamente

### Testes Finais
- [ ] Criar nova receita
- [ ] Editar receita existente
- [ ] Adicionar avaliação
- [ ] Upload de imagem
- [ ] Busca de receitas
- [ ] Logout e login novamente

### Pós-Migração
- [ ] Conta antiga do Railway cancelada (após confirmar que tudo funciona)
- [ ] Backups regulares configurados
- [ ] Documentação atualizada
- [ ] Time/usuários notificados da nova URL

---

## Backup Automático Regular

Para evitar perda de dados, configure backups regulares:

### Opção 1: Cron Job Local

```bash
# Adicione ao crontab (crontab -e)
# Backup semanal todo domingo às 3h
0 3 * * 0 cd /caminho/para/receitas-back && ./scripts/backup-railway.sh
```

### Opção 2: GitHub Actions

Crie `.github/workflows/backup.yml`:

```yaml
name: Database Backup

on:
  schedule:
    - cron: '0 3 * * 0'  # Todo domingo às 3h
  workflow_dispatch:  # Permitir execução manual

jobs:
  backup:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install PostgreSQL client
        run: sudo apt-get install postgresql-client
      
      - name: Run backup
        env:
          DATABASE_URL: ${{ secrets.DATABASE_URL }}
        run: ./scripts/backup-railway.sh
      
      - name: Upload backup
        uses: actions/upload-artifact@v3
        with:
          name: database-backup
          path: backups/*.sql
          retention-days: 30
```

---

## Recursos Adicionais

- [Documentação Railway](https://docs.railway.app/)
- [PostgreSQL Backup Documentation](https://www.postgresql.org/docs/current/backup.html)
- [GORM Documentation](https://gorm.io/docs/)

---

## Suporte

Se encontrar problemas:

1. Verifique a seção [Solução de Problemas](#solução-de-problemas)
2. Execute o script de validação para diagnóstico
3. Consulte os logs do Railway (aba "Deployments" → "View Logs")
4. Verifique as variáveis de ambiente

---

**Última atualização:** Janeiro 2026

