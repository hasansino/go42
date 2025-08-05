---
best-practices: https://www.anthropic.com/engineering/claude-code-best-practices
---

# General

- Ignore blocks starting with [IGNORE] and ending with [/IGNORE]
- Use existing project files as examples and references
- Consult with CONVENTIONS.md and for project-specific conventions
- When you want violate a convention, explain why, and ask for approval

# Workflow

- When you start a task:
  - Build a plan which adheres to the conventions
  - You may deviate from the conventions if it is necessary for the task
  - Explain why you deviate from the conventions
  - Ask for approval of the plan before you start coding

- After you finish a task:
  - run `make generate` to update the generated files
  - run `make lint` and fix any issues
  - run `make test-unit` to run the tests, fix any issues
  - run `make run` to start application and check if it works
    - if it doesn't work, fix the issues
    - if it works, run `make test-integration` to run integration tests
    - stop the application
  - give suggestions for improvements of conventions or this workflow

# Don't

- Never commit anything - commiting is done by the user
- Never write anywhere 'generated with claude code' or similar
- Avoid unnecessary comments, use comments to explain ambiguous code
