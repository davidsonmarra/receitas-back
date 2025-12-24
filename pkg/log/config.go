package log

// Config representa a configuração do logger
type Config struct {
	// Level define o nível mínimo de log: "debug", "info", "warn", "error"
	// Padrão: "info"
	Level string

	// Development define se o logger está em modo desenvolvimento
	// true = logs formatados e coloridos (legível para humanos)
	// false = logs em JSON estruturado (ideal para produção)
	Development bool
}

// NewConfig cria uma configuração padrão para o logger
func NewConfig() Config {
	return Config{
		Level:       "info",
		Development: false,
	}
}
