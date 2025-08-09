package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

var (
	mcpServers = map[string]MCPServer{
		"kwb": {
			Command: "go",
			Args:    []string{"run", "cmd/genkwb/main.go", "-serve", "-index", "ai/index"},
			Env:     map[string]string{},
		},
	}
	enabledMCPServers  = []string{"kwb"}
	defaultPermissions = Permissions{
		Allow: []string{},
		Deny:  []string{},
	}
)

type MCPServerConfig struct {
	Servers map[string]MCPServer `json:"mcpServers"`
}

type MCPServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

type Settings struct {
	Permissions       Permissions `json:"permissions"`
	EnabledMCPServers []string    `json:"enabledMcpjsonServers"`
}

type Permissions struct {
	Allow []string `json:"allow"`
	Deny  []string `json:"deny"`
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

type Subagents struct {
	Agents map[string]string
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
	Subagents    Subagents
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

	content, err := loadContentFromChunks("ai/chunks")
	if err != nil {
		return fmt.Errorf("failed to load content: %w", err)
	}

	// Copy agents to .claude/agents directory
	if err := copyAgentsToClaudeDir(); err != nil {
		fmt.Printf("Warning: failed to copy agents: %v\n", err)
	}

	// Generate .claude/settings.json with MCP servers
	if err := generateClaudeSettings(); err != nil {
		fmt.Printf("Warning: failed to generate .claude/settings.json: %v\n", err)
	}

	if err := generateMCPServerConfig(); err != nil {
		fmt.Printf("Warning: failed to generate .mcp.json: %v\n", err)
	}

	subagents, err := loadSubagents("ai/agents")
	if err != nil {
		// Subagents are optional, so we just log if they're missing
		subagents = &Subagents{
			Agents: make(map[string]string),
			Order:  []string{},
		}
	}

	// Build environment data
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
		Subagents:    *subagents,
	}

	// Override branch in CI
	if envData.IsCI {
		ciBranch := getEnv("CI_BRANCH", "")
		if ciBranch != "" {
			envData.Branch = ciBranch
		}
	}

	for name, provider := range config.Providers {
		if err := generateForProvider(provider, name, envData); err != nil {
			return fmt.Errorf("failed to generate for %s: %w", name, err)
		}
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

func loadContentFromChunks(dir string) (*Content, error) {
	content := &Content{
		Chunks: make(map[string]string),
		Order:  []string{},
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read chunks directory: %w", err)
	}

	// Collect all md files
	var mdFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process .md files
		if strings.HasSuffix(entry.Name(), ".md") {
			mdFiles = append(mdFiles, entry.Name())
		}
	}

	// Sort files for consistent ordering
	sort.Strings(mdFiles)

	// Load each file
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

func generateForProvider(provider ProviderConfig, providerName string, data *TemplateData) error {

	// Load and execute template
	tmplData, err := os.ReadFile(provider.Template)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	// Add Claude-specific flag to template data
	templateData := struct {
		*TemplateData
		IsClaudeProvider bool
	}{
		TemplateData:     data,
		IsClaudeProvider: providerName == "claude",
	}

	tmpl, err := template.New("provider").Parse(string(tmplData))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Write output
	dir := filepath.Dir(provider.Output)
	if dir != "." && dir != "/" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	if err := os.WriteFile(provider.Output, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	fmt.Printf("Generated %s\n", provider.Output)

	return nil
}

func loadSubagents(dir string) (*Subagents, error) {
	subagents := &Subagents{
		Agents: make(map[string]string),
		Order:  []string{},
	}

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return subagents, nil // Directory doesn't exist, return empty
	}

	return subagents, nil // Agents are handled differently now
}

func copyAgentsToClaudeDir() error {
	srcDir := "ai/agents"
	dstDir := ".claude/agents"

	// Check if source directory exists
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return nil // No agents to copy
	}

	// Create destination directory
	if err := os.MkdirAll(dstDir, 0755); err != nil {
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
		if err := os.WriteFile(dstPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", dstPath, err)
		}

		fmt.Printf("Copied agent: %s\n", entry.Name())
	}

	return nil
}

func generateClaudeSettings() error {
	settings := Settings{
		EnabledMCPServers: enabledMCPServers,
		Permissions:       defaultPermissions,
	}

	claudeDir := ".claude"
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create .claude directory: %w", err)
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	settingsPath := filepath.Join(claudeDir, "settings.json")
	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings.json: %w", err)
	}

	fmt.Printf("Generated %s\n", settingsPath)

	return nil
}

func generateMCPServerConfig() error {
	config := MCPServerConfig{
		Servers: mcpServers,
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(".mcp.json", data, 0644); err != nil {
		return fmt.Errorf("failed to write settings.json: %w", err)
	}
	fmt.Printf("Generated %s\n", ".mcp.json")
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
