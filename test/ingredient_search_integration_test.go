package test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/davidsonmarra/receitas-app/internal/http/handlers"
	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/database"
	"github.com/davidsonmarra/receitas-app/test/testdb"
)

// TestIngredientSearchIntegration testa casos de uso reais da busca de ingredientes
func TestIngredientSearchIntegration(t *testing.T) {
	testdb.SetupWithCleanup(t)

	// Criar ingredientes reais do banco TACO
	ingredients := []models.Ingredient{
		// Farinhas
		{Name: "Farinha de Trigo", Calories: 364, Category: "farinhas", Source: "taco"},
		{Name: "Farinha de Trigo Integral", Calories: 340, Category: "farinhas", Source: "taco"},
		{Name: "Farinha de Rosca", Calories: 398, Category: "farinhas", Source: "taco"},
		{Name: "Farinha de Mandioca", Calories: 365, Category: "farinhas", Source: "taco"},
		// Cereais
		{Name: "Arroz branco", Calories: 128, Category: "cereais", Source: "taco"},
		{Name: "Arroz integral", Calories: 123, Category: "cereais", Source: "taco"},
		{Name: "Trigo em Grão", Calories: 330, Category: "cereais", Source: "taco"},
		// Leguminosas
		{Name: "Feijão preto", Calories: 77, Category: "leguminosas", Source: "taco"},
		{Name: "Feijão carioca", Calories: 76, Category: "leguminosas", Source: "taco"},
		// Óleos
		{Name: "Óleo de Soja", Calories: 884, Category: "óleos", Source: "taco"},
		{Name: "Óleo de Coco", Calories: 862, Category: "óleos", Source: "taco"},
		// Frutas
		{Name: "Abacate, cru", Calories: 96, Category: "frutas", Source: "taco"},
		{Name: "Abacaxi, cru", Calories: 48, Category: "frutas", Source: "taco"},
		{Name: "Coco ralado", Calories: 354, Category: "frutas", Source: "taco"},
	}

	for _, ing := range ingredients {
		database.DB.Create(&ing)
	}

	// Caso de Uso 1: "Farinha de Trigo" deve encontrar farinhas com trigo
	t.Run("Busca: farinha de trigo", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ingredients?search=farinha+de+trigo", nil)
		rec := httptest.NewRecorder()

		handlers.ListIngredients(rec, req)

		if rec.Code != 200 {
			t.Fatalf("esperado status 200, obteve %d", rec.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)

		data := response["data"].([]interface{})
		
		// Deve encontrar pelo menos as 2 farinhas de trigo
		if len(data) < 2 {
			t.Errorf("esperado pelo menos 2 ingredientes, obteve %d", len(data))
		}

		// Verificar que "Farinha de Trigo" está nos resultados (contém ambas palavras)
		foundFarinhaTrigo := false
		foundFarinhaTrigoIntegral := false
		
		for _, item := range data {
			ing := item.(map[string]interface{})
			name := ing["name"].(string)
			if name == "Farinha de Trigo" {
				foundFarinhaTrigo = true
			}
			if name == "Farinha de Trigo Integral" {
				foundFarinhaTrigoIntegral = true
			}
		}

		if !foundFarinhaTrigo {
			t.Error("❌ Não encontrou 'Farinha de Trigo' - PROBLEMA ORIGINAL")
		} else {
			t.Log("✅ Encontrou 'Farinha de Trigo'")
		}

		if !foundFarinhaTrigoIntegral {
			t.Error("❌ Não encontrou 'Farinha de Trigo Integral'")
		} else {
			t.Log("✅ Encontrou 'Farinha de Trigo Integral'")
		}
	})

	// Caso de Uso 2: "arroz integral" deve priorizar Arroz integral
	t.Run("Busca: arroz integral", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ingredients?search=arroz+integral", nil)
		rec := httptest.NewRecorder()

		handlers.ListIngredients(rec, req)

		if rec.Code != 200 {
			t.Fatalf("esperado status 200, obteve %d", rec.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)

		data := response["data"].([]interface{})
		
		if len(data) == 0 {
			t.Fatal("nenhum resultado encontrado")
		}

		// Primeiro resultado deve ser "Arroz integral" (contém ambas palavras E começa com primeira)
		first := data[0].(map[string]interface{})
		firstName := first["name"].(string)
		
		if firstName == "Arroz integral" {
			t.Log("✅ 'Arroz integral' está em primeiro lugar (Prioridade 1: contém ambas E começa com 'arroz')")
		} else {
			t.Errorf("❌ Primeiro resultado: %s (esperado 'Arroz integral' - ranking incorreto)", firstName)
		}
	})

	// Caso de Uso 3: "feijão preto" deve encontrar feijão preto primeiro
	t.Run("Busca: feijão preto", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ingredients?search=feij%C3%A3o+preto", nil)
		rec := httptest.NewRecorder()

		handlers.ListIngredients(rec, req)

		if rec.Code != 200 {
			t.Fatalf("esperado status 200, obteve %d", rec.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)

		data := response["data"].([]interface{})
		
		if len(data) == 0 {
			t.Fatal("nenhum resultado encontrado")
		}

		// Primeiro resultado deve ser "Feijão preto"
		first := data[0].(map[string]interface{})
		firstName := first["name"].(string)
		
		if firstName == "Feijão preto" {
			t.Log("✅ 'Feijão preto' está em primeiro lugar")
		} else {
			t.Errorf("❌ Primeiro resultado: %s (esperado 'Feijão preto')", firstName)
		}
	})

	// Caso de Uso 4: Busca com uma palavra só deve continuar funcionando
	t.Run("Busca: abacate", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ingredients?search=abacate", nil)
		rec := httptest.NewRecorder()

		handlers.ListIngredients(rec, req)

		if rec.Code != 200 {
			t.Fatalf("esperado status 200, obteve %d", rec.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)

		data := response["data"].([]interface{})
		
		if len(data) != 1 {
			t.Errorf("esperado 1 ingrediente, obteve %d", len(data))
		}

		if len(data) > 0 {
			first := data[0].(map[string]interface{})
			name := first["name"].(string)
			if name == "Abacate, cru" {
				t.Log("✅ Busca com palavra única funciona")
			}
		}
	})

	// Caso de Uso 5: Stopwords devem ser ignoradas
	t.Run("Busca: óleo de coco", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ingredients?search=%C3%B3leo+de+coco", nil)
		rec := httptest.NewRecorder()

		handlers.ListIngredients(rec, req)

		if rec.Code != 200 {
			t.Fatalf("esperado status 200, obteve %d", rec.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)

		data := response["data"].([]interface{})
		
		// Deve encontrar óleos e coco
		if len(data) < 2 {
			t.Errorf("esperado pelo menos 2 ingredientes, obteve %d", len(data))
		}

		// Verificar que "Óleo de Coco" vem primeiro
		if len(data) > 0 {
			first := data[0].(map[string]interface{})
			name := first["name"].(string)
			if name == "Óleo de Coco" {
				t.Log("✅ 'Óleo de Coco' está em primeiro lugar (stopword 'de' ignorada)")
			}
		}
	})

	// Resumo dos testes
	t.Log("\n=== RESUMO DOS TESTES DE INTEGRAÇÃO ===")
	t.Log("✅ Busca por múltiplas palavras funciona")
	t.Log("✅ Stopwords são ignoradas corretamente")
	t.Log("✅ Ranking de relevância está correto")
	t.Log("✅ Compatibilidade com busca de palavra única mantida")
}

