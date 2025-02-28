package vanity

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunVanity(t *testing.T) {
	// Save original args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Create temp directory for test outputs
	tempDir, err := os.MkdirTemp("", "vanity_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testDataDir := filepath.Join(tempDir, "data", "vanity", "default")
	
	// Test case 1: Basic functionality with no prefix
	t.Run("NoPrefix", func(t *testing.T) {
		// Set up args for this test
		os.Args = []string{"cmd", "-workers", "1"}
		
		// Reset flags for each test
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		
		// Redirect standard logger output for testing
		var logOutput strings.Builder
		log.SetOutput(&logOutput)
		defer log.SetOutput(os.Stderr)
		
		RunVanity()
		
		output := logOutput.String()
		if !strings.Contains(output, "Vanity Onion Address:") {
			t.Errorf("Expected output to contain onion address, got: %s", output)
		}
		if !strings.Contains(output, "Public Key (hex):") {
			t.Errorf("Expected output to contain public key, got: %s", output)
		}
		if !strings.Contains(output, "Expanded Private Key (hex):") {
			t.Errorf("Expected output to contain private key, got: %s", output)
		}
	})
	
	// Test case 2: With save flag
	t.Run("WithSave", func(t *testing.T) {
		// Temporarily replace the data dir path
		os.Setenv("VANITY_DATA_DIR", testDataDir)
		defer os.Unsetenv("VANITY_DATA_DIR")
		
		// Set up args for this test
		os.Args = []string{"cmd", "-workers", "1", "-save"}
		
		// Reset flags for each test
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		
		// Redirect standard logger output for testing
		var logOutput strings.Builder
		log.SetOutput(&logOutput)
		defer log.SetOutput(os.Stderr)
		
		RunVanity()
		
		// Check if files were created
		files := []string{
			filepath.Join(testDataDir, "hs_ed25519_secret_key"),
			filepath.Join(testDataDir, "hs_ed25519_public_key"),
			filepath.Join(testDataDir, "hostname"),
		}
		
		for _, file := range files {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				t.Errorf("Expected file %s was not created", file)
			}
		}
		
		// Check hostname file format
		hostnameData, err := os.ReadFile(filepath.Join(testDataDir, "hostname"))
		if err != nil {
			t.Fatalf("Failed to read hostname file: %v", err)
		}
		
		if !strings.HasSuffix(string(hostnameData), ".onion\n") {
			t.Errorf("Hostname file does not have correct format: %s", string(hostnameData))
		}
	})
	
	// Test case 3: With prefix
	t.Run("WithPrefix", func(t *testing.T) {
		// Set up args for this test - use a very short prefix to avoid long test times
		os.Args = []string{"cmd", "-workers", "1", "-prefix", "a"}
		
		// Reset flags for each test
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		
		// Redirect standard logger output for testing
		var logOutput strings.Builder
		log.SetOutput(&logOutput)
		defer log.SetOutput(os.Stderr)
		
		RunVanity()
		
		output := logOutput.String()
		if !strings.Contains(output, "Vanity Onion Address: a") {
			t.Errorf("Expected address to start with prefix 'a', got: %s", output)
		}
	})
}