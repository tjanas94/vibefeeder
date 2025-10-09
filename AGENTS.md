# AI Rules for VibeFeeder

## TECH STACK

- Go 1.25
- Echo v4
- Templ
- htmx
- Alpine.js
- Tailwind 4
- DaisyUI

## PROJECT STRUCTURE

When introducing changes to the project, always follow the directory structure below:

- `cmd/` - Contains the main application entry point and any subcommands
- `internal/app` - Contains application setup, routing, and middleware
- `internal/shared/database` - Contains database client
- `internal/shared/config` - Contains configuration loading and management logic
- `internal/shared/http` - Contains http server utils
- `internal/shared/view` - Contains shared layout, components and pages
- `internal/{feature}` - Contains feature-specific code, such as handlers, services, models and views for each domain area (e.g., `feed`, `user`, `summary`)

When modifying the directory structure, always update this section.

## TASK_AUTOMATION

**IMPORTANT**: Always use `task` (go-task) instead of running commands directly.

Use `task --list` to see available tasks. Common tasks:

- `task dev` - Development with hot-reload
- `task build` - Production build
- `task lint` - Run all linters
- `task fmt` - Format all code
- `task install-deps` - Install dependencies

## CODING_PRACTICES

### Guidelines for clean code

- Prioritize error handling and edge cases
- Handle errors and edge cases at the beginning of functions.
- Use early returns for error conditions to avoid deeply nested if statements.
- Place the happy path last in the function for improved readability.
- Avoid unnecessary else statements; use if-return pattern instead.
- Use guard clauses to handle preconditions and invalid states early.
- Implement proper error logging and user-friendly error messages.
- Consider using custom error types or error factories for consistent error handling.

### Guidelines for VERSION_CONTROL

#### GIT

- Use conventional commits to create meaningful commit messages
- Use feature branches with descriptive names
- Write meaningful commit messages that explain why changes were made, not just what
- Keep commits focused on single logical changes to facilitate code review and bisection
- Use interactive rebase to clean up history before merging feature branches
- Leverage git hooks to enforce code quality checks before commits and pushes

## FRONTEND

### Guidelines for STYLING

#### TAILWIND

- Use the @layer directive to organize styles into components, utilities, and base layers
- Implement Just-in-Time (JIT) mode for development efficiency and smaller CSS bundles
- Use arbitrary values with square brackets (e.g., w-[123px]) for precise one-off designs
- Leverage the @apply directive in component classes to reuse utility combinations
- Use Tailwind 4 CSS directives for customization: @theme, @plugin and @custom-variant
- Use component extraction for repeated UI patterns instead of copying utility classes
- Leverage the theme() function in CSS for accessing Tailwind theme values
- Implement dark mode with the dark: variant
- Use responsive variants (sm:, md:, lg:, etc.) for adaptive designs
- Leverage state variants (hover:, focus:, active:, etc.) for interactive elements

## BACKEND

### Guidelines for GO

#### ECHO

- Use the middleware system for cross-cutting concerns with proper ordering based on execution requirements
- Implement the context package for request-scoped values and proper cancellation propagation
- Use the validator package for request validation with custom validation rules
- Apply proper route grouping for related endpoints and consistent path prefixing
- Implement structured error handling with custom error types and appropriate HTTP status codes
- Use context timeouts for external service calls to prevent resource leaks

#### TEMPL

- Use view models to separate presentation logic from data fetching
- Keep components focused - delegate business logic to services layer
- Use Tailwind classes for styling, avoid inline styles or templ CSS components
