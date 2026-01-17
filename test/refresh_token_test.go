package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/davidsonmarra/receitas-app/internal/http/handlers"
	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/auth"
	"github.com/davidsonmarra/receitas-app/pkg/database"
	"github.com/davidsonmarra/receitas-app/test/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testUserAgent = "test-agent"

func TestRefreshToken_Success(t *testing.T) {
	cleanup := testdb.Setup(t)
	defer cleanup()

	// Criar usuário
	user := models.User{
		Name:     "Test User",
		Email:    "refresh@test.com",
		Password: "hashedpassword",
		Role:     "user",
	}
	require.NoError(t, database.DB.Create(&user).Error)

	// Criar refresh token
	refreshToken, err := auth.CreateRefreshToken(auth.RefreshTokenInfo{
		UserID:            user.ID,
		Email:             user.Email,
		Role:              user.Role,
		DeviceName:        "Test Device",
		DeviceFingerprint: auth.HashString(testUserAgent),
		IPAddress:         "127.0.0.1",
	})
	require.NoError(t, err)

	// Fazer requisição de refresh
	reqBody := map[string]string{
		"refresh_token": refreshToken,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", testUserAgent)

	w := httptest.NewRecorder()
	handlers.RefreshToken(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotEmpty(t, response["access_token"])
	assert.NotEmpty(t, response["refresh_token"])
	assert.NotEmpty(t, response["expires_in"])

	// Verificar que o novo refresh token é diferente do antigo
	assert.NotEqual(t, refreshToken, response["refresh_token"])
}

func TestRefreshToken_ExpiredToken(t *testing.T) {
	cleanup := testdb.Setup(t)
	defer cleanup()

	// Criar usuário
	user := models.User{
		Name:     "Test User",
		Email:    "expired@test.com",
		Password: "hashedpassword",
		Role:     "user",
	}
	require.NoError(t, database.DB.Create(&user).Error)

	// Gerar um JWT válido primeiro
	validJWT, err := auth.GenerateRefreshTokenJWT(user.ID, user.Email, user.Role)
	require.NoError(t, err)

	// Criar token expirado no banco usando o hash do JWT real
	expiredToken := models.RefreshToken{
		UserID:            user.ID,
		TokenHash:         auth.HashString(validJWT),
		DeviceName:        "Test Device",
		DeviceFingerprint: auth.HashString(testUserAgent),
		IPAddress:         "127.0.0.1",
		ExpiresAt:         time.Now().Add(-1 * time.Hour), // Expirado
	}
	require.NoError(t, database.DB.Create(&expiredToken).Error)

	// Fazer requisição de refresh com o JWT real
	reqBody := map[string]string{
		"refresh_token": validJWT,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", testUserAgent)

	w := httptest.NewRecorder()
	handlers.RefreshToken(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	errorMap := response["error"].(map[string]interface{})
	assert.Contains(t, errorMap["code"], "REFRESH_TOKEN_EXPIRED")
}

func TestRefreshToken_RevokedToken(t *testing.T) {
	cleanup := testdb.Setup(t)
	defer cleanup()

	// Criar usuário
	user := models.User{
		Name:     "Test User",
		Email:    "revoked@test.com",
		Password: "hashedpassword",
		Role:     "user",
	}
	require.NoError(t, database.DB.Create(&user).Error)

	// Criar refresh token
	refreshToken, err := auth.CreateRefreshToken(auth.RefreshTokenInfo{
		UserID:            user.ID,
		Email:             user.Email,
		Role:              user.Role,
		DeviceName:        "Test Device",
		DeviceFingerprint: auth.HashString(testUserAgent),
		IPAddress:         "127.0.0.1",
	})
	require.NoError(t, err)

	// Revogar o token
	err = auth.RevokeRefreshTokenByString(refreshToken)
	require.NoError(t, err)

	// Tentar usar o token revogado
	reqBody := map[string]string{
		"refresh_token": refreshToken,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", testUserAgent)

	w := httptest.NewRecorder()
	handlers.RefreshToken(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	errorMap := response["error"].(map[string]interface{})
	assert.Contains(t, errorMap["code"], "REFRESH_TOKEN_REVOKED")
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	cleanup := testdb.Setup(t)
	defer cleanup()

	// Fazer requisição com token inexistente
	reqBody := map[string]string{
		"refresh_token": "rt_invalid_token_that_does_not_exist",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", testUserAgent)

	w := httptest.NewRecorder()
	handlers.RefreshToken(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	errorMap := response["error"].(map[string]interface{})
	assert.Contains(t, errorMap["code"], "REFRESH_TOKEN_INVALID")
}

func TestRefreshToken_TokenRotation(t *testing.T) {
	cleanup := testdb.Setup(t)
	defer cleanup()

	// Criar usuário
	user := models.User{
		Name:     "Test User",
		Email:    "rotation@test.com",
		Password: "hashedpassword",
		Role:     "user",
	}
	require.NoError(t, database.DB.Create(&user).Error)

	// Criar refresh token
	oldRefreshToken, err := auth.CreateRefreshToken(auth.RefreshTokenInfo{
		UserID:            user.ID,
		Email:             user.Email,
		Role:              user.Role,
		DeviceName:        "Test Device",
		DeviceFingerprint: auth.HashString(testUserAgent),
		IPAddress:         "127.0.0.1",
	})
	require.NoError(t, err)

	// Fazer refresh (primeira vez)
	reqBody := map[string]string{
		"refresh_token": oldRefreshToken,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", testUserAgent)

	w := httptest.NewRecorder()
	handlers.RefreshToken(w, req)

	if w.Code != http.StatusOK {
		t.Logf("First refresh response (should succeed): status=%d body=%s", w.Code, w.Body.String())
	}
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	newRefreshToken := response["refresh_token"].(string)

	// Tentar usar o token antigo novamente (deve falhar - token rotation)
	reqBody2 := map[string]string{
		"refresh_token": oldRefreshToken,
	}
	body2, _ := json.Marshal(reqBody2)
	req2 := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("User-Agent", testUserAgent)

	w2 := httptest.NewRecorder()
	handlers.RefreshToken(w2, req2)

	if w2.Code != http.StatusUnauthorized {
		t.Logf("Second refresh response (should fail): %s", w2.Body.String())
	}
	assert.Equal(t, http.StatusUnauthorized, w2.Code)

	// Usar o novo token deve funcionar
	reqBody3 := map[string]string{
		"refresh_token": newRefreshToken,
	}
	body3, _ := json.Marshal(reqBody3)
	req3 := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body3))
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("User-Agent", testUserAgent)

	w3 := httptest.NewRecorder()
	handlers.RefreshToken(w3, req3)

	if w3.Code != http.StatusOK {
		t.Logf("Third refresh response (should succeed): %s", w3.Body.String())
	}
	assert.Equal(t, http.StatusOK, w3.Code)
}

func TestRevokeAllUserTokens(t *testing.T) {
	cleanup := testdb.Setup(t)
	defer cleanup()

	// Criar usuário
	user := models.User{
		Name:     "Test User",
		Email:    "revokeall@test.com",
		Password: "hashedpassword",
		Role:     "user",
	}
	require.NoError(t, database.DB.Create(&user).Error)

	// Criar múltiplos refresh tokens
	for i := 0; i < 3; i++ {
		_, err := auth.CreateRefreshToken(auth.RefreshTokenInfo{
			UserID:            user.ID,
			Email:             user.Email,
			Role:              user.Role,
			DeviceName:        "Test Device",
			DeviceFingerprint: auth.HashString(testUserAgent),
			IPAddress:         "127.0.0.1",
		})
		require.NoError(t, err)
	}

	// Verificar que existem 3 tokens ativos
	tokens, err := auth.GetUserActiveTokens(user.ID)
	require.NoError(t, err)
	assert.Len(t, tokens, 3)

	// Revogar todos
	err = auth.RevokeAllUserTokens(user.ID)
	require.NoError(t, err)

	// Verificar que não há mais tokens ativos
	tokens, err = auth.GetUserActiveTokens(user.ID)
	require.NoError(t, err)
	assert.Len(t, tokens, 0)
}

func TestMaxRefreshTokensPerUser(t *testing.T) {
	cleanup := testdb.Setup(t)
	defer cleanup()

	// Criar usuário
	user := models.User{
		Name:     "Test User",
		Email:    "maxtoken@test.com",
		Password: "hashedpassword",
		Role:     "user",
	}
	require.NoError(t, database.DB.Create(&user).Error)

	// Criar mais tokens que o limite (5)
	for i := 0; i < 7; i++ {
		_, err := auth.CreateRefreshToken(auth.RefreshTokenInfo{
			UserID:            user.ID,
			Email:             user.Email,
			Role:              user.Role,
			DeviceName:        "Test Device",
			DeviceFingerprint: auth.HashString(testUserAgent),
			IPAddress:         "127.0.0.1",
		})
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // Garantir ordem de criação
	}

	// Aguardar goroutine de limpeza
	time.Sleep(100 * time.Millisecond)

	// Verificar que apenas 5 tokens estão ativos
	tokens, err := auth.GetUserActiveTokens(user.ID)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(tokens), auth.MaxRefreshTokensPerUser)
}

func TestCleanupExpiredTokens(t *testing.T) {
	cleanup := testdb.Setup(t)
	defer cleanup()

	// Criar usuário
	user := models.User{
		Name:     "Test User",
		Email:    "cleanup@test.com",
		Password: "hashedpassword",
		Role:     "user",
	}
	require.NoError(t, database.DB.Create(&user).Error)

	// Criar token expirado
	tokenHash := auth.HashString("expired-token-cleanup")
	expiredToken := models.RefreshToken{
		UserID:            user.ID,
		TokenHash:         tokenHash,
		DeviceName:        "Test Device",
		DeviceFingerprint: auth.HashString(testUserAgent),
		IPAddress:         "127.0.0.1",
		ExpiresAt:         time.Now().Add(-1 * time.Hour),
	}
	require.NoError(t, database.DB.Create(&expiredToken).Error)

	// Criar token válido
	_, err := auth.CreateRefreshToken(auth.RefreshTokenInfo{
		UserID:            user.ID,
		Email:             user.Email,
		Role:              user.Role,
		DeviceName:        "Test Device",
		DeviceFingerprint: auth.HashString(testUserAgent),
		IPAddress:         "127.0.0.1",
	})
	require.NoError(t, err)

	// Executar cleanup
	err = auth.CleanupExpiredTokens()
	require.NoError(t, err)

	// Verificar que apenas o token válido permanece
	var count int64
	database.DB.Model(&models.RefreshToken{}).Where("user_id = ?", user.ID).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestLoginAndRegister_ReturnRefreshToken(t *testing.T) {
	testdb.SetupWithCleanup(t)
	router := setupRouter()

	// Test Register
	user := createTestUser(t, "newuser@test.com", "password123", "New User")
	require.NotNil(t, user)

	// Test Login usando helper que funciona
	loginBody := map[string]string{
		"email":    "newuser@test.com",
		"password": "password123",
	}
	body, _ := json.Marshal(loginBody)
	req := httptest.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", testUserAgent)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var loginResponse map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&loginResponse)
	require.NoError(t, err)

	// Verificar que o login retorna access_token, refresh_token e expires_in
	assert.NotEmpty(t, loginResponse["access_token"])
	assert.NotEmpty(t, loginResponse["refresh_token"])
	assert.NotEmpty(t, loginResponse["expires_in"])

	// Verificar que o refresh token foi salvo no banco
	var tokenCount int64
	err = database.DB.Model(&models.RefreshToken{}).Where("user_id = ?", user.ID).Count(&tokenCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(1), tokenCount, "deve haver exatamente 1 refresh token no banco")
}

func TestAccessToken_CannotBeUsedForRefresh(t *testing.T) {
	cleanup := testdb.Setup(t)
	defer cleanup()

	// Criar usuário
	user := models.User{
		Name:     "Test User",
		Email:    "accesstoken@test.com",
		Password: "hashedpassword",
		Role:     "user",
	}
	require.NoError(t, database.DB.Create(&user).Error)

	// Gerar access token
	accessToken, err := auth.GenerateToken(user.ID, user.Email, user.Role)
	require.NoError(t, err)

	// Tentar usar access token no endpoint de refresh
	reqBody := map[string]string{
		"refresh_token": accessToken,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", testUserAgent)

	w := httptest.NewRecorder()
	handlers.RefreshToken(w, req)

	// Deve falhar porque não é um refresh token válido
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

