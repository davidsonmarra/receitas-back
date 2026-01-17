#!/bin/bash
# Script de Backup via pg_dump
# Cria um dump completo do banco PostgreSQL do Railway
# Uso: ./scripts/backup-railway.sh

set -e  # Para em caso de erro

echo "🔄 Iniciando backup do banco de dados..."

# Verifica se DATABASE_URL está definida
if [ -z "$DATABASE_URL" ]; then
    echo "❌ Erro: DATABASE_URL não está definida"
    echo "💡 Defina a variável: export DATABASE_URL='sua_connection_string'"
    exit 1
fi

# Cria diretório de backups se não existir
BACKUP_DIR="backups"
mkdir -p "$BACKUP_DIR"

# Nome do arquivo com timestamp
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/backup_receitas_$TIMESTAMP.sql"

echo "📦 Criando dump do banco..."
# Flag --no-sync para evitar problemas de versão
pg_dump "$DATABASE_URL" --no-sync 2>/dev/null > "$BACKUP_FILE" || \
  pg_dump "$DATABASE_URL" > "$BACKUP_FILE"

if [ $? -eq 0 ]; then
    FILE_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
    echo "✅ Backup criado com sucesso!"
    echo "📄 Arquivo: $BACKUP_FILE"
    echo "📊 Tamanho: $FILE_SIZE"
    echo ""
    echo "💾 Para restaurar em um novo banco, use:"
    echo "   export NEW_DATABASE_URL='nova_connection_string'"
    echo "   ./scripts/restore-railway.sh $BACKUP_FILE"
else
    echo "❌ Erro ao criar backup"
    exit 1
fi

