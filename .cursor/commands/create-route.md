# Create Route

Create a new HTTP route in the Go backend project.

## Input:

- route path
- handler name
- operation type (read or write) - for rate limiting

## Requirements:

- Create a handler file under `/internal/http/handlers`
- Follow the same style and structure as existing handlers
- Register the route in `/internal/http/routes/routes.go`
- The handler must return a JSON response
- Use logging when appropriate
- Keep the handler small and idiomatic
- Apply appropriate rate limiting based on operation type

## Instructions:

1. Create the handler file with the pattern: `/internal/http/handlers/{handler_name}.go`
2. Implement the handler function following Go best practices
3. Use the `response.JSON()` helper for JSON responses
4. Add appropriate logging with `log.InfoCtx()`, `log.ErrorCtx()`, etc
5. Register the route in the `Setup()` function in `/internal/http/routes/routes.go`
6. Apply rate limiting middleware based on operation type:
   - **Read operations** (GET): Use `.With(customMiddleware.RateLimitRead(rateLimitConfig))`
   - **Write operations** (POST, PUT, DELETE): Use `.With(customMiddleware.RateLimitWrite(rateLimitConfig))`
   - **No specific limit**: Rely only on global rate limit (100/min)

## Rate Limiting Strategy:

- **Global Limit**: 100 requests/minute (applied to all endpoints)
- **Read Limit**: 60 requests/minute (GET endpoints)
- **Write Limit**: 20 requests/minute (POST, PUT, DELETE endpoints)

All limits are per IP address and configurable via environment variables.

## Example:

For a route `/users` with handler `UsersHandler`:

### Handler File: `/internal/http/handlers/users.go`

```go
package handlers

import (
    "net/http"
    "github.com/davidsonmarra/receitas-app/pkg/log"
    "github.com/davidsonmarra/receitas-app/pkg/response"
)

func ListUsers(w http.ResponseWriter, r *http.Request) {
    log.InfoCtx(r.Context(), "listing users")
    
    // Your logic here
    users := []string{"user1", "user2"}
    
    response.JSON(w, http.StatusOK, users)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
    log.InfoCtx(r.Context(), "creating user")
    
    // Your logic here
    
    response.JSON(w, http.StatusCreated, map[string]string{
        "message": "User created",
    })
}
```

### Register in `/internal/http/routes/routes.go`:

```go
// Inside Setup() function, after loading rateLimitConfig

// Users routes with rate limiting
r.Route("/users", func(r chi.Router) {
    // GET - read operation (60/min)
    r.With(customMiddleware.RateLimitRead(rateLimitConfig)).Get("/", handlers.ListUsers)
    
    // POST - write operation (20/min)
    r.With(customMiddleware.RateLimitWrite(rateLimitConfig)).Post("/", handlers.CreateUser)
})
```

## Notes:

- All routes already have global rate limiting (100/min) applied automatically
- Only add specific rate limits (Read/Write) for routes that need stricter control
- Health checks and test endpoints typically don't need specific rate limits
- Rate limiting is per IP address and considers proxy headers (X-Forwarded-For, X-Real-IP)

Do not modify unrelated files.
