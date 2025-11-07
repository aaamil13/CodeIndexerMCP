package python

import (
	"bufio"
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
	currentClass := ""
	var currentClassSymbol *types.Symbol
	var docstringLines []string
	inDocstring := false
	docstringMarker := ""

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

		// Handle docstrings
		if strings.HasPrefix(trimmed, `"""`) || strings.HasPrefix(trimmed, "'''") {
			if !inDocstring {
				inDocstring = true
				docstringMarker = trimmed[:3]
				docstringLines = []string{strings.TrimPrefix(trimmed, docstringMarker)}
				if strings.HasSuffix(trimmed, docstringMarker) && len(trimmed) > 6 {
					// Single-line docstring
					inDocstring = false
					docstringLines = []string{strings.Trim(trimmed, docstringMarker)}
				}
			} else if strings.Contains(trimmed, docstringMarker) {
				inDocstring = false
				docstringLines = append(docstringLines, strings.TrimSuffix(trimmed, docstringMarker))
			}
			continue
		}

		if inDocstring {
			docstringLines = append(docstringLines, trimmed)
			continue
		}

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Get indentation level
		indent := len(line) - len(trimmed)

		// Reset class context if we're back at top level
		if indent == 0 && currentClass != "" {
			currentClass = ""
			currentClassSymbol = nil
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
				Documentation: strings.TrimSpace(strings.Join(docstringLines, "\n")), // Trim documentation
			}

			if parentClasses != "" {
				symbol.Metadata = map[string]interface{}{
					"parent_classes": parentClasses,
				}
			}

			result.Symbols = append(result.Symbols, symbol)
			currentClass = className
			currentClassSymbol = symbol
			docstringLines = []string{}
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
			if currentClass != "" {
				symbolType = types.SymbolTypeMethod
				if currentClassSymbol != nil {
					parentID = &currentClassSymbol.ID
				}
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
				Documentation: strings.TrimSpace(strings.Join(docstringLines, "\n")), // Trim documentation
			}

			// Check for decorators
			if len(result.Symbols) > 0 {
				lastSymbol := result.Symbols[len(result.Symbols)-1]
				if lastSymbol.Type == types.SymbolTypeFunction && strings.HasPrefix(lastSymbol.Name, "@") {
					symbol.Metadata = map[string]interface{}{
						"decorators": []string{lastSymbol.Name},
					}
				}
			}

			result.Symbols = append(result.Symbols, symbol)
			docstringLines = []string{}
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
			continue
		}

		// Check for decorators
		if match := decoratorRegex.FindStringSubmatch(trimmed); match != nil {
			// Store decorator info for next function
			decoratorName := "@" + match[1]
			result.Symbols = append(result.Symbols, &types.Symbol{
				Name:       decoratorName,
				Type:       types.SymbolTypeFunction,
				StartLine:  lineNumber,
				Visibility: types.VisibilityPublic,
			})
			continue
		}

		// Check for variables (simple detection)
		if indent == 0 && !strings.HasPrefix(trimmed, "def") && !strings.HasPrefix(trimmed, "class") {
			if match := varRegex.FindStringSubmatch(trimmed); match != nil {
				varName := match[1]
				if !isKeyword(varName) {
					symbolType := types.SymbolTypeVariable
					// Python constants are typically all caps
					if varName == strings.ToUpper(varName) && strings.ContainsAny(varName, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
						symbolType = types.SymbolTypeConstant
					}

					result.Symbols = append(result.Symbols, &types.Symbol{
						Name:       varName,
						Type:       symbolType,
						StartLine:  lineNumber,
						EndLine:    lineNumber,
						Visibility: p.getVisibility(varName),
						IsExported: p.isExported(varName),
					})
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
