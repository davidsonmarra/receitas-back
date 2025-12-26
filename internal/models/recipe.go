package models

import (
	"time"

	"gorm.io/gorm"
)

// Recipe representa uma receita no sistema
type Recipe struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Title       string         `gorm:"not null;size:200" json:"title" validate:"required,min=3,max=200"`
	Description string         `gorm:"type:text" json:"description"`
	PrepTime    int            `gorm:"not null" json:"prep_time" validate:"required,min=1"` // minutos
	Servings    int            `gorm:"not null;default:1" json:"servings" validate:"required,min=1"`
	Difficulty  string         `gorm:"size:50" json:"difficulty" validate:"omitempty,oneof=fácil média difícil"`
	UserID      *uint          `gorm:"index" json:"user_id,omitempty"` // NULL = receita geral, NOT NULL = receita do usuário
	User        *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt   time.Time      `gorm:"index" json:"created_at"` // Índice para ordenação rápida
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName especifica o nome da tabela no banco de dados
func (Recipe) TableName() string {
	return "recipes"
}
