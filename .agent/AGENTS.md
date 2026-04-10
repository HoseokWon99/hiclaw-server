# Repository Guidelines

## Project Structure & Module Organization
This repository is a small Go service rooted at [main.go](/Users/hoseok/mytools/hiclaw/hiclaw-server/main.go). Packages are organized by responsibility: `config/` loads environment-based settings, `core/` defines shared types and service/store interfaces, `service/` contains business logic, `storage/` implements persistence, `server/` exposes HTTP and WebSocket handlers, `hub/` manages broadcasts, `agent/` calls the external agent webhook, and `tailnet/` handles Tailscale discovery. Tests live beside the code they cover as `*_test.go`.

## Build, Test, and Development Commands
- `go run .` starts the server on `:8080` by default.
- `go build ./...` compiles all packages and catches interface or import regressions.
- `go test ./...` runs the full test suite.
- `gofmt -w .` formats all Go files before review.
- `go test ./... -cover` checks package coverage when touching service or storage logic.

Configuration is driven by env vars such as `HICLAW_HTTP_ADDR`, `HICLAW_SQLITE_PATH`, `HICLAW_AGENT_DEVICE_NAME`, and `HICLAW_AGENT_WEBHOOK_PATH`.

## Coding Style & Naming Conventions
Use standard Go formatting and keep files `gofmt`-clean. Indentation is tabs, not spaces, in `.go` files. Exported identifiers use `PascalCase`; unexported helpers use `camelCase`. Keep interface definitions in `core/`, concrete implementations in leaf packages, and constructor names in the `NewType` form such as `NewChatService` or `NewSQLiteDeviceStore`.

## Testing Guidelines
Use the standard `testing` package with table-driven tests where multiple cases share setup. Name tests `Test<Type>_<Behavior>` as in `TestSQLiteDeviceStore_RegisterAndList`. Prefer in-memory fixtures (`:memory:` SQLite, lightweight mocks) over external dependencies. At the moment, `go test ./...` can fail if required `go.sum` entries are missing; run `go mod tidy` before retrying when dependency metadata changes.

## Commit & Pull Request Guidelines
Recent history uses short `feat:` commits plus a few placeholder `m` commits; prefer the former. Write imperative, scoped messages like `feat: add websocket broadcast guard` or `fix: validate agent reply payload`. PRs should describe the behavioral change, list validation steps, link issues when relevant, and include request/response examples for API changes.

## Security & Configuration Tips
Do not commit `.env`, local database files, or machine-specific secrets. Use a local SQLite path for development, and treat Tailscale and agent endpoint settings as environment-specific configuration rather than hard-coded values.
