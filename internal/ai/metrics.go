package ai

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/internal/model"
)

// MetricsCalculator calculates code quality metrics
type MetricsCalculator struct {
	db *database.Manager
}

// NewMetricsCalculator creates a new metrics calculator
func NewMetricsCalculator(db *database.Manager) *MetricsCalculator {
	return &MetricsCalculator{db: db}
}

// CalculateMetrics calculates metrics for a symbol
func (mc *MetricsCalculator) CalculateMetrics(symbol *model.Symbol) (*model.CodeMetrics, error) {
	// TODO: Implement after DB methods are available
	// // Get the symbol
	// if symbol == nil {
	// 	return nil, fmt.Errorf("symbol cannot be nil")
	// }

	// // Get file
	// file := symbol.File

	// // Extract code
	// code, err := mc.extractCode(file, symbol.Range.Start.Line, symbol.Range.End.Line)
	// if err != nil {
	// 	return nil, err
	// }

	// // Calculate metrics
	// loc := symbol.Range.End.Line - symbol.Range.Start.Line + 1
	// cyclomaticComplexity := mc.calculateCyclomaticComplexity(code, symbol.Language)
	// cognitiveComplexity := mc.calculateCognitiveComplexity(code, symbol.Language)
	// maxNestingDepth := mc.calculateMaxNestingDepth(code, symbol.Language)
	// parameters := mc.countParameters(symbol.Signature)
	// returnStatements := mc.countReturnStatements(code)
	// commentDensity := mc.calculateCommentDensity(code, symbol.Language)
	// hasDocumentation := symbol.Documentation != ""

	// // Calculate maintainability index
	// // Formula: 171 - 5.2 * ln(HV) - 0.23 * CC - 16.2 * ln(LOC)
	// // Simplified version
	// maintainability := 100.0
	// if loc > 0 {
	// 	maintainability = 171.0 - 5.2*float64(cyclomaticComplexity) - 0.23*float64(cyclomaticComplexity) - 16.2*float64(loc)/10.0
	// 	if maintainability < 0 {
	// 		maintainability = 0
	// 	}
	// 	if maintainability > 100 {
	// 		maintainability = 100
	// 	}
	// }

	// // Determine quality
	// quality := mc.determineQuality(cyclomaticComplexity, cognitiveComplexity, maintainability, hasDocumentation)

	// return &model.CodeMetrics{
	// 	FilePath:             file,
	// 	FunctionName:         symbol.Name,
	// 	LinesOfCode:          loc,
	// 	CyclomaticComplexity: cyclomaticComplexity,
	// 	CognitiveComplexity:  cognitiveComplexity,
	// 	MaintainabilityIndex: maintainability,
	// 	Parameters:           parameters,
	// 	ReturnStatements:     returnStatements,
	// 	MaxNestingDepth:      maxNestingDepth,
	// 	CommentDensity:       commentDensity,
	// 	HasDocumentation:     hasDocumentation,
	// 	Quality:              quality,
	// }, nil
	return nil, fmt.Errorf("not implemented")
}

// calculateCyclomaticComplexity calculates cyclomatic complexity
func (mc *MetricsCalculator) calculateCyclomaticComplexity(code, language string) int {
	complexity := 1 // Base complexity

	// Language-specific patterns
	var patterns []string

	switch language {
	case "go":
		patterns = []string{
			`\bif\b`, `\bfor\b`, `\bcase\b`, `\bswitch\b`,
			`&&`, `\|\|`, `\bselect\b`,
		}
	case "python":
		patterns = []string{
			`\bif\b`, `\belif\b`, `\bfor\b`, `\bwhile\b`,
			`\band\b`, `\bor\b`, `\bexcept\b`, `\bwith\b`,
		}
	case "javascript", "typescript":
		patterns = []string{
			`\bif\b`, `\bfor\b`, `\bwhile\b`, `\bcase\b`, `\bswitch\b`,
			`&&`, `\|\|`, `\bcatch\b`, `\?.*:`,
		}
	default:
		patterns = []string{
			`\bif\b`, `\bfor\b`, `\bwhile\b`, `\bcase\b`,
			`&&`, `\|\|`, `\bcatch\b`,
		}
	}

	// Count decision points
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(code, -1)
		complexity += len(matches)
	}

	return complexity
}

// calculateCognitiveComplexity calculates cognitive complexity
func (mc *MetricsCalculator) calculateCognitiveComplexity(code, language string) int {
	// Cognitive complexity adds nesting weight
	complexity := 0
	nestingLevel := 0

	lines := strings.Split(code, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for nesting increase
		if strings.Contains(trimmed, "{") {
			nestingLevel++
		}

		// Count decision points with nesting weight
		patterns := []string{`\bif\b`, `\bfor\b`, `\bwhile\b`, `\bswitch\b`}
		for _, pattern := range patterns {
			if matched, _ := regexp.MatchString(pattern, trimmed); matched {
				complexity += 1 + nestingLevel
			}
		}

		// Check for nesting decrease
		if strings.Contains(trimmed, "}") {
			nestingLevel--
			if nestingLevel < 0 {
				nestingLevel = 0
			}
		}
	}

	return complexity
}

// calculateMaxNestingDepth calculates maximum nesting depth
func (mc *MetricsCalculator) calculateMaxNestingDepth(code, language string) int {
	maxDepth := 0
	currentDepth := 0

	lines := strings.Split(code, "\n")
	for _, line := range lines {
		openBraces := strings.Count(line, "{")
		closeBraces := strings.Count(line, "}")

		currentDepth += openBraces
		if currentDepth > maxDepth {
			maxDepth = currentDepth
		}
		currentDepth -= closeBraces
	}

	return maxDepth
}

// countParameters counts function parameters
func (mc *MetricsCalculator) countParameters(signature string) int {
	if signature == "" {
		return 0
	}

	// Extract parameters from signature
	start := strings.Index(signature, "(")
	end := strings.LastIndex(signature, ")")
	if start == -1 || end == -1 || start >= end {
		return 0
	}

	params := signature[start+1 : end]
	params = strings.TrimSpace(params)

	if params == "" {
		return 0
	}

	// Simple count by commas (not perfect but good enough)
	return strings.Count(params, ",") + 1
}

// countReturnStatements counts return statements
func (mc *MetricsCalculator) countReturnStatements(code string) int {
	re := regexp.MustCompile(`\breturn\b`)
	matches := re.FindAllString(code, -1)
	return len(matches)
}

// calculateCommentDensity calculates comment density
func (mc *MetricsCalculator) calculateCommentDensity(code, language string) float64 {
	totalLines := 0
	commentLines := 0

	lines := strings.Split(code, "\n")
	for _, line := range lines {
		totalLines++
		trimmed := strings.TrimSpace(line)

		// Check for comments based on language
		isComment := false
		switch language {
		case "go", "javascript", "typescript":
			isComment = strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*")
		case "python":
			isComment = strings.HasPrefix(trimmed, "#")
		}

		if isComment {
			commentLines++
		}
	}

	if totalLines == 0 {
		return 0.0
	}

	return float64(commentLines) / float64(totalLines) * 100.0
}

// determineQuality determines overall code quality
func (mc *MetricsCalculator) determineQuality(cyclomatic, cognitive int, maintainability float64, hasDoc bool) string {
	score := 0

	// Low complexity is good
	if cyclomatic <= 10 {
		score += 2
	} else if cyclomatic <= 20 {
		score += 1
	}

	if cognitive <= 15 {
		score += 2
	} else if cognitive <= 30 {
		score += 1
	}

	// High maintainability is good
	if maintainability >= 80 {
		score += 2
	} else if maintainability >= 60 {
		score += 1
	}

	// Documentation is good
	if hasDoc {
		score += 1
	}

	// Determine quality level
	if score >= 6 {
		return "excellent"
	} else if score >= 4 {
		return "good"
	} else if score >= 2 {
		return "fair"
	}
	return "poor"
}

// extractCode extracts code from a file
func (mc *MetricsCalculator) extractCode(filePath string, startLine, endLine int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		if lineNum >= startLine && lineNum <= endLine {
			lines = append(lines, scanner.Text())
		}
		if lineNum > endLine {
			break
		}
	}

	return strings.Join(lines, "\n"), scanner.Err()
}

// CalculateFileMetrics calculates metrics for an entire file
func (mc *MetricsCalculator) CalculateFileMetrics(filePath string) ([]*model.CodeMetrics, error) {
	// TODO: Implement after DB methods are available
	// // Get file
	// file := filePath

	// // Get all symbols in file
	// symbols, err := mc.db.GetSymbolsByFile(file)
	// if err != nil {
	// 	return nil, err
	// }

	// metrics := []*model.CodeMetrics{}
	// for _, symbol := range symbols {
	// 	// TODO: Check symbol kind for function/method
	// 	// if symbol.Type == types.SymbolTypeFunction || symbol.Type == types.SymbolTypeMethod {
	// 	metric, err := mc.CalculateMetrics(symbol)
	// 	if err != nil {
	// 		continue // Skip on error
	// 	}
	// 	metrics = append(metrics, metric)
	// 	// }
	// }

	// return metrics, nil
	return nil, fmt.Errorf("not implemented")
}