package python

import (
	"bufio"
	"log"
	"regexp"
	"strings"
	"strconv"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
)

// Parser is the Python language parser
type Parser struct {
}

// NewParser creates a new Python parser
func NewParser() *Parser {
	return &Parser{}
}

// Language returns the language identifier (e.g., "python")
func (p *Parser) Language() string {
	return "python"
}

// Extensions returns file extensions this parser handles (e.g., [".py"])
func (p *Parser) Extensions() []string {
	return []string{".py"}
}

// Priority returns parser priority (higher = preferred when multiple parsers match)
func (p *Parser) Priority() int {
	return 100
}

// SupportsFramework checks if parser supports specific framework analysis
func (p *Parser) SupportsFramework(framework string) bool {
	return false
}

// Parse parses Python source code
// This is a basic regex-based parser. For production use, consider using tree-sitter-python
func (p *Parser) Parse(content []byte, filePath string) (*parsing.ParseResult, error) {
	result := &parsing.ParseResult{
		Symbols:       []*model.Symbol{},
		Imports:       []*model.Import{},
		Relationships: []*model.Relationship{},
		Metadata:      make(map[string]interface{}),
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	lineNumber := 0
	var docstringLines []string
	inDocstring := false
	docstringMarker := ""
	var pendingDocstring string
	parentStack := []*model.Symbol{}

	classRegex := regexp.MustCompile(`^class\s+(\w+)(\(.*\))?`)
	functionRegex := regexp.MustCompile(`^(async\s+)?def\s+(\w+)\s*\((.*?)\)(\s*->\s*.+)?`)
	importRegex := regexp.MustCompile(`^import\s+(.+)`)
	fromImportRegex := regexp.MustCompile(`^from\s+(.+?)\s+import\s+(.+)`)
	decoratorRegex := regexp.MustCompile(`^@(\w+)`)
	varRegex := regexp.MustCompile(`^(\w+)\s*[:=]`)

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		log.Printf("DEBUG: Line %d: %s (trimmed: %s)", lineNumber, line, trimmed)

		if (strings.HasPrefix(trimmed, `"""`) || strings.HasPrefix(trimmed, "'''")) {
			if !inDocstring {
				inDocstring = true
				docstringMarker = trimmed[:3]
				docstringLines = []string{strings.TrimPrefix(trimmed, docstringMarker)}
				if strings.HasSuffix(trimmed, docstringMarker) && len(trimmed) > 6 {
					pendingDocstring = strings.TrimSpace(strings.Trim(trimmed, docstringMarker))
					log.Printf("DEBUG: Docstring: Single line ended at line %d, content: '%s'", lineNumber, pendingDocstring)
					inDocstring = false
					docstringLines = []string{}
				                } else {
				                    log.Printf("DEBUG: Docstring: Started at line %d, marker: %s, content: %v", lineNumber, docstringMarker, docstringLines)
				                }
				            } else if strings.Contains(trimmed, docstringMarker) {
				                docstringLines = append(docstringLines, strings.TrimSuffix(trimmed, docstringMarker))
				                pendingDocstring = strings.TrimSpace(strings.Join(docstringLines, "\n")) // Finalize pendingDocstring
				                log.Printf("DEBUG: Docstring: Multi-line ended at line %d, content: '%s'", lineNumber, pendingDocstring)
				                inDocstring = false
				                docstringLines = []string{}
				            }
				            // After a docstring ends, ensure pendingDocstring is set
				            if !inDocstring && len(docstringLines) > 0 { // Only finalize if there were lines
				                pendingDocstring = strings.TrimSpace(strings.Join(docstringLines, "\n"))
				                docstringLines = []string{} // Clear docstringLines after finalizing pendingDocstring
				            }
				            continue
				        }
		if inDocstring {
			docstringLines = append(docstringLines, trimmed)
			log.Printf("DEBUG: Docstring: Appending line %d: %s", lineNumber, trimmed)
			continue
		}

		// If we are not in a docstring and there's a pending docstring, and the current line is not a definition, discard it.
		// This handles cases where a docstring is followed by blank lines or comments before a definition.
		if pendingDocstring != "" && trimmed != "" && !strings.HasPrefix(trimmed, "#") &&
			!classRegex.MatchString(trimmed) && !functionRegex.MatchString(trimmed) { // Only check for class/function definitions
			log.Printf("DEBUG: Discarding pending docstring at line %d as no definition followed: '%s'", lineNumber, pendingDocstring)
			pendingDocstring = ""
			docstringLines = []string{}
		}

		if inDocstring {
			docstringLines = append(docstringLines, trimmed)
			log.Printf("DEBUG: Docstring: Appending line %d: %s", lineNumber, trimmed)
			continue
		}



		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := len(line) - len(strings.TrimLeft(line, " "))
		log.Printf("DEBUG: Line %d: Indent %d, parentStack size: %d", lineNumber, indent, len(parentStack))

		for len(parentStack) > 0 {
			indentStr, ok := parentStack[len(parentStack)-1].Metadata["indent"]
			if !ok {
				// Handle error or default value
				break
			}
			parentIndent, err := strconv.Atoi(indentStr)
			if err != nil {
				// Handle error
				break
			}
			if indent <= parentIndent {
				log.Printf("DEBUG: Popping from parentStack. Current indent %d <= parent indent %d (Symbol: %s)", indent, parentIndent, parentStack[len(parentStack)-1].Name)
				parentStack = parentStack[:len(parentStack)-1]
			} else {
				break
			}
		}

		if indent == 0 && len(parentStack) > 0 {
			log.Printf("DEBUG: Clearing parentStack at line %d due to top-level indent.", lineNumber)
			parentStack = []*model.Symbol{}
		}

		if match := classRegex.FindStringSubmatch(trimmed); match != nil {
			className := match[1]
			var parentClasses string
			if len(match) > 2 && match[2] != "" {
				parentClasses = strings.Trim(match[2], "()")
			}

			symbol := &model.Symbol{
				Name:          className,
				Kind:          model.SymbolKindClass,
				File:          "", // File path will be set by the caller
				Range:         model.Range{Start: model.Position{Line: lineNumber, Column: 1, Byte: 1}, End: model.Position{Line: lineNumber, Column: 1, Byte: 1}},
				Visibility:    p.getVisibility(className),
				Documentation: pendingDocstring, // Assign pending docstring
				Metadata:      make(map[string]string),
			}

			if parentClasses != "" {
				symbol.Metadata["parent_classes"] = parentClasses
			}

			result.Symbols = append(result.Symbols, symbol)
			symbol.Metadata["indent"] = strconv.Itoa(indent)
			parentStack = append(parentStack, symbol)
			log.Printf("DEBUG: Class Symbol: %s, Doc: '%s', parentStack size: %d", className, pendingDocstring, len(parentStack))
			continue
		}

		if match := functionRegex.FindStringSubmatch(trimmed); match != nil {
			isAsync := match[1] != ""
			funcName := match[2]
			params := match[3]
			returnType := ""
			if len(match) > 4 {
				returnType = strings.TrimSpace(strings.TrimPrefix(match[4], "->"))
			}

			symbolKind := model.SymbolKindFunction
			var parentID string

			if len(parentStack) > 0 {
				parentSymbol := parentStack[len(parentStack)-1]
				if parentSymbol.Kind == model.SymbolKindClass {
					symbolKind = model.SymbolKindMethod
				} else {
					symbolKind = model.SymbolKindFunction
				}
				parentID = parentSymbol.ID
				log.Printf("DEBUG: Assigning ParentID %s (from stack) to %s %s at line %d", parentID, symbolKind, funcName, lineNumber)
			}

			signature := p.buildSignature(funcName, params, returnType, isAsync)
			symbol := &model.Symbol{
				Name:          funcName,
				Kind:          symbolKind,
				Signature:     signature,
				File:          "", // File path will be set by the caller
				Range:         model.Range{Start: model.Position{Line: lineNumber, Column: 1, Byte: 1}, End: model.Position{Line: lineNumber, Column: 1, Byte: 1}},
				Visibility:    p.getVisibility(funcName),
				Documentation: pendingDocstring, // Assign pending docstring
				Metadata:      make(map[string]string),
			}
			if isAsync {
				symbol.Metadata["is_async"] = "true"
			}

			var decorators []string
			for i := len(result.Symbols) - 1; i >= 0; i-- {
				lastSymbol := result.Symbols[i]
				if lastSymbol.Kind == model.SymbolKindDecorator {
					decorators = append([]string{lastSymbol.Name}, decorators...)
					result.Symbols = result.Symbols[:i]
				} else {
					break
				}
			}

			if len(decorators) > 0 {
				symbol.Metadata["decorators"] = strings.Join(decorators, ",")
				log.Printf("DEBUG: Associated decorators %v with function '%s' at line %d", decorators, funcName, lineNumber)
			}

			symbol.Metadata["indent"] = strconv.Itoa(indent)
			result.Symbols = append(result.Symbols, symbol)
			parentStack = append(parentStack, symbol)
			log.Printf("DEBUG: Function/Method Symbol: %s, Kind: %s, Doc: '%s', parentStack size: %d", funcName, symbolKind, pendingDocstring, len(parentStack))
			continue
		}

		if match := importRegex.FindStringSubmatch(trimmed); match != nil {
			imports := strings.Split(match[1], ",")
			for _, imp := range imports {
				imp = strings.TrimSpace(imp)
				result.Imports = append(result.Imports, &model.Import{
					Path: imp,
					Range: model.Range{
						Start: model.Position{Line: lineNumber, Column: 1, Byte: 1},
						End:   model.Position{Line: lineNumber, Column: 1, Byte: 1},
					},
				})
			}
			log.Printf("DEBUG: Found import: %v at line %d", imports, lineNumber)
			continue
		}

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

			result.Imports = append(result.Imports, &model.Import{
				Path:    source,
				Members: importedNames,
				Range: model.Range{
					Start: model.Position{Line: lineNumber, Column: 1, Byte: 1},
					End:   model.Position{Line: lineNumber, Column: 1, Byte: 1},
				},
			})
			log.Printf("DEBUG: Found from-import: from %s import %v at line %d", source, importedNames, lineNumber)
			continue
		}

		if match := decoratorRegex.FindStringSubmatch(trimmed); match != nil {
			decoratorName := "@" + match[1]
			result.Symbols = append(result.Symbols, &model.Symbol{
				Name:       decoratorName,
				Kind:       model.SymbolKindDecorator,
				File:       "", // File path will be set by the caller
				Range:      model.Range{Start: model.Position{Line: lineNumber, Column: 1, Byte: 1}, End: model.Position{Line: lineNumber, Column: 1, Byte: 1}},
				Visibility: model.VisibilityPublic,
			})
			log.Printf("DEBUG: Found decorator: %s at line %d", decoratorName, lineNumber)
			continue
		}

		if indent == 0 && !strings.HasPrefix(trimmed, "def") && !strings.HasPrefix(trimmed, "class") {
			if match := varRegex.FindStringSubmatch(trimmed); match != nil {
				varName := match[1]
				if !isKeyword(varName) {
					symbolKind := model.SymbolKindVariable

					result.Symbols = append(result.Symbols, &model.Symbol{
						Name:       varName,
						Kind:       symbolKind,
						File:       "", // File path will be set by the caller
						Range:      model.Range{Start: model.Position{Line: lineNumber, Column: 1, Byte: 1}, End: model.Position{Line: lineNumber, Column: 1, Byte: 1}},
						Visibility: p.getVisibility(varName),
					})
					log.Printf("DEBUG: Found variable/constant: %s, Kind: %s at line %d", varName, symbolKind, lineNumber)
				}
			}
		}
	}

	return result, scanner.Err()
}

func (p *Parser) buildSignature(name, params, returnType string, isAsync bool) string {
	sig := ""
	if isAsync {
		sig = "async "
	}
	sig += "def " + name + "(" + params + ")"
	if returnType != "" {
		sig += " -> " + returnType
	}
	sig += ":"
	return sig
}

func (p *Parser) getVisibility(name string) model.Visibility {
	if strings.HasPrefix(name, "__") && !strings.HasSuffix(name, "__") {
		return model.VisibilityInternal
	}
	if strings.HasPrefix(name, "_") {
		return model.VisibilityPrivate
	}
	return model.VisibilityPublic
}

func (p *Parser) getImportType(source string) model.ImportKind {
	stdLibs := []string{
		"os", "sys", "re", "json", "datetime", "collections", "itertools",
		"functools", "pathlib", "typing", "abc", "math", "random", "time",
		"io", "subprocess", "threading", "multiprocessing", "asyncio",
	}

	for _, lib := range stdLibs {
		if source == lib || strings.HasPrefix(source, lib+".") {
			return model.ImportKindStdlib
		}
	}

	if strings.HasPrefix(source, ".") {
		return model.ImportKindLocal
	}

	return model.ImportKindExternal
}

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
