📄 PRD — Go Backend (Phase 1)

1. Product Context

This project will establish the foundation for a backend service written in Go.
In this first phase, the goal is strictly to:

initialize the Go project

define the architecture

run a minimal HTTP server

expose a /test route returning “hello world”

include unit tests

provide Cursor commands for route & test generation

No business logic yet — only the core infrastructure.

2. Project Objectives
   🎯 Phase 1 Goals

Create a clean, scalable Go backend project structure

Implement an HTTP server

Add a /test endpoint

Define testing strategy and add unit tests

Automate developer workflows using Cursor commands

❌ Out of Scope (for now)

Database

Authentication & Authorization

Business rules

Deployment configuration

CI/CD

3. Tech Stack

Go ≥ 1.22

HTTP router: go-chi/chi

Tests: testing + httptest

Logging: Go standard library

Go modules

4. Project Architecture
   /cmd
   /api
   main.go -> API server entrypoint

/internal
/server
server.go -> HTTP server setup

/http
/routes
routes.go -> route registration
/handlers
test.go -> /test handler

/pkg
/response
json.go -> JSON response helpers

/test
test_handler_test.go -> unit tests for handlers

Architectural Principles

/cmd → executables

/internal → core app logic

/pkg → reusable utilities

Handlers are stateless & minimal

Clean separation of concerns

5. Initial Functionality
   Endpoint: /test

Method

GET /test

Response

{
"message": "hello world"
}

Acceptance Criteria

Response must be JSON

HTTP status 200

Header: Content-Type: application/json

6. Non-Functional Requirements

Idiomatic Go code

Expandable architecture

Unit tests required

Simple readable logging

gofmt formatting

No global mutable state

7. Unit Testing Requirements

Tests must:

use testing

use httptest

validate:

HTTP status

JSON body

content type

Expected Behavior
status == 200
message == "hello world"

Test Location
/test

Naming
\*\_test.go

8. Definition of Done (Phase 1)

Phase 1 is complete when:

Server runs via go run ./cmd/api

/test returns JSON “hello world”

go test ./... passes

Architecture matches spec

Cursor commands work

README exists

⚙️ 9. Cursor Commands
🛠 Command 1 — Create Route

Suggested Name

Create Route

Prompt

Create a new HTTP route in the Go backend project.

Input:

- route path
- handler name

Requirements:

- Create a handler file under /internal/http/handlers
- Follow the same style and structure as existing handlers
- Register the route in /internal/http/routes/routes.go
- The handler must return a JSON response
- Use logging when appropriate
- Keep the handler small and idiomatic

Do not modify unrelated files.

🧪 Command 2 — Create Test

Suggested Name

Create Test

Prompt

Create a Go unit test for the given HTTP handler.

Requirements:

- Use testing + httptest
- Assert HTTP status code
- Assert JSON body contents
- Assert Content-Type is application/json
- Place test files under /test
- Filename must end with \*\_test.go
- Follow existing code style and project conventions

🚀 10. Implementation Plan

Initialize Go module

Create folder structure

Implement HTTP server

Add /test handler

Register route

Write tests

Run tests

Create README

📘 11. Minimum README Requirements

Must include:

project description

Go version

how to run

how to test

folder overview

🌱 12. Future Roadmap

RESTful endpoints

Database layer

Authentication & Authorization

Migrations

Observability

CI/CD

Docker

API documentation
