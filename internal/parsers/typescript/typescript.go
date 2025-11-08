package typescript

import (
	"fmt"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
	// sitter "github.com/smacker/go-tree-sitter"
	// "github.com/smacker/go-tree-sitter/typescript/typescript"
)

// TypeScriptParser parses TypeScript/JavaScript using Tree-sitter
type TypeScriptParser struct {
	// parser *sitter.Parser
	// In production, this would have actual Tree-sitter parser instance
}

// NewTypeScriptParser creates a new TypeScript parser
func NewTypeScriptParser() *TypeScriptParser {
	return &TypeScriptParser{}
}

// Language returns the language identifier (e.g., "typescript")
func (p *TypeScriptParser) Language() string {
	return "typescript"
}

// Extensions returns file extensions this parser handles (e.g., [".ts", ".tsx"])
func (p *TypeScriptParser) Extensions() []string {
	return []string{".ts", ".tsx", ".js", ".jsx"}
}

// Priority returns parser priority (higher = preferred when multiple parsers match)
func (p *TypeScriptParser) Priority() int {
	return 100
}

// Parse parses TypeScript/JavaScript content
func (p *TypeScriptParser) Parse(content []byte, filePath string) (*parsing.ParseResult, error) {
	result := &parsing.ParseResult{
		Symbols:  make([]*model.Symbol, 0),
		Imports:  make([]*model.Import, 0),
		Metadata: make(map[string]interface{}),
	}

	// Determine if this is TypeScript or JavaScript
	isTypeScript := strings.HasSuffix(filePath, ".ts") || strings.HasSuffix(filePath, ".tsx")
	isJSX := strings.HasSuffix(filePath, ".tsx") || strings.HasSuffix(filePath, ".jsx")

	result.Metadata["is_typescript"] = isTypeScript
	result.Metadata["is_jsx"] = isJSX

	// In production, this would use actual Tree-sitter:
	/*
		parser := sitter.NewParser()
		if isTypeScript {
			parser.SetLanguage(typescript.GetLanguage())
		} else {
			parser.SetLanguage(javascript.GetLanguage())
		}

		tree, err := parser.ParseCtx(context.Background(), nil, content)
		if err != nil {
			return nil, err
		}
		defer tree.Close()

		// Walk the syntax tree
		cursor := sitter.NewTreeCursor(tree.RootNode())
		defer cursor.Close()

		p.extractSymbols(cursor, content, result)
	*/

	// Fallback: Simple regex-based extraction for demonstration
	p.simpleExtraction(content, result, isTypeScript)

	return result, nil
}

// simpleExtraction provides basic extraction without Tree-sitter (temporary)
func (p *TypeScriptParser) simpleExtraction(content []byte, result *parsing.ParseResult, isTS bool) {
	lines := strings.Split(string(content), "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Function declarations
		if strings.Contains(line, "function ") {
			p.extractFunction(line, i+1, result)
		}

		// Arrow functions
		if strings.Contains(line, "=>") && (strings.Contains(line, "const ") || strings.Contains(line, "let ")) {
			p.extractArrowFunction(line, i+1, result)
		}

		// Class declarations
		if strings.HasPrefix(line, "class ") || strings.Contains(line, " class ") {
			p.extractClass(line, i+1, result)
		}

		// Interface declarations (TypeScript)
		if isTS && (strings.HasPrefix(line, "interface ") || strings.Contains(line, " interface ")) {
			p.extractInterface(line, i+1, result)
		}

		// Type alias (TypeScript)
		if isTS && (strings.HasPrefix(line, "type ") || strings.Contains(line, " type ")) {
			p.extractType(line, i+1, result)
		}

		// Import statements
		if strings.HasPrefix(line, "import ") {
			p.extractImport(line, i+1, result)
		}

		// Export statements
		if strings.HasPrefix(line, "export ") {
			// Handle export statements
		}
	}
}

func (p *TypeScriptParser) extractFunction(line string, lineNum int, result *parsing.ParseResult) {
	// Extract: function name(params): returnType
	parts := strings.Fields(line)
	for i, part := range parts {
		if part == "function" && i+1 < len(parts) {
			nameWithParams := parts[i+1]
			name := strings.Split(nameWithParams, "(")[0]

			symbol := &model.Symbol{
				Name:       name,
				Kind:       model.SymbolKindFunction,
				File:       "", // File path will be set by the caller
				Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
				Visibility: model.VisibilityPublic,
				Signature:  line,
				Metadata: map[string]string{
					"is_exported": fmt.Sprintf("%t", strings.HasPrefix(line, "export")),
				},
			}

			result.Symbols = append(result.Symbols, symbol)
			break
		}
	}
}

func (p *TypeScriptParser) extractArrowFunction(line string, lineNum int, result *parsing.ParseResult) {
	// Extract: const name = (params) => { }
	if strings.Contains(line, "const ") || strings.Contains(line, "let ") {
		parts := strings.Split(line, "=")
		if len(parts) >= 2 {
			namePart := strings.TrimSpace(parts[0])
			name := strings.Fields(namePart)[len(strings.Fields(namePart))-1]

			symbol := &model.Symbol{
				Name:       name,
				Kind:       model.SymbolKindFunction,
				File:       "", // File path will be set by the caller
				Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
				Visibility: model.VisibilityPublic,
				Signature:  line,
				Metadata: map[string]string{
					"is_exported": fmt.Sprintf("%t", strings.HasPrefix(line, "export")),
				},
			}

			result.Symbols = append(result.Symbols, symbol)
		}
	}
}

func (p *TypeScriptParser) extractClass(line string, lineNum int, result *parsing.ParseResult) {
	parts := strings.Fields(line)
	for i, part := range parts {
		if part == "class" && i+1 < len(parts) {
			name := parts[i+1]
			name = strings.TrimSuffix(name, "{")

			symbol := &model.Symbol{
				Name:       name,
				Kind:       model.SymbolKindClass,
				File:       "", // File path will be set by the caller
				Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
				Visibility: model.VisibilityPublic,
				Signature:  line,
				Metadata: map[string]string{
					"is_exported": fmt.Sprintf("%t", strings.HasPrefix(line, "export")),
				},
			}

			result.Symbols = append(result.Symbols, symbol)
			break
		}
	}
}

func (p *TypeScriptParser) extractInterface(line string, lineNum int, result *parsing.ParseResult) {
	parts := strings.Fields(line)
	for i, part := range parts {
		if part == "interface" && i+1 < len(parts) {
			name := parts[i+1]
			name = strings.TrimSuffix(name, "{")

			symbol := &model.Symbol{
				Name:       name,
				Kind:       model.SymbolKindInterface,
				File:       "", // File path will be set by the caller
				Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
				Visibility: model.VisibilityPublic,
				Signature:  line,
				Metadata: map[string]string{
					"is_exported": "true", // Interfaces are always exported in TS
				},
			}

			result.Symbols = append(result.Symbols, symbol)
			break
		}
	}
}

func (p *TypeScriptParser) extractType(line string, lineNum int, result *parsing.ParseResult) {
	parts := strings.Fields(line)
	for i, part := range parts {
		if part == "type" && i+1 < len(parts) {
			name := parts[i+1]
			name = strings.TrimSuffix(name, "=")

			symbol := &model.Symbol{
				Name:       name,
				Kind:       model.SymbolKindClass, // Use class for type aliases
				File:       "",                    // File path will be set by the caller
				Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
				Visibility: model.VisibilityPublic,
				Signature:  line,
				Metadata: map[string]string{
					"is_exported": fmt.Sprintf("%t", strings.HasPrefix(line, "export")),
				},
			}

			result.Symbols = append(result.Symbols, symbol)
			break
		}
	}
}

func (p *TypeScriptParser) extractImport(line string, lineNum int, result *parsing.ParseResult) {
	// Extract: import { something } from 'module'
	// or: import something from 'module'
	parts := strings.Split(line, "from")
	if len(parts) >= 2 {
		modulePart := strings.TrimSpace(parts[1])
		modulePath := strings.Trim(modulePart, `"'`)
		modulePath = strings.TrimSuffix(modulePath, ";")

		// Extract imported names
		importedNames := make([]string, 0)
		importPart := strings.TrimSpace(parts[0])
		importPart = strings.TrimPrefix(importPart, "import")
		importPart = strings.TrimSpace(importPart)

		if strings.Contains(importPart, "{") {
			// Named imports
			namesStr := strings.Trim(importPart, "{}")
			names := strings.Split(namesStr, ",")
			for _, name := range names {
				name = strings.TrimSpace(name)
				if name != "" {
					// Handle "as" aliases
					if strings.Contains(name, " as ") {
						parts := strings.Split(name, " as ")
						importedNames = append(importedNames, strings.TrimSpace(parts[0]))
					} else {
						importedNames = append(importedNames, name)
					}
				}
			}
		} else {
			// Default import
			name := strings.TrimSpace(importPart)
			if name != "" && name != "*" {
				importedNames = append(importedNames, name)
			}
		}

		imp := &model.Import{
			Path:    modulePath,
			Members: importedNames,
			Range: model.Range{
				Start: model.Position{Line: lineNum},
				End:   model.Position{Line: lineNum},
			},
		}

		result.Imports = append(result.Imports, imp)
	}
}

// SupportsFramework checks if this parser supports framework analysis
func (p *TypeScriptParser) SupportsFramework(framework string) bool {
	supported := map[string]bool{
		"react":   true,
		"vue":     true,
		"angular": true,
		"express": true,
		"nest":    true,
	}
	return supported[strings.ToLower(framework)]
}

// TreeSitterExtractor would extract symbols using actual Tree-sitter
// This is the production implementation template:
/*
func (p *TypeScriptParser) extractSymbols(cursor *sitter.TreeCursor, content []byte, result *parsing.ParseResult) {
	node := cursor.CurrentNode()

	switch node.Type() {
	case "function_declaration":
		p.handleFunctionDeclaration(node, content, result)
	case "class_declaration":
		p.handleClassDeclaration(node, content, result)
	case "interface_declaration":
		p.handleInterfaceDeclaration(node, content, result)
	case "type_alias_declaration":
		p.handleTypeAliasDeclaration(node, content, result)
	case "method_definition":
		p.handleMethodDefinition(node, content, result)
	case "import_statement":
		p.handleImportStatement(node, content, result)
	}

	// Recursively process children
	if cursor.GoToFirstChild() {
		p.extractSymbols(cursor, content, result)
		cursor.GoToParent()
	}

	// Process siblings
	for cursor.GoToNextSibling() {
		p.extractSymbols(cursor, content, result)
	}
}

func (p *TypeScriptParser) handleFunctionDeclaration(node *sitter.Node, content []byte, result *parsing.ParseResult) {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return
	}

	name := nameNode.Content(content)

	symbol := &model.Symbol{
		Name:       string(name),
		Kind:       model.SymbolKindFunction,
		Range:      model.Range{Start: model.Position{Line: int(node.StartPoint().Row) + 1}, End: model.Position{Line: int(node.EndPoint().Row) + 1}},
		Visibility: model.VisibilityPublic,
		Signature:  string(node.Content(content)),
	}

	result.Symbols = append(result.Symbols, symbol)
}
*/

// Production Tree-sitter query examples:
const tsQueries = `
; Functions
(function_declaration
  name: (identifier) @function.name) @function

; Arrow functions
(lexical_declaration
  (variable_declarator
    name: (identifier) @function.name
    value: (arrow_function))) @function

; Classes
(class_declaration
  name: (type_identifier) @class.name) @class

; Methods
(method_definition
  name: (property_identifier) @method.name) @method

; Interfaces (TypeScript)
(interface_declaration
  name: (type_identifier) @interface.name) @interface

; Type aliases (TypeScript)
(type_alias_declaration
  name: (type_identifier) @type.name) @type

; Imports
(import_statement
  source: (string) @import.source) @import
`

func init() {
	// In production: Load and compile Tree-sitter queries
	fmt.Println("TypeScript parser initialized (using fallback mode until Tree-sitter is fully integrated)")
}
