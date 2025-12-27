package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/database"
	"github.com/davidsonmarra/receitas-app/pkg/log"
)

func main() {
	// Inicializar logger e database
	logConfig := log.Config{Level: "info", Development: true}
	if err := log.Init(logConfig); err != nil {
		fmt.Printf("Erro ao inicializar logger: %v\n", err)
		os.Exit(1)
	}

	if err := database.Connect(); err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	fmt.Println("🌱 Iniciando seed de ingredientes da Tabela TACO...")

	// Ler arquivo CSV
	csvPath := "alimentos.csv"
	if len(os.Args) > 1 {
		csvPath = os.Args[1]
	}

	ingredients, err := parseTACOCSV(csvPath)
	if err != nil {
		log.Error("failed to parse TACO CSV", "error", err)
		fmt.Printf("❌ Erro ao ler CSV: %v\n", err)
		fmt.Println("\nUso: go run ./cmd/seed-ingredients [caminho_do_csv]")
		fmt.Println("Exemplo: go run ./cmd/seed-ingredients ~/Downloads/alimentos.csv")
		os.Exit(1)
	}

	fmt.Printf("📊 Total de alimentos encontrados: %d\n\n", len(ingredients))

	// Inserir ingredientes no banco
	inserted := 0
	skipped := 0

	for i, ing := range ingredients {
		if err := database.DB.Create(&ing).Error; err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				skipped++
			} else {
				log.Error("failed to create ingredient", "name", ing.Name, "error", err)
			}
		} else {
			inserted++
			if inserted%50 == 0 {
				fmt.Printf("  [%d/%d] ✅ Inseridos: %d, Ignorados: %d\n", i+1, len(ingredients), inserted, skipped)
			}
		}
	}

	fmt.Printf("\n✅ Seed completo!\n")
	fmt.Printf("   Inseridos: %d\n", inserted)
	fmt.Printf("   Ignorados (duplicados): %d\n", skipped)

	var count int64
	database.DB.Model(&models.Ingredient{}).Count(&count)
	fmt.Printf("   Total no banco: %d ingredientes\n", count)
}

// parseTACOCSV lê o arquivo CSV da Tabela TACO e retorna uma lista de ingredientes
func parseTACOCSV(filePath string) ([]models.Ingredient, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	var ingredients []models.Ingredient

	// Ler cabeçalho
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("erro ao ler cabeçalho: %w", err)
	}

	// Mapear índices das colunas
	colMap := make(map[string]int)
	for i, col := range header {
		colMap[col] = i
	}

	// Índices das colunas importantes
	categoriaCol := colMap["Categoria do alimento"]
	descricaoCol := colMap["Descrição dos alimentos"]
	energiaCol := colMap["Energia (kcal)"]
	proteinaCol := colMap["Proteína (g)"]
	lipideosCol := colMap["Lipídeos (g)"]
	carboidratoCol := colMap["Carboidrato (g)"]
	fibraCol := colMap["Fibra Alimentar (g)"]

	lineNumber := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Error("erro ao ler linha", "line", lineNumber, "error", err)
			lineNumber++
			continue
		}

		lineNumber++

		// Validar se tem dados suficientes
		if len(record) < len(header) {
			continue
		}

		// Extrair dados
		categoria := normalizarCategoria(record[categoriaCol])
		descricao := strings.TrimSpace(record[descricaoCol])
		
		// Pular se não tiver descrição
		if descricao == "" {
			continue
		}

		// Converter valores nutricionais
		calorias := parseFloat(record[energiaCol])
		proteina := parseFloat(record[proteinaCol])
		gordura := parseFloat(record[lipideosCol])
		carboidrato := parseFloat(record[carboidratoCol])
		fibra := parseFloat(record[fibraCol])

		// Pular se não tiver calorias (dado incompleto)
		if calorias == 0 {
			continue
		}

		ingredient := models.Ingredient{
			Name:     descricao,
			Category: categoria,
			Calories: calorias,
			Protein:  proteina,
			Fat:      gordura,
			Carbs:    carboidrato,
			Fiber:    fibra,
			Unit:     "g",
			Source:   "taco",
		}

		ingredients = append(ingredients, ingredient)
	}

	return ingredients, nil
}

// parseFloat converte string para float64, tratando valores inválidos
func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	
	// Substituir vírgula por ponto
	s = strings.Replace(s, ",", ".", 1)
	
	// Tratar valores especiais
	if s == "" || s == "NA" || s == "*" || s == "Tr" || s == " " {
		return 0
	}

	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}

	// Garantir que não seja negativo
	if value < 0 {
		return 0
	}

	return value
}

// normalizarCategoria normaliza o nome da categoria
func normalizarCategoria(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	
	// Remover "e derivados"
	s = strings.Replace(s, " e derivados", "", 1)
	
	// Simplificar nomes longos
	switch {
	case strings.Contains(s, "cereais"):
		return "cereais"
	case strings.Contains(s, "verduras") || strings.Contains(s, "hortaliças"):
		return "vegetais"
	case strings.Contains(s, "frutas"):
		return "frutas"
	case strings.Contains(s, "gorduras") || strings.Contains(s, "óleos"):
		return "óleos e gorduras"
	case strings.Contains(s, "pescados") || strings.Contains(s, "frutos do mar"):
		return "peixes e frutos do mar"
	case strings.Contains(s, "carnes"):
		return "carnes"
	case strings.Contains(s, "leite"):
		return "laticínios"
	case strings.Contains(s, "bebidas"):
		return "bebidas"
	case strings.Contains(s, "ovos"):
		return "ovos"
	case strings.Contains(s, "açucarados"):
		return "açúcares e doces"
	case strings.Contains(s, "leguminosas"):
		return "leguminosas"
	case strings.Contains(s, "nozes") || strings.Contains(s, "sementes"):
		return "nozes e sementes"
	case strings.Contains(s, "alimentos preparados"):
		return "preparados"
	case strings.Contains(s, "miscelâneas"):
		return "diversos"
	case strings.Contains(s, "industrializados"):
		return "industrializados"
	default:
		return s
	}
}
