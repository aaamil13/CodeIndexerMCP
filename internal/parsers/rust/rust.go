package rust

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// RustParser parses Rust source code
type RustParser struct {
	*parser.BaseParser
}

// NewParser creates a new Rust parser
func NewParser() *RustParser {
	return &RustParser{
		BaseParser: parser.NewBaseParser("rust", []string{".rs"}, 100),
	}
}

// Parse parses Rust source code
func (p *RustParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
	result := &types.ParseResult{
		Symbols:       make([]*types.Symbol, 0),
		Imports:       make([]*types.Import, 0),
		Relationships: make([]*types.Relationship, 0),
		Metadata:      make(map[string]interface{}),
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Extract use statements
	p.extractUses(lines, result)

	// Extract mods
	p.extractMods(lines, result)

	// Extract structs and enums
	p.extractTypes(contentStr, result)

	// Extract traits
	p.extractTraits(contentStr, result)

	// Extract impl blocks
	p.extractImpls(contentStr, result)

	// Extract functions
	p.extractFunctions(contentStr, result)

	result.Metadata["language"] = "rust"

	return result, nil
}

func (p *RustParser) extractUses(lines []string, result *types.ParseResult) {
	useRe := regexp.MustCompile(`^\s*use\s+([\w:{}*,\s]+);`)

	for i, line := range lines {
		if matches := useRe.FindStringSubmatch(line); matches != nil {
			imp := &types.Import{
				Source:     matches[1],
				LineNumber: i + 1,
			}
			result.Imports = append(result.Imports, imp)
		}
	}
}

func (p *RustParser) extractMods(lines []string, result *types.ParseResult) {
	modRe := regexp.MustCompile(`^\s*(pub\s+)?mod\s+(\w+)`)

	for i, line := range lines {
		if matches := modRe.FindStringSubmatch(line); matches != nil {
			isPublic := matches[1] != ""
			name := matches[2]

			visibility := types.VisibilityPrivate
			if isPublic {
				visibility = types.VisibilityPublic
			}

			symbol := &types.Symbol{
				Name:       name,
				Type:       types.SymbolTypeModule,
				StartLine:  i + 1,
				EndLine:    i + 1,
				Visibility: visibility,
				Signature:  "mod " + name,
			}

			result.Symbols = append(result.Symbols, symbol)
		}
	}
}

func (p *RustParser) extractTypes(content string, result *types.ParseResult) {
	// Struct declaration
	structRe := regexp.MustCompile(`(?m)^\s*(pub\s+)?struct\s+(\w+)(?:<[^>]+>)?`)

	structMatches := structRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range structMatches {
		isPublic := match[2] != -1
		name := content[match[4]:match[5]]

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		visibility := types.VisibilityPrivate
		if isPublic {
			visibility = types.VisibilityPublic
		}

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeStruct,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: visibility,
			Signature:  "struct " + name,
		}

		result.Symbols = append(result.Symbols, symbol)
	}

	// Enum declaration
	enumRe := regexp.MustCompile(`(?m)^\s*(pub\s+)?enum\s+(\w+)(?:<[^>]+>)?`)

	enumMatches := enumRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range enumMatches {
		isPublic := match[2] != -1
		name := content[match[4]:match[5]]

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		visibility := types.VisibilityPrivate
		if isPublic {
			visibility = types.VisibilityPublic
		}

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeEnum,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: visibility,
			Signature:  "enum " + name,
		}

		result.Symbols = append(result.Symbols, symbol)
	}

	// Type alias
	typeRe := regexp.MustCompile(`(?m)^\s*(pub\s+)?type\s+(\w+)\s*=`)

	typeMatches := typeRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range typeMatches {
		isPublic := match[2] != -1
		name := content[match[4]:match[5]]

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		visibility := types.VisibilityPrivate
		if isPublic {
			visibility = types.VisibilityPublic
		}

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeVariable,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: visibility,
			Signature:  "type " + name,
			Metadata: map[string]interface{}{
				"type_alias": true,
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *RustParser) extractTraits(content string, result *types.ParseResult) {
	traitRe := regexp.MustCompile(`(?m)^\s*(pub\s+)?trait\s+(\w+)(?:<[^>]+>)?`)

	matches := traitRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		isPublic := match[2] != -1
		name := content[match[4]:match[5]]

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		visibility := types.VisibilityPrivate
		if isPublic {
			visibility = types.VisibilityPublic
		}

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeInterface, // Traits are like interfaces
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: visibility,
			Signature:  "trait " + name,
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *RustParser) extractImpls(content string, result *types.ParseResult) {
	// impl Trait for Type
	implRe := regexp.MustCompile(`(?m)^\s*impl(?:<[^>]+>)?\s+(\w+)\s+for\s+(\w+)`)

	matches := implRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		traitName := content[match[2]:match[3]]
		typeName := content[match[4]:match[5]]

		// Add relationship
		result.Relationships = append(result.Relationships, &types.Relationship{
			Type:       types.RelationshipImplements,
			SourceName: typeName,
			TargetName: traitName,
		})
	}
}

func (p *RustParser) extractFunctions(content string, result *types.ParseResult) {
	// Function declaration
	funcRe := regexp.MustCompile(`(?m)^\s*(pub\s+)?(?:async\s+)?(?:unsafe\s+)?(?:extern\s+"[^"]+"\s+)?fn\s+(\w+)(?:<[^>]+>)?\s*\(([^)]*)\)(?:\s*->\s*([\w<>&]+))?`)

	matches := funcRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		isPublic := match[2] != -1
		name := content[match[4]:match[5]]
		params := ""
		if match[6] != -1 {
			params = content[match[6]:match[7]]
		}

		returnType := ""
		if match[8] != -1 {
			returnType = content[match[8]:match[9]]
		}

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		sig := "fn " + name + "(" + params + ")"
		if returnType != "" {
			sig += " -> " + returnType
		}

		visibility := types.VisibilityPrivate
		if isPublic {
			visibility = types.VisibilityPublic
		}

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeFunction,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: visibility,
			Signature:  sig,
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}
