package csharp

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
)

// CSharpParser parses C# source code
type CSharpParser struct {
}

// NewParser creates a new C# parser
func NewParser() *CSharpParser {
	return &CSharpParser{}
}

// Language returns the language identifier (e.g., "go", "python", "typescript")
func (p *CSharpParser) Language() string {
	return "csharp"
}

// Extensions returns file extensions this parser handles (e.g., [".cs"])
func (p *CSharpParser) Extensions() []string {
	return []string{".cs"}
}

// Priority returns parser priority (higher = preferred when multiple parsers match)
func (p *CSharpParser) Priority() int {
	return 100
}

// SupportsFramework checks if parser supports specific framework analysis
func (p *CSharpParser) SupportsFramework(framework string) bool {
	return false
}

// Parse parses C# source code
func (p *CSharpParser) Parse(content []byte, filePath string) (*parsing.ParseResult, error) {
	result := &parsing.ParseResult{
		Symbols:       make([]*model.Symbol, 0),
		Imports:       make([]*model.Import, 0),
		Relationships: make([]*model.Relationship, 0),
		Metadata:      make(map[string]interface{}),
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Extract namespace
	p.extractNamespaces(lines, contentStr, result)

	// Extract using directives
	p.extractUsings(lines, result)

	// Extract types (classes, interfaces, structs, enums)
	p.extractTypes(lines, contentStr, result)

	// Extract members (methods, properties, fields)
	p.extractMembers(lines, contentStr, result)

	result.Metadata["language"] = "csharp"

	return result, nil
}

func (p *CSharpParser) extractNamespaces(lines []string, content string, result *parsing.ParseResult) {
	// File-scoped namespace (C# 10+)
	fileScopedNsRe := regexp.MustCompile(`(?m)^\s*namespace\s+([\w.]+)\s*;`)

	// Block-scoped namespace
	blockScopedNsRe := regexp.MustCompile(`(?m)^\s*namespace\s+([\w.]+)\s*{`)

	// Check file-scoped first
	if matches := fileScopedNsRe.FindStringSubmatch(content); matches != nil {
		result.Metadata["namespace"] = matches[1]
		lineNum := strings.Count(content[:strings.Index(content, matches[0])], "\n") + 1

		symbol := &model.Symbol{
			Name:       matches[1],
			Kind:       model.SymbolKindNamespace,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  "namespace " + matches[1],
		}
		result.Symbols = append(result.Symbols, symbol)
		return
	}

	// Check block-scoped namespaces
	nsMatches := blockScopedNsRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range nsMatches {
		name := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &model.Symbol{
			Name:       name,
			            Kind:       model.SymbolKindNamespace,
			            File:       "", // File path will be set by the caller
			            Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: p.findClosingBrace(lines, lineNum-1)}},			Visibility: model.VisibilityPublic,
			Signature:  "namespace " + name,
		}
		result.Symbols = append(result.Symbols, symbol)

		if result.Metadata["namespace"] == nil {
			result.Metadata["namespace"] = name
		}
	}
}

func (p *CSharpParser) extractUsings(lines []string, result *parsing.ParseResult) {
	usingRe := regexp.MustCompile(`^\s*using\s+(static\s+)?(?:([\w.]+)\s*=\s*)?([\w.]+)\s*;`)

	for i, line := range lines {
		if matches := usingRe.FindStringSubmatch(line); matches != nil {
			// isStatic := matches[1] != "" // Declared and not used
			// alias := matches[2] // Declared and not used
			namespace := matches[3]

			imp := &model.Import{
				Path: namespace,
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

func (p *CSharpParser) extractTypes(lines []string, content string, result *parsing.ParseResult) {
	// Type declaration: class, interface, struct, enum, record
	typeRe := regexp.MustCompile(`(?m)^\s*(?:[\w\s,()=]+\s*)*(public|private|protected|internal|protected\s+internal|private\s+protected)?\s*(abstract|sealed|static|partial)?\s*(class|interface|struct|enum|record(?:\s+(?:class|struct))?)\s+(\w+)(?:<[^>]+>)?(?:\s*:\s*([\w<>,\]+))?`)

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
		switch {
		case strings.HasPrefix(typeKind, "class"):
			symbolKind = model.SymbolKindClass
		case strings.HasPrefix(typeKind, "interface"):
			symbolKind = model.SymbolKindInterface
		case strings.HasPrefix(typeKind, "struct"):
			symbolKind = model.SymbolKindStruct
		case typeKind == "enum":
			symbolKind = model.SymbolKindEnum
		case strings.HasPrefix(typeKind, "record"):
			symbolKind = model.SymbolKindClass // Record is a special class
		default:
			symbolKind = model.SymbolKindClass
		}

		sig := typeKind + " " + name
		if inheritance != "" {
			sig += " : " + inheritance
		}

		symbol := &model.Symbol{
			Name:       name,
			Kind:       symbolKind,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: p.findClosingBrace(lines, lineNum-1)}},
			Visibility: p.parseVisibility(visibility),
			Signature:  sig,
		}

		if modifier != "" {
			symbol.Metadata = map[string]string{
				"modifier": modifier,
			}
		}

		result.Symbols = append(result.Symbols, symbol)

		// Add relationships for inheritance
		if inheritance != "" {
			for _, parent := range strings.Split(inheritance, ",") {
				parent = strings.TrimSpace(parent)
				// In C#, first is base class, rest are interfaces
				// We'll just mark all as "extends" for simplicity
				result.Relationships = append(result.Relationships, &model.Relationship{
					Type:       model.RelationshipKindExtends,
					SourceSymbol: name,
					TargetSymbol: parent,
				})
			}
		}
	}
}

func (p *CSharpParser) extractMembers(lines []string, content string, result *parsing.ParseResult) {
	// Method declaration
	methodRe := regexp.MustCompile(`(?m)^\s*(?:[\w\s,()=]+\s*)*(public|private|protected|internal)?\s*(static|virtual|override|abstract|async|extern)?\s*([\w<>[\]?]+)\s+(\w+)\s*(?:<[^>]+>)?\s*\(([^)]*)\)`)

	// Property declaration
	propertyRe := regexp.MustCompile(`(?m)^\s*(public|private|protected|internal)?\s*(static|virtual|override|abstract)?\s*([\w<>[\]?]+)\s+(\w+)\s*{\s*(?:get|set)`)

	// Field declaration
	fieldRe := regexp.MustCompile(`(?m)^\s*(public|private|protected|internal)?\s*(static|readonly|const|volatile)?\s*([\w<>[\]?]+)\s+(\w+)\s*[=;]`)

	// Extract methods
	methodMatches := methodRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range methodMatches {
		visibility := ""
		if match[2] != -1 {
			visibility = content[match[2]:match[3]]
		}

		modifier := ""
		if match[4] != -1 {
			modifier = content[match[4]:match[5]]
		}

		returnType := strings.TrimSpace(content[match[6]:match[7]])
		name := content[match[8]:match[9]]
		params := ""
		if match[10] != -1 {
			params = content[match[10]:match[11]]
		}

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		// Skip if return type is too long or looks weird
		if len(returnType) > 50 || strings.Contains(returnType, "{") {
			continue
		}

		sig := returnType + " " + name + "(" + params + ")"

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindMethod,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: p.findClosingBrace(lines, lineNum-1)}},
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

	// Extract properties
	propertyMatches := propertyRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range propertyMatches {
		visibility := ""
		if match[2] != -1 {
			visibility = content[match[2]:match[3]]
		}

		modifier := ""
		if match[4] != -1 {
			modifier = content[match[4]:match[5]]
		}

		propType := strings.TrimSpace(content[match[6]:match[7]])
		name := content[match[8]:match[9]]

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindProperty,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: p.parseVisibility(visibility),
			Signature:  propType + " " + name + " { get; set; }",
		}

		if modifier != "" {
			symbol.Metadata = map[string]string{
				"modifier": modifier,
			}
		}

		result.Symbols = append(result.Symbols, symbol)
	}

	// Extract fields
	fieldMatches := fieldRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range fieldMatches {
		visibility := ""
		if match[2] != -1 {
			visibility = content[match[2]:match[3]]
		}

		modifier := ""
		if match[4] != -1 {
			modifier = content[match[4]:match[5]]
		}

		fieldType := strings.TrimSpace(content[match[6]:match[7]])
		name := content[match[8]:match[9]]

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		// Skip if type is too long
		if len(fieldType) > 50 {
			continue
		}

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindField,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: p.parseVisibility(visibility),
			Signature:  fieldType + " " + name,
		}

		if modifier != "" {
			symbol.Metadata = map[string]string{
				"modifier": modifier,
			}
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *CSharpParser) parseVisibility(vis string) model.Visibility {
	switch strings.ToLower(strings.TrimSpace(vis)) {
	case "public":
		return model.VisibilityPublic
	case "private":
		return model.VisibilityPrivate
	case "protected":
		return model.VisibilityProtected
	case "internal", "protected internal", "private protected":
		return model.VisibilityInternal
	default:
		return model.VisibilityPrivate // C# default is private
	}
}

func (p *CSharpParser) findClosingBrace(lines []string, startLine int) int {
	braceCount := 0
	for i := startLine; i < len(lines); i++ {
		line := lines[i]
		for _, ch := range line {
			if ch == '{' {
				braceCount++
			} else if ch == '}' {
				braceCount--
				if braceCount == 0 {
					return i + 1
				}
			}
		}
	}
	return startLine + 1
}
