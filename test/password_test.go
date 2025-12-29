package test

import (
	"testing"

	"github.com/davidsonmarra/receitas-app/pkg/auth"
)

func TestHashPassword(t *testing.T) {
	password := "minha-senha-123"

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("erro ao gerar hash: %v", err)
	}

	if hash == "" {
		t.Error("hash não deve ser vazio")
	}

	if hash == password {
		t.Error("hash não deve ser igual à senha em texto")
	}
}

func TestCheckPassword_Success(t *testing.T) {
	password := "minha-senha-123"

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("erro ao gerar hash: %v", err)
	}

	if !auth.CheckPassword(hash, password) {
		t.Error("senha correta deveria ser válida")
	}
}

func TestCheckPassword_WrongPassword(t *testing.T) {
	password := "senha-correta"
	wrongPassword := "senha-errada"

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("erro ao gerar hash: %v", err)
	}

	if auth.CheckPassword(hash, wrongPassword) {
		t.Error("senha incorreta não deveria ser válida")
	}
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	password := "mesma-senha"

	hash1, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("erro ao gerar primeiro hash: %v", err)
	}

	hash2, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("erro ao gerar segundo hash: %v", err)
	}

	// Bcrypt gera salts aleatórios, então hashes devem ser diferentes
	if hash1 == hash2 {
		t.Error("hashes da mesma senha deveriam ser diferentes (salt aleatório)")
	}

	// Mas ambos devem validar a senha original
	if !auth.CheckPassword(hash1, password) {
		t.Error("primeiro hash deveria validar a senha")
	}

	if !auth.CheckPassword(hash2, password) {
		t.Error("segundo hash deveria validar a senha")
	}
}
