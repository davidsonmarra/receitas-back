package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/davidsonmarra/receitas-app/internal/http/middleware"
	"github.com/davidsonmarra/receitas-app/pkg/auth"
	"github.com/davidsonmarra/receitas-app/pkg/log"
	"github.com/davidsonmarra/receitas-app/pkg/response"
	"github.com/davidsonmarra/receitas-app/pkg/validation"
)

// RefreshTokenRequest representa a requisição de refresh token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenResponse representa a resposta de refresh token
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // em segundos
}

// RevokeTokenRequest representa a requisição para revogar um token
type RevokeTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// DeviceInfo representa informações de um dispositivo ativo
type DeviceInfo struct {
	ID           string  `json:"id"`
	DeviceName   string  `json:"device_name"`
	IPAddress    string  `json:"ip_address"`
	LastUsedAt   *string `json:"last_used_at"`
	CreatedAt    string  `json:"created_at"`
	IsCurrent    bool    `json:"is_current"`
}

// DevicesResponse representa a resposta com lista de dispositivos
type DevicesResponse struct {
	Devices []DeviceInfo `json:"devices"`
}

// getDeviceFingerprint gera um fingerprint do dispositivo baseado no User-Agent
func getDeviceFingerprint(r *http.Request) string {
	userAgent := r.Header.Get("User-Agent")
	// Em produção, poderia incluir mais informações como Accept-Language, etc.
	return auth.HashString(userAgent)
}

// getClientIP extrai o IP do cliente da requisição
func getClientIP(r *http.Request) string {
	// Verificar headers de proxy primeiro
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For pode conter múltiplos IPs, pegar o primeiro
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}

	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	// Fallback para RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}

// getDeviceName extrai um nome amigável do dispositivo baseado no User-Agent
func getDeviceName(r *http.Request) string {
	userAgent := r.Header.Get("User-Agent")
	
	// Detectar dispositivos iOS
	if strings.Contains(userAgent, "iPhone") {
		return "iPhone"
	}
	if strings.Contains(userAgent, "iPad") {
		return "iPad"
	}
	
	// Detectar Android
	if strings.Contains(userAgent, "Android") {
		return "Android"
	}
	
	// Detectar navegadores desktop
	if strings.Contains(userAgent, "Windows") {
		return "Windows PC"
	}
	if strings.Contains(userAgent, "Macintosh") {
		return "Mac"
	}
	if strings.Contains(userAgent, "Linux") {
		return "Linux"
	}
	
	return "Dispositivo desconhecido"
}

// RefreshToken renova o access token usando um refresh token
func RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest

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

	// Obter informações do dispositivo
	deviceFingerprint := getDeviceFingerprint(r)
	ipAddress := getClientIP(r)

	// Validar e renovar tokens
	result, err := auth.RefreshAccessToken(req.RefreshToken, deviceFingerprint, ipAddress)
	if err != nil {
		// Mapear erros específicos
		switch err {
		case auth.ErrRefreshTokenNotFound:
			response.ErrorWithCode(w, http.StatusUnauthorized, "Refresh token inválido", "REFRESH_TOKEN_INVALID")
		case auth.ErrRefreshTokenExpired:
			response.ErrorWithCode(w, http.StatusUnauthorized, "Refresh token expirado", "REFRESH_TOKEN_EXPIRED")
		case auth.ErrRefreshTokenRevoked:
			response.ErrorWithCode(w, http.StatusUnauthorized, "Refresh token revogado", "REFRESH_TOKEN_REVOKED")
		case auth.ErrDeviceFingerprintMismatch:
			response.ErrorWithCode(w, http.StatusUnauthorized, "Dispositivo não reconhecido", "DEVICE_MISMATCH")
		default:
			log.ErrorCtx(r.Context(), "erro ao renovar token", "error", err)
			response.Error(w, http.StatusInternalServerError, "Erro ao renovar token")
		}
		return
	}

	log.InfoCtx(r.Context(), "token renovado com sucesso", "ip", ipAddress)

	// Retornar novos tokens
	resp := RefreshTokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	}
	response.JSON(w, http.StatusOK, resp)
}

// RevokeToken revoga um refresh token específico
func RevokeToken(w http.ResponseWriter, r *http.Request) {
	var req RevokeTokenRequest

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

	// Obter user ID do contexto
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Usuário não autenticado")
		return
	}

	// Validar que o token pertence ao usuário antes de revogar
	deviceFingerprint := getDeviceFingerprint(r)
	refreshToken, err := auth.ValidateRefreshToken(req.RefreshToken, deviceFingerprint)
	if err != nil {
		response.ErrorWithCode(w, http.StatusBadRequest, "Refresh token inválido", "REFRESH_TOKEN_INVALID")
		return
	}

	// Verificar se o token pertence ao usuário autenticado
	if refreshToken.UserID != userID {
		response.Error(w, http.StatusForbidden, "Token não pertence ao usuário")
		return
	}

	// Revogar o token
	if err := auth.RevokeRefreshTokenByString(req.RefreshToken); err != nil {
		log.ErrorCtx(r.Context(), "erro ao revogar token", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao revogar token")
		return
	}

	log.InfoCtx(r.Context(), "token revogado", "user_id", userID)
	response.JSON(w, http.StatusOK, map[string]string{"message": "Token revogado com sucesso"})
}

// RevokeAllTokens revoga todos os refresh tokens do usuário
func RevokeAllTokens(w http.ResponseWriter, r *http.Request) {
	// Obter user ID do contexto
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Usuário não autenticado")
		return
	}

	// Revogar todos os tokens
	if err := auth.RevokeAllUserTokens(userID); err != nil {
		log.ErrorCtx(r.Context(), "erro ao revogar todos os tokens", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao revogar tokens")
		return
	}

	log.InfoCtx(r.Context(), "todos os tokens revogados", "user_id", userID)
	response.JSON(w, http.StatusOK, map[string]string{"message": "Todos os tokens foram revogados com sucesso"})
}

// ListDevices lista todos os dispositivos ativos do usuário
func ListDevices(w http.ResponseWriter, r *http.Request) {
	// Obter user ID do contexto
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Usuário não autenticado")
		return
	}

	// Buscar tokens ativos
	tokens, err := auth.GetUserActiveTokens(userID)
	if err != nil {
		log.ErrorCtx(r.Context(), "erro ao buscar dispositivos", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao buscar dispositivos")
		return
	}

	// Obter fingerprint do dispositivo atual
	currentFingerprint := getDeviceFingerprint(r)

	// Converter para response
	devices := make([]DeviceInfo, 0, len(tokens))
	for _, token := range tokens {
		var lastUsedAt *string
		if token.LastUsedAt != nil {
			formatted := token.LastUsedAt.Format("2006-01-02T15:04:05Z07:00")
			lastUsedAt = &formatted
		}

		device := DeviceInfo{
			ID:         token.ID.String(),
			DeviceName: token.DeviceName,
			IPAddress:  token.IPAddress,
			LastUsedAt: lastUsedAt,
			CreatedAt:  token.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			IsCurrent:  token.DeviceFingerprint == currentFingerprint,
		}
		devices = append(devices, device)
	}

	resp := DevicesResponse{
		Devices: devices,
	}
	response.JSON(w, http.StatusOK, resp)
}

