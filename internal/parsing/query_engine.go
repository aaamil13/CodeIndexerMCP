package parsing

import (
    "fmt"
    
    sitter "github.com/smacker/go-tree-sitter"
)

type QueryCapture struct {
    Name  string
    Node  *sitter.Node
    Text  string
}

type QueryMatch struct {
    Captures []*QueryCapture
}

type QueryResult struct {
    Matches []*QueryMatch
    Source  []byte
}

type QueryEngine struct {
    grammarManager *GrammarManager
    queryCache     map[string]*sitter.Query
}

func NewQueryEngine(gm *GrammarManager) *QueryEngine {
    return &QueryEngine{
        grammarManager: gm,
        queryCache:     make(map[string]*sitter.Query),
    }
}

func (qe *QueryEngine) Execute(parseResult *ParseResult, queryString string) (*QueryResult, error) {
    grammar, err := qe.grammarManager.GetLanguage(parseResult.Language)
    if err != nil {
        return nil, err
    }
    
    // Кеширане на заявките
    cacheKey := fmt.Sprintf("%s:%s", parseResult.Language, queryString)
    query, exists := qe.queryCache[cacheKey]
    
    if !exists {
        query, err = sitter.NewQuery([]byte(queryString), grammar)
        if err != nil {
            return nil, fmt.Errorf("invalid query: %w", err)
        }
        qe.queryCache[cacheKey] = query
    }
    
    cursor := sitter.NewQueryCursor()
    defer cursor.Close()
    
    cursor.Exec(query, parseResult.RootNode)
    
    result := &QueryResult{
        Matches: make([]*QueryMatch, 0),
        Source:  parseResult.SourceCode,
    }
    
    for {
        match, ok := cursor.NextMatch()
        if !ok {
            break
        }
        
        qMatch := &QueryMatch{
            Captures: make([]*QueryCapture, 0, len(match.Captures)),
        }
        
        for _, capture := range match.Captures {
            captureName := query.CaptureNameForId(capture.Index)
            text := parseResult.SourceCode[capture.Node.StartByte():capture.Node.EndByte()]
            
            qMatch.Captures = append(qMatch.Captures, &QueryCapture{
                Name: captureName,
                Node: capture.Node,
                Text: string(text),
            })
        }
        
        result.Matches = append(result.Matches, qMatch)
    }
    
    return result, nil
}

func (qe *QueryEngine) ExecuteFromFile(parseResult *ParseResult, queryFilePath string) (*QueryResult, error) {
    // Зареждане на query от .scm файл
    // Имплементирайте четене от queries/ директория
    return nil, fmt.Errorf("not implemented")
}
