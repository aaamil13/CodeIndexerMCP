package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aaamil13/CodeIndexerMCP/internal/core"
	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/internal/lsp"
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
	db, err := database.NewDatabase(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Initialize indexer
	indexer := core.NewIndexer(db)

	// Create LSP server
	server := lsp.NewServer(db, indexer)

	// Start server (reads from stdin, writes to stdout)
	log.Println("LSP Server initialized. Listening for requests...")
	if err := server.Start(os.Stdin, os.Stdout); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
