package php

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
)

// PHPParser parses PHP source code
type PHPParser struct {
}

// NewParser creates a new PHP parser
func NewParser() *PHPParser {
	return &PHPParser{}
}

// Language returns the language identifier (e.g., "php")
func (p *PHPParser) Language() string {
	return "php"
}

// Extensions returns file extensions this parser handles (e.g., [".php"])
func (p *PHPParser) Extensions() []string {
	return []string{”.php"}
}

// Priority returns parser priority (higher = preferred when multiple parsers match)
func (p *PHPParser) Priority() int {
	return 100
}

// SupportsFramework checks if parser supports specific framework analysis
func (p *PHPParser) SupportsFramework(framework string) bool {
	return false
}

// Parse parses PHP source code
func (p *PHPParser) Parse(content []byte, filePath string) (*parsing.ParseResult, error) {
	result := &parsing.ParseResult{
		Symbols:       make([]*model.Symbol, 0),
		Imports:       make([]*model.Import, 0),
		Relationships: make([]*model.Relationship, 0),
		Metadata:      make(map[string]interface{}),
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Extract namespace
	p.extractNamespace(lines, result)

	// Extract use statements (imports)
	p.extractUses(lines, result)

	// Extract classes, interfaces, traits
	p.extractTypes(contentStr, result)

	// Extract functions and methods
	p.extractFunctions(contentStr, result)

	result.Metadata["language"] = "php"

	return result, nil
}

func (p *PHPParser) extractNamespace(lines []string, result *parsing.ParseResult) {
	nsRe := regexp.MustCompile(`^\s*namespace\s+([\w\\]+)\s*;`)

	for i, line := range lines {
		if matches := nsRe.FindStringSubmatch(line); matches != nil {
			namespace := matches[1]
			result.Metadata["namespace"] = namespace

			symbol := &model.Symbol{
				Name:       namespace,
				Kind:       model.SymbolKindNamespace,
				File:       "", // File path will be set by the caller
				Range:      model.Range{Start: model.Position{Line: i + 1}, End: model.Position{Line: i + 1}},
				Visibility: model.VisibilityPublic,
				Signature:  "namespace " + namespace,
			}
			result.Symbols = append(result.Symbols, symbol)
			break
		}
	}
}

func (p *PHPParser) extractUses(lines []string, result *parsing.ParseResult) {
	useRe := regexp.MustCompile(`^\s*use\s+([\w\\]+)(?:\s+as\s+(\w+))?\s*;`)

	for i, line := range lines {
		if matches := useRe.FindStringSubmatch(line); matches != nil {
			importPath := matches[1]
			// alias := matches[2] // Declared and not used

			im := &model.Import{
				Path: importPath,
				Range: model.Range{
					Start: model.Position{Line: i + 1},
					End:   model.Position{Line: i + 1},
				},
			}

			// If alias is still needed, it would need to be stored elsewhere,
			// for example in the Metadata field of the Import struct.
			// For now, removing the assignment to avoid compilation error.

			result.Imports = append(result.Imports, imp)
		}
	}
}

func (p *PHPParser) extractTypes(content string, result *parsing.ParseResult) {
	// Class, interface, trait, enum (PHP 8.1+)
	typeRe := regexp.MustCompile(`(?m)^\s*(abstract|final)?\s*(class|interface|trait|enum)\s+(\w+)(?:\s+extends\s+([\w\\]+))?(?:\s+implements\s+([\w\\,\s]+))?`)

	matches := typeRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		modifier := ""
		if match[2] != -1 {
			modifier = content[match[2]:match[3]]
		}

		typeKind := content[match[4]:match[5]] // class, interface, trait, enum
		name := content[match[6]:match[7]]

		extendsClause := ""
		if match[8] != -1 {
			extendsClause = content[match[8]:match[9]]
		}

		implementsClause := ""
		if match[10] != -1 {
			implementsClause = content[match[10]:match[11]]
		}

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		// Determine symbol type
		var symbolKind model.SymbolKind
		switch typeKind {
		case "class":
			symbolKind = model.SymbolKindClass
		case "interface":
			symbolKind = model.SymbolKindInterface
		case "trait":
			symbolKind = model.SymbolKindInterface // Treat trait as interface
		case "enum":
			symbolKind = model.SymbolKindEnum
		default:
			symbolKind = model.SymbolKindClass
		}

		sig := typeKind + " " + name
		if extendsClause != "" {
			sig += " extends " + extendsClause
		}
		if implementsClause != "" {
			sig += " implements " + implementsClause
		}

		symbol := &model.Symbol{
			Name:       name,
			Kind:       symbolKind,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  sig,
		}

		if modifier != "" {
			symbol.Metadata = map[string]string{
				"modifier": modifier,
			}
		}

		result.Symbols = append(result.Symbols, symbol)

		// Add relationships
		if extendsClause != "" {
			result.Relationships = append(result.Relationships, &model.Relationship{
				Type:       model.RelationshipKindExtends,
				SourceSymbol: name,
				TargetSymbol: extendsClause,
			})
		}

		if implementsClause != "" {
			for _, iface := range strings.Split(implementsClause, ",") {
				iface = strings.TrimSpace(iface)
				result.Relationships = append(result.Relationships, &model.Relationship{
					Type:       model.RelationshipKindImplements,
					SourceSymbol: name,
					TargetSymbol: iface,
				})
			}
		}
	}
}

func (p *PHPParser) extractFunctions(content string, result *parsing.ParseResult) {
	// Function or method declaration
	funcRe := regexp.MustCompile(`(?m)^\s*(public|private|protected)?\s*(static|abstract|final)?\s*function\s+(&)?(\w+)\s*\(([^)]*)\)(?:\s*:\s*([\w\\|?]+))?`)

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

		byRef := false
		if match[6] != -1 {
			byRef = true
		}

		name := content[match[8]:match[9]]
		params := ""
		if match[10] != -1 {
			params = content[match[10]:match[11]]
		}

		returnType := ""
		if match[12] != -1 {
			returnType = content[match[12]:match[13]]
		}

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		sig := "function "
		if byRef {
			sig += "&"
		}
		sig += name + "(" + params + ")"
		if returnType != "" {
			sig += ": " + returnType
		}

		// Determine if method or function based on visibility
		symbolKind := model.SymbolKindFunction
		if visibility != "" {
			symbolKind = model.SymbolKindMethod
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
	}

	// Properties
	propRe := regexp.MustCompile(`(?m)^\s*(public|private|protected)\s*(static|readonly)?\s*([\w\\|?]+)?\s*\$(\w+)`)

	propMatches := propRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range propMatches {
		visibility := content[match[2]:match[3]]

		modifier := ""
		if match[4] != -1 {
			modifier = content[match[4]:match[5]]
		}

		propType := ""
		if match[6] != -1 {
			propType = content[match[6]:match[7]]
		}

		name := content[match[8]:match[9]]

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		sig := "$" + name
		if propType != "" {
			sig = propType + " $" + name
		}

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

func (p *PHPParser) parseVisibility(vis string) model.Visibility {
	switch strings.ToLower(vis) {
	case "public":
		return model.VisibilityPublic
	case "private":
		return model.VisibilityPrivate
	case "protected":
		return model.VisibilityProtected
	default:
		return model.VisibilityPublic // PHP default is public
	}
}