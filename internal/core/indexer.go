package core

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/aaamil13/CodeIndexerMCP/internal/ai"
	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsers/golang"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsers/python"
	"github.com/aaamil13/CodeIndexerMCP/internal/utils"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// Indexer is the main code indexer
type Indexer struct {
	projectPath      string
	db               *database.DB
	parsers          *parser.Registry
	ignoreMatcher    *utils.IgnoreMatcher
	project          *types.Project
	logger           *utils.Logger
	config           *Config
	watcher          *Watcher
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

// Config holds indexer configuration
type Config struct {
	IndexDir    string   // Directory for index data (default: .projectIndex)
	WorkerCount int      // Number of parallel workers (default: CPU count)
	BatchSize   int      // Batch size for database operations
	Exclude     []string // Additional exclude patterns
}

// NewIndexer creates a new indexer for the given project path
func NewIndexer(projectPath string, cfg *Config) (*Indexer, error) {
	if cfg == nil {
		cfg = &Config{
			IndexDir:    ".projectIndex",
			WorkerCount: runtime.NumCPU(),
			BatchSize:   100,
		}
	}

	logger := utils.NewLogger("[Indexer]")

	// Initialize parser registry
	reg := parser.NewRegistry()

	// Register built-in parsers
	if err := reg.Register(golang.NewParser()); err != nil {
		return nil, fmt.Errorf("failed to register Go parser: %w", err)
	}
	if err := reg.Register(python.NewParser()); err != nil {
		return nil, fmt.Errorf("failed to register Python parser: %w", err)
	}

	// Initialize ignore matcher
	ignoreMatcher, err := utils.NewIgnoreMatcher(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create ignore matcher: %w", err)
	}

	indexer := &Indexer{
		projectPath:   projectPath,
		parsers:       reg,
		ignoreMatcher: ignoreMatcher,
		logger:        logger,
		config:        cfg,
	}

	return indexer, nil
}

// Initialize initializes the indexer and database
func (idx *Indexer) Initialize() error {
	idx.logger.Info("Initializing indexer for project:", idx.projectPath)

	// Create index directory
	indexDir := filepath.Join(idx.projectPath, idx.config.IndexDir)
	if err := utils.EnsureDir(indexDir); err != nil {
		return fmt.Errorf("failed to create index directory: %w", err)
	}

	// Open database
	dbPath := filepath.Join(indexDir, "index.db")
	db, err := database.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	idx.db = db

	// Get or create project
	projectName := filepath.Base(idx.projectPath)
	project, err := idx.db.GetProject(idx.projectPath)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	if project == nil {
		// Create new project
		project = &types.Project{
			Path:          idx.projectPath,
			Name:          projectName,
			LanguageStats: make(map[string]int),
			CreatedAt:     time.Now(),
		}

		if err := idx.db.CreateProject(project); err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}

		idx.logger.Info("Created new project:", projectName)
	} else {
		idx.logger.Info("Loaded existing project:", projectName)
	}

	idx.project = project

	// Initialize AI helpers
	idx.contextExtractor = ai.NewContextExtractor(idx.db)
	idx.impactAnalyzer = ai.NewImpactAnalyzer(idx.db)
	idx.metricsCalc = ai.NewMetricsCalculator(idx.db)
	idx.snippetExtractor = ai.NewSnippetExtractor(idx.db)
	idx.usageAnalyzer = ai.NewUsageAnalyzer(idx.db)
	idx.changeTracker = ai.NewChangeTracker(idx.db)
	idx.depGraphBuilder = ai.NewDependencyGraphBuilder(idx.db)
	idx.typeValidator = ai.NewTypeValidator(idx.db)

	return nil
}

// Close closes the indexer and releases resources
func (idx *Indexer) Close() error {
	if idx.db != nil {
		return idx.db.Close()
	}
	return nil
}

// IndexAll indexes all files in the project
func (idx *Indexer) IndexAll() error {
	idx.logger.Info("Starting full index of project")
	startTime := time.Now()

	// Scan for files
	files, err := idx.scanFiles()
	if err != nil {
		return fmt.Errorf("failed to scan files: %w", err)
	}

	idx.logger.Infof("Found %d files to index", len(files))

	// Index files concurrently
	if err := idx.indexFiles(files); err != nil {
		return fmt.Errorf("failed to index files: %w", err)
	}

	// Update project stats
	idx.project.LastIndexed = time.Now()
	if err := idx.db.UpdateProject(idx.project); err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	duration := time.Since(startTime)
	idx.logger.Infof("Indexing completed in %v", duration)

	return nil
}

// IndexFile indexes a single file
func (idx *Indexer) IndexFile(filePath string) error {
	// Make path relative to project
	relPath, err := filepath.Rel(idx.projectPath, filePath)
	if err != nil {
		return err
	}

	// Check if should ignore
	if idx.ignoreMatcher.ShouldIgnore(relPath) {
		return nil
	}

	// Check if we can parse this file
	if !idx.parsers.CanParse(filePath) {
		return nil // Skip unsupported files silently
	}

	idx.logger.Debugf("Indexing file: %s", relPath)

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Calculate hash
	hash := utils.HashBytes(content)

	// Check if file has changed
	existingFile, err := idx.db.GetFileByPath(idx.project.ID, relPath)
	if err != nil {
		return err
	}

	if existingFile != nil && existingFile.Hash == hash {
		// File hasn't changed, skip
		idx.logger.Debugf("File unchanged, skipping: %s", relPath)
		return nil
	}

	// Parse file
	parser, err := idx.parsers.GetParserForFile(filePath)
	if err != nil {
		return err
	}

	parseResult, err := parser.Parse(content, filePath)
	if err != nil {
		idx.logger.Warnf("Failed to parse %s: %v", relPath, err)
		return nil // Don't fail on parse errors
	}

	// Count lines
	lines, _ := utils.CountLines(filePath)

	// Save to database in transaction
	err = idx.db.Transaction(func(tx *database.DB) error {
		// Save file
		file := &types.File{
			ProjectID:    idx.project.ID,
			Path:         filePath,
			RelativePath: relPath,
			Language:     parser.Language(),
			Size:         fileInfo.Size(),
			LinesOfCode:  lines,
			Hash:         hash,
			LastModified: fileInfo.ModTime(),
			LastIndexed:  time.Now(),
		}

		if err := idx.db.SaveFile(file); err != nil {
			return err
		}

		// Delete old symbols/imports for this file
		if existingFile != nil {
			idx.db.DeleteSymbolsByFile(file.ID)
			idx.db.DeleteImportsByFile(file.ID)
		}

		// Save symbols
		for _, symbol := range parseResult.Symbols {
			symbol.FileID = file.ID
			if err := idx.db.SaveSymbol(symbol); err != nil {
				return err
			}
		}

		// Save imports
		for _, imp := range parseResult.Imports {
			imp.FileID = file.ID
			if err := idx.db.SaveImport(imp); err != nil {
				return err
			}
		}

		// Save relationships
		for _, rel := range parseResult.Relationships {
			if err := idx.db.SaveRelationship(rel); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to save parse results: %w", err)
	}

	idx.logger.Debugf("Indexed file: %s (%d symbols, %d imports)",
		relPath, len(parseResult.Symbols), len(parseResult.Imports))

	return nil
}

// scanFiles scans the project directory for files
func (idx *Indexer) scanFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(idx.projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// Check if should ignore this directory
			relPath, _ := filepath.Rel(idx.projectPath, path)
			if relPath != "." && idx.ignoreMatcher.ShouldIgnore(relPath) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if should ignore this file
		relPath, _ := filepath.Rel(idx.projectPath, path)
		if idx.ignoreMatcher.ShouldIgnore(relPath) {
			return nil
		}

		// Check if we can parse this file
		if idx.parsers.CanParse(path) {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// indexFiles indexes multiple files concurrently
func (idx *Indexer) indexFiles(files []string) error {
	numWorkers := idx.config.WorkerCount
	jobs := make(chan string, len(files))
	errors := make(chan error, len(files))
	var wg sync.WaitGroup

	// Start workers
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range jobs {
				if err := idx.IndexFile(filePath); err != nil {
					errors <- fmt.Errorf("failed to index %s: %w", filePath, err)
				}
			}
		}()
	}

	// Send jobs
	for _, file := range files {
		jobs <- file
	}
	close(jobs)

	// Wait for completion
	wg.Wait()
	close(errors)

	// Collect errors
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		// Return first error (could be enhanced to return all)
		return errs[0]
	}

	return nil
}

// SearchSymbols searches for symbols
func (idx *Indexer) SearchSymbols(opts types.SearchOptions) ([]*types.Symbol, error) {
	return idx.db.SearchSymbols(opts)
}

// GetFileStructure returns the structure of a file
func (idx *Indexer) GetFileStructure(filePath string) (*types.FileStructure, error) {
	relPath, err := filepath.Rel(idx.projectPath, filePath)
	if err != nil {
		return nil, err
	}

	file, err := idx.db.GetFileByPath(idx.project.ID, relPath)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, fmt.Errorf("file not found: %s", relPath)
	}

	symbols, err := idx.db.GetSymbolsByFile(file.ID)
	if err != nil {
		return nil, err
	}

	imports, err := idx.db.GetImportsByFile(file.ID)
	if err != nil {
		return nil, err
	}

	return &types.FileStructure{
		FilePath: filePath,
		Language: file.Language,
		Symbols:  symbols,
		Imports:  imports,
	}, nil
}

// GetProjectOverview returns project overview
func (idx *Indexer) GetProjectOverview() (*types.ProjectOverview, error) {
	stats, err := idx.db.Stats()
	if err != nil {
		return nil, err
	}

	return &types.ProjectOverview{
		Project:       idx.project,
		TotalFiles:    stats["files"],
		TotalSymbols:  stats["symbols"],
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
func (idx *Indexer) GetSymbolDetails(symbolName string) (*types.SymbolDetails, error) {
	symbol, err := idx.db.GetSymbolByName(symbolName)
	if err != nil {
		return nil, err
	}
	if symbol == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbolName)
	}

	file, err := idx.db.GetFile(symbol.FileID)
	if err != nil {
		return nil, err
	}

	references, err := idx.db.GetReferencesBySymbol(symbol.ID)
	if err != nil {
		return nil, err
	}

	relationships, err := idx.db.GetRelationshipsForSymbol(symbol.ID)
	if err != nil {
		return nil, err
	}

	return &types.SymbolDetails{
		Symbol:        symbol,
		File:          file,
		References:    references,
		Relationships: relationships,
		Documentation: symbol.Documentation,
	}, nil
}

// FindReferences finds all references to a symbol
func (idx *Indexer) FindReferences(symbolName string) ([]*types.Reference, error) {
	symbol, err := idx.db.GetSymbolByName(symbolName)
	if err != nil {
		return nil, err
	}
	if symbol == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbolName)
	}

	return idx.db.GetReferencesBySymbol(symbol.ID)
}

// GetDependencies returns dependencies for a file
func (idx *Indexer) GetDependencies(filePath string) (*types.DependencyGraph, error) {
	relPath, err := filepath.Rel(idx.projectPath, filePath)
	if err != nil {
		return nil, err
	}

	file, err := idx.db.GetFileByPath(idx.project.ID, relPath)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, fmt.Errorf("file not found: %s", relPath)
	}

	imports, err := idx.db.GetImportsByFile(file.ID)
	if err != nil {
		return nil, err
	}

	// Build dependency graph
	deps := make(map[string][]string)
	deps[relPath] = []string{}

	for _, imp := range imports {
		deps[relPath] = append(deps[relPath], imp.Source)
	}

	return &types.DependencyGraph{
		Root:         relPath,
		Dependencies: deps,
		Imports:      imports,
	}, nil
}

// GetAllFiles returns all indexed files
func (idx *Indexer) GetAllFiles() ([]*types.File, error) {
	return idx.db.GetAllFilesForProject(idx.project.ID)
}

// AI Helper Methods

// GetCodeContext extracts comprehensive context for a symbol
func (idx *Indexer) GetCodeContext(symbolName string, depth int) (*types.CodeContext, error) {
	return idx.contextExtractor.ExtractContext(symbolName, depth)
}

// AnalyzeChangeImpact analyzes the impact of changing a symbol
func (idx *Indexer) AnalyzeChangeImpact(symbolName string) (*types.ChangeImpact, error) {
	return idx.impactAnalyzer.AnalyzeChangeImpact(symbolName)
}

// GetCodeMetrics calculates code quality metrics
func (idx *Indexer) GetCodeMetrics(symbolName string) (*types.CodeMetrics, error) {
	return idx.metricsCalc.CalculateMetrics(symbolName)
}

// ExtractSmartSnippet extracts a self-contained code snippet
func (idx *Indexer) ExtractSmartSnippet(symbolName string) (*types.SmartSnippet, error) {
	return idx.snippetExtractor.ExtractSmartSnippet(symbolName, false)
}

// GetUsageStatistics gets usage statistics for a symbol
func (idx *Indexer) GetUsageStatistics(symbolName string) (*types.SymbolUsageStats, error) {
	return idx.usageAnalyzer.AnalyzeUsage(symbolName)
}

// SuggestRefactorings suggests refactoring opportunities
func (idx *Indexer) SuggestRefactorings(symbolName string) ([]*types.RefactoringOpportunity, error) {
	return idx.impactAnalyzer.SuggestRefactorings(symbolName)
}

// FindUnusedSymbols finds unused symbols in the project
func (idx *Indexer) FindUnusedSymbols() ([]*types.Symbol, error) {
	return idx.usageAnalyzer.FindUnusedSymbols(idx.project.ID)
}

// FindMostUsedSymbols finds the most used symbols
func (idx *Indexer) FindMostUsedSymbols(limit int) ([]*types.SymbolUsageStats, error) {
	return idx.usageAnalyzer.FindMostUsedSymbols(idx.project.ID, limit)
}

// Change Tracking Methods

// SimulateSymbolChange simulates a change without applying it
func (idx *Indexer) SimulateSymbolChange(symbolName string, changeType types.ChangeType, newValue string) (*types.ChangeImpactResult, error) {
	return idx.changeTracker.SimulateChange(symbolName, changeType, newValue)
}

// ValidateChanges validates a set of changes
func (idx *Indexer) ValidateChanges(changes []*types.Change) (*types.ValidationResult, error) {
	return idx.changeTracker.ValidateChanges(changes)
}

// GenerateAutoFixes generates automatic fixes for a change
func (idx *Indexer) GenerateAutoFixes(change *types.Change) ([]*types.AutoFixSuggestion, error) {
	return idx.changeTracker.GenerateAutoFixes(change)
}

// Dependency Graph Methods

// BuildDependencyGraph builds a dependency graph for a symbol
func (idx *Indexer) BuildDependencyGraph(symbolName string, maxDepth int) (*types.DependencyGraph, error) {
	return idx.depGraphBuilder.BuildSymbolDependencyGraph(symbolName, maxDepth)
}

// GetDependencies gets all dependencies for a symbol
func (idx *Indexer) GetDependencies(symbolName string) ([]*types.Symbol, error) {
	return idx.depGraphBuilder.GetDependenciesFor(symbolName)
}

// GetDependents gets all symbols that depend on a symbol
func (idx *Indexer) GetDependents(symbolName string) ([]*types.Symbol, error) {
	return idx.depGraphBuilder.GetDependentsFor(symbolName)
}

// AnalyzeDependencyChain analyzes the full dependency chain
func (idx *Indexer) AnalyzeDependencyChain(symbolName string) (map[string]interface{}, error) {
	return idx.depGraphBuilder.AnalyzeDependencyChain(symbolName)
}

// Type Validation Methods

// ValidateFileTypes validates all types in a file
func (idx *Indexer) ValidateFileTypes(filePath string) (*types.TypeValidation, error) {
	file, err := idx.db.GetFileByPath(filePath, idx.project.ID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}
	return idx.typeValidator.ValidateFile(file.ID)
}

// FindUndefinedUsages finds all undefined symbol usages in a file
func (idx *Indexer) FindUndefinedUsages(filePath string) ([]*types.UndefinedUsage, error) {
	file, err := idx.db.GetFileByPath(filePath, idx.project.ID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}
	return idx.typeValidator.FindUndefinedUsages(file.ID)
}

// CheckMethodExists checks if a method exists on a type
func (idx *Indexer) CheckMethodExists(typeName, methodName string) (*types.MissingMethod, error) {
	return idx.typeValidator.CheckMethodExists(typeName, methodName, idx.project.ID)
}

// CalculateTypeSafetyScore calculates type safety score for a file
func (idx *Indexer) CalculateTypeSafetyScore(filePath string) (*types.TypeSafetyScore, error) {
	file, err := idx.db.GetFileByPath(filePath, idx.project.ID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}
	return idx.typeValidator.CalculateTypeSafetyScore(file.ID)
}
