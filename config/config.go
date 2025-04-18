package config

import (
	"github.com/caarlos0/env"
	"github.com/gofiber/fiber/v2/log"
	"github.com/joho/godotenv"
)

type EnvConfig struct {
	ServerPort   string `env:"SERVER_PORT"`
	MaxFileSize  int    `env:"MAX_FILE_SIZE"`
	CosConfig    CosConfig
	UploadConfig UploadConfig
	RedisConfig  RedisConfig
	DBConfig     DBConfig
}

type CosConfig struct {
	SecretID  string `env:"SECRET_ID"`
	SecretKey string `env:"SECRET_KEY"`
	CosURL    string `env:"COS_URL"`
	Bucket    string `env:"COS_BUCKET"`
	Region    string `env:"COS_REGION"`
}

type UploadConfig struct {
	AccessKeyId     string `env:"ALIBABA_CLOUD_ACCESS_KEY_ID"`
	AccessKeySecret string `env:"ALIBABA_CLOUD_ACCESS_KEY_SECRET"`
	EndPoint        string `env:"ALIBABA_CLOUD_END_POINT"`
}

type DBConfig struct {
	DBHost         string `env:"DB_HOST"`
	DBPort         string `env:"DB_PORT"`
	DBUser         string `env:"DB_USER"`
	DBPassword     string `env:"DB_PASSWORD"`
	DBName         string `env:"DB_NAME"`
	DBSSLMode      string `env:"DB_SSL_MODE"`
	DBMaxIdleConns int    `env:"DB_MAX_IDLE_CONNS"`
	DBMaxOpenConns int    `env:"DB_MAX_OPEN_CONNS"`
}

type RedisConfig struct {
	RedisHost     string `env:"REDIS_HOST"`
	RedisPort     string `env:"REDIS_PORT"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	RedisDB       int    `env:"REDIS_DB"`
}

func NewEnvConfig() *EnvConfig {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	config := &EnvConfig{}
	if err := env.Parse(config); err != nil {
		log.Fatalf("Error parsing env: %v", err)
	}
	cosConfig := &CosConfig{}
	if err := env.Parse(cosConfig); err != nil {
		log.Fatalf("Error parsing env: %v", err)
	}
	uploadConfig := &UploadConfig{}
	if err := env.Parse(uploadConfig); err != nil {
		log.Fatalf("Error parsing env: %v", err)
	}
	dbConfig := &DBConfig{}
	if err := env.Parse(dbConfig); err != nil {
		log.Fatalf("Error parsing env: %v", err)
	}
	redisConfig := &RedisConfig{}
	if err := env.Parse(redisConfig); err != nil {
		log.Fatalf("Error parsing env: %v", err)
	}
	config.CosConfig = *cosConfig
	config.UploadConfig = *uploadConfig
	config.RedisConfig = *redisConfig
	config.DBConfig = *dbConfig
	return config
}
