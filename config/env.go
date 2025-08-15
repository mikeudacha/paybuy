package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PublicHost                    string
	Host                          string
	DBHost                        string
	DBUser                        string
	DBPassword                    string
	DBPort                        string
	DBName                        string
	JWTSecret                     string
	JWTExpirationInSeconds        string
	JWTRefreshSecret              string
	JWTRefreshExpirationInSeconds string
	Environment                   string
	CookieDomain                  string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error load .env file")
	}
	return &Config{
		PublicHost:                    os.Getenv("PUBLIC_HOST"),
		Host:                          os.Getenv("HOST"),
		DBUser:                        os.Getenv("DB_USERNAME"),
		DBPassword:                    os.Getenv("DB_PASSWORD"),
		DBPort:                        os.Getenv("DB_PORT"),
		DBName:                        os.Getenv("DB_NAME"),
		DBHost:                        os.Getenv("DB_HOST"),
		JWTSecret:                     os.Getenv("JWTSecret"),
		JWTExpirationInSeconds:        os.Getenv("JWTExpirationInSeconds"),
		JWTRefreshSecret:              os.Getenv("JWTRefreshSecret"),
		JWTRefreshExpirationInSeconds: os.Getenv("JWTRefreshExpirationInSeconds"),
		Environment:                   os.Getenv("ENVIRONMENT"),
		CookieDomain:                  os.Getenv("COOKIE_DOMAIN"),
	}
}
