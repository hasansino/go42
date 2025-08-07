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
