package model

import "time"

type FileSymbols struct {
    FilePath   string       `json:"file_path"`
    Language   string       `json:"language"`
    Functions  []*Function  `json:"functions,omitempty"`
    Methods    []*Method    `json:"methods,omitempty"`
    Classes    []*Class     `json:"classes,omitempty"`
    Interfaces []*Interface `json:"interfaces,omitempty"`
    Variables  []*Variable  `json:"variables,omitempty"`
    Imports    []*Import    `json:"imports,omitempty"`
    ParseTime  time.Time    `json:"parse_time"`
    ParseError string       `json:"parse_error,omitempty"`
}

func (fs *FileSymbols) AllSymbols() []CodeElement {
    symbols := make([]CodeElement, 0)
    
    for _, f := range fs.Functions {
        symbols = append(symbols, f)
    }
    for _, m := range fs.Methods {
        symbols = append(symbols, m)
    }
    for _, c := range fs.Classes {
        symbols = append(symbols, c)
    }
    for _, i := range fs.Interfaces {
        symbols = append(symbols, i)
    }
    for _, v := range fs.Variables {
        symbols = append(symbols, v)
    }
    
    return symbols
}
