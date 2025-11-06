package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/aaamil13/CodeIndexerMCP/internal/core"
	"github.com/aaamil13/CodeIndexerMCP/internal/mcp"
	"github.com/aaamil13/CodeIndexerMCP/internal/utils"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		printUsage()
		return nil
	}

	command := os.Args[1]

	// Get project path (current directory by default)
	projectPath := "."
	if len(os.Args) > 2 {
		projectPath = os.Args[2]
	}

	// Make absolute
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("invalid project path: %w", err)
	}

	switch command {
	case "index":
		return runIndex(absPath)
	case "watch":
		return runWatch(absPath)
	case "mcp":
		return runMCP(absPath)
	case "search":
		if len(os.Args) < 3 {
			return fmt.Errorf("search requires a query argument")
		}
		query := os.Args[2]
		return runSearch(absPath, query)
	case "overview":
		return runOverview(absPath)
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func runIndex(projectPath string) error {
	fmt.Println("🚀 Code Indexer - Indexing project...")
	fmt.Println("Project:", projectPath)

	indexer, err := core.NewIndexer(projectPath, nil)
	if err != nil {
		return err
	}
	defer indexer.Close()

	if err := indexer.Initialize(); err != nil {
		return err
	}

	if err := indexer.IndexAll(); err != nil {
		return err
	}

	fmt.Println("✅ Indexing completed successfully!")
	return nil
}

func runWatch(projectPath string) error {
	fmt.Println("🔍 Code Indexer - Watch Mode")
	fmt.Println("Project:", projectPath)

	indexer, err := core.NewIndexer(projectPath, nil)
	if err != nil {
		return err
	}
	defer indexer.Close()

	if err := indexer.Initialize(); err != nil {
		return err
	}

	// Initial index
	fmt.Println("Performing initial index...")
	if err := indexer.IndexAll(); err != nil {
		return err
	}
	fmt.Println("✅ Initial indexing complete")

	// Start watching
	fmt.Println("👀 Watching for file changes... (Press Ctrl+C to stop)")
	if err := indexer.Watch(); err != nil {
		return err
	}

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\n🛑 Stopping watcher...")

	if err := indexer.StopWatch(); err != nil {
		return err
	}

	fmt.Println("✅ Watcher stopped")
	return nil
}

func runMCP(projectPath string) error {
	fmt.Println("🚀 Code Indexer MCP Server")
	fmt.Println("Project:", projectPath)

	indexer, err := core.NewIndexer(projectPath, nil)
	if err != nil {
		return err
	}
	defer indexer.Close()

	if err := indexer.Initialize(); err != nil {
		return err
	}

	// Index project on startup
	fmt.Println("Indexing project...")
	if err := indexer.IndexAll(); err != nil {
		return err
	}

	// Start MCP server
	server := mcp.NewServer(indexer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Fprintln(os.Stderr, "\nShutting down...")
		cancel()
	}()

	fmt.Fprintln(os.Stderr, "MCP Server started. Ready to receive requests.")

	if err := server.Start(ctx); err != nil && err != context.Canceled {
		return err
	}

	return nil
}

func runSearch(projectPath string, query string) error {
	indexer, err := core.NewIndexer(projectPath, nil)
	if err != nil {
		return err
	}
	defer indexer.Close()

	if err := indexer.Initialize(); err != nil {
		return err
	}

	opts := types.SearchOptions{
		Query: query,
		Limit: 20,
	}

	symbols, err := indexer.SearchSymbols(opts)
	if err != nil {
		return err
	}

	if len(symbols) == 0 {
		fmt.Println("No symbols found matching:", query)
		return nil
	}

	fmt.Printf("Found %d symbols:\n\n", len(symbols))

	for _, symbol := range symbols {
		fmt.Printf("📍 %s (%s)\n", symbol.Name, symbol.Type)
		if symbol.Signature != "" {
			fmt.Printf("   Signature: %s\n", symbol.Signature)
		}
		fmt.Printf("   Location: Line %d-%d\n", symbol.StartLine, symbol.EndLine)
		if symbol.Documentation != "" {
			fmt.Printf("   Docs: %s\n", truncate(symbol.Documentation, 80))
		}
		fmt.Println()
	}

	return nil
}

func runOverview(projectPath string) error {
	indexer, err := core.NewIndexer(projectPath, nil)
	if err != nil {
		return err
	}
	defer indexer.Close()

	if err := indexer.Initialize(); err != nil {
		return err
	}

	overview, err := indexer.GetProjectOverview()
	if err != nil {
		return err
	}

	fmt.Println("📊 Project Overview")
	fmt.Println("==================")
	fmt.Printf("Name: %s\n", overview.Project.Name)
	fmt.Printf("Path: %s\n", overview.Project.Path)
	fmt.Printf("Total Files: %d\n", overview.TotalFiles)
	fmt.Printf("Total Symbols: %d\n", overview.TotalSymbols)

	if len(overview.LanguageStats) > 0 {
		fmt.Println("\nLanguages:")
		for lang, count := range overview.LanguageStats {
			fmt.Printf("  - %s: %d files\n", lang, count)
		}
	}

	fmt.Printf("\nLast Indexed: %s\n", overview.Project.LastIndexed.Format("2006-01-02 15:04:05"))

	return nil
}

func printUsage() {
	fmt.Println(`Code Indexer MCP - Intelligent code indexer for AI agents

Usage:
  code-indexer <command> [arguments]

Commands:
  index [path]      Index the project at the given path (default: current directory)
  watch [path]      Watch for file changes and auto-index (default: current directory)
  mcp [path]        Start MCP server for the project
  search <query>    Search for symbols in the project
  overview [path]   Show project overview and statistics
  help              Show this help message

Examples:
  code-indexer index .
  code-indexer watch /path/to/project
  code-indexer mcp /path/to/project
  code-indexer search "MyFunction"
  code-indexer overview

For more information, visit: https://github.com/aaamil13/CodeIndexerMCP
`)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
