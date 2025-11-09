package core

import (
    "fmt"
    "os"
    "path/filepath"
    "sync"
    "time"
    
    "github.com/aaamil13/CodeIndexerMCP/internal/ai"
    "github.com/aaamil13/CodeIndexerMCP/internal/database"
    "github.com/aaamil13/CodeIndexerMCP/internal/model"
    "github.com/aaamil13/CodeIndexerMCP/internal/parsing"
    "github.com/aaamil13/CodeIndexerMCP/internal/parsing/extractors"
    "github.com/aaamil13/CodeIndexerMCP/internal/utils"
)

type Indexer struct {
    projectPath    string
    dbManager      *database.Manager
    ignoreMatcher  *utils.IgnoreMatcher
    project        *model.Project
    logger         *utils.Logger
    
    grammarManager *parsing.GrammarManager
    astProvider    *parsing.ASTProvider
    queryEngine    *parsing.QueryEngine
    extractors     map[string]Extractor
    
    watcher        *Watcher
    config         Config
    
    // AI helpers
    contextExtractor *ai.ContextExtractor
    impactAnalyzer   *ai.ImpactAnalyzer
    metricsCalc      *ai.MetricsCalculator
    snippetExtractor *ai.SnippetExtractor
    usageAnalyzer    *ai.UsageAnalyzer
    changeTracker    *ai.ChangeTracker
    depGraphBuilder  *ai.DependencyGraphBuilder
    typeValidator    *ai.TypeValidator
}

type Extractor interface {
    ExtractAll(parseResult *parsing.ParseResult, filePath string) (*model.FileSymbols, error)
}

type Config struct {
    IndexDir     string   // Directory for index data (default: .projectIndex)
    WorkerCount  int      // Number of parallel workers (default: CPU count)
    BatchSize    int      // Batch size for database operations
    ExcludePaths []string // Additional exclude patterns
    IncludeExts  map[string]bool // Extensions to include
}

func NewIndexer(projectPath string, cfg *Config) (*Indexer, error) {
    if cfg == nil {
        cfg = &Config{
            IndexDir:    ".projectIndex",
            WorkerCount: 4, // Default to 4 workers
            BatchSize:   100,
            ExcludePaths: []string{".git", "node_modules", "vendor"},
            IncludeExts: map[string]bool{
                ".go": true, ".py": true, ".ts": true, ".tsx": true,
                ".js": true, ".jsx": true, ".java": true, ".cs": true,
                ".php": true, ".rb": true, ".rs": true, ".kt": true,
                ".swift": true, ".c": true, ".cpp": true, ".cc": true,
                ".sh": true, ".sql": true, ".html": true, ".css": true,
                ".json": true, ".yaml": true, ".yml": true, ".toml": true,
                ".xml": true, ".md": true, ".rst": true,
            },
        }
    }
    
    logger := utils.NewLogger("[Indexer]")
    
    // Create index directory
    indexDir := filepath.Join(projectPath, cfg.IndexDir)
    if err := utils.EnsureDir(indexDir); err != nil {
        return nil, fmt.Errorf("failed to create index directory: %w", err)
    }
    
    // Open database
    dbPath := filepath.Join(indexDir, "index_test.db")
    dbManager, err := database.NewManager(dbPath, logger)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }
    
    grammarManager := parsing.NewGrammarManager()
    astProvider := parsing.NewASTProvider(grammarManager)
    queryEngine := parsing.NewQueryEngine(grammarManager)
    
    ignoreMatcher, err := utils.NewIgnoreMatcher(projectPath)
    if err != nil {
        return nil, fmt.Errorf("failed to create ignore matcher: %w", err)
    }
    
    indexer := &Indexer{
        projectPath:    projectPath,
        dbManager:      dbManager,
        ignoreMatcher:  ignoreMatcher,
        logger:         logger,
        config:         *cfg,
        grammarManager: grammarManager,
        astProvider:    astProvider,
        queryEngine:    queryEngine,
        extractors:     make(map[string]Extractor),
    }
    
    indexer.registerExtractors()
    
    return indexer, nil
}

func (idx *Indexer) registerExtractors() {
    idx.extractors["go"] = extractors.NewGoExtractor(idx.queryEngine)
    idx.extractors["python"] = extractors.NewPythonExtractor(idx.queryEngine)
    // TODO: Register other language extractors
}

// Initialize initializes the indexer and database (now part of NewIndexer)
func (idx *Indexer) Initialize() error {
    idx.logger.Info("Initializing indexer for project:", idx.projectPath)
    
    // Get or create project
    projectName := filepath.Base(idx.projectPath)
    project, err := idx.dbManager.GetProject(idx.projectPath)
    if err != nil {
        return fmt.Errorf("failed to get project: %w", err)
    }
    
    if project == nil {
        // Create new project
        project = &model.Project{
            Path:          idx.projectPath,
            Name:          projectName,
            LanguageStats: make(map[string]int), // Initialize empty map
            CreatedAt:     time.Now(),
            LastIndexed:   time.Time{}, // Initialize to zero time
        }
        
        if err := idx.dbManager.CreateProject(project); err != nil {
            return fmt.Errorf("failed to create project: %w", err)
        }
        idx.project = project // Update idx.project with the newly created project (which now has an ID)
        
        idx.logger.Info("Created new project:", projectName)
    } else {
        // Ensure LanguageStats is not nil if loaded from DB
        if project.LanguageStats == nil {
            project.LanguageStats = make(map[string]int)
        }
        idx.logger.Info("Loaded existing project:", projectName)
    }
    
    idx.project = project
    idx.logger.Debug("Initialized project with ID:", idx.project.ID)
    
    // Initialize AI helpers
    idx.contextExtractor = ai.NewContextExtractor(idx.dbManager)
    idx.impactAnalyzer = ai.NewImpactAnalyzer(idx.dbManager)
    idx.metricsCalc = ai.NewMetricsCalculator(idx.dbManager)
    idx.snippetExtractor = ai.NewSnippetExtractor(idx.dbManager)
    idx.usageAnalyzer = ai.NewUsageAnalyzer(idx.dbManager)
    idx.changeTracker = ai.NewChangeTracker(idx.dbManager)
    idx.depGraphBuilder = ai.NewDependencyGraphBuilder(idx.dbManager)
    idx.typeValidator = ai.NewTypeValidator(idx.dbManager)
    
    return nil
}

// Close closes the indexer and releases resources
func (idx *Indexer) Close() error {
    if idx.dbManager != nil {
        err := idx.dbManager.Close()
        idx.dbManager = nil // Set db to nil after closing
        return err
    }
    return nil
}

// IndexAll indexes all files in the project
func (idx *Indexer) IndexAll() error {
    if idx.dbManager == nil {
        return fmt.Errorf("indexer is closed")
    }
    idx.logger.Info("Starting full index of project")
    startTime := time.Now()
    
    // Scan for files
    files := make(chan string, 100)
    results := make(chan *indexResult, 100)
    
    var wg sync.WaitGroup
    
    // Start workers
    for i := 0; i < idx.config.WorkerCount; i++ {
        wg.Add(1)
        go idx.worker(files, results, &wg)
    }
    
    // File scanner
    go func() {
        defer close(files)
        filepath.Walk(idx.projectPath, func(path string, info os.FileInfo, err error) error {
            if err != nil {
                idx.logger.Errorf("Error walking file %s: %v", path, err)
                return err
            }
            
            if info.IsDir() {
                // Check if should ignore this directory
                relPath, _ := filepath.Rel(idx.projectPath, path)
                if relPath != "." && idx.ignoreMatcher.ShouldIgnore(relPath) {
                    idx.logger.Debugf("Ignoring directory: %s", relPath)
                    return filepath.SkipDir
                }
                return nil
            }
            
            // Check if should ignore this file
            relPath, _ := filepath.Rel(idx.projectPath, path)
            if idx.ignoreMatcher.ShouldIgnore(relPath) {
                idx.logger.Debugf("Ignoring file: %s", relPath)
                return nil
            }
            
            if idx.shouldIndex(path) {
                files <- path
            }
            
            return nil
        })
    }()
    
    // Collector for results
    go func() {
        wg.Wait()
        close(results)
    }()
    
    // Process results
    for result := range results {
        if result.err != nil {
            idx.logger.Errorf("Error indexing %s: %v", result.filePath, result.err)
            continue
        }
        
        if result.symbols != nil {
            if err := idx.dbManager.SaveFileSymbols(result.symbols); err != nil {
                idx.logger.Errorf("Error saving symbols from %s: %v", result.filePath, err)
            }
        }
    }
    
    // Update project stats
    idx.project.LastIndexed = time.Now()
    allFiles, err := idx.dbManager.GetAllFilesForProject(idx.project.ID)
    if err != nil {
        return fmt.Errorf("failed to get all files for language stats: %w", err)
    }
    newLanguageStats := make(map[string]int)
    for _, file := range allFiles {
        newLanguageStats[file.Language]++
    }
    idx.project.LanguageStats = newLanguageStats
    
    if err := idx.dbManager.UpdateProject(idx.project); err != nil {
        return fmt.Errorf("failed to update project: %w", err)
    }
    
    duration := time.Since(startTime)
    idx.logger.Infof("Indexing completed in %v", duration)
    
    return nil
}

type indexResult struct {
    filePath string
    symbols  *model.FileSymbols
    err      error
}

func (idx *Indexer) worker(files <-chan string, results chan<- *indexResult, wg *sync.WaitGroup) {
    defer wg.Done()
    
    for filePath := range files {
        symbols, err := idx.indexFile(filePath)
        results <- &indexResult{
            filePath: filePath,
            symbols:  symbols,
            err:      err,
        }
    }
}

func (idx *Indexer) indexFile(filePath string) (*model.FileSymbols, error) {
    // Determine language
    language := idx.detectLanguage(filePath)
    if language == "" {
        return nil, fmt.Errorf("unsupported file type: %s", filePath)
    }
    
    // Read file content
    content, err := os.ReadFile(filePath)
    if err != nil {
        return nil, err
    }
    
    // Parse
    parseResult, err := idx.astProvider.Parse(language, content)
    if err != nil {
        return nil, err
    }
    defer parseResult.Close()
    
    // Extract symbols
    extractor, exists := idx.extractors[language]
    if !exists {
        return nil, fmt.Errorf("no extractor for language: %s", language)
    }
    
    symbols, err := extractor.ExtractAll(parseResult, filePath)
    if err != nil {
        return nil, err
    }
    
    return symbols, nil
}


// detectLanguage determines the language of a file based on its extension
func (idx *Indexer) detectLanguage(filePath string) string {
    ext := filepath.Ext(filePath)
    
    // Use a map for faster lookup
    extToLang := map[string]string{
        ".go":   "go",
        ".py":   "python",
        ".ts":   "typescript",
        ".tsx":  "typescript",
        ".js":   "javascript",
        ".jsx":  "javascript",
        ".java": "java",
        ".cs":   "csharp",
        ".php":  "php",
        ".rb":   "ruby",
        ".rs":   "rust",
        ".kt":   "kotlin",
        ".swift": "swift",
        ".c":    "c",
        ".cpp":  "cpp",
        ".cc":   "cpp",
        ".sh":   "bash",
        ".sql":  "sql",
        ".html": "html",
        ".css":  "css",
        ".json": "json",
        ".yaml": "yaml",
        ".yml":  "yaml",
        ".toml": "toml",
        ".xml":  "xml",
        ".md":   "markdown",
        ".rst":  "rst",
    }
    
    if lang, ok := extToLang[ext]; ok {
        return lang
    }
    return ""
}

// shouldIndex checks if a file should be indexed based on configuration
func (idx *Indexer) shouldIndex(path string) bool {
    // Check for excluded paths
    for _, exclude := range idx.config.ExcludePaths {
        if matched, _ := filepath.Match(idx.projectPath + string(filepath.Separator) + exclude, path); matched {
            return false
        }
    }
    
    ext := filepath.Ext(path)
    if include, ok := idx.config.IncludeExts[ext]; ok {
        return include
    }
    return false // Default to not indexing if extension is not explicitly included
}

// SearchSymbols searches for symbols
func (idx *Indexer) SearchSymbols(opts model.SearchOptions) ([]*model.Symbol, error) {
	return idx.dbManager.SearchSymbols(opts)
}

// GetFileStructure returns the structure of a file
func (idx *Indexer) GetFileStructure(filePath string) (*model.ParseResult, error) {
    relPath, err := filepath.Rel(idx.projectPath, filePath)
    if err != nil {
        return nil, err
    }
    
    file, err := idx.dbManager.GetFileByPath(idx.project.ID, relPath)
    if err != nil {
        return nil, err
    }
    if file == nil {
        return nil, fmt.Errorf("file not found: %s", relPath)
    }
    
    symbols, err := idx.dbManager.GetSymbolsByFile(file.Path)
    if err != nil {
        return nil, err
    }
    
    imports, err := idx.dbManager.GetImportsByFile(file.Path)
    if err != nil {
        return nil, err
    }
    
    return &model.ParseResult{
        FilePath: filePath,
        Language: file.Language,
        Symbols:  symbols,
        Imports:  imports,
    }, nil
}

// GetProjectOverview returns project overview
func (idx *Indexer) GetProjectOverview() (*model.ProjectOverview, error) {
    totalFiles := 0
    files, err := idx.dbManager.GetAllFilesForProject(idx.project.ID)
    if err != nil {
        return nil, err
    }
    totalFiles = len(files)
    
    totalSymbols := 0
    // TODO: Implement GetTotalSymbols in database.Manager
    // For now, we'll just count symbols from files if available
    for _, file := range files {
        symbols, err := idx.dbManager.GetSymbolsByFile(file.Path)
        if err != nil {
            idx.logger.Warnf("Failed to get symbols for file %s: %v", file.Path, err)
            continue
        }
        totalSymbols += len(symbols)
    }
    
    return &model.ProjectOverview{
        Project:       idx.project,
        TotalFiles:    totalFiles,
        TotalSymbols:  totalSymbols,
        LanguageStats: idx.project.LanguageStats,
    }, nil
}

// Watch starts watching for file changes and auto-indexes
func (idx *Indexer) Watch() error {
	if idx.watcher != nil {
		return fmt.Errorf("watcher already started")
	}

	watcher, err := NewWatcher(idx)
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	idx.watcher = watcher

	if err := watcher.Start(); err != nil {
		return fmt.Errorf("failed to start watcher: %w", err)
	}

	idx.logger.Info("File watching started")
	return nil
}

// StopWatch stops the file watcher
func (idx *Indexer) StopWatch() error {
	if idx.watcher == nil {
		return nil
	}

	err := idx.watcher.Stop()
	idx.watcher = nil
	return err
}

// GetSymbolDetails returns detailed information about a symbol
func (idx *Indexer) GetSymbolDetails(symbolName string) (*model.SymbolDetails, error) {
    symbol, err := idx.dbManager.GetSymbolByName(symbolName)
    if err != nil {
        return nil, err
    }
    if symbol == nil {
        return nil, fmt.Errorf("symbol not found: %s", symbolName)
    }
    
    // file, err := idx.dbManager.GetFile(symbol.FileID) // FileID is not directly available in model.Symbol
    // if err != nil {
    // 	return nil, err
    // }
    
    references, err := idx.dbManager.GetReferencesBySymbol(symbol.ID)
    if err != nil {
        return nil, err
    }
    
    // relationships, err := idx.dbManager.GetRelationshipsForSymbol(symbol.ID) // This method does not exist in database.Manager
    // if err != nil {
    // 	return nil, err
    // }
    
    return &model.SymbolDetails{
        Symbol:        symbol,
        // File:          file,
        References:    references,
        // Relationships: relationships,
        Documentation: symbol.Documentation,
    }, nil
}

// FindReferences finds all references to a symbol
func (idx *Indexer) FindReferences(symbolName string) ([]*model.Reference, error) {
    symbol, err := idx.dbManager.GetSymbolByName(symbolName)
    if err != nil {
        return nil, err
    }
    if symbol == nil {
        return nil, fmt.Errorf("symbol not found: %s", symbolName)
    }
    
    return idx.dbManager.GetReferencesBySymbol(symbol.ID)
}

// GetDependencies gets all dependencies for a symbol
func (idx *Indexer) GetDependencies(symbolName string) ([]*model.Symbol, error) {
	return idx.depGraphBuilder.GetDependenciesFor(symbolName)
}

// GetAllFiles returns all indexed files
func (idx *Indexer) GetAllFiles() ([]*model.File, error) {
    return idx.dbManager.GetAllFilesForProject(idx.project.ID)
}

// AI Helper Methods

// GetCodeContext extracts comprehensive context for a symbol
func (idx *Indexer) GetCodeContext(symbolName string, depth int) (*model.CodeContext, error) {
	return idx.contextExtractor.ExtractContext(symbolName, depth)
}

// AnalyzeChangeImpact analyzes the impact of changing a symbol
func (idx *Indexer) AnalyzeChangeImpact(symbolName string) (*model.ChangeImpact, error) {
	return idx.impactAnalyzer.AnalyzeChangeImpact(symbolName)
}

// GetCodeMetrics calculates code quality metrics
func (idx *Indexer) GetCodeMetrics(symbolName string) (*model.CodeMetrics, error) {
    symbol, err := idx.dbManager.GetSymbolByName(symbolName)
    if err != nil {
        return nil, err
    }
    if symbol == nil {
        return nil, fmt.Errorf("symbol not found: %s", symbolName)
    }
    return idx.metricsCalc.CalculateMetrics(symbol)
}

// ExtractSmartSnippet extracts a self-contained code snippet
func (idx *Indexer) ExtractSmartSnippet(symbolName string) (*model.SmartSnippet, error) {
    symbol, err := idx.dbManager.GetSymbolByName(symbolName)
    if err != nil {
        return nil, err
    }
    if symbol == nil {
        return nil, fmt.Errorf("symbol not found: %s", symbolName)
    }
    return idx.snippetExtractor.ExtractSmartSnippet(symbol, false)
}

// GetUsageStatistics gets usage statistics for a symbol
func (idx *Indexer) GetUsageStatistics(symbolName string) (*model.SymbolUsageStats, error) {
    symbol, err := idx.dbManager.GetSymbolByName(symbolName)
    if err != nil {
        return nil, err
    }
    if symbol == nil {
        return nil, fmt.Errorf("symbol not found: %s", symbolName)
    }
    return idx.usageAnalyzer.AnalyzeUsage(symbol)
}

// SuggestRefactorings suggests refactoring opportunities
func (idx *Indexer) SuggestRefactorings(symbolName string) ([]*model.RefactoringOpportunity, error) {
	return idx.impactAnalyzer.SuggestRefactorings(symbolName)
}

// FindUnusedSymbols finds unused symbols in the project
func (idx *Indexer) FindUnusedSymbols() ([]*model.Symbol, error) {
	return idx.usageAnalyzer.FindUnusedSymbols(fmt.Sprintf("%d", idx.project.ID))
}

// FindMostUsedSymbols finds the most used symbols
func (idx *Indexer) FindMostUsedSymbols(limit int) ([]*model.SymbolUsageStats, error) {
	return idx.usageAnalyzer.FindMostUsedSymbols(fmt.Sprintf("%d", idx.project.ID), limit)
}

// Change Tracking Methods

// SimulateSymbolChange simulates a change without applying it
func (idx *Indexer) SimulateSymbolChange(symbolName string, changeType model.ChangeType, newValue string) (*model.ChangeImpact, error) {
	return idx.changeTracker.SimulateChange(symbolName, changeType, newValue)
}

// ValidateChanges validates a set of changes
func (idx *Indexer) ValidateChanges(changes []*model.Change) (*model.ValidationResult, error) {
	return idx.changeTracker.ValidateChanges(changes)
}

// GenerateAutoFixes generates automatic fixes for a change
func (idx *Indexer) es(change *model.Change) ([]*model.AutoFixSuggestion, error) {
	return idx.changeTracker.GenerateAutoFixes(change)
}

// Dependency Graph Methods

// BuildDependencyGraph builds a dependency graph for a symbol
func (idx *Indexer) BuildDependencyGraph(symbolName string, maxDepth int) (*model.DependencyGraph, error) {
	return idx.depGraphBuilder.BuildSymbolDependencyGraph(symbolName, maxDepth)
}

// GetDependents gets all symbols that depend on a symbol
func (idx *Indexer) GetDependents(symbolName string) ([]*model.Symbol, error) {
	return idx.depGraphBuilder.GetDependentsFor(symbolName)
}

// AnalyzeDependencyChain analyzes the full dependency chain
func (idx *Indexer) AnalyzeDependencyChain(symbolName string) (map[string]interface{}, error) {
	return idx.depGraphBuilder.AnalyzeDependencyChain(symbolName)
}

// Type Validation Methods

// ValidateFileTypes validates all types in a file
func (idx *Indexer) ValidateFileTypes(filePath string) (*model.TypeValidation, error) {
	return idx.typeValidator.ValidateFile(filePath)
}

// FindUndefinedUsages finds all undefined symbol usages in a file
func (idx *Indexer) FindUndefinedUsages(filePath string) ([]*model.UndefinedUsage, error) {
	return idx.typeValidator.FindUndefinedUsages(filePath)
}

// CheckMethodExists checks if a method exists on a type
func (idx *Indexer) CheckMethodExists(typeName, methodName string) (*model.MissingMethod, error) {
	return idx.typeValidator.CheckMethodExists(typeName, methodName, fmt.Sprintf("%d", idx.project.ID))
}

func (idx *Indexer) GetProject() *model.Project {
    return idx.project
}

// IndexFile indexes a single file, making it accessible publicly
func (idx *Indexer) IndexFile(filePath string) error {
    fileSymbols, err := idx.indexFile(filePath)
    if err != nil {
        return err
    }
    if fileSymbols != nil {
        return idx.dbManager.SaveFileSymbols(fileSymbols)
    }
    return nil
}

// CalculateTypeSafetyScore calculates type safety score for a file
func (idx *Indexer) CalculateTypeSafetyScore(filePath string) (*model.TypeSafetyScore, error) {
	return idx.typeValidator.CalculateTypeSafetyScore(filePath)
}
