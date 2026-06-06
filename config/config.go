package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server  ServerConfig
	Redis   RedisConfig
	Gemini  GeminiConfig
	Groq    GroqConfig
	Auth    AuthConfig
	Crypto  CryptoConfig
	Agent   AgentConfig
}

type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type GeminiConfig struct {
	APIKey          string
	Model           string
	MaxOutputTokens int32
	Temperature     float32
}

type GroqConfig struct {
	APIKey          string
	Model           string
	MaxOutputTokens int
	Temperature     float32
}

type AuthConfig struct {
	JWTSecret      string
	TokenExpiry    time.Duration
	MaxSessions    int
}

type CryptoConfig struct {
	AESKey string // 32-byte hex-encoded key for AES-256
}

type AgentConfig struct {
	FreePoolSize int
	ProPoolSize  int
	// Per-state timeouts (from spec)
	InterviewTimeout    time.Duration // 8–15 min window enforced at orchestrator
	GraphBuildTimeout   time.Duration // < 30s
	LogTimeout          time.Duration // 3–5 min
	GraphUpdateTimeout  time.Duration // < 10s
	MoEGatingTimeout    time.Duration // < 5s
	AgentsRunTimeout    time.Duration // < 15s
	VetoResolveTimeout  time.Duration // < 5s
	PlanSynthTimeout    time.Duration // < 10s
	CoachWriteTimeout   time.Duration // < 15s
	GeminiCallTimeout   time.Duration // < 5s (context.WithTimeout on every call)
}

func Load() Config {
	return Config{
		Server: ServerConfig{
			Port:            getEnv("PORT", "8080"),
			ReadTimeout:     15 * time.Second,
			WriteTimeout:    15 * time.Second,
			ShutdownTimeout: 30 * time.Second,
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		Gemini: GeminiConfig{
			APIKey:          getEnv("GEMINI_API_KEY", ""),
			Model:           getEnv("GEMINI_MODEL", "gemini-2.5-flash"),
			MaxOutputTokens: int32(getEnvInt("GEMINI_MAX_TOKENS", 8192)),
			Temperature:     float32(getEnvFloat("GEMINI_TEMPERATURE", 0.7)),
		},
		Groq: GroqConfig{
			APIKey:          getEnv("GROQ_API_KEY", ""),
			Model:           getEnv("GROQ_MODEL", "llama-3.3-70b-versatile"),
			MaxOutputTokens: getEnvInt("GROQ_MAX_TOKENS", 8192),
			Temperature:     float32(getEnvFloat("GROQ_TEMPERATURE", 0.7)),
		},
		Auth: AuthConfig{
			JWTSecret:   mustEnv("JWT_SECRET"),
			TokenExpiry: 24 * time.Hour,
			MaxSessions: 3,
		},
		Crypto: CryptoConfig{
			AESKey: mustEnv("AES_KEY"),
		},
		Agent: AgentConfig{
			FreePoolSize:       1,
			ProPoolSize:        3,
			GraphBuildTimeout:  30 * time.Second,
			LogTimeout:         5 * time.Minute,
			GraphUpdateTimeout: 10 * time.Second,
			MoEGatingTimeout:   5 * time.Second,
			AgentsRunTimeout:   15 * time.Second,
			VetoResolveTimeout: 5 * time.Second,
			PlanSynthTimeout:   10 * time.Second,
			CoachWriteTimeout:  15 * time.Second,
			GeminiCallTimeout:  5 * time.Second,
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("required env var not set: " + key)
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}
