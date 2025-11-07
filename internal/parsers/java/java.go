package java

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// JavaParser parses Java source code
type JavaParser struct {
	*parser.BaseParser
}

// NewParser creates a new Java parser
func NewParser() *JavaParser {
	return &JavaParser{
		BaseParser: parser.NewBaseParser("java", []string{".java"}, 100),
	}
}

// Parse parses Java source code
func (p *JavaParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
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

	// Extract classes and interfaces
	p.extractTypes(lines, contentStr, result)

	// Extract methods and fields
	p.extractMembers(lines, contentStr, result)

	result.Metadata["language"] = "java"

	return result, nil
}

func (p *JavaParser) extractPackage(lines []string, result *types.ParseResult) {
	packageRe := regexp.MustCompile(`^\s*package\s+([\w.]+)\s*;`)

	for i, line := range lines {
		if matches := packageRe.FindStringSubmatch(line); matches != nil { // Fixed: used packageRe
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

func (p *JavaParser) extractImports(lines []string, result *types.ParseResult) {
	importRe := regexp.MustCompile(`^\s*import\s+(static\s+)?([\w.*]+)\s*;`)

	for i, line := range lines {
		if matches := importRe.FindStringSubmatch(line); matches != nil {
			// isStatic := matches[1] != "" // Declared and not used, removed this line
			importPath := matches[2]

			imp := &types.Import{
				Source:     importPath,
				LineNumber: i + 1,
			}

			// Alias field no longer exists in types.Import
			// The `isStatic` information can potentially be stored in Metadata if needed.

			result.Imports = append(result.Imports, imp)
		}
	}
}

func (p *JavaParser) extractTypes(lines []string, content string, result *types.ParseResult) {
	// Class declaration
	classRe := regexp.MustCompile(`(?m)^\s*(public|private|protected)?\s*(abstract|final|static)?\s*(class|interface|enum|@interface)\s+(\w+)(?:\s+extends\s+([\w<>,\s]+))?(?:\s+implements\s+([\w<>,\s]+))?`)

	matches := classRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		visibility := ""
		if match[2] != -1 {
			visibility = content[match[2]:match[3]]
		}

		modifier := ""
		if match[4] != -1 {
			modifier = content[match[4]:match[5]]
		}

		typeKind := content[match[6]:match[7]] // class, interface, enum, @interface
		name := content[match[8]:match[9]]

		extendsClause := ""
		if match[10] != -1 {
			extendsClause = content[match[10]:match[11]]
		}

		implementsClause := ""
		if match[12] != -1 {
			implementsClause = content[match[12]:match[13]]
		}

		// Calculate line number
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		// Determine symbol type
		var symbolType types.SymbolType
		switch typeKind {
		case "class":
			symbolType = types.SymbolTypeClass
		case "interface":
			symbolType = types.SymbolTypeInterface
		case "enum":
			symbolType = types.SymbolTypeEnum
		case "@interface":
			symbolType = types.SymbolTypeInterface // Annotation
		default:
			symbolType = types.SymbolTypeClass
		}

		// Build signature
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

		// Add relationships for extends
		if extendsClause != "" {
			for _, parent := range strings.Split(extendsClause, ",") {
				parent = strings.TrimSpace(parent)
				result.Relationships = append(result.Relationships, &types.Relationship{
					Type:       types.RelationshipExtends,
					SourceName: name,
					TargetName: parent,
				})
			}
		}

		// Add relationships for implements
		if implementsClause != "" {
			for _, iface := range strings.Split(implementsClause, ",") {
				iface = strings.TrimSpace(iface)
				result.Relationships = append(result.Relationships, &types.Relationship{
					Type:       types.RelationshipImplements,
					SourceName: name,
					TargetName: iface,
				})
			}
		}
	}
}

func (p *JavaParser) extractMembers(lines []string, content string, result *types.ParseResult) {
	// Method declaration
	methodRe := regexp.MustCompile(`(?m)^\s*(public|private|protected)?\s*(static|final|abstract|synchronized|native)?\s*([\w<>\[\],\s]+)\s+(\w+)\s*\(([^)]*)\)(?:\s+throws\s+([\w,\s]+))?`)

	// Field declaration
	fieldRe := regexp.MustCompile(`(?m)^\s*(public|private|protected)?\s*(static|final|transient|volatile)?\s*([\w<>\[\],\s]+)\s+(\w+)\s*[=;]`)

	// Extract methods
	methodMatches := methodRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range methodMatches {
		// Skip if this looks like a field (has = or ; right after)
		afterMatch := match[1]
		if afterMatch < len(content) && (content[afterMatch] == '=' || content[afterMatch] == ';') {
			continue
		}

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

		throwsClause := ""
		if match[12] != -1 {
			throwsClause = content[match[12]:match[13]]
		}

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		// Skip constructor-like patterns if return type looks weird
		if len(returnType) > 50 || strings.Contains(returnType, "{") {
			continue
		}

		sig := returnType + " " + name + "(" + params + ")"
		if throwsClause != "" {
			sig += " throws " + throwsClause
		}

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

		// Skip if type is too long (likely not a field)
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

func (p *JavaParser) parseVisibility(vis string) types.Visibility {
	switch strings.ToLower(vis) {
	case "public":
		return types.VisibilityPublic
	case "private":
		return types.VisibilityPrivate
	case "protected":
		return types.VisibilityProtected
	default:
		return types.VisibilityPackage // Java default is package-private
	}
}

func (p *JavaParser) findClosingBrace(lines []string, startLine int) int {
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
