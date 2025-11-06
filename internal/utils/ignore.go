package utils

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// DefaultIgnorePatterns are patterns to ignore by default
var DefaultIgnorePatterns = []string{
	".git",
	".projectIndex",
	"node_modules",
	"dist",
	"build",
	"target",
	"bin",
	"obj",
	".idea",
	".vscode",
	"*.log",
	"*.tmp",
	".DS_Store",
	"__pycache__",
	"*.pyc",
	".pytest_cache",
	"coverage",
	".nyc_output",
}

// IgnoreMatcher checks if paths should be ignored
type IgnoreMatcher struct {
	patterns []string
}

// NewIgnoreMatcher creates a new ignore matcher
func NewIgnoreMatcher(projectPath string) (*IgnoreMatcher, error) {
	patterns := make([]string, len(DefaultIgnorePatterns))
	copy(patterns, DefaultIgnorePatterns)

	// Try to load .gitignore
	gitignorePath := filepath.Join(projectPath, ".gitignore")
	if _, err := os.Stat(gitignorePath); err == nil {
		gitPatterns, err := loadGitignore(gitignorePath)
		if err == nil {
			patterns = append(patterns, gitPatterns...)
		}
	}

	// Try to load .indexerignore
	indexerIgnorePath := filepath.Join(projectPath, ".indexerignore")
	if _, err := os.Stat(indexerIgnorePath); err == nil {
		indexerPatterns, err := loadGitignore(indexerIgnorePath)
		if err == nil {
			patterns = append(patterns, indexerPatterns...)
		}
	}

	return &IgnoreMatcher{
		patterns: patterns,
	}, nil
}

// ShouldIgnore checks if a path should be ignored
func (im *IgnoreMatcher) ShouldIgnore(path string) bool {
	base := filepath.Base(path)

	for _, pattern := range im.patterns {
		if matchPattern(pattern, path) || matchPattern(pattern, base) {
			return true
		}
	}

	return false
}

// loadGitignore loads patterns from a gitignore-style file
func loadGitignore(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		patterns = append(patterns, line)
	}

	return patterns, scanner.Err()
}

// matchPattern performs simple pattern matching
func matchPattern(pattern, path string) bool {
	// Simple implementation - can be enhanced with proper glob matching
	if strings.Contains(pattern, "*") {
		// Handle wildcards
		if strings.HasPrefix(pattern, "*.") {
			ext := pattern[1:]
			return strings.HasSuffix(path, ext)
		}
		if strings.HasSuffix(pattern, "*") {
			prefix := pattern[:len(pattern)-1]
			return strings.HasPrefix(path, prefix) || strings.Contains(path, "/"+prefix)
		}
	}

	// Exact match or contains match
	return path == pattern || filepath.Base(path) == pattern || strings.Contains(path, "/"+pattern+"/") || strings.HasSuffix(path, "/"+pattern)
}
