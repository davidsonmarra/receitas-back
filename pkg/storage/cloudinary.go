package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// ImageService interface para serviços de imagem (permite mocking)
type ImageService interface {
	UploadImage(ctx context.Context, params UploadImageParams) (*UploadResult, error)
	DeleteImage(ctx context.Context, publicID string) error
	GetOptimizedURL(publicID string, width, height int, quality string) (string, error)
	GetImageVariants(publicID string) map[string]string
}

// ServiceFactory é uma função que cria um ImageService
// Pode ser substituída nos testes para retornar um mock
var ServiceFactory func() (ImageService, error) = func() (ImageService, error) {
	return NewCloudinaryService()
}

// CloudinaryService gerencia upload de imagens para o Cloudinary
type CloudinaryService struct {
	cld *cloudinary.Cloudinary
}

// NewCloudinaryService cria uma nova instância do serviço
func NewCloudinaryService() (*CloudinaryService, error) {
	// Cloudinary URL format: cloudinary://API_KEY:API_SECRET@CLOUD_NAME
	cloudinaryURL := os.Getenv("CLOUDINARY_URL")
	if cloudinaryURL == "" {
		return nil, fmt.Errorf("CLOUDINARY_URL não configurado")
	}

	cld, err := cloudinary.NewFromURL(cloudinaryURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao inicializar Cloudinary: %w", err)
	}

	// Verificar se o cloud name foi extraído corretamente
	if cld.Config.Cloud.CloudName == "" {
		return nil, fmt.Errorf("CloudName vazio - verifique formato da CLOUDINARY_URL (deve ser: cloudinary://API_KEY:API_SECRET@CLOUD_NAME)")
	}

	return &CloudinaryService{cld: cld}, nil
}

// UploadImageParams parâmetros para upload de imagem
type UploadImageParams struct {
	File      multipart.File
	FileName  string
	Folder    string // ex: "recipes", "ingredients"
	PublicID  string // ID único (opcional, será gerado se vazio)
	MaxWidth  int    // largura máxima (0 = sem limite)
	MaxHeight int    // altura máxima (0 = sem limite)
}

// UploadResult resultado do upload
type UploadResult struct {
	PublicID  string `json:"public_id"`
	URL       string `json:"url"`
	SecureURL string `json:"secure_url"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Format    string `json:"format"`
	Bytes     int    `json:"bytes"`
}

// UploadImage faz upload de uma imagem para o Cloudinary
func (s *CloudinaryService) UploadImage(ctx context.Context, params UploadImageParams) (*UploadResult, error) {
	if s.cld == nil {
		return nil, fmt.Errorf("serviço cloudinary não inicializado")
	}

	// Validar tipo de arquivo
	if err := validateImageFile(params.FileName); err != nil {
		return nil, err
	}

	// IMPORTANTE: Voltar o cursor do arquivo para o início
	// O arquivo pode ter sido lido antes e estar no final
	if seeker, ok := params.File.(io.Seeker); ok {
		if _, err := seeker.Seek(0, 0); err != nil {
			return nil, fmt.Errorf("erro ao posicionar cursor do arquivo: %w", err)
		}
	}

	// Ler o arquivo para um buffer
	// Isso garante que o conteúdo está completamente carregado
	fileBytes, err := io.ReadAll(params.File)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo: %w", err)
	}

	if len(fileBytes) == 0 {
		return nil, fmt.Errorf("arquivo vazio - nenhum byte foi lido")
	}

	// Gerar PublicID único se não fornecido
	publicID := params.PublicID
	if publicID == "" {
		publicID = generatePublicID(params.FileName)
	}

	// Não adicionar folder ao PublicID se já está configurado no Folder
	// O Cloudinary cuida disso automaticamente

	// Configurar parâmetros de upload seguindo a documentação oficial
	// https://cloudinary.com/documentation/go_quick_start
	uploadParams := uploader.UploadParams{
		PublicID:       publicID,
		Folder:         params.Folder,
		ResourceType:   "image",
		Overwrite:      api.Bool(true),  // Usar api.Bool conforme documentação
		UniqueFilename: api.Bool(false), // Usar api.Bool conforme documentação
	}

	// Adicionar transformações se especificadas
	if params.MaxWidth > 0 || params.MaxHeight > 0 {
		uploadParams.Transformation = buildTransformation(params.MaxWidth, params.MaxHeight)
	}

	// Converter bytes para io.Reader (formato aceito pelo Cloudinary)
	fileReader := bytes.NewReader(fileBytes)

	// Fazer upload usando o Reader
	result, err := s.cld.Upload.Upload(ctx, fileReader, uploadParams)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer upload: %w", err)
	}

	if result == nil {
		return nil, fmt.Errorf("cloudinary retornou resultado nulo")
	}

	if result.PublicID == "" || result.SecureURL == "" {
		return nil, fmt.Errorf("cloudinary retornou dados vazios - PublicID: '%s', SecureURL: '%s' - verifique CLOUDINARY_URL", result.PublicID, result.SecureURL)
	}

	return &UploadResult{
		PublicID:  result.PublicID,
		URL:       result.URL,
		SecureURL: result.SecureURL,
		Width:     result.Width,
		Height:    result.Height,
		Format:    result.Format,
		Bytes:     result.Bytes,
	}, nil
}

// DeleteImage deleta uma imagem do Cloudinary
func (s *CloudinaryService) DeleteImage(ctx context.Context, publicID string) error {
	if s.cld == nil {
		return fmt.Errorf("serviço cloudinary não inicializado")
	}

	if publicID == "" {
		return fmt.Errorf("publicID não pode ser vazio")
	}

	_, err := s.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: "image",
	})
	if err != nil {
		return fmt.Errorf("erro ao deletar imagem: %w", err)
	}

	return nil
}

// GetOptimizedURL retorna URL otimizada da imagem com transformações
func (s *CloudinaryService) GetOptimizedURL(publicID string, width, height int, quality string) (string, error) {
	if s.cld == nil {
		return "", fmt.Errorf("serviço cloudinary não inicializado")
	}

	if publicID == "" {
		return "", fmt.Errorf("publicID não pode ser vazio")
	}

	// Validar parâmetros
	if width <= 0 || height <= 0 {
		return "", fmt.Errorf("largura e altura devem ser maiores que zero")
	}

	if width > 5000 || height > 5000 {
		return "", fmt.Errorf("dimensões muito grandes (máximo: 5000x5000)")
	}

	// Construir URL com transformações
	transformation := fmt.Sprintf("w_%d,h_%d,c_fill,q_%s,f_auto", width, height, quality)
	url := fmt.Sprintf("https://res.cloudinary.com/%s/image/upload/%s/%s",
		s.cld.Config.Cloud.CloudName,
		transformation,
		publicID,
	)

	return url, nil
}

// GetImageVariants retorna URLs otimizadas em diferentes tamanhos
func (s *CloudinaryService) GetImageVariants(publicID string) map[string]string {
	if publicID == "" {
		return map[string]string{}
	}

	cloudName := s.cld.Config.Cloud.CloudName
	variants := make(map[string]string)

	// Thumbnail (pequeno para listagem)
	variants["thumbnail"] = fmt.Sprintf("https://res.cloudinary.com/%s/image/upload/w_150,h_150,c_fill,q_auto,f_auto/%s", cloudName, publicID)
	
	// Small
	variants["small"] = fmt.Sprintf("https://res.cloudinary.com/%s/image/upload/w_400,h_400,c_fit,q_auto,f_auto/%s", cloudName, publicID)
	
	// Medium
	variants["medium"] = fmt.Sprintf("https://res.cloudinary.com/%s/image/upload/w_800,h_800,c_fit,q_auto,f_auto/%s", cloudName, publicID)
	
	// Large
	variants["large"] = fmt.Sprintf("https://res.cloudinary.com/%s/image/upload/w_1200,h_1200,c_fit,q_auto,f_auto/%s", cloudName, publicID)
	
	// Original
	variants["original"] = fmt.Sprintf("https://res.cloudinary.com/%s/image/upload/%s", cloudName, publicID)

	return variants
}

// buildTransformation constrói string de transformação para Cloudinary
func buildTransformation(maxWidth, maxHeight int) string {
	if maxWidth == 0 && maxHeight == 0 {
		return "q_auto,f_auto" // auto quality e formato
	}

	parts := []string{"c_limit", "q_auto", "f_auto"}
	if maxWidth > 0 {
		parts = append(parts, fmt.Sprintf("w_%d", maxWidth))
	}
	if maxHeight > 0 {
		parts = append(parts, fmt.Sprintf("h_%d", maxHeight))
	}

	return strings.Join(parts, ",")
}

// validateImageFile valida se o arquivo é uma imagem válida
func validateImageFile(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	validExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".bmp":  true,
	}

	if !validExtensions[ext] {
		return fmt.Errorf("formato de arquivo não suportado: %s (permitidos: jpg, jpeg, png, gif, webp, bmp)", ext)
	}

	return nil
}

// generatePublicID gera um ID público único para a imagem
func generatePublicID(filename string) string {
	// Remover extensão
	name := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Limpar nome (remover caracteres especiais)
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, name)

	// Adicionar timestamp para garantir unicidade
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s_%d", name, timestamp)
}

// ValidateImageSize valida o tamanho do arquivo
func ValidateImageSize(file io.Reader, maxSizeMB int) error {
	// Ler primeiro para contar bytes
	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo: %w", err)
	}

	sizeMB := len(data) / (1024 * 1024)
	if sizeMB > maxSizeMB {
		return fmt.Errorf("imagem muito grande: %dMB (máximo: %dMB)", sizeMB, maxSizeMB)
	}

	return nil
}
