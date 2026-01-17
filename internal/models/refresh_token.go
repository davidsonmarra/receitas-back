package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RefreshToken representa um token de refresh no sistema
type RefreshToken struct {
	ID                uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	UserID            uint           `gorm:"not null;index:idx_refresh_tokens_user_id" json:"user_id"`
	User              *User          `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	TokenHash         string         `gorm:"uniqueIndex;not null;size:64" json:"-"` // SHA256 hash, nunca retornar
	DeviceName        string         `gorm:"size:255" json:"device_name"`
	DeviceFingerprint string         `gorm:"size:255" json:"-"` // Nunca retornar por segurança
	IPAddress         string         `gorm:"size:45" json:"ip_address"`
	ExpiresAt         time.Time      `gorm:"not null;index:idx_refresh_tokens_expires_at" json:"expires_at"`
	LastUsedAt        *time.Time     `json:"last_used_at"`
	RevokedAt         *time.Time     `json:"revoked_at"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName especifica o nome da tabela no banco de dados
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// BeforeCreate hook do GORM - gera UUID antes de criar
func (rt *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	return nil
}

// IsExpired verifica se o token está expirado
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsRevoked verifica se o token foi revogado
func (rt *RefreshToken) IsRevoked() bool {
	return rt.RevokedAt != nil
}

// IsValid verifica se o token é válido (não expirado e não revogado)
func (rt *RefreshToken) IsValid() bool {
	return !rt.IsExpired() && !rt.IsRevoked()
}

// UpdateLastUsed atualiza o timestamp de último uso
func (rt *RefreshToken) UpdateLastUsed() {
	now := time.Now()
	rt.LastUsedAt = &now
}

// Revoke marca o token como revogado
func (rt *RefreshToken) Revoke() {
	now := time.Now()
	rt.RevokedAt = &now
}

