package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

func setupTestDB(t *testing.T) (*DB, string) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	return db, dbPath
}

func TestCreateProject(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	project := &types.Project{
		Name: "test-project",
		Path: "/test/path",
	}

	err := db.CreateProject(project)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	if project.ID == 0 {
		t.Error("Expected project ID to be set")
	}

	// Verify project was created
	retrieved, err := db.GetProject(project.Path)
	if err != nil {
		t.Fatalf("GetProject failed: %v", err)
	}

	if retrieved.Name != project.Name {
		t.Errorf("Expected name %s, got %s", project.Name, retrieved.Name)
	}
	if retrieved.Path != project.Path {
		t.Errorf("Expected path %s, got %s", project.Path, retrieved.Path)
	}
}

func TestCreateFile(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	project := &types.Project{Name: "test", Path: "/test"}
	db.CreateProject(project)

	file := &types.File{
		ProjectID: project.ID,
		Path:      "/test/file.go",
		Language:  "go",
		Size:      1024,
	}

	err := db.SaveFile(file)
	if err != nil {
		t.Fatalf("SaveFile failed: %v", err)
	}

	if file.ID == 0 {
		t.Error("Expected file ID to be set")
	}

	// Verify file was created
	retrieved, err := db.GetFile(file.ID)
	if err != nil {
		t.Fatalf("GetFile failed: %v", err)
	}

	if retrieved.Path != file.Path {
		t.Errorf("Expected path %s, got %s", file.Path, retrieved.Path)
	}
	if retrieved.Language != file.Language {
		t.Errorf("Expected language %s, got %s", file.Language, retrieved.Language)
	}
}

func TestCreateSymbol(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	project := &types.Project{Name: "test", Path: "/test"}
	db.CreateProject(project)

	file := &types.File{ProjectID: project.ID, Path: "/test/file.go", Language: "go"}
	db.SaveFile(file)

	symbol := &types.Symbol{
		FileID:        file.ID,
		Name:          "TestFunc",
		Type:          types.SymbolTypeFunction,
		Signature:     "func TestFunc() error",
		StartLine:     10,
		EndLine:       20,
		Visibility:    types.VisibilityPublic,
		IsExported:    true,
		Documentation: "TestFunc does testing",
	}

	err := db.SaveSymbol(symbol)
	if err != nil {
		t.Fatalf("SaveSymbol failed: %v", err)
	}

	if symbol.ID == 0 {
		t.Error("Expected symbol ID to be set")
	}

	// Verify symbol was created
	retrieved, err := db.GetSymbol(symbol.ID)
	if err != nil {
		t.Fatalf("GetSymbol failed: %v", err)
	}

	if retrieved.Name != symbol.Name {
		t.Errorf("Expected name %s, got %s", symbol.Name, retrieved.Name)
	}
	if retrieved.Type != symbol.Type {
		t.Errorf("Expected type %s, got %s", symbol.Type, retrieved.Type)
	}
	if retrieved.Signature != symbol.Signature {
		t.Errorf("Expected signature %s, got %s", symbol.Signature, retrieved.Signature)
	}
}

func TestSearchSymbols(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	project := &types.Project{Name: "test", Path: "/test"}
	db.CreateProject(project)

	file := &types.File{ProjectID: project.ID, Path: "/test/file.go", Language: "go"}
	db.SaveFile(file)

	// Create multiple symbols
	symbols := []*types.Symbol{
		{FileID: file.ID, Name: "TestFunc", Type: types.SymbolTypeFunction},
		{FileID: file.ID, Name: "TestStruct", Type: types.SymbolTypeStruct},
		{FileID: file.ID, Name: "TestMethod", Type: types.SymbolTypeMethod},
	}

	for _, sym := range symbols {
		db.SaveSymbol(sym)
	}

	// Search for symbols
	results, err := db.SearchSymbols(types.SearchOptions{
		Query:     "Test",
		ProjectID: project.ID,
	})
	if err != nil {
		t.Fatalf("SearchSymbols failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// Search with type filter
	funcType := types.SymbolTypeFunction
	results, err = db.SearchSymbols(types.SearchOptions{
		Query:     "Test",
		Type:      &funcType,
		ProjectID: project.ID,
	})
	if err != nil {
		t.Fatalf("SearchSymbols with type filter failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result with function filter, got %d", len(results))
	}
	if results[0].Name != "TestFunc" {
		t.Errorf("Expected TestFunc, got %s", results[0].Name)
	}
}

func TestCreateImport(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	project := &types.Project{Name: "test", Path: "/test"}
	db.CreateProject(project)

	file := &types.File{ProjectID: project.ID, Path: "/test/file.go", Language: "go"}
	db.SaveFile(file)

	imp := &types.Import{
		FileID: file.ID,
		Source: "fmt",
	}

	err := db.SaveImport(imp)
	if err != nil {
		t.Fatalf("SaveImport failed: %v", err)
	}

	// Verify import was created
	imports, err := db.GetImportsByFile(file.ID)
	if err != nil {
		t.Fatalf("GetImportsByFile failed: %v", err)
	}

	if len(imports) != 1 {
		t.Fatalf("Expected 1 import, got %d", len(imports))
	}

	if imports[0].Source != imp.Source {
		t.Errorf("Expected import source %s, got %s", imp.Source, imports[0].Source)
	}
}

func TestCreateRelationship(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	project := &types.Project{Name: "test", Path: "/test"}
	db.CreateProject(project)

	file := &types.File{ProjectID: project.ID, Path: "/test/file.go", Language: "go"}
	db.SaveFile(file)

	parent := &types.Symbol{FileID: file.ID, Name: "Parent", Type: types.SymbolTypeClass}
	db.SaveSymbol(parent)

	child := &types.Symbol{FileID: file.ID, Name: "Child", Type: types.SymbolTypeClass}
	db.SaveSymbol(child)

	rel := &types.Relationship{
		FromSymbolID: child.ID,
		ToSymbolID:   parent.ID,
		Type:         types.RelationshipExtends, // Changed to RelationshipExtends
	}

	err := db.SaveRelationship(rel)
	if err != nil {
		t.Fatalf("SaveRelationship failed: %v", err)
	}

	// Verify relationship was created
	relationships, err := db.GetRelationshipsForSymbol(child.ID)
	if err != nil {
		t.Fatalf("GetRelationshipsForSymbol failed: %v", err)
	}

	if len(relationships) != 1 {
		t.Fatalf("Expected 1 relationship, got %d", len(relationships))
	}

	if relationships[0].Type != rel.Type {
		t.Errorf("Expected type %s, got %s", rel.Type, relationships[0].Type)
	}
}

func TestCreateReference(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	project := &types.Project{Name: "test", Path: "/test"}
	db.CreateProject(project)

	file := &types.File{ProjectID: project.ID, Path: "/test/file.go", Language: "go"}
	db.SaveFile(file)

	symbol := &types.Symbol{FileID: file.ID, Name: "MyFunc", Type: types.SymbolTypeFunction}
	db.SaveSymbol(symbol)

	ref := &types.Reference{
		SymbolID:      symbol.ID,
		FileID:        file.ID,
		LineNumber:    25,
		ColumnNumber:  10,
		ReferenceType: "call",
	}

	err := db.SaveReference(ref)
	if err != nil {
		t.Fatalf("SaveReference failed: %v", err)
	}

	// Verify reference was created
	references, err := db.GetReferencesBySymbol(symbol.ID)
	if err != nil {
		t.Fatalf("GetReferencesBySymbol failed: %v", err)
	}

	if len(references) != 1 {
		t.Fatalf("Expected 1 reference, got %d", len(references))
	}

	if references[0].LineNumber != ref.LineNumber {
		t.Errorf("Expected line %d, got %d", ref.LineNumber, references[0].LineNumber)
	}
}

func TestGetSymbolByName(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	project := &types.Project{Name: "test", Path: "/test"}
	db.CreateProject(project)

	file := &types.File{ProjectID: project.ID, Path: "/test/file.go", Language: "go"}
	db.SaveFile(file)

	symbol := &types.Symbol{
		FileID: file.ID,
		Name:   "UniqueFunc",
		Type:   types.SymbolTypeFunction,
	}
	db.SaveSymbol(symbol)

	// Get by name
	retrieved, err := db.GetSymbolByName("UniqueFunc")
	if err != nil {
		t.Fatalf("GetSymbolByName failed: %v", err)
	}

	if retrieved.Name != symbol.Name {
		t.Errorf("Expected name %s, got %s", symbol.Name, retrieved.Name)
	}
}

func TestDeleteFile(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	project := &types.Project{Name: "test", Path: "/test"}
	db.CreateProject(project)

	file := &types.File{ProjectID: project.ID, Path: "/test/file.go", Language: "go"}
	db.SaveFile(file)

	// Delete file
	err := db.DeleteFile(file.ID)
	if err != nil {
		t.Fatalf("DeleteFile failed: %v", err)
	}

	// Verify file was deleted
	_, err = db.GetFile(file.ID)
	if err == nil {
		t.Error("Expected error when getting deleted file")
	}
}

func TestUpdateSymbol(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	project := &types.Project{Name: "test", Path: "/test"}
	db.CreateProject(project)

	file := &types.File{ProjectID: project.ID, Path: "/test/file.go", Language: "go"}
	db.SaveFile(file)

	symbol := &types.Symbol{
		FileID: file.ID,
		Name:   "OldName",
		Type:   types.SymbolTypeFunction,
	}
	db.SaveSymbol(symbol)

	// Update symbol
	symbol.Name = "NewName"
	symbol.Documentation = "Updated docs"
	err := db.SaveSymbol(symbol) // SaveSymbol also handles updates
	if err != nil {
		t.Fatalf("SaveSymbol failed: %v", err)
	}

	// Verify update
	retrieved, err := db.GetSymbol(symbol.ID)
	if err != nil {
		t.Fatalf("GetSymbol failed: %v", err)
	}

	if retrieved.Name != "NewName" {
		t.Errorf("Expected name NewName, got %s", retrieved.Name)
	}
	if retrieved.Documentation != "Updated docs" {
		t.Errorf("Expected updated docs, got %s", retrieved.Documentation)
	}
}

func TestGetAllFilesForProject(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	project := &types.Project{Name: "test", Path: "/test"}
	db.CreateProject(project)

	// Create multiple files
	files := []*types.File{
		{ProjectID: project.ID, Path: "/test/file1.go", Language: "go"},
		{ProjectID: project.ID, Path: "/test/file2.py", Language: "python"},
		{ProjectID: project.ID, Path: "/test/file3.go", Language: "go"},
	}

	for _, file := range files {
		db.SaveFile(file)
	}

	// Get all files
	allFiles, err := db.GetAllFilesForProject(project.ID)
	if err != nil {
		t.Fatalf("GetAllFilesForProject failed: %v", err)
	}

	if len(allFiles) != 3 {
		t.Errorf("Expected 3 files, got %d", len(allFiles))
	}
}

func TestDatabasePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "persist.db")

	// Create database and add data
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	project := &types.Project{Name: "persist-test", Path: "/test"}
	db.CreateProject(project)
	db.Close()

	// Reopen database
	db2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to reopen database: %v", err)
	}
	defer db2.Close()

	// Verify data persisted
	retrieved, err := db2.GetProject(project.Path)
	if err != nil {
		t.Fatalf("GetProject failed after reopen: %v", err)
	}

	if retrieved.Name != project.Name {
		t.Errorf("Expected name %s after reopen, got %s", project.Name, retrieved.Name)
	}

	// Verify database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file should exist")
	}
}
