package cpp

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
)

// CppParser parses C++ source code
type CppParser struct {
}

// NewParser creates a new C++ parser
func NewParser() *CppParser {
	return &CppParser{}
}

// Language returns the language identifier (e.g., "go", "python", "typescript")
func (p *CppParser) Language() string {
	return "cpp"
}

// Extensions returns file extensions this parser handles (e.g., [".cpp", ".h"])
func (p *CppParser) Extensions() []string {
	return []string{'.cpp', '.cc', '.cxx', '.hpp', '.h', '.hxx'}
}

// Priority returns parser priority (higher = preferred when multiple parsers match)
func (p *CppParser) Priority() int {
	return 100
}

// SupportsFramework checks if parser supports specific framework analysis
func (p *CppParser) SupportsFramework(framework string) bool {
	return false
}

// Parse parses C++ source code
func (p *CppParser) Parse(content []byte, filePath string) (*parsing.ParseResult, error) {
	result := &parsing.ParseResult{
		Symbols:       make([]*model.Symbol, 0),
		Imports:       make([]*model.Import, 0),
		Relationships: make([]*model.Relationship, 0),
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

func (p *CppParser) extractIncludes(lines []string, result *parsing.ParseResult) {
	includeRe := regexp.MustCompile(`^\s*#\s*include\s+[<\"]([^>\"]+)[>\"]`)

	for i, line := range lines {
		if matches := includeRe.FindStringSubmatch(line); matches != nil {
			imp := &model.Import{
				Path: matches[1],
				Range: model.Range{
					Start: model.Position{Line: i + 1},
					End:   model.Position{Line: i + 1},
				},
			}
			result.Imports = append(result.Imports, imp)
		}
	}
}

func (p *CppParser) extractNamespaces(content string, result *parsing.ParseResult) {
	nsRe := regexp.MustCompile(`(?m)^\s*namespace\s+(\w+)(?:\s*::\s*\w+)*\s*{`)

	matches := nsRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindNamespace,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  "namespace " + name,
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *CppParser) extractClasses(content string, result *parsing.ParseResult) {
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

		var symbolKind model.SymbolKind
		if typeKind == "struct" {
			symbolKind = model.SymbolKindStruct
		} else {
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
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
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

func (p *CppParser) extractFunctions(content string, result *parsing.ParseResult) {
	// Function declaration/definition
	// Including member functions and operators
	funcRe := regexp.MustCompile(`(?m)^\s*(?:virtual|static|inline|explicit|constexpr|friend)?\s*([\w:<>,
\s\*&]+?)\s+(\w+|operator\s*[+\-*/=<>!&|]+)\s*\(([^)]*)\)\s*(?:const|override|final|noexcept)?\s*(?:{|;|.=)`) // Added '=' to the end of the regex

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
		symbolKind := model.SymbolKindFunction
		if strings.Contains(name, "::") || strings.HasPrefix(name, "operator") {
			symbolKind = model.SymbolKindMethod
		}

		symbol := &model.Symbol{
			Name:       name,
			Kind:       symbolKind,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  sig,
		}

		result.Symbols = append(result.Symbols, symbol)
	}

	// Constructor/Destructor
	ctorRe := regexp.MustCompile(`(?m)^\s*(?:explicit\s+)?(\w+)::\1\s*\(([^)]*)\)`) // Removed trailing '='
	dtorRe := regexp.MustCompile(`(?m)^\s*(?:virtual\s+)?(\w+):~\1\s*\(\)`)

	// Constructors
	ctorMatches := ctorRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range ctorMatches {
		className := content[match[2]:match[3]]
		params := ""
		if match[4] != -1 {
			params = content[match[4]:match[5]]
		}

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &model.Symbol{
			Name:       className + "::" + className,
			Kind:       model.SymbolKindConstructor,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  className + "(" + params + ")",
		}

		result.Symbols = append(result.Symbols, symbol)
	}

	// Destructors
	dtorMatches := dtorRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range dtorMatches {
		className := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &model.Symbol{
			Name:       className + "::~" + className,
			Kind:       model.SymbolKindMethod, // Destructors are methods
			File:       "",                     // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  "~" + className + "()",
			Metadata: map[string]string{
				"destructor": "true",
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *CppParser) extractTemplates(content string, result *parsing.ParseResult) {
	// Template class/function
	templateRe := regexp.MustCompile(`(?m)^\s*template\s*<([^>]+)>\s*(?:class|struct|typename)\s+(\w+)`)

	matches := templateRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		templateParams := content[match[2]:match[3]]
		name := content[match[4]:match[5]]

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindClass, // Templates are often classes or functions, here we generalize to class
			File:       "",                    // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  "template<" + templateParams + "> class " + name,
			Metadata: map[string]string{
				"template": "true",
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}