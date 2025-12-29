package main

import (
	"fmt"
	"os"

	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/auth"
	"github.com/davidsonmarra/receitas-app/pkg/database"
	"github.com/davidsonmarra/receitas-app/pkg/log"
)

func main() {
	// Inicializar logger
	logConfig := log.Config{
		Level:       "info",
		Development: true,
	}
	if err := log.Init(logConfig); err != nil {
		fmt.Printf("❌ Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Conectar database
	if err := database.Connect(); err != nil {
		log.Error("failed to connect to database", "error", err)
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	fmt.Println("🔍 Verificando se já existe admin...")

	// Verificar se já existe admin
	var existingAdmin models.User
	err := database.DB.Where("role = ?", "admin").First(&existingAdmin).Error
	if err == nil {
		log.Info("admin already exists", "email", existingAdmin.Email)
		fmt.Printf("✅ Admin já existe: %s\n", existingAdmin.Email)
		fmt.Printf("   Nome: %s\n", existingAdmin.Name)
		fmt.Printf("   ID: %d\n", existingAdmin.ID)
		return
	}

	fmt.Println("📝 Criando admin...")

	// Obter credenciais (variáveis de ambiente ou defaults)
	email := os.Getenv("ADMIN_EMAIL")
	if email == "" {
		email = "admin@receitas.com"
	}

	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		password = "admin123" // ⚠️ TROCAR EM PRODUÇÃO!
	}

	name := os.Getenv("ADMIN_NAME")
	if name == "" {
		name = "Administrador"
	}

	// Hash da senha
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		log.Error("failed to hash password", "error", err)
		fmt.Printf("❌ Failed to hash password: %v\n", err)
		os.Exit(1)
	}

	// Criar admin
	admin := models.User{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
		Role:     "admin",
	}

	if err := database.DB.Create(&admin).Error; err != nil {
		log.Error("failed to create admin", "error", err)
		fmt.Printf("❌ Failed to create admin: %v\n", err)
		os.Exit(1)
	}

	log.Info("admin created successfully", "email", admin.Email, "id", admin.ID)

	fmt.Println("\n✅ Admin criado com sucesso!")
	fmt.Println("═══════════════════════════════════════")
	fmt.Printf("   Nome:  %s\n", admin.Name)
	fmt.Printf("   Email: %s\n", admin.Email)
	fmt.Printf("   Senha: %s\n", password)
	fmt.Printf("   ID:    %d\n", admin.ID)
	fmt.Println("═══════════════════════════════════════")
	fmt.Println("⚠️  IMPORTANTE: TROCAR SENHA EM PRODUÇÃO!")
	fmt.Println()
}
