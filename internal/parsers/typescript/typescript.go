package typescript

import (
	"fmt"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
	// sitter "github.com/smacker/go-tree-sitter"
	// "github.com/smacker/go-tree-sitter/typescript/typescript"
)

// TypeScriptParser parses TypeScript/JavaScript using Tree-sitter
type TypeScriptParser struct {
	*parser.BaseParser
	// parser *sitter.Parser
	// In production, this would have actual Tree-sitter parser instance
}

// NewTypeScriptParser creates a new TypeScript parser
func NewTypeScriptParser() *TypeScriptParser {
	return &TypeScriptParser{
		BaseParser: parser.NewBaseParser(
			"typescript",
			[]string{".ts", ".tsx", ".js", ".jsx"},
			100, // High priority for Tree-sitter
		),
	}
}

// Parse parses TypeScript/JavaScript content
func (p *TypeScriptParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
	result := &types.ParseResult{
		Symbols:       make([]*types.Symbol, 0),
		Imports:       make([]*types.Import, 0),
		Relationships: make([]*types.Relationship, 0),
		Metadata:      make(map[string]interface{}),
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
func (p *TypeScriptParser) simpleExtraction(content []byte, result *types.ParseResult, isTS bool) {
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

func (p *TypeScriptParser) extractFunction(line string, lineNum int, result *types.ParseResult) {
	// Extract: function name(params): returnType
	parts := strings.Fields(line)
	for i, part := range parts {
		if part == "function" && i+1 < len(parts) {
			nameWithParams := parts[i+1]
			name := strings.Split(nameWithParams, "(")[0]

			symbol := &types.Symbol{
				Name:       name,
				Type:       types.SymbolTypeFunction,
				StartLine:  lineNum,
				EndLine:    lineNum,
				Visibility: types.VisibilityPublic,
				Signature:  line,
				IsExported: strings.HasPrefix(line, "export"),
			}

			result.Symbols = append(result.Symbols, symbol)
			break
		}
	}
}

func (p *TypeScriptParser) extractArrowFunction(line string, lineNum int, result *types.ParseResult) {
	// Extract: const name = (params) => { }
	if strings.Contains(line, "const ") || strings.Contains(line, "let ") {
		parts := strings.Split(line, "=")
		if len(parts) >= 2 {
			namePart := strings.TrimSpace(parts[0])
			name := strings.Fields(namePart)[len(strings.Fields(namePart))-1]

			symbol := &types.Symbol{
				Name:       name,
				Type:       types.SymbolTypeFunction,
				StartLine:  lineNum,
				EndLine:    lineNum,
				Visibility: types.VisibilityPublic,
				Signature:  line,
				IsExported: strings.HasPrefix(line, "export"),
			}

			result.Symbols = append(result.Symbols, symbol)
		}
	}
}

func (p *TypeScriptParser) extractClass(line string, lineNum int, result *types.ParseResult) {
	parts := strings.Fields(line)
	for i, part := range parts {
		if part == "class" && i+1 < len(parts) {
			name := parts[i+1]
			name = strings.TrimSuffix(name, "{")

			symbol := &types.Symbol{
				Name:       name,
				Type:       types.SymbolTypeClass,
				StartLine:  lineNum,
				EndLine:    lineNum,
				Visibility: types.VisibilityPublic,
				Signature:  line,
				IsExported: strings.HasPrefix(line, "export"),
			}

			result.Symbols = append(result.Symbols, symbol)
			break
		}
	}
}

func (p *TypeScriptParser) extractInterface(line string, lineNum int, result *types.ParseResult) {
	parts := strings.Fields(line)
	for i, part := range parts {
		if part == "interface" && i+1 < len(parts) {
			name := parts[i+1]
			name = strings.TrimSuffix(name, "{")

			symbol := &types.Symbol{
				Name:       name,
				Type:       types.SymbolTypeInterface,
				StartLine:  lineNum,
				EndLine:    lineNum,
				Visibility: types.VisibilityPublic,
				Signature:  line,
				IsExported: true, // Interfaces are always exported in TS
			}

			result.Symbols = append(result.Symbols, symbol)
			break
		}
	}
}

func (p *TypeScriptParser) extractType(line string, lineNum int, result *types.ParseResult) {
	parts := strings.Fields(line)
	for i, part := range parts {
		if part == "type" && i+1 < len(parts) {
			name := parts[i+1]
			name = strings.TrimSuffix(name, "=")

			symbol := &types.Symbol{
				Name:       name,
				Type:       types.SymbolTypeClass, // Use class for type aliases
				StartLine:  lineNum,
				EndLine:    lineNum,
				Visibility: types.VisibilityPublic,
				Signature:  line,
				IsExported: strings.HasPrefix(line, "export"),
			}

			result.Symbols = append(result.Symbols, symbol)
			break
		}
	}
}

func (p *TypeScriptParser) extractImport(line string, lineNum int, result *types.ParseResult) {
	// Extract: import { something } from 'module'
	// or: import something from 'module'
	parts := strings.Split(line, "from")
	if len(parts) >= 2 {
		modulePart := strings.TrimSpace(parts[1])
		modulePath := strings.Trim(modulePart, `'"`)
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

		imp := &types.Import{
			FileID:        0, // Will be set by indexer
			Source:        modulePath,
			ImportedNames: importedNames,
			LineNumber:    lineNum,
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
func (p *TypeScriptParser) extractSymbols(cursor *sitter.TreeCursor, content []byte, result *types.ParseResult) {
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

func (p *TypeScriptParser) handleFunctionDeclaration(node *sitter.Node, content []byte, result *types.ParseResult) {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return
	}

	name := nameNode.Content(content)

	symbol := &types.Symbol{
		Name:       string(name),
		Type:       types.SymbolTypeFunction,
		StartLine:  int(node.StartPoint().Row) + 1,
		EndLine:    int(node.EndPoint().Row) + 1,
		Visibility: types.VisibilityPublic,
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
