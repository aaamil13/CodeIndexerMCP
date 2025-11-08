package java

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
)

// JavaParser parses Java source code
type JavaParser struct {
}

// NewParser creates a new Java parser
func NewParser() *JavaParser {
	return &JavaParser{}
}

// Language returns the language identifier (e.g., "java")
func (p *JavaParser) Language() string {
	return "java"
}

// Extensions returns file extensions this parser handles (e.g., [".java"])
func (p *JavaParser) Extensions() []string {
	return []string{`.java`}
}

// Priority returns parser priority (higher = preferred when multiple parsers match)
func (p *JavaParser) Priority() int {
	return 100
}

// SupportsFramework checks if parser supports specific framework analysis
func (p *JavaParser) SupportsFramework(framework string) bool {
	return false
}

// Parse parses Java source code
func (p *JavaParser) Parse(content []byte, filePath string) (*parsing.ParseResult, error) {
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

	// Extract classes and interfaces
	p.extractTypes(lines, contentStr, result)

	// Extract methods and fields
	p.extractMembers(lines, contentStr, result)

	result.Metadata["language"] = "java"

	return result, nil
}

func (p *JavaParser) extractPackage(lines []string, result *parsing.ParseResult) {
	packageRe := regexp.MustCompile(`^\s*package\s+([\w.]+)\s*;`)

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

func (p *JavaParser) extractImports(lines []string, result *parsing.ParseResult) {
	importRe := regexp.MustCompile(`^\s*import\s+(static\s+)?([\w.*]+)\s*;`)

	for i, line := range lines {
		if matches := importRe.FindStringSubmatch(line); matches != nil {
			// isStatic := matches[1] != "" // Declared and not used, removed this line
			importPath := matches[2]

			imp := &model.Import{
				Path: importPath,
				Range: model.Range{
					Start: model.Position{Line: i + 1},
					End:   model.Position{Line: i + 1},
				},
			}

			// Alias field no longer exists in types.Import
			// The `isStatic` information can potentially be stored in Metadata if needed.

			result.Imports = append(result.Imports, imp)
		}
	}
}

func (p *JavaParser) extractTypes(lines []string, content string, result *parsing.ParseResult) {
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
		var symbolKind model.SymbolKind
		switch typeKind {
		case "class":
			symbolKind = model.SymbolKindClass
		case "interface":
			symbolKind = model.SymbolKindInterface
		case "enum":
			symbolKind = model.SymbolKindEnum
		case "@interface":
			symbolKind = model.SymbolKindInterface // Annotation
		default:
			symbolKind = model.SymbolKindClass
		}

		// Build signature
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

		// Add relationships for extends
		if extendsClause != "" {
			for _, parent := range strings.Split(extendsClause, ",") {
				parent = strings.TrimSpace(parent)
				result.Relationships = append(result.Relationships, &model.Relationship{
					Type:       model.RelationshipKindExtends,
					SourceSymbol: name,
					TargetSymbol: parent,
				})
			}
		}

		// Add relationships for implements
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

func (p *JavaParser) extractMembers(lines []string, content string, result *parsing.ParseResult) {
	// Method declaration
	methodRe := regexp.MustCompile(`(?m)^\s*(public|private|protected)?\s*(static|final|abstract|synchronized|native)?\s*([\w<>[\]\s]+)\s+(\w+)\s*\(([^)]*)\)(?:\s+throws\s+([\w,\s]+))?`)

	// Field declaration
	fieldRe := regexp.MustCompile(`(?m)^\s*(public|private|protected)?\s*(static|final|transient|volatile)?\s*([\w<>[\]\s]+)\s+(\w+)\s*[=;]`)

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

func (p *JavaParser) parseVisibility(vis string) model.Visibility {
	switch strings.ToLower(vis) {
	case "public":
		return model.VisibilityPublic
	case "private":
		return model.VisibilityPrivate
	case "protected":
		return model.VisibilityProtected
	default:
		return model.VisibilityPackage // Java default is package-private
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
