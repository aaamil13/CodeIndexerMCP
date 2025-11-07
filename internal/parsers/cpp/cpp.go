package cpp

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// CppParser parses C++ source code
type CppParser struct {
	*parser.BaseParser
}

// NewParser creates a new C++ parser
func NewParser() *CppParser {
	return &CppParser{
		BaseParser: parser.NewBaseParser("cpp", []string{".cpp", ".cc", ".cxx", ".hpp", ".h", ".hxx"}, 100),
	}
}

// Parse parses C++ source code
func (p *CppParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
	result := &types.ParseResult{
		Symbols:       make([]*types.Symbol, 0),
		Imports:       make([]*types.Import, 0),
		Relationships: make([]*types.Relationship, 0),
		Metadata:      make(map[string]interface{}),
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Extract includes
	p.extractIncludes(lines, result)

	// Extract namespaces
	p.extractNamespaces(contentStr, result)

	// Extract classes and structs
	p.extractClasses(contentStr, result)

	// Extract functions
	p.extractFunctions(contentStr, result)

	// Extract templates
	p.extractTemplates(contentStr, result)

	result.Metadata["language"] = "cpp"
	result.Metadata["type"] = "source"
	if strings.HasSuffix(filePath, ".h") || strings.HasSuffix(filePath, ".hpp") || strings.HasSuffix(filePath, ".hxx") {
		result.Metadata["type"] = "header"
	}

	return result, nil
}

func (p *CppParser) extractIncludes(lines []string, result *types.ParseResult) {
	includeRe := regexp.MustCompile(`^\s*#\s*include\s+[<"]([^>"]+)[>"]`)

	for i, line := range lines {
		if matches := includeRe.FindStringSubmatch(line); matches != nil {
			imp := &types.Import{
				Source: matches[1],
				Line:   i + 1,
			}
			result.Imports = append(result.Imports, imp)
		}
	}
}

func (p *CppParser) extractNamespaces(content string, result *types.ParseResult) {
	nsRe := regexp.MustCompile(`(?m)^\s*namespace\s+(\w+)(?:\s*::\s*\w+)*\s*{`)

	matches := nsRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeNamespace,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  "namespace " + name,
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *CppParser) extractClasses(content string, result *types.ParseResult) {
	// Class/struct declaration
	classRe := regexp.MustCompile(`(?m)^\s*(?:template\s*<[^>]+>\s*)?(class|struct)\s+(?:__declspec\([^)]+\)\s+)?(\w+)(?:\s*:\s*((?:public|protected|private)\s+[\w:,\s<>]+))?(?:\s*{|;)`)

	matches := classRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		typeKind := content[match[2]:match[3]] // class or struct
		name := content[match[4]:match[5]]

		inheritance := ""
		if match[6] != -1 {
			inheritance = content[match[6]:match[7]]
		}

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		var symbolType types.SymbolType
		if typeKind == "struct" {
			symbolType = types.SymbolTypeStruct
		} else {
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
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  sig,
		}

		result.Symbols = append(result.Symbols, symbol)

		// Add inheritance relationships
		if inheritance != "" {
			// Parse: public BaseClass, private Interface1
			parts := strings.Split(inheritance, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				// Remove access specifier
				for _, prefix := range []string{"public ", "protected ", "private "} {
					part = strings.TrimPrefix(part, prefix)
				}
				part = strings.TrimSpace(part)

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

func (p *CppParser) extractFunctions(content string, result *types.ParseResult) {
	// Function declaration/definition
	// Including member functions and operators
	funcRe := regexp.MustCompile(`(?m)^\s*(?:virtual|static|inline|explicit|constexpr|friend)?\s*([\w:<>,\s\*&]+?)\s+(\w+|operator\s*[+\-*/=<>!&|]+)\s*\(([^)]*)\)\s*(?:const|override|final|noexcept)?\s*(?:{|;|=)`)

	matches := funcRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		returnType := strings.TrimSpace(content[match[2]:match[3]])
		name := strings.TrimSpace(content[match[4]:match[5]])
		params := ""
		if match[6] != -1 {
			params = content[match[6]:match[7]]
		}

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		// Skip if return type is too long or looks like template/preprocessor
		if len(returnType) > 100 || strings.Contains(returnType, "#") || strings.Contains(returnType, "template") {
			continue
		}

		// Skip control flow keywords
		if name == "if" || name == "while" || name == "for" || name == "switch" {
			continue
		}

		// Skip destructors (they'll often be caught by this pattern)
		if strings.HasPrefix(name, "~") {
			continue
		}

		sig := returnType + " " + name + "(" + params + ")"

		// Determine if it's a method (has :: in name) or function
		symbolType := types.SymbolTypeFunction
		if strings.Contains(name, "::") || strings.HasPrefix(name, "operator") {
			symbolType = types.SymbolTypeMethod
		}

		symbol := &types.Symbol{
			Name:       name,
			Type:       symbolType,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  sig,
		}

		result.Symbols = append(result.Symbols, symbol)
	}

	// Constructor/Destructor
	ctorRe := regexp.MustCompile(`(?m)^\s*(?:explicit\s+)?(\w+)::\1\s*\(([^)]*)\)`)
	dtorRe := regexp.MustCompile(`(?m)^\s*(?:virtual\s+)?(\w+)::~\1\s*\(\)`)

	// Constructors
	ctorMatches := ctorRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range ctorMatches {
		className := content[match[2]:match[3]]
		params := ""
		if match[4] != -1 {
			params = content[match[4]:match[5]]
		}

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &types.Symbol{
			Name:       className + "::" + className,
			Type:       types.SymbolTypeConstructor,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  className + "(" + params + ")",
		}

		result.Symbols = append(result.Symbols, symbol)
	}

	// Destructors
	dtorMatches := dtorRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range dtorMatches {
		className := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &types.Symbol{
			Name:       className + "::~" + className,
			Type:       types.SymbolTypeMethod,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  "~" + className + "()",
			Metadata: map[string]interface{}{
				"destructor": true,
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *CppParser) extractTemplates(content string, result *types.ParseResult) {
	// Template class/function
	templateRe := regexp.MustCompile(`(?m)^\s*template\s*<([^>]+)>\s*(?:class|struct|typename)\s+(\w+)`)

	matches := templateRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		templateParams := content[match[2]:match[3]]
		name := content[match[4]:match[5]]

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeClass,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  "template<" + templateParams + "> class " + name,
			Metadata: map[string]interface{}{
				"template": true,
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}
