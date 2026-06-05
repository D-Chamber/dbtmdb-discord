package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DiscordToken  string
	TMDBToken     string // API Read Access Token from TMDB (used as Bearer token)
	CacheDuration time.Duration
	LogLevel      string
	MaxResults    int
}

func Load() *Config {
	discordToken := os.Getenv("DISCORD_TOKEN")

	// TMDB_ACCESS_TOKEN is the preferred env var name (API Read Access Token / Bearer token).
	// TMDB_API_KEY is accepted as a fallback for backward compatibility.
	tmdbToken := os.Getenv("TMDB_ACCESS_TOKEN")
	if tmdbToken == "" {
		tmdbToken = os.Getenv("TMDB_API_KEY")
	}

	if discordToken == "" || tmdbToken == "" {
		log.Fatal("DISCORD_TOKEN and TMDB_ACCESS_TOKEN (or TMDB_API_KEY) must be set in environment or .env file")
	}

	cacheMinutes := getEnvInt("CACHE_DURATION_MINUTES", 15)
	return &Config{
		DiscordToken:  discordToken,
		TMDBToken:     tmdbToken,
		CacheDuration: time.Duration(cacheMinutes) * time.Minute,
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		MaxResults:    getEnvInt("MAX_RESULTS", 5),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
