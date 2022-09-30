package config

import (
	"fmt"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type APIConfig struct {
	Debug            bool          `envconfig:"DEBUG" default:"false"`
	Host             string        `envconfig:"HOST" default:"0.0.0.0"`
	GrpcPort         string        `envconfig:"GRPC_PORT" default:"50051"`
	RestPort         string        `envconfig:"REST_PORT" default:"8000"`
	ProductsCacheTtl time.Duration `envconfig:"PRODUCTS_CACHE_TTL" default:"1m"`
}

type DatabaseConfig struct {
	DSN  string `envconfig:"DATABASE_DSN" required:"true"`
	Name string `envconfig:"DATABASE_NAME" required:"true"`
}

type TracingConfig struct {
	ExporterAddress string `envconfig:"TRACING_EXPORTER_ADDRESS"`
	ExporterPort    string `envconfig:"TRACING_EXPORTER_PORT"`
}

type SentryConfig struct {
	DSN string `envconfig:"SENTRY_DSN"`
	Env string `envconfig:"SENTRY_ENV" default:"dev"`
}

type Config struct {
	API      *APIConfig
	Database *DatabaseConfig
	Tracing  *TracingConfig
	Sentry   *SentryConfig
}

var (
	once   sync.Once
	config *Config
)

// GetConfig Загружает конфиг из .env файла и возвращает объект конфигурации
// В случае, если не передать параметр `envfiles`, берется `.env` файл из корня проекта.
func GetConfig(envfiles ...string) (*Config, error) {
	var err error
	once.Do(func() {
		_ = godotenv.Load(envfiles...)

		var c Config
		err = envconfig.Process("", &c)
		if err != nil {
			err = fmt.Errorf("error parse config from env variables: %w", err)
			return
		}
		config = &c
	})

	return config, err
}
