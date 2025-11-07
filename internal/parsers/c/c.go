package c

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// CParser parses C source code
type CParser struct {
	*parser.BaseParser
}

// NewParser creates a new C parser
func NewParser() *CParser {
	return &CParser{
		BaseParser: parser.NewBaseParser("c", []string{".c", ".h"}, 100),
	}
}

// Parse parses C source code
func (p *CParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
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

	// Extract function declarations and definitions
	p.extractFunctions(contentStr, result)

	// Extract structs, unions, enums
	p.extractTypes(contentStr, result)

	// Extract typedefs
	p.extractTypedefs(contentStr, result)

	// Extract global variables
	p.extractGlobals(contentStr, result)

	// Extract defines
	p.extractDefines(lines, result)

	result.Metadata["language"] = "c"
	result.Metadata["type"] = "source"
	if strings.HasSuffix(filePath, ".h") {
		result.Metadata["type"] = "header"
	}

	return result, nil
}

func (p *CParser) extractIncludes(lines []string, result *types.ParseResult) {
	includeRe := regexp.MustCompile(`^\s*#\s*include\s+[<"]([^>"]+)[>"]`)

	for i, line := range lines {
		if matches := includeRe.FindStringSubmatch(line); matches != nil {
			imp := &types.Import{
				Source:     matches[1],
				LineNumber: i + 1,
			}
			result.Imports = append(result.Imports, imp)
		}
	}
}

func (p *CParser) extractFunctions(content string, result *types.ParseResult) {
	// Function definition/declaration
	// Matches: returnType functionName(params)
	funcRe := regexp.MustCompile(`(?m)^\s*(static|extern|inline)?\s*([\w\s\*]+?)\s+(\w+)\s*\(([^)]*)\)\s*(?:{|;)`)

	matches := funcRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		modifier := ""
		if match[2] != -1 {
			modifier = strings.TrimSpace(content[match[2]:match[3]])
		}

		returnType := strings.TrimSpace(content[match[4]:match[5]])
		name := content[match[6]:match[7]]
		params := ""
		if match[8] != -1 {
			params = content[match[8]:match[9]]
		}

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		// Skip if return type is too long or looks like a macro
		if len(returnType) > 100 || strings.Contains(returnType, "#") {
			continue
		}

		// Skip common false positives
		if name == "if" || name == "while" || name == "for" || name == "switch" {
			continue
		}

		sig := returnType + " " + name + "(" + params + ")"

		visibility := types.VisibilityPublic
		if modifier == "static" {
			visibility = types.VisibilityPrivate // static in C means file-local
		}

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeFunction,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: visibility,
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

func (p *CParser) extractTypes(content string, result *types.ParseResult) {
	// Struct declaration
	structRe := regexp.MustCompile(`(?m)^\s*(typedef\s+)?struct\s+(\w+)?\s*{`)

	structMatches := structRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range structMatches {
		name := ""
		if match[4] != -1 {
			name = content[match[4]:match[5]]
		} else {
			// Anonymous struct in typedef
			name = "anonymous"
		}

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeStruct,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  "struct " + name,
		}

		result.Symbols = append(result.Symbols, symbol)
	}

	// Union declaration
	unionRe := regexp.MustCompile(`(?m)^\s*(typedef\s+)?union\s+(\w+)?\s*{`)

	unionMatches := unionRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range unionMatches {
		name := ""
		if match[4] != -1 {
			name = content[match[4]:match[5]]
		} else {
			name = "anonymous"
		}

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeStruct, // Use struct type for unions too
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  "union " + name,
		}

		result.Symbols = append(result.Symbols, symbol)
	}

	// Enum declaration
	enumRe := regexp.MustCompile(`(?m)^\s*(typedef\s+)?enum\s+(\w+)?\s*{`)

	enumMatches := enumRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range enumMatches {
		name := ""
		if match[4] != -1 {
			name = content[match[4]:match[5]]
		} else {
			name = "anonymous"
		}

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeEnum,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  "enum " + name,
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *CParser) extractTypedefs(content string, result *types.ParseResult) {
	// typedef typename newname;
	typedefRe := regexp.MustCompile(`(?m)^\s*typedef\s+([^;{]+?)\s+(\w+)\s*;`)

	matches := typedefRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		baseType := strings.TrimSpace(content[match[2]:match[3]])
		name := content[match[4]:match[5]]

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeVariable, // Use variable for typedef
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  "typedef " + baseType + " " + name,
			Metadata: map[string]interface{}{
				"typedef": true,
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *CParser) extractGlobals(content string, result *types.ParseResult) {
	// Global variable declaration
	// This is tricky because it can look like a function
	globalRe := regexp.MustCompile(`(?m)^\s*(static|extern|const|volatile)?\s*([\w\s\*]+?)\s+(\w+)\s*(?:=|;)`)

	matches := globalRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		modifier := ""
		if match[2] != -1 {
			modifier = strings.TrimSpace(content[match[2]:match[3]])
		}

		varType := strings.TrimSpace(content[match[4]:match[5]])
		name := content[match[6]:match[7]]

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		// Skip if type is too long
		if len(varType) > 100 {
			continue
		}

		// Skip preprocessor directives
		if strings.HasPrefix(name, "#") {
			continue
		}

		visibility := types.VisibilityPublic
		if modifier == "static" {
			visibility = types.VisibilityPrivate
		}

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeVariable,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: visibility,
			Signature:  varType + " " + name,
		}

		if modifier != "" {
			symbol.Metadata = map[string]interface{}{
				"modifier": modifier,
			}
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *CParser) extractDefines(lines []string, result *types.ParseResult) {
	defineRe := regexp.MustCompile(`^\s*#\s*define\s+(\w+)(?:\([^)]*\))?\s*(.*)`)

	for i, line := range lines {
		if matches := defineRe.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			value := strings.TrimSpace(matches[2])

			// Skip common guards
			if strings.HasSuffix(name, "_H") || strings.HasSuffix(name, "_H_") {
				continue
			}

			sig := "#define " + name
			if value != "" && len(value) < 50 {
				sig += " " + value
			}

			symbol := &types.Symbol{
				Name:       name,
				Type:       types.SymbolTypeConstant,
				StartLine:  i + 1,
				EndLine:    i + 1,
				Visibility: types.VisibilityPublic,
				Signature:  sig,
				Metadata: map[string]interface{}{
					"define": true,
				},
			}

			result.Symbols = append(result.Symbols, symbol)
		}
	}
}
