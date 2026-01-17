package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/auth"
	"github.com/davidsonmarra/receitas-app/pkg/database"
	"github.com/davidsonmarra/receitas-app/pkg/log"
	"github.com/davidsonmarra/receitas-app/pkg/response"
	"github.com/davidsonmarra/receitas-app/pkg/validation"
	"gorm.io/gorm"
)

// RegisterRequest representa os dados de cadastro
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginRequest representa os dados de login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse representa a resposta de autenticação
type AuthResponse struct {
	User         models.User `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int         `json:"expires_in"` // em segundos
	
	// Manter compatibilidade com versões antigas
	Token string `json:"token,omitempty"` // Deprecated: usar access_token
}

// Register cadastra um novo usuário
func Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ValidationError(w, "Formato de dados inválido.")
		return
	}

	// Validar os dados
	if errs := validation.ValidateStruct(req); errs != nil {
		message := validation.FormatErrors(errs)
		response.ValidationError(w, message)
		return
	}

	// Verificar se o email já existe
	var existingUser models.User
	err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error
	if err == nil {
		response.ValidationError(w, "E-mail já cadastrado.")
		return
	} else if err != gorm.ErrRecordNotFound {
		log.ErrorCtx(r.Context(), "failed to check existing email", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao verificar e-mail")
		return
	}

	// Hash da senha
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		log.ErrorCtx(r.Context(), "failed to hash password", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao processar senha")
		return
	}

	// Criar usuário
	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to create user", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao criar usuário")
		return
	}

	// Gerar access token JWT (incluindo role)
	accessToken, err := auth.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		log.ErrorCtx(r.Context(), "failed to generate access token", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao gerar token")
		return
	}

	// Gerar refresh token
	refreshToken, err := auth.CreateRefreshToken(auth.RefreshTokenInfo{
		UserID:            user.ID,
		Email:             user.Email,
		Role:              user.Role,
		DeviceName:        getDeviceName(r),
		DeviceFingerprint: getDeviceFingerprint(r),
		IPAddress:         getClientIP(r),
	})
	if err != nil {
		log.ErrorCtx(r.Context(), "failed to generate refresh token", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao gerar token")
		return
	}

	log.InfoCtx(r.Context(), "user registered", "id", user.ID, "email", user.Email, "role", user.Role)

	authResponse := AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    auth.GetAccessTokenDurationSeconds(),
		Token:        accessToken, // Compatibilidade
	}
	response.JSON(w, http.StatusCreated, authResponse)
}

// Login autentica um usuário
func Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ValidationError(w, "Formato de dados inválido.")
		return
	}

	// Validar os dados
	if errs := validation.ValidateStruct(req); errs != nil {
		message := validation.FormatErrors(errs)
		response.ValidationError(w, message)
		return
	}

	// Buscar usuário por email
	var user models.User
	err := database.DB.Where("email = ?", req.Email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(w, http.StatusUnauthorized, "E-mail ou senha inválidos")
		} else {
			log.ErrorCtx(r.Context(), "failed to find user", "error", err)
			response.Error(w, http.StatusInternalServerError, "Erro ao buscar usuário")
		}
		return
	}

	// Verificar senha
	if !auth.CheckPassword(user.Password, req.Password) {
		response.Error(w, http.StatusUnauthorized, "E-mail ou senha inválidos")
		return
	}

	// Gerar access token JWT (incluindo role)
	accessToken, err := auth.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		log.ErrorCtx(r.Context(), "failed to generate access token", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao gerar token")
		return
	}

	// Gerar refresh token
	refreshToken, err := auth.CreateRefreshToken(auth.RefreshTokenInfo{
		UserID:            user.ID,
		Email:             user.Email,
		Role:              user.Role,
		DeviceName:        getDeviceName(r),
		DeviceFingerprint: getDeviceFingerprint(r),
		IPAddress:         getClientIP(r),
	})
	if err != nil {
		log.ErrorCtx(r.Context(), "failed to generate refresh token", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao gerar token")
		return
	}

	log.InfoCtx(r.Context(), "user logged in", "id", user.ID, "email", user.Email, "role", user.Role)

	authResponse := AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    auth.GetAccessTokenDurationSeconds(),
		Token:        accessToken, // Compatibilidade
	}
	response.JSON(w, http.StatusOK, authResponse)
}

// LogoutRequest representa a requisição de logout (opcional)
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token,omitempty"` // Opcional: revogar token específico
}

// Logout invalida o token do usuário
func Logout(w http.ResponseWriter, r *http.Request) {
	// Extrair token do header Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		response.Error(w, http.StatusUnauthorized, "Token não fornecido")
		return
	}

	// Remover "Bearer " do início
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		response.Error(w, http.StatusUnauthorized, "Formato de token inválido")
		return
	}

	// Validar token para obter a expiração e user ID
	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Token inválido")
		return
	}

	// Adicionar access token à blacklist
	auth.AddToBlacklist(tokenString, claims.ExpiresAt.Time)

	// Tentar ler refresh token do body (opcional)
	var req LogoutRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	// Se refresh token foi fornecido, revogar apenas ele
	if req.RefreshToken != "" {
		if err := auth.RevokeRefreshTokenByString(req.RefreshToken); err != nil {
			log.ErrorCtx(r.Context(), "erro ao revogar refresh token específico", "error", err)
		} else {
			log.InfoCtx(r.Context(), "refresh token específico revogado", "user_id", claims.UserID)
		}
	} else {
		// Se não foi fornecido, revogar todos os tokens do usuário
		// (comportamento mais seguro para logout completo)
		if err := auth.RevokeAllUserTokens(claims.UserID); err != nil {
			log.ErrorCtx(r.Context(), "erro ao revogar refresh tokens", "error", err)
		} else {
			log.InfoCtx(r.Context(), "todos os refresh tokens revogados", "user_id", claims.UserID)
		}
	}

	log.InfoCtx(r.Context(), "user logged out", "user_id", claims.UserID)
	response.JSON(w, http.StatusOK, map[string]string{"message": "Logout realizado com sucesso"})
}
