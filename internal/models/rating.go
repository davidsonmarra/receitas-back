package models

import (
	"time"

	"gorm.io/gorm"
)

// Rating representa uma avaliação de uma receita por um usuário
type Rating struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	RecipeID  uint           `gorm:"not null;index:idx_recipe_user,priority:1;index" json:"recipe_id" validate:"required"`
	UserID    uint           `gorm:"not null;index:idx_recipe_user,priority:2;index" json:"user_id" validate:"required"`
	Score     int            `gorm:"not null;check:score >= 1 AND score <= 5" json:"score" validate:"required,min=1,max=5"`
	Comment   string         `gorm:"type:text" json:"comment,omitempty" validate:"omitempty,max=1000"`
	Recipe    *Recipe        `gorm:"foreignKey:RecipeID" json:"recipe,omitempty"`
	User      *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName especifica o nome da tabela no banco de dados
func (Rating) TableName() string {
	return "ratings"
}

// BeforeCreate hook para garantir que o score está no range correto
func (r *Rating) BeforeCreate(tx *gorm.DB) error {
	if r.Score < 1 || r.Score > 5 {
		return gorm.ErrInvalidValue
	}
	return nil
}

// BeforeUpdate hook para garantir que o score está no range correto
func (r *Rating) BeforeUpdate(tx *gorm.DB) error {
	if r.Score < 1 || r.Score > 5 {
		return gorm.ErrInvalidValue
	}
	return nil
}

