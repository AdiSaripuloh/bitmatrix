package bitmatrix

// BitReader reads a single bit from the matrix.
type BitReader interface {
	Has(row, col uint32) bool
}

// BitWriter sets or clears a single bit in the matrix.
type BitWriter interface {
	Set(row, col uint32)
	Clear(row, col uint32)
}

// FlagChecker performs multi-flag checks against a single entity.
type FlagChecker interface {
	HasAll(row uint32, cols ...uint32) bool
	HasAny(row uint32, cols ...uint32) bool
}

// ColReader queries an entire column across all entities.
type ColReader interface {
	CountInCol(col uint32) int
	ColAnd(colA, colB uint32) *Bitmap
	ColOr(colA, colB uint32) *Bitmap
	ColAndNot(colA, colB uint32) *Bitmap
}

// ColWriter mutates an entire column across all entities.
type ColWriter interface {
	ClearCol(col uint32)
	SetColChunk(col, chunk uint32)
}

// Grower expands the matrix row capacity.
type Grower interface {
	Grow(newRows uint32)
}

// ReadOnlyMatrix is the read-only surface of the matrix.
// Accept this in functions that query but never mutate.
type ReadOnlyMatrix interface {
	BitReader
	FlagChecker
	ColReader
}

// Matrix is the full read/write surface of the matrix.
type Matrix interface {
	ReadOnlyMatrix
	BitWriter
	ColWriter
	Grower
	// Ensure sets flag col for entity row, growing if row is out of range.
	// Prefer this over Set when the entity ID may exceed current bounds.
	Ensure(row, col uint32)
}

// CellReader reads a multi-bit cell value from the matrix.
type CellReader interface {
	Get(row, col uint32) uint64
}

// CellWriter sets or clears a multi-bit cell in the matrix.
type CellWriter interface {
	Set(row, col uint32, val uint64)
	Clear(row, col uint32)
}

// ReadOnlyMultiMatrix is the read-only surface of a multi-bit matrix.
// Accept this in functions that query but never mutate.
type ReadOnlyMultiMatrix interface {
	CellReader
	FlagChecker
}

// MultiMatrix is the full read/write surface of a multi-bit matrix.
type MultiMatrix interface {
	ReadOnlyMultiMatrix
	CellWriter
	Grower
	// Ensure stores val at (row, col), growing if row is out of range.
	Ensure(row, col uint32, val uint64)
}

// Compile-time interface checks.
var _ Matrix = (*BitMatrix)(nil)
var _ Matrix = (*AtomicBitMatrix)(nil)
var _ ReadOnlyMatrix = (*BitMatrix)(nil)
var _ ReadOnlyMatrix = (*AtomicBitMatrix)(nil)
var _ MultiMatrix = (*MultiBitMatrix)(nil)
var _ MultiMatrix = (*AtomicMultiBitMatrix)(nil)
var _ ReadOnlyMultiMatrix = (*MultiBitMatrix)(nil)
var _ ReadOnlyMultiMatrix = (*AtomicMultiBitMatrix)(nil)
