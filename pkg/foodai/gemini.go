package foodai

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/davidsonmarra/receitas-app/pkg/log"
)

// GeminiClient cliente para Google Gemini Vision API
type GeminiClient struct {
	apiKey  string
	baseURL string
}

// GeminiRequest estrutura de request para Gemini API
type GeminiRequest struct {
	Contents         []Content         `json:"contents"`
	GenerationConfig *GenerationConfig `json:"generationConfig,omitempty"`
}

// GenerationConfig configurações de geração do Gemini
type GenerationConfig struct {
	Temperature     float64 `json:"temperature"`     // 0.0 = mais determinístico
	TopP            float64 `json:"topP"`            // Limita diversidade
	TopK            int     `json:"topK"`            // Reduz opções aleatórias
	MaxOutputTokens int     `json:"maxOutputTokens"` // Limite de tokens de saída
}

// Content conteúdo da mensagem
type Content struct {
	Parts []Part `json:"parts"`
}

// Part parte da mensagem (texto ou imagem)
type Part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *InlineData `json:"inlineData,omitempty"`
}

// InlineData dados inline (imagem em base64)
type InlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"` // base64
}

// GeminiResponse resposta da API do Gemini
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// FoodDetectionResult resultado da detecção de alimentos
type FoodDetectionResult struct {
	Foods []DetectedFood `json:"foods"`
}

// DetectedFood alimento detectado na imagem
type DetectedFood struct {
	Name       string  `json:"name"`           // Nome em português
	Confidence float64 `json:"confidence"`     // 0-1
	Quantity   float64 `json:"quantity_grams"` // Estimativa em gramas
}

// NewGeminiClient cria um novo cliente Gemini
func NewGeminiClient() *GeminiClient {
	return &GeminiClient{
		apiKey:  os.Getenv("GEMINI_API_KEY"),
		baseURL: "https://generativelanguage.googleapis.com/v1beta/models/gemini-3-flash-preview:generateContent",
	}
}

// AnalyzeFood analisa uma imagem de comida e retorna os alimentos detectados
func (c *GeminiClient) AnalyzeFood(imageData []byte) (*FoodDetectionResult, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY não configurado")
	}

	// Converter imagem para base64
	imageBase64 := base64.StdEncoding.EncodeToString(imageData)

	// Obter contexto TACO
	tacoContext := GetTACOContext()

	// Criar prompt otimizado com contexto TACO
	prompt := tacoContext + `
═══════════════════════════════════════════════════════════════
TAREFA: Analise esta imagem e identifique TODOS os alimentos visíveis.

INSTRUÇÕES:
1. Identifique cada alimento separadamente
2. Use os nomes EXATOS da lista TACO acima quando corresponder
3. Estime quantidade em GRAMAS (considere: colher ~15g, prato ~26cm)
4. Confidence 0.0-1.0: claro(0.9-1.0), parcial(0.7-0.9), difícil(0.5-0.7)
5. Liste molhos/temperos separadamente se visíveis

FORMATO JSON (sem markdown):
{
  "foods": [
    {"name": "nome EXATO TACO ou português", "confidence": 0.95, "quantity_grams": 150}
  ]
}

EXEMPLOS:
✅ "Arroz, integral, cozido"
✅ "Feijão, preto, cozido"  
❌ "arroz" (genérico)
❌ "rice" (inglês)

RETORNE APENAS O JSON.`

	// Montar request com configurações otimizadas
	request := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
					{
						InlineData: &InlineData{
							MimeType: "image/jpeg",
							Data:     imageBase64,
						},
					},
				},
			},
		},
		GenerationConfig: &GenerationConfig{
			Temperature:     0.2,    // Baixa temperatura = mais preciso e consistente
			TopP:            0.8,    // Limita diversidade de tokens
			TopK:            40,     // Reduz opções aleatórias
			MaxOutputTokens: 8192,   // Espaço suficiente para respostas JSON complexas
		},
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar request: %w", err)
	}

	// Fazer request para Gemini API
	url := fmt.Sprintf("%s?key=%s", c.baseURL, c.apiKey)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestJSON))
	if err != nil {
		return nil, fmt.Errorf("erro ao chamar API do Gemini: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Error("erro na resposta do Gemini", "status", resp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	// Parse resposta do Gemini
	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("erro ao parsear resposta do Gemini: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("resposta vazia do Gemini")
	}

	// Extrair JSON do texto
	text := geminiResp.Candidates[0].Content.Parts[0].Text

	// Limpar possíveis markdown
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	// Parse resultado final
	var result FoodDetectionResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		log.Error("erro ao parsear resultado", "text", text, "error", err)
		return nil, fmt.Errorf("erro ao parsear resultado: %w - texto: %s", err, text)
	}

	log.Info("alimentos detectados pelo Gemini", "count", len(result.Foods))

	return &result, nil
}

