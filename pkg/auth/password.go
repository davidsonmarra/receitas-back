package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword gera um hash bcrypt da senha
// Usa cost 12 para balanceamento entre segurança e performance
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword compara uma senha em texto com seu hash bcrypt
func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

