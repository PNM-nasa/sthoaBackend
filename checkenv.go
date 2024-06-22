package main

import (
	"log"
	"os"
)

func checkKeyEnv(keys []string) {
	for _, key := range keys {
		val := os.Getenv(key)
		if val == "" {
			log.Printf("error: key %s not found \n", key)
		}
	}
}
