package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppEnv   string
	LogLevel string

	DatabaseURL string

	WhatsAppDBPath string

	StorageDriver    string
	StorageLocalDir  string
	StorageEndpoint  string
	StorageAccessKey string
	StorageSecretKey string
	StorageBucket    string
	StorageUseSSL    bool

	GeminiAPIKey  string
	GeminiModel   string
	GeminiModelTx string

	LLMDailyBudget     int
	LLMMaxOutputTokens int32

	MaxAudioSizeBytes int64
	MaxImageSizeBytes int64
	MaxPDFSizeBytes   int64

	DraftExpiryMinutes int

	ConfidenceAutoSave   float64
	ConfidenceAskConfirm float64
}

func Load() (*Config, error) {
	cfg := &Config{}

	cfg.AppEnv = getEnv("APP_ENV", "development")
	cfg.LogLevel = getEnv("LOG_LEVEL", "info")

	cfg.DatabaseURL = getEnv("DATABASE_URL", "")
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	cfg.WhatsAppDBPath = getEnv("WHATSAPP_DB_PATH", "./data/whatsapp.db")

	cfg.StorageDriver = getEnv("STORAGE_DRIVER", "local")
	cfg.StorageLocalDir = getEnv("STORAGE_LOCAL_DIR", "./data/media")
	cfg.StorageEndpoint = getEnv("STORAGE_ENDPOINT", "localhost:9000")
	cfg.StorageAccessKey = getEnv("STORAGE_ACCESS_KEY", "minioadmin")
	cfg.StorageSecretKey = getEnv("STORAGE_SECRET_KEY", "minioadmin")
	cfg.StorageBucket = getEnv("STORAGE_BUCKET", "gonami")
	cfg.StorageUseSSL = getEnvBool("STORAGE_USE_SSL", false)

	cfg.GeminiAPIKey = getEnv("GEMINI_API_KEY", "")
	cfg.GeminiModel = getEnv("GEMINI_MODEL", "gemini-2.0-flash")
	cfg.GeminiModelTx = getEnv("GEMINI_MODEL_TEXT", "")

	cfg.LLMDailyBudget = getEnvInt("LLM_DAILY_BUDGET", 300)
	cfg.LLMMaxOutputTokens = int32(getEnvInt("LLM_MAX_OUTPUT_TOKENS", 150))

	maxAudioMB := getEnvInt("MAX_AUDIO_SIZE_MB", 10)
	maxImageMB := getEnvInt("MAX_IMAGE_SIZE_MB", 10)
	maxPDFMB := getEnvInt("MAX_PDF_SIZE_MB", 20)
	cfg.MaxAudioSizeBytes = int64(maxAudioMB) * 1024 * 1024
	cfg.MaxImageSizeBytes = int64(maxImageMB) * 1024 * 1024
	cfg.MaxPDFSizeBytes = int64(maxPDFMB) * 1024 * 1024

	cfg.DraftExpiryMinutes = getEnvInt("DRAFT_EXPIRY_MINUTES", 15)

	cfg.ConfidenceAutoSave = getEnvFloat("CONFIDENCE_AUTO_SAVE", 0.80)
	cfg.ConfidenceAskConfirm = getEnvFloat("CONFIDENCE_ASK_CONFIRM", 0.50)

	return cfg, nil
}

func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
}

func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return defaultVal
	}
	return b
}

func getEnvInt(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return i
}

func getEnvFloat(key string, defaultVal float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return defaultVal
	}
	return f
}
