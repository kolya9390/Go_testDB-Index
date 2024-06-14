package config

import (
	"github.com/joho/godotenv"
)

type Config struct {
	GRPCPort string   `yaml:"grpc_port"`
	GRPCHost string   `yml:grpc_host`
	Postgres ConfigDB `yaml:"postgres"`
}

type ConfigDB struct {
	User          string `yaml:"user"`
	Password      string `yaml:"password"`
	Host          string `yaml:"host"`
	Port          string `yaml:"port"`
	Database      string `yaml:"database"`
	MigrationsDir string `yaml:"migrationsDir"`
}

func LoadConfigFromEnv(path string) (Config, error) {
	env, err := godotenv.Read(path)
	if err != nil {
		return Config{}, err
	}
	return Config{
		GRPCHost: env["GRPC_HOST"],
		GRPCPort: env["GRPC_PORT"],
		Postgres: ConfigDB{
			User:          env["DB_USER"],
			Password:      env["DB_PASSWORD"],
			Host:          env["DB_HOST"],
			Port:          env["DB_PORT"],
			Database:      env["DB_NAME"],
			MigrationsDir: env["DB_MIGRATIONS_DIR"],
		},
	}, nil
}
