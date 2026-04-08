package config

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	HTTPListenAddr   string
	MediaMTXHTTPURL  string
	AllowedStreamKey string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPListenAddr:   getEnv("HTTP_LISTEN_ADDR", ":8080"),
		MediaMTXHTTPURL:  getEnv("MEDIAMTX_HTTP_URL", "http://localhost:8889"),
		AllowedStreamKey: os.Getenv("STREAM_KEY"),
	}

	if cfg.AllowedStreamKey == "" {
		return Config{}, errors.New("STREAM_KEY is required")
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
