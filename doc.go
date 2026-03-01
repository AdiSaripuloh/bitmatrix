// Package bitmatrix provides a zero-allocation 2D matrix for mapping
// (row, col) pairs to single bits or small multi-bit values. It is designed
// for high-throughput flag/permission checks where millions of entities each
// have up to thousands of boolean flags or small integer states.
//
// The matrix is stored in column-major order so that all entities sharing
// the same flag occupy contiguous memory, making flag-centric queries
// (e.g. "does user X have permission Y?") extremely cache-friendly.
//
// # Quick start
//
//	bm := bitmatrix.New(
//	    bitmatrix.WithRows(1_000_000),
//	    bitmatrix.WithCols(1024),
//	)
//
//	bm.Set(42, 7)          // grant flag 7 to entity 42
//	bm.Has(42, 7)          // true
//	bm.Clear(42, 7)        // revoke flag 7 from entity 42
//
// # Atomic variant
//
// For concurrent writes without an external mutex, use [NewAtomic]:
//
//	abm := bitmatrix.NewAtomic(bitmatrix.WithRows(1_000_000))
//	abm.Set(42, 7)  // atomic CAS loop
//	abm.Has(42, 7)  // atomic load
//
// # Multi-bit matrix
//
// [NewMultiBitMatrix] stores a configurable number of bits per cell (1, 2,
// 4, 8, 16, or 32) instead of a single bit:
//
//	m := bitmatrix.NewMultiBitMatrix(
//	    bitmatrix.WithBitsPerCell(4),
//	    bitmatrix.WithMultiBitRows(1_000_000),
//	)
//	m.Set(42, 0, 15)   // store 4-bit value
//	m.Get(42, 0)        // 15
//
// [NewAtomicMultiBitMatrix] adds lock-free concurrent access using the same
// CAS-loop pattern as [NewAtomic]:
//
//	am := bitmatrix.NewAtomicMultiBitMatrix(
//	    bitmatrix.WithBitsPerCell(4),
//	    bitmatrix.WithMultiBitRows(1_000_000),
//	)
//	am.Set(42, 0, 15)  // atomic CAS loop
//	am.Get(42, 0)       // atomic load
//
// See the project README for detailed benchmarks, memory layout, and
// design rationale: https://github.com/AdiSaripuloh/bitmatrix
package bitmatrix
