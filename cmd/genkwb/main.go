package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	defaultIndexPath = "ai/index"
	serverName       = "kwb"
	serverVersion    = "0.1.0"
)

type KnowledgeServer struct {
	index bleve.Index
}

type Document struct {
	ID      string `json:"id"`
	Path    string `json:"path"`
	Content string `json:"content"`
	Type    string `json:"type"`
}

func main() {
	var (
		indexPath = flag.String("index", defaultIndexPath, "Path to Bleve index")
		buildMode = flag.Bool("build", false, "Build/rebuild index")
		serve     = flag.Bool("serve", false, "Start MCP server")
	)
	flag.Parse()

	if *buildMode {
		if err := buildIndex(*indexPath); err != nil {
			slog.Error("failed to build index", slog.Any("error", err))
			os.Exit(1)
		}
		slog.Info("index built successfully", slog.String("path", *indexPath))
		return
	}

	if *serve {
		if err := runServer(*indexPath); err != nil {
			slog.Error("failed to run server", slog.Any("error", err))
			os.Exit(1)
		}
		return
	}

	flag.Usage()
}

func buildIndex(indexPath string) error {
	err := os.RemoveAll(indexPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing old index: %w", err)
	}

	mapping := bleve.NewIndexMapping()
	index, err := bleve.New(indexPath, mapping)
	if err != nil {
		return fmt.Errorf("creating index: %w", err)
	}
	defer func(index bleve.Index) {
		_ = index.Close()
	}(index)

	extensions := map[string]bool{
		".go":    true,
		".md":    true,
		".yaml":  true,
		".yml":   true,
		".mod":   true,
		".sum":   true,
		".proto": true,
		".sql":   true,
		".json":  true,
		".toml":  true,
		".env":   true,
		".sh":    true,
	}

	excludeDirs := []string{
		".git",
		"vendor",
		"node_modules",
		".idea",
		".vscode",
		"dist",
		"build",
		"bin",
		indexPath,
	}

	fileCount := 0
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories
		if info.IsDir() {
			for _, excl := range excludeDirs {
				if strings.Contains(path, excl) {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Check if file should be indexed
		shouldIndex := false
		ext := filepath.Ext(path)
		if extensions[ext] {
			shouldIndex = true
		} else if info.Name() == "Makefile" || info.Name() == "Dockerfile" || info.Name() == ".gitignore" {
			shouldIndex = true
		}

		if !shouldIndex {
			return nil
		}

		// Skip files in excluded directories
		for _, excl := range excludeDirs {
			if strings.Contains(path, excl) {
				return nil
			}
		}

		// Skip very large files (>1MB)
		if info.Size() > 1024*1024 {
			slog.Warn("skipping large file", "path", path, "size", info.Size())
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			slog.Warn("failed to read file", "path", path, "error", err)
			return nil
		}

		doc := Document{
			ID:      path,
			Path:    path,
			Content: string(content),
			Type:    getFileType(path),
		}

		if err := index.Index(doc.ID, doc); err != nil {
			slog.Warn("failed to index file", "path", path, "error", err)
			return nil
		}

		fileCount++
		slog.Debug("indexed file", "path", path)
		return nil
	})

	if err != nil {
		return fmt.Errorf("walking directory: %w", err)
	}

	count, _ := index.DocCount()
	slog.Info("indexing complete", "documents", count, "files_processed", fileCount)
	return nil
}

func getFileType(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".go":
		return "code"
	case ".md":
		return "documentation"
	case ".yaml", ".yml":
		return "config"
	case "":
		if strings.Contains(path, "Makefile") {
			return "makefile"
		}
		return "other"
	default:
		return "other"
	}
}

func runServer(indexPath string) error {
	index, err := bleve.Open(indexPath)
	if err != nil {
		return fmt.Errorf("opening index: %w", err)
	}
	defer func(index bleve.Index) {
		_ = index.Close()
	}(index)

	ks := &KnowledgeServer{index: index}

	mcpServer := server.NewMCPServer(
		serverName,
		serverVersion,
	)

	searchTool := mcp.NewTool("search",
		mcp.WithDescription("Search the knowledge base"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
		mcp.WithNumber("limit", mcp.Description("Maximum results (default: 10)")),
	)
	mcpServer.AddTool(searchTool, ks.searchHandler)

	getFileTool := mcp.NewTool("get_file",
		mcp.WithDescription("Get full content of a specific file"),
		mcp.WithString("path", mcp.Required(), mcp.Description("File path")),
	)
	mcpServer.AddTool(getFileTool, ks.getFileHandler)

	listFilesTool := mcp.NewTool("list_files",
		mcp.WithDescription("List all indexed files"),
		mcp.WithString("type", mcp.Description("Filter by type: code, documentation, config")),
	)
	mcpServer.AddTool(listFilesTool, ks.listFilesHandler)

	return server.ServeStdio(mcpServer)
}

func (s *KnowledgeServer) searchHandler(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("Invalid query parameter"), nil
	}

	limit := 10
	if l := request.GetFloat("limit", 10); l > 0 {
		limit = int(l)
	}

	bleveQuery := bleve.NewQueryStringQuery(query)
	searchRequest := bleve.NewSearchRequestOptions(bleveQuery, limit, 0, false)
	searchRequest.Fields = []string{"Path", "Type"}

	result, err := s.index.Search(searchRequest)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search error: %v", err)), nil
	}

	output := fmt.Sprintf("Found %d results:\n\n", result.Total)
	for i, hit := range result.Hits {
		output += fmt.Sprintf("%d. %s (score: %.2f)\n", i+1, hit.ID, hit.Score)

		// Skip preview for now as Document method returns different type
		output += "\n"
	}

	return mcp.NewToolResultText(output), nil
}

func (s *KnowledgeServer) getFileHandler(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	path, err := request.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("Invalid path parameter"), nil
	}

	// Read file directly since we can't get content from index easily
	content, err := os.ReadFile(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("File not found: %s", path)), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

func (s *KnowledgeServer) listFilesHandler(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	fileType := request.GetString("type", "")

	var q query.Query
	if fileType != "" {
		termQuery := bleve.NewTermQuery(fileType)
		termQuery.SetField("Type")
		q = termQuery
	} else {
		q = bleve.NewMatchAllQuery()
	}

	searchRequest := bleve.NewSearchRequestOptions(q, 1000, 0, false)
	searchRequest.Fields = []string{"Path", "Type"}

	result, err := s.index.Search(searchRequest)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search error: %v", err)), nil
	}

	output := fmt.Sprintf("Total files: %d\n\n", len(result.Hits))
	for _, hit := range result.Hits {
		output += fmt.Sprintf("- %s\n", hit.ID)
	}

	return mcp.NewToolResultText(output), nil
}
