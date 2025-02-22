// File: cheeseburger/testgen/context_extractor.go
package testgen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
)

// builtInFunctions is a set of known built-in functions and common standard library functions to ignore.
var builtInFunctions = map[string]bool{
	"append":  true,
	"cap":     true,
	"close":   true,
	"complex": true,
	"copy":    true,
	"delete":  true,
	"imag":    true,
	"len":     true,
	"make":    true,
	"new":     true,
	"panic":   true,
	"print":   true,
	"println": true,
	"real":    true,
	"recover": true,
	// Optionally, add more common functions if needed.
}

// DependencyInfo holds extracted context for a dependency.
type DependencyInfo struct {
	Name       string
	DocComment string
	Signature  string
	SourceFile string
}

// ExtractCalledFunctions parses the given source code and returns a list
// of function names that are called inside it, excluding built-in and standard library functions.
func ExtractCalledFunctions(source string) ([]string, error) {
	fset := token.NewFileSet()
	// Wrap the source code in a dummy package.
	file, err := parser.ParseFile(fset, "target.go", "package main\n"+source, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing source: %w", err)
	}

	functions := make(map[string]bool)
	ast.Inspect(file, func(n ast.Node) bool {
		// Look for call expressions.
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		// Identify the called function.
		switch fun := call.Fun.(type) {
		case *ast.Ident:
			if !builtInFunctions[fun.Name] {
				functions[fun.Name] = true
			}
		case *ast.SelectorExpr:
			// For calls like pkg.Func, record the function name.
			if fun.Sel != nil && !builtInFunctions[fun.Sel.Name] {
				functions[fun.Sel.Name] = true
			}
		}
		return true
	})

	// Convert map keys to slice.
	var deps []string
	for name := range functions {
		deps = append(deps, name)
	}
	return deps, nil
}

// LookupDependencyDocumentation searches the project directory for the given function name,
// and if found, extracts its documentation and signature.
func LookupDependencyDocumentation(rootDir, funcName string) (*DependencyInfo, error) {
	var depInfo *DependencyInfo
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			// Skip files that cannot be parsed.
			return nil
		}

		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Name.Name != funcName {
				continue
			}

			// Get documentation comment if available.
			doc := ""
			if funcDecl.Doc != nil {
				doc = funcDecl.Doc.Text()
			}

			// Build a simple signature.
			var sigBuf bytes.Buffer
			sigBuf.WriteString("func ")
			sigBuf.WriteString(funcDecl.Name.Name)
			sigBuf.WriteString("(")
			if funcDecl.Type.Params != nil {
				params := []string{}
				for _, field := range funcDecl.Type.Params.List {
					typeStr := fmt.Sprintf("%s", field.Type)
					var names []string
					for _, name := range field.Names {
						names = append(names, name.Name)
					}
					if len(names) > 0 {
						params = append(params, fmt.Sprintf("%s %s", strings.Join(names, ", "), typeStr))
					} else {
						params = append(params, typeStr)
					}
				}
				sigBuf.WriteString(strings.Join(params, ", "))
			}
			sigBuf.WriteString(")")
			if funcDecl.Type.Results != nil {
				results := []string{}
				for _, field := range funcDecl.Type.Results.List {
					typeStr := fmt.Sprintf("%s", field.Type)
					if len(field.Names) > 0 {
						var names []string
						for _, name := range field.Names {
							names = append(names, name.Name)
						}
						results = append(results, fmt.Sprintf("%s %s", strings.Join(names, ", "), typeStr))
					} else {
						results = append(results, typeStr)
					}
				}
				sigBuf.WriteString(" ")
				if len(results) > 1 {
					sigBuf.WriteString("(" + strings.Join(results, ", ") + ")")
				} else {
					sigBuf.WriteString(strings.Join(results, ", "))
				}
			}

			depInfo = &DependencyInfo{
				Name:       funcName,
				DocComment: strings.TrimSpace(doc),
				Signature:  sigBuf.String(),
				SourceFile: path,
			}
			// Found the first match; stop walking.
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if depInfo == nil {
		return nil, fmt.Errorf("function %s not found", funcName)
	}
	return depInfo, nil
}

// BuildEnhancedPrompt builds a prompt that includes the original function source
// plus additional context for each non-built-in dependency found in the code.
func BuildEnhancedPrompt(functionCode, context, projectRoot string) (string, error) {
	// Instruct the LLM to produce table-driven tests with header comments
	// that explain which uncovered branches or edge cases are being tested.
	prompt := fmt.Sprintf(
		`Generate table-driven Go test cases for the following function.
Consider edge cases and uncovered branch conditions.
Use subtests and include header comments that describe the purpose of each test case.
Ignore standard library functions; focus on custom logic only.

Function:
%s`,
		functionCode,
	)
	if context != "" {
		prompt += fmt.Sprintf("\n\nAdditional context: %s", context)
	}

	// Extract dependency names.
	depNames, err := ExtractCalledFunctions(functionCode)
	if err != nil {
		return "", fmt.Errorf("failed to extract dependencies: %w", err)
	}
	if len(depNames) == 0 {
		return prompt, nil
	}

	// Append dependency context only for dependencies we successfully find.
	var depContextLines []string
	for _, depName := range depNames {
		depInfo, err := LookupDependencyDocumentation(projectRoot, depName)
		if err != nil {
			// Skip dependencies that are not found.
			continue
		}
		depContextLines = append(depContextLines, fmt.Sprintf(
			"\nFunction %s (from %s):\nSignature: %s\nDocumentation: %s",
			depInfo.Name, depInfo.SourceFile, depInfo.Signature, depInfo.DocComment,
		))
	}
	if len(depContextLines) > 0 {
		prompt += "\n\n--- Dependency Context (only custom functions) ---"
		prompt += strings.Join(depContextLines, "\n")
	}

	return prompt, nil
}
