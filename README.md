# bitmatrix

[![Go Reference](https://pkg.go.dev/badge/github.com/AdiSaripuloh/bitmatrix.svg)](https://pkg.go.dev/github.com/AdiSaripuloh/bitmatrix)
[![Go Report Card](https://goreportcard.com/badge/github.com/AdiSaripuloh/bitmatrix)](https://goreportcard.com/report/github.com/AdiSaripuloh/bitmatrix)
[![CI](https://github.com/AdiSaripuloh/bitmatrix/actions/workflows/ci.yml/badge.svg)](https://github.com/AdiSaripuloh/bitmatrix/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Zero-overhead 2D bit matrix for flagging and small-value storage at scale.**

Answer "does entity X have flag Y?" in **sub-nanosecond to 10 nanoseconds** with zero allocations, zero GC pressure, and zero dependencies.

```go
bm := bitmatrix.New(
    bitmatrix.WithRows(1_000_000),
    bitmatrix.WithCols(1024),
)

bm.Set(57403, 42)                // grant flag 42 to entity 57403
bm.Has(57403, 42)                // true — 0.7 ns, one division + one shift

bm.ColAnd(42, 99)                // entities with BOTH flag 42 and 99
bm.CountInCol(42)                // how many entities have flag 42
bm.ClearCol(42)                  // revoke flag 42 from everyone — ~1 μs
```

Need more than one bit per cell? Use `MultiBitMatrix` for small integer values:

```go
m := bitmatrix.NewMultiBitMatrix(
    bitmatrix.WithBitsPerCell(4),            // 4-bit cells (0–15)
    bitmatrix.WithMultiBitRows(1_000_000),
)
m.Set(42, 0, 15)                 // store 4-bit value — 2.3 ns
m.Get(42, 0)                     // 15 — 1.3 ns
```

## Why

Most Go cache and bitmap libraries are either **1D** (one bitset per flag, no native matrix operations), **compressed** (overhead per check for dense data), or a **full database engine** (transactions, schemas, many dependencies for one bit check).

bitmatrix is none of those. It's a single data structure that maps `(row, col) → bits` using pure arithmetic on contiguous memory.

## Install

```bash
go get github.com/AdiSaripuloh/bitmatrix
```

Requires Go 1.21+. Zero external dependencies.

## How It Works

### Memory Layout

bitmatrix stores data in **column-major** order. Each flag (column) owns a contiguous flat slice of `uint64` words.

For `BitMatrix`, each `uint64` holds exactly 64 entities (1 bit per entity):

```
Flag 0:   [ word 0 ][ word 1 ][ word 2 ] ... [ word N ]
Flag 1:   [ word 0 ][ word 1 ][ word 2 ] ... [ word N ]
...
Flag 1023:[ word 0 ][ word 1 ][ word 2 ] ... [ word N ]

Words are grouped into chunks for bulk operations:
Each chunk = ChunkSize entities = ChunkSize/64 uint64 words = 1 KB (default)
```

For `MultiBitMatrix`, each `uint64` holds `64 / bitsPerCell` cells. For example, with 4-bit cells each word holds 16 cells.

### Why This Is Fast

`Has(57403, 42)` compiles down to two arithmetic operations:

```
wordIdx = 57403 / BitsPerWord   // = 57403 >> 6  → which uint64 word
bitPos  = 57403 % BitsPerWord   // = 57403 & 63  → which bit within that word

return data[42][wordIdx] >> bitPos & 1 != 0
```

Two instructions. No branches. No pointer chasing. No hash computation. No lock acquisition. The CPU prefetcher can predict every memory access because the data is contiguous.

### Why Column-Major

Permission/flag checks cluster by flag, not by entity. A middleware checking `can_read` runs that check for thousands of different users. Column-major layout keeps `flag[42]`'s data hot in L1/L2 cache across those checks.

```
Row-major:  check flag 42 for user 0     → cache line at offset 0
            check flag 42 for user 57403 → cache line at offset 7.3 MB (MISS)

Col-major:  check flag 42 for user 0     → chunk 0 of flag 42 (1 KB)
            check flag 42 for user 57403 → chunk 7 of flag 42 (1 KB, likely cached)
```

## API

### Construction

```go
// BitMatrix — 1 bit per cell (boolean flags).
bm := bitmatrix.New()
bm := bitmatrix.New(
    bitmatrix.WithRows(500_000),
    bitmatrix.WithCols(256),
    bitmatrix.WithChunkSize(4096),
)

// Lock-free atomic variant for concurrent writes.
abm := bitmatrix.NewAtomic(
    bitmatrix.WithRows(1_000_000),
    bitmatrix.WithCols(64),
)

// MultiBitMatrix — configurable bits per cell (1, 2, 4, 8, 16, 32).
m := bitmatrix.NewMultiBitMatrix(
    bitmatrix.WithBitsPerCell(4),
    bitmatrix.WithMultiBitRows(1_000_000),
    bitmatrix.WithMultiBitCols(256),
)

// Lock-free atomic variant.
am := bitmatrix.NewAtomicMultiBitMatrix(
    bitmatrix.WithBitsPerCell(4),
    bitmatrix.WithMultiBitRows(1_000_000),
)
```

### Options

**BitMatrix / AtomicBitMatrix:**

| Option | Default | Constraint |
|---|---|---|
| `WithRows(n uint32)` | `1 << 20` (1M) | `[MinRows, MaxRows]` |
| `WithCols(n uint32)` | `1024` | `[MinCols, MaxCols]` |
| `WithChunkSize(n uint32)` | `8192` | `>= 64`, multiple of `BitsPerWord` |

**MultiBitMatrix / AtomicMultiBitMatrix:**

| Option | Default | Constraint |
|---|---|---|
| `WithMultiBitRows(n uint32)` | `1 << 20` (1M) | `[MinRows, MaxRows]` |
| `WithMultiBitCols(n uint32)` | `1024` | `[MinCols, MaxCols]` |
| `WithMultiBitChunkSize(n uint32)` | `8192` | `>= 64`, multiple of `BitsPerWord` |
| `WithBitsPerCell(n uint32)` | `2` | Power of 2 dividing 64: `1, 2, 4, 8, 16, 32` |

All options silently ignore values that violate their constraints, keeping the default.

### Constants

```go
bitmatrix.BitsPerWord   // 64   — bits per uint64 word
bitmatrix.ChunkSize     // 8192 — default entities per chunk
bitmatrix.MaxRows       // 1 << 20
bitmatrix.MaxCols       // 1024
bitmatrix.MinRows       // 1
bitmatrix.MinCols       // 1
bitmatrix.MinChunkSize  // 64
bitmatrix.MaxBitsPerCell // 32
```

### Interfaces

Interfaces follow Go convention: small, single-responsibility, composed into larger ones.

**BitMatrix interfaces:**

```go
// Primitive interfaces — accept only what you need.
type BitReader   interface { Has(row, col uint32) bool }
type BitWriter   interface { Set(row, col uint32); Clear(row, col uint32) }
type FlagChecker interface { HasAll(row uint32, cols ...uint32) bool; HasAny(row uint32, cols ...uint32) bool }
type ColReader   interface { CountInCol(col uint32) int; ColAnd(colA, colB uint32) *Bitmap; ColOr(colA, colB uint32) *Bitmap; ColAndNot(colA, colB uint32) *Bitmap }
type ColWriter   interface { ClearCol(col uint32); SetColChunk(col, chunk uint32) }
type Grower      interface { Grow(newRows uint32) }

// Composed interfaces.
type ReadOnlyMatrix interface { BitReader; FlagChecker; ColReader }
type Matrix         interface { ReadOnlyMatrix; BitWriter; ColWriter; Grower; Ensure(row, col uint32) }
```

Both `*BitMatrix` and `*AtomicBitMatrix` implement `Matrix` and `ReadOnlyMatrix`.

**MultiBitMatrix interfaces:**

```go
// Primitive interfaces.
type CellReader interface { Get(row, col uint32) uint64 }
type CellWriter interface { Set(row, col uint32, val uint64); Clear(row, col uint32) }

// Composed interfaces.
type ReadOnlyMultiMatrix interface { CellReader; FlagChecker }
type MultiMatrix         interface { ReadOnlyMultiMatrix; CellWriter; Grower; Ensure(row, col uint32, val uint64) }
```

Both `*MultiBitMatrix` and `*AtomicMultiBitMatrix` implement `MultiMatrix` and `ReadOnlyMultiMatrix`.

**Use the narrowest interface that satisfies your function:**

```go
// Only checks permissions — no mutation needed.
func canRead(m BitReader, userID uint32) bool {
    return m.Has(userID, PermRead)
}

// Reporting — reads columns but never writes.
func report(m ColReader, col uint32) {
    fmt.Println(m.CountInCol(col))
}

// Full access — entity ID may exceed current bounds.
func provision(m Matrix, userID uint32) {
    m.Ensure(userID, PermRead)
}

// Read a multi-bit value — works with both MultiBitMatrix and AtomicMultiBitMatrix.
func getLevel(m CellReader, userID, col uint32) uint64 {
    return m.Get(userID, col)
}
```

### Accessor Methods

All matrix types expose their dimensions via getter methods:

```go
bm.Rows()        // current row capacity (grows with Grow/Ensure)
bm.Cols()        // number of columns
m.BitsPerCell()  // bits per cell (always 1 for BitMatrix)
bitmap.Rows()    // row range of the source matrix
```

### BitMatrix Operations

```go
func (bm *BitMatrix) Has(row, col uint32) bool   // 0.7 ns
func (bm *BitMatrix) Set(row, col uint32)         // 2.0 ns — row must be within bounds
func (bm *BitMatrix) Clear(row, col uint32)       // 1.9 ns — row must be within bounds
func (bm *BitMatrix) Ensure(row, col uint32)      // 2.0 ns within bounds; grows if row >= rows
```

`Set` and `Clear` require the row to be within the current matrix bounds. Use `Ensure` when
the entity ID is not guaranteed to fit — it grows the matrix automatically if needed.

### Multi-Flag Checks

```go
// Does entity have ALL of these flags?
func (bm *BitMatrix) HasAll(row uint32, cols ...uint32) bool

// Does entity have ANY of these flags?
func (bm *BitMatrix) HasAny(row uint32, cols ...uint32) bool
```

`HasAll` and `HasAny` are also available on `MultiBitMatrix` and `AtomicMultiBitMatrix`, where they check for non-zero cell values.

### Bulk Column Operations

These operate on an entire flag across all entities. Because each flag is a contiguous slice, these are sequential memory scans — the CPU prefetcher's best case.

```go
func (bm *BitMatrix) CountInCol(col uint32) int          // popcount across all words
func (bm *BitMatrix) ClearCol(col uint32)                // zero all words in column
func (bm *BitMatrix) SetColChunk(col, chunk uint32)      // fill one chunk (ChunkSize entities)
```

`CountInCol` and `ClearCol` are also available on `MultiBitMatrix` and `AtomicMultiBitMatrix`.

### Set Operations Across Flags

```go
func (bm *BitMatrix) ColAnd(colA, colB uint32) *Bitmap    // entities with BOTH flags
func (bm *BitMatrix) ColOr(colA, colB uint32) *Bitmap     // entities with EITHER flag
func (bm *BitMatrix) ColAndNot(colA, colB uint32) *Bitmap // entities with colA but NOT colB
```

### Bitmap

```go
func (b *Bitmap) Has(row uint32) bool          // is this entity set?
func (b *Bitmap) Count() int                   // total set entities
func (b *Bitmap) ForEach(f func(row uint32))   // iterate set entities
func (b *Bitmap) Rows() uint32                 // row range from source matrix
```

### Growth

```go
// Grow expands to newRows row capacity. newRows is a count, not an index —
// to make index row valid, call Grow(row+1). Existing bits are untouched.
// Note: Grow does not enforce MaxRows.
func (bm *BitMatrix) Grow(newRows uint32)

// Ensure is the safe alternative to Set: grows to fit row if needed, then sets.
// Prefer this over calling Grow + Set manually.
func (bm *BitMatrix) Ensure(row, col uint32)
```

### MultiBitMatrix Operations

`MultiBitMatrix` stores a configurable number of bits per cell (1, 2, 4, 8, 16, or 32) instead of a single boolean flag. Cell values never span two `uint64` words because `bitsPerCell` is restricted to powers of 2 that evenly divide 64.

```go
func (m *MultiBitMatrix) Get(row, col uint32) uint64              // 1.3 ns
func (m *MultiBitMatrix) Set(row, col uint32, val uint64)         // 2.3 ns — truncates to bitsPerCell
func (m *MultiBitMatrix) Clear(row, col uint32)                   // sets cell to zero
func (m *MultiBitMatrix) Ensure(row, col uint32, val uint64)      // grows if row >= rows, then sets
func (m *MultiBitMatrix) CountInCol(col uint32) int               // non-zero cells in column
```

`Set` silently truncates `val` to the cell's bit width. For example, with 4-bit cells, `Set(row, col, 0xFF)` stores `0xF`.

## Use Cases

**Authorization** — "Can user X perform action Y?"

```go
perms := bitmatrix.New(bitmatrix.WithRows(1_000_000), bitmatrix.WithCols(1024))
perms.Set(userID, PermReadDocs)

if perms.Has(userID, PermReadDocs) {
    // allow
}
```

**Feature Flags** — "Is feature Y enabled for user X?"

```go
features := bitmatrix.New(bitmatrix.WithRows(1_000_000), bitmatrix.WithCols(256))

// Roll out to first 8192 users (one chunk)
features.SetColChunk(FeatureNewUI, 0)

// Kill switch
features.ClearCol(FeatureNewUI)
```

**Gaming Achievements** — "Has player X unlocked achievement Y?"

```go
achievements := bitmatrix.New(bitmatrix.WithRows(1_000_000), bitmatrix.WithCols(512))
achievements.Set(playerID, AchievFirstBlood)

// "How many players unlocked this?"
count := achievements.CountInCol(AchievFirstBlood)
```

**RBAC / Permission Levels** — "What access level does user X have on resource Y?"

```go
// 2-bit cells: 0=none, 1=read, 2=write, 3=admin
acl := bitmatrix.NewMultiBitMatrix(
    bitmatrix.WithBitsPerCell(2),
    bitmatrix.WithMultiBitRows(1_000_000),
    bitmatrix.WithMultiBitCols(1024),
)
acl.Set(userID, resourceID, 3) // admin access
level := acl.Get(userID, resourceID)
```

**IoT Capabilities** — "Does device X support capability Y?"

```go
caps := bitmatrix.New(bitmatrix.WithRows(1_000_000), bitmatrix.WithCols(128))

// "All devices with Bluetooth AND WiFi"
both := caps.ColAnd(CapBluetooth, CapWiFi)
both.ForEach(func(deviceID uint32) { /* ... */ })
```

**Search Pre-filtering** — "Which documents have tag Y?"

```go
tags := bitmatrix.New(bitmatrix.WithRows(1_000_000), bitmatrix.WithCols(1024))

// Narrow search space before expensive vector similarity
candidates := tags.ColAnd(TagGolang, TagPerformance)
candidates.ForEach(func(docID uint32) { /* feed to vector search */ })
```

## Memory Usage

**BitMatrix** (1 bit per cell):

```
Memory = ceil(rows / ChunkSize) × ChunkSize / 8 × cols

Examples:
  1M entities × 1K flags    = 128 MB
  1M entities × 256 flags   = 32 MB
  500K entities × 128 flags = 8 MB
  100K entities × 64 flags  = 0.8 MB
```

**MultiBitMatrix** (N bits per cell):

```
Memory = ceil(rows / ChunkSize) × ceil(ChunkSize / (64/bitsPerCell)) × 8 × cols

Examples (bitsPerCell=4):
  1M entities × 1K cols     = 512 MB
  1M entities × 256 cols    = 128 MB
  100K entities × 64 cols   = 3.2 MB
```

No per-entity metadata. No pointers. No hash tables. Just bits.

## Concurrency

**Reads are safe without locks.** `Has()` and `Get()` perform a single memory load (one `uint64` read) and a bit mask. No locks, no CAS, no contention.

**Writes require synchronization.** `Set()` and `Clear()` modify a `uint64` word shared by multiple entities. Two options:

```go
// Option A: External lock (simplest, fine for infrequent writes)
mu.Lock()
bm.Set(entity, flag)
mu.Unlock()

// Option B: Atomic variant (lock-free writes, slightly slower)
abm := bitmatrix.NewAtomic(bitmatrix.WithRows(1_000_000), bitmatrix.WithCols(1024))
abm.Set(entity, flag)    // atomic OR via CAS loop
abm.Clear(entity, flag)  // atomic AND-NOT via CAS loop
abm.Has(entity, flag)    // atomic Load
```

The same pattern applies to `MultiBitMatrix`:

```go
am := bitmatrix.NewAtomicMultiBitMatrix(
    bitmatrix.WithBitsPerCell(4),
    bitmatrix.WithMultiBitRows(1_000_000),
)
am.Set(entity, col, val)  // atomic CAS loop
am.Get(entity, col)       // atomic Load
```

For the common auth/flagging pattern (read-heavy, write-rare), Option A with a `sync.RWMutex` is ideal.

### Grow / Ensure Caveat

`Grow()` and `Ensure()` reallocate the underlying memory and are **not atomic**, even on `AtomicBitMatrix` and `AtomicMultiBitMatrix`. If concurrent `Grow` or `Ensure` calls are possible, protect them with an external lock:

```go
mu.Lock()
abm.Grow(newSize)
mu.Unlock()
```

Once the matrix is large enough for your data, all `Has`/`Set`/`Clear`/`Get` calls on atomic variants remain lock-free.

## When to Use Which Type

| Question | Type |
|---|---|
| "Does entity X have flag Y?" (yes/no) | `BitMatrix` |
| "What is entity X's level/state for Y?" (small int) | `MultiBitMatrix` |
| Concurrent writes without external lock? | `AtomicBitMatrix` / `AtomicMultiBitMatrix` |
| Read-heavy, write-rare with external lock? | `BitMatrix` / `MultiBitMatrix` |

## When NOT to Use bitmatrix

- **Sparse data** — If only 100 out of 1 billion entities have a flag, you're wasting memory. A compressed bitmap is a better fit.
- **Unknown/unbounded key space** — If you don't know how many entities you'll have, a hash-based cache is more appropriate.
- **Need eviction/TTL** — bitmatrix holds everything. If you need to expire data, use an LRU cache.
- **Need rich queries** — If you need SQL-like filtering across mixed types, use a columnar store or database.
- **Row-major access pattern** — If your hot path is "list all flags for entity X" rather than "check flag Y for entity X", the column-major layout is the wrong orientation.
- **Large per-cell values** — `MultiBitMatrix` supports up to 32 bits per cell. For larger values, use a regular slice or map.

## Design Decisions

**`BitsPerWord = 64`** — Each `uint64` stores 64 entities. The word index is `row / BitsPerWord` and the bit position is `row % BitsPerWord`. Both reduce to single shift/AND instructions since `BitsPerWord` is a compile-time constant.

**Dynamic chunk size** — `ChunkSize` defaults to 8192 but is configurable via `WithChunkSize`. Smaller chunks (e.g. 4096) keep more data in L1 for sparse access patterns. Larger chunks (e.g. 16384) reduce overhead for dense bulk operations. `wordIdx` remains `row / BitsPerWord` regardless of chunk size — chunk boundaries only affect bulk operations (`SetColChunk`, `Grow`).

**Column-major layout** — Flag checks cluster by flag ID. A middleware checking one permission does so for many different entities. Column-major keeps that flag's data hot.

**No compression** — Compressed bitmap formats use adaptive containers and dispatch on every access. For dense data the dispatch overhead exceeds the memory savings. bitmatrix assumes dense data and skips the dispatch entirely.

**No eviction** — Eviction adds 48+ bytes of metadata per entity and causes cache misses when evicted entities return. For bounded populations, pre-allocation is cheaper than eviction machinery.

**No dependencies** — The entire library is `[]uint64` arithmetic. No SIMD, no assembly, no external packages.

**Generalized addressing** — `BitMatrix` and `MultiBitMatrix` share the same `baseMatrix` storage engine. The addressing arithmetic is parameterized on `bitsPerCell`. When `bitsPerCell=1`, the generalized formulas collapse to the simple 1-bit case with no overhead.

## Benchmarks

```
cpu: Apple M2

BitMatrix:
BenchmarkHas-8             1000000000    0.76 ns/op         0 B/op    0 allocs/op
BenchmarkSet-8              625091022    2.04 ns/op         0 B/op    0 allocs/op
BenchmarkClear-8            614040549    1.94 ns/op         0 B/op    0 allocs/op
BenchmarkHasAll_5flags-8    202021831    5.66 ns/op         0 B/op    0 allocs/op
BenchmarkHasAny_5flags-8    284921958    4.20 ns/op         0 B/op    0 allocs/op
BenchmarkCountInCol_1M-8       171336    7037 ns/op         0 B/op    0 allocs/op
BenchmarkClearCol_1M-8         170006    7090 ns/op         0 B/op    0 allocs/op
BenchmarkColAnd_1M-8           155294    7222 ns/op    131104 B/op    2 allocs/op
BenchmarkEnsure_WithinBounds-8 1000000000       2.00 ns/op         0 B/op    0 allocs/op
BenchmarkAtomicHas-8       1000000000    0.79 ns/op         0 B/op    0 allocs/op
BenchmarkAtomicSet-8        154550314    7.72 ns/op         0 B/op    0 allocs/op

MultiBitMatrix (4-bit cells):
BenchmarkMultiBitGet-8     1000000000    1.28 ns/op         0 B/op    0 allocs/op
BenchmarkMultiBitSet-8     1000000000    2.25 ns/op         0 B/op    0 allocs/op
BenchmarkAtomicMultiBitGet-8 1000000000  1.30 ns/op         0 B/op    0 allocs/op
BenchmarkAtomicMultiBitSet-8  437493141  8.20 ns/op         0 B/op    0 allocs/op
```

`ColAnd`/`ColOr`/`ColAndNot` allocate exactly 2 times (the returned `Bitmap` struct and its backing `[]uint64`). `Grow` and `Ensure` (when growing) allocate exactly 1 time (one contiguous slab covering all columns). All other operations are zero-alloc.

## License

MIT
