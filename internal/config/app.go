package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	Env                string      `yaml:"env" required:"true"`
	HttpConfig         *HttpServer `yaml:"http_server" required:"true"`
	S3Config           *S3Config
	MaxCachedVideosUrl int `yaml:"max_cached_videos_url" required:"true"`
}

type HttpServer struct {
	Host        string        `yaml:"host" required:"true"`
	Port        string        `yaml:"port" default:"8080"`
	Timeout     time.Duration `yaml:"timeout" default:"5s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" default:"30s"`
}

type S3Config struct {
	Region     string
	AccessKey  string
	SecretKey  string
	BucketName string
	Endpoint   string
}

func MustLoadAppConfig() *AppConfig {
	err := godotenv.Load()
	if err != nil {
		log.Printf("No .env file found, proceeding with environment variables")
	}
	configPath := os.Getenv("APP_CONFIG_PATH")
	if configPath == "" {
		log.Fatal("APP_CONFIG_PATH environment variable not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatal("APP_CONFIG_PATH does not exist")
	}

	var appConfig AppConfig
	if err := cleanenv.ReadConfig(configPath, &appConfig); err != nil {
		log.Fatal(err)
	}

	s3Config := &S3Config{
		Region:     os.Getenv("S3_REGION"),
		AccessKey:  os.Getenv("S3_ACCESS_KEY"),
		SecretKey:  os.Getenv("S3_SECRET_KEY"),
		BucketName: os.Getenv("S3_BUCKET"),
		Endpoint:   os.Getenv("S3_ENDPOINT"),
	}

	appConfig.S3Config = s3Config

	return &appConfig
}
