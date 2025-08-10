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