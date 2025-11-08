package css

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
)

// CSSParser parses CSS/SCSS/SASS source code
type CSSParser struct {
}

// NewParser creates a new CSS parser
func NewParser() *CSSParser {
	return &CSSParser{}
}

// Language returns the language identifier (e.g., "css")
func (p *CSSParser) Language() string {
	return "css"
}

// Extensions returns file extensions this parser handles (e.g., [".css", ".scss"])
func (p *CSSParser) Extensions() []string {
	return []string{".css", ".scss", ".sass", ".less"}
}

// Priority returns parser priority (higher = preferred when multiple parsers match)
func (p *CSSParser) Priority() int {
	return 50
}

// SupportsFramework checks if parser supports specific framework analysis
func (p *CSSParser) SupportsFramework(framework string) bool {
	return false
}

// Parse parses CSS content
func (p *CSSParser) Parse(content []byte, filePath string) (*parsing.ParseResult, error) {
	result := &parsing.ParseResult{
		Symbols:  make([]*model.Symbol, 0),
		Imports:  make([]*model.Import, 0),
		Metadata: make(map[string]interface{}),
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

func (p *CSSParser) extractImports(content string, result *parsing.ParseResult) {
	// @import "file.css" or @import url("file.css")
	importRe := regexp.MustCompile(`@import\s+(?:url\()?["']([^"']+)["']\)?`)

	matches := importRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		source := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		imp := &model.Import{
			Path: source,
			Range: model.Range{
				Start: model.Position{Line: lineNum},
				End:   model.Position{Line: lineNum},
			},
		}

		result.Imports = append(result.Imports, imp)
	}
}

func (p *CSSParser) extractSelectors(content string, result *parsing.ParseResult) {
	// CSS rule: selector { properties }
	// Match class selectors, ID selectors, element selectors
	selectorRe := regexp.MustCompile(`(?m)^\s*([.#]?["\w-]+\s*(?:[.#:]["\w-]+\s*)*\s*(?:[>+~]\s*["\w.-]+)*)\s*{`)

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
		var symbolKind model.SymbolKind
		if strings.HasPrefix(selector, ".") {
			symbolKind = model.SymbolKindVariable // Class
		} else if strings.HasPrefix(selector, "#") {
			symbolKind = model.SymbolKindConstant // ID
		} else {
			symbolKind = model.SymbolKindVariable // Element or complex selector
		}

		symbol := &model.Symbol{
			Name:       selector,
			Kind:       symbolKind,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  selector + " { }",
			Metadata: map[string]string{
				"selector": "true",
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *CSSParser) extractVariables(content string, result *parsing.ParseResult) {
	// CSS custom properties: --variable-name
	varRe := regexp.MustCompile(`(--["\w-]+)\s*:`)

	seen := make(map[string]bool)

	matches := varRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]

		if seen[name] {
			continue
		}
		seen[name] = true

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindVariable,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  name,
			Metadata: map[string]string{
				"css_variable": "true",
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}

	// SCSS/LESS variables: $variable-name or @variable-name
	scssVarRe := regexp.MustCompile(`([$@]["\w-]+)\s*:`)

	scssMatches := scssVarRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range scssMatches {
		name := content[match[2]:match[3]]

		if seen[name] {
			continue
		}
		seen[name] = true

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindVariable,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  name,
			Metadata: map[string]string{
				"preprocessor_variable": "true",
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *CSSParser) extractKeyframes(content string, result *parsing.ParseResult) {
	// @keyframes animation-name
	keyframesRe := regexp.MustCompile(`@(?:-webkit-|-moz-|-o-)?keyframes\s+(["\w-]+)`) // Added " to regex

	matches := keyframesRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindFunction,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  "@keyframes " + name,
			Metadata: map[string]string{
				"keyframes": "true",
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *CSSParser) extractMediaQueries(content string, result *parsing.ParseResult) {
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

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindVariable,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  "@media " + query,
			Metadata: map[string]string{
				"media_query": "true",
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}
