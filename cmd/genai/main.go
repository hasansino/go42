package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

const claudeConfigPath = ".claude/settings.json"
const claudeConfig = `{
  "permissions": {
    "allow": [
      	"Edit", 
		"Glob", 
		"Grep", 
		"LS", 
		"MultiEdit",
		"Read", 
		"Task", 
		"TodoWrite", 
		"WebFetch", 
		"WebSearch", 
		"Write",
      	"Bash",
      	"mcp__kwb__search", "mcp__kwb__list_files", "mcp__kwb__get_file"
    ],
    "deny": []
  }, 
  "enabledMcpjsonServers": [
    "kwb"
  ]
}`

const claudeMCPConfigPath = ".mcp.json"
const claudeMCPConfig = `{
  "mcpServers": {
    "kwb": {
      "command": "go",
      "args": [
        "run",
        "cmd/genkwb/main.go",
        "-serve",
        "-index",
        "ai/index"
      ]
    }
  }
}`

const crushConfigPath = ".crush.json"
const crushConfig = `{
  "$schema": "https://charm.land/crush.json",
  "lsp": {
    "go": {
      "command": "gopls"
    }
  },
  "mcp": {
    "filesystem": {
      "type": "stdio",
      "command": "go",
      "args": ["run", "cmd/genkwb/main.go", "-serve", "-index", "ai/index"],
	  "env": {},
    }
  },
  "permissions": {
    "allowed_tools": []
  }
}
`

const geminiConfigPath = ".gemini/settings.json"
const geminiConfig = `{
  "coreTools": [
	"LSTool", 
	"ReadFileTool", 
	"WriteFileTool", 
	"GrepTool", 
	"GlobTool", 
	"EditTool", 
	"ReadManyFilesTool", 
	"ShellTool", 
	"WebFetchTool", 
	"WebSearchTool", 
	"MemoryTool",
	"mcp__kwb__search", "mcp__kwb__list_files", "mcp__kwb__get_file"
  ],
  "excludeTools": [],
  "maxSessionTurns": 10,
  "maxSessionDuration": 600,
  "checkpointing": {"enabled": true},
  "autoAccept": true,
  "mcpServers": {
    "kwb": {
      "command": "go",
      "args": ["run", "cmd/genkwb/main.go", "-serve", "-index", "ai/index"],
      "env": {},
      "timeout": 30000,
      "trust": true
    }
  },
  "allowMCPServers": ["kwb"],
  "usageStatisticsEnabled": false
}`

var configs = map[string]string{
	claudeConfigPath:    claudeConfig,
	claudeMCPConfigPath: claudeMCPConfig,
	crushConfigPath:     crushConfig,
	geminiConfigPath:    geminiConfig,
}

// ----

type Config struct {
	Project   ProjectConfig             `yaml:"project"`
	Providers map[string]ProviderConfig `yaml:"providers"`
}

type ProjectConfig struct {
	Name        string                 `yaml:"name"`
	Language    string                 `yaml:"language"`
	Description string                 `yaml:"description"`
	Metadata    map[string]interface{} `yaml:"metadata"`
}

type ProviderConfig struct {
	Template string `yaml:"template"`
	Output   string `yaml:"output"`
}

type Content struct {
	Chunks map[string]string
	Order  []string
}

type TemplateData struct {
	Project      string
	Language     string
	Description  string
	Metadata     map[string]interface{}
	Branch       string
	IsCI         bool
	PRNumber     string
	CommitSHA    string
	BuildURL     string
	TargetBranch string
	Content      Content
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	config, err := loadConfig("ai/genai.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	content, err := loadChunks("ai/chunks")
	if err != nil {
		return fmt.Errorf("failed to load content: %w", err)
	}

	envData := &TemplateData{
		Project:      config.Project.Name,
		Language:     config.Project.Language,
		Description:  config.Project.Description,
		Metadata:     config.Project.Metadata,
		Branch:       getEnv("BRANCH", "main"),
		IsCI:         getEnv("CI", "") == "true",
		PRNumber:     getEnv("PR_NUMBER", ""),
		CommitSHA:    getEnv("COMMIT_SHA", ""),
		BuildURL:     getEnv("BUILD_URL", ""),
		TargetBranch: getEnv("TARGET_BRANCH", "main"),
		Content:      *content,
	}

	if envData.IsCI {
		ciBranch := getEnv("CI_BRANCH", "")
		if ciBranch != "" {
			envData.Branch = ciBranch
		}
	}

	for name, provider := range config.Providers {
		if err := generateProvider(provider, envData); err != nil {
			return fmt.Errorf("failed to generate for %s: %w", name, err)
		}
	}

	if err := exportConfigs(); err != nil {
		return fmt.Errorf("failed to export configs: %w", err)
	}

	if err := copyAgents(); err != nil {
		fmt.Printf("Warning: failed to copy agents: %v\n", err)
	}

	return nil
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func loadChunks(dir string) (*Content, error) {
	content := &Content{
		Chunks: make(map[string]string),
		Order:  []string{},
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read chunks directory: %w", err)
	}

	var mdFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".md") {
			mdFiles = append(mdFiles, entry.Name())
		}
	}

	sort.Strings(mdFiles)

	for _, filename := range mdFiles {
		path := filepath.Join(dir, filename)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", filename, err)
		}

		// Use filename without extension as key
		key := strings.TrimSuffix(filename, ".md")
		content.Chunks[key] = string(data)
		content.Order = append(content.Order, key)
	}

	if len(content.Chunks) == 0 {
		return nil, fmt.Errorf("no content chunks found in %s", dir)
	}

	return content, nil
}

func generateProvider(provider ProviderConfig, data *TemplateData) error {
	tmplData, err := os.ReadFile(provider.Template)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	tmpl, err := template.New("provider").Parse(string(tmplData))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	dir := filepath.Dir(provider.Output)
	if dir != "." && dir != "/" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	if err := os.WriteFile(provider.Output, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	fmt.Printf("Generated %s\n", provider.Output)

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func exportConfigs() error {
	for path, content := range configs {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", path, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("failed to write %s: %w", path, err)
		}
		fmt.Printf("Exported config to %s\n", path)
	}
	return nil
}

func copyAgents() error {
	srcDir := "ai/agents"
	dstDir := ".claude/agents"

	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return nil // no agents to copy
	}

	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return fmt.Errorf("failed to create .claude/agents: %w", err)
	}

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("failed to read agents directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		// Read source file
		data, err := os.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", srcPath, err)
		}

		// Write to destination
		if err := os.WriteFile(dstPath, data, 0o644); err != nil {
			return fmt.Errorf("failed to write %s: %w", dstPath, err)
		}

		fmt.Printf("Copied agent: %s\n", entry.Name())
	}

	return nil
}
