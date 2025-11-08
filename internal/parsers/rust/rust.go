package rust

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
)

// RustParser parses Rust source code
type RustParser struct {
}

// NewParser creates a new Rust parser
func NewParser() *RustParser {
	return &RustParser{}
}

// Language returns the language identifier (e.g., "rust")
func (p *RustParser) Language() string {
	return "rust"
}

// Extensions returns file extensions this parser handles (e.g., [".rs"])
func (p *RustParser) Extensions() []string {
	return []string{".rs"}
}

// Priority returns parser priority (higher = preferred when multiple parsers match)
func (p *RustParser) Priority() int {
	return 100
}

// SupportsFramework checks if parser supports specific framework analysis
func (p *RustParser) SupportsFramework(framework string) bool {
	return false
}

// Parse parses Rust source code
func (p *RustParser) Parse(content []byte, filePath string) (*parsing.ParseResult, error) {
	result := &parsing.ParseResult{
		Symbols:       make([]*model.Symbol, 0),
		Imports:       make([]*model.Import, 0),
		Relationships: make([]*model.Relationship, 0),
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

func (p *RustParser) extractUses(lines []string, result *parsing.ParseResult) {
	useRe := regexp.MustCompile(`^\s*use\s+([\w:{}*,\s]+);`)

	for i, line := range lines {
		if matches := useRe.FindStringSubmatch(line); matches != nil {
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

func (p *RustParser) extractMods(lines []string, result *parsing.ParseResult) {
	modRe := regexp.MustCompile(`^\s*(pub\s+)?mod\s+(\w+)`)

	for i, line := range lines {
		if matches := modRe.FindStringSubmatch(line); matches != nil {
			isPublic := matches[1] != ""
			name := matches[2]

			visibility := model.VisibilityPrivate
			if isPublic {
				visibility = model.VisibilityPublic
			}

			symbol := &model.Symbol{
				Name:       name,
				Kind:       model.SymbolKindModule,
				File:       "", // File path will be set by the caller
				Range:      model.Range{Start: model.Position{Line: i + 1}, End: model.Position{Line: i + 1}},
				Visibility: visibility,
				Signature:  "mod " + name,
			}

			result.Symbols = append(result.Symbols, symbol)
		}
	}
}

func (p *RustParser) extractTypes(content string, result *parsing.ParseResult) {
	// Struct declaration
	structRe := regexp.MustCompile(`(?m)^\s*(pub\s+)?struct\s+(\w+)(?:<[^>]+>)?`)

	structMatches := structRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range structMatches {
		isPublic := match[2] != -1
		name := content[match[4]:match[5]]

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		visibility := model.VisibilityPrivate
		if isPublic {
			visibility = model.VisibilityPublic
		}

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindStruct,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
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

		visibility := model.VisibilityPrivate
		if isPublic {
			visibility = model.VisibilityPublic
		}

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindEnum,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
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

		visibility := model.VisibilityPrivate
		if isPublic {
			visibility = model.VisibilityPublic
		}

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindVariable,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: visibility,
			Signature:  "type " + name,
			Metadata: map[string]string{
				"type_alias": "true",
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *RustParser) extractTraits(content string, result *parsing.ParseResult) {
	traitRe := regexp.MustCompile(`(?m)^\s*(pub\s+)?trait\s+(\w+)(?:<[^>]+>)?`)

	matches := traitRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		isPublic := match[2] != -1
		name := content[match[4]:match[5]]

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		visibility := model.VisibilityPrivate
		if isPublic {
			visibility = model.VisibilityPublic
		}

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindInterface, // Traits are like interfaces
			File:       "",                        // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: visibility,
			Signature:  "trait " + name,
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *RustParser) extractImpls(content string, result *parsing.ParseResult) {
	// impl Trait for Type
	implRe := regexp.MustCompile(`(?m)^\s*impl(?:<[^>]+>)?\s+(\w+)\s+for\s+(\w+)`)

	matches := implRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		traitName := content[match[2]:match[3]]
		typeName := content[match[4]:match[5]]

		// Add relationship
		result.Relationships = append(result.Relationships, &model.Relationship{
			Type:       model.RelationshipKindImplements,
			SourceSymbol: typeName,
			TargetSymbol: traitName,
		})
	}
}

func (p *RustParser) extractFunctions(content string, result *parsing.ParseResult) {
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

		visibility := model.VisibilityPrivate
		if isPublic {
			visibility = model.VisibilityPublic
		}

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindFunction,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: visibility,
			Signature:  sig,
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}