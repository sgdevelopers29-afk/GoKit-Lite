# Contributing to GoKit-Lite

Thank you for your interest in contributing to GoKit-Lite! This guide will help you get started with the contribution workflow, coding standards, and review process.

---

## Table of Contents

- [Getting Started](#getting-started)
- [Fork Workflow](#fork-workflow)
- [Branch Naming](#branch-naming)
- [Commit Conventions](#commit-conventions)
- [Pull Request Checklist](#pull-request-checklist)
- [Coding Standards](#coding-standards)
- [Documentation Standards](#documentation-standards)
- [Testing Requirements](#testing-requirements)
- [Review Process](#review-process)
- [Good First Issues](#good-first-issues)

---

## Getting Started

1. **Fork** the repository on GitHub.
2. **Clone** your fork locally:
   ```bash
   git clone https://github.com/<your-username>/GoKit-Lite.git
   cd GoKit-Lite
   ```
3. **Add the upstream remote:**
   ```bash
   git remote add upstream https://github.com/sgdevelopers29-afk/GoKit-Lite.git
   ```
4. **Install Go 1.25+** — see [go.dev/dl](https://go.dev/dl/).
5. **Verify your setup:**
   ```bash
   go test ./...
   ```

---

## Fork Workflow

All contributions follow the **fork-and-pull** model:

1. **Sync your fork** with upstream before starting new work:
   ```bash
   git checkout develop
   git fetch upstream
   git merge upstream/develop
   ```

2. **Create a feature branch** from `develop` (never from `main`):
   ```bash
   git checkout -b feature/my-feature develop
   ```

3. **Make your changes**, commit, and push to your fork:
   ```bash
   git push origin feature/my-feature
   ```

4. **Open a Pull Request** on GitHub, targeting the `develop` branch.

5. **Address review feedback**, push additional commits, and wait for approval.

> **Note:** The `main` branch is reserved for releases. All PRs must target `develop`.

---

## Branch Naming

Use descriptive, prefixed branch names:

| Prefix | Use Case | Example |
|--------|----------|---------|
| `feature/` | New functionality | `feature/redis-cache-adapter` |
| `fix/` | Bug fixes | `fix/validator-nil-pointer` |
| `docs/` | Documentation only | `docs/improve-auth-examples` |
| `refactor/` | Code restructuring (no behavior change) | `refactor/cache-generics` |
| `test/` | Test additions or improvements | `test/ratelimit-benchmarks` |
| `chore/` | Maintenance tasks (CI, deps, tooling) | `chore/github-actions-ci` |

**Rules:**
- Use lowercase with hyphens as separators.
- Keep names concise but descriptive.
- Include a related issue number when applicable: `fix/42-expired-token-panic`.

---

## Commit Conventions

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### Types

| Type | Description |
|------|-------------|
| `feat` | A new feature |
| `fix` | A bug fix |
| `docs` | Documentation changes only |
| `test` | Adding or updating tests |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `perf` | Performance improvement |
| `chore` | Maintenance (CI, deps, tooling) |

### Scope

Use the package name as the scope when applicable:

```
feat(auth): add token refresh support
fix(validator): handle nil pointer in nested struct validation
docs(cache): add TTL usage examples
test(monitor): add concurrent access benchmarks
```

### Rules

- Use the **imperative mood** ("add", "fix", "update" — not "added", "fixed", "updated").
- Keep the first line under **72 characters**.
- Reference issue numbers in the footer: `Closes #42`.
- One logical change per commit — avoid combining unrelated changes.

---

## Pull Request Checklist

Before submitting your PR, ensure you have completed the following:

- [ ] **Branch is up to date** with `develop`.
- [ ] **All tests pass** locally: `go test ./...`
- [ ] **Race detector passes:** `go test -race ./...`
- [ ] **Code is formatted:** `gofmt -w .`
- [ ] **Vet passes:** `go vet ./...`
- [ ] **New code has tests** with meaningful assertions.
- [ ] **GoDoc comments** are present on all exported types and functions.
- [ ] **No unrelated changes** are included in the diff.
- [ ] **PR description** clearly explains the change and links to related issues.
- [ ] **CHANGELOG.md** is updated for user-facing changes.

### PR Description Template

```markdown
## Summary

Brief description of the change.

## Motivation

Why is this change needed?

## Changes

- List of specific changes made.

## Testing

How was this tested?

## Related Issues

Closes #<issue-number>
```

---

## Coding Standards

### General

- Follow [Effective Go](https://go.dev/doc/effective_go) and the [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments).
- Run `gofmt` before every commit — unformatted code will not be merged.
- Run `go vet` to catch common mistakes.
- Keep functions focused and small. If a function exceeds ~50 lines, consider refactoring.
- Prefer returning `error` over panicking.

### Naming

- Use idiomatic Go names: `MixedCaps` for exported, `mixedCaps` for unexported.
- Avoid stuttering: `cache.New()` not `cache.NewCache()`.
- Use descriptive variable names — `err`, `ctx`, `ok` are acceptable for their standard uses.
- Acronyms should be all-caps: `UserID`, `HTTPHandler`, `JSONResponse`.

### Error Handling

- Return errors rather than logging and continuing.
- Use sentinel errors (`var ErrNotFound = errors.New(...)`) for well-known failure modes.
- Wrap errors with context using `fmt.Errorf("operation: %w", err)`.
- Document error conditions in GoDoc comments.

### Concurrency

- All public APIs that may be used concurrently **must** be goroutine-safe.
- Prefer `sync/atomic` over `sync.Mutex` where applicable (see the `monitor` package).
- Use `sync.RWMutex` when read operations significantly outnumber writes (see the `cache` package).
- Document thread-safety guarantees in GoDoc comments.

### Package Design

- One directory per package.
- Keep the public API surface minimal — export only what users need.
- Place related types in separate files when they grow large (e.g., `errors.go`, `types.go`).
- Follow the "default instance" pattern used by `auth` and `monitor` for convenience APIs.

---

## Documentation Standards

### GoDoc Comments

Every exported type, function, method, and package must have a GoDoc comment:

```go
// Validate inspects every exported field of data using struct tags and returns
// an error for the first rule violation it finds (fail-fast).
//
// data must be a struct or a non-nil pointer to a struct.
func Validate(data any) error {
```

**Rules:**
- Start with the name of the symbol being documented.
- Use complete sentences.
- Document parameters, return values, and error conditions.
- Include code examples in `example_test.go` files for complex APIs.

### Markdown Files

- Use ATX-style headings (`#`, `##`, `###`).
- Use fenced code blocks with language identifiers (` ```go `, ` ```bash `).
- Keep lines under 100 characters where practical.
- Use tables for structured comparisons.
- Add a table of contents for documents exceeding ~100 lines.

### Per-Package Documentation

Each package should have a corresponding `docs/<package>.md` with:
- Overview and purpose
- API reference with usage examples
- Error handling guide
- Thread-safety notes (if applicable)

---

## Testing Requirements

### Unit Tests

- Every package must have a `<package>_test.go` file.
- Test both success and failure paths.
- Use table-driven tests where appropriate.
- Test edge cases: nil inputs, empty strings, zero values, boundary conditions.
- Use `t.Helper()` in test helper functions.
- Avoid `time.Sleep` in tests — use channels, timeouts, or `testing.Short()`.

### Example Tests

- Provide `example_test.go` with runnable `Example*` functions for key APIs.
- These serve as both documentation and regression tests.

### Benchmarks

- Performance-sensitive code should have benchmarks in `<package>_benchmark_test.go`.
- Use `b.ReportAllocs()` to track allocations.
- Run benchmarks before and after changes: `go test -bench=. -benchmem ./...`

### Running Tests

```bash
# All tests
go test ./...

# With race detector
go test -race ./...

# Verbose output
go test -v ./...

# Specific package
go test ./validator/...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Review Process

1. **Automated checks** — CI must pass (formatting, vetting, tests, race detection).
2. **Maintainer review** — at least one maintainer must approve the PR.
3. **Feedback cycle** — address all review comments. Mark resolved conversations.
4. **Squash and merge** — PRs are squash-merged into `develop` with a clean commit message.

### What Reviewers Look For

- Correctness and completeness
- Test coverage for new and changed code
- GoDoc comments on exported symbols
- Adherence to existing patterns and conventions
- No unnecessary dependencies
- Backward compatibility (no breaking changes without discussion)

### Response Times

- Maintainers aim to provide initial review within **3 business days**.
- If you haven't received feedback within a week, feel free to ping the PR.

---

## Good First Issues

New to GoKit-Lite? Look for issues labeled [`good first issue`](https://github.com/sgdevelopers29-afk/GoKit-Lite/labels/good%20first%20issue):

- Improving error messages
- Adding runnable examples to `example_test.go` files
- Writing additional benchmarks
- Fixing documentation typos or improving clarity
- Adding missing test cases

---

## License

By contributing to GoKit-Lite, you agree that your contributions will be licensed under the [MIT License](LICENSE).

---

Thank you for contributing to GoKit-Lite! 🎉
