package auth

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType define o tipo de token JWT
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// Claims representa as informações armazenadas no JWT
type Claims struct {
	UserID    uint      `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`       // 'user' ou 'admin'
	TokenType TokenType `json:"token_type"` // 'access' ou 'refresh'
	jwt.RegisteredClaims
}

var jwtSecret []byte

// Durações configuráveis dos tokens (em minutos)
var (
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
)

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Para desenvolvimento, usar um secret padrão (não recomendado para produção)
		secret = "default-secret-change-in-production"
	}
	jwtSecret = []byte(secret)

	// Configurar duração do access token (padrão: 15 minutos)
	accessDuration := os.Getenv("ACCESS_TOKEN_DURATION")
	if accessDuration == "" {
		AccessTokenDuration = 15 * time.Minute
	} else {
		minutes, err := strconv.Atoi(accessDuration)
		if err != nil || minutes <= 0 {
			AccessTokenDuration = 15 * time.Minute
		} else {
			AccessTokenDuration = time.Duration(minutes) * time.Minute
		}
	}

	// Configurar duração do refresh token (padrão: 30 dias = 43200 minutos)
	refreshDuration := os.Getenv("REFRESH_TOKEN_DURATION")
	if refreshDuration == "" {
		RefreshTokenDuration = 30 * 24 * time.Hour
	} else {
		minutes, err := strconv.Atoi(refreshDuration)
		if err != nil || minutes <= 0 {
			RefreshTokenDuration = 30 * 24 * time.Hour
		} else {
			RefreshTokenDuration = time.Duration(minutes) * time.Minute
		}
	}
}

// GenerateToken gera um novo access token JWT para o usuário
// Token expira em 15 minutos (configurável via ACCESS_TOKEN_DURATION)
// Inclui role para facilitar verificações rápidas (mas sempre verificar banco para segurança)
func GenerateToken(userID uint, email string, role string) (string, error) {
	return generateTokenWithType(userID, email, role, TokenTypeAccess, AccessTokenDuration)
}

// GenerateRefreshTokenJWT gera um novo refresh token JWT (usado internamente)
// Token expira em 30 dias (configurável via REFRESH_TOKEN_DURATION)
// Este é apenas o JWT - o token completo é gerenciado pelo serviço de refresh
func GenerateRefreshTokenJWT(userID uint, email string, role string) (string, error) {
	return generateTokenWithType(userID, email, role, TokenTypeRefresh, RefreshTokenDuration)
}

// generateTokenWithType gera um token JWT com tipo e duração específicos
func generateTokenWithType(userID uint, email string, role string, tokenType TokenType, duration time.Duration) (string, error) {
	claims := Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// GetAccessTokenDurationSeconds retorna a duração do access token em segundos
// Útil para retornar ao cliente o tempo de expiração
func GetAccessTokenDurationSeconds() int {
	return int(AccessTokenDuration.Seconds())
}

// ValidateToken valida um token JWT e retorna suas claims
func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validar que o método de assinatura é o esperado
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de assinatura inválido")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("token inválido")
}
