### MCP Tools Usage Guide: go42x-kwb vs gopls

This guide explains when to use the **go42x-kwb** (Knowledge Base) MCP server versus the **gopls** (Go Language Server) MCP server for analyzing and working with Go code.

#### Overview

Both tools serve different purposes in Go code analysis:
- **go42x-kwb**: A fast, indexed search tool for exploring codebases through keyword and pattern matching
- **gopls**: A semantic Go analysis tool that understands Go syntax, types, and relationships

### When to Use go42x-kwb

The Knowledge Base server excels at **fast text-based searches** and **initial code exploration**.

#### Best Use Cases

1. **Initial Codebase Exploration**
   - When you need to quickly understand what files exist in a project
   - Getting an overview of the codebase structure
   - Finding files by type (code, documentation, config)
   ```
   Example: "Show me all configuration files in the project"
   Tool: mcp__go42x-kwb__list_files with type="config"
   ```

2. **Keyword and Pattern Search**
   - Finding all occurrences of a specific string or pattern
   - Searching for TODO comments, error messages, or specific text
   - Looking for configuration values or environment variables
   ```
   Example: "Find all files mentioning 'database connection'"
   Tool: mcp__go42x-kwb__search with query="database connection"
   ```

3. **Quick File Content Retrieval**
   - When you know the exact file path and need its contents
   - Reading configuration files, documentation, or scripts
   - Accessing non-Go files (YAML, JSON, Markdown, etc.)
   ```
   Example: "Show me the README file"
   Tool: mcp__go42x-kwb__get_file with path="README.md"
   ```

4. **Cross-Language Searches**
   - Searching across mixed codebases (Go, JavaScript, Python, etc.)
   - Finding patterns in build scripts, CI/CD configs, and documentation
   - Exploring test fixtures and data files

#### Strengths

- ✅ Very fast search across large codebases
- ✅ Works with any file type, not just Go
- ✅ Great for text pattern matching
- ✅ Efficient for initial exploration
- ✅ Lightweight and quick to query

#### Limitations

- ❌ No semantic understanding of Go code
- ❌ Cannot find references or implementations
- ❌ No type information or code relationships
- ❌ No understanding of Go imports or packages

### When to Use gopls

The Go Language Server provides **deep semantic analysis** and **code intelligence** for Go projects.

#### Best Use Cases

1. **Understanding Go Code Structure**
   - Analyzing package dependencies and imports
   - Understanding module and workspace layout
   - Getting package API summaries
   ```
   Example: "What packages does this project contain?"
   Tool: mcp__gopls__go_workspace
   ```

2. **Finding Symbol Definitions and References**
   - Locating where a function, type, or variable is defined
   - Finding all usages of a specific symbol
   - Tracing method calls and type usage
   ```
   Example: "Find all references to the Server.Run method"
   Tool: mcp__gopls__go_symbol_references with symbol="Server.Run"
   ```

3. **Semantic Code Search**
   - Finding symbols by name with fuzzy matching
   - Searching for types, interfaces, functions across packages
   - Locating implementations of interfaces
   ```
   Example: "Find all types with 'Handler' in their name"
   Tool: mcp__gopls__go_search with query="handler"
   ```

4. **Code Context and Dependencies**
   - Understanding file dependencies within a package
   - Analyzing cross-file relationships
   - Getting context about imports and usage
   ```
   Example: "What does server.go depend on?"
   Tool: mcp__gopls__go_file_context with file="/path/to/server.go"
   ```

5. **Package API Analysis**
   - Understanding public APIs of packages
   - Exploring third-party dependencies
   - Reviewing exported types and functions
   ```
   Example: "Show me the public API of the storage package"
   Tool: mcp__gopls__go_package_api with packagePaths=["example.com/storage"]
   ```

6. **Code Quality and Diagnostics**
   - Finding compilation errors and issues
   - Checking for type errors
   - Validating code changes
   ```
   Example: "Check for errors in the edited files"
   Tool: mcp__gopls__go_diagnostics with files=["/path/to/file.go"]
   ```

#### Strengths

- ✅ Deep semantic understanding of Go code
- ✅ Accurate symbol resolution and type information
- ✅ Understands Go imports, packages, and modules
- ✅ Can find references, implementations, and dependencies
- ✅ Provides compilation diagnostics

#### Limitations

- ❌ Only works with Go code
- ❌ Slower for simple text searches
- ❌ Requires valid Go code to analyze
- ❌ More resource-intensive than text search

### Decision Flow Chart

```
Start: What do you need to do?
│
├─> Need to search for text/keywords?
│   └─> Use go42x-kwb__search
│
├─> Need to list/explore files?
│   └─> Use go42x-kwb__list_files
│
├─> Need to read a specific file?
│   ├─> Is it a Go file you'll analyze?
│   │   └─> Use Read tool (built-in) + gopls__go_file_context
│   └─> Just need contents?
│       └─> Use go42x-kwb__get_file
│
├─> Need to find Go symbols/types?
│   └─> Use gopls__go_search
│
├─> Need to find symbol references?
│   └─> Use gopls__go_symbol_references
│
├─> Need to understand package structure?
│   └─> Use gopls__go_workspace or gopls__go_package_api
│
└─> Need to check for Go errors?
    └─> Use gopls__go_diagnostics
```

### Practical Examples

#### Example 1: Understanding a New Codebase
```
1. Start with go42x-kwb__list_files to see project structure
2. Use go42x-kwb__search to find main entry points
3. Switch to gopls__go_workspace for Go module information
4. Use gopls__go_package_api to understand key packages
```

#### Example 2: Finding and Fixing a Bug
```
1. Use go42x-kwb__search to find error messages or relevant keywords
2. Use gopls__go_symbol_references to trace function calls
3. Use gopls__go_file_context to understand dependencies
4. After editing, use gopls__go_diagnostics to verify fixes
```

#### Example 3: Adding a New Feature
```
1. Use gopls__go_search to find similar existing features
2. Use gopls__go_package_api to understand available APIs
3. Use go42x-kwb__search to find examples in tests
4. Use gopls__go_diagnostics after implementation
```

### Performance Considerations

**Use go42x-kwb for:**
- Initial exploration (fast overview)
- Broad text searches across many files
- Non-Go file access
- Quick keyword lookups

**Use gopls for:**
- Precise symbol location
- Understanding code relationships
- Type-safe refactoring preparation
- Compilation checking

### Summary Rules

1. **Start with go42x-kwb** when exploring unknown codebases
2. **Use go42x-kwb** for text/pattern searches
3. **Switch to gopls** when you need semantic understanding
4. **Use gopls** for Go-specific analysis and refactoring
5. **Combine both** for comprehensive code analysis
6. **Prefer go42x-kwb** for non-Go files
7. **Prefer gopls** for Go symbol resolution and type information

Remember: go42x-kwb is your "grep on steroids" while gopls is your "Go code intelligence engine". Use them together for maximum effectiveness!
