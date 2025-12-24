# Create Route

Create a new HTTP route in the Go backend project.

## Input:

- route path
- handler name

## Requirements:

- Create a handler file under `/internal/http/handlers`
- Follow the same style and structure as existing handlers
- Register the route in `/internal/http/routes/routes.go`
- The handler must return a JSON response
- Use logging when appropriate
- Keep the handler small and idiomatic

## Instructions:

1. Create the handler file with the pattern: `/internal/http/handlers/{handler_name}.go`
2. Implement the handler function following Go best practices
3. Use the `response.JSON()` helper for JSON responses
4. Add appropriate logging
5. Register the route in the `Setup()` function in `/internal/http/routes/routes.go`

## Example:

For a route `/users` with handler `UsersHandler`:

- Create: `/internal/http/handlers/users.go`
- Implement: `func UsersHandler(w http.ResponseWriter, r *http.Request)`
- Register: `r.Get("/users", handlers.UsersHandler)` in routes.go

Do not modify unrelated files.
