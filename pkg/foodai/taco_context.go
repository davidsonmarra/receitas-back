package foodai

import (
	"strings"
	"sync"

	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/database"
	"github.com/davidsonmarra/receitas-app/pkg/log"
)

var (
	// Cache de ingredientes por categoria
	ingredientCache      map[string][]string
	ingredientCacheMutex sync.RWMutex
	ingredientCacheInit  sync.Once
)

// GetTACOContext retorna contexto dos ingredientes TACO para o prompt
func GetTACOContext() string {
	ingredientCacheInit.Do(loadIngredientCache)

	ingredientCacheMutex.RLock()
	defer ingredientCacheMutex.RUnlock()

	var builder strings.Builder
	builder.WriteString("BANCO DE DADOS TACO - Use estes nomes EXATOS quando identificar alimentos:\n\n")

	// Categorias em ordem de prioridade (mais usadas primeiro)
	priorityCategories := []string{
		"cereais",
		"leguminosas",
		"carnes",
		"vegetais",
		"frutas",
		"laticínios",
		"ovos",
		"peixes e frutos do mar",
		"óleos e gorduras",
		"açúcares e doces",
	}

	for _, category := range priorityCategories {
		if ingredients, ok := ingredientCache[category]; ok && len(ingredients) > 0 {
			builder.WriteString("📦 ")
			builder.WriteString(strings.ToUpper(category))
			builder.WriteString(":\n")

			// Limitar a 12 ingredientes mais comuns por categoria (reduz tamanho do prompt)
			limit := 12
			if len(ingredients) < limit {
				limit = len(ingredients)
			}

			for i := 0; i < limit; i++ {
				builder.WriteString("  - ")
				builder.WriteString(ingredients[i])
				builder.WriteString("\n")
			}
			builder.WriteString("\n")
		}
	}

	builder.WriteString("IMPORTANTE: Se identificar um alimento que corresponde a algum nome acima, use EXATAMENTE esse nome.\n")
	builder.WriteString("Se não encontrar correspondência, use o nome em português mais próximo.\n")

	return builder.String()
}

// loadIngredientCache carrega os ingredientes do banco e agrupa por categoria
func loadIngredientCache() {
	log.Info("carregando cache de ingredientes TACO")

	ingredientCache = make(map[string][]string)

	var ingredients []models.Ingredient

	// Buscar todos os ingredientes com dados nutricionais completos
	if err := database.DB.
		Select("name, category").
		Where("calories > 0").
		Order("category ASC, name ASC").
		Find(&ingredients).Error; err != nil {
		log.Error("erro ao carregar ingredientes para cache", "error", err)
		return
	}

	// Agrupar por categoria
	for _, ing := range ingredients {
		category := strings.ToLower(strings.TrimSpace(ing.Category))
		if category != "" {
			ingredientCache[category] = append(ingredientCache[category], ing.Name)
		}
	}

	// Log estatísticas
	totalIngredients := 0
	for category, items := range ingredientCache {
		totalIngredients += len(items)
		log.Debug("cache TACO", "category", category, "count", len(items))
	}

	log.Info("cache de ingredientes TACO carregado",
		"categories", len(ingredientCache),
		"total_ingredients", totalIngredients)
}

// GetTopIngredientsByCategory retorna os N ingredientes mais relevantes de uma categoria
func GetTopIngredientsByCategory(category string, limit int) []string {
	ingredientCacheInit.Do(loadIngredientCache)

	ingredientCacheMutex.RLock()
	defer ingredientCacheMutex.RUnlock()

	category = strings.ToLower(strings.TrimSpace(category))

	if ingredients, ok := ingredientCache[category]; ok {
		if len(ingredients) <= limit {
			return ingredients
		}
		return ingredients[:limit]
	}

	return []string{}
}

// RefreshCache atualiza o cache (útil se ingredientes forem adicionados)
func RefreshCache() {
	log.Info("atualizando cache de ingredientes TACO")

	ingredientCacheMutex.Lock()
	defer ingredientCacheMutex.Unlock()

	ingredientCache = nil
	ingredientCacheInit = sync.Once{}

	// Recarregar
	loadIngredientCache()
}

