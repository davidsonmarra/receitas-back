# Scripts de Backup e Migração

Esta pasta contém scripts para fazer backup e restauração do banco de dados PostgreSQL.

## 📁 Scripts Disponíveis

### Backup

- **`backup-railway.sh`** - Cria backup SQL completo usando pg_dump (Recomendado)
  ```bash
  export DATABASE_URL="sua_connection_string"
  ./scripts/backup-railway.sh
  ```

- **`backup-json.sh`** - Exporta dados para arquivos JSON (útil para inspeção visual)
  ```bash
  export DATABASE_URL="sua_connection_string"
  ./scripts/backup-json.sh
  ```

### Restauração

- **`restore-railway.sh`** - Restaura backup SQL em novo banco
  ```bash
  export NEW_DATABASE_URL="nova_connection_string"
  ./scripts/restore-railway.sh backups/backup_receitas_20260116_143022.sql
  ```

- **`restore-json.sh`** - Importa dados dos arquivos JSON
  ```bash
  export NEW_DATABASE_URL="nova_connection_string"
  ./scripts/restore-json.sh backups/json/backup_20260116_143500
  ```

### Validação

- **`validate-backup.sh`** - Valida integridade dos dados após restauração
  ```bash
  export NEW_DATABASE_URL="nova_connection_string"
  ./scripts/validate-backup.sh
  ```

## 🚀 Uso Rápido

### Migração Completa em 3 Passos

```bash
# 1. Fazer backup do banco atual
export DATABASE_URL="postgresql://user:pass@host:5432/db"
./scripts/backup-railway.sh

# 2. Restaurar em novo banco
export NEW_DATABASE_URL="postgresql://new_user:new_pass@new_host:5432/new_db"
./scripts/restore-railway.sh backups/backup_receitas_*.sql

# 3. Validar dados
./scripts/validate-backup.sh
```

## 📖 Documentação Completa

Veja o [MIGRATION_GUIDE.md](../MIGRATION_GUIDE.md) para instruções detalhadas, solução de problemas e checklist completo.

## ⚠️ Importante

- Os scripts precisam de permissão de execução: `chmod +x scripts/*.sh`
- PostgreSQL client tools devem estar instalados
- Sempre valide os dados após restauração
- Faça backup antes de qualquer operação crítica

