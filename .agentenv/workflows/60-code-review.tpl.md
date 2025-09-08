### Code Review

Comprehensive code evaluation for quality, security, and maintainability.

#### Phase 1: Preparation
- **Context understanding**: Purpose of changes
- **Requirements review**: What should be achieved
- **Scope assessment**: Size and impact of changes
- **History check**: Related previous changes

#### Phase 2: Structure Review
- **Architecture fit**: Alignment with design
- **Module organization**: Proper separation
- **Dependencies**: Appropriate and minimal
- **Naming conventions**: Clear and consistent

#### Phase 3: Code Quality
- **Readability**: Clear and self-documenting
- **Complexity**: Simple as possible
- **DRY principle**: No unnecessary duplication
- **SOLID principles**: Proper abstraction

#### Phase 4: Functionality
- **Correctness**: Logic is sound
- **Edge cases**: Handled appropriately
- **Error handling**: Comprehensive and graceful
- **Performance**: No obvious bottlenecks

#### Phase 5: Security
- Input validation and sanitization
- Authentication and authorization
- No hardcoded secrets
- SQL injection prevention
- XSS protection

#### Phase 6: Testing
- **Coverage**: Adequate test cases
- **Quality**: Tests actually test behavior
- **Edge cases**: Boundary conditions covered
- **Maintainability**: Tests are clear

#### Review Checklist
- [ ] Code compiles without warnings
- [ ] All tests pass
- [ ] Documentation updated
- [ ] No commented-out code
- [ ] No debug statements
- [ ] Consistent formatting
- [ ] Clear commit messages

#### Feedback Guidelines
- **Be constructive**: Suggest improvements
- **Be specific**: Point to exact lines
- **Provide examples**: Show better approaches
- **Acknowledge good**: Highlight positives
- **Prioritize**: Distinguish must-fix from nice-to-have

#### Severity Levels
- **🔴 Critical**: Must fix before merge
- **🟡 Major**: Should fix before merge
- **🔵 Minor**: Can fix in follow-up
- **💡 Suggestion**: Consider for improvement