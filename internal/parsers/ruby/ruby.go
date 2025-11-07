package ruby

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// RubyParser parses Ruby source code
type RubyParser struct {
	*parser.BaseParser
}

// NewParser creates a new Ruby parser
func NewParser() *RubyParser {
	return &RubyParser{
		BaseParser: parser.NewBaseParser("ruby", []string{".rb"}, 100),
	}
}

// Parse parses Ruby source code
func (p *RubyParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
	result := &types.ParseResult{
		Symbols:       make([]*types.Symbol, 0),
		Imports:       make([]*types.Import, 0),
		Relationships: make([]*types.Relationship, 0),
		Metadata:      make(map[string]interface{}),
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Extract requires and includes
	p.extractRequires(lines, result)

	// Extract modules
	p.extractModules(contentStr, result)

	// Extract classes
	p.extractClasses(contentStr, result)

	// Extract methods
	p.extractMethods(contentStr, result)

	result.Metadata["language"] = "ruby"

	return result, nil
}

func (p *RubyParser) extractRequires(lines []string, result *types.ParseResult) {
	requireRe := regexp.MustCompile(`^\s*require\s+['"]([^'"]+)['"]`)
	requireRelRe := regexp.MustCompile(`^\s*require_relative\s+['"]([^'"]+)['"]`)
	gemRe := regexp.MustCompile(`^\s*gem\s+['"]([^'"]+)['"]`)

	for i, line := range lines {
		// require
		if matches := requireRe.FindStringSubmatch(line); matches != nil {
			imp := &types.Import{
				Source: matches[1],
				Line:   i + 1,
			}
			result.Imports = append(result.Imports, imp)
			continue
		}

		// require_relative
		if matches := requireRelRe.FindStringSubmatch(line); matches != nil {
			imp := &types.Import{
				Source: matches[1],
				Line:   i + 1,
				Alias:  "relative",
			}
			result.Imports = append(result.Imports, imp)
			continue
		}

		// gem
		if matches := gemRe.FindStringSubmatch(line); matches != nil {
			imp := &types.Import{
				Source: matches[1],
				Line:   i + 1,
				Alias:  "gem",
			}
			result.Imports = append(result.Imports, imp)
		}
	}
}

func (p *RubyParser) extractModules(content string, result *types.ParseResult) {
	moduleRe := regexp.MustCompile(`(?m)^\s*module\s+([A-Z][\w:]*)\s*$`)

	matches := moduleRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeModule,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  "module " + name,
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *RubyParser) extractClasses(content string, result *types.ParseResult) {
	classRe := regexp.MustCompile(`(?m)^\s*class\s+([A-Z][\w:]*)\s*(?:<\s*([\w:]+))?`)

	matches := classRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]

		parent := ""
		if match[4] != -1 {
			parent = content[match[4]:match[5]]
		}

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		sig := "class " + name
		if parent != "" {
			sig += " < " + parent
		}

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeClass,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  sig,
		}

		result.Symbols = append(result.Symbols, symbol)

		// Add inheritance relationship
		if parent != "" {
			result.Relationships = append(result.Relationships, &types.Relationship{
				Type:       types.RelationshipTypeExtends,
				SourceName: name,
				TargetName: parent,
			})
		}
	}
}

func (p *RubyParser) extractMethods(content string, result *types.ParseResult) {
	// Instance method
	methodRe := regexp.MustCompile(`(?m)^\s*def\s+(self\.)?(\w+[!?]?)\s*(?:\(([^)]*)\))?`)

	matches := methodRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		isClassMethod := false
		if match[2] != -1 {
			isClassMethod = true
		}

		name := content[match[4]:match[5]]
		params := ""
		if match[6] != -1 {
			params = content[match[6]:match[7]]
		}

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		sig := "def "
		if isClassMethod {
			sig += "self."
		}
		sig += name
		if params != "" {
			sig += "(" + params + ")"
		}

		visibility := types.VisibilityPublic
		// Check if method is private or protected (would need context analysis)
		// For now, assume public

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeMethod,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: visibility,
			Signature:  sig,
		}

		if isClassMethod {
			symbol.Metadata = map[string]interface{}{
				"class_method": true,
			}
		}

		result.Symbols = append(result.Symbols, symbol)
	}

	// attr_accessor, attr_reader, attr_writer
	attrRe := regexp.MustCompile(`(?m)^\s*(attr_accessor|attr_reader|attr_writer)\s+:(\w+)`)

	attrMatches := attrRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range attrMatches {
		attrType := content[match[2]:match[3]]
		name := content[match[4]:match[5]]

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeProperty,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  attrType + " :" + name,
			Metadata: map[string]interface{}{
				"attribute": attrType,
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}
