package config

import (
	"os"
)

type Config struct {
	ServerPort           string
	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	RedisAddr            string
	RedisPassword        string
	JWTSecret            string
	OpenAIAPIKey         string
	SQLitePath           string
	UploadDir            string
	DefaultAdminEmail    string
	DefaultAdminPassword string
	FrontendBaseURL      string
	PayPalClientID       string
	PayPalSecret         string
	PayPalBaseURL        string
}

func Load() *Config {
	return &Config{
		ServerPort:           getEnvAny([]string{"SERVER_PORT", "PORT"}, "8080"),
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5432"),
		DBUser:               getEnv("DB_USER", "postgres"),
		DBPassword:           getEnv("DB_PASSWORD", "postgres"),
		DBName:               getEnv("DB_NAME", "chinese_learning"),
		RedisAddr:            getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:        getEnv("REDIS_PASSWORD", ""),
		JWTSecret:            getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		OpenAIAPIKey:         getEnv("OPENAI_API_KEY", ""),
		SQLitePath:           getEnv("SQLITE_PATH", "chinese_learning.db"),
		UploadDir:            getEnv("UPLOAD_DIR", "./uploads"),
		DefaultAdminEmail:    getEnv("DEFAULT_ADMIN_EMAIL", "admin@chineseapp.com"),
		DefaultAdminPassword: getEnv("DEFAULT_ADMIN_PASSWORD", "Admin123456"),
		FrontendBaseURL:      getEnv("FRONTEND_BASE_URL", "https://chinese-learning-app-jade.vercel.app"),
		PayPalClientID:       getEnv("PAYPAL_CLIENT_ID", ""),
		PayPalSecret:         getEnv("PAYPAL_SECRET", ""),
		PayPalBaseURL:        getEnv("PAYPAL_BASE_URL", "https://api-m.sandbox.paypal.com"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAny(keys []string, fallback string) string {
	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok && value != "" {
			return value
		}
	}
	return fallback
}
