package main

import (
	"bytes"
	"cheeseburger/vanity"
	"crypto/ed25519"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
)

const cliVersion = "1.0.0"

//go:embed tor-bins/tor-linux-x86_64
var torBinary []byte

type VanityKey struct {
	OnionAddress string `json:"onion_address"`
	PublicKey    string `json:"public_key"`
	PrivateKey   string `json:"private_key"`
	Attempts     uint64 `json:"attempts"`
	Timestamp    string `json:"timestamp"`
}

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
		serve()
	case "mvc":
		serveMvc()
	default:
		fmt.Printf("Unknown command: %s\n\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	helpText := `Usage: cheeseburger <command> [options]
Commands:
  help                           Display this help message.
  version                        Show version information.
  vanity    [options]            Generate a vanity onion address (e.g., vanity --prefix test [--save]).
  serve <static_directory> [--vanity-name <name>]  
                                 Run the Tor hidden service for a static file server.
  mvc [--vanity-name <name>]     Run the Tor hidden service for the MVC blog application.
`
	fmt.Println(helpText)
}

// serve sets up Tor integration and runs a static file server.
// It expects a static directory as the second argument, with an optional "--vanity-name" parameter.
func serve() {
	if len(os.Args) < 3 || (len(os.Args) > 3 && os.Args[3] == "--vanity-name" && len(os.Args) < 5) {
		fmt.Println("Error: static directory path and optional --vanity-name are required for serve command")
		os.Exit(1)
	}
	staticDir := os.Args[2]
	runTorHiddenService(func() {
		log.Printf("Starting static file server on port 8080 serving directory: %s", staticDir)
		if err := startStaticHTTPServer(":8080", staticDir); err != nil {
			log.Printf("Static server error: %v", err)
		}
	})
}

// serveMvc sets up Tor integration and runs the MVC blog service.
func serveMvc() {
	runTorHiddenService(func() {
		// Open or initialize the Badger DB at data/badger
		opts := badger.DefaultOptions("data/badger")
		db, err := badger.Open(opts)
		if err != nil {
			log.Fatalf("Failed to open Badger DB: %v", err)
		}
		defer db.Close()

		// Setup MVC routes using the Badger DB instance.
		router := routesSetupMVCRoutes(db)
		if router == nil {
			log.Fatal("Failed to setup MVC routes")
		}

		log.Println("Starting MVC blog service on port 8080 via Tor")
		if err := http.ListenAndServe(":8080", router); err != nil {
			log.Fatalf("MVC server error: %v", err)
		}
	})
}

// runTorHiddenService performs the common Tor integration steps for both static and MVC services.
// It accepts a function (serveFunc) that starts the desired HTTP server.
func runTorHiddenService(serveFunc func()) {
	// Determine vanity key path based on optional vanity name.
	vanityName := ""
	for i, arg := range os.Args {
		if arg == "--vanity-name" && i+1 < len(os.Args) {
			vanityName = os.Args[i+1]
			break
		}
	}
	_ = vanityName

	persistentKeyPath := ""
	if vanityName != "" {
		persistentKeyPath = filepath.Join("data", "vanity", vanityName, "vanity.json")
	} else {
		persistentKeyPath = filepath.Join("data", "vanity", "default", "vanity.json")
	}
	persistent := false
	if _, err := os.Stat(persistentKeyPath); err == nil {
		persistent = true
	}

	var dataDir, hsDir, torrcPath string
	var tempParentDir string
	var err error

	if persistent {
		// Use the vanity key directory directly as the hidden service directory.
		hsDir = filepath.Join(getCurrentDirectory(), filepath.Dir(persistentKeyPath))
		log.Printf("Using hidden service directory: %s", hsDir)
		log.Printf("Using vanity key from: %s", persistentKeyPath)

		// Verify vanity key file.
		keyData, err := os.ReadFile(persistentKeyPath)
		if err != nil {
			log.Fatalf("Failed to read vanity key file: %v", err)
		}
		var vk VanityKey
		if err := json.Unmarshal(keyData, &vk); err != nil {
			log.Fatalf("Failed to unmarshal vanity key JSON: %v", err)
		}
		log.Printf("Using vanity key with onion address: %s", vk.OnionAddress)

		// Validate secret key file.
		secretKeyPath := filepath.Join(hsDir, "hs_ed25519_secret_key")
		secretKeyData, err := os.ReadFile(secretKeyPath)
		if err != nil {
			log.Fatalf("Failed to read secret key file: %v", err)
		}
		log.Printf("Secret key file size: %d bytes", len(secretKeyData))
		secretHeader := make([]byte, 32)
		copy(secretHeader, []byte("== ed25519v1-secret: type0 =="))
		if len(secretKeyData) != 96 || !bytes.Equal(secretKeyData[:32], secretHeader) {
			log.Fatalf("Secret key file has invalid format")
		}
		log.Printf("Secret key header verified: %x", secretKeyData[:32])
		privateScalar := secretKeyData[32:64]
		derivedPriv := ed25519.NewKeyFromSeed(privateScalar)
		derivedPub := derivedPriv.Public().(ed25519.PublicKey)
		log.Printf("Derived public key (hex): %x", derivedPub)

		// Validate public key file.
		publicKeyPath := filepath.Join(hsDir, "hs_ed25519_public_key")
		publicKeyData, err := os.ReadFile(publicKeyPath)
		if err != nil {
			log.Fatalf("Failed to read public key file: %v", err)
		}
		log.Printf("Public key file size: %d bytes", len(publicKeyData))
		publicHeader := make([]byte, 32)
		copy(publicHeader, []byte("== ed25519v1-public: type0 =="))
		if len(publicKeyData) != 64 || !bytes.Equal(publicKeyData[:32], publicHeader) {
			log.Fatalf("Public key file has invalid format")
		}
		log.Printf("Public key header verified: %x", publicKeyData[:32])
		storedPub := publicKeyData[32:64]
		if !bytes.Equal(storedPub, derivedPub) {
			log.Fatalf("Public key mismatch. Stored key does not match key derived from secret key")
		}
		log.Printf("Public key verified: matches key derived from secret key")

		hostnamePath := filepath.Join(hsDir, "hostname")
		hostnameData, err := os.ReadFile(hostnamePath)
		if err != nil {
			log.Fatalf("Failed to read hostname file: %v", err)
		}
		hostname := strings.TrimSpace(string(hostnameData))
		if hostname != vk.OnionAddress {
			log.Fatalf("Hostname mismatch. Expected %s but found %s", vk.OnionAddress, hostname)
		}
		log.Printf("Hostname verified: %s", hostname)

		// Set permissions.
		os.Chmod(hsDir, 0700)
		os.Chmod(secretKeyPath, 0600)
		os.Chmod(publicKeyPath, 0600)
		os.Chmod(hostnamePath, 0600)

		// For persistent mode, use a temporary directory for Tor's DataDirectory.
		tempParentDir, err = os.MkdirTemp("", "tor-example-")
		if err != nil {
			log.Fatalf("Failed to create temp directory: %v", err)
		}
		dataDir = filepath.Join(tempParentDir, "data")
		if err := os.MkdirAll(dataDir, 0700); err != nil {
			log.Fatalf("Failed to create data directory: %v", err)
		}
		torrcPath = filepath.Join(tempParentDir, "torrc")
	} else {
		// Temporary mode: create temporary directories.
		tempParentDir, err = os.MkdirTemp("", "tor-example-")
		if err != nil {
			log.Fatalf("Failed to create temp directory: %v", err)
		}
		dataDir = filepath.Join(tempParentDir, "data")
		hsDir = filepath.Join(getCurrentDirectory(), filepath.Dir(persistentKeyPath))
		if err := os.MkdirAll(dataDir, 0700); err != nil {
			log.Fatalf("Failed to create data directory: %v", err)
		}
		if err := os.MkdirAll(hsDir, 0700); err != nil {
			log.Fatalf("Failed to create hidden service directory: %v", err)
		}
		torrcPath = filepath.Join(tempParentDir, "torrc")
	}

	// Create torrc config file.
	torrcContent := fmt.Sprintf(`
# Write Tor's runtime data here
DataDirectory %s

# Open a SOCKS port for local connections (optional)
SocksPort 9050

# Our hidden service
HiddenServiceDir %s
HiddenServicePort 80 127.0.0.1:8080

# Log notice to stdout
Log notice stdout
`, dataDir, hsDir)
	log.Printf("Writing torrc to: %s", torrcPath)
	log.Printf("Using hidden service directory: %s", hsDir)
	log.Printf("Torrc content:\n%s", torrcContent)
	if err := os.WriteFile(torrcPath, []byte(torrcContent), 0600); err != nil {
		log.Fatalf("Failed to write torrc file: %v", err)
	}

	// Start the HTTP service (either static or MVC) in a separate goroutine.
	go serveFunc()

	// Write the embedded Tor binary to a temporary file.
	tmpTor, err := os.CreateTemp("", "tor-")
	if err != nil {
		log.Fatalf("Failed to create temporary file for embedded tor: %v", err)
	}
	if _, err = tmpTor.Write(torBinary); err != nil {
		log.Fatalf("Failed to write embedded tor binary to temp file: %v", err)
	}
	if err = tmpTor.Chmod(0755); err != nil {
		log.Fatalf("Failed to set executable permissions on embedded tor binary: %v", err)
	}
	tmpTorPath := tmpTor.Name()
	tmpTor.Close()

	// Run Tor with the generated config.
	cmd := exec.Command(tmpTorPath, "-f", torrcPath)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to get stdout pipe for tor: %v", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalf("Failed to get stderr pipe for tor: %v", err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start Tor process: %v", err)
	}
	// Read and print Tor logs.
	go io.Copy(os.Stdout, stdoutPipe)
	go io.Copy(os.Stderr, stderrPipe)
	// Wait for Tor to generate the onion hostname file.
	time.Sleep(5 * time.Second)
	hostnamePath := filepath.Join(hsDir, "hostname")
	hostnameBytes, err := os.ReadFile(hostnamePath)
	if err != nil {
		log.Printf("Could not read onion hostname yet: %v", err)
		log.Printf("Tor may still be starting. Waiting a bit longer...")
		time.Sleep(5 * time.Second)
		hostnameBytes, err = os.ReadFile(hostnamePath)
		if err != nil {
			log.Fatalf("Still no hostname file: %v", err)
		}
	}
	hostname := strings.TrimSpace(string(hostnameBytes))
	log.Printf("Your onion service is live at: %s", hostname)
	log.Printf("Press Ctrl+C to stop.\n")
	// Wait for Tor to exit.
	if err := cmd.Wait(); err != nil {
		log.Printf("Tor process exited with an error: %v", err)
	}
	// Cleanup temporary directories and files if not persistent.
	if !persistent {
		os.RemoveAll(tempParentDir)
		os.Remove(tmpTorPath)
		log.Println("Cleaned up temporary directories and embedded tor binary. Exiting.")
	}
}

func getCurrentDirectory() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}
	return dir
}

func startStaticHTTPServer(addr, staticDir string) error {
	handler := http.FileServer(http.Dir(staticDir))
	http.Handle("/", handler)
	return http.ListenAndServe(addr, nil)
}

func routesSetupMVCRoutes(db *badger.DB) http.Handler {
	// Placeholder for actual MVC router setup.
	// Replace this with the actual implementation from your routes package.
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("MVC blog service"))
	})
	return mux
}
