# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `BitMatrix` with column-major layout and configurable rows, cols, and chunk size
- `AtomicBitMatrix` for lock-free concurrent writes via CAS loops
- `MultiBitMatrix` with configurable bits per cell (1, 2, 4, 8, 16, 32)
- `AtomicMultiBitMatrix` for lock-free concurrent multi-bit writes via CAS loops
- Core operations: `Has`, `Set`, `Clear`, `Ensure`
- Multi-bit cell operations: `Get`, `Set` (with value), `Clear`, `Ensure` (with value)
- Multi-flag checks: `HasAll`, `HasAny` (works on both BitMatrix and MultiBitMatrix)
- Bulk column operations: `CountInCol`, `ClearCol`, `SetColChunk`
- Set operations across flags: `ColAnd`, `ColOr`, `ColAndNot`
- `Bitmap` result type with `Has`, `Count`, `ForEach`, and `Rows`
- `Grow` for dynamic row expansion
- Accessor methods: `Rows`, `Cols`, `BitsPerCell` on all matrix types
- Functional options: `WithRows`, `WithCols`, `WithChunkSize`
- MultiBitMatrix options: `WithMultiBitRows`, `WithMultiBitCols`, `WithMultiBitChunkSize`, `WithBitsPerCell`
- BitMatrix interfaces: `BitReader`, `BitWriter`, `FlagChecker`, `ColReader`, `ColWriter`, `Grower`, `ReadOnlyMatrix`, `Matrix`
- MultiBitMatrix interfaces: `CellReader`, `CellWriter`, `ReadOnlyMultiMatrix`, `MultiMatrix`
- `doc.go` with package-level documentation for pkg.go.dev
- `.golangci.yml` with strict linter configuration
- GitHub Actions CI workflow (test, lint, coverage across Go 1.21–1.23)
- `CONTRIBUTING.md` with development and PR guidelines
- Example tests for `New`, `Has`, `HasAll`, `ColAnd`, `NewAtomic`, `NewMultiBitMatrix`, `NewAtomicMultiBitMatrix`
- Fuzz tests: `FuzzSetHasRoundTrip`, `FuzzClearRoundTrip`, `FuzzMultiBitSetGetRoundTrip`
- Race-detection test for concurrent `Grow`
- Makefile targets: `test-race`, `lint`, `cover`, `fuzz`, `clean`
- README badges: Go Reference, Go Report Card, CI, License

### Fixed
- `WithRows(MaxRows)` and `WithCols(MaxCols)` were silently ignored due to exclusive upper bound check
