package testdb

import (
	"errors"

	"github.com/davidsonmarra/receitas-app/pkg/foodai"
)

// MockGeminiClient cliente mock para testes
type MockGeminiClient struct {
	ShouldReturnError bool
	MockResult        *foodai.FoodDetectionResult
}

// AnalyzeFood implementação mock
func (m *MockGeminiClient) AnalyzeFood(imageData []byte) (*foodai.FoodDetectionResult, error) {
	if m.ShouldReturnError {
		return nil, errors.New("mock gemini error")
	}

	if m.MockResult != nil {
		return m.MockResult, nil
	}

	// Resultado padrão para testes
	return &foodai.FoodDetectionResult{
		Foods: []foodai.DetectedFood{
			{Name: "arroz branco", Confidence: 0.95, Quantity: 150},
			{Name: "feijão preto", Confidence: 0.88, Quantity: 100},
		},
	}, nil
}

