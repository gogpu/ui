# Contributing to gogpu/ui

> Thank you for your interest in contributing to **gogpu/ui**!

---

## Code of Conduct

This project follows the [Go Community Code of Conduct](https://go.dev/conduct). Be respectful, constructive, and inclusive.

---

## Getting Started

### Prerequisites

- **Go 1.25+** (latest stable recommended)
- **Git** for version control
- Familiarity with Go modules and testing

### Setup

```bash
# Clone the repository
git clone https://github.com/gogpu/ui.git
cd ui

# Verify setup
go build ./...
go test ./...
```

---

## How to Contribute

### Reporting Issues

Before opening an issue:

1. **Search existing issues** to avoid duplicates
2. **Check the roadmap** in [ROADMAP.md](ROADMAP.md)
3. **Provide context**: Go version, OS, reproduction steps

**Issue templates:**
- **Bug Report** — Unexpected behavior
- **Feature Request** — New functionality
- **Question** — Usage clarification

### Submitting Pull Requests

1. **Fork** the repository
2. **Create a branch** from `main`:
   ```bash
   git checkout -b feat/your-feature
   ```
3. **Make changes** following our code standards
4. **Write tests** for new functionality
5. **Run checks**:
   ```bash
   go fmt ./...
   go test ./...
   golangci-lint run
   ```
6. **Commit** with conventional messages
7. **Push** and open a Pull Request

---

## Development Standards

### Code Style

- **Formatting**: `gofmt` (enforced)
- **Linting**: `golangci-lint` with project config
- **Naming**: Follow [Go naming conventions](https://go.dev/doc/effective_go#names)

### Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

[optional body]
```

**Types:**
| Type | Purpose |
|------|---------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation |
| `refactor` | Code restructuring |
| `test` | Adding/updating tests |
| `chore` | Maintenance tasks |

**Examples:**
```
feat(widgets): add Button component
fix(layout): correct Flexbox alignment calculation
docs(readme): update installation instructions
```

### Testing

- **Unit tests** for all public functions
- **Coverage target**: 70%+
- **Table-driven tests** preferred

```go
func TestButton_Click(t *testing.T) {
    tests := []struct {
        name    string
        input   ButtonConfig
        wantErr bool
    }{
        // test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic...
        })
    }
}
```

---

## API Design Guidelines

### Backward Compatibility

We follow strict backward compatibility rules (see [docs/VERSIONING.md](docs/VERSIONING.md)):

1. **Functional Options** for configurable types:
   ```go
   func NewButton(text string, opts ...ButtonOption) *Button
   ```

2. **Interface Extension** for optional capabilities:
   ```go
   type Focusable interface {
       Widget
       Focus()
       Blur()
   }
   ```

3. **Config Structs** with zero-value defaults:
   ```go
   type ButtonConfig struct {
       Text     string
       Disabled bool  // zero value = enabled
   }
   ```

### Package Organization

| Package | Visibility | Can Change? |
|---------|------------|-------------|
| `core/` | Public | Stable |
| `widgets/` | Public | Stable |
| `internal/` | Private | Yes |
| `experimental/` | Public | Yes (unstable) |

---

## Pull Request Checklist

Before submitting:

- [ ] Code compiles without errors
- [ ] All tests pass
- [ ] New tests added for new functionality
- [ ] `go fmt` applied
- [ ] `golangci-lint run` passes
- [ ] Documentation updated if needed
- [ ] Commit messages follow convention
- [ ] No breaking changes to public API (or discussed first)

---

## Review Process

1. **Automated checks** must pass (CI)
2. **Maintainer review** for code quality
3. **Discussion** for design decisions
4. **Squash merge** to main

Typical review focuses on:
- Correctness and edge cases
- API design and consistency
- Performance implications
- Test coverage

---

## Areas for Contribution

### Good First Issues

Look for issues labeled `good first issue`:
- Documentation improvements
- Test coverage additions
- Small bug fixes

### Larger Contributions

For significant changes, **open an issue first** to discuss:
- New widgets
- Layout algorithms
- Theme implementations
- Accessibility features

---

## Questions?

- **GitHub Issues** — Technical questions
- **GitHub Discussions** — General discussion
- **ROADMAP.md** — Project direction

---

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).

---

*Thank you for helping make gogpu/ui better!*
