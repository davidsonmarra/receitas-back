package test

import (
	"testing"
	"time"

	"github.com/davidsonmarra/receitas-app/pkg/auth"
)

func TestGenerateToken(t *testing.T) {
	userID := uint(123)
	email := "test@example.com"

	token, err := auth.GenerateToken(userID, email, "user")
	if err != nil {
		t.Fatalf("erro ao gerar token: %v", err)
	}

	if token == "" {
		t.Error("token não deve ser vazio")
	}
}

func TestValidateToken_Success(t *testing.T) {
	userID := uint(456)
	email := "valid@example.com"

	token, err := auth.GenerateToken(userID, email, "user")
	if err != nil {
		t.Fatalf("erro ao gerar token: %v", err)
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		t.Fatalf("erro ao validar token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("esperado userID %d, obteve %d", userID, claims.UserID)
	}

	if claims.Email != email {
		t.Errorf("esperado email %s, obteve %s", email, claims.Email)
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	invalidToken := "token.invalido.xyz"

	_, err := auth.ValidateToken(invalidToken)
	if err == nil {
		t.Error("esperava erro ao validar token inválido")
	}
}

func TestValidateToken_EmptyToken(t *testing.T) {
	_, err := auth.ValidateToken("")
	if err == nil {
		t.Error("esperava erro ao validar token vazio")
	}
}

func TestValidateToken_Expiration(t *testing.T) {
	userID := uint(789)
	email := "expiring@example.com"

	token, err := auth.GenerateToken(userID, email, "user")
	if err != nil {
		t.Fatalf("erro ao gerar token: %v", err)
	}

	// Validar que o token tem tempo de expiração configurado
	claims, err := auth.ValidateToken(token)
	if err != nil {
		t.Fatalf("erro ao validar token: %v", err)
	}

	// Verificar que expira em aproximadamente 24 horas
	expectedExpiration := time.Now().Add(24 * time.Hour)
	diff := claims.ExpiresAt.Time.Sub(expectedExpiration)

	// Permitir 1 minuto de diferença (tolerância para execução do teste)
	if diff > time.Minute || diff < -time.Minute {
		t.Errorf("expiração incorreta: esperado ~24h, diferença: %v", diff)
	}
}
