package bash

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// BashParser parses Bash/Shell script source code
type BashParser struct {
	*parser.BaseParser
}

// NewParser creates a new Bash parser
func NewParser() *BashParser {
	return &BashParser{
		BaseParser: parser.NewBaseParser("bash", []string{".sh", ".bash"}, 100),
	}
}

// Parse parses Bash source code
func (p *BashParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
	result := &types.ParseResult{
		Symbols:       make([]*types.Symbol, 0),
		Imports:       make([]*types.Import, 0),
		Relationships: make([]*types.Relationship, 0),
		Metadata:      make(map[string]interface{}),
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

func (p *BashParser) extractSources(lines []string, result *types.ParseResult) {
	sourceRe := regexp.MustCompile(`^\s*(?:source|\.)\s+([^\s#]+)`)

	for i, line := range lines {
		if matches := sourceRe.FindStringSubmatch(line); matches != nil {
			imp := &types.Import{
				Source:     matches[1],
				LineNumber: i + 1,
			}
			result.Imports = append(result.Imports, imp)
		}
	}
}

func (p *BashParser) extractFunctions(content string, result *types.ParseResult) {
	// Function declaration: function name() { or name() {
	funcRe := regexp.MustCompile(`(?m)^\s*(?:function\s+)?(\w+)\s*\(\s*\)\s*{`)

	matches := funcRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeFunction,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  "function " + name + "()",
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *BashParser) extractVariables(lines []string, result *types.ParseResult) {
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

			symbol := &types.Symbol{
				Name:       name,
				Type:       types.SymbolTypeVariable,
				StartLine:  i + 1,
				EndLine:    i + 1,
				Visibility: types.VisibilityPublic,
				Signature:  name,
			}

			result.Symbols = append(result.Symbols, symbol)
		}
	}
}
