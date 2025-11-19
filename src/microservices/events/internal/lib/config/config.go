package config

type Config struct {
	Port         string
	KafkaBrokers string
	DSN          string
}

func NewConfig() (*Config, error) {
	return &Config{}, nil
}
