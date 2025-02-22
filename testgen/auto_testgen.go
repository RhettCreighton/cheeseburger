// File: cheeseburger/testgen/auto_testgen.go
package testgen

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// CoverageTarget represents a single uncovered function region from the JSON coverage output.
type CoverageTarget struct {
	File             string `json:"file"`
	Function         string `json:"function"`
	FuncStart        int    `json:"funcStart"`
	FuncEnd          int    `json:"funcEnd"`
	UncoveredRegions []struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"uncoveredRegions"`
}

// AutoGenerateTests reads the coverage JSON file, extracts functions with low coverage,
// and for each function, extracts its source code and passes it to GenerateTests.
// In this limited prototype, only the first uncovered target is processed.
func AutoGenerateTests(coverageJSONPath, testOutputDir string) error {
	// Remove any existing output directory.
	if err := os.RemoveAll(testOutputDir); err != nil {
		return fmt.Errorf("failed to remove existing output directory %s: %v", testOutputDir, err)
	}
	// Create a fresh output directory.
	if err := os.MkdirAll(testOutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %v", testOutputDir, err)
	}

	// Read the coverage JSON file.
	rawData, err := ioutil.ReadFile(coverageJSONPath)
	if err != nil {
		return fmt.Errorf("failed to read coverage JSON: %v", err)
	}

	// Filter out any non-JSON preamble by finding the first '['.
	dataStr := string(rawData)
	startIdx := strings.Index(dataStr, "[")
	if startIdx == -1 {
		return fmt.Errorf("no JSON array found in coverage output")
	}
	dataStr = dataStr[startIdx:]
	data := []byte(dataStr)

	var targets []CoverageTarget
	if err := json.Unmarshal(data, &targets); err != nil {
		return fmt.Errorf("failed to parse coverage JSON: %v", err)
	}

	if len(targets) == 0 {
		return fmt.Errorf("no coverage targets found")
	}

	// Process only the first uncovered target for a controlled test.
	target := targets[0]
	fmt.Printf("Processing %s:%s (lines %d-%d)\n", target.File, target.Function, target.FuncStart, target.FuncEnd)

	functionCode, err := extractFunctionCode(target.File, target.Function, target.FuncStart, target.FuncEnd)
	if err != nil {
		return fmt.Errorf("could not extract function code for %s: %v", target.Function, err)
	}

	// Build additional context for the prompt (e.g., which lines are uncovered)
	var regions []string
	for _, r := range target.UncoveredRegions {
		regions = append(regions, fmt.Sprintf("lines %d-%d", r.Start, r.End))
	}
	context := fmt.Sprintf("Uncovered regions: %s", strings.Join(regions, ", "))

	// Determine the test file path.
	// We create a default file name in the output directory based on the base file name.
	testFilePath := filepath.Join(testOutputDir, filepath.Base(target.File))
	// Optionally, you could append "_test.go" if desired.

	// Generate tests for the function.
	if err := GenerateTests(functionCode, context, testFilePath); err != nil {
		return fmt.Errorf("error generating tests for %s: %v", target.Function, err)
	}

	fmt.Printf("Tests generated for %s in %s\n", target.Function, testFilePath)
	return nil
}

// extractFunctionCode uses a simple line-based extraction to return the source code
// of the function from filePath between funcStart and funcEnd.
// If reading the file fails with a "not a directory" error (or file not found),
// it will try stripping a "cheeseburger/" prefix.
func extractFunctionCode(filePath, funcName string, funcStart, funcEnd int) (string, error) {
	// Resolve to an absolute path.
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path for %s: %v", filePath, err)
	}
	// Try reading the file.
	data, err := ioutil.ReadFile(absPath)
	if err != nil && (os.IsNotExist(err) || strings.Contains(err.Error(), "not a directory")) {
		// Fallback: if the path starts with "cheeseburger/", remove that prefix.
		if strings.HasPrefix(filePath, "cheeseburger/") {
			adjustedPath := strings.TrimPrefix(filePath, "cheeseburger/")
			absPath, err = filepath.Abs(adjustedPath)
			if err != nil {
				return "", fmt.Errorf("failed to resolve absolute path for adjusted %s: %v", adjustedPath, err)
			}
			data, err = ioutil.ReadFile(absPath)
			if err != nil {
				return "", fmt.Errorf("file not found at both %s and %s: %v", filePath, adjustedPath, err)
			}
		} else {
			return "", fmt.Errorf("failed to read file %s: %v", absPath, err)
		}
	} else if err != nil {
		return "", fmt.Errorf("failed to read file %s: %v", absPath, err)
	}

	lines := strings.Split(string(data), "\n")
	if funcStart-1 < 0 || funcEnd > len(lines) {
		return "", fmt.Errorf("invalid line numbers: %d-%d", funcStart, funcEnd)
	}

	functionLines := lines[funcStart-1 : funcEnd]
	return strings.Join(functionLines, "\n"), nil
}
