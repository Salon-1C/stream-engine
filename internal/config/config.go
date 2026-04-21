package config

import (
	"fmt"
	"os"
)

type Config struct {
	HTTPListenAddr  string
	MediaMTXHTTPURL string
	// AllowedStreamKey is optional. When empty, any /live/<key> path is accepted.
	AllowedStreamKey string
	RabbitMQURL      string
	RabbitMQQueue    string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPListenAddr:   getEnv("HTTP_LISTEN_ADDR", ":8080"),
		MediaMTXHTTPURL:  getEnv("MEDIAMTX_HTTP_URL", "http://localhost:8889"),
		AllowedStreamKey: os.Getenv("STREAM_KEY"),
		RabbitMQURL:      os.Getenv("RABBITMQ_URL"),
		RabbitMQQueue:    getEnv("RABBITMQ_QUEUE", "recordings.ready"),
	}

	if cfg.MediaMTXHTTPURL == "" {
		return Config{}, fmt.Errorf("MEDIAMTX_HTTP_URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
