# 🔄 Sistema de Backup e Migração - Guia Rápido

> Sistema completo para backup, restauração e migração do banco de dados PostgreSQL do Railway

## 🚀 Início Rápido

### 1. Fazer Backup (3 minutos)

```bash
# Configure a URL do banco atual
export DATABASE_URL="sua_connection_string_do_railway"

# Execute o backup
./scripts/backup-railway.sh
```

✅ Seu backup estará em: `backups/backup_receitas_TIMESTAMP.sql`

### 2. Criar Nova Conta Railway

1. Crie nova conta Railway (email diferente ou via GitHub/Google)
2. Crie novo projeto → Provision PostgreSQL
3. Copie a nova DATABASE_URL

### 3. Restaurar Dados (5 minutos)

```bash
# Configure a URL do novo banco
export NEW_DATABASE_URL="nova_connection_string_do_railway"

# Restaure o backup
./scripts/restore-railway.sh backups/backup_receitas_TIMESTAMP.sql

# Valide os dados
./scripts/validate-backup.sh
```

## 📚 Documentação Completa

- **[MIGRATION_GUIDE.md](MIGRATION_GUIDE.md)** - Guia detalhado de migração

  - Pré-requisitos e instalação
  - Métodos de backup (SQL e JSON)
  - Passo-a-passo completo
  - Solução de problemas
  - Checklist de migração

- **[BACKUP_IMPLEMENTATION_SUMMARY.md](BACKUP_IMPLEMENTATION_SUMMARY.md)** - Resumo técnico

  - Arquivos criados
  - Funcionalidades implementadas
  - Fluxo de uso
  - Considerações técnicas

- **[scripts/README.md](scripts/README.md)** - Documentação dos scripts

  - Descrição de cada script
  - Exemplos de uso

- **[backups/README.md](backups/README.md)** - Gestão de backups
  - Estrutura de arquivos
  - Segurança
  - Limpeza de backups antigos

## 🛠️ Scripts Disponíveis

### Backup

| Script              | Descrição                     | Uso                           |
| ------------------- | ----------------------------- | ----------------------------- |
| `backup-railway.sh` | Backup SQL completo (pg_dump) | **Recomendado** para migração |
| `backup-json.sh`    | Backup em JSON                | Útil para inspeção visual     |
| `quick-backup.sh`   | Backup rápido comprimido      | Para backups rotineiros       |
| `check-db-size.sh`  | Verifica tamanho do banco     | Antes de fazer backup         |

### Restauração

| Script               | Descrição            | Uso                                  |
| -------------------- | -------------------- | ------------------------------------ |
| `restore-railway.sh` | Restaura backup SQL  | Rápido e confiável                   |
| `restore-json.sh`    | Restaura backup JSON | Portável                             |
| `validate-backup.sh` | Valida integridade   | **Sempre executar** após restauração |

## 📦 Estrutura de Arquivos

```
receitas-back/
├── scripts/
│   ├── backup-railway.sh      ⭐ Backup SQL (recomendado)
│   ├── backup-json.sh          📝 Backup JSON
│   ├── restore-railway.sh      💾 Restaurar SQL
│   ├── restore-json.sh         📥 Restaurar JSON
│   ├── validate-backup.sh      ✅ Validar dados
│   ├── check-db-size.sh        📊 Ver tamanho
│   ├── quick-backup.sh         ⚡ Backup rápido
│   └── README.md               📖 Docs dos scripts
│
├── cmd/
│   ├── backup-db/main.go       🔧 Exportador JSON
│   └── restore-db/main.go      🔧 Importador JSON
│
├── backups/
│   ├── .gitkeep
│   ├── README.md               📖 Gestão de backups
│   ├── backup_*.sql            (gerados)
│   └── json/                   (gerados)
│
├── MIGRATION_GUIDE.md          📘 Guia completo
├── BACKUP_IMPLEMENTATION_SUMMARY.md  📋 Resumo técnico
└── BACKUP_README.md            👈 Você está aqui
```

## ⚡ Exemplos Práticos

### Migração Completa (Modo Fácil)

```bash
# 1. Backup do banco atual
export DATABASE_URL="postgresql://user:pass@containers-us-west-123.railway.app:7432/railway"
./scripts/backup-railway.sh
# Resultado: backups/backup_receitas_20260116_143022.sql

# 2. Criar nova conta Railway e provisionar PostgreSQL
# 3. Copiar nova DATABASE_URL

# 4. Restaurar no novo banco
export NEW_DATABASE_URL="postgresql://newuser:newpass@containers-us-west-456.railway.app:7432/railway"
./scripts/restore-railway.sh backups/backup_receitas_20260116_143022.sql

# 5. Validar
./scripts/validate-backup.sh
```

### Verificar Tamanho Antes de Migrar

```bash
export DATABASE_URL="sua_connection_string"
./scripts/check-db-size.sh
```

**Output esperado:**

```
🔍 Verificando tamanho do banco de dados...

📊 Tamanho do Banco de Dados:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 size
 42 MB

📋 Tamanho por Tabela:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 schemaname | tablename          | size
 public     | ingredients        | 25 MB
 public     | recipes            | 12 MB
 public     | recipe_ingredients | 3 MB
 public     | ratings            | 1.5 MB
 public     | users              | 512 kB

💾 Limite Railway Free Tier: 512 MB
✅ Banco está dentro do limite (8% usado)
```

### Backup com Inspeção Visual (JSON)

```bash
# Fazer backup JSON
export DATABASE_URL="sua_connection_string"
./scripts/backup-json.sh

# Ver dados exportados
cat backups/json/backup_20260116_143500/users.json | jq '.[0]'
```

### Backup Rápido Comprimido

```bash
export DATABASE_URL="sua_connection_string"
./scripts/quick-backup.sh
# Resultado: backups/backup_20260116_143022.sql.gz (comprimido)
```

## 🔒 Segurança e Boas Práticas

### ✅ O que FAZER

- ✅ Sempre validar após restauração (`validate-backup.sh`)
- ✅ Guardar backups em local seguro (não Git)
- ✅ Configurar backup automático semanal
- ✅ Testar restauração periodicamente
- ✅ Verificar tamanho antes de fazer backup
- ✅ Comprimir backups grandes

### ❌ O que NÃO FAZER

- ❌ Não versionar backups no Git (dados sensíveis)
- ❌ Não compartilhar backups publicamente
- ❌ Não fazer backup sem validar após
- ❌ Não deletar backup antigo antes de validar o novo
- ❌ Não migrar sem verificar tamanho do banco

## 🆘 Precisa de Ajuda?

### Problemas Comuns

**"pg_dump: command not found"**

```bash
# macOS
brew install postgresql

# Ubuntu/Debian
sudo apt-get install postgresql-client
```

**"permission denied" nos scripts**

```bash
chmod +x scripts/*.sh
```

**Backup muito grande (> 512MB)**

- Limpe dados antigos
- Use compressão: `./scripts/quick-backup.sh`
- Considere upgrade do plano Railway

### Documentação Detalhada

Consulte [MIGRATION_GUIDE.md](MIGRATION_GUIDE.md) para:

- Solução de problemas detalhada
- Configuração da aplicação após migração
- Checklist completo
- Backup automático (cron/GitHub Actions)

## 📊 Status da Implementação

✅ **Completo e Testado**

| Recurso                | Status |
| ---------------------- | ------ |
| Backup SQL (pg_dump)   | ✅     |
| Backup JSON            | ✅     |
| Restauração SQL        | ✅     |
| Restauração JSON       | ✅     |
| Validação de dados     | ✅     |
| Verificação de tamanho | ✅     |
| Documentação completa  | ✅     |
| Scripts com permissões | ✅     |

## 🎯 Próximos Passos

1. **Agora:** Teste o sistema fazendo um backup

   ```bash
   export DATABASE_URL="sua_url"
   ./scripts/backup-railway.sh
   ```

2. **Quando precisar migrar:** Siga o [MIGRATION_GUIDE.md](MIGRATION_GUIDE.md)

3. **Após migrar:** Configure backup automático (veja guia)

## 💡 Dicas

- **Backup regular:** Configure cron job para backup semanal
- **Teste de restauração:** Teste restauração em banco local periodicamente
- **Monitore tamanho:** Execute `check-db-size.sh` mensalmente
- **Cloudinary:** Imagens não precisam backup, URLs já estão no banco
- **JWT Secret:** Use o mesmo para manter tokens válidos

---

**📅 Implementado em:** Janeiro 16, 2026  
**🔧 Versão:** 1.0  
**📝 Autor:** Sistema de Backup Automatizado

**⭐ Comece agora:** `./scripts/backup-railway.sh`
