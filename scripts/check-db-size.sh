#!/bin/bash
# Script para verificar o tamanho do banco de dados
# Útil antes de fazer backup para garantir que cabe no limite do Railway free tier (512MB)
# Uso: ./scripts/check-db-size.sh

set -e

echo "🔍 Verificando tamanho do banco de dados..."
echo ""

# Verifica se DATABASE_URL está definida
if [ -z "$DATABASE_URL" ]; then
    echo "❌ Erro: DATABASE_URL não está definida"
    echo "💡 Defina a variável: export DATABASE_URL='sua_connection_string'"
    exit 1
fi

# Tamanho total do banco
echo "📊 Tamanho do Banco de Dados:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
psql "$DATABASE_URL" -c "SELECT pg_size_pretty(pg_database_size(current_database())) as size;"

echo ""
echo "📋 Tamanho por Tabela:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
psql "$DATABASE_URL" -c "
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size,
    pg_total_relation_size(schemaname||'.'||tablename) AS size_bytes
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY size_bytes DESC;
"

echo ""
echo "💾 Limite Railway Free Tier: 512 MB"
echo ""

# Calcular tamanho total em bytes
total_bytes=$(psql "$DATABASE_URL" -t -c "SELECT pg_database_size(current_database());")
total_bytes=$(echo "$total_bytes" | tr -d ' ')

# Limite do Railway em bytes (512 MB)
limit_bytes=536870912

if [ "$total_bytes" -lt "$limit_bytes" ]; then
    usage_percent=$((total_bytes * 100 / limit_bytes))
    echo "✅ Banco está dentro do limite ($usage_percent% usado)"
else
    echo "⚠️  AVISO: Banco excede o limite do Railway free tier!"
    echo "   Considere limpar dados antigos ou fazer upgrade do plano"
fi

