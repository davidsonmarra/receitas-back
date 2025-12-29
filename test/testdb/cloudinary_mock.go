package testdb

import (
	"context"
	"fmt"
	"io"

	"github.com/davidsonmarra/receitas-app/pkg/storage"
)

// MockCloudinaryService é um mock do CloudinaryService para testes
// Implementa a interface storage.ImageService
type MockCloudinaryService struct {
	ShouldFailUpload bool
	ShouldFailDelete bool
}

// Garantir que MockCloudinaryService implementa storage.ImageService
var _ storage.ImageService = (*MockCloudinaryService)(nil)

// NewMockCloudinaryService cria uma nova instância do mock
func NewMockCloudinaryService() *MockCloudinaryService {
	return &MockCloudinaryService{
		ShouldFailUpload: false,
		ShouldFailDelete: false,
	}
}

// UploadImage simula o upload de uma imagem
func (m *MockCloudinaryService) UploadImage(ctx context.Context, params storage.UploadImageParams) (*storage.UploadResult, error) {
	// Validações que o serviço real faz
	if params.File == nil {
		return nil, fmt.Errorf("arquivo não pode ser nulo")
	}

	// Ler o arquivo para verificar se está vazio
	data, err := io.ReadAll(params.File)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("arquivo vazio - nenhum byte foi lido")
	}

	if params.FileName == "" {
		return nil, fmt.Errorf("nome do arquivo não pode ser vazio")
	}

	// Validar extensão do arquivo
	validExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	isValid := false
	for _, ext := range validExtensions {
		if len(params.FileName) >= len(ext) && params.FileName[len(params.FileName)-len(ext):] == ext {
			isValid = true
			break
		}
	}
	if !isValid {
		return nil, fmt.Errorf("formato de arquivo não suportado. Use: jpg, jpeg, png, gif ou webp")
	}

	if m.ShouldFailUpload {
		return nil, fmt.Errorf("erro simulado de upload")
	}

	// Retornar resultado simulado
	return &storage.UploadResult{
		PublicID:  "mock/test-image-123",
		SecureURL: "https://res.cloudinary.com/mock/image/upload/v1234567890/mock/test-image-123.jpg",
		URL:       "http://res.cloudinary.com/mock/image/upload/v1234567890/mock/test-image-123.jpg",
		Format:    "jpg",
		Width:     1024,
		Height:    768,
		Bytes:     len(data),
	}, nil
}

// DeleteImage simula a exclusão de uma imagem
func (m *MockCloudinaryService) DeleteImage(ctx context.Context, publicID string) error {
	if publicID == "" {
		return fmt.Errorf("publicID não pode ser vazio")
	}

	if m.ShouldFailDelete {
		return fmt.Errorf("erro simulado ao deletar imagem")
	}

	return nil
}

// GetOptimizedURL retorna uma URL otimizada simulada
func (m *MockCloudinaryService) GetOptimizedURL(publicID string, width, height int, quality string) (string, error) {
	if publicID == "" {
		return "", fmt.Errorf("publicID não pode ser vazio")
	}

	// Retornar URL simulada com os parâmetros
	url := fmt.Sprintf("https://res.cloudinary.com/mock/image/upload/w_%d,h_%d,q_%s/%s", 
		width, height, quality, publicID)
	
	return url, nil
}

// GetImageVariants retorna variantes simuladas de uma imagem
func (m *MockCloudinaryService) GetImageVariants(publicID string) map[string]string {
	if publicID == "" {
		return map[string]string{}
	}

	return map[string]string{
		"thumbnail": fmt.Sprintf("https://res.cloudinary.com/mock/image/upload/w_150,h_150,c_fill/%s", publicID),
		"small":     fmt.Sprintf("https://res.cloudinary.com/mock/image/upload/w_400,h_400,c_fit/%s", publicID),
		"medium":    fmt.Sprintf("https://res.cloudinary.com/mock/image/upload/w_800,h_800,c_fit/%s", publicID),
		"large":     fmt.Sprintf("https://res.cloudinary.com/mock/image/upload/w_1200,h_1200,c_fit/%s", publicID),
		"original":  fmt.Sprintf("https://res.cloudinary.com/mock/image/upload/%s", publicID),
	}
}

