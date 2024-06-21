package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func checkenv() {
	log.Println("Loading env")
	if err := godotenv.Load(); err == nil {
		log.Println("good : Found file .env")
	} else {
		log.Println("error: No .env file found")
	}
}

func checkKeyEnv(keys []string) {
	for _, key := range keys {
		val := os.Getenv(key)
		if val == "" {
			log.Printf("error: key %s not found \n", key)
		}
	}
}
