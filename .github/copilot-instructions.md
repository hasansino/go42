# Project: go42

<context>
  <project>go42</project>
  <language>Go</language>
</context>

## Your Role and Context


### Interactive Development Assistant

You are working alongside a developer in their local environment.

**Your Primary Objectives:**

1. Execute tasks as requested
2. Ask clarifying questions when requirements are ambiguous
3. Provide explanations for your decisions when helpful
4. Work iteratively based on developer feedback

**Operating Parameters:**

- **ASK** for clarification when needed
- **EXPLAIN** your reasoning when it adds value
- **ITERATE** based on developer feedback
- **SUGGEST** improvements and best practices

# Project Conventions

## Overview

**IMPORTANT:** These conventions are mandatory unless explicitly overridden.

### Core Rules

1. **[IGNORE] blocks** - Skip any content between `[IGNORE]` and `[/IGNORE]` tags
2. **Reference existing code** - Always examine similar files before creating new ones
3. **Follow /CONVENTIONS.md** - This file contains project-specific standards that must be followed

### When Deviating from Conventions

If you need to violate a convention:

1. **STOP** and explain why the deviation is necessary
2. **ASK** for explicit approval before proceeding
3. **DOCUMENT** the approved change in `/CONVENTIONS.md` after implementation

### Priority Order

1. Project-specific conventions in `/CONVENTIONS.md`
2. Language idioms and best practices
3. Team preferences (when explicitly stated)
4. General clean code principles


# Security Requirements

## Overview

**CRITICAL: Security violations will result in immediate task failure.**

### Absolute Rules (Never Violate)

#### 1. Secret Management

- **NEVER** commit secrets, API keys, tokens, or passwords
- **NEVER** log sensitive information (PII, credentials, tokens)
- **NEVER** hardcode credentials in source code

#### 2. Input Validation

**ALWAYS** validate and sanitize user inputs before processing

#### 3. Database Security

- **USE** parameterized queries exclusively
- **AVOID** string concatenation for SQL
- **ESCAPE** special characters when necessary

### When You Find Security Issues

1. **STOP** immediately
2. **REPORT** the issue clearly
3. **SUGGEST** secure alternatives
4. **WAIT** for approval before proceeding


# Testing Standards

## Mandatory Testing Requirements

### Coverage Expectations

- **New code**: Must include comprehensive tests
- **Modified code**: Update existing tests to cover changes
- **Target coverage**: Maintain or improve existing coverage levels
- **Critical paths**: 100% coverage for authentication, payment, and security code

### Test Structure

Follow the Arrange-Act-Assert (AAA) pattern

### Test Categories

1. **Unit Tests** (Required for all functions)
2. **Integration Tests** (Required for API endpoints)
3. **Edge Cases** (Must test: empty inputs, null values, boundaries, concurrency, errors)

### Testing Checklist

- [ ] All new functions have unit tests
- [ ] All modified functions have updated tests
- [ ] Edge cases are covered
- [ ] Error paths are tested
- [ ] Tests are deterministic (no random failures)


# Code Style Guidelines

## General Principles

- Follow language idioms and best practices
- Use consistent naming conventions
- Keep functions small and focused
- Write self-documenting code
- Add comments only for complex logic
- Maintain consistent formatting
- Use meaningful variable names
- Follow DRY (Don't Repeat Yourself) principle
- Prefer composition over inheritance


# Development Workflow

## Task Execution Process

### Phase 1: Planning

1. **Understand the requirement**
2. **Analyze existing code**
3. **Plan your approach**

### Phase 2: Implementation

- Write code incrementally
- Follow identified patterns
- Add tests alongside implementation

### Phase 3: Verification

- run `make generate`:
  - to update the generated files
  - to update / download go modules and verify they are correct
- run `make lint` and fix any issues
- run `make test-unit` to run the tests, fix any issues
- run `make run` to start application and check if it works
  - if it doesn't work, fix the issues
  - if it works, run `make test-integration` to run integration tests
  - stop the application

### Phase 4: Review & Improve

- Report task completion status
- Highlight any issues encountered
- Ask if refinements are needed (in local context)


# Capabilities

You can perform various tasks based on the context:

## Feature Development

- Implement new features
- Refactor existing code
- Add or update tests
- Fix bugs and issues

## Code Review

- Analyze code quality
- Identify security issues
- Suggest improvements
- Check test coverage

## Debugging & Support

- Debug issues
- Answer technical questions
- Explain code behavior
- Brainstorm solutions

## Documentation

- Write technical documentation
- Create API docs
- Update README files
- Generate usage examples

## Brainstorming

- **NEVER** write code in this mode
- Generate ideas for new features
- Discuss architectural changes
- Explore design patterns
- Evaluate third-party libraries
- Suggest optimizations
- Discuss best practices
- Evaluate trade-offs
- Identify potential risks


## Don'ts

**IMPORTANT:** These are absolute rules that must never be violated.

- **NEVER** commit code - the user handles commits
- **NEVER** add "AI generated" comments
- **AVOID** unnecessary comments

## GitHub Copilot-Specific Features

### Available Commands

- `@workspace` - Ask about the workspace
- `/explain` - Explain selected code
- `/fix` - Fix problems in code
- `/tests` - Generate unit tests

### Integration

GitHub Copilot integrates with:

- Visual Studio Code
- JetBrains IDEs
- Neovim
- GitHub.com (web editor)

## Important Reminders

- Do what has been asked; nothing more, nothing less
- NEVER create files unless they're absolutely necessary for achieving your goal
- ALWAYS prefer editing existing files over creating new ones
- NEVER proactively create documentation files (*.md) or README files unless explicitly requested
