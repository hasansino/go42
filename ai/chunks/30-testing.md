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