package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidsonmarra/receitas-app/pkg/pagination"
)

// Testes de unidade do pacote pagination

func TestExtractParams_DefaultValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/recipes", nil)
	params := pagination.ExtractParams(req)

	if params.Page != pagination.DefaultPage {
		t.Errorf("Expected page %d, got %d", pagination.DefaultPage, params.Page)
	}

	if params.Limit != pagination.DefaultLimit {
		t.Errorf("Expected limit %d, got %d", pagination.DefaultLimit, params.Limit)
	}
}

func TestExtractParams_ValidValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/recipes?page=2&limit=50", nil)
	params := pagination.ExtractParams(req)

	if params.Page != 2 {
		t.Errorf("Expected page 2, got %d", params.Page)
	}

	if params.Limit != 50 {
		t.Errorf("Expected limit 50, got %d", params.Limit)
	}
}

func TestExtractParams_NegativeValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/recipes?page=-1&limit=-10", nil)
	params := pagination.ExtractParams(req)

	if params.Page != pagination.DefaultPage {
		t.Errorf("Negative page should default to %d, got %d", pagination.DefaultPage, params.Page)
	}

	if params.Limit != pagination.DefaultLimit {
		t.Errorf("Negative limit should default to %d, got %d", pagination.DefaultLimit, params.Limit)
	}
}

func TestExtractParams_ZeroValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/recipes?page=0&limit=0", nil)
	params := pagination.ExtractParams(req)

	if params.Page != pagination.DefaultPage {
		t.Errorf("Zero page should default to %d, got %d", pagination.DefaultPage, params.Page)
	}

	if params.Limit != pagination.DefaultLimit {
		t.Errorf("Zero limit should default to %d, got %d", pagination.DefaultLimit, params.Limit)
	}
}

func TestExtractParams_ExceedsMaxLimit(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/recipes?page=1&limit=500", nil)
	params := pagination.ExtractParams(req)

	if params.Limit != pagination.MaxLimit {
		t.Errorf("Limit should be capped at %d, got %d", pagination.MaxLimit, params.Limit)
	}
}

func TestExtractParams_InvalidValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/recipes?page=abc&limit=xyz", nil)
	params := pagination.ExtractParams(req)

	if params.Page != pagination.DefaultPage {
		t.Errorf("Invalid page should default to %d, got %d", pagination.DefaultPage, params.Page)
	}

	if params.Limit != pagination.DefaultLimit {
		t.Errorf("Invalid limit should default to %d, got %d", pagination.DefaultLimit, params.Limit)
	}
}

func TestCalculateOffset(t *testing.T) {
	tests := []struct {
		page     int
		limit    int
		expected int
	}{
		{1, 20, 0},     // Primeira página
		{2, 20, 20},    // Segunda página
		{3, 20, 40},    // Terceira página
		{1, 10, 0},     // Primeira página, limite menor
		{5, 10, 40},    // Quinta página, limite menor
		{10, 100, 900}, // Página alta, limite alto
	}

	for _, tt := range tests {
		params := pagination.Params{Page: tt.page, Limit: tt.limit}
		offset := pagination.CalculateOffset(params)

		if offset != tt.expected {
			t.Errorf("For page=%d, limit=%d, expected offset %d, got %d",
				tt.page, tt.limit, tt.expected, offset)
		}
	}
}

func TestBuildMetadata(t *testing.T) {
	tests := []struct {
		name           string
		page           int
		limit          int
		total          int64
		expectedPages  int
		expectedHasNext bool
		expectedHasPrev bool
	}{
		{"first page with data", 1, 20, 100, 5, true, false},
		{"middle page", 3, 20, 100, 5, true, true},
		{"last page", 5, 20, 100, 5, false, true},
		{"single page", 1, 20, 10, 1, false, false},
		{"empty result", 1, 20, 0, 0, false, false},
		{"partial last page", 3, 20, 45, 3, false, true},
		{"exact pages", 2, 25, 100, 4, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := pagination.Params{Page: tt.page, Limit: tt.limit}
			metadata := pagination.BuildMetadata(params, tt.total)

			if metadata.Page != tt.page {
				t.Errorf("Expected page %d, got %d", tt.page, metadata.Page)
			}

			if metadata.Limit != tt.limit {
				t.Errorf("Expected limit %d, got %d", tt.limit, metadata.Limit)
			}

			if metadata.Total != tt.total {
				t.Errorf("Expected total %d, got %d", tt.total, metadata.Total)
			}

			if metadata.TotalPages != tt.expectedPages {
				t.Errorf("Expected total_pages %d, got %d", tt.expectedPages, metadata.TotalPages)
			}

			if metadata.HasNext != tt.expectedHasNext {
				t.Errorf("Expected has_next %v, got %v", tt.expectedHasNext, metadata.HasNext)
			}

			if metadata.HasPrev != tt.expectedHasPrev {
				t.Errorf("Expected has_prev %v, got %v", tt.expectedHasPrev, metadata.HasPrev)
			}
		})
	}
}

func TestBuildResponse(t *testing.T) {
	data := []string{"item1", "item2", "item3"}
	params := pagination.Params{Page: 1, Limit: 20}
	total := int64(100)

	response := pagination.BuildResponse(data, params, total)

	if response.Data == nil {
		t.Error("Expected data to be set")
	}

	if response.Pagination.Page != 1 {
		t.Errorf("Expected page 1, got %d", response.Pagination.Page)
	}

	if response.Pagination.Total != 100 {
		t.Errorf("Expected total 100, got %d", response.Pagination.Total)
	}

	if response.Pagination.TotalPages != 5 {
		t.Errorf("Expected total_pages 5, got %d", response.Pagination.TotalPages)
	}
}

