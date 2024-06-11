package config


type Config struct {
	GRPCApiAddr string `yaml:"GRPCApiAddr"`
	Postgres    ConfigDB `yaml:"postgres"`
}

type ConfigDB struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Database string `yaml:"database"`
}