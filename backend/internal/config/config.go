package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr    string
	CORSOrigins []string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	NWSUserAgent  string
	USNOAppID     string
	WeatherAPIKey string
	OWMAPIKey     string
	AISStreamKey  string
}

func Load() (*Config, error) {
	// Best-effort load .env from cwd and parent
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	c := &Config{
		HTTPAddr:      env("HTTP_ADDR", ":8080"),
		CORSOrigins:   splitCSV(env("CORS_ORIGINS", "http://localhost:5173")),
		DBHost:        env("DB_HOST", "127.0.0.1"),
		DBPort:        env("DB_PORT", "3306"),
		DBUser:        env("DB_USER", "fishingapp"),
		DBPassword:    env("DB_PASSWORD", "fishingapp"),
		DBName:        env("DB_NAME", "fishingapp"),
		NWSUserAgent:  env("NWS_USER_AGENT", "(puget-sound-fishing-app, contact@example.com)"),
		USNOAppID:     env("USNO_APP_ID", "psfish"),
		WeatherAPIKey: os.Getenv("WEATHERAPI_KEY"),
		OWMAPIKey:     os.Getenv("OWM_API_KEY"),
		AISStreamKey:  os.Getenv("AISSTREAM_API_KEY"),
	}
	return c, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true&charset=utf8mb4",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
