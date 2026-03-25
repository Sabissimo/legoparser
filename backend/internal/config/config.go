package config

import (
	"os"
	"time"
)

type Config struct {
	Port           string
	DatabaseURL    string
	ChromeWSURL    string
	CronSchedule   string
	WoltLat        string
	WoltLon        string
	RateLimitDelay time.Duration
	MigrationsPath string
}

func Load() *Config {
	cfg := &Config{
		Port:           getEnv("PORT", "8081"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://parser:parserpass@localhost:5433/lego_parser?sslmode=disable"),
		ChromeWSURL:    getEnv("CHROME_WS_URL", "ws://localhost:9222"),
		CronSchedule:   getEnv("CRON_SCHEDULE", "0 0 6 * * *"),
		WoltLat:        getEnv("WOLT_LAT", "41.7151"),
		WoltLon:        getEnv("WOLT_LON", "44.8271"),
		MigrationsPath: getEnv("MIGRATIONS_PATH", "migrations"),
	}

	delay, err := time.ParseDuration(getEnv("RATE_LIMIT_DELAY", "500ms"))
	if err != nil {
		delay = 500 * time.Millisecond
	}
	cfg.RateLimitDelay = delay

	return cfg
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
