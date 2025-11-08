package powershell

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
)

// PowerShellParser parses PowerShell source code
type PowerShellParser struct {
}

// NewParser creates a new PowerShell parser
func NewParser() *PowerShellParser {
	return &PowerShellParser{}
}

// Language returns the language identifier (e.g., "powershell")
func (p *PowerShellParser) Language() string {
	return "powershell"
}

// Extensions returns file extensions this parser handles (e.g., [".ps1", ".psm1"])
func (p *PowerShellParser) Extensions() []string {
	return []string{".ps1", ".psm1", ".psd1"}
}

// Priority returns parser priority (higher = preferred when multiple parsers match)
func (p *PowerShellParser) Priority() int {
	return 100
}

// SupportsFramework checks if parser supports specific framework analysis
func (p *PowerShellParser) SupportsFramework(framework string) bool {
	return false
}

// Parse parses PowerShell source code
func (p *PowerShellParser) Parse(content []byte, filePath string) (*parsing.ParseResult, error) {
	result := &parsing.ParseResult{
		Symbols:       make([]*model.Symbol, 0),
		Imports:       make([]*model.Import, 0),
		Relationships: make([]*model.Relationship, 0),
		Metadata:      make(map[string]interface{}),
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Extract imports (Import-Module, using)
	p.extractImports(lines, result)

	// Extract functions
	p.extractFunctions(contentStr, result)

	// Extract classes (PowerShell 5.0+)
	p.extractClasses(contentStr, result)

	// Extract variables
	p.extractVariables(lines, result)

	result.Metadata["language"] = "powershell"

	return result, nil
}

func (p *PowerShellParser) extractImports(lines []string, result *parsing.ParseResult) {
	importModuleRe := regexp.MustCompile(`^\s*Import-Module\s+([^\s#]+)`)
	usingRe := regexp.MustCompile(`^\s*using\s+(?:module|namespace)\s+([^\s#]+)`)

	for i, line := range lines {
		// Import-Module
		if matches := importModuleRe.FindStringSubmatch(line); matches != nil {
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

		// using module/namespace
		if matches := usingRe.FindStringSubmatch(line); matches != nil {
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

func (p *PowerShellParser) extractFunctions(content string, result *parsing.ParseResult) {
	// Function declaration
	funcRe := regexp.MustCompile(`(?im)^\s*function\s+([\w-]+)\s*(?:\(([^)]*)\))?\s*{`)

	matches := funcRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]
		params := ""
		if match[4] != -1 {
			params = content[match[4]:match[5]]
		}

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		sig := "function " + name
		if params != "" {
			sig += "(" + params + ")"
		}

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindFunction,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  sig,
		}

		result.Symbols = append(result.Symbols, symbol)
	}

	// Filter (advanced function with [CmdletBinding])
	filterRe := regexp.MustCompile(`(?im)^\s*filter\s+([\w-]+)\s*{`)

	filterMatches := filterRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range filterMatches {
		name := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindFunction,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  "filter " + name,
			Metadata: map[string]string{
				"filter": "true",
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *PowerShellParser) extractClasses(content string, result *parsing.ParseResult) {
	// Class declaration (PowerShell 5.0+)
	classRe := regexp.MustCompile(`(?im)^\s*class\s+(\w+)(?:\s*:\s*([\w,\s]+))?\s*{`)

	matches := classRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]

			inheritance := ""
		if match[4] != -1 {
			inheritance = content[match[4]:match[5]]
		}

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		sig := "class " + name
		if inheritance != "" {
			sig += " : " + inheritance
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

		// Add inheritance
		if inheritance != "" {
			parts := strings.Split(inheritance, ",")
			for _, parent := range parts {
				parent = strings.TrimSpace(parent)
				if parent != "" {
					result.Relationships = append(result.Relationships, &model.Relationship{
						Type:       model.RelationshipKindExtends,
						SourceSymbol: name,
						TargetSymbol: parent,
					})
				}
			}
		}
	}

	// Enum declaration
	enumRe := regexp.MustCompile(`(?im)^\s*enum\s+(\w+)\s*{`)

	enumMatches := enumRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range enumMatches {
		name := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindEnum,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  "enum " + name,
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *PowerShellParser) extractVariables(lines []string, result *parsing.ParseResult) {
	// Script-level variable: $script:VarName or $global:VarName
	varRe := regexp.MustCompile(`^\s*\$(?:script|global):([A-Z]\w*)\s*=\s*`)

	seen := make(map[string]bool)

	for i, line := range lines {
		if matches := varRe.FindStringSubmatch(line); matches != nil {
			name := matches[1]

			if seen[name] {
				continue
			}
			seen[name] = true

			symbol := &model.Symbol{
				Name:       "$" + name,
				Kind:       model.SymbolKindVariable,
				File:       "", // File path will be set by the caller
				Range:      model.Range{Start: model.Position{Line: i + 1}, End: model.Position{Line: i + 1}},
				Visibility: model.VisibilityPublic,
				Signature:  "$" + name,
			}

			result.Symbols = append(result.Symbols, symbol)
		}
	}
}