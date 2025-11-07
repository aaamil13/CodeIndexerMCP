package csharp

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// CSharpParser parses C# source code
type CSharpParser struct {
	*parser.BaseParser
}

// NewParser creates a new C# parser
func NewParser() *CSharpParser {
	return &CSharpParser{
		BaseParser: parser.NewBaseParser("csharp", []string{".cs"}, 100),
	}
}

// Parse parses C# source code
func (p *CSharpParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
	result := &types.ParseResult{
		Symbols:       make([]*types.Symbol, 0),
		Imports:       make([]*types.Import, 0),
		Relationships: make([]*types.Relationship, 0),
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

func (p *CSharpParser) extractNamespaces(lines []string, content string, result *types.ParseResult) {
	// File-scoped namespace (C# 10+)
	fileScopedNsRe := regexp.MustCompile(`(?m)^\s*namespace\s+([\w.]+)\s*;`)

	// Block-scoped namespace
	blockScopedNsRe := regexp.MustCompile(`(?m)^\s*namespace\s+([\w.]+)\s*{`)

	// Check file-scoped first
	if matches := fileScopedNsRe.FindStringSubmatch(content); matches != nil {
		result.Metadata["namespace"] = matches[1]
		lineNum := strings.Count(content[:strings.Index(content, matches[0])], "\n") + 1

		symbol := &types.Symbol{
			Name:       matches[1],
			Type:       types.SymbolTypeNamespace,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
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

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeNamespace,
			StartLine:  lineNum,
			EndLine:    p.findClosingBrace(lines, lineNum-1),
			Visibility: types.VisibilityPublic,
			Signature:  "namespace " + name,
		}
		result.Symbols = append(result.Symbols, symbol)

		if result.Metadata["namespace"] == nil {
			result.Metadata["namespace"] = name
		}
	}
}

func (p *CSharpParser) extractUsings(lines []string, result *types.ParseResult) {
	usingRe := regexp.MustCompile(`^\s*using\s+(static\s+)?(?:([\w.]+)\s*=\s*)?([\w.]+)\s*;`)

	for i, line := range lines {
		if matches := usingRe.FindStringSubmatch(line); matches != nil {
			// isStatic := matches[1] != "" // Declared and not used
			// alias := matches[2] // Declared and not used
			namespace := matches[3]

			imp := &types.Import{
				Source:     namespace,
				LineNumber: i + 1,
			}

			// Alias field no longer exists in types.Import
			// If needed, the alias logic would need to be stored elsewhere
			// or handled differently by the MCP agents.

			result.Imports = append(result.Imports, imp)
		}
	}
}

func (p *CSharpParser) extractTypes(lines []string, content string, result *types.ParseResult) {
	// Type declaration: class, interface, struct, enum, record
	typeRe := regexp.MustCompile(`(?m)^\s*(?:\[[\w\s,()=]+\]\s*)*(public|private|protected|internal|protected\s+internal|private\s+protected)?\s*(abstract|sealed|static|partial)?\s*(class|interface|struct|enum|record(?:\s+(?:class|struct))?)\s+(\w+)(?:<[^>]+>)?(?:\s*:\s*([\w<>,\s]+))?`)

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
		var symbolType types.SymbolType
		switch {
		case strings.HasPrefix(typeKind, "class"):
			symbolType = types.SymbolTypeClass
		case strings.HasPrefix(typeKind, "interface"):
			symbolType = types.SymbolTypeInterface
		case strings.HasPrefix(typeKind, "struct"):
			symbolType = types.SymbolTypeStruct
		case typeKind == "enum":
			symbolType = types.SymbolTypeEnum
		case strings.HasPrefix(typeKind, "record"):
			symbolType = types.SymbolTypeClass // Record is a special class
		default:
			symbolType = types.SymbolTypeClass
		}

		sig := typeKind + " " + name
		if inheritance != "" {
			sig += " : " + inheritance
		}

		symbol := &types.Symbol{
			Name:       name,
			Type:       symbolType,
			StartLine:  lineNum,
			EndLine:    p.findClosingBrace(lines, lineNum-1),
			Visibility: p.parseVisibility(visibility),
			Signature:  sig,
		}

		if modifier != "" {
			symbol.Metadata = map[string]interface{}{
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
				result.Relationships = append(result.Relationships, &types.Relationship{
					Type:       types.RelationshipExtends, // Corrected constant name
					SourceName: name,
					TargetName: parent,
				})
			}
		}
	}
}

func (p *CSharpParser) extractMembers(lines []string, content string, result *types.ParseResult) {
	// Method declaration
	methodRe := regexp.MustCompile(`(?m)^\s*(?:\[[\w\s,()=]+\]\s*)*(public|private|protected|internal)?\s*(static|virtual|override|abstract|async|extern)?\s*([\w<>\[\]?]+)\s+(\w+)\s*(?:<[^>]+>)?\s*\(([^)]*)\)`)

	// Property declaration
	propertyRe := regexp.MustCompile(`(?m)^\s*(public|private|protected|internal)?\s*(static|virtual|override|abstract)?\s*([\w<>\[\]?]+)\s+(\w+)\s*{\s*(?:get|set)`)

	// Field declaration
	fieldRe := regexp.MustCompile(`(?m)^\s*(public|private|protected|internal)?\s*(static|readonly|const|volatile)?\s*([\w<>\[\]?]+)\s+(\w+)\s*[=;]`)

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

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeMethod,
			StartLine:  lineNum,
			EndLine:    p.findClosingBrace(lines, lineNum-1),
			Visibility: p.parseVisibility(visibility),
			Signature:  sig,
		}

		if modifier != "" {
			symbol.Metadata = map[string]interface{}{
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

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeProperty,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: p.parseVisibility(visibility),
			Signature:  propType + " " + name + " { get; set; }",
		}

		if modifier != "" {
			symbol.Metadata = map[string]interface{}{
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

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeField,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: p.parseVisibility(visibility),
			Signature:  fieldType + " " + name,
		}

		if modifier != "" {
			symbol.Metadata = map[string]interface{}{
				"modifier": modifier,
			}
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *CSharpParser) parseVisibility(vis string) types.Visibility {
	switch strings.ToLower(strings.TrimSpace(vis)) {
	case "public":
		return types.VisibilityPublic
	case "private":
		return types.VisibilityPrivate
	case "protected":
		return types.VisibilityProtected
	case "internal", "protected internal", "private protected":
		return types.VisibilityInternal
	default:
		return types.VisibilityPrivate // C# default is private
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
