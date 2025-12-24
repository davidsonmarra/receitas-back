# Create Test

Create a Go unit test for the given HTTP handler.

## Requirements:

- Use `testing` + `httptest`
- Assert HTTP status code
- Assert JSON body contents
- Assert Content-Type is `application/json`
- Place test files under `/test`
- Filename must end with `*_test.go`
- Follow existing code style and project conventions

## Instructions:

1. Create the test file with the pattern: `/test/{handler_name}_test.go`
2. Import necessary packages: `testing`, `net/http`, `net/http/httptest`, `encoding/json`
3. Create test function following the pattern: `func Test{HandlerName}(t *testing.T)`
4. Use `httptest.NewRecorder()` to capture the response
5. Assert:
   - Status code using `rr.Code`
   - Content-Type header using `rr.Header().Get("Content-Type")`
   - JSON body by decoding `rr.Body`

## Example:

For testing `UsersHandler`:

- Create: `/test/users_handler_test.go`
- Function: `func TestUsersHandler(t *testing.T)`
- Validate all required assertions

Follow the same structure as existing tests in the `/test` directory.
