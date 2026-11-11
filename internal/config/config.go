package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/sophrone/ssg/internal/page"
)

type Config struct {
	Title      string                `yaml:"title"`
	BaseURL    string                `yaml:"base_url"`
	Navigation []page.NavigationLink `yaml:"navigation"`
}

func Load(path string) (*Config, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(source, &config); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &config, nil
}
