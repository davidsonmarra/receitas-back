#!/bin/bash
# Script rápido de backup com data única
# Uso: ./scripts/quick-backup.sh

set -e

echo "⚡ Backup Rápido do Banco de Dados"
echo ""

# Verifica DATABASE_URL
if [ -z "$DATABASE_URL" ]; then
    echo "❌ DATABASE_URL não definida"
    exit 1
fi

# Criar diretório
mkdir -p backups

# Backup com timestamp
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="backups/backup_$TIMESTAMP.sql"

echo "📦 Criando backup..."
pg_dump "$DATABASE_URL" > "$BACKUP_FILE"

# Comprimir para economizar espaço
echo "🗜️  Comprimindo..."
gzip "$BACKUP_FILE"

FILE_SIZE=$(du -h "$BACKUP_FILE.gz" | cut -f1)
echo "✅ Pronto! Arquivo: $BACKUP_FILE.gz ($FILE_SIZE)"

