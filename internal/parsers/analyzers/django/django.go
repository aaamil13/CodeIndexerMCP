package django

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
)

// DjangoAnalyzer analyzes Django framework patterns
type DjangoAnalyzer struct{}

// NewDjangoAnalyzer creates a new Django analyzer
func NewDjangoAnalyzer() *DjangoAnalyzer {
	return &DjangoAnalyzer{}
}

// Framework returns the framework name
func (a *DjangoAnalyzer) Framework() string {
	return "django"
}

// Language returns the target language
func (a *DjangoAnalyzer) Language() string {
	return "python"
}

// DetectFramework detects if file uses Django
func (a *DjangoAnalyzer) DetectFramework(content []byte, filePath string) bool {
	contentStr := string(content)

	// Check for Django imports
	djangoImports := []string{
		"from django",
		"import django",
		"django.db.models",
		"django.views",
		"django.urls",
		"django.http",
		"django.contrib",
	}

	for _, imp := range djangoImports {
		if strings.Contains(contentStr, imp) {
			return true
		}
	}

	// Check for Django-specific patterns
	if strings.Contains(contentStr, "models.Model") ||
		strings.Contains(contentStr, "models.CharField") ||
		strings.Contains(contentStr, "urlpatterns") {
		return true
	}

	return false
}

// Analyze analyzes Django-specific patterns
func (a *DjangoAnalyzer) Analyze(result *model.ParseResult, content []byte) (*model.FrameworkInfo, error) {
	info := &model.FrameworkInfo{
		Name: "django",
	}

	contentStr := string(content)

	// Detect Django version (from imports)
	a.detectVersion(contentStr, info)

	// Extract models
	a.extractModels(result, contentStr, info)

	// Extract views
	a.extractViews(result, contentStr, info)

	// Extract URL patterns
	a.extractURLPatterns(contentStr, info)

	// Detect Django REST Framework
	if strings.Contains(contentStr, "rest_framework") {
		// info.Patterns = append(info.Patterns, "Django REST Framework")
		a.extractSerializers(result, contentStr, info)
	}

	// Detect forms
	a.extractForms(result, contentStr, info)

	// Detect admin customizations
	a.detectAdmin(contentStr, info)

	// Check for common issues
	a.checkIssues(contentStr, info)

	return info, nil
}

func (a *DjangoAnalyzer) detectVersion(content string, info *model.FrameworkInfo) {
	// Detection based on imports and patterns
	if strings.Contains(content, "from django.urls import path") {
		info.Version = "2.0+" // path() was introduced in Django 2.0
	} else if strings.Contains(content, "from django.conf.urls import url") {
		info.Version = "1.x"
	}

	// Async views indicate Django 3.1+
	if strings.Contains(content, "async def") && strings.Contains(content, "django") {
		info.Version = "3.1+"
	}
}

func (a *DjangoAnalyzer) extractModels(result *model.ParseResult, content string, info *model.FrameworkInfo) {
	// Find classes that inherit from models.Model
	for _, symbol := range result.Symbols {
		if symbol.Kind != model.SymbolKindClass {
			continue
		}

		// Check if it's a Django model
		if a.isModelClass(symbol.Name, content) {
			model := a.parseModel(symbol, content)
			if model != nil {
				// info.Models = append(info.Models, model) // Removed as FrameworkInfo does not have Models
			}
		}
	}
}

func (a *DjangoAnalyzer) isModelClass(className string, content string) bool {
	// Look for: class ClassName(models.Model) or class ClassName(Model)
	patterns := []string{
		`class\s+` + className + `\s*\(\s*models\.Model\s*\)`,
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

func (a *DjangoAnalyzer) parseModel(symbol *model.Symbol, content string) *model.Model {
	model := &model.Model{
		Name:        symbol.Name,
		Fields:      make([]*model.ModelField, 0),
		Relations:   make([]*model.ModelRelation, 0),
		Indexes:     make([]string, 0),
		Validations: make([]*model.ModelValidation, 0),
	}

	// Extract table name from Meta class
	metaRe := regexp.MustCompile(`class\s+Meta:[\s\S]*?db_table\s*=\s*["']([^"']+)["']`)
	if matches := metaRe.FindStringSubmatch(content); len(matches) >= 2 {
		model.Table = matches[1]
	}

	// Extract fields
	lines := strings.Split(content, "\n")
	inClass := false

	for _, line := range lines { // Removed 'i' from here
		trimmed := strings.TrimSpace(line)

		// Find class definition
		if strings.Contains(trimmed, "class "+symbol.Name) {
			inClass = true
			continue
		}

		if !inClass {
			continue
		}

		// Check if we've left the class
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
			break
		}

		// Check if this is a field definition
		if a.isFieldDefinition(trimmed) {
			field := a.parseField(trimmed)
			if field != nil {
				model.Fields = append(model.Fields, field)
			}
		}

		// Check for relationships
		if relation := a.parseRelation(trimmed); relation != nil {
			model.Relations = append(model.Relations, relation)
		}
	}

	return model
}

func (a *DjangoAnalyzer) isFieldDefinition(line string) bool {
	fieldTypes := []string{
		"models.CharField",
		"models.IntegerField",
		"models.TextField",
		"models.DateField",
		"models.DateTimeField",
		"models.BooleanField",
		"models.EmailField",
		"models.URLField",
		"models.FileField",
		"models.ImageField",
		"models.DecimalField",
		"models.FloatField",
		"models.JSONField",
		"models.UUIDField",
	}

	for _, fieldType := range fieldTypes {
		if strings.Contains(line, fieldType) {
			return true
		}
	}

	return false
}

func (a *DjangoAnalyzer) parseField(line string) *model.ModelField {
	// Extract: field_name = models.FieldType(...)
	parts := strings.Split(line, "=")
	if len(parts) < 2 {
		return nil
	}

	fieldName := strings.TrimSpace(parts[0])
	fieldDef := strings.TrimSpace(parts[1])

	field := &model.ModelField{
		Name: fieldName,
	}

	// Extract field type
	if strings.Contains(fieldDef, "CharField") {
		field.Type = "string"
		field.DatabaseType = "VARCHAR"
	} else if strings.Contains(fieldDef, "IntegerField") {
		field.Type = "int"
		field.DatabaseType = "INTEGER"
	} else if strings.Contains(fieldDef, "TextField") {
		field.Type = "string"
		field.DatabaseType = "TEXT"
	} else if strings.Contains(fieldDef, "DateTimeField") {
		field.Type = "datetime"
		field.DatabaseType = "DATETIME"
	} else if strings.Contains(fieldDef, "BooleanField") {
		field.Type = "bool"
		field.DatabaseType = "BOOLEAN"
	} else if strings.Contains(fieldDef, "EmailField") {
		field.Type = "string"
		field.DatabaseType = "VARCHAR"
	}

	// Check field options
	field.Nullable = strings.Contains(fieldDef, "null=True")
	field.Unique = strings.Contains(fieldDef, "unique=True")
	field.Primary = strings.Contains(fieldDef, "primary_key=True")

	// Extract default value
	defaultRe := regexp.MustCompile(`default=([^,)]+)`)
	if matches := defaultRe.FindStringSubmatch(fieldDef); len(matches) >= 2 {
		field.Default = strings.TrimSpace(matches[1])
	}

	return field
}

func (a *DjangoAnalyzer) parseRelation(line string) *model.ModelRelation {
	relationTypes := map[string]string{
		"ForeignKey":     "belongs_to",
		"OneToOneField":  "has_one",
		"ManyToManyField": "many_to_many",
	}

	for fieldType, relationType := range relationTypes {
		if strings.Contains(line, "models."+fieldType) {
			relation := &model.ModelRelation{
				Type: relationType,
			}

			// Extract related model
			re := regexp.MustCompile(`models\.` + fieldType + `\s*\(\s*["']?([A-Za-z0-9_.]+)["']?`)
			if matches := re.FindStringSubmatch(line); len(matches) >= 2 {
				relation.RelatedModel = matches[1]
			}

			return relation
		}
	}

	return nil
}

func (a *DjangoAnalyzer) extractViews(result *model.ParseResult, content string, info *model.FrameworkInfo) {
	viewTypes := []string{
		"View", "TemplateView", "ListView", "DetailView",
		"CreateView", "UpdateView", "DeleteView", "FormView",
	}

	for _, symbol := range result.Symbols {
		if symbol.Type == model.SymbolTypeClass {
			// Check if it's a view class
			for _, viewType := range viewTypes {
				if strings.Contains(content, "class "+symbol.Name) &&
					strings.Contains(content, viewType) {
					component := &model.FrameworkComponent{
						Type:     "view",
						Name:     symbol.Name,
						Symbol:   symbol,
						Metadata: make(map[string]interface{}),
					}
					component.Metadata["view_type"] = viewType
					info.Components = append(info.Components, component)
					break
				}
			}
		}

		// Function-based views
		if symbol.Type == model.SymbolTypeFunction {
			// Check if function has request parameter
			if strings.Contains(symbol.Signature, "request") {
				component := &model.FrameworkComponent{
					Type:     "view",
					Name:     symbol.Name,
					Symbol:   symbol,
					Metadata: make(map[string]interface{}),
				}
				component.Metadata["view_type"] = "function"
				info.Components = append(info.Components, component)
			}
		}
	}
}

func (a *DjangoAnalyzer) extractURLPatterns(content string, info *model.FrameworkInfo) {
	// Find: path('route/', view, name='name')
	pathRe := regexp.MustCompile(`path\s*\(\s*["']([^"']+)["']\s*,\s*([^,]+)`)
	matches := pathRe.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			route := &model.Route{
				Path:    match[1],
				Handler: strings.TrimSpace(match[2]),
			}

			// Extract name parameter
			nameRe := regexp.MustCompile(`name\s*=\s*["']([^"']+)["']`)
			if nameMatches := nameRe.FindStringSubmatch(content); len(nameMatches) >= 2 {
				route.Description = nameMatches[1]
			}

			info.Routes = append(info.Routes, route)
		}
	}
}

func (a *DjangoAnalyzer) extractSerializers(result *model.ParseResult, content string, info *model.FrameworkInfo) {
	for _, symbol := range result.Symbols {
		if symbol.Type == model.SymbolTypeClass {
			if strings.Contains(content, "Serializer") {
				component := &model.FrameworkComponent{
					Type:     "serializer",
					Name:     symbol.Name,
					Symbol:   symbol,
					Metadata: make(map[string]interface{}),
				}
				info.Components = append(info.Components, component)
			}
		}
	}
}

func (a *DjangoAnalyzer) extractForms(result *model.ParseResult, content string, info *model.FrameworkInfo) {
	for _, symbol := range result.Symbols {
		if symbol.Type == model.SymbolTypeClass {
			if strings.Contains(content, "forms.Form") ||
				strings.Contains(content, "forms.ModelForm") {
				component := &model.FrameworkComponent{
					Type:     "form",
					Name:     symbol.Name,
					Symbol:   symbol,
					Metadata: make(map[string]interface{}),
				}
				info.Components = append(info.Components, component)
			}
		}
	}
}

func (a *DjangoAnalyzer) detectAdmin(content string, info *model.FrameworkInfo) {
	if strings.Contains(content, "admin.site.register") ||
		strings.Contains(content, "admin.ModelAdmin") {
		// info.Patterns = append(info.Patterns, "Django Admin customization") // Removed as FrameworkInfo does not have Patterns
	}
}

func (a *DjangoAnalyzer) checkIssues(content string, info *model.FrameworkInfo) {
	// Check for common Django anti-patterns

	// Raw SQL usage
	if strings.Contains(content, ".raw(") || strings.Contains(content, "execute(") {
		// info.Warnings = append(info.Warnings, "Raw SQL detected - consider using ORM")
	}

	// Missing CSRF protection
	if strings.Contains(content, "@csrf_exempt") {
		// info.Warnings = append(info.Warnings, "CSRF protection disabled - security risk")
	}

	// N+1 query problem indicators
	if strings.Contains(content, "for ") && strings.Contains(content, ".all()") {
		// info.Warnings = append(info.Warnings, "Possible N+1 query problem - use select_related/prefetch_related")
	}

	// Missing __str__ method in models
	if strings.Contains(content, "models.Model") && !strings.Contains(content, "def __str__") {
		// info.Warnings = append(info.Warnings, "Model missing __str__ method")
	}
}
