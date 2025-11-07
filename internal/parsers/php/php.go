package php

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// PHPParser parses PHP source code
type PHPParser struct {
	*parser.BaseParser
}

// NewParser creates a new PHP parser
func NewParser() *PHPParser {
	return &PHPParser{
		BaseParser: parser.NewBaseParser("php", []string{".php"}, 100),
	}
}

// Parse parses PHP source code
func (p *PHPParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
	result := &types.ParseResult{
		Symbols:       make([]*types.Symbol, 0),
		Imports:       make([]*types.Import, 0),
		Relationships: make([]*types.Relationship, 0),
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

func (p *PHPParser) extractNamespace(lines []string, result *types.ParseResult) {
	nsRe := regexp.MustCompile(`^\s*namespace\s+([\w\\]+)\s*;`)

	for i, line := range lines {
		if matches := nsRe.FindStringSubmatch(line); matches != nil {
			namespace := matches[1]
			result.Metadata["namespace"] = namespace

			symbol := &types.Symbol{
				Name:       namespace,
				Type:       types.SymbolTypeNamespace,
				StartLine:  i + 1,
				EndLine:    i + 1,
				Visibility: types.VisibilityPublic,
				Signature:  "namespace " + namespace,
			}
			result.Symbols = append(result.Symbols, symbol)
			break
		}
	}
}

func (p *PHPParser) extractUses(lines []string, result *types.ParseResult) {
	useRe := regexp.MustCompile(`^\s*use\s+([\w\\]+)(?:\s+as\s+(\w+))?\s*;`)

	for i, line := range lines {
		if matches := useRe.FindStringSubmatch(line); matches != nil {
			importPath := matches[1]
			alias := matches[2]

			imp := &types.Import{
				Source: importPath,
				Line:   i + 1,
			}

			if alias != "" {
				imp.Alias = alias
			}

			result.Imports = append(result.Imports, imp)
		}
	}
}

func (p *PHPParser) extractTypes(content string, result *types.ParseResult) {
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
		var symbolType types.SymbolType
		switch typeKind {
		case "class":
			symbolType = types.SymbolTypeClass
		case "interface":
			symbolType = types.SymbolTypeInterface
		case "trait":
			symbolType = types.SymbolTypeInterface // Treat trait as interface
		case "enum":
			symbolType = types.SymbolTypeEnum
		default:
			symbolType = types.SymbolTypeClass
		}

		sig := typeKind + " " + name
		if extendsClause != "" {
			sig += " extends " + extendsClause
		}
		if implementsClause != "" {
			sig += " implements " + implementsClause
		}

		symbol := &types.Symbol{
			Name:       name,
			Type:       symbolType,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  sig,
		}

		if modifier != "" {
			symbol.Metadata = map[string]interface{}{
				"modifier": modifier,
			}
		}

		result.Symbols = append(result.Symbols, symbol)

		// Add relationships
		if extendsClause != "" {
			result.Relationships = append(result.Relationships, &types.Relationship{
				Type:       types.RelationshipTypeExtends,
				SourceName: name,
				TargetName: extendsClause,
			})
		}

		if implementsClause != "" {
			for _, iface := range strings.Split(implementsClause, ",") {
				iface = strings.TrimSpace(iface)
				result.Relationships = append(result.Relationships, &types.Relationship{
					Type:       types.RelationshipTypeImplements,
					SourceName: name,
					TargetName: iface,
				})
			}
		}
	}
}

func (p *PHPParser) extractFunctions(content string, result *types.ParseResult) {
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
		symbolType := types.SymbolTypeFunction
		if visibility != "" {
			symbolType = types.SymbolTypeMethod
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

func (p *PHPParser) parseVisibility(vis string) types.Visibility {
	switch strings.ToLower(vis) {
	case "public":
		return types.VisibilityPublic
	case "private":
		return types.VisibilityPrivate
	case "protected":
		return types.VisibilityProtected
	default:
		return types.VisibilityPublic // PHP default is public
	}
}
