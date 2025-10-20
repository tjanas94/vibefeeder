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
- `internal/app/` - Application setup, routing, and middleware configuration
- `internal/shared/` - Shared utilities and infrastructure
  - `ai/` - AI client integration for generating summaries
  - `auth/` - Authentication and authorization utilities
  - `config/` - Configuration loading and validation (env vars, .env support)
  - `database/` - Database client with Supabase types and queries
  - `errors/` - Custom error types and error handling utilities
  - `logger/` - slog-based structured logging with Echo middleware integration
  - `models/` - Shared domain models used across features
  - `ssrf/` - SSRF protection validator for HTTP requests (blocks private IPs, localhost, cloud metadata)
  - `validator/` - Request validation utilities
  - `view/` - Templ renderer (echo.Renderer implementation) and shared components
- `internal/{feature}/` - Feature-specific code organized by domain
  - `dashboard/` - Dashboard feature (handler, models, views)
  - `feed/` - Feed management (handler, service, repository, models, views, constants)
  - `fetcher/` - RSS/Atom feed fetching service (service, repository, calculations)
  - `static/` - Static file serving (currently empty)
  - `summary/` - Content summarization (handler, service, models, views, prompts, errors)
  - Each feature may contain: handlers, services, repositories, models, views, and feature-specific utilities
- `web/` - Frontend assets
  - `styles/` - Tailwind CSS and styling files

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
- Separate pure functions from side-effecting functions.
