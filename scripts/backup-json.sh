#!/bin/bash
# Script de Backup em formato JSON
# Exporta cada tabela para um arquivo JSON separado
# Uso: ./scripts/backup-json.sh

set -e  # Para em caso de erro

echo "🔄 Iniciando backup em formato JSON..."

# Verifica se DATABASE_URL está definida
if [ -z "$DATABASE_URL" ]; then
    echo "❌ Erro: DATABASE_URL não está definida"
    echo "💡 Defina a variável: export DATABASE_URL='sua_connection_string'"
    exit 1
fi

# Cria diretório de backups se não existir
BACKUP_DIR="backups/json"
mkdir -p "$BACKUP_DIR"

# Timestamp para o backup
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FOLDER="$BACKUP_DIR/backup_$TIMESTAMP"
mkdir -p "$BACKUP_FOLDER"

echo "📦 Exportando tabelas para JSON..."

# Executa o comando Go para exportar
cd "$(dirname "$0")/.."
go run cmd/backup-db/main.go "$BACKUP_FOLDER"

if [ $? -eq 0 ]; then
    echo "✅ Backup JSON criado com sucesso!"
    echo "📁 Pasta: $BACKUP_FOLDER"
    echo ""
    echo "Arquivos criados:"
    ls -lh "$BACKUP_FOLDER"
    echo ""
    echo "💾 Para restaurar em um novo banco, use:"
    echo "   export NEW_DATABASE_URL='nova_connection_string'"
    echo "   ./scripts/restore-json.sh $BACKUP_FOLDER"
else
    echo "❌ Erro ao criar backup JSON"
    exit 1
fi

