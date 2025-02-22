// File: cheeseburger/testgen/testgen.go
package testgen

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GenerateTests accepts the source code of a function along with optional context,
// calls the LLM to generate test code, cleans up the output, replaces placeholder module names,
// and then appends it to the specified test file.
func GenerateTests(functionCode, context, testFilePath string) error {
	// Build the prompt.
	prompt := BuildPrompt(functionCode, context)
	fmt.Println("Sending prompt to LLM:\n", prompt)

	// Call the LLM API to generate test code.
	generatedTest, err := CallLLM(prompt)
	if err != nil {
		return fmt.Errorf("error generating tests: %v", err)
	}

	// Clean up the generated test code:
	// Use a regular expression to locate the first occurrence of a line starting with "package".
	re := regexp.MustCompile(`(?m)^package\s`)
	loc := re.FindStringIndex(generatedTest)
	cleanedTest := generatedTest
	if loc != nil {
		// Trim off any content before the "package" declaration.
		cleanedTest = generatedTest[loc[0]:]
	} else {
		// If no package declaration is found, force a default one.
		cleanedTest = "package autotest\n\n" + generatedTest
	}

	// Replace any placeholder module name "your_project" with the actual module name.
	moduleName, err := getModuleName()
	if err == nil && moduleName != "" {
		cleanedTest = strings.ReplaceAll(cleanedTest, "your_project", moduleName)
	} else {
		// If unable to determine, log a warning.
		fmt.Println("Warning: could not determine module name from go.mod; leaving placeholders unchanged")
	}

	fmt.Println("Generated test code (cleaned):\n", cleanedTest)

	// Write or append the cleaned test code to the test file.
	if err := appendToFile(testFilePath, cleanedTest); err != nil {
		return fmt.Errorf("error writing test code: %v", err)
	}

	fmt.Printf("Test code successfully written to %s\n", testFilePath)
	return nil
}

// appendToFile appends content to the file at filePath. It creates the file if it does not exist.
func appendToFile(filePath, content string) error {
	// Ensure the directory exists.
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Append to the file (or create it if not exists).
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString("\n\n" + content); err != nil {
		return err
	}
	return nil
}

// getModuleName reads the go.mod file from the project root and returns the module name.
func getModuleName() (string, error) {
	file, err := os.Open("go.mod")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			// The line should be like: "module github.com/yourusername/cheeseburger"
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("module name not found in go.mod")
}
