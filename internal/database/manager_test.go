package database_test

import (
	"database/sql"
	"encoding/json" // Added json import
	"path/filepath"
	"testing"
	"time"

	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*database.Manager, string) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	
	dbManager, err := database.NewManager(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := dbManager.Close()
		assert.NoError(t, err, "Failed to close database manager")
	})
	return dbManager, dbPath
}

func TestManager_CreateAndGetProject(t *testing.T) {
	db, _ := setupTestDB(t)

	projectPath := "/test/project"
	projectName := "test-project"

	// Create project
	project := &model.Project{
		Path:          projectPath,
		Name:          projectName,
		LanguageStats: map[string]int{"go": 10, "python": 5},
		CreatedAt:     time.Now().Add(-24 * time.Hour),
		LastIndexed:   time.Now().Add(-1 * time.Hour),
	}
	err := db.SaveProject(project) // Changed to SaveProject
	require.NoError(t, err)
	assert.NotZero(t, project.ID)

	// Get project
	fetchedProject, err := db.GetProjectByPath(projectPath) // Changed to GetProjectByPath
	require.NoError(t, err)
	require.NotNil(t, fetchedProject)

	assert.Equal(t, project.ID, fetchedProject.ID)
	assert.Equal(t, project.Path, fetchedProject.Path)
	assert.Equal(t, project.Name, fetchedProject.Name)
	assert.Equal(t, project.LanguageStats, fetchedProject.LanguageStats)
	// Compare time with some tolerance due to possible precision issues in SQLite
	assert.WithinDuration(t, project.CreatedAt, fetchedProject.CreatedAt, time.Second)
	assert.WithinDuration(t, project.LastIndexed, fetchedProject.LastIndexed, time.Second)
}

func TestManager_UpdateProject(t *testing.T) {
	db, _ := setupTestDB(t)

	projectPath := "/test/project"
	projectName := "test-project"

	project := &model.Project{
		Path:          projectPath,
		Name:          projectName,
		LanguageStats: map[string]int{"go": 10},
		CreatedAt:     time.Now().Add(-24 * time.Hour),
		LastIndexed:   time.Now().Add(-1 * time.Hour),
	}
	err := db.SaveProject(project) // Changed to SaveProject
	require.NoError(t, err)

	// Update project
	project.Name = "updated-project"
	project.LanguageStats["python"] = 20
	project.LastIndexed = time.Now()
	// No direct UpdateProject, SaveProject handles updates on conflict
	err = db.SaveProject(project) // Changed to SaveProject
	require.NoError(t, err)

	fetchedProject, err := db.GetProjectByPath(projectPath) // Changed to GetProjectByPath
	require.NoError(t, err)
	require.NotNil(t, fetchedProject)

	assert.Equal(t, project.Name, fetchedProject.Name)
	assert.Equal(t, project.LanguageStats, fetchedProject.LanguageStats)
	assert.WithinDuration(t, project.LastIndexed, fetchedProject.LastIndexed, time.Second)
}

func TestManager_SaveFileAndGet(t *testing.T) {
	db, _ := setupTestDB(t)

	projectPath := "/test/project"
	projectName := "test-project"
	project := &model.Project{Path: projectPath, Name: projectName}
	db.SaveProject(project) // Changed to SaveProject

	filePath := filepath.Join(projectPath, "main.go")
	relPath := "main.go"
	file := &model.File{
		ProjectID:    project.ID,
		Path:         filePath,
		RelativePath: relPath,
		Language:     "go",
		Size:         1024,
		LinesOfCode:  50,
		Hash:         "abcdef123",
		LastModified: time.Now().Add(-time.Hour),
		LastIndexed:  time.Now(),
	}

	err := db.SaveFile(file)
	require.NoError(t, err)
	assert.NotZero(t, file.ID)

	fetchedFile, err := db.GetFileByPath(project.ID, relPath)
	require.NoError(t, err)
	require.NotNil(t, fetchedFile)

	assert.Equal(t, file.ID, fetchedFile.ID)
	assert.Equal(t, file.Path, fetchedFile.Path)
	assert.Equal(t, file.RelativePath, fetchedFile.RelativePath)
	assert.Equal(t, file.Language, fetchedFile.Language)
	assert.Equal(t, file.Size, fetchedFile.Size)
	assert.Equal(t, file.LinesOfCode, fetchedFile.LinesOfCode)
	assert.Equal(t, file.Hash, fetchedFile.Hash)
	assert.WithinDuration(t, file.LastModified, fetchedFile.LastModified, time.Second)
	assert.WithinDuration(t, file.LastIndexed, fetchedFile.LastIndexed, time.Second)
}

func TestManager_SaveSymbolAndGet(t *testing.T) {
	db, _ := setupTestDB(t)

	// Create project and file first
	project := &model.Project{Path: "/test/project", Name: "test-project"}
	db.SaveProject(project) // Changed to SaveProject
	file := &model.File{ProjectID: project.ID, Path: "/test/project/main.go", RelativePath: "main.go", Language: "go"}
	db.SaveFile(file)

	symbol := &model.Symbol{
		FileID:        file.ID, // Added FileID
		Name:          "MyFunc",
		Kind:          model.SymbolKindFunction,
		FilePath:      file.Path, // Changed from File to FilePath
		Language:      "go",
		Signature:     "func MyFunc()",
		Documentation: "A test function",
		Visibility:    model.VisibilityPublic,
		LineNumber:    1, // Added
		ColumnNumber:  1, // Added
		EndLineNumber: 5, // Added
		EndColumnNumber: 1, // Added
		Parent:        "", // Added
		ContentHash:   "hash123",
		Status:        model.StatusCompleted,
		Priority:      5, 
		AssignedAgent: "AI",
		CreatedAt:     time.Now().Add(-time.Minute),
		UpdatedAt:     time.Now(),
		Metadata:      map[string]string{"key": "value"},
	}

	err := db.Transaction(func(tx *sql.Tx) error {
		return db.SaveSymbolTx(tx, symbol)
	})
	require.NoError(t, err)

	fetchedSymbol, err := db.GetSymbolByID(symbol.ID) // Changed to GetSymbolByID
	require.NoError(t, err)
	require.NotNil(t, fetchedSymbol)

	assert.NotZero(t, symbol.ID) // ID is now auto-increment
	assert.Equal(t, file.ID, fetchedSymbol.FileID)
	assert.Equal(t, symbol.Name, fetchedSymbol.Name)
	assert.Equal(t, symbol.Kind, fetchedSymbol.Kind)
	assert.Equal(t, symbol.FilePath, fetchedSymbol.FilePath) // Changed from File to FilePath
	assert.Equal(t, symbol.Language, fetchedSymbol.Language)
	assert.Equal(t, symbol.Signature, fetchedSymbol.Signature)
	assert.Equal(t, symbol.Documentation, fetchedSymbol.Documentation)
	assert.Equal(t, symbol.Visibility, fetchedSymbol.Visibility)
	assert.Equal(t, symbol.LineNumber, fetchedSymbol.LineNumber) // Added
	assert.Equal(t, symbol.ColumnNumber, fetchedSymbol.ColumnNumber) // Added
	assert.Equal(t, symbol.EndLineNumber, fetchedSymbol.EndLineNumber) // Added
	assert.Equal(t, symbol.EndColumnNumber, fetchedSymbol.EndColumnNumber) // Added
	assert.Equal(t, symbol.Parent, fetchedSymbol.Parent) // Added
	assert.Equal(t, symbol.ContentHash, fetchedSymbol.ContentHash)
	assert.Equal(t, symbol.Status, fetchedSymbol.Status)
	assert.Equal(t, symbol.Priority, fetchedSymbol.Priority)
	assert.Equal(t, symbol.AssignedAgent, fetchedSymbol.AssignedAgent)
	assert.WithinDuration(t, symbol.CreatedAt, fetchedSymbol.CreatedAt, time.Second)
	assert.WithinDuration(t, symbol.UpdatedAt, fetchedSymbol.UpdatedAt, time.Second)
	assert.Equal(t, symbol.Metadata, fetchedSymbol.Metadata)
}

func TestManager_SaveFunctionAndGet(t *testing.T) {
	db, _ := setupTestDB(t)

	project := &model.Project{Path: "/test/project", Name: "test-project"}
	db.SaveProject(project) // Changed to SaveProject
	file := &model.File{ProjectID: project.ID, Path: "/test/project/main.go", RelativePath: "main.go", Language: "go"}
	db.SaveFile(file)

	function := &model.Function{
		Symbol: model.Symbol{
			FileID:        file.ID, // Added FileID
			Name:        "Add",
			Kind:        model.SymbolKindFunction,
			FilePath:    file.Path, // Changed from File to FilePath
			Language:    "go",
			Signature:   "(a int, b int) int",
			LineNumber:    10, // Added
			ColumnNumber:  1, // Added
			EndLineNumber: 15, // Added
			EndColumnNumber: 1, // Added
			ContentHash: "funcHash1",
		},
		Parameters: []model.Parameter{
			{Name: "a", Type: "int"},
			{Name: "b", Type: "int"},
		},
		ReturnType: "int",
		Body:       "return a + b",
	}

	err := db.Transaction(func(tx *sql.Tx) error {
		return db.SaveSymbolTx(tx, &function.Symbol) // Changed to SaveSymbolTx
	})
	require.NoError(t, err)

	// Verify symbol is saved
	fetchedSymbol, err := db.GetSymbolByID(function.Symbol.ID) // Changed to GetSymbolByID
	require.NoError(t, err)
	require.NotNil(t, fetchedSymbol)
	assert.Equal(t, function.Name, fetchedSymbol.Name)

	// TODO: Add actual GetFunctionDetails/GetParameters to Manager to fully test
}

func TestManager_SaveClassAndGet(t *testing.T) {
	db, _ := setupTestDB(t)

	project := &model.Project{Path: "/test/project", Name: "test-project"}
	db.SaveProject(project) // Changed to SaveProject
	file := &model.File{ProjectID: project.ID, Path: "/test/project/main.go", RelativePath: "main.go", Language: "go"}
	db.SaveFile(file)

	class := &model.Class{
		Symbol: model.Symbol{
			FileID:        file.ID, // Added FileID
			Name:        "MyClass",
			Kind:        model.SymbolKindClass,
			FilePath:    file.Path, // Changed from File to FilePath
			Language:    "go",
			Signature:   "type MyClass struct",
			LineNumber:    20, // Added
			ColumnNumber:  1, // Added
			EndLineNumber: 30, // Added
			EndColumnNumber: 1, // Added
			ContentHash: "classHash1",
		},
		IsAbstract: false,
		Fields: []model.Field{
			{Name: "Field1", Type: "string", Visibility: string(model.VisibilityPublic)},
		},
	}

	err := db.Transaction(func(tx *sql.Tx) error {
		return db.SaveSymbolTx(tx, &class.Symbol) // Changed to SaveSymbolTx
	})
	require.NoError(t, err)

	// Verify symbol is saved
	fetchedSymbol, err := db.GetSymbolByID(class.Symbol.ID) // Changed to GetSymbolByID
	require.NoError(t, err)
	require.NotNil(t, fetchedSymbol)
	assert.Equal(t, class.Name, fetchedSymbol.Name)

	// TODO: Add actual GetClassDetails/GetFields to Manager to fully test
}

func TestManager_SaveMethod(t *testing.T) {
	db, _ := setupTestDB(t)

	project := &model.Project{Path: "/test/project", Name: "test-project"}
	db.SaveProject(project) // Changed to SaveProject
	file := &model.File{ProjectID: project.ID, Path: "/test/project/main.go", RelativePath: "main.go", Language: "go"}
	db.SaveFile(file)

	method := &model.Method{
		Function: model.Function{
			Symbol: model.Symbol{
				FileID:        file.ID, // Added FileID
				Name:        "MyMethod",
				Kind:        model.SymbolKindMethod,
				FilePath:    file.Path, // Changed from File to FilePath
				Language:    "go",
				Signature:   "(c *MyClass) MyMethod()",
				LineNumber:    35, // Added
				ColumnNumber:  1, // Added
				EndLineNumber: 40, // Added
				EndColumnNumber: 1, // Added
				ContentHash: "methodHash1",
			},
			Parameters: []model.Parameter{
				{Name: "param1", Type: "string"},
			},
			ReturnType: "error",
			Body:       "return nil",
		},
		ReceiverType: "*MyClass",
		IsStatic:     false,
	}

	err := db.Transaction(func(tx *sql.Tx) error {
		return db.SaveSymbolTx(tx, &method.Symbol) // Changed to SaveSymbolTx
	})
	require.NoError(t, err)

	// Verify symbol is saved
	fetchedSymbol, err := db.GetSymbolByID(method.Symbol.ID) // Changed to GetSymbolByID
	require.NoError(t, err)
	require.NotNil(t, fetchedSymbol)
	assert.Equal(t, method.Name, fetchedSymbol.Name)
}

func TestManager_HasSymbolChanged(t *testing.T) {
	db, _ := setupTestDB(t)

	project := &model.Project{Path: "/test/project", Name: "test-project"}
	db.SaveProject(project)
	file := &model.File{ProjectID: project.ID, Path: "/test/project/main.go", RelativePath: "main.go", Language: "go"}
	db.SaveFile(file)

	initialHash := "initialHash"
	updatedHash := "updatedHash"

	symbol := &model.Symbol{
		FileID:      file.ID,
		Name:        "TestFunc",
		Kind:        model.SymbolKindFunction,
		FilePath:    file.Path,
		Language:    "go",
		ContentHash: initialHash,
		LineNumber:    1,
		ColumnNumber:  1,
		EndLineNumber: 5,
		EndColumnNumber: 1,
	}

	// Save initial symbol
	err := db.Transaction(func(tx *sql.Tx) error {
		return db.SaveSymbolTx(tx, symbol)
	})
	require.NoError(t, err)

	// Same hash, should not be changed
	changed, err := db.HasSymbolChanged(symbol.ID, initialHash)
	require.NoError(t, err)
	assert.False(t, changed, "Expected symbol not to be changed with same hash")

	// Different hash, should be changed
	changed, err = db.HasSymbolChanged(symbol.ID, updatedHash)
	require.NoError(t, err)
	assert.True(t, changed, "Expected symbol to be changed with different hash")
}

func TestManager_SaveSymbolIfChanged(t *testing.T) {
	db, _ := setupTestDB(t)

	project := &model.Project{Path: "/test/project", Name: "test-project"}
	db.SaveProject(project) // Changed to SaveProject
	file := &model.File{ProjectID: project.ID, Path: "/test/project/main.go", RelativePath: "main.go", Language: "go"}
	db.SaveFile(file)

	hashV1 := "hashV1"
	hashV2 := "hashV2"

	// New symbol
	symbolV1 := &model.Symbol{
		FileID:      file.ID, // Added FileID
		Name:        "FuncV1",
		Kind:        model.SymbolKindFunction,
		FilePath:    file.Path, // Changed to FilePath
		Language:    "go",
		ContentHash: hashV1,
		LineNumber:    1, // Added
		ColumnNumber:  1, // Added
		EndLineNumber: 5, // Added
		EndColumnNumber: 1, // Added
	}
	saved, err := db.SaveSymbolIfChanged(symbolV1)
	require.NoError(t, err)
	assert.True(t, saved, "Expected symbol to be saved initially")

	// No change, should not save
	saved, err = db.SaveSymbolIfChanged(symbolV1)
	require.NoError(t, err)
	assert.False(t, saved, "Expected symbol not to be saved if hash is same")

	// Change hash, should save
	symbolV2 := &model.Symbol{
		FileID:      file.ID, // Added FileID
		Name:        "FuncV2", // Name change also
		Kind:        model.SymbolKindFunction,
		FilePath:    file.Path, // Changed to FilePath
		Language:    "go",
		ContentHash: hashV2,
		LineNumber:    1, // Added
		ColumnNumber:  1, // Added
		EndLineNumber: 5, // Added
		EndColumnNumber: 1, // Added
	}
	saved, err = db.SaveSymbolIfChanged(symbolV2)
	require.NoError(t, err)
	assert.True(t, saved, "Expected symbol to be saved if hash is different")

	fetchedSymbol, err := db.GetSymbolByID(symbolV2.ID) // Changed to GetSymbolByID
	require.NoError(t, err)
	assert.NotNil(t, fetchedSymbol)
	assert.Equal(t, "FuncV2", fetchedSymbol.Name) // Verify name update
	assert.Equal(t, hashV2, fetchedSymbol.ContentHash)
}

func TestManager_SaveFileSymbols(t *testing.T) {
	db, _ := setupTestDB(t)

	project := &model.Project{Path: "/test/project", Name: "test-project"}
	db.SaveProject(project) // Changed to SaveProject
	file := &model.File{ProjectID: project.ID, Path: "/test/project/main.go", RelativePath: "main.go", Language: "go"}
	db.SaveFile(file)

	func1 := &model.Symbol{ // Changed from model.Function to model.Symbol
		FileID: file.ID, Name: "F1", Kind: model.SymbolKindFunction, FilePath: file.Path, Language: "go", ContentHash: "h1", LineNumber: 1, ColumnNumber: 1, EndLineNumber: 5, EndColumnNumber: 1,
	}
	class1 := &model.Symbol{ // Changed from model.Class to model.Symbol
		FileID: file.ID, Name: "C1", Kind: model.SymbolKindClass, FilePath: file.Path, Language: "go", ContentHash: "h2", LineNumber: 1, ColumnNumber: 1, EndLineNumber: 5, EndColumnNumber: 1,
	}
	method1 := &model.Symbol{ // Changed from model.Method to model.Symbol
		FileID: file.ID, Name: "M1", Kind: model.SymbolKindMethod, FilePath: file.Path, Language: "go", ContentHash: "h3", LineNumber: 1, ColumnNumber: 1, EndLineNumber: 5, EndColumnNumber: 1,
	}

	symbolsJSON, err := json.Marshal([]*model.Symbol{func1, class1, method1}) // Marshal symbols to JSON
	require.NoError(t, err)

	fileSymbols := &model.FileSymbols{
		FileID: file.ID,
		SymbolsJSON: symbolsJSON,
	}

	err = db.Transaction(func(tx *sql.Tx) error {
		return db.SaveFileSymbolsTx(tx, fileSymbols)
	})
	require.NoError(t, err)

	// Verify symbols are saved
	fetchedFileSymbols, err := db.GetFileSymbolsByFileID(file.ID)
	require.NoError(t, err)
	assert.NotNil(t, fetchedFileSymbols)

	var fetchedSymbols []*model.Symbol
	err = json.Unmarshal(fetchedFileSymbols.SymbolsJSON, &fetchedSymbols)
	require.NoError(t, err)
	assert.Len(t, fetchedSymbols, 3)

	// Update symbols and save again
	func2 := &model.Symbol{ // Changed from model.Function to model.Symbol
		FileID: file.ID, Name: "F2", Kind: model.SymbolKindFunction, FilePath: file.Path, Language: "go", ContentHash: "h4", LineNumber: 1, ColumnNumber: 1, EndLineNumber: 5, EndColumnNumber: 1,
	}
	
	symbolsJSON, err = json.Marshal([]*model.Symbol{func2}) // Marshal updated symbols to JSON
	require.NoError(t, err)
	fileSymbols.SymbolsJSON = symbolsJSON

	err = db.Transaction(func(tx *sql.Tx) error {
		return db.SaveFileSymbolsTx(tx, fileSymbols)
	})
	require.NoError(t, err)

	fetchedFileSymbols, err = db.GetFileSymbolsByFileID(file.ID)
	require.NoError(t, err)
	assert.NotNil(t, fetchedFileSymbols)

	err = json.Unmarshal(fetchedFileSymbols.SymbolsJSON, &fetchedSymbols)
	require.NoError(t, err)
	assert.Len(t, fetchedSymbols, 1) // Only func2 should remain
	assert.Equal(t, "F2", fetchedSymbols[0].Name)
}

func TestManager_AITaskMethods(t *testing.T) {
	db, _ := setupTestDB(t)

	// Create a project
	project := &model.Project{Path: "/test/project", Name: "test-project"}
	db.SaveProject(project) // Changed to SaveProject
	file := &model.File{ProjectID: project.ID, Path: "/test/project/main.go", RelativePath: "main.go", Language: "go"}
	db.SaveFile(file)

	// Create a symbol
	symbol := &model.Symbol{
		FileID:   file.ID, // Added FileID
		Name:     "AISymbol",
		Kind:     model.SymbolKindFunction,
		FilePath: file.Path, // Changed to FilePath
		Language: "go",
		LineNumber:    1, // Added
		ColumnNumber:  1, // Added
		EndLineNumber: 5, // Added
		EndColumnNumber: 1, // Added
		Status:   model.StatusPlanned,
	}
	err := db.Transaction(func(tx *sql.Tx) error {
		return db.SaveSymbolTx(tx, symbol)
	})
	require.NoError(t, err)

	fetchedSymbol, err := db.GetSymbolByID(symbol.ID) // Changed to GetSymbolByID
	require.NoError(t, err)
	assert.Equal(t, model.StatusPlanned, fetchedSymbol.Status) // Should be Planned not InProgress for now
}
