package ruby

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
)

// RubyParser parses Ruby source code
type RubyParser struct {
}

// NewParser creates a new Ruby parser
func NewParser() *RubyParser {
	return &RubyParser{}
}

// Language returns the language identifier (e.g., "ruby")
func (p *RubyParser) Language() string {
	return "ruby"
}

// Extensions returns file extensions this parser handles (e.g., [".rb"])
func (p *RubyParser) Extensions() []string {
	return []string{".rb"}
}

// Priority returns parser priority (higher = preferred when multiple parsers match)
func (p *RubyParser) Priority() int {
	return 100
}

// SupportsFramework checks if parser supports specific framework analysis
func (p *RubyParser) SupportsFramework(framework string) bool {
	return false
}

// Parse parses Ruby source code
func (p *RubyParser) Parse(content []byte, filePath string) (*parsing.ParseResult, error) {
	result := &parsing.ParseResult{
		Symbols:       make([]*model.Symbol, 0),
		Imports:       make([]*model.Import, 0),
		Relationships: make([]*model.Relationship, 0),
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

func (p *RubyParser) extractRequires(lines []string, result *parsing.ParseResult) {
	requireRe := regexp.MustCompile(`^\s*require\s+['"]([^'"]+)['"]`)
	requireRelRe := regexp.MustCompile(`^\s*require_relative\s+['"]([^'"]+)['"]`)
	gemRe := regexp.MustCompile(`^\s*gem\s+['"]([^'"]+)['"]`)

	for i, line := range lines {
		// require
		if matches := requireRe.FindStringSubmatch(line); matches != nil {
			imp := &model.Import{
				Path: matches[1],
				Range: model.Range{
					Start: model.Position{Line: i + 1},
					End:   model.Position{Line: i + 1},
				},
			}
			result.Imports = append(result.Imports, imp)
			continue
		}

		// require_relative
		if matches := requireRelRe.FindStringSubmatch(line); matches != nil {
			imp := &model.Import{
				Path: matches[1],
				Range: model.Range{
					Start: model.Position{Line: i + 1},
					End:   model.Position{Line: i + 1},
				},
			}
			result.Imports = append(result.Imports, imp)
			continue
		}

		// gem
		if matches := gemRe.FindStringSubmatch(line); matches != nil {
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

func (p *RubyParser) extractModules(content string, result *parsing.ParseResult) {
	moduleRe := regexp.MustCompile(`(?m)^\s*module\s+([A-Z][\w:]*)\s*$`)

	matches := moduleRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindModule,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  "module " + name,
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *RubyParser) extractClasses(content string, result *parsing.ParseResult) {
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

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindClass,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  sig,
		}

		result.Symbols = append(result.Symbols, symbol)

		// Add inheritance relationship
		if parent != "" {
			result.Relationships = append(result.Relationships, &model.Relationship{
				Type:       model.RelationshipKindExtends,
				SourceSymbol: name,
				TargetSymbol: parent,
			})
		}
	}
}

func (p *RubyParser) extractMethods(content string, result *parsing.ParseResult) {
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

		visibility := model.VisibilityPublic
		// Check if method is private or protected (would need context analysis)
		// For now, assume public

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindMethod,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: visibility,
			Signature:  sig,
		}

		if isClassMethod {
			symbol.Metadata = map[string]string{
				"class_method": "true",
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

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindProperty,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  attrType + " :" + name,
			Metadata: map[string]string{
				"attribute": attrType,
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}