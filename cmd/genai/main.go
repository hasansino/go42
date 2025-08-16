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

		"mcp__gopls__go_workspace",
		"mcp__gopls__go_search",
		"mcp__gopls__go_file_context",
		"mcp__gopls__go_package_api",
		"mcp__gopls__go_symbol_references",
		"mcp__gopls__go_diagnostics",

		"mcp__github__create_issue",
		"mcp__github__update_issue",
		"mcp__github__get_issue",
		"mcp__github__get_issue_comments",
		"mcp__github__add_issue_comment",
		"mcp__github__list_issues",
		"mcp__github__search_issues",
		"mcp__github__list_issue_types",
		"mcp__github__add_sub_issue",
		"mcp__github__remove_sub_issue",
		"mcp__github__reprioritize_sub_issue",
		"mcp__github__list_sub_issues",
		"mcp__github__assign_copilot_to_issue",

		"mcp__github__create_pull_request",
		"mcp__github__update_pull_request",
		"mcp__github__update_pull_request_branch",
		"mcp__github__get_pull_request",
		"mcp__github__get_pull_request_comments",
		"mcp__github__get_pull_request_diff",
		"mcp__github__get_pull_request_files",
		"mcp__github__get_pull_request_reviews",
		"mcp__github__get_pull_request_status",
		"mcp__github__list_pull_requests",
		"mcp__github__search_pull_requests",
		"mcp__github__merge_pull_request",

		"mcp__github__create_pending_pull_request_review",
		"mcp__github__add_comment_to_pending_review",
		"mcp__github__submit_pending_pull_request_review",
		"mcp__github__delete_pending_pull_request_review",
		"mcp__github__create_and_submit_pull_request_review",
		"mcp__github__request_copilot_review",

		"mcp__github__create_repository",
		"mcp__github__fork_repository",
		"mcp__github__search_repositories",

		"mcp__github__create_branch",
		"mcp__github__list_branches",
		"mcp__github__get_commit",
		"mcp__github__list_commits",

		"mcp__github__list_workflows",
		"mcp__github__run_workflow",
		"mcp__github__get_workflow_run",
		"mcp__github__list_workflow_runs",
		"mcp__github__list_workflow_jobs",
		"mcp__github__rerun_workflow_run",
		"mcp__github__rerun_failed_jobs",
		"mcp__github__cancel_workflow_run",
		"mcp__github__get_workflow_run_logs",
		"mcp__github__delete_workflow_run_logs",
		"mcp__github__get_workflow_run_usage",
		"mcp__github__list_workflow_run_artifacts",
		"mcp__github__download_workflow_run_artifact",
		"mcp__github__get_job_logs",

		"mcp__github__list_code_scanning_alerts",
		"mcp__github__get_code_scanning_alert",
		"mcp__github__list_dependabot_alerts",
		"mcp__github__get_dependabot_alert",
		"mcp__github__list_secret_scanning_alerts",
		"mcp__github__get_secret_scanning_alert",

		"mcp__github__list_notifications",
		"mcp__github__get_notification_details",
		"mcp__github__dismiss_notification",
		"mcp__github__manage_notification_subscription",
		"mcp__github__manage_repository_notification_subscription",
		"mcp__github__mark_all_notifications_read",

		"mcp__github__search_code",
		"mcp__github__search_users",
		"mcp__github__search_orgs",

		"mcp__github__get_discussion",
		"mcp__github__get_discussion_comments",
		"mcp__github__list_discussions",
		"mcp__github__list_discussion_categories",

		"mcp__github__get_tag",
		"mcp__github__list_tags",
		"mcp__github__list_releases",
		"mcp__github__get_latest_release",

		"mcp__github__get_me",
		"mcp__github__get_teams",
		"mcp__github__get_team_members",

		"mcp__kwb__search", "mcp__kwb__list_files", "mcp__kwb__get_file"
	],
    "deny": []
  }, 
  "enabledMcpjsonServers": [
    "gopls", "github", "kwb"
  ]
}`

const claudeMCPConfigPath = ".mcp.json"
const claudeMCPConfig = `{
  "mcpServers": {
    "gopls": {
      "command": "gopls",
      "args": ["mcp"]
    },
    "github": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-e",
        "GITHUB_PERSONAL_ACCESS_TOKEN",
        "-e",
        "GITHUB_TOOLSETS=all",
        "ghcr.io/github/github-mcp-server"
      ],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "$GH_MCP_KEY"
      }
    },
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

    "mcp__gopls__go_workspace",
    "mcp__gopls__go_search",
    "mcp__gopls__go_file_context",
    "mcp__gopls__go_package_api",
    "mcp__gopls__go_symbol_references",
    "mcp__gopls__go_diagnostics",

    "mcp__github__create_issue",
    "mcp__github__update_issue",
    "mcp__github__get_issue",
    "mcp__github__get_issue_comments",
    "mcp__github__add_issue_comment",
    "mcp__github__list_issues",
    "mcp__github__search_issues",
    "mcp__github__list_issue_types",
    "mcp__github__add_sub_issue",
    "mcp__github__remove_sub_issue",
    "mcp__github__reprioritize_sub_issue",
    "mcp__github__list_sub_issues",
    "mcp__github__assign_copilot_to_issue",

    "mcp__github__create_pull_request",
    "mcp__github__update_pull_request",
    "mcp__github__update_pull_request_branch",
    "mcp__github__get_pull_request",
    "mcp__github__get_pull_request_comments",
    "mcp__github__get_pull_request_diff",
    "mcp__github__get_pull_request_files",
    "mcp__github__get_pull_request_reviews",
    "mcp__github__get_pull_request_status",
    "mcp__github__list_pull_requests",
    "mcp__github__search_pull_requests",
    "mcp__github__merge_pull_request",

    "mcp__github__create_pending_pull_request_review",
    "mcp__github__add_comment_to_pending_review",
    "mcp__github__submit_pending_pull_request_review",
    "mcp__github__delete_pending_pull_request_review",
    "mcp__github__create_and_submit_pull_request_review",
    "mcp__github__request_copilot_review",

    "mcp__github__create_repository",
    "mcp__github__fork_repository",
    "mcp__github__search_repositories",

    "mcp__github__create_branch",
    "mcp__github__list_branches",
    "mcp__github__get_commit",
    "mcp__github__list_commits",

    "mcp__github__list_workflows",
    "mcp__github__run_workflow",
    "mcp__github__get_workflow_run",
    "mcp__github__list_workflow_runs",
    "mcp__github__list_workflow_jobs",
    "mcp__github__rerun_workflow_run",
    "mcp__github__rerun_failed_jobs",
    "mcp__github__cancel_workflow_run",
    "mcp__github__get_workflow_run_logs",
    "mcp__github__delete_workflow_run_logs",
    "mcp__github__get_workflow_run_usage",
    "mcp__github__list_workflow_run_artifacts",
    "mcp__github__download_workflow_run_artifact",
    "mcp__github__get_job_logs",

    "mcp__github__list_code_scanning_alerts",
    "mcp__github__get_code_scanning_alert",
    "mcp__github__list_dependabot_alerts",
    "mcp__github__get_dependabot_alert",
    "mcp__github__list_secret_scanning_alerts",
    "mcp__github__get_secret_scanning_alert",

    "mcp__github__list_notifications",
    "mcp__github__get_notification_details",
    "mcp__github__dismiss_notification",
    "mcp__github__manage_notification_subscription",
    "mcp__github__manage_repository_notification_subscription",
    "mcp__github__mark_all_notifications_read",

    "mcp__github__search_code",
    "mcp__github__search_users",
    "mcp__github__search_orgs",

    "mcp__github__get_discussion",
    "mcp__github__get_discussion_comments",
    "mcp__github__list_discussions",
    "mcp__github__list_discussion_categories",

    "mcp__github__get_tag",
    "mcp__github__list_tags",
    "mcp__github__list_releases",
    "mcp__github__get_latest_release",

    "mcp__github__get_me",
    "mcp__github__get_teams",
    "mcp__github__get_team_members",

    "mcp__kwb__search", "mcp__kwb__list_files", "mcp__kwb__get_file"
  ],
  "excludeTools": [],
  "maxSessionTurns": 10,
  "maxSessionDuration": 600,
  "checkpointing": {"enabled": true},
  "autoAccept": true,
  "mcpServers": {
    "gopls": {
      "command": "gopls",
      "args": ["mcp"],
      "env": {},
      "timeout": 30000,
      "trust": true
    },
    "github": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-e",
        "GITHUB_PERSONAL_ACCESS_TOKEN",
        "-e",
        "GITHUB_TOOLSETS=all",
        "ghcr.io/github/github-mcp-server"
      ],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "$GH_MCP_KEY"
      },
      "timeout": 30000,
      "trust": true
    },
    "kwb": {
      "command": "go",
      "args": ["run", "cmd/genkwb/main.go", "-serve", "-index", "ai/index"],
      "env": {},
      "timeout": 30000,
      "trust": true
    }
  },
  "allowMCPServers": ["gopls", "github", "kwb"],
  "usageStatisticsEnabled": false
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
    "github": {
      "type": "stdio",
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-e",
        "GITHUB_PERSONAL_ACCESS_TOKEN",
        "-e",
        "GITHUB_TOOLSETS=all",
        "ghcr.io/github/github-mcp-server"
      ],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "${GH_MCP_KEY}"
      }
    },
    "kwb": {
      "type": "stdio",
      "command": "go",
      "args": ["run", "cmd/genkwb/main.go", "-serve", "-index", "ai/index"],
	  "env": {}
    }
  },
  "permissions": {
    "allowed_tools": [
      "view",
      "ls",
      "grep",
      "edit",

      "mcp__gopls__go_workspace",
      "mcp__gopls__go_search",
      "mcp__gopls__go_file_context",
      "mcp__gopls__go_package_api",
      "mcp__gopls__go_symbol_references",
      "mcp__gopls__go_diagnostics",

      "mcp__github__create_issue",
      "mcp__github__update_issue",
      "mcp__github__get_issue",
      "mcp__github__get_issue_comments",
      "mcp__github__add_issue_comment",
      "mcp__github__list_issues",
      "mcp__github__search_issues",
      "mcp__github__list_issue_types",
      "mcp__github__add_sub_issue",
      "mcp__github__remove_sub_issue",
      "mcp__github__reprioritize_sub_issue",
      "mcp__github__list_sub_issues",
      "mcp__github__assign_copilot_to_issue",

      "mcp__github__create_pull_request",
      "mcp__github__update_pull_request",
      "mcp__github__update_pull_request_branch",
      "mcp__github__get_pull_request",
      "mcp__github__get_pull_request_comments",
      "mcp__github__get_pull_request_diff",
      "mcp__github__get_pull_request_files",
      "mcp__github__get_pull_request_reviews",
      "mcp__github__get_pull_request_status",
      "mcp__github__list_pull_requests",
      "mcp__github__search_pull_requests",
      "mcp__github__merge_pull_request",

      "mcp__github__create_pending_pull_request_review",
      "mcp__github__add_comment_to_pending_review",
      "mcp__github__submit_pending_pull_request_review",
      "mcp__github__delete_pending_pull_request_review",
      "mcp__github__create_and_submit_pull_request_review",
      "mcp__github__request_copilot_review",

      "mcp__github__create_repository",
      "mcp__github__fork_repository",
      "mcp__github__search_repositories",

      "mcp__github__create_branch",
      "mcp__github__list_branches",
      "mcp__github__get_commit",
      "mcp__github__list_commits",

      "mcp__github__list_workflows",
      "mcp__github__run_workflow",
      "mcp__github__get_workflow_run",
      "mcp__github__list_workflow_runs",
      "mcp__github__list_workflow_jobs",
      "mcp__github__rerun_workflow_run",
      "mcp__github__rerun_failed_jobs",
      "mcp__github__cancel_workflow_run",
      "mcp__github__get_workflow_run_logs",
      "mcp__github__delete_workflow_run_logs",
      "mcp__github__get_workflow_run_usage",
      "mcp__github__list_workflow_run_artifacts",
      "mcp__github__download_workflow_run_artifact",
      "mcp__github__get_job_logs",

      "mcp__github__list_code_scanning_alerts",
      "mcp__github__get_code_scanning_alert",
      "mcp__github__list_dependabot_alerts",
      "mcp__github__get_dependabot_alert",
      "mcp__github__list_secret_scanning_alerts",
      "mcp__github__get_secret_scanning_alert",

      "mcp__github__list_notifications",
      "mcp__github__get_notification_details",
      "mcp__github__dismiss_notification",
      "mcp__github__manage_notification_subscription",
      "mcp__github__manage_repository_notification_subscription",
      "mcp__github__mark_all_notifications_read",

      "mcp__github__search_code",
      "mcp__github__search_users",
      "mcp__github__search_orgs",

      "mcp__github__get_discussion",
      "mcp__github__get_discussion_comments",
      "mcp__github__list_discussions",
      "mcp__github__list_discussion_categories",

      "mcp__github__get_tag",
      "mcp__github__list_tags",
      "mcp__github__list_releases",
      "mcp__github__get_latest_release",

      "mcp__github__get_me",
      "mcp__github__get_teams",
      "mcp__github__get_team_members",

      "mcp__kwb__search", "mcp__kwb__list_files", "mcp__kwb__get_file"
    ]
  }
}`

var configs = map[string]string{
	claudeConfigPath:    claudeConfig,
	claudeMCPConfigPath: claudeMCPConfig,
	geminiConfigPath:    geminiConfig,
	crushConfigPath:     crushConfig,
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
