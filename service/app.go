package service

import (
	"cheeseburger/app/routes"
	"log"
	"net/http"

	"github.com/dgraph-io/badger/v4"
)

// RunAppServer starts the MVC blog service
func RunAppServer(args []string) {
	// Extract vanity name if provided
	var vanityName string
	for i := 0; i < len(args); i++ {
		if args[i] == "--vanity-name" && i+1 < len(args) {
			vanityName = args[i+1]
			// Remove the flag and value from args
			args = append(args[:i], args[i+2:]...)
			break
		}
	}

	// Set up the database and router
	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatalf("Failed to open Badger DB: %v", err)
	}
	defer db.Close()

	router := routes.SetupMVCRoutes(db)
	if router == nil {
		log.Fatal("Failed to setup MVC routes")
	}

	// Start the server with Tor
	log.Println("Starting MVC blog service on port 8080")
	runTorHiddenService(vanityName, func() {
		if err := http.ListenAndServe(":8080", router); err != nil {
			log.Fatalf("MVC server error: %v", err)
		}
	})
}
