package main

import (
	"flag"
	// "fmt" // Imported and not used
	"log"
	"os"

	"github.com/aaamil13/CodeIndexerMCP/internal/core"
	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/internal/lsp"
	"github.com/aaamil13/CodeIndexerMCP/internal/utils"
)

func main() {
	// Parse command-line flags
	dbPath := flag.String("db", "./codeindex.db", "Path to database file")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	// Setup logging
	if *debug {
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	} else {
		log.SetFlags(0)
		log.SetOutput(os.Stderr) // LSP communication is on stdin/stdout
	}

	log.Printf("Starting CodeIndexer LSP Server...")
	log.Printf("Database: %s", *dbPath)

	// Initialize database
	logger := utils.NewLogger("[LSP Server]")
	db, err := database.NewManager(*dbPath, logger)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Get project path for indexer
	projectPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	// Initialize indexer (Indexer manages its own DB internally)
	// Passing nil for config to use default
	indexer, err := core.NewIndexer(projectPath, nil) // Corrected arguments and return values
	if err != nil {
		log.Fatalf("Failed to create indexer: %v", err)
	}

	// Create LSP server
	server := lsp.NewServer(db, indexer)

	// Start server (reads from stdin, writes to stdout)
	log.Println("LSP Server initialized. Listening for requests...")
	if err := server.Start(os.Stdin, os.Stdout); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
