package service

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v4"
)

var osExit = os.Exit

// HandleCommand handles MVC subcommands and returns an exit code.
func HandleCommand(args []string) int {
	if len(args) < 1 {
		printMvcHelp()
		osExit(1)
		return 1
	}

	cmd := args[0]
	switch cmd {
	case "serve":
		RunAppServer(args[1:])
		return 0
	case "clean":
		clean()
		return 0
	case "init":
		initDb()
		return 0
	case "backup":
		backup()
		return 0
	case "restore":
		if len(args) < 2 {
			fmt.Println("Error: backup file path required for restore")
			osExit(1)
			return 1
		}
		return restore(args[1])
	case "help":
		printMvcHelp()
		return 0
	default:
		fmt.Printf("Unknown mvc command: %s\n\n", cmd)
		printMvcHelp()
		osExit(1)
		return 1
	}
}

// printMvcHelp prints help for MVC subcommands.
func printMvcHelp() {
	helpText := `Usage: cheeseburger mvc

Commands:
  serve [--vanity-name <name>]    Run the blog service (always runs as Tor hidden service)
  clean                           Clean the blog database
  init                            Initialize a new empty database
  backup                          Create a backup of the database
  restore [file]                  Restore database from backup
  help                            Display this help message
`
	fmt.Println(helpText)
}

// clean removes the database.
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
		fmt.Printf("Failed to clean database: %v\n", err)
		return
	}
	fmt.Println("Database cleaned successfully")
}

// initDb initializes a new empty database.
func initDb() {
	if _, err := os.Stat(dbPath); err == nil {
		fmt.Println("Database already exists. Use 'clean' first if you want to reinitialize.")
		return
	}

	if err := os.MkdirAll(dbPath, 0755); err != nil {
		fmt.Printf("Failed to create database directory: %v\n", err)
		return
	}

	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		return
	}
	defer db.Close()

	fmt.Println("Database initialized successfully")
}

// backup creates a backup of the database.
func backup() {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println("No database exists to backup")
		return
	}

	backupDir := "data/backups"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		fmt.Printf("Failed to create backup directory: %v\n", err)
		return
	}

	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		return
	}
	defer db.Close()

	backupFile := filepath.Join(backupDir, fmt.Sprintf("backup_%d.db", time.Now().Unix()))
	f, err := os.Create(backupFile)
	if err != nil {
		fmt.Printf("Failed to create backup file: %v\n", err)
		return
	}
	defer f.Close()

	if _, err := db.Backup(f, 0); err != nil {
		fmt.Printf("Failed to backup database: %v\n", err)
		return
	}

	fmt.Printf("Database backed up successfully to %s\n", backupFile)
}

// restore restores the database from a backup.
func restore(backupFile string) int {
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		fmt.Printf("Backup file does not exist: %s\n", backupFile)
		return 1
	}

	if _, err := os.Stat(dbPath); err == nil {
		fmt.Print("Existing database found. Do you want to replace it? [y/N] ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Operation cancelled")
			return 1
		}
		if err := os.RemoveAll(dbPath); err != nil {
			fmt.Printf("Failed to remove existing database: %v\n", err)
			return 1
		}
	}

	if err := os.MkdirAll(dbPath, 0755); err != nil {
		fmt.Printf("Failed to create database directory: %v\n", err)
		return 1
	}

	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		return 1
	}
	defer db.Close()

	f, err := os.Open(backupFile)
	if err != nil {
		fmt.Printf("Failed to open backup file: %v\n", err)
		return 1
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		fmt.Printf("Failed to stat backup file: %v\n", err)
		return 1
	}
	if fi.Size() == 0 {
		fmt.Printf("Backup file is empty: %s\n", backupFile)
		return 1
	}

	err = func() error {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic occurred during restore: %v", r)
			}
		}()
		return db.Load(f, 4)
	}()
	if err != nil {
		fmt.Printf("Failed to restore database: %v\n", err)
		return 1
	}

	fmt.Println("Database restored successfully")
	return 0
}
