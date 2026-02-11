package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	HomeDir    string
	OutputDir  string
	DBPath     string
	LogPath    string
	SchemaPath string
}

func Load() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	outputDir := filepath.Join(homeDir, "Downloads", "extracted_jobs")

	cfg := &Config{
		HomeDir:    homeDir,
		OutputDir:  outputDir,
		DBPath:     filepath.Join(outputDir, "jobs.db"),
		LogPath:    filepath.Join(homeDir, "Downloads", "extractor.log"),
		SchemaPath: filepath.Join(homeDir, "Projects", "text-extractor", "native-host", "schema.sql"),
	}

	return cfg, nil
}

func (c *Config) EnsureDirectories() error {
	return os.MkdirAll(c.OutputDir, 0755)
}
