package types

// FrameworkInfo contains framework-specific information
type FrameworkInfo struct {
	Name            string                 `json:"name"`             // e.g., "react", "django", "spring"
	Version         string                 `json:"version"`          // Framework version if detected
	Type            string                 `json:"type"`             // "frontend", "backend", "fullstack", "testing"
	Components      []*FrameworkComponent  `json:"components"`       // Framework-specific components
	Routes          []*Route               `json:"routes"`           // Web routes/endpoints
	Models          []*Model               `json:"models"`           // Data models
	Dependencies    []string               `json:"dependencies"`     // Framework dependencies
	Configuration   map[string]interface{} `json:"configuration"`    // Framework config
	Patterns        []string               `json:"patterns"`         // Detected patterns
	BestPractices   []string               `json:"best_practices"`   // Detected best practices
	Warnings        []string               `json:"warnings"`         // Potential issues
}

// FrameworkComponent represents a framework-specific component
type FrameworkComponent struct {
	Type        string                 `json:"type"`         // "component", "controller", "service", "middleware"
	Name        string                 `json:"name"`
	Symbol      *Symbol                `json:"symbol"`       // Associated symbol
	Metadata    map[string]interface{} `json:"metadata"`     // Component-specific metadata
	Props       []*ComponentProp       `json:"props"`        // Props/inputs
	Events      []string               `json:"events"`       // Events/outputs
	Lifecycle   []string               `json:"lifecycle"`    // Lifecycle methods
	Decorators  []string               `json:"decorators"`   // Decorators/annotations
}

// ComponentProp represents a component property/prop
type ComponentProp struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
}

// Route represents a web route/endpoint
type Route struct {
	Path        string            `json:"path"`
	Method      string            `json:"method"`      // GET, POST, PUT, DELETE, etc.
	Handler     string            `json:"handler"`     // Handler function/method name
	HandlerSymbol *Symbol         `json:"handler_symbol,omitempty"`
	Middleware  []string          `json:"middleware"`
	Parameters  []*RouteParameter `json:"parameters"`
	Description string            `json:"description,omitempty"`
}

// RouteParameter represents a route parameter
type RouteParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`     // "path", "query", "body", "header"
	DataType    string `json:"data_type"` // "string", "int", "object", etc.
	Required    bool   `json:"required"`
	Description string `json:"description,omitempty"`
}

// Model represents a data model (ORM, database schema, etc.)
type Model struct {
	Name        string              `json:"name"`
	Symbol      *Symbol             `json:"symbol"`
	Table       string              `json:"table,omitempty"`      // Database table name
	Fields      []*ModelField       `json:"fields"`
	Relations   []*ModelRelation    `json:"relations"`
	Indexes     []string            `json:"indexes"`
	Validations []*ModelValidation  `json:"validations"`
}

// ModelField represents a model field
type ModelField struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	DatabaseType string `json:"database_type,omitempty"`
	Nullable    bool   `json:"nullable"`
	Primary     bool   `json:"primary"`
	Unique      bool   `json:"unique"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
}

// ModelRelation represents a relationship between models
type ModelRelation struct {
	Type         string `json:"type"`          // "has_one", "has_many", "belongs_to", "many_to_many"
	RelatedModel string `json:"related_model"`
	ForeignKey   string `json:"foreign_key,omitempty"`
	Through      string `json:"through,omitempty"` // For many-to-many
}

// ModelValidation represents a model validation rule
type ModelValidation struct {
	Field   string `json:"field"`
	Rule    string `json:"rule"`    // "required", "email", "min_length", etc.
	Value   string `json:"value,omitempty"`
	Message string `json:"message,omitempty"`
}

// FrameworkPattern represents a detected framework pattern
type FrameworkPattern struct {
	Name        string   `json:"name"`        // "MVC", "REST", "GraphQL", "Microservices"
	Type        string   `json:"type"`        // "architectural", "design", "integration"
	Confidence  float64  `json:"confidence"`  // 0.0 - 1.0
	Evidence    []string `json:"evidence"`    // Why pattern was detected
	Files       []string `json:"files"`       // Files implementing pattern
	Description string   `json:"description"`
}

// LanguageFeature represents language-specific features detected
type LanguageFeature struct {
	Name        string  `json:"name"`        // "async/await", "generics", "decorators"
	Language    string  `json:"language"`
	UsageCount  int     `json:"usage_count"`
	Examples    []string `json:"examples"`
	BestPractice string `json:"best_practice,omitempty"`
}

// ParseResult extension to include frameworks
func (pr *ParseResult) AddFramework(framework *FrameworkInfo) {
	if pr.Frameworks == nil {
		pr.Frameworks = make([]*FrameworkInfo, 0)
	}
	pr.Frameworks = append(pr.Frameworks, framework)
}
