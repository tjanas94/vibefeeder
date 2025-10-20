# Testing Guide

This document describes how to run tests in the VibeFeeder project.

## Prerequisites

Ensure you have all dependencies installed:

```bash
task install-deps
```

## Test Types

VibeFeeder uses two types of tests:

1. **Unit Tests** - Go tests using the standard `testing` library and Testify
2. **E2E Tests** - End-to-end tests using Playwright

## Running All Tests

To run both unit and E2E tests:

```bash
task test
```

## Unit Tests

Unit tests are written in Go using the standard `testing` library with Testify for assertions and mocks.

### Run Unit Tests

```bash
task test:unit
```

This command runs all Go tests with:

- Verbose output (`-v`)
- Race condition detection (`-race`)
- Coverage report generation (`-coverprofile=coverage.out`)

### View Coverage Report

After running unit tests, you can view the coverage report in your browser:

```bash
task test:unit:coverage
```

This will:

1. Run all unit tests
2. Generate a coverage report
3. Open the HTML coverage report in your default browser

### Writing Unit Tests

Unit tests should be placed in `*_test.go` files alongside the code they test.

Example structure:

```
internal/
  shared/
    validator/
      validator.go
      validator_test.go
```

## E2E Tests

End-to-end tests use Playwright to test the full application flow in real browser (Chrome).

### Run E2E Tests

```bash
task test:e2e
```

This runs all Playwright tests in headless mode.

### Run E2E Tests in Headed Mode

To see the browser while tests run:

```bash
task test:e2e:headed
```

### Run E2E Tests in UI Mode

For interactive test development and debugging:

```bash
task test:e2e:ui
```

This opens Playwright's UI mode where you can:

- Run tests interactively
- See step-by-step execution
- Time-travel through test steps
- View test traces

### Debug E2E Tests

To debug tests with Playwright Inspector:

```bash
task test:e2e:debug
```

### View E2E Test Report

After running E2E tests, view the detailed HTML report:

```bash
task test:e2e:report
```

### E2E Test Structure

E2E tests are located in the `tests/e2e/` directory:

```
tests/
  e2e/
    fixtures/       # Test fixtures and setup
    pages/          # Page Object Model classes
    *.spec.ts       # Test specifications
```

### Test Configuration

Playwright configuration is defined in `playwright.config.ts` at the project root.

## Test Organization

### Unit Tests

- Located alongside source code in `*_test.go` files
- Use table-driven tests where appropriate
- Follow Go testing conventions
- Utilize Testify for assertions and mocks

### E2E Tests

- Located in `tests/e2e/` directory
- Use Page Object Model pattern for better maintainability
- Test files follow `*.spec.ts` naming convention
- Fixtures in `tests/e2e/fixtures/` for shared setup
- Page objects in `tests/e2e/pages/` for reusable page interactions

## Best Practices

### Unit Tests

1. **Test file naming**: Use `<filename>_test.go` convention
2. **Test function naming**: Use `Test<FunctionName>` convention
3. **Table-driven tests**: Use for testing multiple scenarios
4. **Error handling**: Test both success and error cases
5. **Edge cases**: Include boundary conditions and edge cases
6. **Race conditions**: Always run with `-race` flag

### E2E Tests

1. **Page Object Model**: Use page objects for better maintainability
2. **Fixtures**: Use fixtures for authentication and shared setup
3. **Selectors**: Prefer data-testid attributes over CSS selectors
4. **Waiting**: Use Playwright's auto-waiting instead of explicit waits
5. **Isolation**: Ensure tests can run independently
6. **Clean up**: Use fixtures to handle setup and teardown

## Continuous Integration

Tests are run automatically in the CI/CD pipeline using GitHub Actions. Both unit and E2E tests must pass before code can be merged.

## Troubleshooting

### Unit Tests Failing

1. Check test output for specific error messages
2. Run with verbose flag: `go test -v ./...`
3. Check for race conditions: `go test -race ./...`
4. Verify dependencies are up to date: `task install-deps`

### E2E Tests Failing

1. Check browser console logs in the test report
2. Run in headed mode to see visual feedback: `task test:e2e:headed`
3. Use UI mode for interactive debugging: `task test:e2e:ui`
4. Check screenshots and videos in test artifacts
5. Ensure the application is properly built before testing

### Common Issues

**Port already in use**

- Stop any running instances of the application
- Check for zombie processes: `pkill -f vibefeeder`

**Browser not installed**

- Install Playwright browsers: `npx playwright install`

**Dependencies missing**

- Reinstall dependencies: `task install-deps`

## Additional Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Playwright Documentation](https://playwright.dev/)
- [Playwright Best Practices](https://playwright.dev/docs/best-practices)
