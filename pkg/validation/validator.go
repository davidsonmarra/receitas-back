package validation

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidateStruct valida uma struct e retorna erros formatados em português
func ValidateStruct(s interface{}) []string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var errors []string
	for _, err := range err.(validator.ValidationErrors) {
		errors = append(errors, translateError(err))
	}

	return errors
}

// translateError traduz um erro de validação para português
func translateError(err validator.FieldError) string {
	field := translateField(err.Field())
	
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("o %s é obrigatório", field)
	case "min":
		if err.Type().String() == "string" {
			return fmt.Sprintf("o %s deve ter no mínimo %s caracteres", field, err.Param())
		}
		return fmt.Sprintf("o %s deve ser no mínimo %s", field, err.Param())
	case "max":
		if err.Type().String() == "string" {
			return fmt.Sprintf("o %s deve ter no máximo %s caracteres", field, err.Param())
		}
		return fmt.Sprintf("o %s deve ser no máximo %s", field, err.Param())
	case "oneof":
		options := strings.ReplaceAll(err.Param(), " ", ", ")
		return fmt.Sprintf("o %s deve ser uma das opções: %s", field, options)
	case "email":
		return fmt.Sprintf("o %s deve ser um email válido", field)
	case "url":
		return fmt.Sprintf("o %s deve ser uma URL válida", field)
	case "numeric":
		return fmt.Sprintf("o %s deve ser um número", field)
	case "alpha":
		return fmt.Sprintf("o %s deve conter apenas letras", field)
	case "alphanum":
		return fmt.Sprintf("o %s deve conter apenas letras e números", field)
	default:
		return fmt.Sprintf("o %s é inválido", field)
	}
}

// translateField traduz o nome do campo para português
func translateField(field string) string {
	translations := map[string]string{
		"Title":       "título",
		"Description": "descrição",
		"PrepTime":    "tempo de preparo",
		"Servings":    "número de porções",
		"Difficulty":  "dificuldade",
	}

	if translated, ok := translations[field]; ok {
		return translated
	}

	// Se não encontrar tradução, retorna em minúsculas
	return strings.ToLower(field)
}

// FormatErrors formata uma lista de erros retornando apenas o primeiro
func FormatErrors(errors []string) string {
	if len(errors) == 0 {
		return ""
	}

	// Retornar apenas o primeiro erro
	firstError := errors[0]
	// Primeira letra maiúscula e adicionar ponto final
	return strings.ToUpper(string(firstError[0])) + firstError[1:] + "."
}

