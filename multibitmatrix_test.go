package bitmatrix

import (
	"fmt"
	"testing"
)

// ── Constructor ─────────────────────────────────────────────────────────

func TestNewMultiBitMatrix_Defaults(t *testing.T) {
	m := NewMultiBitMatrix()
	if m.rows != MaxRows {
		t.Errorf("rows = %d, want %d", m.rows, MaxRows)
	}
	if m.cols != MaxCols {
		t.Errorf("cols = %d, want %d", m.cols, MaxCols)
	}
	if m.chunkSize != ChunkSize {
		t.Errorf("chunkSize = %d, want %d", m.chunkSize, ChunkSize)
	}
	if m.bitsPerCell != 2 {
		t.Errorf("bitsPerCell = %d, want 2", m.bitsPerCell)
	}
	if m.cellsPerWord != 32 {
		t.Errorf("cellsPerWord = %d, want 32", m.cellsPerWord)
	}
	if m.valueMask != 0x3 {
		t.Errorf("valueMask = %x, want 0x3", m.valueMask)
	}
}

func TestNewMultiBitMatrix_WithOptions(t *testing.T) {
	m := NewMultiBitMatrix(
		WithMultiBitRows(10000),
		WithMultiBitCols(64),
		WithMultiBitChunkSize(4096),
		WithBitsPerCell(4),
	)
	if m.rows != 10000 {
		t.Errorf("rows = %d, want 10000", m.rows)
	}
	if m.cols != 64 {
		t.Errorf("cols = %d, want 64", m.cols)
	}
	if m.chunkSize != 4096 {
		t.Errorf("chunkSize = %d, want 4096", m.chunkSize)
	}
	if m.bitsPerCell != 4 {
		t.Errorf("bitsPerCell = %d, want 4", m.bitsPerCell)
	}
	if m.cellsPerWord != 16 {
		t.Errorf("cellsPerWord = %d, want 16", m.cellsPerWord)
	}
	if m.valueMask != 0xF {
		t.Errorf("valueMask = %x, want 0xF", m.valueMask)
	}
}

func TestNewMultiBitMatrix_InvalidBitsPerCell(t *testing.T) {
	// Invalid values should leave the default (2) unchanged.
	for _, n := range []uint32{0, 3, 5, 6, 7, 9, 15, 33, 64, 128} {
		m := NewMultiBitMatrix(WithBitsPerCell(n))
		if m.bitsPerCell != 2 {
			t.Errorf("WithBitsPerCell(%d): bitsPerCell = %d, want default 2", n, m.bitsPerCell)
		}
	}
}

// ── Set / Get round-trip ────────────────────────────────────────────────

func TestMultiBit_SetGet(t *testing.T) {
	for _, bpc := range []uint32{1, 2, 4, 8, 16, 32} {
		t.Run(bpcName(bpc), func(t *testing.T) {
			m := NewMultiBitMatrix(WithMultiBitRows(1000), WithMultiBitCols(16), WithBitsPerCell(bpc))
			maxVal := uint64((1 << bpc) - 1)

			// Set a few values at different rows.
			pairs := [][2]uint32{{0, 0}, {1, 1}, {63, 2}, {64, 3}, {65, 4}, {999, 15}}
			for _, p := range pairs {
				m.Set(p[0], p[1], maxVal)
				got := m.Get(p[0], p[1])
				if got != maxVal {
					t.Errorf("Get(%d,%d) = %d, want %d", p[0], p[1], got, maxVal)
				}
			}
		})
	}
}

// ── Clear ───────────────────────────────────────────────────────────────

func TestMultiBit_Clear(t *testing.T) {
	m := NewMultiBitMatrix(WithMultiBitRows(100), WithMultiBitCols(4), WithBitsPerCell(4))
	m.Set(10, 2, 15)
	m.Clear(10, 2)
	if got := m.Get(10, 2); got != 0 {
		t.Errorf("after Clear: Get(10,2) = %d, want 0", got)
	}
}

// ── Value overflow / truncation ─────────────────────────────────────────

func TestMultiBit_ValueOverflow(t *testing.T) {
	for _, bpc := range []uint32{1, 2, 4, 8, 16, 32} {
		t.Run(bpcName(bpc), func(t *testing.T) {
			m := NewMultiBitMatrix(WithMultiBitRows(100), WithMultiBitCols(4), WithBitsPerCell(bpc))
			maxVal := uint64((1 << bpc) - 1)
			overflow := maxVal + 1 // one bit too many

			m.Set(0, 0, overflow)
			got := m.Get(0, 0)
			if got != 0 {
				t.Errorf("overflow value %d stored as %d, want 0 (truncated)", overflow, got)
			}

			// All bits set beyond width should be masked off.
			m.Set(1, 0, ^uint64(0))
			got = m.Get(1, 0)
			if got != maxVal {
				t.Errorf("^uint64(0) stored as %d, want %d", got, maxVal)
			}
		})
	}
}

// ── Word boundaries ─────────────────────────────────────────────────────

func TestMultiBit_WordBoundaries(t *testing.T) {
	for _, bpc := range []uint32{1, 2, 4, 8, 16, 32} {
		t.Run(bpcName(bpc), func(t *testing.T) {
			cpw := uint32(64) / bpc
			m := NewMultiBitMatrix(WithMultiBitRows(cpw*3), WithMultiBitCols(4), WithBitsPerCell(bpc))
			maxVal := uint64((1 << bpc) - 1)

			// Last cell of word 0, first cell of word 1, last cell of word 1.
			boundaries := []uint32{cpw - 1, cpw, 2*cpw - 1}
			for _, row := range boundaries {
				m.Set(row, 0, maxVal)
				got := m.Get(row, 0)
				if got != maxVal {
					t.Errorf("row %d: Get = %d, want %d", row, got, maxVal)
				}
			}
		})
	}
}

// ── Chunk boundaries ────────────────────────────────────────────────────

func TestMultiBit_ChunkBoundaries(t *testing.T) {
	m := NewMultiBitMatrix(
		WithMultiBitRows(ChunkSize*2+100),
		WithMultiBitCols(4),
		WithBitsPerCell(4),
		WithMultiBitChunkSize(ChunkSize),
	)
	rows := []uint32{ChunkSize - 1, ChunkSize, ChunkSize + 1, ChunkSize*2 - 1, ChunkSize * 2}
	for _, row := range rows {
		val := uint64(row % 16)
		m.Set(row, 0, val)
		got := m.Get(row, 0)
		if got != val {
			t.Errorf("row %d: Get = %d, want %d", row, got, val)
		}
	}
}

// ── Neighbour isolation ─────────────────────────────────────────────────

func TestMultiBit_SetDoesNotAffectNeighbours(t *testing.T) {
	for _, bpc := range []uint32{1, 2, 4, 8, 16, 32} {
		t.Run(bpcName(bpc), func(t *testing.T) {
			cpw := uint32(64) / bpc
			m := NewMultiBitMatrix(WithMultiBitRows(cpw*2), WithMultiBitCols(4), WithBitsPerCell(bpc))
			maxVal := uint64((1 << bpc) - 1)

			// Set the middle cell of word 0.
			mid := cpw / 2
			m.Set(mid, 0, maxVal)

			// Neighbours must remain zero.
			if mid > 0 {
				if got := m.Get(mid-1, 0); got != 0 {
					t.Errorf("row %d (before): %d, want 0", mid-1, got)
				}
			}
			if mid+1 < cpw {
				if got := m.Get(mid+1, 0); got != 0 {
					t.Errorf("row %d (after): %d, want 0", mid+1, got)
				}
			}
			// Different column must remain zero.
			if got := m.Get(mid, 1); got != 0 {
				t.Errorf("col 1: %d, want 0", got)
			}
		})
	}
}

// ── HasAll / HasAny ─────────────────────────────────────────────────────

func TestMultiBit_HasAll(t *testing.T) {
	m := NewMultiBitMatrix(WithMultiBitRows(100), WithMultiBitCols(8), WithBitsPerCell(4))
	m.Set(5, 0, 1)
	m.Set(5, 1, 3)
	m.Set(5, 2, 15)
	// col 3 left at 0

	if !m.HasAll(5, 0, 1, 2) {
		t.Error("HasAll(5, 0,1,2) = false, want true")
	}
	if m.HasAll(5, 0, 1, 3) {
		t.Error("HasAll(5, 0,1,3) = true, want false")
	}
}

func TestMultiBit_HasAny(t *testing.T) {
	m := NewMultiBitMatrix(WithMultiBitRows(100), WithMultiBitCols(8), WithBitsPerCell(4))
	m.Set(5, 2, 7)
	// cols 0, 1, 3 left at 0

	if !m.HasAny(5, 0, 1, 2) {
		t.Error("HasAny(5, 0,1,2) = false, want true")
	}
	if m.HasAny(5, 0, 1, 3) {
		t.Error("HasAny(5, 0,1,3) = true, want false")
	}
}

// ── CountInCol ──────────────────────────────────────────────────────────

func TestMultiBit_CountInCol(t *testing.T) {
	m := NewMultiBitMatrix(WithMultiBitRows(200), WithMultiBitCols(4), WithBitsPerCell(4))
	m.Set(0, 0, 1)
	m.Set(50, 0, 2)
	m.Set(100, 0, 15)
	m.Set(199, 0, 3)

	if got := m.CountInCol(0); got != 4 {
		t.Errorf("CountInCol(0) = %d, want 4", got)
	}
	if got := m.CountInCol(1); got != 0 {
		t.Errorf("CountInCol(1) = %d, want 0", got)
	}
}

// ── ClearCol ────────────────────────────────────────────────────────────

func TestMultiBit_ClearCol(t *testing.T) {
	m := NewMultiBitMatrix(WithMultiBitRows(200), WithMultiBitCols(4), WithBitsPerCell(4))
	for row := uint32(0); row < 200; row++ {
		m.Set(row, 1, uint64(row%16))
	}
	m.ClearCol(1)

	for row := uint32(0); row < 200; row++ {
		if got := m.Get(row, 1); got != 0 {
			t.Errorf("after ClearCol: Get(%d,1) = %d, want 0", row, got)
		}
	}
}

// ── Grow ────────────────────────────────────────────────────────────────

func TestMultiBit_Grow(t *testing.T) {
	m := NewMultiBitMatrix(WithMultiBitRows(100), WithMultiBitCols(4), WithBitsPerCell(4))
	m.Set(99, 0, 7)
	m.Set(50, 1, 3)

	m.Grow(500)
	if m.rows != 500 {
		t.Errorf("rows = %d, want 500", m.rows)
	}

	// Existing data preserved.
	if got := m.Get(99, 0); got != 7 {
		t.Errorf("after Grow: Get(99,0) = %d, want 7", got)
	}
	if got := m.Get(50, 1); got != 3 {
		t.Errorf("after Grow: Get(50,1) = %d, want 3", got)
	}

	// New region accessible and zero.
	if got := m.Get(499, 0); got != 0 {
		t.Errorf("new row: Get(499,0) = %d, want 0", got)
	}
	m.Set(499, 0, 15)
	if got := m.Get(499, 0); got != 15 {
		t.Errorf("after Set: Get(499,0) = %d, want 15", got)
	}
}

func TestMultiBit_Grow_Noop(t *testing.T) {
	m := NewMultiBitMatrix(WithMultiBitRows(100), WithMultiBitCols(4), WithBitsPerCell(4))
	m.Grow(50) // smaller, should be no-op
	if m.rows != 100 {
		t.Errorf("rows = %d, want 100", m.rows)
	}
}

// ── Ensure ──────────────────────────────────────────────────────────────

func TestMultiBit_Ensure(t *testing.T) {
	m := NewMultiBitMatrix(WithMultiBitRows(100), WithMultiBitCols(4), WithBitsPerCell(4))
	m.Set(10, 0, 5)

	// Within bounds — should just set.
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

// ── Fuzz ────────────────────────────────────────────────────────────────

func FuzzMultiBitSetGetRoundTrip(f *testing.F) {
	f.Add(uint32(0), uint32(0), uint64(0))
	f.Add(uint32(999), uint32(15), uint64(0xFFFF))
	f.Add(uint32(31), uint32(7), uint64(3))

	m := NewMultiBitMatrix(WithMultiBitRows(1000), WithMultiBitCols(16), WithBitsPerCell(16))
	f.Fuzz(func(t *testing.T, row uint32, col uint32, val uint64) {
		row %= 1000
		col %= 16
		masked := val & m.valueMask

		m.Set(row, col, val)
		got := m.Get(row, col)
		if got != masked {
			t.Errorf("Set(%d,%d,%d) then Get = %d, want %d", row, col, val, got, masked)
		}
		m.Clear(row, col)
	})
}

// ── Benchmarks ──────────────────────────────────────────────────────────

func BenchmarkMultiBitGet(b *testing.B) {
	m := NewMultiBitMatrix(WithMultiBitRows(1_000_000), WithMultiBitCols(128), WithBitsPerCell(4))
	m.Set(500_000, 64, 15)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkU64 = m.Get(500_000, 64)
	}
}

func BenchmarkMultiBitSet(b *testing.B) {
	m := NewMultiBitMatrix(WithMultiBitRows(1_000_000), WithMultiBitCols(128), WithBitsPerCell(4))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Set(500_000, 64, 15)
	}
}

// ── Helpers ─────────────────────────────────────────────────────────────

func bpcName(bpc uint32) string {
	return fmt.Sprintf("%dbpc", bpc)
}
