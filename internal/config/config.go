package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// struct tags -> They tell Go libraries how to read data into struct fields, So, when we load the YAML file, Go knows how to fill these values into your struct.
type HTTPServer struct {
	Address string `yaml:"address" env-requried:"true"`
}

type Config struct {
	//means `what will be the value of this -> from where we are getting called struct tags`
	Env          string               `yaml:"env" env:"ENV" env-requried:"true"`
	Storage_path string               `yaml:"storage_path" env-requried:"true"`
	HTTPServer   `yaml:"http_server"` //struct embed
}

func MustLoad() *Config {
	var configPath string
	// If we run the app like: CONFIG_PATH=config/local.yaml go run cmd/go-server/main.go. CONFIG_PATH would be picked up here. Check if there’s an environment variable named CONFIG_PATH already set in the system.
	configPath = os.Getenv("CONFIG_PATH")

	// If no environment variable, it checks if we gave a command-line flag, like: go run cmd/go-server/main.go --config config/local.yam
	if configPath == "" {
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
	// cleanenv is an external library that reads our YAML file and fills in the struct automatically — just like dotenv fills process.env in Node.
	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("can not read config file: %s", err.Error())
	}

	return &cfg
}
