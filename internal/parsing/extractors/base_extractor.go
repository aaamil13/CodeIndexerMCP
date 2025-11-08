package extractors

import (
    "crypto/sha256"
    "fmt"
    "time"
    
    "github.com/aaamil13/CodeIndexerMCP/internal/model"
    sitter "github.com/smacker/go-tree-sitter"
)

type BaseExtractor struct {
    Language string
}

func (be *BaseExtractor) GenerateID(kind, name, file string, pos model.Position) string {
    data := fmt.Sprintf("%s:%s:%s:%d:%d", kind, name, file, pos.Line, pos.Column)
    hash := sha256.Sum256([]byte(data))
    return fmt.Sprintf("%x", hash[:8])
}

// 💡 ПОДОБРЕНИЕ #5: Изчисляване на Content Hash
func (be *BaseExtractor) ComputeContentHash(content string) string {
    hash := sha256.Sum256([]byte(content))
    return fmt.Sprintf("%x", hash)
}

// Проверка дали символът е променен
func (be *BaseExtractor) HasContentChanged(oldHash, newContent string) bool {
    newHash := be.ComputeContentHash(newContent)
    return oldHash != newHash
}

func (be *BaseExtractor) NodeToPosition(node *sitter.Node) model.Position {
    start := node.StartPoint()
    return model.Position{
        Line:   int(start.Row) + 1,
        Column: int(start.Column) + 1,
        Byte:   int(node.StartByte()),
    }
}

func (be *BaseExtractor) NodeToRange(node *sitter.Node) model.Range {
    start := node.StartPoint()
    end := node.EndPoint()
    
    return model.Range{
        Start: model.Position{
            Line:   int(start.Row) + 1,
            Column: int(start.Column) + 1,
            Byte:   int(node.StartByte()),
        },
        End: model.Position{
            Line:   int(end.Row) + 1,
            Column: int(end.Column) + 1,
            Byte:   int(node.EndByte()),
        },
    }
}

func (be *BaseExtractor) ExtractText(node *sitter.Node, source []byte) string {
    return string(source[node.StartByte():node.EndByte()])
}

func (be *BaseExtractor) ExtractDocumentation(node *sitter.Node, source []byte) string {
    // Търси коментари преди node
    // Имплементация зависи от езика
    return ""
}

func (be *BaseExtractor) ExtractStatusFromComments(node *sitter.Node, source []byte) model.DevelopmentStatus {
    // Търси специални коментари като:
    // // STATUS: planned
    // // STATUS: in_progress
    // Имплементация...
    return ""
}

func (be *BaseExtractor) ExtractPriorityFromComments(node *sitter.Node, source []byte) int {
    // Търси коментари като: // PRIORITY: 5
    return 0
}
