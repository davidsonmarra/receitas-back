package models

import (
	"time"

	"gorm.io/gorm"
)

// User representa um usuário no sistema
type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Name      string         `gorm:"not null;size:100" json:"name" validate:"required,min=3,max=100"`
	Email     string         `gorm:"uniqueIndex;not null;size:255" json:"email" validate:"required,email"`
	Password  string         `gorm:"not null" json:"-" validate:"required,min=6"`
	Role      string         `gorm:"default:'user';size:20" json:"role"` // 'user' ou 'admin'
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName especifica o nome da tabela no banco de dados
func (User) TableName() string {
	return "users"
}

// IsAdmin verifica se o usuário é admin (preparado para futuro)
// Descomentar quando implementar sistema de admin completo
// func (u *User) IsAdmin() bool {
// 	return u.Role == "admin"
// }
