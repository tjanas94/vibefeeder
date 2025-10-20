## BACKEND

### Guidelines for GO

#### ECHO

- Use the middleware system for cross-cutting concerns with proper ordering based on execution requirements
- Implement the context package for request-scoped values and proper cancellation propagation
- Use the validator package for request validation with custom validation rules
- Apply proper route grouping for related endpoints and consistent path prefixing
- Implement structured error handling with custom error types and appropriate HTTP status codes
- Use context timeouts for external service calls to prevent resource leaks
