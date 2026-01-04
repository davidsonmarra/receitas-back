package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/database"
	"github.com/davidsonmarra/receitas-app/pkg/log"
	"github.com/davidsonmarra/receitas-app/pkg/pagination"
	"github.com/davidsonmarra/receitas-app/pkg/response"
	"github.com/davidsonmarra/receitas-app/pkg/validation"
)

// splitSearchTerms divide o termo de busca em palavras válidas, removendo stopwords
func splitSearchTerms(search string) []string {
	// Stopwords comuns em português
	stopwords := map[string]bool{
		"de": true, "da": true, "do": true, "das": true, "dos": true,
		"e": true, "ou": true, "com": true, "em": true, "a": true,
		"o": true, "as": true, "os": true, "para": true,
	}

	// Normalizar: lowercase e trim
	search = strings.TrimSpace(strings.ToLower(search))

	// Dividir em palavras
	words := strings.Fields(search)

	// Filtrar palavras válidas (>= 3 caracteres e não stopwords)
	var validWords []string
	for _, word := range words {
		word = strings.TrimSpace(word)
		// Manter palavras com 3+ caracteres que não sejam stopwords
		if len(word) >= 3 && !stopwords[word] {
			validWords = append(validWords, word)
		}
	}

	return validWords
}

// ListIngredients lista todos ingredientes com filtros e paginação
func ListIngredients(w http.ResponseWriter, r *http.Request) {
	params := pagination.ExtractParams(r)

	query := database.DB.Model(&models.Ingredient{})

	// Filtro por nome e categoria com ranking de relevância
	if search := r.URL.Query().Get("search"); search != "" {
		// Normalizar busca: lowercase e trim
		search = strings.TrimSpace(strings.ToLower(search))

		// Dividir em palavras válidas
		searchWords := splitSearchTerms(search)

		// Se não houver palavras válidas, usar busca original (termo completo)
		if len(searchWords) == 0 {
			searchPattern := "%" + search + "%"
			query = query.Where(
				"LOWER(name) LIKE ? OR LOWER(category) LIKE ?",
				searchPattern, searchPattern,
			)
		} else {
			// Construir WHERE com múltiplas palavras (OR)
			var conditions []string
			var args []interface{}

			for _, word := range searchWords {
				pattern := "%" + word + "%"
				conditions = append(conditions, "(LOWER(name) LIKE ? OR LOWER(category) LIKE ?)")
				args = append(args, pattern, pattern)
			}

			whereClause := strings.Join(conditions, " OR ")
			query = query.Where(whereClause, args...)

			// Ordenar por relevância: quanto mais palavras no nome, melhor
			var orderCases []string
			var orderArgs []interface{}

			// Prioridade 1: Nome contém TODAS as palavras (maior relevância)
			var allWordsConditions []string
			for _, word := range searchWords {
				allWordsConditions = append(allWordsConditions, "LOWER(name) LIKE ?")
				orderArgs = append(orderArgs, "%"+word+"%")
			}
			orderCases = append(orderCases, "WHEN "+strings.Join(allWordsConditions, " AND ")+" THEN 1")

			// Prioridade 2: Nome começa com a primeira palavra
			orderCases = append(orderCases, "WHEN LOWER(name) LIKE ? THEN 2")
			orderArgs = append(orderArgs, searchWords[0]+"%")

			// Prioridade 3: Nome contém a primeira palavra
			orderCases = append(orderCases, "WHEN LOWER(name) LIKE ? THEN 3")
			orderArgs = append(orderArgs, "%"+searchWords[0]+"%")

			// Prioridade 4: Categoria contém alguma palavra
			orderCases = append(orderCases, "WHEN LOWER(category) LIKE ? THEN 4")
			orderArgs = append(orderArgs, "%"+searchWords[0]+"%")

			orderCases = append(orderCases, "ELSE 5 END")
			orderSQL := "CASE " + strings.Join(orderCases, " ") + " "

			query = query.Order(database.DB.Raw(orderSQL, orderArgs...))
		}
	} else {
		// Sem busca, ordenar alfabeticamente
		query = query.Order("name ASC")
	}

	// Filtro adicional por categoria (AND com search)
	if category := r.URL.Query().Get("category"); category != "" {
		query = query.Where("category = ?", strings.ToLower(category))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to count ingredients", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to count ingredients")
		return
	}

	var ingredients []models.Ingredient
	offset := pagination.CalculateOffset(params)

	if err := query.Limit(params.Limit).Offset(offset).
		Find(&ingredients).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to list ingredients", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to list ingredients")
		return
	}

	log.InfoCtx(r.Context(), "ingredients listed", "total", total, "returned", len(ingredients), "search", r.URL.Query().Get("search"), "search_words", len(splitSearchTerms(r.URL.Query().Get("search"))))
	paginatedResponse := pagination.BuildResponse(ingredients, params, total)
	response.JSON(w, http.StatusOK, paginatedResponse)
}

// GetIngredient retorna um ingrediente por ID
func GetIngredient(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var ingredient models.Ingredient
	if err := database.DB.First(&ingredient, id).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Ingredient not found")
		return
	}

	response.JSON(w, http.StatusOK, ingredient)
}

// CreateIngredient cria um novo ingrediente (admin only)
func CreateIngredient(w http.ResponseWriter, r *http.Request) {
	var ingredient models.Ingredient

	if err := json.NewDecoder(r.Body).Decode(&ingredient); err != nil {
		response.ValidationError(w, "Formato de dados inválido.")
		return
	}

	if errs := validation.ValidateStruct(ingredient); errs != nil {
		message := validation.FormatErrors(errs)
		response.ValidationError(w, message)
		return
	}

	// Normalizar categoria para lowercase
	ingredient.Category = strings.ToLower(ingredient.Category)
	if ingredient.Source == "" {
		ingredient.Source = "manual"
	}

	if err := database.DB.Create(&ingredient).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to create ingredient", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to create ingredient")
		return
	}

	log.InfoCtx(r.Context(), "ingredient created", "id", ingredient.ID, "name", ingredient.Name)
	response.JSON(w, http.StatusCreated, ingredient)
}

// UpdateIngredient atualiza um ingrediente (admin only)
func UpdateIngredient(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var ingredient models.Ingredient
	if err := database.DB.First(&ingredient, id).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Ingredient not found")
		return
	}

	var updateData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		response.ValidationError(w, "Formato de dados inválido.")
		return
	}

	// Normalizar categoria se presente
	if category, ok := updateData["category"].(string); ok {
		updateData["category"] = strings.ToLower(category)
	}

	if err := database.DB.Model(&ingredient).Updates(updateData).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to update ingredient", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to update ingredient")
		return
	}

	log.InfoCtx(r.Context(), "ingredient updated", "id", ingredient.ID)
	response.JSON(w, http.StatusOK, ingredient)
}

// DeleteIngredient remove um ingrediente (admin only)
func DeleteIngredient(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Verificar se ingrediente está em uso em alguma receita
	var count int64
	database.DB.Model(&models.RecipeIngredient{}).Where("ingredient_id = ?", id).Count(&count)

	if count > 0 {
		response.ValidationError(w, "Não é possível deletar ingrediente em uso em receitas.")
		return
	}

	if err := database.DB.Delete(&models.Ingredient{}, id).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to delete ingredient", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to delete ingredient")
		return
	}

	log.InfoCtx(r.Context(), "ingredient deleted", "id", id)
	response.JSON(w, http.StatusOK, map[string]string{"message": "Ingredient deleted"})
}

// GetCategories retorna lista de categorias disponíveis
func GetCategories(w http.ResponseWriter, r *http.Request) {
	var categories []string

	database.DB.Model(&models.Ingredient{}).
		Distinct("category").
		Where("category != ''").
		Order("category ASC").
		Pluck("category", &categories)

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"categories": categories,
	})
}
