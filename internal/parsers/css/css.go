package css

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// CSSParser parses CSS/SCSS/SASS source code
type CSSParser struct {
	*parser.BaseParser
}

// NewParser creates a new CSS parser
func NewParser() *CSSParser {
	return &CSSParser{
		BaseParser: parser.NewBaseParser("css", []string{".css", ".scss", ".sass", ".less"}, 50),
	}
}

// Parse parses CSS content
func (p *CSSParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
	result := &types.ParseResult{
		Symbols:       make([]*types.Symbol, 0),
		Imports:       make([]*types.Import, 0),
		Relationships: make([]*types.Relationship, 0),
		Metadata:      make(map[string]interface{}),
	}

	contentStr := string(content)

	// Extract @import statements
	p.extractImports(contentStr, result)

	// Extract CSS rules (selectors)
	p.extractSelectors(contentStr, result)

	// Extract CSS variables (custom properties)
	p.extractVariables(contentStr, result)

	// Extract @keyframes
	p.extractKeyframes(contentStr, result)

	// Extract @media queries
	p.extractMediaQueries(contentStr, result)

	result.Metadata["language"] = "css"
	if strings.HasSuffix(filePath, ".scss") || strings.HasSuffix(filePath, ".sass") {
		result.Metadata["preprocessor"] = "sass"
	} else if strings.HasSuffix(filePath, ".less") {
		result.Metadata["preprocessor"] = "less"
	}

	return result, nil
}

func (p *CSSParser) extractImports(content string, result *types.ParseResult) {
	// @import "file.css" or @import url("file.css")
	importRe := regexp.MustCompile(`@import\s+(?:url\()?["']([^"']+)["']\)?`)

	matches := importRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		source := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		imp := &types.Import{
			Source: source,
			Line:   lineNum,
		}

		result.Imports = append(result.Imports, imp)
	}
}

func (p *CSSParser) extractSelectors(content string, result *types.ParseResult) {
	// CSS rule: selector { properties }
	// Match class selectors, ID selectors, element selectors
	selectorRe := regexp.MustCompile(`(?m)^\s*([.#]?[\w-]+(?:[.#:][\w-]+)*(?:\s*[>+~]\s*[\w.-]+)*)\s*{`)

	seen := make(map[string]bool)

	matches := selectorRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		selector := strings.TrimSpace(content[match[2]:match[3]])

		// Skip @-rules
		if strings.HasPrefix(selector, "@") {
			continue
		}

		// Skip if already seen
		if seen[selector] {
			continue
		}
		seen[selector] = true

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		// Determine type based on selector
		var symbolType types.SymbolType
		if strings.HasPrefix(selector, ".") {
			symbolType = types.SymbolTypeVariable // Class
		} else if strings.HasPrefix(selector, "#") {
			symbolType = types.SymbolTypeConstant // ID
		} else {
			symbolType = types.SymbolTypeVariable // Element or complex selector
		}

		symbol := &types.Symbol{
			Name:       selector,
			Type:       symbolType,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  selector + " { }",
			Metadata: map[string]interface{}{
				"selector": true,
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *CSSParser) extractVariables(content string, result *types.ParseResult) {
	// CSS custom properties: --variable-name
	varRe := regexp.MustCompile(`(--[\w-]+)\s*:`)

	seen := make(map[string]bool)

	matches := varRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]

		if seen[name] {
			continue
		}
		seen[name] = true

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeVariable,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  name,
			Metadata: map[string]interface{}{
				"css_variable": true,
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}

	// SCSS/LESS variables: $variable-name or @variable-name
	scssVarRe := regexp.MustCompile(`([$@][\w-]+)\s*:`)

	scssMatches := scssVarRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range scssMatches {
		name := content[match[2]:match[3]]

		if seen[name] {
			continue
		}
		seen[name] = true

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeVariable,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  name,
			Metadata: map[string]interface{}{
				"preprocessor_variable": true,
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *CSSParser) extractKeyframes(content string, result *types.ParseResult) {
	// @keyframes animation-name
	keyframesRe := regexp.MustCompile(`@(?:-webkit-|-moz-|-o-)?keyframes\s+([\w-]+)`)

	matches := keyframesRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeFunction,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  "@keyframes " + name,
			Metadata: map[string]interface{}{
				"keyframes": true,
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *CSSParser) extractMediaQueries(content string, result *types.ParseResult) {
	// @media query
	mediaRe := regexp.MustCompile(`@media\s+([^{]+)`)

	seen := make(map[string]bool)

	matches := mediaRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		query := strings.TrimSpace(content[match[2]:match[3]])

		if seen[query] {
			continue
		}
		seen[query] = true

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		// Create a simple name from query
		name := "@media " + query
		if len(name) > 50 {
			name = name[:50] + "..."
		}

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeVariable,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  "@media " + query,
			Metadata: map[string]interface{}{
				"media_query": true,
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}
