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
// In this prototype, only the first uncovered target is processed.
func AutoGenerateTests(coverageJSONPath, testOutputDir string) error {
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

	// Process only the first uncovered target.
	target := targets[0]
	fmt.Printf("Processing %s:%s (lines %d-%d)\n", target.File, target.Function, target.FuncStart, target.FuncEnd)

	// Adjust the file path if the first component is not a directory.
	parts := strings.Split(target.File, string(filepath.Separator))
	if len(parts) > 0 {
		if info, err := os.Stat(parts[0]); err == nil && !info.IsDir() {
			// Remove the first component.
			target.File = strings.Join(parts[1:], string(filepath.Separator))
		}
	}

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
	// If testOutputDir is provided, use it; otherwise, write the test file in the same directory as the source file.
	var testFilePath string
	if testOutputDir != "" {
		testFilePath = filepath.Join(testOutputDir, strings.TrimSuffix(filepath.Base(target.File), ".go")+"_autotest_test.go")
	} else {
		testFilePath = filepath.Join(filepath.Dir(target.File), strings.TrimSuffix(filepath.Base(target.File), ".go")+"_autotest_test.go")
	}

	// Generate tests for the function.
	if err := GenerateTests(functionCode, context, testFilePath); err != nil {
		return fmt.Errorf("error generating tests for %s: %v", target.Function, err)
	}

	fmt.Printf("Tests generated for %s in %s\n", target.Function, testFilePath)
	return nil
}

// extractFunctionCode uses a simple line-based extraction to return the source code
// of the function from filePath between funcStart and funcEnd.
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
