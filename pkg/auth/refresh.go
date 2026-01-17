package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/database"
	"github.com/davidsonmarra/receitas-app/pkg/log"
	"gorm.io/gorm"
)

var (
	// ErrRefreshTokenNotFound indica que o token não foi encontrado
	ErrRefreshTokenNotFound = errors.New("refresh token não encontrado")
	// ErrRefreshTokenExpired indica que o token expirou
	ErrRefreshTokenExpired = errors.New("refresh token expirado")
	// ErrRefreshTokenRevoked indica que o token foi revogado
	ErrRefreshTokenRevoked = errors.New("refresh token revogado")
	// ErrRefreshTokenInvalid indica que o token é inválido
	ErrRefreshTokenInvalid = errors.New("refresh token inválido")
	// ErrDeviceFingerprintMismatch indica que o device fingerprint não corresponde
	ErrDeviceFingerprintMismatch = errors.New("device fingerprint não corresponde")
)

// Configurações de refresh token
var (
	MaxRefreshTokensPerUser  int
	EnableDeviceFingerprint  bool
)

func init() {
	// Configurar limite de tokens por usuário (padrão: 5)
	maxTokens := os.Getenv("MAX_REFRESH_TOKENS_PER_USER")
	if maxTokens == "" {
		MaxRefreshTokensPerUser = 5
	} else {
		max, err := strconv.Atoi(maxTokens)
		if err != nil || max <= 0 {
			MaxRefreshTokensPerUser = 5
		} else {
			MaxRefreshTokensPerUser = max
		}
	}

	// Configurar validação de device fingerprint (padrão: true)
	enableFingerprint := os.Getenv("ENABLE_DEVICE_FINGERPRINT")
	if enableFingerprint == "" || enableFingerprint == "true" {
		EnableDeviceFingerprint = true
	} else {
		EnableDeviceFingerprint = false
	}
}

// RefreshTokenInfo contém informações para criar um refresh token
type RefreshTokenInfo struct {
	UserID            uint
	Email             string
	Role              string
	DeviceName        string
	DeviceFingerprint string
	IPAddress         string
}

// RefreshTokenResult contém os tokens gerados após refresh
type RefreshTokenResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

// generateRandomToken gera um token aleatório seguro
func generateRandomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// hashToken gera SHA256 hash do token
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}

// HashString gera SHA256 hash de uma string (exportada para uso externo)
func HashString(s string) string {
	hash := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", hash)
}

// CreateRefreshToken cria um novo refresh token e o persiste no banco
func CreateRefreshToken(info RefreshTokenInfo) (string, error) {
	// Gerar token aleatório
	token, err := generateRandomToken()
	if err != nil {
		return "", fmt.Errorf("erro ao gerar token: %w", err)
	}

	// Prefixo para identificar tokens de refresh facilmente
	fullToken := "rt_" + token

	// Hash do token para armazenar no banco
	tokenHash := hashToken(fullToken)

	// Criar registro no banco
	refreshToken := models.RefreshToken{
		UserID:            info.UserID,
		TokenHash:         tokenHash,
		DeviceName:        info.DeviceName,
		DeviceFingerprint: info.DeviceFingerprint,
		IPAddress:         info.IPAddress,
		ExpiresAt:         time.Now().Add(RefreshTokenDuration),
	}

	if err := database.DB.Create(&refreshToken).Error; err != nil {
		return "", fmt.Errorf("erro ao salvar refresh token: %w", err)
	}

	// Limitar número de tokens por usuário
	go func() {
		if err := RevokeOldTokens(info.UserID, MaxRefreshTokensPerUser); err != nil {
			log.Error("erro ao revogar tokens antigos", "user_id", info.UserID, "error", err)
		}
	}()

	return fullToken, nil
}

// ValidateRefreshToken valida um refresh token e retorna suas informações
func ValidateRefreshToken(token string, deviceFingerprint string) (*models.RefreshToken, error) {
	// Calcular hash do token
	tokenHash := hashToken(token)

	// Buscar token no banco
	var refreshToken models.RefreshToken
	err := database.DB.Where("token_hash = ?", tokenHash).
		Preload("User").
		First(&refreshToken).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, fmt.Errorf("erro ao buscar refresh token: %w", err)
	}

	// Verificar se foi revogado
	if refreshToken.IsRevoked() {
		return nil, ErrRefreshTokenRevoked
	}

	// Verificar se expirou
	if refreshToken.IsExpired() {
		return nil, ErrRefreshTokenExpired
	}

	// Validar device fingerprint se habilitado
	if EnableDeviceFingerprint && deviceFingerprint != "" {
		if refreshToken.DeviceFingerprint != deviceFingerprint {
			// Log de segurança: possível tentativa de uso de token roubado
			log.Warn("device fingerprint mismatch",
				"user_id", refreshToken.UserID,
				"token_id", refreshToken.ID,
				"expected", refreshToken.DeviceFingerprint,
				"received", deviceFingerprint)
			return nil, ErrDeviceFingerprintMismatch
		}
	}

	return &refreshToken, nil
}

// RefreshAccessToken valida um refresh token e gera novos tokens (rotation)
func RefreshAccessToken(token string, deviceFingerprint string, ipAddress string) (*RefreshTokenResult, error) {
	// Validar refresh token
	refreshToken, err := ValidateRefreshToken(token, deviceFingerprint)
	if err != nil {
		return nil, err
	}

	// Atualizar last_used_at
	refreshToken.UpdateLastUsed()
	if err := database.DB.Save(refreshToken).Error; err != nil {
		log.Error("erro ao atualizar last_used_at", "error", err)
	}

	// Gerar novo access token
	accessToken, err := GenerateToken(refreshToken.UserID, refreshToken.User.Email, refreshToken.User.Role)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar access token: %w", err)
	}

	// Token rotation: revogar token antigo e criar novo
	if err := RevokeRefreshToken(refreshToken.TokenHash); err != nil {
		log.Error("erro ao revogar token antigo", "error", err)
	}

	// Criar novo refresh token
	newRefreshToken, err := CreateRefreshToken(RefreshTokenInfo{
		UserID:            refreshToken.UserID,
		Email:             refreshToken.User.Email,
		Role:              refreshToken.User.Role,
		DeviceName:        refreshToken.DeviceName,
		DeviceFingerprint: deviceFingerprint,
		IPAddress:         ipAddress,
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao criar novo refresh token: %w", err)
	}

	return &RefreshTokenResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    GetAccessTokenDurationSeconds(),
	}, nil
}

// RevokeRefreshToken revoga um refresh token específico pelo hash
func RevokeRefreshToken(tokenHash string) error {
	result := database.DB.Model(&models.RefreshToken{}).
		Where("token_hash = ? AND revoked_at IS NULL", tokenHash).
		Update("revoked_at", time.Now())

	if result.Error != nil {
		return fmt.Errorf("erro ao revogar token: %w", result.Error)
	}

	return nil
}

// RevokeRefreshTokenByString revoga um refresh token específico pela string do token
func RevokeRefreshTokenByString(token string) error {
	tokenHash := hashToken(token)
	return RevokeRefreshToken(tokenHash)
}

// RevokeAllUserTokens revoga todos os refresh tokens de um usuário
func RevokeAllUserTokens(userID uint) error {
	result := database.DB.Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", time.Now())

	if result.Error != nil {
		return fmt.Errorf("erro ao revogar todos os tokens: %w", result.Error)
	}

	log.Info("todos os tokens do usuário foram revogados", "user_id", userID, "count", result.RowsAffected)
	return nil
}

// RevokeOldTokens mantém apenas os N tokens mais recentes de um usuário
func RevokeOldTokens(userID uint, keepCount int) error {
	// Buscar todos os tokens não revogados do usuário, ordenados por data de criação
	var tokens []models.RefreshToken
	err := database.DB.Where("user_id = ? AND revoked_at IS NULL", userID).
		Order("created_at DESC").
		Find(&tokens).Error

	if err != nil {
		return fmt.Errorf("erro ao buscar tokens: %w", err)
	}

	// Se tiver mais tokens que o limite, revogar os mais antigos
	if len(tokens) > keepCount {
		tokensToRevoke := tokens[keepCount:]
		for _, token := range tokensToRevoke {
			if err := RevokeRefreshToken(token.TokenHash); err != nil {
				log.Error("erro ao revogar token antigo", "token_id", token.ID, "error", err)
			}
		}
		log.Info("tokens antigos revogados", "user_id", userID, "revoked_count", len(tokensToRevoke))
	}

	return nil
}

// GetUserActiveTokens retorna todos os tokens ativos de um usuário
func GetUserActiveTokens(userID uint) ([]models.RefreshToken, error) {
	var tokens []models.RefreshToken
	err := database.DB.Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now()).
		Order("last_used_at DESC NULLS LAST, created_at DESC").
		Find(&tokens).Error

	if err != nil {
		return nil, fmt.Errorf("erro ao buscar tokens ativos: %w", err)
	}

	return tokens, nil
}

// CleanupExpiredTokens remove tokens expirados do banco de dados
func CleanupExpiredTokens() error {
	result := database.DB.Unscoped().
		Where("expires_at < ?", time.Now()).
		Delete(&models.RefreshToken{})

	if result.Error != nil {
		return fmt.Errorf("erro ao limpar tokens expirados: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		log.Info("tokens expirados removidos", "count", result.RowsAffected)
	}

	return nil
}

// StartRefreshTokenCleanup inicia um job periódico para limpar tokens expirados
func StartRefreshTokenCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			if err := CleanupExpiredTokens(); err != nil {
				log.Error("erro ao limpar tokens expirados", "error", err)
			}
		}
	}()
	log.Info("job de limpeza de refresh tokens iniciado", "interval", interval)
}

