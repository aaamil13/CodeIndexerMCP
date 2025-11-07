package swift

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// SwiftParser parses Swift source code
type SwiftParser struct {
	*parser.BaseParser
}

// NewParser creates a new Swift parser
func NewParser() *SwiftParser {
	return &SwiftParser{
		BaseParser: parser.NewBaseParser("swift", []string{".swift"}, 100),
	}
}

// Parse parses Swift source code
func (p *SwiftParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
	result := &types.ParseResult{
		Symbols:       make([]*types.Symbol, 0),
		Imports:       make([]*types.Import, 0),
		Relationships: make([]*types.Relationship, 0),
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

func (p *SwiftParser) extractImports(lines []string, result *types.ParseResult) {
	importRe := regexp.MustCompile(`^\s*import\s+([\w.]+)`)

	for i, line := range lines {
		if matches := importRe.FindStringSubmatch(line); matches != nil {
			imp := &types.Import{
				Source:     matches[1],
				LineNumber: i + 1,
			}
			result.Imports = append(result.Imports, imp)
		}
	}
}

func (p *SwiftParser) extractTypes(content string, result *types.ParseResult) {
	// Class, struct, protocol, enum, actor
	typeRe := regexp.MustCompile(`(?m)^\s*(?:@[\w.()]+\s+)*(public|private|fileprivate|internal|open)?\s*(final|static)?\s*(class|struct|protocol|enum|actor)\s+(\w+)(?:\s*:\s*([\w,\s&]+))?`)

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
		switch typeKind {
		case "class", "actor":
			symbolType = types.SymbolTypeClass
		case "struct":
			symbolType = types.SymbolTypeStruct
		case "protocol":
			symbolType = types.SymbolTypeInterface
		case "enum":
			symbolType = types.SymbolTypeEnum
		default:
			symbolType = types.SymbolTypeClass
		}

		sig := typeKind + " " + name
		if inheritance != "" {
			sig += ": " + inheritance
		}

		symbol := &types.Symbol{
			Name:       name,
			Type:       symbolType,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: p.parseVisibility(visibility),
			Signature:  sig,
		}

		if modifier != "" {
			symbol.Metadata = map[string]interface{}{
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
					result.Relationships = append(result.Relationships, &types.Relationship{
						Type:       types.RelationshipExtends,
						SourceName: name,
						TargetName: part,
					})
				}
			}
		}
	}
}

func (p *SwiftParser) extractExtensions(content string, result *types.ParseResult) {
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

		symbol := &types.Symbol{
			Name:       name + " (extension)",
			Type:       types.SymbolTypeClass,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  sig,
			Metadata: map[string]interface{}{
				"extension": true,
			},
		}

		result.Symbols = append(result.Symbols, symbol)

		// Add conformances
		if protocols != "" {
			parts := strings.Split(protocols, ",")
			for _, protocol := range parts {
				protocol = strings.TrimSpace(protocol)
				if protocol != "" {
					result.Relationships = append(result.Relationships, &types.Relationship{
						Type:       types.RelationshipImplements,
						SourceName: name,
						TargetName: protocol,
					})
				}
			}
		}
	}
}

func (p *SwiftParser) extractFunctions(content string, result *types.ParseResult) {
	// Function/method declaration
	funcRe := regexp.MustCompile(`(?m)^\s*(?:@[\w.()]+\s+)*(public|private|fileprivate|internal|open)?\s*(static|class|mutating|override|final)?\s*func\s+(\w+)(?:<[^>]+>)?\s*\(([^)]*)\)(?:\s*(?:async|throws|rethrows))?\s*(?:->\s*([\w<>?]+))?`)

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

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeFunction,
			StartLine:  lineNum,
			EndLine:    lineNum,
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

		symbol := &types.Symbol{
			Name:       "init",
			Type:       types.SymbolTypeConstructor,
			StartLine:  lineNum,
			EndLine:    lineNum,
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
}

func (p *SwiftParser) extractProperties(content string, result *types.ParseResult) {
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

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeProperty,
			StartLine:  lineNum,
			EndLine:    lineNum,
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
}

func (p *SwiftParser) parseVisibility(vis string) types.Visibility {
	switch strings.ToLower(vis) {
	case "public", "open":
		return types.VisibilityPublic
	case "private":
		return types.VisibilityPrivate
	case "fileprivate":
		return types.VisibilityPrivate // File-scoped
	case "internal":
		return types.VisibilityInternal
	default:
		return types.VisibilityInternal // Swift default is internal
	}
}
