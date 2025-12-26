package pagination

import (
	"math"
	"net/http"
	"strconv"
)

// Constantes de configuração
const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100 // Previne requests muito grandes
	MinLimit     = 1
)

// Params representa os parâmetros de paginação
type Params struct {
	Page  int
	Limit int
}

// Metadata contém informações sobre a paginação
type Metadata struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// Response é a estrutura padrão para respostas paginadas
type Response struct {
	Data       interface{} `json:"data"`
	Pagination Metadata    `json:"pagination"`
}

// ExtractParams extrai e valida os parâmetros de paginação da requisição
func ExtractParams(r *http.Request) Params {
	page := extractInt(r.URL.Query().Get("page"), DefaultPage)
	limit := extractInt(r.URL.Query().Get("limit"), DefaultLimit)

	// Validar page (mínimo 1)
	if page < 1 {
		page = DefaultPage
	}

	// Validar limit (entre MinLimit e MaxLimit)
	if limit < MinLimit {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	return Params{
		Page:  page,
		Limit: limit,
	}
}

// extractInt converte string para int, retornando defaultValue em caso de erro
func extractInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}

	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}

	return val
}

// CalculateOffset calcula o offset para a query baseado nos parâmetros
func CalculateOffset(params Params) int {
	return (params.Page - 1) * params.Limit
}

// BuildMetadata constrói os metadados de paginação
func BuildMetadata(params Params, total int64) Metadata {
	totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))
	
	// Evitar divisão por zero
	if params.Limit == 0 {
		totalPages = 0
	}

	return Metadata{
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    params.Page < totalPages,
		HasPrev:    params.Page > 1,
	}
}

// BuildResponse constrói a resposta paginada completa
func BuildResponse(data interface{}, params Params, total int64) Response {
	return Response{
		Data:       data,
		Pagination: BuildMetadata(params, total),
	}
}

