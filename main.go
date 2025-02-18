package main

import (
	"cheeseburger/service"
	"cheeseburger/vanity"
	"fmt"
	"os"
	"strings"
)

const cliVersion = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	cmd := strings.ToLower(os.Args[1])
	switch cmd {
	case "help":
		printHelp()
	case "version":
		fmt.Printf("cheeseburger version %s\n", cliVersion)
	case "vanity":
		// Remove the subcommand so flag parsing in vanity.RunVanity works correctly.
		os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
		vanity.RunVanity()
	case "serve":
		if len(os.Args) < 3 {
			fmt.Println("Error: static directory path required for serve command")
			os.Exit(1)
		}
		staticDir := os.Args[2]
		// Extract vanity name if provided
		var vanityName string
		for i := 3; i < len(os.Args); i++ {
			if os.Args[i] == "--vanity-name" && i+1 < len(os.Args) {
				vanityName = os.Args[i+1]
				break
			}
		}
		service.RunStaticTorServer(staticDir, vanityName)
	case "mvc":
		// Pass all args after "mvc" to the handler
		service.HandleCommand(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	helpText := `Usage: cheeseburger <command> [options]

Commands:
  help                           Display this help message
  version                        Show version information
  vanity [options]              Generate a vanity onion address (e.g., vanity --prefix test [--save])
  serve <static_directory>      Run static file server with Tor hidden service
  mvc                           MVC blog commands:
    serve [--vanity-name <name>]  Run the blog service (runs as Tor hidden service)
    clean                         Clean the database
    init                          Initialize database
    backup                        Backup database
    restore [file]               Restore from backup
    help                         Show MVC help
`
	fmt.Println(helpText)
}
