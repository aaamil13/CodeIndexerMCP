package html

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// HTMLParser parses HTML source code
type HTMLParser struct {
	*parser.BaseParser
}

// NewParser creates a new HTML parser
func NewParser() *HTMLParser {
	return &HTMLParser{
		BaseParser: parser.NewBaseParser("html", []string{".html", ".htm"}, 50),
	}
}

// Parse parses HTML content
func (p *HTMLParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
	result := &types.ParseResult{
		Symbols:       make([]*types.Symbol, 0),
		Imports:       make([]*types.Import, 0),
		Relationships: make([]*types.Relationship, 0),
		Metadata:      make(map[string]interface{}),
	}

	contentStr := string(content)

	// Extract elements with IDs
	p.extractIDs(contentStr, result)

	// Extract elements with classes
	p.extractClasses(contentStr, result)

	// Extract script/link tags (imports)
	p.extractImports(contentStr, result)

	result.Metadata["language"] = "html"

	return result, nil
}

func (p *HTMLParser) extractIDs(content string, result *types.ParseResult) {
	// Match id="..." or id='...'
	idRe := regexp.MustCompile(`<(\w+)[^>]*\sid=["']([^"']+)["']`)

	matches := idRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		tag := content[match[2]:match[3]]
		id := content[match[4]:match[5]]

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &types.Symbol{
			Name:       "#" + id,
			Type:       types.SymbolTypeVariable,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  "<" + tag + " id=\"" + id + "\">",
			Metadata: map[string]interface{}{
				"tag": tag,
				"id":  id,
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *HTMLParser) extractClasses(content string, result *types.ParseResult) {
	// Match class="..." or class='...'
	classRe := regexp.MustCompile(`<(\w+)[^>]*\sclass=["']([^"']+)["']`)

	seen := make(map[string]bool)

	matches := classRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		tag := content[match[2]:match[3]]
		classes := content[match[4]:match[5]]

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		// Split multiple classes
		for _, className := range strings.Fields(classes) {
			className = strings.TrimSpace(className)
			if className == "" {
				continue
			}

			key := className
			if seen[key] {
				continue
			}
			seen[key] = true

			symbol := &types.Symbol{
				Name:       "." + className,
				Type:       types.SymbolTypeVariable,
				StartLine:  lineNum,
				EndLine:    lineNum,
				Visibility: types.VisibilityPublic,
				Signature:  "<" + tag + " class=\"" + className + "\">",
				Metadata: map[string]interface{}{
					"tag":   tag,
					"class": className,
				},
			}

			result.Symbols = append(result.Symbols, symbol)
		}
	}
}

func (p *HTMLParser) extractImports(content string, result *types.ParseResult) {
	// Script tags
	scriptRe := regexp.MustCompile(`<script[^>]*\ssrc=["']([^"']+)["']`)

	scriptMatches := scriptRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range scriptMatches {
		src := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		imp := &types.Import{
			Source: src,
			Line:   lineNum,
			Alias:  "script",
		}

		result.Imports = append(result.Imports, imp)
	}

	// Link tags (CSS)
	linkRe := regexp.MustCompile(`<link[^>]*\shref=["']([^"']+)["']`)

	linkMatches := linkRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range linkMatches {
		href := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		// Only add if it's a stylesheet
		if strings.Contains(content[match[0]:match[1]], "stylesheet") {
			imp := &types.Import{
				Source: href,
				Line:   lineNum,
				Alias:  "stylesheet",
			}

			result.Imports = append(result.Imports, imp)
		}
	}
}
