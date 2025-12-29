package handlers

import (
	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/database"
)

// isAdmin verifica se um userID é admin
// Retorna false em caso de erro (fail secure)
func isAdmin(userID uint) bool {
	var user models.User
	if err := database.DB.Select("role").First(&user, userID).Error; err != nil {
		return false // Fail secure: em caso de erro, negar acesso
	}
	return user.Role == "admin"
}

// getUserRole retorna o role de um usuário
// Retorna "user" como default em caso de erro
func getUserRole(userID uint) string {
	var user models.User
	if err := database.DB.Select("role").First(&user, userID).Error; err != nil {
		return "user" // Default: usuário comum
	}
	return user.Role
}
