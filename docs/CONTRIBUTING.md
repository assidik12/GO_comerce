# Contributing to Catalyst

Thank you for your interest in contributing to Catalyst! Every contribution — bug reports,
feature requests, code improvements, or documentation fixes — is highly valued.

> **Before writing code**, please read:
> - [`docs/STANDARDS.md`](./STANDARDS.md) — coding standards that every PR must follow
> - [`docs/ARCHITECTURE.md`](./ARCHITECTURE.md) — architecture and tech stack overview
> - [`docs/CONCEPTS.md`](./CONCEPTS.md) — design decisions and rationale

---

## How Can I Help?

- 🐞 **Report Bugs** — something doesn't work as expected
- ✨ **Suggest Features** — new functionality or improvements
- 🔧 **Write Code** — fix bugs or implement features
- 📝 **Improve Documentation** — clarify docs, fix typos, add examples

---

## Reporting Bugs

1. **Check existing issues** — search [Issues](https://github.com/assidik12/catalyst/issues) first to avoid duplicates.
2. **Create a new issue** — include a clear title, detailed description, and steps to reproduce.
3. **Provide context** — Go version, OS, error logs, and expected vs actual behavior.

---

## Suggesting Features

Open a [GitHub Issue](https://github.com/assidik12/catalyst/issues) with:
- What problem does this solve?
- How should it work?
- Are there alternative approaches?

---

## Pull Request Workflow

### 1. Fork & Clone

```bash
git clone https://github.com/<your-username>/catalyst.git
cd catalyst
```

### 2. Create a Feature Branch

Always branch from `main`. Use descriptive names:

```bash
git checkout -b feat/add-payment-gateway
git checkout -b fix/stock-race-condition
git checkout -b docs/update-api-reference
```

### 3. Follow Coding Standards

All code must follow [`docs/STANDARDS.md`](./STANDARDS.md):

- **Import order**: stdlib → external → internal
- **Error handling**: use sentinel errors from `internal/domain/errors.go`
- **Naming**: Go conventions (PascalCase exported, camelCase unexported)
- **Testing**: every change must include at least one failure scenario test
- **Layer rules**: don't put business logic in handlers or SQL in services

### 4. Run Tests

```bash
# All tests
go test ./...

# With race detector (required for concurrency-related changes)
go test -race ./...

# Coverage report
go test -coverprofile=coverage.out ./...
```

### 5. Commit with Conventional Commits

```bash
# Format: <type>(<scope>): <description>
git commit -m "feat(transaction): add idempotency key support"
git commit -m "fix(product): prevent negative stock on concurrent decrement"
git commit -m "test(user): add missing failure scenarios for user service"
```

| Type | When |
|---|---|
| `feat` | New feature |
| `fix` | Bug fix |
| `refactor` | Code change without feature or bug fix |
| `test` | Adding or improving tests |
| `docs` | Documentation only |
| `chore` | Maintenance (deps, config, etc.) |
| `perf` | Performance improvement |

**Rules:**
- Description in **English**, **imperative mood** ("add", not "added")
- Max 72 characters on the first line
- Scope: domain/layer name (`product`, `transaction`, `handler`, etc.)

### 6. Push & Open PR

```bash
git push origin feat/add-payment-gateway
```

Then open a Pull Request on GitHub with:
- Clear description of what changed and why
- Link to related issue (if any)
- Screenshots for UI-related changes

---

## Language Convention

| Context | Language |
|---|---|
| PR description, discussion, review comments | Bahasa Indonesia |
| Code (variables, functions, structs) | English |
| Code comments (`//`) | English |
| Commit messages | English |
| File and directory names | English |

---

## Review Process

Every PR will be reviewed against:
1. **Clean Architecture compliance** — no cross-layer violations
2. **Error handling** — sentinel errors, no swallowed errors
3. **Concurrency safety** — transactions, singleflight, mutex where needed
4. **Security** — no SQL injection, no sensitive data in logs
5. **Test coverage** — at least one failure scenario per test file

PRs with **BLOCKER** issues will be rejected. See the full checklist in the
[Code Reviewer Agent](../ai_agent/review_agent.md).

---

## Questions?

Open a [Discussion](https://github.com/assidik12/catalyst/discussions) or reach out
via the issue tracker. We're happy to help!
