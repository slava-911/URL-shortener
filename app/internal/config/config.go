package config

import (
	"flag"
	"log"
	"os"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

// Так как это приложение будет в докере, то не нужно в env:"IS_DEBUG" писать название приложения, например, так
// env:"URL-shortener_IS_DEBUG", то есть не приходится писать конкретно что за переменная окружения. А если даже
// приложение не будет в докере, то можно запустить через systemd и там сделать scoped переменные окружения доступные
// только этому приложению.
type Config struct {
	IsDebug       bool `env:"IS_DEBUG" env-default:"false"`
	IsDevelopment bool `env:"IS_DEV" env-default:"false"`
	HTTP          struct {
		IP           string        `yaml:"ip" env:"HTTP-IP"`
		Port         int           `yaml:"port" env:"HTTP-PORT"`
		ReadTimeout  time.Duration `yaml:"read_timeout" env:"HTTP-READ-TIMEOUT"`
		WriteTimeout time.Duration `yaml:"write_timeout" env:"HTTP-WHITE-TIMEOUT"`
		CORS         struct {
			AllowedMethods     []string `yaml:"allowed_methods" env:"HTTP-CORS-ALLOWED-METHODS"`
			AllowedOrigins     []string `yaml:"allowed_origins" env:"HTTP-CORS-ALLOWED-ORIGINS"`
			AllowCredentials   bool     `yaml:"allow_credentials" env:"HTTP-CORS-ALLOW-CREDENTIALS"`
			AllowedHeaders     []string `yaml:"allowed_headers" env:"HTTP-CORS-ALLOWED-HEADERS"`
			OptionsPassthrough bool     `yaml:"options_passthrough" env:"HTTP-CORS-OPTIONS-PASSTHROUGH"`
			ExposedHeaders     []string `yaml:"exposed_headers" env:"HTTP-CORS-EXPOSED-HEADERS"`
			Debug              bool     `yaml:"debug" env:"HTTP-CORS-DEBUG"`
		} `yaml:"cors"`
	} `yaml:"http"`
	AppConfig struct {
		LogLevel  string `env:"LOG_LEVEL" env-default:"trace"`
		AdminUser struct {
			Email    string `env:"ADMIN_EMAIL" env-default:"admin"`
			Password string `env:"ADMIN_PWD" env-default:"admin"`
		}
	}
	JWT struct {
		Secret string `env:"JWT_SECRET" env-required:"true"`
	}
	PostgreSQL struct {
		Username string `env:"DB_USERNAME" env-required:"true"`
		Password string `env:"DB_PASSWORD" env-required:"true"`
		Host     string `env:"DB_HOST" env-required:"true"`
		Port     string `env:"DB_PORT" env-required:"true"`
		Database string `env:"DB_DATABASE" env-required:"true"`
	}
}

const (
	EnvConfigPathName  = "CONFIG-PATH"
	FlagConfigPathName = "config"
)

var configPath string
var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		flag.StringVar(&configPath, FlagConfigPathName, "configs/config.local.yaml", "this is app config file")
		flag.Parse()

		log.Print("config init")

		if configPath == "" {
			configPath = os.Getenv(EnvConfigPathName)
		}

		if configPath == "" {
			log.Fatal("config path is required")
		}

		if err := godotenv.Load("../.env"); err != nil {
			log.Fatalf("error loading env variables: %v", err)
		}

		instance = &Config{}

		if err := cleanenv.ReadConfig(configPath, instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			log.Println(help)
			log.Fatal(err)
		}

	})
	return instance
}
