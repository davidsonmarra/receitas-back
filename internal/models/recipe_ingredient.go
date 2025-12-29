package models

import "time"

// RecipeIngredient representa o relacionamento entre receitas e ingredientes
// Permite especificar quantidade, unidade e notas para cada ingrediente em uma receita
type RecipeIngredient struct {
	ID           uint       `gorm:"primarykey" json:"id"`
	RecipeID     uint       `gorm:"not null;index" json:"recipe_id" validate:"required"`
	Recipe       Recipe     `gorm:"foreignKey:RecipeID" json:"-"`
	IngredientID uint       `gorm:"not null;index" json:"ingredient_id" validate:"required"`
	Ingredient   Ingredient `gorm:"foreignKey:IngredientID" json:"ingredient"`
	Quantity     float64    `gorm:"not null" json:"quantity" validate:"required,gt=0"`
	Unit         string     `gorm:"not null;size:50" json:"unit" validate:"required"`
	Notes        string     `gorm:"size:200" json:"notes,omitempty"`
	Order        int        `gorm:"default:0" json:"order"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// TableName especifica o nome da tabela no banco de dados
func (RecipeIngredient) TableName() string {
	return "recipe_ingredients"
}
