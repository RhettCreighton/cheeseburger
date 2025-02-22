// cheeseburger/coverage/parser.go
package coverage

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// FileCoverage holds the total and covered statements for a file.
type FileCoverage struct {
	FileName     string
	TotalStmts   int
	CoveredStmts int
}

// parseCoverageLine processes one line from the coverage profile.
// Expected format (after the first line):
//
//	<filename>:<startLine>.<startCol>,<endLine>.<endCol> <numStmts> <count>
//
// Example:
//
//	cheeseburger/app/models/comment.go:12.34,15.56 3 1
func parseCoverageLine(line string) (string, int, int, error) {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return "", 0, 0, fmt.Errorf("invalid coverage line: %s", line)
	}

	// The first part contains the file name and code range.
	fileAndRange := parts[0]

	// The following parts are: number of statements and hit count.
	stats := parts[1:]
	if len(stats) != 2 {
		return "", 0, 0, fmt.Errorf("unexpected stats format: %v", stats)
	}
	numStmts, err := strconv.Atoi(stats[0])
	if err != nil {
		return "", 0, 0, err
	}
	count, err := strconv.Atoi(stats[1])
	if err != nil {
		return "", 0, 0, err
	}

	// The file name is the part before the colon.
	fileParts := strings.Split(fileAndRange, ":")
	if len(fileParts) < 1 {
		return "", 0, 0, fmt.Errorf("invalid file part in line: %s", line)
	}
	fileName := fileParts[0]
	return fileName, numStmts, count, nil
}

// parseCoverageFile reads the coverage file and aggregates data per file.
func parseCoverageFile(filename string) (map[string]*FileCoverage, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	coverageData := make(map[string]*FileCoverage)
	firstLine := true

	for scanner.Scan() {
		line := scanner.Text()
		// Skip the first line which specifies the coverage mode (e.g., "mode: set").
		if firstLine {
			firstLine = false
			continue
		}

		fileName, numStmts, count, err := parseCoverageLine(line)
		if err != nil {
			fmt.Printf("Error parsing line: %v\n", err)
			continue
		}

		// Initialize or update the file's coverage data.
		fc, exists := coverageData[fileName]
		if !exists {
			fc = &FileCoverage{FileName: fileName}
			coverageData[fileName] = fc
		}
		fc.TotalStmts += numStmts
		// If count > 0, assume the statements are covered.
		if count > 0 {
			fc.CoveredStmts += numStmts
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return coverageData, nil
}

// RunCoverage is the entry point for the coverage command.
// It automatically runs "go test -coverprofile=<tempfile>" to generate a fresh coverage profile,
// parses the output, displays a summary, and then cleans up the temporary file.
func RunCoverage(args []string) {
	// Determine test packages (default: "./...")
	testArgs := []string{"test", "-coverprofile", ""}
	if len(args) >= 1 && args[0] != "" {
		// Allow user to specify packages or test arguments if needed.
		testArgs = append(testArgs, args...)
	} else {
		// Default test target.
		testArgs = append(testArgs, "./...")
	}

	// Create a temporary file for coverage output.
	tmpFile, err := os.CreateTemp("", "coverage-*.out")
	if err != nil {
		fmt.Printf("Error creating temporary coverage file: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(tmpFile.Name()) // Ensure it gets removed.
	tmpFile.Close()                 // We'll let the test command write to it.

	// Update the testArgs to include the temporary file name.
	testArgs[2] = tmpFile.Name()

	fmt.Printf("Running tests to generate coverage report...\n")
	cmd := exec.Command("go", testArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running tests: %v\n", err)
		os.Exit(1)
	}

	coverageData, err := parseCoverageFile(tmpFile.Name())
	if err != nil {
		fmt.Printf("Error parsing coverage file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Coverage Summary:")
	for _, data := range coverageData {
		percentage := 0.0
		if data.TotalStmts > 0 {
			percentage = (float64(data.CoveredStmts) / float64(data.TotalStmts)) * 100
		}
		fmt.Printf("%s: %.2f%% (%d/%d statements covered)\n", data.FileName, percentage, data.CoveredStmts, data.TotalStmts)
	}
}
