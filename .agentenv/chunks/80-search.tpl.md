### Navigation Guide

This guide explains how to effectively search and navigate codebases using available tools, from specialized MCP servers to native search utilities.

#### Tool Categories

**Specialized MCP Servers** - Language-specific or domain-specific tools that provide semantic understanding:
- Language servers (gopls for Go, typescript-language-server, rust-analyzer, etc.)
- Knowledge base servers (go42x-kwb or similar indexed search tools)
- Documentation servers and API explorers

**Native Search Tools** - Built-in utilities for text and pattern matching:
- **Grep** - Fast regex-based content search across files
- **Glob** - File pattern matching and discovery
- **Read** - Direct file content access
- **Task** - Complex multi-step search operations

#### Choosing the Right Search Tool

**Start with specialized tools if available**, then fall back to native tools when needed.

**For initial exploration:**
- Try knowledge base tools (like go42x-kwb) for indexed searches
- Use Glob to discover file structure and patterns
- Use Grep for broad keyword searches
- Example: Finding all configuration files → Try KB tool's list_files, fallback to Glob with "**/*.{json,yaml,toml}"

**For text and pattern searches:**
- Try knowledge base search functions first (faster for indexed content)
- Fall back to Grep for complex regex patterns
- Use Task for multi-round iterative searches
- Example: Finding error messages → KB search, then Grep with pattern "error|Error|ERROR"

**For semantic code understanding:**
- Use language-specific servers when available (gopls, typescript-language-server, etc.)
- These understand imports, types, symbols, and relationships
- Fall back to Grep + Read for basic symbol searches
- Example: Finding function references → Language server's find-references, fallback to Grep

**For file reading:**
- Use specialized getters if they provide additional context
- Fall back to Read tool for direct access
- Combine with language servers for semantic context
- Example: Reading a config file → KB get_file for metadata, or Read for raw content

#### Search Strategy Decision Tree

```
What do you need to find?
│
├─> Files by name/type?
│   ├─> Try: KB list_files or similar
│   └─> Fallback: Glob with patterns
│
├─> Text/keywords in files?
│   ├─> Try: KB search functions
│   └─> Fallback: Grep with regex
│
├─> Code symbols/definitions?
│   ├─> Try: Language server search
│   └─> Fallback: Grep + Read combination
│
├─> Symbol references/usages?
│   ├─> Try: Language server references
│   └─> Fallback: Grep across codebase
│
├─> Package/module structure?
│   ├─> Try: Language server workspace analysis
│   └─> Fallback: Glob + Read manifest files
│
└─> Complex multi-step search?
    └─> Use: Task tool for autonomous searching
```

#### Practical Search Examples

**Example: Understanding a new codebase**
1. Check for specialized tools (language servers, KB tools)
2. Use Glob to map file structure ("**/*.{js,ts,py,go,java}")
3. Use Grep to find entry points (pattern: "main|Main|entry|start")
4. Read configuration files to understand setup
5. Use language servers for dependency analysis if available

**Example: Finding and fixing a bug**
1. Search error messages with KB tool or Grep
2. Use language server to trace symbol references
3. Fall back to Grep for text-based call tracking
4. Read relevant files for context
5. Verify fixes with language diagnostics or tests

**Example: Adding a new feature**
1. Search for similar features with semantic search
2. Fall back to Grep for pattern matching
3. Use Glob to find related test files
4. Read API documentation and examples
5. Use language server for type checking if available

#### Performance Tips

**Fast operations:**
- Indexed KB searches (milliseconds)
- Glob for file discovery (fast for patterns)
- Read for known file paths (instant)

**Moderate operations:**
- Grep on small-medium codebases (seconds)
- Language server queries (depends on project size)

**Slower operations:**
- Grep on very large codebases (may take longer)
- Complex Task operations (multiple rounds)
- Initial language server indexing

#### Search Tool Priority Order

1. **Specialized MCP tools** (if available and applicable)
    - Language servers for semantic understanding
    - Knowledge base for indexed searches
    - Domain-specific tools

2. **Native search tools** (always available)
    - Grep for content search
    - Glob for file patterns
    - Read for direct access
    - Task for complex searches

3. **Combination strategies**
    - Use specialized tools for precision
    - Use native tools for breadth
    - Combine both for comprehensive analysis

**Key principles:**
- Start specific (specialized tools) then go broad (native tools)
- Use semantic search for code understanding
- Use text search for patterns and keywords
- Combine tools for best results
- Always have native tools as fallback
