package bash

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
)

// BashParser parses Bash/Shell script source code
type BashParser struct {
}

// NewParser creates a new Bash parser
func NewParser() *BashParser {
	return &BashParser{}
}

// Language returns the language identifier (e.g., "go", "python", "typescript")
func (p *BashParser) Language() string {
	return "bash"
}

// Extensions returns file extensions this parser handles (e.g., [".sh", ".bash"])
func (p *BashParser) Extensions() []string {
	return []string{".sh", ".bash"}
}

// Priority returns parser priority (higher = preferred when multiple parsers match)
func (p *BashParser) Priority() int {
	return 100
}

// SupportsFramework checks if parser supports specific framework analysis
func (p *BashParser) SupportsFramework(framework string) bool {
	return false
}

// Parse parses Bash source code
func (p *BashParser) Parse(content []byte, filePath string) (*parsing.ParseResult, error) {
	result := &parsing.ParseResult{
		Symbols: make([]*model.Symbol, 0),
		Imports: make([]*model.Import, 0),
		// Relationships: make([]*model.Relationship, 0), // Relationships are not directly extracted by bash parser
		Metadata: make(map[string]interface{}),
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Extract source/. commands
	p.extractSources(lines, result)

	// Extract functions
	p.extractFunctions(contentStr, result)

	// Extract variables
	p.extractVariables(lines, result)

	result.Metadata["language"] = "bash"

	return result, nil
}

func (p *BashParser) extractSources(lines []string, result *parsing.ParseResult) {
	sourceRe := regexp.MustCompile(`^\s*(?:source|\.)\s+([^\s#]+)`)

	for i, line := range lines {
		if matches := sourceRe.FindStringSubmatch(line); matches != nil {
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

func (p *BashParser) extractFunctions(content string, result *parsing.ParseResult) {
	// Function declaration: function name() { or name() {
	funcRe := regexp.MustCompile(`(?m)^\s*(?:function\s+)?(\w+)\s*\(\s*\)\s*{`)

	matches := funcRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindFunction,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  "function " + name + "()",
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *BashParser) extractVariables(lines []string, result *parsing.ParseResult) {
	// Variable assignment: VAR=value or declare VAR=value
	varRe := regexp.MustCompile(`^\s*(?:declare\s+(?:-[a-zA-Z]+\s+)?|export\s+|local\s+)?([A-Z_][A-Z0-9_]*)\s*=`)

	seen := make(map[string]bool)

	for i, line := range lines {
		if matches := varRe.FindStringSubmatch(line); matches != nil {
			name := matches[1]

			// Skip if already seen
			if seen[name] {
				continue
			}
			seen[name] = true

			// Skip common shell variables
			if name == "PATH" || name == "HOME" || name == "USER" {
				continue
			}

			symbol := &model.Symbol{
				Name:       name,
				Kind:       model.SymbolKindVariable,
				File:       "", // File path will be set by the caller
				Range:      model.Range{Start: model.Position{Line: i + 1}, End: model.Position{Line: i + 1}},
				Visibility: model.VisibilityPublic,
				Signature:  name,
			}

			result.Symbols = append(result.Symbols, symbol)
		}
	}
}
