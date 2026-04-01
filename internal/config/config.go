package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Port        int    `yaml:"port"`
	ContextSize int    `yaml:"context_size"`
	ModelPath   string `yaml:"model_path"`
}

type APIConfig struct {
	Endpoint string `yaml:"endpoint"`
	APIKey   string `yaml:"api_key"`
	Model    string `yaml:"model"`
}

type ExecSafety struct {
	AllowList  []string `yaml:"allow_list"`
	DeniedList []string `yaml:"denied_list"`
}

type Config struct {
	Mode   string       `yaml:"mode"`
	Server ServerConfig `yaml:"server"`
	API    APIConfig    `yaml:"api"`
	Safety ExecSafety   `yaml:"safety"`
	Language string     `yaml:"language"`
}

func DefaultConfig() *Config {
	return &Config{
		Mode: "api",
		Server: ServerConfig{
			Port:        8398,
			ContextSize: 2048,
		},
		API: APIConfig{
			Endpoint: "https://api.openai.com/v1",
			Model:    "gpt-4o-mini",
		},
		Safety: ExecSafety{
			AllowList:  []string{HomeDir()},
			DeniedList: []string{},
		},
		Language: "en",
	}
}

func Load() (*Config, error) {
	path := ConfigFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}
	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Save() error {
	if err := EnsureDir(AppDir()); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigFilePath(), data, 0600)
}
