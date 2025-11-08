package golang

import (
	"go/ast"
	goparser "go/parser" // Alias standard library parser
	"go/token"
	"strings"
	"fmt"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
)

// Parser is the Go language parser
type Parser struct {
}

// NewParser creates a new Go parser
func NewParser() *Parser {
	return &Parser{}
}

// Language returns the language identifier (e.g., "go")
func (p *Parser) Language() string {
	return "go"
}

// Extensions returns file extensions this parser handles (e.g., [".go"])
func (p *Parser) Extensions() []string {
	return []string{".go"}
}

// Priority returns parser priority (higher = preferred when multiple parsers match)
func (p *Parser) Priority() int {
	return 100
}

// SupportsFramework checks if parser supports specific framework analysis
func (p *Parser) SupportsFramework(framework string) bool {
	return false
}

// Parse parses Go source code
func (p *Parser) Parse(content []byte, filePath string) (*parsing.ParseResult, error) {
	fset := token.NewFileSet()

	// Parse with comments
	file, err := goparser.ParseFile(fset, filePath, content, goparser.ParseComments)
	if err != nil {
		return nil, err
	}

	result := &parsing.ParseResult{
		Symbols:       []*model.Symbol{},
		Imports:       []*model.Import{},
		Relationships: []*model.Relationship{},
		Metadata:      make(map[string]interface{}),
	}

	// Extract package name
	if file.Name != nil {
		result.Metadata["package"] = file.Name.Name
	}

	// Extract imports
	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, `""`)

		// Detect if it's stdlib
		if !strings.Contains(importPath, ".") {
			// importKind = model.ImportKindStdlib
		}

		result.Imports = append(result.Imports, &model.Import{
			Path: importPath,
			Range: model.Range{
				Start: model.Position{Line: fset.Position(imp.Pos()).Line},
				End:   model.Position{Line: fset.Position(imp.End()).Line},
			},
		})
	}

	// Walk the AST
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			symbol := p.extractFunction(node, fset, file)
			result.Symbols = append(result.Symbols, symbol)

		case *ast.GenDecl:
			// Handle type, const, var declarations
			symbols := p.extractGenDecl(node, fset, file)
			result.Symbols = append(result.Symbols, symbols...)
		}

		return true
	})

	return result, nil
}

// extractFunction extracts function/method information
func (p *Parser) extractFunction(fn *ast.FuncDecl, fset *token.FileSet, file *ast.File) *model.Symbol {
	symbol := &model.Symbol{
		Name:      fn.Name.Name,
		Kind:      model.SymbolKindFunction,
		File:      "", // File path will be set by the caller
		Range:     model.Range{Start: model.Position{Line: fset.Position(fn.Pos()).Line}, End: model.Position{Line: fset.Position(fn.End()).Line}},
		Signature: p.buildFunctionSignature(fn),
	}

	// Check if it's a method
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		symbol.Kind = model.SymbolKindMethod
	}

	// Determine visibility (exported = public)
	if ast.IsExported(fn.Name.Name) {
		symbol.Visibility = model.VisibilityPublic
		symbol.Metadata = map[string]string{"is_exported": "true"}
	} else {
		symbol.Visibility = model.VisibilityPrivate
		symbol.Metadata = map[string]string{"is_exported": "false"}
	}

	// Extract documentation
	if fn.Doc != nil {
		symbol.Documentation = strings.TrimSpace(fn.Doc.Text())
	}

	return symbol
}

// extractGenDecl extracts type, const, and var declarations
func (p *Parser) extractGenDecl(decl *ast.GenDecl, fset *token.FileSet, file *ast.File) []*model.Symbol {
	var symbols []*model.Symbol

	for _, spec := range decl.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			symbol := p.extractTypeSpec(s, decl, fset)
			if symbol != nil {
				symbols = append(symbols, symbol)
			}

		case *ast.ValueSpec:
			// Variables or constants
			for _, name := range s.Names {
				symbolKind := model.SymbolKindVariable
				if decl.Tok == token.CONST {
					symbolKind = model.SymbolKindConstant
				}

				symbol := &model.Symbol{
					Name: name.Name,
					Kind: symbolKind,
					File: "", // File path will be set by the caller
					Range: model.Range{
						Start: model.Position{Line: fset.Position(name.Pos()).Line},
						End:   model.Position{Line: fset.Position(name.End()).Line},
					},
				}

				if ast.IsExported(name.Name) {
					symbol.Visibility = model.VisibilityPublic
					symbol.Metadata = map[string]string{"is_exported": "true"}
				} else {
					symbol.Visibility = model.VisibilityPrivate
					symbol.Metadata = map[string]string{"is_exported": "false"}
				}

				if decl.Doc != nil {
					symbol.Documentation = strings.TrimSpace(decl.Doc.Text())
				}

				symbols = append(symbols, symbol)
			}
		}
	}

	return symbols
}

// extractTypeSpec extracts type definitions (struct, interface, etc.)
func (p *Parser) extractTypeSpec(spec *ast.TypeSpec, decl *ast.GenDecl, fset *token.FileSet) *model.Symbol {
	symbol := &model.Symbol{
		Name: spec.Name.Name,
		Kind: model.SymbolKindTypeAlias,
		File: "", // File path will be set by the caller
		Range: model.Range{
			Start: model.Position{Line: fset.Position(spec.Pos()).Line},
			End:   model.Position{Line: fset.Position(spec.End()).Line},
		},
	}

	// Determine specific type
	switch spec.Type.(type) {
	case *ast.StructType:
		symbol.Kind = model.SymbolKindStruct
	case *ast.InterfaceType:
		symbol.Kind = model.SymbolKindInterface
	}

	// Visibility
	if ast.IsExported(spec.Name.Name) {
		symbol.Visibility = model.VisibilityPublic
		symbol.Metadata = map[string]string{"is_exported": "true"}
	} else {
		symbol.Visibility = model.VisibilityPrivate
		symbol.Metadata = map[string]string{"is_exported": "false"}
	}

	// Documentation
	if decl.Doc != nil {
		symbol.Documentation = strings.TrimSpace(decl.Doc.Text())
	}

	return symbol
}

// buildFunctionSignature builds a function signature string
func (p *Parser) buildFunctionSignature(fn *ast.FuncDecl) string {
	var sig strings.Builder

	sig.WriteString("func ")

	// Add receiver for methods
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		sig.WriteString("(")
		for i, field := range fn.Recv.List {
			if i > 0 {
				sig.WriteString(", ")
			}
			if len(field.Names) > 0 {
				sig.WriteString(field.Names[0].Name + " ")
			}
			sig.WriteString(p.exprToString(field.Type))
		}
		sig.WriteString(") ")
	}

	sig.WriteString(fn.Name.Name)
	sig.WriteString("(")

	// Parameters
	if fn.Type.Params != nil {
		for i, field := range fn.Type.Params.List {
			if i > 0 {
				sig.WriteString(", ")
			}
			for j, name := range field.Names {
				if j > 0 {
					sig.WriteString(", ")
				}
				sig.WriteString(name.Name)
			}
			if len(field.Names) > 0 { // Only add space if there are names
				sig.WriteString(" ")
			}
			sig.WriteString(p.exprToString(field.Type))
		}
	}

	sig.WriteString(")")

	// Return type
	if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
		sig.WriteString(" ")
		if len(fn.Type.Results.List) > 1 {
			sig.WriteString("(")
		}
		for i, field := range fn.Type.Results.List {
			if i > 0 {
				sig.WriteString(", ")
			}
			// Result fields can have names, but often don't
			for j, name := range field.Names {
				if j > 0 {
					sig.WriteString(", ")
				}
				sig.WriteString(name.Name)
			}
			if len(field.Names) > 0 {
				sig.WriteString(" ")
			}
			sig.WriteString(p.exprToString(field.Type))
		}
		if len(fn.Type.Results.List) > 1 {
			sig.WriteString(")")
		}
	}

	return sig.String()
}

// exprToString converts an AST expression to its string representation
func (p *Parser) exprToString(expr ast.Expr) string {
	var sb strings.Builder
	// This is a simplified conversion. A full implementation would need to handle
	// various expression types (e.g., StarExpr, ArrayType, MapType, FuncType, etc.)
	// For now, we'll just use the string representation of the expression.
	switch e := expr.(type) {
	case *ast.Ident:
		sb.WriteString(e.Name)
	case *ast.StarExpr:
		sb.WriteString("*" + p.exprToString(e.X))
	case *ast.ArrayType:
		sb.WriteString("[]" + p.exprToString(e.Elt))
	case *ast.MapType:
		sb.WriteString("map[" + p.exprToString(e.Key) + "]" + p.exprToString(e.Value))
	case *ast.FuncType:
		sb.WriteString("func")
		// TODO: Recursively build function type signature
	default:
		sb.WriteString(fmt.Sprintf("%v", expr))
	}
	return sb.String()
}
