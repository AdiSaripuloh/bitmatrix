package bitmatrix

import "testing"

// matrixSizes covers the full range from a single-cell matrix to the
// 1M × 1K default, letting you observe how memory and latency scale.
var matrixSizes = []struct {
	name string
	rows uint32
	cols uint32
}{
	{"1x1", 1, 1},
	{"64x1", 64, 1},
	{"1Kx4", 1_000, 4},
	{"8Kx64", 8_192, 64},
	{"100Kx256", 100_000, 256},
	{"1Mx1K", 1_000_000, 1_024},
}

// BenchmarkSizes_New measures construction cost (allocations + time).
func BenchmarkSizes_New(b *testing.B) {
	for _, s := range matrixSizes {
		b.Run(s.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = New(WithRows(s.rows), WithCols(s.cols))
			}
		})
	}
}

// BenchmarkSizes_Has measures single-bit read at each matrix size.
// Memory footprint grows but Has latency should stay flat —
// it is always two arithmetic ops on one cache line.
func BenchmarkSizes_Has(b *testing.B) {
	for _, s := range matrixSizes {
		b.Run(s.name, func(b *testing.B) {
			bm := New(WithRows(s.rows), WithCols(s.cols))
			row := s.rows / 2
			col := s.cols / 2
			bm.Set(row, col)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				sink = bm.Has(row, col)
			}
		})
	}
}

// BenchmarkSizes_Set measures single-bit write at each matrix size.
func BenchmarkSizes_Set(b *testing.B) {
	for _, s := range matrixSizes {
		b.Run(s.name, func(b *testing.B) {
			bm := New(WithRows(s.rows), WithCols(s.cols))
			row := s.rows / 2
			col := s.cols / 2
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bm.Set(row, col)
			}
		})
	}
}

// BenchmarkSizes_Ensure_WithinBounds measures Ensure when the row already exists.
func BenchmarkSizes_Ensure_WithinBounds(b *testing.B) {
	for _, s := range matrixSizes {
		b.Run(s.name, func(b *testing.B) {
			bm := New(WithRows(s.rows), WithCols(s.cols))
			row := s.rows / 2
			col := s.cols / 2
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bm.Ensure(row, col)
			}
		})
	}
}

// BenchmarkSizes_CountInCol measures popcount across all words in a column.
// This scales linearly with rows — it is the operation most sensitive to size.
func BenchmarkSizes_CountInCol(b *testing.B) {
	for _, s := range matrixSizes {
		b.Run(s.name, func(b *testing.B) {
			bm := New(WithRows(s.rows), WithCols(s.cols))
			col := s.cols / 2
			bm.SetColChunk(col, 0)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = bm.CountInCol(col)
			}
		})
	}
}

// BenchmarkSizes_ColAnd measures the AND set operation across two columns.
// Allocates one Bitmap per call — allocation size scales with rows.
func BenchmarkSizes_ColAnd(b *testing.B) {
	for _, s := range matrixSizes {
		if s.cols < 2 {
			continue // need at least 2 columns for ColAnd
		}
		b.Run(s.name, func(b *testing.B) {
			bm := New(WithRows(s.rows), WithCols(s.cols))
			bm.SetColChunk(0, 0)
			bm.SetColChunk(1, 0)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = bm.ColAnd(0, 1)
			}
		})
	}
}
