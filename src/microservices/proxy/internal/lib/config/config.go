package config

type Config struct {
	Port             string
	MonolithURL      string
	MoviesServiceURL string
	EventsServiceURL string
	GradualMigration bool
}

func NewConfig() (*Config, error) {
	return &Config{}, nil
}
