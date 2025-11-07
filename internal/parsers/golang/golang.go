package golang

import (
	"go/ast"
	goparser "go/parser" // Alias standard library parser
	"go/token"
	"strings"
	"fmt" // Added fmt import for exprToString default case

	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// Parser is the Go language parser
type Parser struct {
	*parser.BaseParser
}

// NewParser creates a new Go parser
func NewParser() *Parser {
	return &Parser{
		BaseParser: parser.NewBaseParser("go", []string{".go"}, 100),
	}
}

// Parse parses Go source code
func (p *Parser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
	fset := token.NewFileSet()

	// Parse with comments
	file, err := goparser.ParseFile(fset, filePath, content, goparser.ParseComments)
	if err != nil {
		return nil, err
	}

	result := &types.ParseResult{
		Symbols:       []*types.Symbol{},
		Imports:       []*types.Import{},
		Relationships: []*types.Relationship{},
		Metadata:      make(map[string]interface{}),
	}

	// Extract package name
	if file.Name != nil {
		result.Metadata["package"] = file.Name.Name
	}

	// Extract imports
	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")
		importType := types.ImportTypeExternal

		// Detect if it's stdlib
		if !strings.Contains(importPath, ".") {
			importType = types.ImportTypeStdlib
		}

	result.Imports = append(result.Imports, &types.Import{
			Source:     importPath,
			ImportType: importType,
			LineNumber: fset.Position(imp.Pos()).Line,
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
func (p *Parser) extractFunction(fn *ast.FuncDecl, fset *token.FileSet, file *ast.File) *types.Symbol {
	symbol := &types.Symbol{
		Name:      fn.Name.Name,
		Type:      types.SymbolTypeFunction,
		StartLine: fset.Position(fn.Pos()).Line,
		EndLine:   fset.Position(fn.End()).Line,
		Signature: p.buildFunctionSignature(fn),
	}

	// Check if it's a method
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		symbol.Type = types.SymbolTypeMethod
	}

	// Determine visibility (exported = public)
	if ast.IsExported(fn.Name.Name) {
		symbol.Visibility = types.VisibilityPublic
		symbol.IsExported = true
	} else {
		symbol.Visibility = types.VisibilityPrivate
		symbol.IsExported = false
	}

	// Extract documentation
	if fn.Doc != nil {
		symbol.Documentation = strings.TrimSpace(fn.Doc.Text())
	}

	return symbol
}

// extractGenDecl extracts type, const, and var declarations
func (p *Parser) extractGenDecl(decl *ast.GenDecl, fset *token.FileSet, file *ast.File) []*types.Symbol {
	var symbols []*types.Symbol

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
				symbolType := types.SymbolTypeVariable
				if decl.Tok == token.CONST {
					symbolType = types.SymbolTypeConstant
				}

				symbol := &types.Symbol{
					Name:      name.Name,
					Type:      symbolType,
					StartLine: fset.Position(name.Pos()).Line,
					EndLine:   fset.Position(name.End()).Line,
				}

				if ast.IsExported(name.Name) {
					symbol.Visibility = types.VisibilityPublic
					symbol.IsExported = true
				} else {
					symbol.Visibility = types.VisibilityPrivate
					symbol.IsExported = false
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
func (p *Parser) extractTypeSpec(spec *ast.TypeSpec, decl *ast.GenDecl, fset *token.FileSet) *types.Symbol {
	symbol := &types.Symbol{
		Name:      spec.Name.Name,
		Type:      types.SymbolTypeType,
		StartLine: fset.Position(spec.Pos()).Line,
		EndLine:   fset.Position(spec.End()).Line,
	}

	// Determine specific type
	switch spec.Type.(type) {
	case *ast.StructType:
		symbol.Type = types.SymbolTypeStruct
	case *ast.InterfaceType:
		symbol.Type = types.SymbolTypeInterface
	}

	// Visibility
	if ast.IsExported(spec.Name.Name) {
		symbol.Visibility = types.VisibilityPublic
		symbol.IsExported = true
	} else {
		symbol.Visibility = types.VisibilityPrivate
		symbol.IsExported = false
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
	if fn.Recv != nil {
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
