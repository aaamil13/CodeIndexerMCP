package powershell

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// PowerShellParser parses PowerShell source code
type PowerShellParser struct {
	*parser.BaseParser
}

// NewParser creates a new PowerShell parser
func NewParser() *PowerShellParser {
	return &PowerShellParser{
		BaseParser: parser.NewBaseParser("powershell", []string{".ps1", ".psm1", ".psd1"}, 100),
	}
}

// Parse parses PowerShell source code
func (p *PowerShellParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
	result := &types.ParseResult{
		Symbols:       make([]*types.Symbol, 0),
		Imports:       make([]*types.Import, 0),
		Relationships: make([]*types.Relationship, 0),
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

func (p *PowerShellParser) extractImports(lines []string, result *types.ParseResult) {
	importModuleRe := regexp.MustCompile(`^\s*Import-Module\s+([^\s#]+)`)
	usingRe := regexp.MustCompile(`^\s*using\s+(?:module|namespace)\s+([^\s#]+)`)

	for i, line := range lines {
		// Import-Module
		if matches := importModuleRe.FindStringSubmatch(line); matches != nil {
			imp := &types.Import{
				Source: matches[1],
				Line:   i + 1,
				Alias:  "module",
			}
			result.Imports = append(result.Imports, imp)
			continue
		}

		// using module/namespace
		if matches := usingRe.FindStringSubmatch(line); matches != nil {
			imp := &types.Import{
				Source: matches[1],
				Line:   i + 1,
			}
			result.Imports = append(result.Imports, imp)
		}
	}
}

func (p *PowerShellParser) extractFunctions(content string, result *types.ParseResult) {
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

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeFunction,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
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

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeFunction,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  "filter " + name,
			Metadata: map[string]interface{}{
				"filter": true,
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *PowerShellParser) extractClasses(content string, result *types.ParseResult) {
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

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeClass,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  sig,
		}

		result.Symbols = append(result.Symbols, symbol)

		// Add inheritance
		if inheritance != "" {
			parts := strings.Split(inheritance, ",")
			for _, parent := range parts {
				parent = strings.TrimSpace(parent)
				if parent != "" {
					result.Relationships = append(result.Relationships, &types.Relationship{
						Type:       types.RelationshipTypeExtends,
						SourceName: name,
						TargetName: parent,
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

func (p *PowerShellParser) extractVariables(lines []string, result *types.ParseResult) {
	// Script-level variable: $script:VarName or $global:VarName
	varRe := regexp.MustCompile(`^\s*\$(?:script|global):([A-Z]\w*)\s*=`)

	seen := make(map[string]bool)

	for i, line := range lines {
		if matches := varRe.FindStringSubmatch(line); matches != nil {
			name := matches[1]

			if seen[name] {
				continue
			}
			seen[name] = true

			symbol := &types.Symbol{
				Name:       "$" + name,
				Type:       types.SymbolTypeVariable,
				StartLine:  i + 1,
				EndLine:    i + 1,
				Visibility: types.VisibilityPublic,
				Signature:  "$" + name,
			}

			result.Symbols = append(result.Symbols, symbol)
		}
	}
}
