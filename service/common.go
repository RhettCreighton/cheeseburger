package service

import (
	"log"
	"os"
)

// Database path - variable to allow testing with different paths
var dbPath = "data/badger"

func getCurrentDirectory() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}
	return dir
}
