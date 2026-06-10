package config

import (
	"os"
)

// Config holds runtime configuration loaded from environment variables.
type Config struct {
	Port         string
	DatabasePath string
	BaseURL      string
	StaticDir    string
	TemplateDir  string
	DocsDir      string
	ChangelogPath string
}

// Load reads configuration from the environment with sensible defaults.
func Load() Config {
	return Config{
		Port:         envOr("PORT", "8080"),
		DatabasePath: envOr("DATABASE_PATH", "./data/payments.db"),
		BaseURL:      envOr("BASE_URL", "http://localhost:8080"),
		StaticDir:    envOr("STATIC_DIR", "web/static"),
		TemplateDir:  envOr("TEMPLATE_DIR", "web/templates"),
		DocsDir:       envOr("DOCS_DIR", "docs"),
		ChangelogPath: envOr("CHANGELOG_PATH", "CHANGELOG.md"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
