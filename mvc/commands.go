package mvc

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"cheeseburger/routes"

	"github.com/dgraph-io/badger/v4"
)

const dbPath = "data/badger"

// HandleCommand handles MVC subcommands
func HandleCommand(args []string) {
	if len(args) < 1 {
		printMvcHelp()
		os.Exit(1)
	}

	cmd := args[0]
	switch cmd {
	case "serve":
		serve(args[1:])
	case "clean":
		clean()
	case "init":
		initDb()
	case "backup":
		backup()
	case "restore":
		if len(args) < 2 {
			fmt.Println("Error: backup file path required for restore")
			os.Exit(1)
		}
		restore(args[1])
	case "help":
		printMvcHelp()
	default:
		fmt.Printf("Unknown mvc command: %s\n\n", cmd)
		printMvcHelp()
		os.Exit(1)
	}
}

// printMvcHelp prints help for MVC subcommands
func printMvcHelp() {
	helpText := `Usage: cheeseburger mvc <command> [options]

Commands:
  serve [--vanity-name <name>]    Run the blog service (always runs as Tor hidden service)
  clean                           Clean the blog database
  init                           Initialize a new empty database
  backup                         Create a backup of the database
  restore [file]                 Restore database from backup
  help                           Display this help message
`
	fmt.Println(helpText)
}

// serve starts the MVC blog service
func serve(args []string) {
	// Open or initialize the Badger DB
	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatalf("Failed to open Badger DB: %v", err)
	}
	defer db.Close()

	// Setup MVC routes using the Badger DB instance
	router := routes.SetupMVCRoutes(db)
	if router == nil {
		log.Fatal("Failed to setup MVC routes")
	}

	log.Println("Starting MVC blog service on port 8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("MVC server error: %v", err)
	}
}

// clean removes the database
func clean() {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println("Database is already clean (does not exist)")
		return
	}

	fmt.Print("Are you sure you want to clean the database? This cannot be undone. [y/N] ")
	var response string
	fmt.Scanln(&response)
	if response != "y" && response != "Y" {
		fmt.Println("Operation cancelled")
		return
	}

	if err := os.RemoveAll(dbPath); err != nil {
		log.Fatalf("Failed to clean database: %v", err)
	}
	fmt.Println("Database cleaned successfully")
}

// initDb initializes a new empty database
func initDb() {
	if _, err := os.Stat(dbPath); err == nil {
		fmt.Println("Database already exists. Use 'clean' first if you want to reinitialize.")
		return
	}

	if err := os.MkdirAll(dbPath, 0755); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	fmt.Println("Database initialized successfully")
}

// backup creates a backup of the database
func backup() {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println("No database exists to backup")
		return
	}

	backupDir := "data/backups"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		log.Fatalf("Failed to create backup directory: %v", err)
	}

	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	backupFile := filepath.Join(backupDir, fmt.Sprintf("backup_%d.db", time.Now().Unix()))
	f, err := os.Create(backupFile)
	if err != nil {
		log.Fatalf("Failed to create backup file: %v", err)
	}
	defer f.Close()

	if _, err := db.Backup(f, 0); err != nil {
		log.Fatalf("Failed to backup database: %v", err)
	}

	fmt.Printf("Database backed up successfully to %s\n", backupFile)
}

// restore restores the database from a backup
func restore(backupFile string) {
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		fmt.Printf("Backup file does not exist: %s\n", backupFile)
		return
	}

	if _, err := os.Stat(dbPath); err == nil {
		fmt.Print("Existing database found. Do you want to replace it? [y/N] ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Operation cancelled")
			return
		}
		if err := os.RemoveAll(dbPath); err != nil {
			log.Fatalf("Failed to remove existing database: %v", err)
		}
	}

	if err := os.MkdirAll(dbPath, 0755); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	f, err := os.Open(backupFile)
	if err != nil {
		log.Fatalf("Failed to open backup file: %v", err)
	}
	defer f.Close()

	if err := db.Load(f, 4); err != nil {
		log.Fatalf("Failed to restore database: %v", err)
	}

	fmt.Println("Database restored successfully")
}
