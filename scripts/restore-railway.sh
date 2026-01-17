#!/bin/bash
# Script de Restauração via psql
# Restaura um dump SQL para o banco PostgreSQL
# Uso: ./scripts/restore-railway.sh <arquivo_backup.sql>

set -e  # Para em caso de erro

echo "🔄 Iniciando restauração do banco de dados..."

# Verifica se o arquivo de backup foi fornecido
if [ -z "$1" ]; then
    echo "❌ Erro: Arquivo de backup não fornecido"
    echo "💡 Uso: ./scripts/restore-railway.sh <arquivo_backup.sql>"
    echo ""
    echo "Exemplos de arquivos disponíveis:"
    ls -lh backups/*.sql 2>/dev/null || echo "  Nenhum backup SQL encontrado"
    exit 1
fi

BACKUP_FILE="$1"

# Verifica se o arquivo existe
if [ ! -f "$BACKUP_FILE" ]; then
    echo "❌ Erro: Arquivo não encontrado: $BACKUP_FILE"
    exit 1
fi

# Verifica se NEW_DATABASE_URL está definida
if [ -z "$NEW_DATABASE_URL" ]; then
    echo "❌ Erro: NEW_DATABASE_URL não está definida"
    echo "💡 Defina a variável: export NEW_DATABASE_URL='nova_connection_string'"
    exit 1
fi

echo "📄 Arquivo de backup: $BACKUP_FILE"
echo "📊 Tamanho: $(du -h "$BACKUP_FILE" | cut -f1)"
echo ""

# Pergunta de confirmação
read -p "⚠️  Isso irá SOBRESCREVER todos os dados no novo banco. Continuar? (s/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[SsYy]$ ]]; then
    echo "❌ Restauração cancelada"
    exit 0
fi

echo "📦 Restaurando dump no novo banco..."
psql "$NEW_DATABASE_URL" < "$BACKUP_FILE"

if [ $? -eq 0 ]; then
    echo "✅ Restauração concluída com sucesso!"
    echo ""
    echo "🔍 Próximos passos:"
    echo "   1. Execute o script de validação: ./scripts/validate-backup.sh"
    echo "   2. Atualize as variáveis de ambiente da sua aplicação"
    echo "   3. Teste as funcionalidades principais"
else
    echo "❌ Erro ao restaurar backup"
    exit 1
fi

