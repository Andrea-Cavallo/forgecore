# Go 1.24 Best Practices

## Project Structure

```
myproject/
├── cmd/            # Main entrypoints (one subdir per binary)
│   └── myapp/
│       └── main.go
├── internal/       # Private packages (not importable externally)
├── pkg/            # Public reusable packages
├── api/            # API definitions (protobuf, OpenAPI)
├── config/         # Configuration files
├── scripts/        # Build and utility scripts
└── go.mod
```

- Keep `main.go` minimal — delegate logic to packages
- Use `internal/` to enforce encapsulation within the module
- One package per directory; package name matches directory name

## Modules & Dependencies

```bash
go mod init github.com/yourorg/yourrepo
go mod tidy          # Remove unused dependencies
go mod vendor        # Vendor dependencies for reproducible builds
```

- Pin dependencies with exact versions in `go.sum`
- Avoid `replace` directives in production modules
- Use `go get -u ./...` with care; prefer explicit version upgrades
- Go 1.24: use `tool` directive in `go.mod` to manage tool dependencies

```go
// go.mod
tool (
    golang.org/x/tools/cmd/stringer v0.x.x
)
```

## Code Style

- Follow `gofmt` / `goimports` — non-negotiable, enforce via CI
- Max line length: 100 characters (soft limit)
- Use `go vet` and `staticcheck` as baseline linters
- Preferred linter: `golangci-lint` with `.golangci.yml` config

### Naming Conventions

| Kind | Style | Example |
|------|-------|---------|
| Package | lowercase, no underscores | `httputil` |
| Exported | PascalCase | `UserService` |
| Unexported | camelCase | `parseToken` |
| Constants | PascalCase or ALL_CAPS for truly global | `MaxRetries` |
| Interfaces | Noun or `-er` suffix | `Reader`, `UserStore` |
| Errors | `Err` prefix for sentinel values | `ErrNotFound` |
| Error types | `Error` suffix | `ValidationError` |

## Error Handling

```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("loading config: %w", err)
}

// Sentinel errors for callers to match
var ErrNotFound = errors.New("not found")

// Check with errors.Is / errors.As
if errors.Is(err, ErrNotFound) { ... }

// Custom error type
type ValidationError struct {
    Field   string
    Message string
}
func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed on %s: %s", e.Field, e.Message)
}
```

- Never ignore errors — use `_` only when intentional and documented
- Wrap with `%w` to preserve the error chain
- Return errors as the last return value
- Avoid `panic` in library code; reserve it for truly unrecoverable states

## Interfaces

```go
// Define interfaces where they are used (consumer side), not where implemented
type UserStore interface {
    Get(ctx context.Context, id string) (*User, error)
    Save(ctx context.Context, u *User) error
}

// Prefer small, focused interfaces
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

- Accept interfaces, return concrete types
- Keep interfaces small (1–3 methods where possible)
- Embed interfaces to compose larger ones

## Concurrency

```go
// Always pass context for cancellation
func FetchUser(ctx context.Context, id string) (*User, error) { ... }

// Use errgroup for concurrent work with error propagation
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error { return fetchA(ctx) })
g.Go(func() error { return fetchB(ctx) })
if err := g.Wait(); err != nil { ... }

// Protect shared state
type SafeCounter struct {
    mu sync.Mutex
    v  map[string]int
}
```

- Always propagate `context.Context` as first parameter
- Prefer channels for ownership transfer, mutexes for shared state
- Use `sync.Once` for one-time initialization
- Avoid `goroutine` leaks — always have a way to stop goroutines
- Use `go tool vet -race` / run tests with `-race` flag

## Testing

```go
// Table-driven tests
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"positive", 1, 2, 3},
        {"negative", -1, -2, -3},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            got := Add(tc.a, tc.b)
            if got != tc.want {
                t.Errorf("Add(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.want)
            }
        })
    }
}
```

- Use `testify/assert` or standard `testing` — be consistent
- Table-driven tests for all non-trivial functions
- Use `t.Parallel()` for independent tests
- Integration tests: build-tag `//go:build integration`
- Benchmarks: use `testing.B` and run with `-benchmem`
- Go 1.24: use `testing/synctest` for testing concurrent code

```bash
go test ./...                    # Run all tests
go test -race ./...              # Race detection
go test -cover ./...             # Coverage
go test -run TestFoo -v ./...    # Run specific test
```

## Performance

- Profile before optimizing: `pprof`, `go tool trace`
- Avoid premature allocation: reuse buffers with `sync.Pool`
- Prefer `strings.Builder` over `+` concatenation in loops
- Use `slices` and `maps` packages (stdlib since 1.21) for generic utilities
- Benchmark critical paths with `testing.B`

```go
// Preallocate slices when size is known
s := make([]int, 0, expectedLen)

// Use sync.Pool for short-lived allocations
var bufPool = sync.Pool{
    New: func() any { return new(bytes.Buffer) },
}
```

## Go 1.24 Specific Features

### Generic Type Aliases
```go
// Type aliases can now be generic
type Stack[T any] = []T
```

### Improved `encoding/json` (v2 via `encoding/json/v2`)
- Use `omitzero` tag option for zero-value omission
- Prefer the new `json.Marshal` / `json.Unmarshal` from v2 for stricter behavior

### Swiss Table Maps
- Maps now use Swiss Table internally — no code changes required
- Benefits: better cache performance on large maps

### Tool Dependencies in `go.mod`
```go
// go.mod
tool golang.org/x/tools/cmd/stringer
```
```bash
go tool stringer -type=MyEnum   // Run managed tools
```

### Timer Changes
- `time.Timer` and `time.Ticker` no longer require manual drain after `Stop()`
- Safe to call `Reset` without draining the channel first

## Security

- Validate all external input at system boundaries
- Use `crypto/rand` for random secrets, never `math/rand`
- Avoid `os/exec` with user-supplied input; prefer structured APIs
- Use `net/http` timeouts explicitly:

```go
client := &http.Client{
    Timeout: 10 * time.Second,
}
srv := &http.Server{
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
}
```

- Run `govulncheck ./...` in CI to detect known vulnerabilities

## Documentation

```go
// Package comment goes before package declaration
// Package auth provides JWT-based authentication utilities.
package auth

// Exported symbols must have doc comments starting with the symbol name
// NewService creates an AuthService with the provided config.
func NewService(cfg Config) *Service { ... }
```

- Every exported symbol needs a doc comment
- Use `go doc` and `pkg.go.dev` conventions
- Examples in `_test.go` files with `func Example...()` are testable

## CI Checklist

```bash
go build ./...
go vet ./...
go test -race ./...
golangci-lint run
govulncheck ./...
```
