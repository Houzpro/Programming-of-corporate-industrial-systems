package config

import (
	"time"
)

type Config struct {
	Game   GameConfig   `json:"game"`
	Server ServerConfig `json:"server"`
}

type GameConfig struct {
	CodeLength  int `json:"codeLength"`
	MaxPlayers  int `json:"maxPlayers"`
	MaxAttempts int `json:"maxAttempts"`
}

type ServerConfig struct {
	EnableTimeout   bool          `json:"enableTimeout"`
	InputTimeout    time.Duration `json:"inputTimeout"`
	WarningTime     time.Duration `json:"warningTime"`
	WarningInterval time.Duration `json:"warningInterval"`
}

func DefaultConfig() *Config {
	return &Config{
		Game: GameConfig{
			CodeLength:  4,
			MaxPlayers:  4,
			MaxAttempts: 10,
		},
		Server: ServerConfig{
			EnableTimeout:   false,
			InputTimeout:    30 * time.Second,
			WarningTime:     10 * time.Second,
			WarningInterval: 5 * time.Second,
		},
	}
}

func LoadConfig(filename string) (*Config, error) {
	config := DefaultConfig()
	return config, nil
}

func SaveConfig(config *Config, filename string) error {
	return nil
}
