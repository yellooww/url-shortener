package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `yaml:"env" env:"ENV" env-default:"local"`
	DatabaseURL string `yaml:"database_url" env:"DATABASE_URL"`
	HTTPServer  `yaml:"http_server"`
	Auth        `yaml:"auth"`
}

type Auth struct {
	User     string `yaml:"user" env:"AUTH_USER"`
	Password string `yaml:"password" env:"AUTH_PASSWORD"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file %s does not exist", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		cfg.DatabaseURL = dbURL
		log.Println("Using DATABASE_URL from enviroment")
	}

	if user := os.Getenv("AUTH_USER"); user != "" {
		cfg.Auth.User = user
		log.Println("Using AUTH_USER from environment")
	}

	if password := os.Getenv("AUTH_PASSWORD"); password != "" {
		cfg.Auth.Password = password
		log.Println("Using AUTH_PASSWORD from environment")
	}

	if cfg.Auth.User == "" || cfg.Auth.Password == "" {
		log.Fatal("Username and Password is empty")
	}

	return &cfg
}
