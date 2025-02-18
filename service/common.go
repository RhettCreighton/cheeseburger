package service

import (
	"log"
	"os"
)

const dbPath = "data/badger"

func getCurrentDirectory() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}
	return dir
}
