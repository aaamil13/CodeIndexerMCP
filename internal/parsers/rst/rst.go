package rst

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// ReStructuredTextParser parses reStructuredText documentation files
type ReStructuredTextParser struct {
	*parser.BaseParser
}

// NewParser creates a new reStructuredText parser
func NewParser() *ReStructuredTextParser {
	return &ReStructuredTextParser{
		BaseParser: parser.NewBaseParser("rst", []string{".rst", ".rest"}, 50),
	}
}

// Parse parses reStructuredText content
func (p *ReStructuredTextParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
	result := &types.ParseResult{
		Symbols:       make([]*types.Symbol, 0),
		Imports:       make([]*types.Import, 0),
		Relationships: make([]*types.Relationship, 0),
		Metadata:      make(map[string]interface{}),
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Extract sections (headers)
	p.extractSections(lines, result)

	// Extract directives
	p.extractDirectives(lines, result)

	// Extract references
	p.extractReferences(lines, result)

	result.Metadata["language"] = "rst"
	result.Metadata["type"] = "documentation"

	return result, nil
}

func (p *ReStructuredTextParser) extractSections(lines []string, result *types.ParseResult) {
	// RST headers: underline with =, -, ~, etc.
	// Title
	// =====
	// or
	// ======
	// Title
	// ======

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if line == "" {
			continue
		}

		// Check if next line is an underline
		if i+1 < len(lines) {
			nextLine := lines[i+1]
			if p.isUnderline(nextLine) && len(nextLine) >= len(strings.TrimSpace(line)) {
				title := strings.TrimSpace(line)
				if title != "" {
					level := p.getHeaderLevel(nextLine[0])

					symbol := &types.Symbol{
						Name:       title,
						Type:       types.SymbolTypeVariable,
						StartLine:  i + 1,
						EndLine:    i + 1,
						Visibility: types.VisibilityPublic,
						Signature:  title,
						Metadata: map[string]interface{}{
							"header":       true,
							"header_level": level,
						},
					}

					result.Symbols = append(result.Symbols, symbol)
					i++ // Skip the underline
					continue
				}
			}
		}

		// Check if this is an overline (previous line is also a line of symbols)
		if i > 0 {
			prevLine := lines[i-1]
			if i+1 < len(lines) {
				nextLine := lines[i+1]
				if p.isUnderline(prevLine) && p.isUnderline(nextLine) &&
					len(prevLine) >= len(strings.TrimSpace(line)) &&
					len(nextLine) >= len(strings.TrimSpace(line)) {

					title := strings.TrimSpace(line)
					if title != "" {
						level := p.getHeaderLevel(prevLine[0])

						symbol := &types.Symbol{
							Name:       title,
							Type:       types.SymbolTypeVariable,
							StartLine:  i + 1,
							EndLine:    i + 1,
							Visibility: types.VisibilityPublic,
							Signature:  title,
							Metadata: map[string]interface{}{
								"header":       true,
								"header_level": level,
								"overline":     true,
							},
						}

						result.Symbols = append(result.Symbols, symbol)
					}
				}
			}
		}
	}
}

func (p *ReStructuredTextParser) isUnderline(line string) bool {
	if len(line) == 0 {
		return false
	}

	// Check if line consists of a single repeated character
	firstChar := line[0]
	underlineChars := "=-~`:#\"'^_*+"

	if !strings.ContainsRune(underlineChars, rune(firstChar)) {
		return false
	}

	for _, ch := range line {
		if ch != rune(firstChar) && ch != ' ' {
			return false
		}
	}

	return true
}

func (p *ReStructuredTextParser) getHeaderLevel(char byte) int {
	// Common convention for header levels in RST
	levels := map[byte]int{
		'#': 1, // With overline
		'*': 1, // With overline
		'=': 2,
		'-': 3,
		'^': 4,
		'"': 5,
		'~': 6,
	}

	if level, ok := levels[char]; ok {
		return level
	}
	return 3 // Default
}

func (p *ReStructuredTextParser) extractDirectives(lines []string, result *types.ParseResult) {
	// Directives: .. directive:: content
	directiveRe := regexp.MustCompile(`^\.\.\s+([\w-]+)::\s*(.*)`)

	for i, line := range lines {
		if matches := directiveRe.FindStringSubmatch(line); matches != nil {
			directive := matches[1]
			content := matches[2]

			// Skip common directives that are not interesting
			if directive == "note" || directive == "warning" || directive == "code-block" {
				continue
			}

			name := directive
			if content != "" {
				name = directive + ": " + content
			}

			symbol := &types.Symbol{
				Name:       name,
				Type:       types.SymbolTypeFunction,
				StartLine:  i + 1,
				EndLine:    i + 1,
				Visibility: types.VisibilityPublic,
				Signature:  ".. " + directive + "::",
				Metadata: map[string]interface{}{
					"directive": directive,
				},
			}

			result.Symbols = append(result.Symbols, symbol)
		}
	}
}

func (p *ReStructuredTextParser) extractReferences(lines []string, result *types.ParseResult) {
	// Reference definitions: .. _label: or .. _label:
	refRe := regexp.MustCompile(`^\.\.\s+_([^:]+):\s*(.*)`)

	for i, line := range lines {
		if matches := refRe.FindStringSubmatch(line); matches != nil {
			label := matches[1]
			target := matches[2]

			sig := "_" + label
			if target != "" {
				sig += ": " + target
			}

			symbol := &types.Symbol{
				Name:       label,
				Type:       types.SymbolTypeConstant,
				StartLine:  i + 1,
				EndLine:    i + 1,
				Visibility: types.VisibilityPublic,
				Signature:  sig,
				Metadata: map[string]interface{}{
					"reference": true,
				},
			}

			result.Symbols = append(result.Symbols, symbol)
		}
	}
}
