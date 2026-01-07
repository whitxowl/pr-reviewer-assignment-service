package config

import (
	"fmt"
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	Env           string `env:"ENV" env-default:"local"`
	HTTPServer    `env-prefix:"HTTP_"`
	StorageConfig `env-prefix:"DB_"`
}

type HTTPServer struct {
	Address     string        `env:"ADDRESS" env-default:"localhost:8080"`
	Timeout     time.Duration `env:"TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `env:"IDLE_TIMEOUT" env-default:"60s"`
}

type StorageConfig struct {
	Host           string `env:"HOST" env-required:"true"`
	Port           string `env:"PORT" env-required:"true"`
	User           string `env:"USER" env-required:"true"`
	Password       string `env:"PASSWORD" env-required:"true"`
	Database       string `env:"NAME" env-required:"true"`
	SSLMode        string `env:"SSLMODE" env-default:"disable"`
	MaxConnections int    `env:"MAX_CONNECTIONS" env-default:"25"`
}

func (s *StorageConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		s.User,
		s.Password,
		s.Host,
		s.Port,
		s.Database,
		s.SSLMode)
}

func MustLoadConfig() *Config {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Failed to read environment variables: %s", err)
	}

	return &cfg
}
