package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env   string `yaml:"env" env-default:"local"`
	Redis struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`
	UserStoragePath string `yaml:"user_storage_path" env-required:"true"`
	AdminLogin      string `yaml:"admin_login"`
	AdminPassword   string `yaml:"admin_password"`
	SecretKey       string `yaml:"secret_key"`
	Port            string `yaml:"port"`
	LogLevel        string `yaml:"log_level"`
	Hasher          string `yaml:"hasher"`
	Grpc            struct {
		Port    string        `yaml:"port"`
		Timeout time.Duration `yaml:"timeout"`
	} `yaml:"grpc"`
}

func MustLoad() *Config {
	configPath := fetchConfig()

	if configPath == "" {
		panic("config path not found")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file not found")
	}

	config := &Config{}

	if err := cleanenv.ReadConfig(configPath, config); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return config
}

func fetchConfig() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		return os.Getenv("CONFIG_PATH")
	}

	return res
}
