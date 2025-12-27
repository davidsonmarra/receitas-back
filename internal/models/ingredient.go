package models

import "time"

// Ingredient representa um ingrediente no sistema
// Contém informações nutricionais baseadas em 100g do alimento
type Ingredient struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Name      string    `gorm:"uniqueIndex;not null;size:200" json:"name" validate:"required"`
	Calories  float64   `gorm:"not null" json:"calories" validate:"gte=0"`
	Protein   float64   `gorm:"default:0" json:"protein" validate:"gte=0"`
	Carbs     float64   `gorm:"default:0" json:"carbs" validate:"gte=0"`
	Fat       float64   `gorm:"default:0" json:"fat" validate:"gte=0"`
	Fiber     float64   `gorm:"default:0" json:"fiber,omitempty" validate:"gte=0"`
	Category  string    `gorm:"size:100;index" json:"category"`
	Unit      string    `gorm:"size:50;default:'g'" json:"unit"`
	Source    string    `gorm:"size:50" json:"source"` // "taco", "manual"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName especifica o nome da tabela no banco de dados
func (Ingredient) TableName() string {
	return "ingredients"
}

