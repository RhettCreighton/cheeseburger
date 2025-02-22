// File: cheeseburger/testgen/prompt.go
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

// BuildPrompt builds an enhanced prompt from a function’s source code and optional context.
// It extracts dependency context (custom function signatures and documentation) to aid test generation.
func BuildPrompt(functionCode, context string) string {
	// Updated prompt instructions: note that uncovered branches and edge cases should be addressed.
	prompt := fmt.Sprintf(
		`Generate table-driven Go test cases for the following function.
Consider uncovered branch conditions, edge cases, and use subtests where applicable.
Include header comments in the test file that summarize which scenarios are being tested.
Only include custom functions in the dependency context; ignore built-ins.

Function:
%s`,
		functionCode,
	)
	if context != "" {
		prompt += fmt.Sprintf("\n\nAdditional context: %s", context)
	}

	// Append enhanced dependency context.
	var depContextLines []string
	depNames, err := extractCalledFunctions(functionCode)
	if err != nil {
		// If dependency extraction fails, simply note the error.
		prompt += fmt.Sprintf("\n\nError extracting dependencies: %v", err)
		return prompt
	}
	if len(depNames) == 0 {
		prompt += "\n\nNo custom dependencies found."
		return prompt
	}
	for _, depName := range depNames {
		// Skip built-in functions.
		if builtInFunctions[depName] {
			continue
		}
		depInfo, err := lookupDependencyDocumentation("cheeseburger", depName)
		if err != nil {
			// Skip dependencies not found.
			continue
		}
		depContextLines = append(depContextLines, fmt.Sprintf(
			"\nFunction %s (from %s):\nSignature: %s\nDocumentation: %s",
			depInfo.Name, depInfo.SourceFile, depInfo.Signature, depInfo.DocComment,
		))
	}
	if len(depContextLines) > 0 {
		prompt += "\n\n--- Dependency Context (custom functions only) ---"
		prompt += strings.Join(depContextLines, "\n")
	}
	return prompt
}

// extractCalledFunctions parses the given source code and returns a list of function names called within.
func extractCalledFunctions(source string) ([]string, error) {
	fset := token.NewFileSet()
	// Wrap the source code in a dummy package.
	file, err := parser.ParseFile(fset, "target.go", "package main\n"+source, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing source: %w", err)
	}
	functions := make(map[string]bool)
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fun := call.Fun.(type) {
		case *ast.Ident:
			if !builtInFunctions[fun.Name] {
				functions[fun.Name] = true
			}
		case *ast.SelectorExpr:
			if fun.Sel != nil && !builtInFunctions[fun.Sel.Name] {
				functions[fun.Sel.Name] = true
			}
		}
		return true
	})
	var deps []string
	for name := range functions {
		deps = append(deps, name)
	}
	return deps, nil
}

// lookupDependencyDocumentation searches the given project root for a Go file that defines funcName,
// then extracts its documentation and signature.
func lookupDependencyDocumentation(rootDir, funcName string) (*DependencyInfo, error) {
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
			doc := ""
			if funcDecl.Doc != nil {
				doc = funcDecl.Doc.Text()
			}
			var sigBuf bytes.Buffer
			sigBuf.WriteString("func ")
			sigBuf.WriteString(funcDecl.Name.Name)
			sigBuf.WriteString("(")
			if funcDecl.Type.Params != nil {
				var params []string
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
				var results []string
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
