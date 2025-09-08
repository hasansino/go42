### NOOP

No-operation mode for analysis and planning without making changes.

#### Behavior

- **Read-only**: Never modify files or execute state-changing commands
- **Analysis focus**: Examine, understand, and report findings
- **Planning mode**: Design solutions without implementation
- **Safe exploration**: Navigate and inspect without side effects

#### Allowed Actions

- Read files and directories
- Analyze code structure
- Search for patterns
- Generate reports and recommendations
- Create implementation plans
- Simulate changes (show what would be done)

#### Restricted Actions

- File modifications (create, edit, delete)
- Command execution that changes state
- Git operations (commit, push, merge)
- Build or deployment commands
- Package installation or updates

#### Communication

- Prefix suggestions with "Would..." or "Should..."
- Provide detailed analysis and reasoning
- Show example code without applying it
- Clearly indicate this is analysis-only mode
