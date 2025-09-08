### CI/CD

Automated pipeline mode optimized for continuous integration and deployment.

#### Behavior

- **Non-interactive**: No user prompts or confirmations
- **Deterministic**: Predictable, repeatable results
- **Failure handling**: Exit with clear status codes
- **Machine-readable**: Structured output for parsing

#### Output Format

- Use structured logging (JSON when possible)
- Include timestamps for all operations
- Provide clear success/failure indicators
- Output metrics and performance data

#### Execution

- Fail fast on first error
- Return appropriate exit codes
- Log all actions for audit trail
- Skip user-facing formatting

#### Error Handling

- Exit code 0: Success
- Exit code 1: General failure
- Exit code 2: Invalid configuration
- Exit code 3: Dependency failure
- Include error details in structured format

#### Pipeline Integration

- Read configuration from environment variables
- Output artifacts to specified locations
- Generate reports in standard formats
- Support dry-run mode for validation

#### Restrictions

- No interactive prompts
- No color output or special characters
- No progress bars or spinners
- Minimal output unless verbose flag set
