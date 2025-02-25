package main

import (
	"cheeseburger/coverage"
	"cheeseburger/service"
	"cheeseburger/vanity"
	"flag"
	"fmt"
	"os"
	"strings"
)

var exit = os.Exit

const cliVersion = "1.0.0"

var RealMain = main
var PrintHelp = printHelp
var CliVersion = cliVersion

func main() {
	exitCode := run()
	exit(exitCode)
}

func run() int {
	if len(os.Args) < 2 {
		fmt.Println("Usage: cheeseburger <command>")
		return 1
	}

	cmd := strings.ToLower(os.Args[1])
	switch cmd {
	case "help":
		printHelp()
		return 0
	case "version":
		fmt.Printf("cheeseburger version %s\n", cliVersion)
		return 0
	case "vanity":
		// Remove the subcommand so flag parsing in vanity.RunVanity works correctly.
		os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
		vanity.RunVanity()
		return 0
	case "serve":
		if len(os.Args) < 3 {
			fmt.Println("Error: static directory path required for serve command")
			return 1
		}
		staticDir := os.Args[2]
		var vanityName string
		for i := 3; i < len(os.Args); i++ {
			if os.Args[i] == "--vanity-name" && i+1 < len(os.Args) {
				vanityName = os.Args[i+1]
				break
			}
		}
		service.RunStaticTorServer(staticDir, vanityName)
		return 0
	case "mvc":
		return service.HandleCommand(os.Args[2:])
	case "coverage":
		// Create a flag set for the coverage command.
		coverageFlags := flag.NewFlagSet("coverage", flag.ExitOnError)
		jsonFlag := coverageFlags.Bool("json", false, "Output coverage targets in JSON format")
		// Parse the flags starting from os.Args[2:].
		err := coverageFlags.Parse(os.Args[2:])
		if err != nil {
			fmt.Printf("Error parsing coverage flags: %v\n", err)
			return 1
		}
		// Build remaining args and include --json if the flag is set.
		remainingArgs := coverageFlags.Args()
		if *jsonFlag {
			remainingArgs = append(remainingArgs, "--json")
		}
		coverage.RunCoverage(remainingArgs)
		return 0
	default:
		fmt.Printf("Unknown command: %s\n\n", os.Args[1])
		printHelp()
		return 1
	}
}

func printHelp() {
	helpText := `Usage: cheeseburger <command> [options]

Commands:
  help                           Display this help message
  version                        Show version information
  vanity [options]               Generate a vanity onion address (e.g., vanity --prefix test [--save])
  serve <static_directory>       Run static file server with Tor hidden service
  mvc                           MVC blog commands:
    serve [--vanity-name <name>] Run the blog service (runs as Tor hidden service)
    clean                        Clean the database
    init                         Initialize database
    backup                       Backup database
    restore [file]               Restore from backup
    help                         Show MVC help
  coverage [--json]              Automatically run tests to generate a temporary coverage report and display its summary.
`
	fmt.Println(helpText)
}
