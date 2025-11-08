package parsing

import (
    "fmt"
    "sync"
    
    sitter "github.com/smacker/go-tree-sitter"
    "github.com/smacker/go-tree-sitter/c"
    "github.com/smacker/go-tree-sitter/cpp"
    "github.com/smacker/go-tree-sitter/golang"
    "github.com/smacker/go-tree-sitter/java"
    "github.com/smacker/go-tree-sitter/javascript"
    "github.com/smacker/go-tree-sitter/python"
    "github.com/smacker/go-tree-sitter/rust"
    "github.com/smacker/go-tree-sitter/typescript/typescript"
)

type GrammarManager struct {
    grammars map[string]*sitter.Language
    mu       sync.RWMutex
}

func NewGrammarManager() *GrammarManager {
    gm := &GrammarManager{
        grammars: make(map[string]*sitter.Language),
    }
    gm.initBuiltinGrammars()
    return gm
}

func (gm *GrammarManager) initBuiltinGrammars() {
    // Built-in граматики от go-tree-sitter
    gm.grammars["go"] = golang.GetLanguage()
    gm.grammars["python"] = python.GetLanguage()
    gm.grammars["typescript"] = typescript.GetLanguage()
    gm.grammars["javascript"] = javascript.GetLanguage()
    gm.grammars["java"] = java.GetLanguage()
    gm.grammars["c"] = c.GetLanguage()
    gm.grammars["cpp"] = cpp.GetLanguage()
    gm.grammars["rust"] = rust.GetLanguage()
}

func (gm *GrammarManager) GetLanguage(lang string) (*sitter.Language, error) {
    gm.mu.RLock()
    defer gm.mu.RUnlock()
    
    grammar, exists := gm.grammars[lang]
    if !exists {
        return nil, fmt.Errorf("language not supported: %s", lang)
    }
    return grammar, nil
}

func (gm *GrammarManager) GetSupportedLanguages() []string {
    gm.mu.RLock()
    defer gm.mu.RUnlock()
    
    langs := make([]string, 0, len(gm.grammars))
    for lang := range gm.grammars {
        langs = append(langs, lang)
    }
    return langs
}
