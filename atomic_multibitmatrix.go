package bitmatrix

import "sync/atomic"

// AtomicMultiBitMatrix wraps MultiBitMatrix with atomic read/write operations.
// Use when Get/Set/Clear may race without an external mutex.
//
// Get, Set, and Clear are safe for concurrent use. However, Grow and Ensure
// (which calls Grow internally) are NOT atomic — they reallocate the
// underlying slices without synchronization. If concurrent Grow/Ensure calls
// are possible, protect them with an external lock (e.g. sync.RWMutex).
type AtomicMultiBitMatrix struct {
	MultiBitMatrix
}

// NewAtomicMultiBitMatrix creates an AtomicMultiBitMatrix. Accepts the same
// options as NewMultiBitMatrix.
func NewAtomicMultiBitMatrix(opts ...MultiBitMatrixOption) *AtomicMultiBitMatrix {
	return &AtomicMultiBitMatrix{*NewMultiBitMatrix(opts...)}
}

// Get atomically loads the word and returns the multi-bit value at (row, col).
func (a *AtomicMultiBitMatrix) Get(row, col uint32) uint64 {
	wi := a.cellWordIdx(row)
	off := a.cellBitOffset(row)
	return (atomic.LoadUint64(&a.data[col][wi]) >> off) & a.valueMask
}

// Set atomically stores val into cell (row, col) using a CAS loop.
// If val exceeds the cell's bit width it is silently truncated.
func (a *AtomicMultiBitMatrix) Set(row, col uint32, val uint64) {
	ptr := &a.data[col][a.cellWordIdx(row)]
	off := a.cellBitOffset(row)
	v := val & a.valueMask
	mask := a.valueMask << off
	for {
		old := atomic.LoadUint64(ptr)
		nw := (old &^ mask) | (v << off)
		if atomic.CompareAndSwapUint64(ptr, old, nw) {
			return
		}
	}
}

// Clear atomically sets cell (row, col) to zero using a CAS loop.
func (a *AtomicMultiBitMatrix) Clear(row, col uint32) {
	ptr := &a.data[col][a.cellWordIdx(row)]
	mask := a.valueMask << a.cellBitOffset(row)
	for {
		old := atomic.LoadUint64(ptr)
		if atomic.CompareAndSwapUint64(ptr, old, old&^mask) {
			return
		}
	}
}

// Ensure sets val at (row, col), growing the matrix first if row is out of
// range. The grow is not atomic — use an external lock if concurrent Ensure
// calls are possible. The subsequent Set is atomic.
func (a *AtomicMultiBitMatrix) Ensure(row, col uint32, val uint64) {
	if row >= a.rows {
		a.Grow(row + 1)
	}
	a.Set(row, col, val)
}
