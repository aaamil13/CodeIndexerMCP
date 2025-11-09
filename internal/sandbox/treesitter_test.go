package sandbox

import (
    "fmt"
    "testing"
    
    sitter "github.com/smacker/go-tree-sitter"
    "github.com/smacker/go-tree-sitter/golang"
)

func TestTreeSitterBasic(t *testing.T) {
    parser := sitter.NewParser()
    defer parser.Close()
    
    parser.SetLanguage(golang.GetLanguage())
    
    sourceCode := []byte(`
package main

func hello() string {
    return "world"
}
`)
    
    tree := parser.Parse(nil, sourceCode)
    defer tree.Close()
    
    root := tree.RootNode()
    
    fmt.Printf("Root: %s\n", root.Type())
    fmt.Printf("Children: %d\n", root.ChildCount())
    
    if root.Type() != "source_file" {
        t.Errorf("Expected source_file, got %s", root.Type())
    }
}

func TestQueryExecution(t *testing.T) {
    parser := sitter.NewParser()
    defer parser.Close()
    
    parser.SetLanguage(golang.GetLanguage())
    
    source := []byte(`
package main

func add(a, b int) int {
    return a + b
}

func subtract(a, b int) int {
    return a - b
}
`)
    
    tree := parser.Parse(nil, source)
    defer tree.Close()
    
    // Заявка за намиране на всички функции
    queryStr := `(function_declaration name: (identifier) @func_name)`
    
    query, err := sitter.NewQuery([]byte(queryStr), golang.GetLanguage())
    if err != nil {
        t.Fatal(err)
    }
    defer query.Close()
    
    cursor := sitter.NewQueryCursor()
    defer cursor.Close()
    
    cursor.Exec(query, tree.RootNode())
    
    funcCount := 0
    for {
        match, ok := cursor.NextMatch()
        if !ok {
            break
        }
        
        for _, capture := range match.Captures {
            funcCount++
            funcName := source[capture.Node.StartByte():capture.Node.EndByte()]
            fmt.Printf("Found function: %s\n", funcName)
        }
    }
    
    if funcCount != 2 {
        t.Errorf("Expected 2 functions, found %d", funcCount)
    }
}
