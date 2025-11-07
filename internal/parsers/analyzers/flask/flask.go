package flask

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// FlaskAnalyzer analyzes Flask framework patterns
type FlaskAnalyzer struct{}

// NewFlaskAnalyzer creates a new Flask analyzer
func NewFlaskAnalyzer() *FlaskAnalyzer {
	return &FlaskAnalyzer{}
}

// Framework returns the framework name
func (a *FlaskAnalyzer) Framework() string {
	return "flask"
}

// Language returns the target language
func (a *FlaskAnalyzer) Language() string {
	return "python"
}

// DetectFramework detects if file uses Flask
func (a *FlaskAnalyzer) DetectFramework(content []byte, filePath string) bool {
	contentStr := string(content)

	// Check for Flask imports
	flaskImports := []string{
		"from flask import",
		"import flask",
		"Flask(__name__)",
	}

	for _, imp := range flaskImports {
		if strings.Contains(contentStr, imp) {
			return true
		}
	}

	// Check for Flask decorators
	if strings.Contains(contentStr, "@app.route") ||
		strings.Contains(contentStr, "@blueprint.route") {
		return true
	}

	return false
}

// Analyze analyzes Flask-specific patterns
func (a *FlaskAnalyzer) Analyze(result *types.ParseResult, content []byte) (*types.FrameworkInfo, error) {
	info := &types.FrameworkInfo{
		Name:         "flask",
		Type:         "backend",
		Components:   make([]*types.FrameworkComponent, 0),
		Routes:       make([]*types.Route, 0),
		Dependencies: make([]string, 0),
		Patterns:     make([]string, 0),
		Warnings:     make([]string, 0),
	}

	contentStr := string(content)

	// Extract routes
	a.extractRoutes(result, contentStr, info)

	// Detect blueprints
	a.detectBlueprints(contentStr, info)

	// Detect extensions
	a.detectExtensions(contentStr, info)

	// Detect Flask-RESTful
	if strings.Contains(contentStr, "flask_restful") || strings.Contains(contentStr, "Api(app)") {
		info.Patterns = append(info.Patterns, "Flask-RESTful")
		a.extractRESTfulResources(result, contentStr, info)
	}

	// Detect Flask-SQLAlchemy
	if strings.Contains(contentStr, "flask_sqlalchemy") {
		info.Patterns = append(info.Patterns, "Flask-SQLAlchemy")
		a.extractModels(result, contentStr, info)
	}

	// Check for common issues
	a.checkIssues(contentStr, info)

	return info, nil
}

func (a *FlaskAnalyzer) extractRoutes(result *types.ParseResult, content string, info *types.FrameworkInfo) {
	lines := strings.Split(content, "\n")

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Find @app.route or @blueprint.route decorators
		if strings.Contains(line, "@app.route") || strings.Contains(line, "@blueprint.route") {
			route := a.parseRoute(line, lines, i, result)
			if route != nil {
				info.Routes = append(info.Routes, route)
			}
		}
	}
}

func (a *FlaskAnalyzer) parseRoute(decoratorLine string, lines []string, lineIndex int, result *types.ParseResult) *types.Route {
	route := &types.Route{
		Method:     "GET", // Default
		Middleware: make([]string, 0),
		Parameters: make([]*types.RouteParameter, 0),
	}

	// Extract path: @app.route('/path')
	pathRe := regexp.MustCompile(`@\w+\.route\s*\(\s*["']([^"']+)["']`)
	if matches := pathRe.FindStringSubmatch(decoratorLine); len(matches) >= 2 {
		route.Path = matches[1]
	}

	// Extract methods: methods=['POST', 'GET']
	methodsRe := regexp.MustCompile(`methods\s*=\s*\[([^\]]+)\]`)
	if matches := methodsRe.FindStringSubmatch(decoratorLine); len(matches) >= 2 {
		methodsStr := matches[1]
		methods := strings.Split(methodsStr, ",")
		if len(methods) > 0 {
			firstMethod := strings.Trim(strings.TrimSpace(methods[0]), `'"`)
			route.Method = firstMethod
		}
	}

	// Find the function definition (next non-empty line)
	for j := lineIndex + 1; j < len(lines); j++ {
		line := strings.TrimSpace(lines[j])
		if line == "" || strings.HasPrefix(line, "@") {
			continue
		}

		if strings.HasPrefix(line, "def ") {
			// Extract function name
			funcRe := regexp.MustCompile(`def\s+(\w+)\s*\(`)
			if matches := funcRe.FindStringSubmatch(line); len(matches) >= 2 {
				route.Handler = matches[1]
			}

			// Extract route parameters from path
			route.Parameters = a.extractRouteParameters(route.Path)
			break
		}
	}

	return route
}

func (a *FlaskAnalyzer) extractRouteParameters(path string) []*types.RouteParameter {
	params := make([]*types.RouteParameter, 0)

	// Find <param> or <type:param> patterns
	paramRe := regexp.MustCompile(`<(?:(\w+):)?(\w+)>`)
	matches := paramRe.FindAllStringSubmatch(path, -1)

	for _, match := range matches {
		param := &types.RouteParameter{
			Type:     "path",
			Required: true,
		}

		if len(match) >= 3 {
			if match[1] != "" {
				param.DataType = match[1] // int, string, path, etc.
			} else {
				param.DataType = "string"
			}
			param.Name = match[2]
		}

		params = append(params, param)
	}

	return params
}

func (a *FlaskAnalyzer) detectBlueprints(content string, info *types.FrameworkInfo) {
	// Find Blueprint definitions
	blueprintRe := regexp.MustCompile(`(\w+)\s*=\s*Blueprint\s*\(`)
	matches := blueprintRe.FindAllStringSubmatch(content, -1)

	if len(matches) > 0 {
		info.Patterns = append(info.Patterns, "Flask Blueprints")
		for _, match := range matches {
			if len(match) >= 2 {
				component := &types.FrameworkComponent{
					Type:     "blueprint",
					Name:     match[1],
					Metadata: make(map[string]interface{}),
				}
				info.Components = append(info.Components, component)
			}
		}
	}
}

func (a *FlaskAnalyzer) detectExtensions(content string, info *types.FrameworkInfo) {
	extensions := map[string]string{
		"flask_login":       "Flask-Login (Authentication)",
		"flask_wtf":         "Flask-WTF (Forms)",
		"flask_mail":        "Flask-Mail",
		"flask_migrate":     "Flask-Migrate (Database migrations)",
		"flask_cors":        "Flask-CORS",
		"flask_jwt":         "Flask-JWT (JSON Web Tokens)",
		"flask_socketio":    "Flask-SocketIO (WebSockets)",
		"flask_caching":     "Flask-Caching",
	}

	for extensionImport, description := range extensions {
		if strings.Contains(content, extensionImport) {
			info.Patterns = append(info.Patterns, description)
		}
	}
}

func (a *FlaskAnalyzer) extractRESTfulResources(result *types.ParseResult, content string, info *types.FrameworkInfo) {
	// Find classes that inherit from Resource
	for _, symbol := range result.Symbols {
		if symbol.Type == types.SymbolTypeClass {
			if strings.Contains(content, "class "+symbol.Name) &&
				strings.Contains(content, "Resource") {
				component := &types.FrameworkComponent{
					Type:     "resource",
					Name:     symbol.Name,
					Symbol:   symbol,
					Metadata: make(map[string]interface{}),
				}

				// Find HTTP methods (get, post, put, delete, etc.)
				methods := []string{}
				for _, methodSymbol := range result.Symbols {
					if methodSymbol.Type == types.SymbolTypeMethod {
						httpMethods := []string{"get", "post", "put", "delete", "patch"}
						for _, httpMethod := range httpMethods {
							if methodSymbol.Name == httpMethod {
								methods = append(methods, httpMethod)
							}
						}
					}
				}
				component.Metadata["http_methods"] = methods

				info.Components = append(info.Components, component)
			}
		}
	}
}

func (a *FlaskAnalyzer) extractModels(result *types.ParseResult, content string, info *types.FrameworkInfo) {
	// Find classes that inherit from db.Model
	for _, symbol := range result.Symbols {
		if symbol.Type == types.SymbolTypeClass {
			if a.isModelClass(symbol.Name, content) {
				model := &types.Model{
					Name:   symbol.Name,
					Symbol: symbol,
					Fields: make([]*types.ModelField, 0),
				}

				// Extract fields (simplified)
				lines := strings.Split(content, "\n")
				inClass := false

				for _, line := range lines {
					trimmed := strings.TrimSpace(line)

					if strings.Contains(trimmed, "class "+symbol.Name) {
						inClass = true
						continue
					}

					if inClass && strings.Contains(trimmed, "db.Column") {
						// Extract field
						if parts := strings.Split(trimmed, "="); len(parts) >= 1 {
							fieldName := strings.TrimSpace(parts[0])
							field := &types.ModelField{
								Name: fieldName,
							}

							// Detect field type
							if strings.Contains(trimmed, "db.String") {
								field.Type = "string"
							} else if strings.Contains(trimmed, "db.Integer") {
								field.Type = "int"
							} else if strings.Contains(trimmed, "db.Boolean") {
								field.Type = "bool"
							} else if strings.Contains(trimmed, "db.DateTime") {
								field.Type = "datetime"
							}

							model.Fields = append(model.Fields, field)
						}
					}

					// Exit class
					if inClass && len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
						break
					}
				}

				info.Models = append(info.Models, model)
			}
		}
	}
}

func (a *FlaskAnalyzer) isModelClass(className string, content string) bool {
	patterns := []string{
		`class\s+` + className + `\s*\(\s*db\.Model\s*\)`,
		`class\s+` + className + `\s*\(\s*Model\s*\)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(content) {
			return true
		}
	}

	return false
}

func (a *FlaskAnalyzer) checkIssues(content string, info *types.FrameworkInfo) {
	// Check for debug mode in production
	if strings.Contains(content, "app.run(debug=True)") {
		info.Warnings = append(info.Warnings, "Debug mode enabled - should be disabled in production")
	}

	// Check for hardcoded secret keys
	if strings.Contains(content, "SECRET_KEY = ") && !strings.Contains(content, "os.environ") {
		info.Warnings = append(info.Warnings, "Hardcoded SECRET_KEY - use environment variables")
	}

	// Check for missing CSRF protection
	if !strings.Contains(content, "flask_wtf") && !strings.Contains(content, "CSRFProtect") {
		info.Warnings = append(info.Warnings, "No CSRF protection detected - consider using Flask-WTF")
	}

	// Check for SQL injection risks
	if strings.Contains(content, ".execute(") && strings.Contains(content, "+") {
		info.Warnings = append(info.Warnings, "Possible SQL injection risk - use parameterized queries")
	}

	// Check for missing error handlers
	if !strings.Contains(content, "@app.errorhandler") {
		info.Warnings = append(info.Warnings, "No error handlers defined - add @app.errorhandler")
	}
}
