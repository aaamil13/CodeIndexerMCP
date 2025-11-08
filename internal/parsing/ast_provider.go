package parsing

import (
	"fmt"
	"sync"
	
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/aaamil13/CodeIndexerMCP/internal/model"
)

type ParseResult struct {
    Tree        *sitter.Tree
    Language    string
    SourceCode  []byte
    RootNode    *sitter.Node
    ParseErrors []ParseError
    
    // Higher-level model elements
    Symbols       []*model.Symbol       `json:"symbols"`
    Imports       []*model.Import       `json:"imports"`
    Relationships []*model.Relationship `json:"relationships"`
    Metadata      map[string]interface{} `json:"metadata"`
}

type ParseError struct {
    Message  string
    Line     int
    Column   int
    Byte     uint32
}

type ASTProvider struct {
    grammarManager *GrammarManager
    parserPool     sync.Pool
}

func NewASTProvider(gm *GrammarManager) *ASTProvider {
    return &ASTProvider{
        grammarManager: gm,
        parserPool: sync.Pool{
            New: func() interface{} {
                return sitter.NewParser()
            },
        },
    }
}

func (ap *ASTProvider) Parse(language string, content []byte) (*ParseResult, error) {
    grammar, err := ap.grammarManager.GetLanguage(language)
    if err != nil {
        return nil, err
    }
    
    // Вземане на parser от pool
    parser := ap.parserPool.Get().(*sitter.Parser)
    defer ap.parserPool.Put(parser)
    
    parser.SetLanguage(grammar)
    
    tree := parser.Parse(nil, content)
    if tree == nil {
        return nil, fmt.Errorf("failed to parse %s code", language)
    }
    
    root := tree.RootNode()
    
    result := &ParseResult{
        Tree:       tree,
        Language:   language,
        SourceCode: content,
        RootNode:   root,
    }
    
    // Проверка за parse errors
    if root.HasError() {
        result.ParseErrors = ap.extractErrors(root, content)
    }
    
    return result, nil
}

func (ap *ASTProvider) extractErrors(node *sitter.Node, source []byte) []ParseError {
    var errors []ParseError
    
    var walk func(*sitter.Node)
    walk = func(n *sitter.Node) {
        if n.Type() == "ERROR" || n.IsMissing() {
            errors = append(errors, ParseError{
                Message: fmt.Sprintf("Syntax error at node: %s", n.Type()),
                Line:    int(n.StartPoint().Row) + 1,
                Column:  int(n.StartPoint().Column) + 1,
                Byte:    n.StartByte(),
            })
        }
        
        for i := 0; i < int(n.ChildCount()); i++ {
            walk(n.Child(i))
        }
    }
    
    walk(node)
    return errors
}

func (ap *ASTProvider) ParseIncremental(oldTree *sitter.Tree, language string, newContent []byte, edits []sitter.EditInput) (*ParseResult, error) {
    grammar, err := ap.grammarManager.GetLanguage(language)
    if err != nil {
        return nil, err
    }
    
    parser := ap.parserPool.Get().(*sitter.Parser)
    defer ap.parserPool.Put(parser)
    
    parser.SetLanguage(grammar)
    
    // Прилагане на промените
    for _, edit := range edits {
        oldTree.Edit(edit)
    }
    
    newTree := parser.Parse(oldTree, newContent)
    
    result := &ParseResult{
        Tree:       newTree,
        Language:   language,
        SourceCode: newContent,
        RootNode:   newTree.RootNode(),
    }
    
    return result, nil
}

func (pr *ParseResult) Close() {
    if pr.Tree != nil {
        pr.Tree.Close()
    }
}
