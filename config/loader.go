package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func Get() *Config {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error when load .env file configuration: ", err.Error())
	}

	return &Config{
		Server: Server{
			Host: os.Getenv("SERVER_HOST"),
			Port: os.Getenv("SERVER_PORT"),
		},
		Telegram: Telegram{
			Token: os.Getenv("TELEGRAM_BOT_TOKEN"),
		},
		Gemini: Gemini{
			Key:   os.Getenv("GEMINI_KEY"),
			Model: os.Getenv("GEMINI_MODEL"),
		},
	}
}
