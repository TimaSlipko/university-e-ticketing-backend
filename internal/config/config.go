package config

import (
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type (
	Config struct {
		Server   ServerConfig   `envconfig:"SERVER"`
		Database DatabaseConfig `envconfig:"DB"`
		Redis    RedisConfig    `envconfig:"REDIS"`
		JWT      JWTConfig      `envconfig:"JWT"`
		Payment  Payment        `envconfig:"PAYMENT"`
	}

	ServerConfig struct {
		Port         string        `envconfig:"PORT" default:"8080"`
		Host         string        `envconfig:"HOST" default:"0.0.0.0"`
		ReadTimeout  time.Duration `envconfig:"READ_TIMEOUT" default:"10s"`
		WriteTimeout time.Duration `envconfig:"WRITE_TIMEOUT" default:"10s"`
	}

	DatabaseConfig struct {
		Host     string `envconfig:"HOST" default:"localhost"`
		Port     string `envconfig:"PORT" default:"5432"`
		User     string `envconfig:"USER" default:"postgres"`
		Password string `envconfig:"PASSWORD" required:"false"`
		Name     string `envconfig:"NAME" default:"e_ticketing_dev"`
		SSLMode  string `envconfig:"SSL_MODE" default:"disable"`
		MaxConns int    `envconfig:"MAX_CONNS" default:"25"`
		MaxIdle  int    `envconfig:"MAX_IDLE" default:"5"`
	}

	RedisConfig struct {
		Host     string `envconfig:"HOST" default:"localhost"`
		Port     string `envconfig:"PORT" default:"6379"`
		Password string `envconfig:"PASSWORD" required:"false"`
		DB       int    `envconfig:"DB" default:"0"`
	}

	JWTConfig struct {
		Secret          string        `envconfig:"SECRET"`
		AccessDuration  time.Duration `envconfig:"ACCESS_DURATION" default:"15m"`
		RefreshDuration time.Duration `envconfig:"REFRESH_DURATION" default:"168h"` // 7 days
		Issuer          string        `envconfig:"ISSUER" default:"e-ticketing-system"`
	}

	Payment struct {
		IsMocked bool `envconfig:"IS_MOCKED" default:"true"`
	}
)

func Load() *Config {
	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	return &cfg
}
