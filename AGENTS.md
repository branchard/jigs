# AGENTS.md

## Project Overview

Jigs is a CLI tool written in Go for interactively managing `.env` files. It reads one or more template files (e.g., `.env.dist`, `.env.dev`), prompts the user for any variable values that are missing, and writes the result to `.env`. If `.env` already exists, only variables not yet defined in it are prompted.

## Repository Structure

```
cmd/jigs/main.go                — CLI entrypoint: argument parsing, orchestration
internal/dotenv/dotenv.go       — .env file parser and writer
internal/dotenv/dotenv_test.go  — Unit tests for the dotenv package
internal/prompt/prompt.go       — Interactive stdin/stdout prompt for variable values
internal/prompt/prompt_test.go  — Unit tests for the prompt package
e2e/e2e_test.go                 — End-to-end tests: build the binary and run it in temp dirs
.github/workflows/ci.yaml      — CI pipeline: test, build Docker image, publish to GHCR
```

- `cmd/jigs/main.go` is the only binary entrypoint. It wires together the `dotenv` and `prompt` packages.
- `internal/dotenv/` handles parsing `.env` files into ordered entries (key-value pairs, comments, blank lines), querying variables, mutating values, and serializing back to disk. It preserves file structure (comments, blank lines) through round-trips.
- `internal/prompt/` provides a single function `ForValue` that reads from an `io.Reader` and writes to an `io.Writer`, making it testable without a real terminal.

## Language and Build

- **Language**: Go (1.26+)
- **Module path**: `github.com/branchard/jigs`
- **No external dependencies** — stdlib only.
- **Build**: `go build -o ./build/jigs ./cmd/jigs` (or `make build`)
- **Test**: `go test ./...` (or `make test`)
- **Development shell**: `make dev` (runs a Docker container with the source mounted)

## Code Conventions

- Standard Go project layout: `cmd/` for binaries, `internal/` for private packages.
- Packages under `internal/` accept interfaces (`io.Reader`, `io.Writer`) rather than concrete types where possible, to support testing.
- Tests use the `testing` package with table-driven tests and `t.TempDir()` for file I/O.
- Errors are wrapped with `fmt.Errorf("context: %w", err)` for traceability.
- No third-party dependencies. Keep it that way unless there is a strong reason.

## How the Tool Works

1. Source files passed as CLI arguments are parsed in order. Variables are collected with first-occurrence-wins semantics for default values.
2. If `.env` exists, it is parsed and its variable keys are collected.
3. Any variable present in the source files but absent from `.env` is prompted interactively. Variables with a non-empty default show it in brackets (e.g., `DB_HOST [localhost]: `). Pressing Enter accepts the default.
4. Prompted values are appended to the existing `.env` (or a new one is created). Existing entries in `.env` are never modified or reordered.

## Testing

Run all tests:

```
go test ./...
```

Tests are colocated with source files (`_test.go` suffix, same package). The `dotenv` tests use temp files for parse/write round-trip verification. The `prompt` tests use `strings.Reader` and `bytes.Buffer` to simulate terminal I/O. The `e2e` tests compile the `jigs` binary, run it in isolated temp directories with piped stdin, and assert on stdout/stderr and the generated `.env` file content.

All tests must pass before merging any change.

## CI Pipeline

The GitHub Actions workflow (`.github/workflows/ci.yaml`) runs on pushes to `main`, version tags (`v*`), and pull requests against `main`. It has two jobs:

1. **test** — sets up Go and runs `go test ./...`.
2. **build-and-publish** — builds a multi-platform Docker image (`linux/amd64` and `linux/arm64`) and pushes it to GitHub Container Registry (`ghcr.io/branchard/jigs`). On pull requests the image is built but not pushed. Image tags are derived automatically from the Git ref (branch name, PR number, semver from tags, short SHA).

Authentication to GHCR uses the built-in `GITHUB_TOKEN` — no extra secrets are required.

## Common Tasks

### Adding a new .env syntax feature

Modify `parseLine` and/or `unquote` in `internal/dotenv/dotenv.go`. Add corresponding table-driven test cases in `dotenv_test.go`. If the feature affects serialization, also update `quoteIfNeeded` and the `Write` method, and verify with a round-trip test.

### Changing prompt behavior

Modify `ForValue` in `internal/prompt/prompt.go`. The function signature uses `io.Reader`/`io.Writer` — keep it that way so tests don't need a real terminal.

### Adding CLI flags or options

Modify `cmd/jigs/main.go`. The current implementation uses raw `os.Args` with no flag library. If flags are needed, prefer the stdlib `flag` package before reaching for a third-party library.
