package html

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
)

// HTMLParser parses HTML source code
type HTMLParser struct {
}

// NewParser creates a new HTML parser
func NewParser() *HTMLParser {
	return &HTMLParser{}
}

// Language returns the language identifier (e.g., "html")
func (p *HTMLParser) Language() string {
	return "html"
}

// Extensions returns file extensions this parser handles (e.g., [".html", ".htm"])
func (p *HTMLParser) Extensions() []string {
	return []string{".html", ".htm"}
}

// Priority returns parser priority (higher = preferred when multiple parsers match)
func (p *HTMLParser) Priority() int {
	return 50
}

// SupportsFramework checks if parser supports specific framework analysis
func (p *HTMLParser) SupportsFramework(framework string) bool {
	return false
}

// Parse parses HTML content
func (p *HTMLParser) Parse(content []byte, filePath string) (*parsing.ParseResult, error) {
	result := &parsing.ParseResult{
		Symbols:  make([]*model.Symbol, 0),
		Imports:  make([]*model.Import, 0),
		Metadata: make(map[string]interface{}),
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

func (p *HTMLParser) extractIDs(content string, result *parsing.ParseResult) {
	// Match id="..." or id='...'
	idRe := regexp.MustCompile(`<(\\w+)[^>]*\sid=["']([^"']+)["']`)

	matches := idRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		tag := content[match[2]:match[3]]
		id := content[match[4]:match[5]]

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &model.Symbol{
			Name:       "#" + id,
			Kind:       model.SymbolKindVariable,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  "<" + tag + " id=\"" + id + "\">",
			Metadata: map[string]string{
				"tag": tag,
				"id":  id,
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *HTMLParser) extractClasses(content string, result *parsing.ParseResult) {
	// Match class="..." or class='...'
	classRe := regexp.MustCompile(`<(\\w+)[^>]*\sclass=["']([^"']+)["']`)

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

			symbol := &model.Symbol{
				Name:       "." + className,
				Kind:       model.SymbolKindVariable,
				File:       "", // File path will be set by the caller
				Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
				Visibility: model.VisibilityPublic,
				Signature:  "<" + tag + " class=\"" + className + "\">",
				Metadata: map[string]string{
					"tag":   tag,
					"class": className,
				},
			}

			result.Symbols = append(result.Symbols, symbol)
		}
	}
}

func (p *HTMLParser) extractImports(content string, result *parsing.ParseResult) {
	// Script tags
	scriptRe := regexp.MustCompile(`<script[^>]*\ssrc=["']([^"']+)["']`)

	scriptMatches := scriptRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range scriptMatches {
		src := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		imp := &model.Import{
			Path: src,
			Range: model.Range{
				Start: model.Position{Line: lineNum},
				End:   model.Position{Line: lineNum},
			},
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
			imp := &model.Import{
				Path: href,
				Range: model.Range{
					Start: model.Position{Line: lineNum},
					End:   model.Position{Line: lineNum},
				},
			}

			result.Imports = append(result.Imports, imp)
		}
	}
}
