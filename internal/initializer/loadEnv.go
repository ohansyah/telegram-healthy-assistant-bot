package initializer

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	fmt.Println("Load Env...")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
