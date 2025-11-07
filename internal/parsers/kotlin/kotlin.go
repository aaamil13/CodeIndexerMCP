package kotlin

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// KotlinParser parses Kotlin source code
type KotlinParser struct {
	*parser.BaseParser
}

// NewParser creates a new Kotlin parser
func NewParser() *KotlinParser {
	return &KotlinParser{
		BaseParser: parser.NewBaseParser("kotlin", []string{".kt", ".kts"}, 100),
	}
}

// Parse parses Kotlin source code
func (p *KotlinParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
	result := &types.ParseResult{
		Symbols:       make([]*types.Symbol, 0),
		Imports:       make([]*types.Import, 0),
		Relationships: make([]*types.Relationship, 0),
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

func (p *KotlinParser) extractPackage(lines []string, result *types.ParseResult) {
	packageRe := regexp.MustCompile(`^\s*package\s+([\w.]+)`)

	for i, line := range lines {
		if matches := packageRe.FindStringSubmatch(line); matches != nil {
			result.Metadata["package"] = matches[1]

			symbol := &types.Symbol{
				Name:       matches[1],
				Type:       types.SymbolTypePackage,
				StartLine:  i + 1,
				EndLine:    i + 1,
				Visibility: types.VisibilityPublic,
				Signature:  "package " + matches[1],
			}
			result.Symbols = append(result.Symbols, symbol)
			break
		}
	}
}

func (p *KotlinParser) extractImports(lines []string, result *types.ParseResult) {
	importRe := regexp.MustCompile(`^\s*import\s+([\w.*]+)(?:\s+as\s+(\w+))?`)

	for i, line := range lines {
		if matches := importRe.FindStringSubmatch(line); matches != nil {
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

func (p *KotlinParser) extractTypes(content string, result *types.ParseResult) {
	// Class, interface, object, data class, sealed class, enum class
	typeRe := regexp.MustCompile(`(?m)^\s*(?:@[\w.()]+\s+)*(public|private|protected|internal)?\s*(abstract|open|final|sealed|data|enum|annotation)?\s*(class|interface|object)\s+(\w+)(?:\s*:\s*([\w<>,\s()]+))?`)

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
		var symbolType types.SymbolType
		switch typeKind {
		case "class":
			symbolType = types.SymbolTypeClass
		case "interface":
			symbolType = types.SymbolTypeInterface
		case "object":
			symbolType = types.SymbolTypeClass // Singleton object
		default:
			symbolType = types.SymbolTypeClass
		}

		if modifier == "enum" {
			symbolType = types.SymbolTypeEnum
		}

		sig := ""
		if modifier != "" {
			sig = modifier + " "
		}
		sig += typeKind + " " + name
		if inheritance != "" {
			sig += " : " + inheritance
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
			// First is typically base class, rest are interfaces
			parts := strings.Split(inheritance, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				// Remove constructor calls
				if idx := strings.Index(part, "("); idx != -1 {
					part = part[:idx]
				}

				if part != "" {
					result.Relationships = append(result.Relationships, &types.Relationship{
						Type:       types.RelationshipTypeExtends,
						SourceName: name,
						TargetName: part,
					})
				}
			}
		}
	}
}

func (p *KotlinParser) extractFunctions(content string, result *types.ParseResult) {
	// Function declaration
	funcRe := regexp.MustCompile(`(?m)^\s*(?:@[\w.()]+\s+)*(public|private|protected|internal)?\s*(override|open|inline|suspend|operator|infix)?\s*fun\s+(?:<[^>]+>\s+)?(\w+)\s*\(([^)]*)\)(?:\s*:\s*([\w<>?]+))?`)

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
}

func (p *KotlinParser) extractProperties(content string, result *types.ParseResult) {
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

func (p *KotlinParser) parseVisibility(vis string) types.Visibility {
	switch strings.ToLower(vis) {
	case "public":
		return types.VisibilityPublic
	case "private":
		return types.VisibilityPrivate
	case "protected":
		return types.VisibilityProtected
	case "internal":
		return types.VisibilityInternal
	default:
		return types.VisibilityPublic // Kotlin default is public
	}
}
