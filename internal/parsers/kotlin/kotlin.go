package kotlin

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
)

// KotlinParser parses Kotlin source code
type KotlinParser struct {
}

// NewParser creates a new Kotlin parser
func NewParser() *KotlinParser {
	return &KotlinParser{}
}

// Language returns the language identifier (e.g., "kotlin")
func (p *KotlinParser) Language() string {
	return "kotlin"
}

// Extensions returns file extensions this parser handles (e.g., [".kt", ".kts"])
func (p *KotlinParser) Extensions() []string {
	return []string{".kt", ".kts"}
}

// Priority returns parser priority (higher = preferred when multiple parsers match)
func (p *KotlinParser) Priority() int {
	return 100
}

// SupportsFramework checks if parser supports specific framework analysis
func (p *KotlinParser) SupportsFramework(framework string) bool {
	return false
}

// Parse parses Kotlin source code
func (p *KotlinParser) Parse(content []byte, filePath string) (*parsing.ParseResult, error) {
	result := &parsing.ParseResult{
		Symbols:       make([]*model.Symbol, 0),
		Imports:       make([]*model.Import, 0),
		Relationships: make([]*model.Relationship, 0),
		Metadata:      make(map[string]interface{}),
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Extract package
	p.extractPackage(lines, result)

	// Extract imports
	p.extractImports(lines, result)

	// Extract classes, interfaces, objects
	p.extractTypes(contentStr, result)

	// Extract functions
	p.extractFunctions(contentStr, result)

	// Extract properties
	p.extractProperties(contentStr, result)

	result.Metadata["language"] = "kotlin"

	return result, nil
}

func (p *KotlinParser) extractPackage(lines []string, result *parsing.ParseResult) {
	packageRe := regexp.MustCompile(`^\s*package\s+([\w.]+)`)

	for i, line := range lines {
		if matches := packageRe.FindStringSubmatch(line); matches != nil {
			result.Metadata["package"] = matches[1]

			symbol := &model.Symbol{
				Name:       matches[1],
				Kind:       model.SymbolKindPackage,
				File:       "", // File path will be set by the caller
				Range:      model.Range{Start: model.Position{Line: i + 1}, End: model.Position{Line: i + 1}},
				Visibility: model.VisibilityPublic,
				Signature:  "package " + matches[1],
			}
			result.Symbols = append(result.Symbols, symbol)
			break
		}
	}
}

func (p *KotlinParser) extractImports(lines []string, result *parsing.ParseResult) {
	importRe := regexp.MustCompile(`^\s*import\s+([\w.*]+)(?:\s+as\s+(\w+))?`)

	for i, line := range lines {
		if matches := importRe.FindStringSubmatch(line); matches != nil {
			importPath := matches[1]
			// alias := matches[2] // Declared and not used

			imp := &model.Import{
				Path: importPath,
				Range: model.Range{
					Start: model.Position{Line: i + 1},
					End:   model.Position{Line: i + 1},
				},
			}

			// Alias field no longer exists in types.Import
			// If needed, the alias logic would need to be stored elsewhere
			// or handled differently by the MCP agents.

			result.Imports = append(result.Imports, imp)
		}
	}
}

func (p *KotlinParser) extractTypes(content string, result *parsing.ParseResult) {
	// Class, interface, object, data class, sealed class, enum class
	typeRe := regexp.MustCompile(`(?m)^\s*(?:@[\w.()]+)?\s*(?:abstract|open|final|sealed|data|enum|annotation)?\s*(class|interface|object)\s+(\w+)(?:\s*:\s*([\w<>,
()]+))?`)

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

		typeKind := content[match[6]:match[7]] // class, interface, object
		name := content[match[8]:match[9]]

		inheritance := ""
		if match[10] != -1 {
			inheritance = content[match[10]:match[11]]
		}

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		// Determine symbol type
		var symbolKind model.SymbolKind
		switch typeKind {
		case "class":
			symbolKind = model.SymbolKindClass
		case "interface":
			symbolKind = model.SymbolKindInterface
		case "object":
			symbolKind = model.SymbolKindClass // Singleton object
		default:
			symbolKind = model.SymbolKindClass
		}

		if modifier == "enum" {
			symbolKind = model.SymbolKindEnum
		}

		sig := ""
		if modifier != "" {
			sig = modifier + " "
		}
		sig += typeKind + " " + name
		if inheritance != "" {
			sig += " : " + inheritance
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
			// First is typically base class, rest are interfaces
			parts := strings.Split(inheritance, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				// Remove constructor calls
				if idx := strings.Index(part, "("); idx != -1 {
					part = part[:idx]
				}

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

func (p *KotlinParser) extractFunctions(content string, result *parsing.ParseResult) {
	// Function declaration
	funcRe := regexp.MustCompile(`(?m)^\s*(?:@[\w.()]+)?\s*(?:override|open|inline|suspend|operator|infix)?\s*fun\s+(?:<[^>]+>\s+)?(\w+)\s*\(([^)]*)\)(?:\s*:\s*([\w<>?]+))?`)

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

		sig := "fun " + name + "(" + params + ")"
		if returnType != "" {
			sig += ": " + returnType
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
}

func (p *KotlinParser) extractProperties(content string, result *parsing.ParseResult) {
	// Property declaration: val/var
	propRe := regexp.MustCompile(`(?m)^\s*(public|private|protected|internal)?\s*(override|const|lateinit)?\s*(val|var)\s+(\w+)\s*:\s*([\w<>?]+)`)

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

		valOrVar := content[match[6]:match[7]]
		name := content[match[8]:match[9]]
		propType := content[match[10]:match[11]]

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		sig := valOrVar + " " + name + ": " + propType

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

func (p *KotlinParser) parseVisibility(vis string) model.Visibility {
	switch strings.ToLower(vis) {
	case "public":
		return model.VisibilityPublic
	case "private":
		return model.VisibilityPrivate
	case "protected":
		return model.VisibilityProtected
	case "internal":
		return model.VisibilityInternal
	default:
		return model.VisibilityPublic // Kotlin default is public
	}
}