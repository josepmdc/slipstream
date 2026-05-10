package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	AceStreamTimeout       time.Duration `env:"ACESTREAM_TIMEOUT"        envDefault:"30s"`
	AceStreamBaseURL       string        `env:"ACESTREAM_BASE_URL"       envDefault:"http://127.0.0.1:6878"`
	AceStreamEndpoint      string        `env:"ACESTREAM_ENDPOINT"       envDefault:"/ace/manifest.m3u8"`
	Host                   string        `env:"HOST"                     envDefault:"localhost"`
	Port                   int           `env:"PORT"                     envDefault:"8080"`
	PublicBaseURL          string        `env:"PUBLIC_BASE_URL"          envDefault:"http://localhost:8080"`
	SegmentCacheMaxSize    int           `env:"SEGMENT_CACHE_MAX_SIZE"   envDefault:"10_000"`
	SegmentCacheExpiration time.Duration `env:"SEGMENT_CACHE_EXPIRATION" envDefault:"60s"`
}

func Load() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}
	return &cfg, nil
}
