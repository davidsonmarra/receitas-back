#!/bin/bash
# Script de Restauração via JSON
# Importa arquivos JSON para o banco PostgreSQL
# Uso: ./scripts/restore-json.sh <pasta_backup>

set -e  # Para em caso de erro

echo "🔄 Iniciando restauração via JSON..."

# Verifica se a pasta de backup foi fornecida
if [ -z "$1" ]; then
    echo "❌ Erro: Pasta de backup não fornecida"
    echo "💡 Uso: ./scripts/restore-json.sh <pasta_backup>"
    echo ""
    echo "Pastas disponíveis:"
    ls -d backups/json/backup_* 2>/dev/null || echo "  Nenhum backup JSON encontrado"
    exit 1
fi

BACKUP_FOLDER="$1"

# Verifica se a pasta existe
if [ ! -d "$BACKUP_FOLDER" ]; then
    echo "❌ Erro: Pasta não encontrada: $BACKUP_FOLDER"
    exit 1
fi

# Verifica se NEW_DATABASE_URL está definida
if [ -z "$NEW_DATABASE_URL" ]; then
    echo "❌ Erro: NEW_DATABASE_URL não está definida"
    echo "💡 Defina a variável: export NEW_DATABASE_URL='nova_connection_string'"
    exit 1
fi

echo "📁 Pasta de backup: $BACKUP_FOLDER"
echo "📄 Arquivos encontrados:"
ls -lh "$BACKUP_FOLDER"/*.json
echo ""

# Pergunta de confirmação
read -p "⚠️  Isso irá SOBRESCREVER todos os dados no novo banco. Continuar? (s/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[SsYy]$ ]]; then
    echo "❌ Restauração cancelada"
    exit 0
fi

echo "📦 Importando dados do JSON..."

# Executa o comando Go para importar
cd "$(dirname "$0")/.."
export DATABASE_URL="$NEW_DATABASE_URL"
go run cmd/restore-db/main.go "$BACKUP_FOLDER"

if [ $? -eq 0 ]; then
    echo "✅ Restauração concluída com sucesso!"
    echo ""
    echo "🔍 Próximos passos:"
    echo "   1. Execute o script de validação: ./scripts/validate-backup.sh"
    echo "   2. Atualize as variáveis de ambiente da sua aplicação"
    echo "   3. Teste as funcionalidades principais"
else
    echo "❌ Erro ao restaurar backup JSON"
    exit 1
fi

