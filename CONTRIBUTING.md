# Contributing to bitmatrix

Thanks for your interest in contributing! Here's how to get started.

## Development

### Prerequisites

- Go 1.21+
- [golangci-lint](https://golangci-lint.run/welcome/install/) (for linting)

### Running tests

```bash
make test
```

### Running benchmarks

```bash
make bench          # all benchmarks
make bench-core     # core operation benchmarks only
make bench-sizes    # size-scaling benchmarks only
```

### Linting

```bash
make lint
```

### Coverage

```bash
make cover
```

## Pull Request Guidelines

1. **Tests required** — every PR must include tests for new or changed behavior.
2. **Benchmarks for performance changes** — if your PR affects hot paths, include before/after benchmark results.
3. **One concern per PR** — keep PRs focused on a single change.
4. **Code style** — follow existing patterns and run `gofmt`. The codebase avoids external dependencies; keep it that way.

## Reporting Issues

Open an issue on GitHub with:
- What you expected to happen
- What actually happened
- Minimal reproduction steps (ideally a failing test)
