package service

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// testDbPath is used to override the default database path during tests
var testDbPath string

func init() {
	// Save original database path
	testDbPath = dbPath
}

func captureOutput(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the function
	f()

	// Restore stdout and close pipe
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func mockStdin(input string, f func()) {
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	// Write input in a goroutine to avoid blocking
	go func() {
		w.Write([]byte(input))
		w.Close()
	}()

	// Run the function
	f()

	// Restore stdin
	os.Stdin = oldStdin
}

func setupTestDB(t *testing.T) string {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	dbPath = filepath.Join(tmpDir, "test.db")
	t.Cleanup(func() {
		dbPath = testDbPath // Restore original path
	})
	return tmpDir
}

func TestHandleCommand(t *testing.T) {
	setupTestDB(t)

	tests := []struct {
		name           string
		args           []string
		expectedOutput string
		expectedExit   int
	}{
		{
			name:           "no arguments",
			args:           []string{},
			expectedOutput: "Usage: cheeseburger mvc\n\nCommands:",
			expectedExit:   1,
		},
		{
			name:           "help command",
			args:           []string{"help"},
			expectedOutput: "Usage: cheeseburger mvc\n\nCommands:",
			expectedExit:   0,
		},
		{
			name:           "unknown command",
			args:           []string{"unknown"},
			expectedOutput: "Unknown mvc command: unknown",
			expectedExit:   1,
		},
		{
			name:           "restore without file",
			args:           []string{"restore"},
			expectedOutput: "Error: backup file path required for restore",
			expectedExit:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var exitCode int
			oldOsExit := osExit
			defer func() { osExit = oldOsExit }()
			osExit = func(code int) {
				exitCode = code
				panic("exit")
			}

			output := captureOutput(func() {
				defer func() {
					if r := recover(); r != nil {
						if r != "exit" {
							panic(r)
						}
					}
				}()
				HandleCommand(tt.args)
			})

			assert.Contains(t, output, tt.expectedOutput)
			if tt.expectedExit > 0 {
				assert.Equal(t, tt.expectedExit, exitCode)
			}
		})
	}
}

func TestInitDb(t *testing.T) {
	setupTestDB(t)

	t.Run("initialize new database", func(t *testing.T) {
		output := captureOutput(func() {
			initDb()
		})

		assert.Contains(t, output, "Database initialized successfully")
		assert.DirExists(t, dbPath)
	})

	t.Run("initialize existing database", func(t *testing.T) {
		output := captureOutput(func() {
			initDb()
		})

		assert.Contains(t, output, "Database already exists")
	})
}

func TestClean(t *testing.T) {
	setupTestDB(t)

	t.Run("clean non-existent database", func(t *testing.T) {
		output := captureOutput(func() {
			clean()
		})

		assert.Contains(t, output, "Database is already clean")
	})

	t.Run("clean existing database - confirmed", func(t *testing.T) {
		// Create test database first
		initDb()
		assert.DirExists(t, dbPath)

		var output string
		// Mock user input "y" for confirmation
		mockStdin("y\n", func() {
			output = captureOutput(func() {
				clean()
			})
		})

		assert.Contains(t, output, "Database cleaned successfully")
		assert.NoDirExists(t, dbPath)
	})

	t.Run("clean existing database - cancelled", func(t *testing.T) {
		// Create test database first
		initDb()
		assert.DirExists(t, dbPath)

		var output string
		// Mock user input "n" for confirmation
		mockStdin("n\n", func() {
			output = captureOutput(func() {
				clean()
			})
		})

		assert.Contains(t, output, "Operation cancelled")
		assert.DirExists(t, dbPath)
	})
}

func TestBackup(t *testing.T) {
	setupTestDB(t)

	t.Run("backup non-existent database", func(t *testing.T) {
		output := captureOutput(func() {
			backup()
		})

		assert.Contains(t, output, "No database exists to backup")
	})

	t.Run("backup existing database", func(t *testing.T) {
		// Create and initialize test database
		initDb()
		assert.DirExists(t, dbPath)

		output := captureOutput(func() {
			backup()
		})

		assert.Contains(t, output, "Database backed up successfully")
		assert.DirExists(t, "data/backups")
	})
}

func TestRestore(t *testing.T) {
	tmpDir := setupTestDB(t)

	t.Run("restore non-existent backup", func(t *testing.T) {
		output := captureOutput(func() {
			restore("nonexistent.db")
		})

		assert.Contains(t, output, "Backup file does not exist")
	})

	t.Run("restore to clean state", func(t *testing.T) {
		// Create a test backup file
		backupFile := filepath.Join(tmpDir, "test_backup.db")
		err := os.WriteFile(backupFile, []byte("test backup data"), 0644)
		assert.NoError(t, err)

		output := captureOutput(func() {
			restore(backupFile)
		})

		assert.Contains(t, output, "Database restored successfully")
	})

	t.Run("restore with existing database - confirmed", func(t *testing.T) {
		// Create test database and backup
		initDb()
		backupFile := filepath.Join(tmpDir, "test_backup.db")
		err := os.WriteFile(backupFile, []byte("test backup data"), 0644)
		assert.NoError(t, err)

		var output string
		// Mock user input "y" for confirmation
		mockStdin("y\n", func() {
			output = captureOutput(func() {
				restore(backupFile)
			})
		})

		assert.Contains(t, output, "Database restored successfully")
	})

	t.Run("restore with existing database - cancelled", func(t *testing.T) {
		// Create test database and backup
		initDb()
		backupFile := filepath.Join(tmpDir, "test_backup.db")
		err := os.WriteFile(backupFile, []byte("test backup data"), 0644)
		assert.NoError(t, err)

		var output string
		// Mock user input "n" for confirmation
		mockStdin("n\n", func() {
			output = captureOutput(func() {
				restore(backupFile)
			})
		})

		assert.Contains(t, output, "Operation cancelled")
	})
}
