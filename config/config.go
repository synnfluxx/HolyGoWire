package config

import (
	"TextMeByte/internal/logger"
	"os"

	"github.com/BurntSushi/toml"
)


type Config struct {
	BindAddr    string `toml:"bind_addr"`    
	DatabaseURL string `toml:"database_url"` 
	SecretKey   string `toml:"secret_key"`   
	Mode        string `toml:"mode"`         
	LogLevel    string `toml:"log_level"`    
}

func NewConfig() *Config {
	return &Config{
		BindAddr: ":3131", 
	}
}

func ParseConfig(cfg *Config, configPath string) {
	_, err := toml.DecodeFile(configPath, cfg)
	if err != nil {
		logger.Log.Fatalf("Failed to parse configuration file %s: %v", configPath, err)
	}

	os.Setenv("BIND_ADDR", cfg.BindAddr)
	os.Setenv("DATABASE_URL", cfg.DatabaseURL)
	os.Setenv("SECRET_KEY", cfg.SecretKey)
	os.Setenv("MODE", cfg.Mode)
	os.Setenv("LOG_LEVEL", cfg.LogLevel)
}