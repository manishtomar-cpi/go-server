package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPServer struct {
	Address string
}

type Config struct {
	//means `what will be the value of this -> from where we are getting called struct tags`
	Env          string               `yaml:"env" env:"ENV" env-requried:"true"`
	Storage_path string               `yaml:"storage_path" env-requried:"true"`
	HTTPServer   `yaml:"http_server"` //struct embed
}

func MustLoad() *Config {
	var configPath string
	configPath = os.Getenv("CONFIG_PATH")

	if configPath == "" { // if config not passed in envs we check from flag mean the start command go run -"flags_here"
		flags := flag.String("config", "", "path to the cofig file")
		flag.Parse()
		configPath = *flags //because flags is the pointer

		if configPath == "" {
			log.Fatal("Config path is not set")
		}
	}

	//if file is not present in the folder
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exists: %s", configPath)
	}

	var cfg Config

	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("can not read config file: %s", err.Error())
	}

	return &cfg
}
