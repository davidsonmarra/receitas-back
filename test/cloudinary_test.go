package test

import (
	"bytes"
	"context"
	"mime/multipart"
	"os"
	"testing"

	"github.com/davidsonmarra/receitas-app/pkg/storage"
)

// mockFile cria um multipart.File mock para testes
type mockFile struct {
	*bytes.Reader
}

func (m *mockFile) Close() error {
	return nil
}

func newMockFile(data []byte) multipart.File {
	return &mockFile{
		Reader: bytes.NewReader(data),
	}
}

func TestNewCloudinaryService_MissingURL(t *testing.T) {
	// Salvar valor original
	originalURL := os.Getenv("CLOUDINARY_URL")
	defer os.Setenv("CLOUDINARY_URL", originalURL)

	// Remover CLOUDINARY_URL
	os.Unsetenv("CLOUDINARY_URL")

	_, err := storage.NewCloudinaryService()
	if err == nil {
		t.Error("esperava erro quando CLOUDINARY_URL não está configurado")
	}

	if err.Error() != "CLOUDINARY_URL não configurado" {
		t.Errorf("mensagem de erro incorreta: %v", err)
	}
}

func TestNewCloudinaryService_InvalidFormat(t *testing.T) {
	// Salvar valor original
	originalURL := os.Getenv("CLOUDINARY_URL")
	defer os.Setenv("CLOUDINARY_URL", originalURL)

	// URL inválida
	os.Setenv("CLOUDINARY_URL", "invalid-url")

	_, err := storage.NewCloudinaryService()
	if err == nil {
		t.Error("esperava erro com URL inválida")
	}
}

func TestValidateImageFile_ValidExtensions(t *testing.T) {
	validFiles := []string{
		"image.jpg",
		"photo.jpeg",
		"icon.png",
		"animation.gif",
		"modern.webp",
		"bitmap.bmp",
	}

	for _, filename := range validFiles {
		// Criar um arquivo mock
		mockData := []byte("fake image data")
		file := newMockFile(mockData)

		// A validação é feita internamente no UploadImage
		// Aqui testamos apenas a extensão
		ctx := context.Background()

		// Criar serviço mock (vai falhar se CLOUDINARY_URL não estiver configurada)
		if os.Getenv("CLOUDINARY_URL") != "" {
			service, err := storage.NewCloudinaryService()
			if err != nil {
				t.Skipf("Pulando teste - CLOUDINARY_URL não configurada: %v", err)
				return
			}

			params := storage.UploadImageParams{
				File:     file,
				FileName: filename,
				Folder:   "test",
			}

			// Este teste vai falhar no upload, mas não na validação do arquivo
			_, err = service.UploadImage(ctx, params)
			// Esperamos que falhe mas não por causa da extensão
			if err != nil && err.Error() == "formato de arquivo não suportado" {
				t.Errorf("arquivo válido foi rejeitado: %s", filename)
			}
		}
	}
}

func TestValidateImageFile_InvalidExtensions(t *testing.T) {
	invalidFiles := []string{
		"document.pdf",
		"archive.zip",
		"video.mp4",
		"script.js",
		"style.css",
	}

	for _, filename := range invalidFiles {
		mockData := []byte("fake file data")
		file := newMockFile(mockData)

		ctx := context.Background()

		if os.Getenv("CLOUDINARY_URL") != "" {
			service, err := storage.NewCloudinaryService()
			if err != nil {
				t.Skipf("Pulando teste - CLOUDINARY_URL não configurada: %v", err)
				return
			}

			params := storage.UploadImageParams{
				File:     file,
				FileName: filename,
				Folder:   "test",
			}

			_, err = service.UploadImage(ctx, params)
			if err == nil {
				t.Errorf("arquivo inválido foi aceito: %s", filename)
			}
		}
	}
}

func TestUploadImage_EmptyFile(t *testing.T) {
	if os.Getenv("CLOUDINARY_URL") == "" {
		t.Skip("CLOUDINARY_URL não configurada - pulando teste de integração")
	}

	service, err := storage.NewCloudinaryService()
	if err != nil {
		t.Fatalf("erro ao criar serviço: %v", err)
	}

	// Arquivo vazio
	emptyFile := newMockFile([]byte{})

	params := storage.UploadImageParams{
		File:     emptyFile,
		FileName: "empty.jpg",
		Folder:   "test",
	}

	_, err = service.UploadImage(context.Background(), params)
	if err == nil {
		t.Error("esperava erro ao fazer upload de arquivo vazio")
	}

	if err.Error() != "arquivo vazio - nenhum byte foi lido" {
		t.Errorf("mensagem de erro incorreta: %v", err)
	}
}

func TestDeleteImage_EmptyPublicID(t *testing.T) {
	if os.Getenv("CLOUDINARY_URL") == "" {
		t.Skip("CLOUDINARY_URL não configurada - pulando teste de integração")
	}

	service, err := storage.NewCloudinaryService()
	if err != nil {
		t.Fatalf("erro ao criar serviço: %v", err)
	}

	err = service.DeleteImage(context.Background(), "")
	if err == nil {
		t.Error("esperava erro ao deletar imagem com publicID vazio")
	}

	if err.Error() != "publicID não pode ser vazio" {
		t.Errorf("mensagem de erro incorreta: %v", err)
	}
}

func TestGetOptimizedURL_EmptyPublicID(t *testing.T) {
	if os.Getenv("CLOUDINARY_URL") == "" {
		t.Skip("CLOUDINARY_URL não configurada - pulando teste de integração")
	}

	service, err := storage.NewCloudinaryService()
	if err != nil {
		t.Fatalf("erro ao criar serviço: %v", err)
	}

	_, err = service.GetOptimizedURL("", 300, 300, "auto")
	if err == nil {
		t.Error("esperava erro ao gerar URL com publicID vazio")
	}

	if err.Error() != "publicID não pode ser vazio" {
		t.Errorf("mensagem de erro incorreta: %v", err)
	}
}

func TestGetOptimizedURL_ValidParams(t *testing.T) {
	if os.Getenv("CLOUDINARY_URL") == "" {
		t.Skip("CLOUDINARY_URL não configurada - pulando teste de integração")
	}

	service, err := storage.NewCloudinaryService()
	if err != nil {
		t.Fatalf("erro ao criar serviço: %v", err)
	}

	url, err := service.GetOptimizedURL("test/image_123", 600, 400, "auto")
	if err != nil {
		t.Fatalf("erro ao gerar URL otimizada: %v", err)
	}

	if url == "" {
		t.Error("URL otimizada não deve ser vazia")
	}

	// Verificar se a URL contém os parâmetros esperados
	expectedParts := []string{
		"res.cloudinary.com",
		"w_600",
		"h_400",
		"q_auto",
		"f_auto",
		"c_fill",
	}

	for _, part := range expectedParts {
		if !contains(url, part) {
			t.Errorf("URL não contém parte esperada: %s\nURL: %s", part, url)
		}
	}
}

// contains verifica se uma string contém uma substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
