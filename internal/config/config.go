package config

import (
	"flag"
	"os"
	"rip/internal/app"
	"rip/internal/app/graphql"
	httpapp "rip/internal/app/http"
	"rip/internal/repository/postgres"
	"rip/internal/repository/s3minio"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string            `yaml:"env" env-default:"local"`
	App        app.Config        `yaml:"app"`
	HTTP       httpapp.Config    `yaml:"http_server"`
	Postgresql postgres.Config   `yaml:"postgresql"`
	S3minio    s3minio.Config    `yaml:"s3minio"`
	GraphQL    graphqlapp.Config `yaml:"graphql"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	return MustLoadPath(configPath)
}

func MustLoadPath(configPath string) *Config {
	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	return &cfg
}

// fetchConfigPath fetches config path from command line flag or environment variable.
// Priority: flag > env > default.
// Default value is empty string.
func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "config/config.yaml", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
