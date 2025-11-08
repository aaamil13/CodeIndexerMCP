package swift

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
)

// SwiftParser parses Swift source code
type SwiftParser struct {
}

// NewParser creates a new Swift parser
func NewParser() *SwiftParser {
	return &SwiftParser{}
}

// Language returns the language identifier (e.g., "swift")
func (p *SwiftParser) Language() string {
	return "swift"
}

// Extensions returns file extensions this parser handles (e.g., ".swift")
func (p *SwiftParser) Extensions() []string {
	return []string{".swift"}
}

// Priority returns parser priority (higher = preferred when multiple parsers match)
func (p *SwiftParser) Priority() int {
	return 100
}

// SupportsFramework checks if parser supports specific framework analysis
func (p *SwiftParser) SupportsFramework(framework string) bool {
	return false
}

// Parse parses Swift source code
func (p *SwiftParser) Parse(content []byte, filePath string) (*parsing.ParseResult, error) {
	result := &parsing.ParseResult{
		Symbols:       make([]*model.Symbol, 0),
		Imports:       make([]*model.Import, 0),
		Relationships: make([]*model.Relationship, 0),
		Metadata:      make(map[string]interface{}),
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Extract imports
	p.extractImports(lines, result)

	// Extract types (class, struct, protocol, enum, actor)
	p.extractTypes(contentStr, result)

	// Extract extensions
	p.extractExtensions(contentStr, result)

	// Extract functions
	p.extractFunctions(contentStr, result)

	// Extract properties
	p.extractProperties(contentStr, result)

	result.Metadata["language"] = "swift"

	return result, nil
}

func (p *SwiftParser) extractImports(lines []string, result *parsing.ParseResult) {
	importRe := regexp.MustCompile(`^\s*import\s+([\w.]+)`)

	for i, line := range lines {
		if matches := importRe.FindStringSubmatch(line); matches != nil {
			imp := &model.Import{
				Path: matches[1],
				Range: model.Range{
					Start: model.Position{Line: i + 1},
					End:   model.Position{Line: i + 1},
				},
			}
			result.Imports = append(result.Imports, imp)
		}
	}
}

func (p *SwiftParser) extractTypes(content string, result *parsing.ParseResult) {
	// Class, struct, protocol, enum, actor
	typeRe := regexp.MustCompile(`(?m)^\s*(?:@[\w.()]+)?\s*(?:final|static)?\s*(class|struct|protocol|enum|actor)\s+(\w+)(?:\s*:\s*([\w,\s&]+))?`)

	matches := typeRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		visibility := ""
		if match[2] != -1 {
			visibility = content[match[2]:match[3]]
		}

		modifier := ""
		if match[4] != -1 {
			modifier = content[match[4]:match[5]]
		}

		typeKind := content[match[6]:match[7]]
		name := content[match[8]:match[9]]

		inheritance := ""
		if match[10] != -1 {
			inheritance = content[match[10]:match[11]]
		}

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		// Determine symbol type
		var symbolKind model.SymbolKind
		switch typeKind {
		case "class", "actor":
			symbolKind = model.SymbolKindClass
		case "struct":
			symbolKind = model.SymbolKindStruct
		case "protocol":
			symbolKind = model.SymbolKindInterface
		case "enum":
			symbolKind = model.SymbolKindEnum
		default:
			symbolKind = model.SymbolKindClass
		}

		sig := typeKind + " " + name
		if inheritance != "" {
			sig += ": " + inheritance
		}

		symbol := &model.Symbol{
			Name:       name,
			Kind:       symbolKind,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: p.parseVisibility(visibility),
			Signature:  sig,
		}

		if modifier != "" {
			symbol.Metadata = map[string]string{
				"modifier": modifier,
			}
		}

		result.Symbols = append(result.Symbols, symbol)

		// Add relationships
		if inheritance != "" {
			parts := strings.Split(inheritance, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				// Remove generic constraints
				part = strings.Split(part, "&")[0]
				part = strings.TrimSpace(part)

				if part != "" {
					result.Relationships = append(result.Relationships, &model.Relationship{
						Type:       model.RelationshipKindExtends,
						SourceSymbol: name,
						TargetSymbol: part,
					})
				}
			}
		}
	}
}

func (p *SwiftParser) extractExtensions(content string, result *parsing.ParseResult) {
	extRe := regexp.MustCompile(`(?m)^\s*extension\s+(\w+)(?:\s*:\s*([\w,\s&]+))?`)

	matches := extRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]

		protocols := ""
		if match[4] != -1 {
			protocols = content[match[4]:match[5]]
		}

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		sig := "extension " + name
		if protocols != "" {
			sig += ": " + protocols
		}

		symbol := &model.Symbol{
			Name:       name + " (extension)",
			Kind:       model.SymbolKindClass, // Extensions are applied to classes/structs/enums
			File:       "",                    // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  sig,
			Metadata: map[string]string{
				"extension": "true",
			},
		}

		result.Symbols = append(result.Symbols, symbol)

		// Add conformances
		if protocols != "" {
			parts := strings.Split(protocols, ",")
			for _, protocol := range parts {
				protocol = strings.TrimSpace(protocol)
				if protocol != "" {
					result.Relationships = append(result.Relationships, &model.Relationship{
						Type:       model.RelationshipKindImplements,
						SourceSymbol: name,
						TargetSymbol: protocol,
					})
				}
			}
		}
	}
}

func (p *SwiftParser) extractFunctions(content string, result *parsing.ParseResult) {
	// Function/method declaration
	funcRe := regexp.MustCompile(`(?m)^\s*(?:@[\w.()]+)?\s*(?:static|class|mutating|override|final)?\s*func\s+(\w+)(?:<[^>]+>)?\s*\(([^)]*)\)(?:\s*(?:async|throws|rethrows))?\s*(?:->\s*([\w<>?]+))?`)

	matches := funcRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		visibility := ""
		if match[2] != -1 {
			visibility = content[match[2]:match[3]]
		}

		modifier := ""
		if match[4] != -1 {
			modifier = content[match[4]:match[5]]
		}

		name := content[match[6]:match[7]]
		params := ""
		if match[8] != -1 {
			params = content[match[8]:match[9]]
		}

		returnType := ""
		if match[10] != -1 {
			returnType = content[match[10]:match[11]]
		}

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		sig := "func " + name + "(" + params + ")"
		if returnType != "" {
			sig += " -> " + returnType
		}

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindFunction,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: p.parseVisibility(visibility),
			Signature:  sig,
		}

		if modifier != "" {
			symbol.Metadata = map[string]string{
				"modifier": modifier,
			}
		}

		result.Symbols = append(result.Symbols, symbol)
	}

	// Init (constructor)
	initRe := regexp.MustCompile(`(?m)^\s*(public|private|fileprivate|internal|open)?\s*(required|convenience)?\s*init(?:\?|!)?\s*\(([^)]*)\)`)

	initMatches := initRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range initMatches {
		visibility := ""
		if match[2] != -1 {
			visibility = content[match[2]:match[3]]
		}

		modifier := ""
		if match[4] != -1 {
			modifier = content[match[4]:match[5]]
		}

		params := ""
		if match[6] != -1 {
			params = content[match[6]:match[7]]
		}

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		sig := "init(" + params + ")"

		symbol := &model.Symbol{
			Name:       "init",
			Kind:       model.SymbolKindConstructor,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: p.parseVisibility(visibility),
			Signature:  sig,
		}

		if modifier != "" {
			symbol.Metadata = map[string]string{
				"modifier": modifier,
			}
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *SwiftParser) extractProperties(content string, result *parsing.ParseResult) {
	// Property declaration: var/let
	propRe := regexp.MustCompile(`(?m)^\s*(public|private|fileprivate|internal|open)?\s*(static|class|lazy)?\s*(var|let)\s+(\w+)\s*:\s*([\w<>?[\]]+)`)

	matches := propRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		visibility := ""
		if match[2] != -1 {
			visibility = content[match[2]:match[3]]
		}

		modifier := ""
		if match[4] != -1 {
			modifier = content[match[4]:match[5]]
		}

		varOrLet := content[match[6]:match[7]]
		name := content[match[8]:match[9]]
		propType := content[match[10]:match[11]]

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		sig := varOrLet + " " + name + ": " + propType

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindProperty,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: p.parseVisibility(visibility),
			Signature:  sig,
		}

		if modifier != "" {
			symbol.Metadata = map[string]string{
				"modifier": modifier,
			}
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *SwiftParser) parseVisibility(vis string) model.Visibility {
	switch strings.ToLower(vis) {
	case "public", "open":
		return model.VisibilityPublic
	case "private":
		return model.VisibilityPrivate
	case "fileprivate":
		return model.VisibilityPrivate // File-scoped
	case "internal":
		return model.VisibilityInternal
	default:
		return model.VisibilityInternal // Swift default is internal
	}
}
