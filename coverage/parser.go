// cheeseburger/coverage/parser.go
package coverage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// LineRange represents a range of lines in a file.
type LineRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// mergeLineRanges merges overlapping or adjacent line ranges.
func mergeLineRanges(ranges []LineRange) []LineRange {
	if len(ranges) == 0 {
		return ranges
	}
	// Sort by start line.
	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].Start < ranges[j].Start
	})
	merged := []LineRange{ranges[0]}
	for _, r := range ranges[1:] {
		last := &merged[len(merged)-1]
		// If ranges overlap or are adjacent, merge them.
		if r.Start <= last.End+1 {
			if r.End > last.End {
				last.End = r.End
			}
		} else {
			merged = append(merged, r)
		}
	}
	return merged
}

// FileCoverage holds the total and covered statements for a file,
// and also records uncovered line ranges.
type FileCoverage struct {
	FileName     string
	TotalStmts   int
	CoveredStmts int
	Uncovered    []LineRange
}

// parseCoverageLine processes one line from the coverage profile.
// Expected format (after the first line):
//
//	<filename>:<startLine>.<startCol>,<endLine>.<endCol> <numStmts> <count>
//
// Example:
//
//	cheeseburger/app/models/comment.go:12.34,15.56 3 1
//
// It returns the file name, start line, end line, number of statements,
// hit count, and error if any.
func parseCoverageLine(line string) (string, int, int, int, int, error) {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return "", 0, 0, 0, 0, fmt.Errorf("invalid coverage line: %s", line)
	}

	// The first part contains the file name and code range.
	fileAndRange := parts[0]

	// The following parts are: number of statements and hit count.
	stats := parts[1:]
	if len(stats) != 2 {
		return "", 0, 0, 0, 0, fmt.Errorf("unexpected stats format: %v", stats)
	}
	numStmts, err := strconv.Atoi(stats[0])
	if err != nil {
		return "", 0, 0, 0, 0, err
	}
	count, err := strconv.Atoi(stats[1])
	if err != nil {
		return "", 0, 0, 0, 0, err
	}

	// Split the file part to get the file name and range.
	fileParts := strings.Split(fileAndRange, ":")
	if len(fileParts) < 2 {
		return "", 0, 0, 0, 0, fmt.Errorf("invalid file part in line: %s", line)
	}
	fileName := fileParts[0]
	rangePart := fileParts[1] // expected format: "12.34,15.56"
	rangeParts := strings.Split(rangePart, ",")
	if len(rangeParts) != 2 {
		return "", 0, 0, 0, 0, fmt.Errorf("invalid range format in line: %s", line)
	}

	// Extract start line.
	startTokens := strings.Split(rangeParts[0], ".")
	startLine, err := strconv.Atoi(startTokens[0])
	if err != nil {
		return "", 0, 0, 0, 0, fmt.Errorf("invalid start line in line: %s", line)
	}

	// Extract end line.
	endTokens := strings.Split(rangeParts[1], ".")
	endLine, err := strconv.Atoi(endTokens[0])
	if err != nil {
		return "", 0, 0, 0, 0, fmt.Errorf("invalid end line in line: %s", line)
	}

	return fileName, startLine, endLine, numStmts, count, nil
}

var resolvedPathCache = make(map[string]string)

// resolveSourcePath checks if the given file path exists. If not,
// it attempts to remove a "cheeseburger/" prefix and checks again.
func resolveSourcePath(filePath string) string {
	if cached, ok := resolvedPathCache[filePath]; ok {
		return cached
	}
	resolved := filePath
	if _, err := os.Stat(filePath); err != nil {
		const prefix = "cheeseburger/"
		if strings.HasPrefix(filePath, prefix) {
			altPath := strings.TrimPrefix(filePath, prefix)
			if _, err := os.Stat(altPath); err == nil {
				resolved = altPath
			}
		}
	}
	resolvedPathCache[filePath] = resolved
	return resolved
}

// parseCoverageFile reads the coverage file and aggregates data per file.
// It also collects the uncovered line ranges (where count == 0).
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

		fileName, startLine, endLine, numStmts, count, err := parseCoverageLine(line)
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
		} else {
			// Record the uncovered range.
			fc.Uncovered = append(fc.Uncovered, LineRange{Start: startLine, End: endLine})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return coverageData, nil
}

// printCoverageSummary prints the overall coverage for each file.
func printCoverageSummary(coverageData map[string]*FileCoverage) {
	// Gather the keys so we can sort the output.
	var files []string
	for f := range coverageData {
		files = append(files, f)
	}
	sort.Strings(files)

	fmt.Println("Coverage Summary:")
	for _, file := range files {
		data := coverageData[file]
		percentage := 0.0
		if data.TotalStmts > 0 {
			percentage = (float64(data.CoveredStmts) / float64(data.TotalStmts)) * 100
		}
		fmt.Printf("%s: %.2f%% (%d/%d statements covered)\n",
			data.FileName, percentage, data.CoveredStmts, data.TotalStmts)
	}
}

// printLowCoverage prints a list of files with coverage below a specified threshold.
func printLowCoverage(coverageData map[string]*FileCoverage, threshold float64) {
	var lowCoverageFiles []string
	for _, data := range coverageData {
		percentage := 0.0
		if data.TotalStmts > 0 {
			percentage = (float64(data.CoveredStmts) / float64(data.TotalStmts)) * 100
		}
		if percentage < threshold {
			lowCoverageFiles = append(lowCoverageFiles, fmt.Sprintf("%s: %.2f%%", data.FileName, percentage))
		}
	}

	if len(lowCoverageFiles) > 0 {
		fmt.Printf("\nFiles with coverage below %.2f%%:\n", threshold)
		sort.Strings(lowCoverageFiles)
		for _, entry := range lowCoverageFiles {
			fmt.Println(entry)
		}
	} else {
		fmt.Printf("\nAll files meet the %.2f%% coverage threshold.\n", threshold)
	}
}

// rangesOverlap checks if two line ranges overlap.
func rangesOverlap(aStart, aEnd, bStart, bEnd int) bool {
	return aStart <= bEnd && bStart <= aEnd
}

// Target represents a function (or method) along with its uncovered regions.
type Target struct {
	FileName         string      `json:"file"`
	FuncName         string      `json:"function"`
	FuncStart        int         `json:"funcStart"`
	FuncEnd          int         `json:"funcEnd"`
	UncoveredRegions []LineRange `json:"uncoveredRegions"`
}

// collectTargets traverses the AST and collects functions that have uncovered regions.
func collectTargets(fileName string, fset *token.FileSet, node ast.Node, mergedRanges []LineRange) []Target {
	var targets []Target
	ast.Inspect(node, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}
		fnStart := fset.Position(fn.Pos()).Line
		fnEnd := fset.Position(fn.End()).Line
		var overlapping []LineRange
		for _, ur := range mergedRanges {
			if rangesOverlap(fnStart, fnEnd, ur.Start, ur.End) {
				overlapping = append(overlapping, ur)
			}
		}
		if len(overlapping) > 0 {
			targets = append(targets, Target{
				FileName:         fileName,
				FuncName:         fn.Name.Name,
				FuncStart:        fnStart,
				FuncEnd:          fnEnd,
				UncoveredRegions: overlapping,
			})
		}
		return true
	})
	return targets
}

// printTargets outputs the target functions. If asJSON is true, it prints in JSON format.
func printTargets(targets []Target, asJSON bool) {
	if asJSON {
		data, err := json.MarshalIndent(targets, "", "  ")
		if err != nil {
			fmt.Printf("Error marshalling targets: %v\n", err)
			return
		}
		fmt.Println(string(data))
	} else {
		fmt.Println("\nUncovered Regions Mapped to Functions:")
		for _, t := range targets {
			fmt.Printf("File %s: Function %s (lines %d-%d) has uncovered regions: ",
				t.FileName, t.FuncName, t.FuncStart, t.FuncEnd)
			for i, r := range t.UncoveredRegions {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("%d-%d", r.Start, r.End)
			}
			fmt.Println()
		}
	}
}

// printUncoveredFunctions maps uncovered regions to functions using static analysis.
// It now supports an optional JSON output mode if the "--json" flag is passed.
func printUncoveredFunctions(coverageData map[string]*FileCoverage, asJSON bool) {
	var allTargets []Target
	fmt.Println()
	for fileName, fc := range coverageData {
		if len(fc.Uncovered) == 0 {
			continue
		}
		mergedRanges := mergeLineRanges(fc.Uncovered)
		resolvedPath := resolveSourcePath(fileName)
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, resolvedPath, nil, 0)
		if err != nil {
			fmt.Printf("Error parsing source file %s: %v\n", fileName, err)
			continue
		}
		targets := collectTargets(fileName, fset, node, mergedRanges)
		allTargets = append(allTargets, targets...)
	}
	printTargets(allTargets, asJSON)
}

// RunCoverage is the entry point for the coverage command.
// It runs "go test -coverprofile=<tempfile>" to generate a coverage profile,
// parses the output, displays a summary, and then outputs the uncovered function targets.
func RunCoverage(args []string) {
	// Check for a JSON output flag.
	jsonOutput := false
	var remainingArgs []string
	for _, arg := range args {
		if arg == "--json" {
			jsonOutput = true
		} else {
			remainingArgs = append(remainingArgs, arg)
		}
	}

	// Determine test packages (default: "./...")
	testArgs := []string{"test", "-coverprofile", ""}
	if len(remainingArgs) >= 1 && remainingArgs[0] != "" {
		// Allow user to specify packages or test arguments if needed.
		testArgs = append(testArgs, remainingArgs...)
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
	tmpFile.Close()                 // Let the test command write to it.

	// Update the testArgs to include the temporary file name.
	testArgs[2] = tmpFile.Name()

	fmt.Println("Running tests to generate coverage report...")
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

	printCoverageSummary(coverageData)
	// Print files with less than 80% coverage (threshold can be adjusted).
	printLowCoverage(coverageData, 80.0)
	// Map uncovered regions to functions, using the JSON flag if set.
	printUncoveredFunctions(coverageData, jsonOutput)
}
