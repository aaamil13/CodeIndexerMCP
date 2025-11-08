package react

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
)

// ReactAnalyzer analyzes React framework patterns
type ReactAnalyzer struct{}

// NewReactAnalyzer creates a new React analyzer
func NewReactAnalyzer() *ReactAnalyzer {
	return &ReactAnalyzer{}
}

// Framework returns the framework name
func (a *ReactAnalyzer) Framework() string {
	return "react"
}

// Language returns the target language
func (a *ReactAnalyzer) Language() string {
	return "typescript" // Also works for JavaScript
}

// DetectFramework detects if file uses React
func (a *ReactAnalyzer) DetectFramework(content []byte, filePath string) bool {
	contentStr := string(content)

	// Check for React imports
	if strings.Contains(contentStr, "from 'react'") ||
		strings.Contains(contentStr, "from \"react\"") ||
		strings.Contains(contentStr, "require('react')") ||
		strings.Contains(contentStr, "require(\"react\")") {
		return true
	}

	// Check for JSX syntax
	if (strings.HasSuffix(filePath, ".jsx") || strings.HasSuffix(filePath, ".tsx")) &&
		(strings.Contains(contentStr, "<") && strings.Contains(contentStr, "/>")) {
		return true
	}

	// Check for common React patterns
	reactPatterns := []string{
		"useState",
		"useEffect",
		"useContext",
		"React.Component",
		"React.FC",
		"React.Fragment",
	}

	for _, pattern := range reactPatterns {
		if strings.Contains(contentStr, pattern) {
			return true
		}
	}

	return false
}

// Analyze analyzes React-specific patterns
func (a *ReactAnalyzer) Analyze(result *model.ParseResult, content []byte) (*model.FrameworkInfo, error) {
	info := &model.FrameworkInfo{
		Name:         "react",
		Type:         "frontend",
		Components:   make([]*model.FrameworkComponent, 0),
		Routes:       make([]*model.Route, 0),
		Dependencies: make([]string, 0),
		Patterns:     make([]string, 0),
		Warnings:     make([]string, 0),
	}

	contentStr := string(content)

	// Detect React version (from imports)
	a.detectVersion(contentStr, info)

	// Extract components
	a.extractComponents(result, contentStr, info)

	// Detect hooks usage
	a.detectHooks(contentStr, info)

	// Detect React Router routes
	a.detectRoutes(contentStr, info)

	// Detect patterns and best practices
	a.detectPatterns(contentStr, info)

	// Check for common issues
	a.checkIssues(contentStr, info)

	return info, nil
}

func (a *ReactAnalyzer) detectVersion(content string, info *model.FrameworkInfo) {
	// Simple version detection from imports
	if strings.Contains(content, "import React") {
		info.Version = "16.8+" // Assumes hooks are available
	}
	if strings.Contains(content, "React.FC") || strings.Contains(content, "FunctionComponent") {
		info.Version = "16.8+" // Function components with TypeScript
	}
}

func (a *ReactAnalyzer) extractComponents(result *model.ParseResult, content string, info *model.FrameworkInfo) {
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Function components: function ComponentName() or const ComponentName = () =>
		if a.isFunctionComponent(line) {
			component := a.parseFunctionComponent(line, i+1, content)
			if component != nil {
				info.Components = append(info.Components, component)
			}
		}

		// Class components: class ComponentName extends React.Component
		if a.isClassComponent(line) {
			component := a.parseClassComponent(line, i+1, result)
			if component != nil {
				info.Components = append(info.Components, component)
			}
		}
	}
}

func (a *ReactAnalyzer) isFunctionComponent(line string) bool {
	// Must start with uppercase (React convention)
	if strings.HasPrefix(line, "function ") {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			name := parts[1]
			if len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z' {
				return true
			}
		}
	}

	if strings.Contains(line, "const ") && strings.Contains(line, "=") {
		parts := strings.Split(line, "=")
		if len(parts) >= 1 {
			namePart := strings.TrimSpace(parts[0])
			namePart = strings.TrimPrefix(namePart, "const ")
			namePart = strings.TrimPrefix(namePart, "export ")
			name := strings.TrimSpace(namePart)
			if len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z' {
				return strings.Contains(line, "=>")
			}
		}
	}

	return false
}

func (a *ReactAnalyzer) isClassComponent(line string) bool {
	return strings.Contains(line, "class ") &&
		(strings.Contains(line, "extends React.Component") ||
			strings.Contains(line, "extends Component") ||
			strings.Contains(line, "extends PureComponent"))
}

func (a *ReactAnalyzer) parseFunctionComponent(line string, lineNum int, content string) *model.FrameworkComponent {
	// Extract component name
	var name string
	if strings.HasPrefix(line, "function ") {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			name = strings.Split(parts[1], "(")[0]
		}
	} else if strings.Contains(line, "const ") {
		parts := strings.Split(line, "=")
		namePart := strings.TrimPrefix(parts[0], "const ")
		namePart = strings.TrimPrefix(namePart, "export ")
		name = strings.TrimSpace(namePart)
		name = strings.Split(name, ":")[0] // Remove type annotation
	}

	if name == "" {
		return nil
	}

	component := &model.FrameworkComponent{
		Type:       "function_component",
		Name:       name,
		Props:      make([]*model.ComponentProp, 0),
		Events:     make([]string, 0),
		Lifecycle:  make([]string, 0),
		Decorators: make([]string, 0),
		Metadata:   make(map[string]interface{}),
	}

	// Extract props (look for interface or type definition)
	propsType := a.findPropsType(name, content)
	if propsType != "" {
		component.Metadata["props_type"] = propsType
		// Could parse props interface here
	}

	// Detect hooks used
	hooks := a.findHooksInComponent(name, content)
	component.Lifecycle = hooks

	return component
}

func (a *ReactAnalyzer) parseClassComponent(line string, lineNum int, result *model.ParseResult) *model.FrameworkComponent {
	parts := strings.Fields(line)
	var name string
	for i, part := range parts {
		if part == "class" && i+1 < len(parts) {
			name = parts[i+1]
			break
		}
	}

	if name == "" {
		return nil
	}

	component := &model.FrameworkComponent{
		Type:       "class_component",
		Name:       name,
		Props:      make([]*model.ComponentProp, 0),
		Events:     make([]string, 0),
		Lifecycle:  make([]string, 0),
		Decorators: make([]string, 0),
		Metadata:   make(map[string]interface{}),
	}

	// Detect lifecycle methods
	lifecycleMethods := []string{
		"componentDidMount",
		"componentDidUpdate",
		"componentWillUnmount",
		"shouldComponentUpdate",
		"getDerivedStateFromProps",
		"getSnapshotBeforeUpdate",
	}

	// Find which lifecycle methods are used
	for _, symbol := range result.Symbols {
		if symbol.Type == model.SymbolTypeMethod {
			for _, method := range lifecycleMethods {
				if symbol.Name == method {
					component.Lifecycle = append(component.Lifecycle, method)
				}
			}
		}
	}

	return component
}

func (a *ReactAnalyzer) findPropsType(componentName string, content string) string {
	// Look for: interface ComponentNameProps or type ComponentNameProps
	patterns := []string{
		`interface\s+` + componentName + `Props`,
		`type\s+` + componentName + `Props`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(content) {
			return componentName + "Props"
		}
	}

	return ""
}

func (a *ReactAnalyzer) findHooksInComponent(componentName string, content string) []string {
	hooks := make([]string, 0)
	hookPatterns := map[string]string{
		"useState":       `useState\(`,
		"useEffect":      `useEffect\(`,
		"useContext":     `useContext\(`,
		"useReducer":     `useReducer\(`,
		"useCallback":    `useCallback\(`,
		"useMemo":        `useMemo\(`,
		"useRef":         `useRef\(`,
		"useLayoutEffect": `useLayoutEffect\(`,
		"useImperativeHandle": `useImperativeHandle\(`,
	}

	// Find component body (simplified - would need proper parsing)
	componentStart := strings.Index(content, componentName)
	if componentStart == -1 {
		return hooks
	}

	// Search for hooks in component
	for hookName, pattern := range hookPatterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(content[componentStart:]) {
			hooks = append(hooks, hookName)
		}
	}

	return hooks
}

func (a *ReactAnalyzer) detectHooks(content string, info *model.FrameworkInfo) {
	hooks := []string{
		"useState", "useEffect", "useContext", "useReducer",
		"useCallback", "useMemo", "useRef", "useLayoutEffect",
		"useImperativeHandle", "useDebugValue",
	}

	for _, hook := range hooks {
		if strings.Contains(content, hook+"(") {
			info.Patterns = append(info.Patterns, "React Hooks: "+hook)
		}
	}

	// Check for custom hooks
	customHookRe := regexp.MustCompile(`const\s+use[A-Z]\w+\s*=`)
	if customHookRe.MatchString(content) {
		info.Patterns = append(info.Patterns, "Custom Hooks detected")
	}
}

func (a *ReactAnalyzer) detectRoutes(content string, info *model.FrameworkInfo) {
	// Detect React Router usage
	if !strings.Contains(content, "react-router") &&
		!strings.Contains(content, "Route") {
		return
	}

	// Find Route components
	routeRe := regexp.MustCompile(`<Route\s+path=["']([^"']+)["']`)
	matches := routeRe.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			route := &model.Route{
				Path:   match[1],
				Method: "GET", // HTTP method not applicable for client-side routing
			}

			// Try to find component
			componentRe := regexp.MustCompile(`component=\{([^}]+)\}`)
			compMatches := componentRe.FindStringSubmatch(content)
			if len(compMatches) >= 2 {
				route.Handler = compMatches[1]
			}

			info.Routes = append(info.Routes, route)
		}
	}
}

func (a *ReactAnalyzer) detectPatterns(content string, info *model.FrameworkInfo) {
	// Context API
	if strings.Contains(content, "createContext") || strings.Contains(content, "useContext") {
		info.Patterns = append(info.Patterns, "Context API")
	}

	// Redux
	if strings.Contains(content, "redux") || strings.Contains(content, "useSelector") {
		info.Patterns = append(info.Patterns, "Redux State Management")
	}

	// Higher-Order Components
	if strings.Contains(content, "withRouter") || strings.Contains(content, "HOC") {
		info.Patterns = append(info.Patterns, "Higher-Order Components")
	}

	// Render Props
	if strings.Contains(content, "render={") && strings.Contains(content, "=>") {
		info.Patterns = append(info.Patterns, "Render Props Pattern")
	}

	// Styled Components
	if strings.Contains(content, "styled-components") || strings.Contains(content, "styled.") {
		info.Patterns = append(info.Patterns, "Styled Components")
	}

	// TypeScript
	if strings.Contains(content, "React.FC") || strings.Contains(content, "FunctionComponent") {
		info.Patterns = append(info.Patterns, "TypeScript with React")
	}
}

func (a *ReactAnalyzer) checkIssues(content string, info *model.FrameworkInfo) {
	// Check for common anti-patterns

	// Missing key in lists
	if strings.Contains(content, ".map(") && !strings.Contains(content, "key=") {
		info.Warnings = append(info.Warnings, "Possible missing 'key' prop in list rendering")
	}

	// Direct state mutation
	if strings.Contains(content, "this.state.") && strings.Contains(content, "=") {
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if strings.Contains(line, "this.state.") &&
				strings.Contains(line, "=") &&
				!strings.Contains(line, "this.setState") {
				info.Warnings = append(info.Warnings, "Possible direct state mutation (use setState)")
				break
			}
		}
	}

	// Large component (> 300 lines)
	lines := strings.Split(content, "\n")
	if len(lines) > 300 {
		info.Warnings = append(info.Warnings, "Large component detected (consider splitting)")
	}

	// Missing useCallback for event handlers
	if strings.Contains(content, "onClick") && !strings.Contains(content, "useCallback") {
		info.Warnings = append(info.Warnings, "Consider using useCallback for event handlers")
	}
}
