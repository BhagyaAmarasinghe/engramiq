package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	LLM         LLMConfig
	Storage     StorageConfig
	Search      SearchConfig
}

type ServerConfig struct {
	Port        string
	CORSOrigins string
}

type DatabaseConfig struct {
	URL            string
	MaxConnections int
	MaxIdleTime    time.Duration
}

type RedisConfig struct {
	URL         string
	PoolSize    int
	DialTimeout time.Duration
}

type JWTConfig struct {
	Secret           string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
}

type LLMConfig struct {
	Provider     string
	APIKey       string
	Model        string
	Temperature  float64
	MaxTokens    int
	Timeout      time.Duration
	StripPII     bool
}

type StorageConfig struct {
	Provider      string
	Endpoint      string
	AccessKey     string
	SecretKey     string
	BucketName    string
	UseSSL        bool
}

type SearchConfig struct {
	ElasticsearchURL string
	Index           string
}

func Load() *Config {
	return &Config{
		Environment: getEnvOrDefault("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Port:        getEnvOrDefault("PORT", "8080"),
			CORSOrigins: getEnvOrDefault("CORS_ORIGINS", "http://localhost:3000"),
		},
		Database: DatabaseConfig{
			URL:            getEnvOrDefault("DATABASE_URL", "postgresql://user:pass@localhost:5432/engramiq?sslmode=disable"),
			MaxConnections: getEnvAsInt("DB_MAX_CONNECTIONS", 25),
			MaxIdleTime:    getEnvAsDuration("DB_MAX_IDLE_TIME", "15m"),
		},
		Redis: RedisConfig{
			URL:         getEnvOrDefault("REDIS_URL", "redis://localhost:6379"),
			PoolSize:    getEnvAsInt("REDIS_POOL_SIZE", 10),
			DialTimeout: getEnvAsDuration("REDIS_DIAL_TIMEOUT", "5s"),
		},
		JWT: JWTConfig{
			Secret:           getEnvOrDefault("JWT_SECRET", "your-secret-key-change-in-production"),
			AccessTokenTTL:   getEnvAsDuration("JWT_ACCESS_TTL", "15m"),
			RefreshTokenTTL:  getEnvAsDuration("JWT_REFRESH_TTL", "168h"), // 7 days
		},
		LLM: LLMConfig{
			Provider:     getEnvOrDefault("LLM_PROVIDER", "openai"),
			APIKey:       os.Getenv("OPENAI_API_KEY"),
			Model:        getEnvOrDefault("LLM_MODEL", "gpt-4-turbo-preview"),
			Temperature:  getEnvAsFloat("LLM_TEMPERATURE", 0.3),
			MaxTokens:    getEnvAsInt("LLM_MAX_TOKENS", 2000),
			Timeout:      getEnvAsDuration("LLM_TIMEOUT", "60s"),
			StripPII:     getEnvAsBool("LLM_STRIP_PII", true),
		},
		Storage: StorageConfig{
			Provider:      getEnvOrDefault("STORAGE_PROVIDER", "minio"),
			Endpoint:      getEnvOrDefault("STORAGE_ENDPOINT", "localhost:9000"),
			AccessKey:     getEnvOrDefault("STORAGE_ACCESS_KEY", "minioadmin"),
			SecretKey:     getEnvOrDefault("STORAGE_SECRET_KEY", "minioadmin"),
			BucketName:    getEnvOrDefault("STORAGE_BUCKET", "engramiq"),
			UseSSL:        getEnvAsBool("STORAGE_USE_SSL", false),
		},
		Search: SearchConfig{
			ElasticsearchURL: getEnvOrDefault("ELASTICSEARCH_URL", "http://localhost:9200"),
			Index:           getEnvOrDefault("ELASTICSEARCH_INDEX", "engramiq"),
		},
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := strings.ToLower(os.Getenv(key))
	if valueStr == "" {
		return defaultValue
	}
	return valueStr == "true" || valueStr == "yes" || valueStr == "1"
}

func getEnvAsDuration(key string, defaultValue string) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		valueStr = defaultValue
	}
	duration, err := time.ParseDuration(valueStr)
	if err != nil {
		d, _ := time.ParseDuration(defaultValue)
		return d
	}
	return duration
}