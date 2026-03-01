package bitmatrix

import (
	"sync"
	"testing"
)

// ── Basic round-trip ────────────────────────────────────────────────────

func TestAtomicMultiBitMatrix_SetGetClear(t *testing.T) {
	for _, bpc := range []uint32{1, 2, 4, 8, 16, 32} {
		t.Run(bpcName(bpc), func(t *testing.T) {
			m := NewAtomicMultiBitMatrix(WithMultiBitRows(1000), WithMultiBitCols(16), WithBitsPerCell(bpc))
			maxVal := uint64((1 << bpc) - 1)

			pairs := [][2]uint32{{0, 0}, {1, 1}, {63, 2}, {64, 3}, {65, 4}, {999, 15}}
			for _, p := range pairs {
				m.Set(p[0], p[1], maxVal)
				if got := m.Get(p[0], p[1]); got != maxVal {
					t.Errorf("Get(%d,%d) = %d, want %d", p[0], p[1], got, maxVal)
				}
			}

			// Clear and verify zero.
			for _, p := range pairs {
				m.Clear(p[0], p[1])
				if got := m.Get(p[0], p[1]); got != 0 {
					t.Errorf("after Clear: Get(%d,%d) = %d, want 0", p[0], p[1], got)
				}
			}
		})
	}
}

// ── Concurrent — different rows ─────────────────────────────────────────

func TestAtomicMultiBitMatrix_Concurrent(t *testing.T) {
	const (
		rows       = 10000
		cols       = 8
		goroutines = 16
		bpc        = 4
	)
	m := NewAtomicMultiBitMatrix(WithMultiBitRows(rows), WithMultiBitCols(cols), WithBitsPerCell(bpc))
	maxVal := uint64((1 << bpc) - 1)

	var wg sync.WaitGroup
	rowsPerG := uint32(rows / goroutines)
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		start := uint32(g) * rowsPerG
		go func(start uint32) {
			defer wg.Done()
			for r := start; r < start+rowsPerG; r++ {
				val := uint64(r%16) & maxVal
				for c := uint32(0); c < cols; c++ {
					m.Set(r, c, val)
				}
			}
		}(start)
	}
	wg.Wait()

	// Verify all values.
	for g := 0; g < goroutines; g++ {
		start := uint32(g) * rowsPerG
		for r := start; r < start+rowsPerG; r++ {
			want := uint64(r%16) & maxVal
			for c := uint32(0); c < cols; c++ {
				if got := m.Get(r, c); got != want {
					t.Errorf("Get(%d,%d) = %d, want %d", r, c, got, want)
				}
			}
		}
	}
}

// ── Concurrent — same cell ──────────────────────────────────────────────

func TestAtomicMultiBitMatrix_ConcurrentSameCell(t *testing.T) {
	const (
		goroutines = 16
		bpc        = 4
	)
	m := NewAtomicMultiBitMatrix(WithMultiBitRows(100), WithMultiBitCols(4), WithBitsPerCell(bpc))

	// Use only odd values so the final assertion is non-vacuous.
	written := make(map[uint64]bool, goroutines)
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		val := uint64(g*2+1) & ((1 << bpc) - 1)
		written[val] = true
		go func(v uint64) {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				m.Set(0, 0, v)
			}
		}(val)
	}
	wg.Wait()

	// The final value must be one of the values that was actually written.
	got := m.Get(0, 0)
	if !written[got] {
		t.Errorf("Get(0,0) = %d, not one of the written values", got)
	}
}

// ── Concurrent — neighbour isolation ────────────────────────────────────

func TestAtomicMultiBitMatrix_ConcurrentNeighbourIsolation(t *testing.T) {
	for _, bpc := range []uint32{2, 4, 8, 16, 32} {
		t.Run(bpcName(bpc), func(t *testing.T) {
			cpw := uint32(64) / bpc
			// Allocate enough rows for at least 2 words so we test within-word neighbours.
			m := NewAtomicMultiBitMatrix(WithMultiBitRows(cpw*2), WithMultiBitCols(4), WithBitsPerCell(bpc))
			maxVal := uint64((1 << bpc) - 1)

			// Pick two adjacent cells in the same word.
			rowA := cpw/2 - 1
			rowB := cpw / 2

			const iterations = 5000
			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				defer wg.Done()
				for i := 0; i < iterations; i++ {
					m.Set(rowA, 0, maxVal)
				}
			}()
			go func() {
				defer wg.Done()
				for i := 0; i < iterations; i++ {
					m.Set(rowB, 0, maxVal)
				}
			}()
			wg.Wait()

			if got := m.Get(rowA, 0); got != maxVal {
				t.Errorf("row %d: Get = %d, want %d", rowA, got, maxVal)
			}
			if got := m.Get(rowB, 0); got != maxVal {
				t.Errorf("row %d: Get = %d, want %d", rowB, got, maxVal)
			}

			// Untouched neighbours must still be zero.
			if rowA > 0 {
				if got := m.Get(rowA-1, 0); got != 0 {
					t.Errorf("row %d (before A): Get = %d, want 0", rowA-1, got)
				}
			}
			if rowB+1 < cpw*2 {
				if got := m.Get(rowB+1, 0); got != 0 {
					t.Errorf("row %d (after B): Get = %d, want 0", rowB+1, got)
				}
			}
			// Different column must remain zero.
			if got := m.Get(rowA, 1); got != 0 {
				t.Errorf("col 1 for row %d: Get = %d, want 0", rowA, got)
			}
		})
	}
}

// ── Ensure ──────────────────────────────────────────────────────────────

func TestAtomicMultiBitMatrix_Ensure(t *testing.T) {
	m := NewAtomicMultiBitMatrix(WithMultiBitRows(100), WithMultiBitCols(4), WithBitsPerCell(4))
	m.Set(10, 0, 5)

	// Within bounds — should just set atomically.
	m.Ensure(10, 1, 3)
	if got := m.Get(10, 1); got != 3 {
		t.Errorf("Ensure within bounds: Get(10,1) = %d, want 3", got)
	}

	// Out of bounds — should grow and set.
	m.Ensure(500, 2, 9)
	if m.rows <= 500 {
		t.Errorf("rows = %d, want > 500", m.rows)
	}
	if got := m.Get(500, 2); got != 9 {
		t.Errorf("Ensure out of bounds: Get(500,2) = %d, want 9", got)
	}

	// Existing data preserved after grow.
	if got := m.Get(10, 0); got != 5 {
		t.Errorf("after Ensure grow: Get(10,0) = %d, want 5", got)
	}
}

// ── Benchmarks ──────────────────────────────────────────────────────────

var sinkU64 uint64 // prevent compiler from eliminating benchmark calls

func BenchmarkAtomicMultiBitGet(b *testing.B) {
	m := NewAtomicMultiBitMatrix(WithMultiBitRows(1_000_000), WithMultiBitCols(128), WithBitsPerCell(4))
	m.Set(500_000, 64, 15)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkU64 = m.Get(500_000, 64)
	}
}

func BenchmarkAtomicMultiBitSet(b *testing.B) {
	m := NewAtomicMultiBitMatrix(WithMultiBitRows(1_000_000), WithMultiBitCols(128), WithBitsPerCell(4))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Set(500_000, 64, 15)
	}
}
