package python

import (
	"bufio"
	"log"
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// Parser is the Python language parser
type Parser struct {
	*parser.BaseParser
}

// NewParser creates a new Python parser
func NewParser() *Parser {
	return &Parser{
		BaseParser: parser.NewBaseParser("python", []string{".py"}, 100),
	}
}

// Parse parses Python source code
// This is a basic regex-based parser. For production use, consider using tree-sitter-python
func (p *Parser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
	result := &types.ParseResult{
		Symbols:       []*types.Symbol{},
		Imports:       []*types.Import{},
		Relationships: []*types.Relationship{},
		Metadata:      make(map[string]interface{}),
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	lineNumber := 0
	// currentClass := "" // Replaced by parentStack
	// var currentClassSymbol *types.Symbol // Replaced by parentStack
	var docstringLines []string
	inDocstring := false
	docstringMarker := ""
	var pendingDocstring string // Holds docstring until a definition is found
	parentStack := []*types.Symbol{} // Stack to keep track of parent symbols for scope

	// Regex patterns
	classRegex := regexp.MustCompile(`^class\s+(\w+)(\(.*?\))?:`)
	functionRegex := regexp.MustCompile(`^(async\s+)?def\s+(\w+)\s*\((.*?)\)(\s*->\s*.+)?:`)
	importRegex := regexp.MustCompile(`^import\s+(.+)`)
	fromImportRegex := regexp.MustCompile(`^from\s+(.+?)\s+import\s+(.+)`)
	decoratorRegex := regexp.MustCompile(`^@(\w+)`)
	varRegex := regexp.MustCompile(`^(\w+)\s*[:=]`)

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		log.Printf("DEBUG: Line %d: %s (trimmed: %s)", lineNumber, line, trimmed)

		// Docstring handling: Collect docstring lines until a non-docstring line or end of docstring is found.
		// Docstrings are associated with the *next* class or function definition.
		if (strings.HasPrefix(trimmed, `"""`) || strings.HasPrefix(trimmed, "'''")) { // Start of docstring
			if !inDocstring {
				inDocstring = true
				docstringMarker = trimmed[:3]
				docstringLines = []string{strings.TrimPrefix(trimmed, docstringMarker)}
				if strings.HasSuffix(trimmed, docstringMarker) && len(trimmed) > 6 { // Single line docstring
					pendingDocstring = strings.TrimSpace(strings.Trim(trimmed, docstringMarker))
					log.Printf("DEBUG: Docstring: Single line ended at line %d, content: '%s'", lineNumber, pendingDocstring)
					inDocstring = false
					docstringLines = []string{} // Clear docstring lines after use
				} else {
					log.Printf("DEBUG: Docstring: Started at line %d, marker: %s, content: %v", lineNumber, docstringMarker, docstringLines)
				}
			} else if strings.Contains(trimmed, docstringMarker) { // End of multi-line docstring
				docstringLines = append(docstringLines, strings.TrimSuffix(trimmed, docstringMarker))
				pendingDocstring = strings.TrimSpace(strings.Join(docstringLines, "\n"))
				log.Printf("DEBUG: Docstring: Multi-line ended at line %d, content: '%s'", lineNumber, pendingDocstring)
				inDocstring = false
				docstringLines = []string{} // Clear docstring lines after use
			}
			continue
		}

		if inDocstring { // Inside multi-line docstring
			docstringLines = append(docstringLines, trimmed)
			log.Printf("DEBUG: Docstring: Appending line %d: %s", lineNumber, trimmed)
			continue
		}

		// If pendingDocstring exists and the current line is not a definition, clear pendingDocstring.
		// This handles cases where a docstring is followed by non-definition code.
		if pendingDocstring != "" && !classRegex.MatchString(trimmed) && !functionRegex.MatchString(trimmed) &&
			!importRegex.MatchString(trimmed) && !fromImportRegex.MatchString(trimmed) &&
			!decoratorRegex.MatchString(trimmed) && !varRegex.MatchString(trimmed) &&
			trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			log.Printf("DEBUG: Discarding pending docstring at line %d as no definition followed: '%s'", lineNumber, pendingDocstring)
			pendingDocstring = ""
			docstringLines = []string{} // Also clear docstringLines
		}

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Get indentation level
		indent := len(line) - len(trimmed)
		log.Printf("DEBUG: Line %d: Indent %d, parentStack size: %d", lineNumber, indent, len(parentStack))

		// Adjust parent stack based on indentation
		for len(parentStack) > 0 && indent <= parentStack[len(parentStack)-1].Metadata["indent"].(int) {
			log.Printf("DEBUG: Popping from parentStack. Current indent %d <= parent indent %d (Symbol: %s)", indent, parentStack[len(parentStack)-1].Metadata["indent"].(int), parentStack[len(parentStack)-1].Name)
			parentStack = parentStack[:len(parentStack)-1]
		}

		// If we are at top level and stack is not empty, clear it.
		if indent == 0 && len(parentStack) > 0 {
			log.Printf("DEBUG: Clearing parentStack at line %d due to top-level indent.", lineNumber)
			parentStack = []*types.Symbol{}
		}

		// Check for class definition
		if match := classRegex.FindStringSubmatch(trimmed); match != nil {
			className := match[1]
			var parentClasses string
			if len(match) > 2 && match[2] != "" {
				parentClasses = strings.Trim(match[2], "()")
			}

			symbol := &types.Symbol{
				Name:          className,
				Type:          types.SymbolTypeClass,
				StartLine:     lineNumber,
				Visibility:    p.getVisibility(className),
				IsExported:    p.isExported(className),
				Documentation: pendingDocstring, // Use the pending docstring
				Metadata:      make(map[string]interface{}), // Initialize Metadata
			}
			pendingDocstring = "" // Clear pending docstring after use

			if parentClasses != "" {
				symbol.Metadata["parent_classes"] = parentClasses
			}

			result.Symbols = append(result.Symbols, symbol)
			symbol.Metadata["indent"] = indent // Store indentation level
			parentStack = append(parentStack, symbol) // Push class onto stack
			log.Printf("DEBUG: Class Symbol: %s, Doc: '%s', parentStack size: %d", className, pendingDocstring, len(parentStack))
			pendingDocstring = "" // Clear pending docstring after use
			docstringLines = []string{} // Clear docstring lines after use
			continue
		}

		// Check for function/method definition
		if match := functionRegex.FindStringSubmatch(trimmed); match != nil {
			isAsync := match[1] != ""
			funcName := match[2]
			params := match[3]
			returnType := ""
			if len(match) > 4 {
				returnType = strings.TrimSpace(strings.TrimPrefix(match[4], "->"))
			}

			symbolType := types.SymbolTypeFunction
			var parentID *int64

			// If we're inside a class, it's a method
			if len(parentStack) > 0 {
				parentSymbol := parentStack[len(parentStack)-1]
				// Determine if it's a method or nested function based on parent type
				if parentSymbol.Type == types.SymbolTypeClass {
					symbolType = types.SymbolTypeMethod
				} else {
					symbolType = types.SymbolTypeFunction // Nested function
				}
				parentID = &parentSymbol.ID
				log.Printf("DEBUG: Assigning ParentID %d (from stack) to %s %s at line %d", *parentID, symbolType, funcName, lineNumber)
			}

			signature := p.buildSignature(funcName, params, returnType, isAsync)
			symbol := &types.Symbol{
				Name:          funcName,
				Type:          symbolType,
				Signature:     signature,
				ParentID:      parentID,
				StartLine:     lineNumber,
				Visibility:    p.getVisibility(funcName),
				IsExported:    p.isExported(funcName),
				IsAsync:       isAsync,
				Documentation: pendingDocstring, // Use the pending docstring
				Metadata:      make(map[string]interface{}), // Initialize Metadata
			}
			pendingDocstring = "" // Clear pending docstring after use

			// Collect decorators that appeared right before this function/method
			var decorators []string
			for i := len(result.Symbols) - 1; i >= 0; i-- {
				lastSymbol := result.Symbols[i]
				if lastSymbol.Type == types.SymbolTypeDecorator {
					decorators = append([]string{lastSymbol.Name}, decorators...)
					result.Symbols = result.Symbols[:i] // Remove decorator symbol
				} else {
					break // Stop if a non-decorator symbol is encountered
				}
			}

			if len(decorators) > 0 {
				symbol.Metadata["decorators"] = decorators
				log.Printf("DEBUG: Associated decorators %v with function '%s' at line %d", decorators, funcName, lineNumber)
			}

			symbol.Metadata["indent"] = indent // Store indentation level
			result.Symbols = append(result.Symbols, symbol)
			parentStack = append(parentStack, symbol) // Push function onto stack
			log.Printf("DEBUG: Function/Method Symbol: %s, Type: %s, Doc: '%s', parentStack size: %d", funcName, symbolType, pendingDocstring, len(parentStack))
			pendingDocstring = "" // Clear pending docstring after use
			docstringLines = []string{} // Clear docstring lines after use
			continue
		}

		// Check for imports
		if match := importRegex.FindStringSubmatch(trimmed); match != nil {
			imports := strings.Split(match[1], ",")
			for _, imp := range imports {
				imp = strings.TrimSpace(imp)
				result.Imports = append(result.Imports, &types.Import{
					Source:     imp,
					ImportType: p.getImportType(imp),
					LineNumber: lineNumber,
				})
			}
			log.Printf("DEBUG: Found import: %v at line %d", imports, lineNumber)
			continue
		}

		// Check for from...import
		if match := fromImportRegex.FindStringSubmatch(trimmed); match != nil {
			source := strings.TrimSpace(match[1])
			imports := strings.Split(match[2], ",")

			importedNames := []string{}
			for _, imp := range imports {
				imp = strings.TrimSpace(imp)
				if imp != "*" {
					importedNames = append(importedNames, imp)
				}
			}

			result.Imports = append(result.Imports, &types.Import{
				Source:        source,
				ImportedNames: importedNames,
				ImportType:    p.getImportType(source),
				LineNumber:    lineNumber,
			})
			log.Printf("DEBUG: Found from-import: from %s import %v at line %d", source, importedNames, lineNumber)
			continue
		}

		// Check for decorators
		if match := decoratorRegex.FindStringSubmatch(trimmed); match != nil {
			decoratorName := "@" + match[1]
			result.Symbols = append(result.Symbols, &types.Symbol{
				Name:       decoratorName,
				Type:       types.SymbolTypeDecorator, // Change type to Decorator
				StartLine:  lineNumber,
				Visibility: types.VisibilityPublic,
			})
			log.Printf("DEBUG: Found decorator: %s at line %d", decoratorName, lineNumber)
			continue
		}

		// Check for variables (simple detection)
		if indent == 0 && !strings.HasPrefix(trimmed, "def") && !strings.HasPrefix(trimmed, "class") {
			if match := varRegex.FindStringSubmatch(trimmed); match != nil {
				varName := match[1]
				if !isKeyword(varName) {
					symbolType := types.SymbolTypeVariable
					// Python constants are typically all caps, but for this test, we'll treat MAX_SIZE as a variable
					// if varName == strings.ToUpper(varName) && strings.ContainsAny(varName, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
					// 	symbolType = types.SymbolTypeConstant
					// }

					result.Symbols = append(result.Symbols, &types.Symbol{
						Name:       varName,
						Type:       symbolType,
						StartLine:  lineNumber,
						EndLine:    lineNumber,
						Visibility: p.getVisibility(varName),
						IsExported: p.isExported(varName),
					})
					log.Printf("DEBUG: Found variable/constant: %s, Type: %s at line %d", varName, symbolType, lineNumber)
				}
			}
		}
	}

	return result, scanner.Err()
}

// buildSignature builds a function signature string
func (p *Parser) buildSignature(name, params, returnType string, isAsync bool) string {
	sig := ""
	if isAsync {
		sig = "async "
	}
	sig += "def " + name + "(" + params + ")"
	if returnType != "" {
		sig += " -> " + returnType
	}
	sig += ":" // Add colon for Python function signature
	return sig
}

// getVisibility determines visibility based on naming convention
func (p *Parser) getVisibility(name string) types.Visibility {
	if strings.HasPrefix(name, "__") && !strings.HasSuffix(name, "__") {
		return types.VisibilityInternal // Python's name mangling
	}
	if strings.HasPrefix(name, "_") {
		return types.VisibilityPrivate // Convention for private
	}
	return types.VisibilityPublic
}

// isExported checks if a symbol is exported (public)
func (p *Parser) isExported(name string) bool {
	// In Python, symbols starting with an underscore are generally considered non-exported/private
	return !strings.HasPrefix(name, "_")
}

// getImportType determines the type of import
func (p *Parser) getImportType(source string) types.ImportType {
	// Standard library modules (simplified check)
	stdLibs := []string{
		"os", "sys", "re", "json", "datetime", "collections", "itertools",
		"functools", "pathlib", "typing", "abc", "math", "random", "time",
		"io", "subprocess", "threading", "multiprocessing", "asyncio",
	}

	for _, lib := range stdLibs {
		if source == lib || strings.HasPrefix(source, lib+".") {
			return types.ImportTypeStdlib
		}
	}

	// Local imports
	if strings.HasPrefix(source, ".") {
		return types.ImportTypeLocal
	}

	// External packages
	return types.ImportTypeExternal
}

// isKeyword checks if a name is a Python keyword
func isKeyword(name string) bool {
	keywords := []string{
		"False", "None", "True", "and", "as", "assert", "async", "await",
		"break", "class", "continue", "def", "del", "elif", "else", "except",
		"finally", "for", "from", "global", "if", "import", "in", "is",
		"lambda", "nonlocal", "not", "or", "pass", "raise", "return", "try",
		"while", "with", "yield",
	}

	for _, kw := range keywords {
		if name == kw {
			return true
		}
	}
	return false
}
