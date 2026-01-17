#!/bin/bash
# Script de Validação do Banco de Dados
# Verifica a integridade dos dados após restauração
# Uso: ./scripts/validate-backup.sh

set -e  # Para em caso de erro

echo "🔍 Iniciando validação do banco de dados..."
echo ""

# Verifica se NEW_DATABASE_URL está definida
if [ -z "$NEW_DATABASE_URL" ]; then
    echo "⚠️  NEW_DATABASE_URL não está definida, usando DATABASE_URL"
    if [ -z "$DATABASE_URL" ]; then
        echo "❌ Erro: Nenhuma DATABASE_URL definida"
        exit 1
    fi
    DB_URL="$DATABASE_URL"
else
    DB_URL="$NEW_DATABASE_URL"
fi

echo "📊 Contagem de registros por tabela:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Função para contar registros
count_records() {
    table=$1
    count=$(psql "$DB_URL" -t -c "SELECT COUNT(*) FROM $table;")
    printf "  %-20s %s registros\n" "$table:" "$count"
}

# Contar registros em cada tabela
count_records "users"
count_records "ingredients"
count_records "recipes"
count_records "recipe_ingredients"
count_records "ratings"

echo ""
echo "🔗 Verificando integridade referencial:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Verificar foreign keys órfãs em recipes
orphan_recipes=$(psql "$DB_URL" -t -c "SELECT COUNT(*) FROM recipes WHERE user_id IS NOT NULL AND user_id NOT IN (SELECT id FROM users);")
if [ "$orphan_recipes" -gt 0 ]; then
    echo "  ❌ Recipes com user_id inválido: $orphan_recipes"
else
    echo "  ✅ Recipes → Users: OK"
fi

# Verificar foreign keys órfãs em recipe_ingredients (recipe_id)
orphan_ri_recipe=$(psql "$DB_URL" -t -c "SELECT COUNT(*) FROM recipe_ingredients WHERE recipe_id NOT IN (SELECT id FROM recipes);")
if [ "$orphan_ri_recipe" -gt 0 ]; then
    echo "  ❌ Recipe Ingredients com recipe_id inválido: $orphan_ri_recipe"
else
    echo "  ✅ Recipe Ingredients → Recipes: OK"
fi

# Verificar foreign keys órfãs em recipe_ingredients (ingredient_id)
orphan_ri_ingredient=$(psql "$DB_URL" -t -c "SELECT COUNT(*) FROM recipe_ingredients WHERE ingredient_id NOT IN (SELECT id FROM ingredients);")
if [ "$orphan_ri_ingredient" -gt 0 ]; then
    echo "  ❌ Recipe Ingredients com ingredient_id inválido: $orphan_ri_ingredient"
else
    echo "  ✅ Recipe Ingredients → Ingredients: OK"
fi

# Verificar foreign keys órfãs em ratings (recipe_id)
orphan_ratings_recipe=$(psql "$DB_URL" -t -c "SELECT COUNT(*) FROM ratings WHERE recipe_id NOT IN (SELECT id FROM recipes);")
if [ "$orphan_ratings_recipe" -gt 0 ]; then
    echo "  ❌ Ratings com recipe_id inválido: $orphan_ratings_recipe"
else
    echo "  ✅ Ratings → Recipes: OK"
fi

# Verificar foreign keys órfãs em ratings (user_id)
orphan_ratings_user=$(psql "$DB_URL" -t -c "SELECT COUNT(*) FROM ratings WHERE user_id NOT IN (SELECT id FROM users);")
if [ "$orphan_ratings_user" -gt 0 ]; then
    echo "  ❌ Ratings com user_id inválido: $orphan_ratings_user"
else
    echo "  ✅ Ratings → Users: OK"
fi

echo ""
echo "🔢 Verificando sequences (auto-increment):"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

check_sequence() {
    table=$1
    sequence=$2
    
    max_id=$(psql "$DB_URL" -t -c "SELECT COALESCE(MAX(id), 0) FROM $table;")
    seq_val=$(psql "$DB_URL" -t -c "SELECT last_value FROM $sequence;")
    
    if [ "$seq_val" -ge "$max_id" ]; then
        echo "  ✅ $sequence: $seq_val (max ID: $max_id)"
    else
        echo "  ⚠️  $sequence: $seq_val (max ID: $max_id) - Precisa atualizar!"
    fi
}

check_sequence "users" "users_id_seq"
check_sequence "ingredients" "ingredients_id_seq"
check_sequence "recipes" "recipes_id_seq"
check_sequence "recipe_ingredients" "recipe_ingredients_id_seq"
check_sequence "ratings" "ratings_id_seq"

echo ""
echo "📋 Verificando índices:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

indices_count=$(psql "$DB_URL" -t -c "SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public';")
echo "  📌 Índices criados: $indices_count"

# Listar alguns índices importantes
echo ""
echo "  Principais índices:"
psql "$DB_URL" -c "SELECT tablename, indexname FROM pg_indexes WHERE schemaname = 'public' ORDER BY tablename, indexname;" | head -n 20

echo ""
echo "✅ Validação concluída!"
echo ""
echo "💡 Próximos passos:"
echo "   1. Verifique se todos os registros estão corretos"
echo "   2. Teste o login de usuários"
echo "   3. Verifique se as imagens estão carregando"
echo "   4. Atualize as variáveis de ambiente da aplicação"

