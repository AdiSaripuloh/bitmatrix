package bitmatrix

import "sync/atomic"

// AtomicBitMatrix wraps BitMatrix with atomic read/write operations.
// Use when Set/Clear may race without an external mutex.
//
// Has, Set, and Clear are safe for concurrent use. However, Grow and Ensure
// (which calls Grow internally) are NOT atomic — they reallocate the
// underlying slices without synchronization. If concurrent Grow/Ensure calls
// are possible, protect them with an external lock (e.g. sync.RWMutex).
type AtomicBitMatrix struct {
	BitMatrix
}

// NewAtomic creates an AtomicBitMatrix. Accepts the same options as New.
func NewAtomic(opts ...Option) *AtomicBitMatrix {
	return &AtomicBitMatrix{*New(opts...)}
}

// Has loads the word atomically before testing the bit.
func (a *AtomicBitMatrix) Has(row, col uint32) bool {
	return atomic.LoadUint64(&a.data[col][a.cellWordIdx(row)])>>a.cellBitOffset(row)&1 != 0
}

// Ensure sets flag col for entity row, growing the matrix first if row is
// out of range. The grow is not atomic — use an external lock if concurrent
// Ensure calls are possible. The subsequent Set is atomic.
func (a *AtomicBitMatrix) Ensure(row, col uint32) {
	if row >= a.rows {
		a.Grow(row + 1)
	}
	a.Set(row, col)
}

// Set atomically ORs the target bit using a CAS loop.
func (a *AtomicBitMatrix) Set(row, col uint32) {
	ptr := &a.data[col][a.cellWordIdx(row)]
	mask := uint64(1) << a.cellBitOffset(row)
	for {
		old := atomic.LoadUint64(ptr)
		if atomic.CompareAndSwapUint64(ptr, old, old|mask) {
			return
		}
	}
}

// Clear atomically ANDs-away the target bit using a CAS loop.
func (a *AtomicBitMatrix) Clear(row, col uint32) {
	ptr := &a.data[col][a.cellWordIdx(row)]
	mask := ^(uint64(1) << a.cellBitOffset(row))
	for {
		old := atomic.LoadUint64(ptr)
		if atomic.CompareAndSwapUint64(ptr, old, old&mask) {
			return
		}
	}
}
