package testgen

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GenerateTests accepts the source code of a function along with optional context,
// calls the LLM to generate test code, cleans up the output, replaces placeholder module names,
// and then writes it to the specified test file, overwriting any existing content.
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
	// Locate the first occurrence of a line starting with "package".
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
		fmt.Println("Warning: could not determine module name from go.mod; leaving placeholders unchanged")
	}

	fmt.Println("Generated test code (cleaned):\n", cleanedTest)

	// Write the cleaned test code to the test file, overwriting any existing content.
	if err := writeToFile(testFilePath, cleanedTest); err != nil {
		return fmt.Errorf("error writing test code: %v", err)
	}

	fmt.Printf("Test code successfully written to %s\n", testFilePath)
	return nil
}

// writeToFile writes content to the file at filePath, ensuring the directory exists.
func writeToFile(filePath, content string) error {
	// Ensure the directory exists.
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, []byte(content), 0644)
}
