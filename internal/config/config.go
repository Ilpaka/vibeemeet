package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	LiveKit     LiveKitConfig
	Log         LogConfig
}

type ServerConfig struct {
	Port         int
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	DSN             string
	MaxConnections  int
	MaxIdleTime     time.Duration
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	Issuer        string
}

type LiveKitConfig struct {
	URL         string // Внутренний URL для бэкенда
	FrontendURL string // Публичный URL для фронтенда
	APIKey      string
	APISecret   string
	HostIP      string // IP адрес хоста для локальной сети
	Port        string // Порт LiveKit для клиентов (внешний)
}

type LogConfig struct {
	Level string
}

func Load() (*Config, error) {
	// Загрузка .env файла (если существует)
	_ = godotenv.Load()

	cfg := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
		},
		Database: DatabaseConfig{
			DSN:             getEnv("DATABASE_DSN", "postgres://appuser:apppass123@localhost:5432/app_database?sslmode=disable"),
			MaxConnections:  getEnvAsInt("DATABASE_MAX_CONNECTIONS", 25),
			MaxIdleTime:     getEnvAsDuration("DATABASE_MAX_IDLE_TIME", 5*time.Minute),
			ConnMaxLifetime: getEnvAsDuration("DATABASE_CONN_MAX_LIFETIME", 1*time.Hour),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			AccessSecret:  getEnv("JWT_ACCESS_SECRET", "your-access-secret-key-change-in-production"),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-key-change-in-production"),
			AccessTTL:     getEnvAsDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL:    getEnvAsDuration("JWT_REFRESH_TTL", 7*24*time.Hour),
			Issuer:        getEnv("JWT_ISSUER", "video-conference"),
		},
		LiveKit: LiveKitConfig{
			URL:         getEnv("LIVEKIT_URL", "ws://localhost:7880"),
			FrontendURL: getEnv("LIVEKIT_FRONTEND_URL", ""),
			APIKey:      getEnv("LIVEKIT_API_KEY", "devkey"),
			APISecret:   getEnv("LIVEKIT_API_SECRET", "secret"),
			HostIP:      getEnv("HOST_IP", GetLocalIP()),
			Port:        getEnv("LIVEKIT_PORT", "7880"),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.JWT.AccessSecret == "" || c.JWT.RefreshSecret == "" {
		return fmt.Errorf("JWT secrets must be set")
	}
	if c.Database.DSN == "" {
		return fmt.Errorf("database DSN must be set")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// GetLocalIP возвращает первый не-localhost IPv4 адрес машины
func GetLocalIP() string {
	// Сначала проверяем переменную окружения
	if ip := os.Getenv("HOST_IP"); ip != "" {
		return ip
	}

	// Пытаемся определить IP автоматически
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}

	// Список приоритетных сетей для локальной сети
	priorityPrefixes := []string{"192.168.", "10.", "172."}

	var fallbackIP string

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip := ipnet.IP.String()

				// Проверяем приоритетные сети
				for _, prefix := range priorityPrefixes {
					if strings.HasPrefix(ip, prefix) {
						return ip
					}
				}

				// Сохраняем как fallback
				if fallbackIP == "" {
					fallbackIP = ip
				}
			}
		}
	}

	if fallbackIP != "" {
		return fallbackIP
	}

	return "localhost"
}
