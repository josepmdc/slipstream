package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	AcestreamTimeout  time.Duration `env:"ACESTREAM_TIMEOUT"  envDefault:"30s"`
	AcestreamBaseURL  string        `env:"ACESTREAM_BASE_URL" envDefault:"http://127.0.0.1:6878"`
	AcestreamEndpoint string        `env:"ACESTREAM_ENDPOINT" envDefault:"/ace/manifest.m3u8"`
	Host              string        `env:"HOST"               envDefault:"localhost"`
	Port              int           `env:"PORT"               envDefault:"8080"`
	PublicBaseURL     string        `env:"PUBLIC_BASE_URL"    envDefault:"http://localhost:8080"`
}

func Load() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}
	return &cfg, nil
}
