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

- `cmd/vibefeeder/` - Application entry point (main.go with bootstrap and graceful shutdown)
- `cmd/gen-asset-manifest/` - Asset manifest generation tool for cache busting
- `internal/app/` - Application setup, routing, and middleware configuration
- `internal/container/` - Dependency injection container
- `internal/shared/` - Shared utilities and infrastructure
  - `ai/` - AI client integration for generating summaries
  - `assets/` - Asset manifest and cache busting utilities
  - `auth/` - Authentication infrastructure (middleware, session management, user session types)
  - `config/` - Configuration loading and validation (env vars, .env support)
  - `csrf/` - CSRF protection utilities
  - `database/` - Database client with Supabase types and queries
  - `errors/` - Custom error types and error handling utilities
  - `events/` - Event logging infrastructure
  - `logger/` - slog-based structured logging with Echo middleware integration
  - `models/` - Shared domain models used across features
  - `ssrf/` - SSRF protection validator for HTTP requests (blocks private IPs, localhost, cloud metadata)
  - `validator/` - Request validation utilities
  - `view/` - Templ renderer (echo.Renderer implementation) and shared components
- `internal/{feature}/` - Feature-specific code organized by domain
  - `auth/` - Authentication feature (handler, models, views)
  - `dashboard/` - Dashboard feature (handler, models, views)
  - `feed/` - Feed management (handler, service, repository, models, views, constants)
  - `fetcher/` - RSS/Atom feed fetching service (service, repository, calculations)
  - `summary/` - Content summarization (handler, service, models, views, prompts, errors)
  - Each feature may contain: handlers, services, repositories, models, views, and feature-specific utilities
- `web/` - Frontend assets
  - `icons/` - Application icons and favicons
  - `styles/` - Tailwind CSS and styling files

When modifying the directory structure, always update this section.

## TASK_AUTOMATION

**IMPORTANT**: Always use `task` (go-task) instead of running commands directly.

Use `task --list` to see available tasks. Common tasks:

- `task dev` - Development with hot-reload
- `task build` - Production build
- `task test:unit` - Run Go unit tests
- `task lint` - Run all linters
- `task fmt` - Format all code
- `task install-deps` - Install dependencies

## TESTING

**IMPORTANT FOR AI AGENTS**:

- ALWAYS run `task test:unit` after making code changes
- NEVER run `task test:e2e` - E2E tests are for CI/CD only
- Unit tests are fast and should be run frequently during development
- Write tests for new functions and modified logic

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
- Separate pure functions from side-effecting functions.
