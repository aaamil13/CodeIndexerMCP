package golang

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// Parser is the Go language parser
type Parser struct{}

// NewParser creates a new Go parser
func NewParser() *Parser {
	return &Parser{}
}

// Language returns the language name
func (p *Parser) Language() string {
	return "go"
}

// Extensions returns supported file extensions
func (p *Parser) Extensions() []string {
	return []string{".go"}
}

// CanParse checks if this parser can handle the file
func (p *Parser) CanParse(filePath string) bool {
	ext := filepath.Ext(filePath)
	return ext == ".go"
}

// Parse parses Go source code
func (p *Parser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
	fset := token.NewFileSet()

	// Parse with comments
	file, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
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
		symbol.Documentation = fn.Doc.Text()
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
					symbol.Documentation = decl.Doc.Text()
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
		symbol.Documentation = decl.Doc.Text()
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
		// Simplified - would need proper formatting
		sig.WriteString("receiver")
		sig.WriteString(") ")
	}

	sig.WriteString(fn.Name.Name)
	sig.WriteString("(")

	// Parameters (simplified)
	if fn.Type.Params != nil {
		for i, field := range fn.Type.Params.List {
			if i > 0 {
				sig.WriteString(", ")
			}
			// Add parameter names/types (simplified)
			if len(field.Names) > 0 {
				sig.WriteString(field.Names[0].Name)
			} else {
				sig.WriteString("_")
			}
		}
	}

	sig.WriteString(")")

	// Return type (simplified)
	if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
		sig.WriteString(" ")
		if len(fn.Type.Results.List) > 1 {
			sig.WriteString("(...)")
		} else {
			sig.WriteString("...")
		}
	}

	return sig.String()
}
