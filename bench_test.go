package bitmatrix

import "testing"

var sink bool // prevent compiler from eliminating benchmark calls

func BenchmarkHas(b *testing.B) {
	bm := New(WithRows(1_000_000), WithCols(1024))
	bm.Set(57403, 42)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = bm.Has(57403, 42)
	}
}

func BenchmarkSet(b *testing.B) {
	bm := New(WithRows(1_000_000), WithCols(1024))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bm.Set(57403, 42)
	}
}

func BenchmarkClear(b *testing.B) {
	bm := New(WithRows(1_000_000), WithCols(1024))
	bm.Set(57403, 42)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bm.Clear(57403, 42)
	}
}

func BenchmarkHasAll_5flags(b *testing.B) {
	bm := New(WithRows(1_000_000), WithCols(1024))
	for col := uint32(0); col < 5; col++ {
		bm.Set(57403, col)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = bm.HasAll(57403, 0, 1, 2, 3, 4)
	}
}

func BenchmarkHasAny_5flags(b *testing.B) {
	bm := New(WithRows(1_000_000), WithCols(1024))
	bm.Set(57403, 4)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = bm.HasAny(57403, 0, 1, 2, 3, 4)
	}
}

func BenchmarkCountInCol_1M(b *testing.B) {
	bm := New(WithRows(1_000_000), WithCols(1024))
	bm.SetColChunk(42, 0)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bm.CountInCol(42)
	}
}

func BenchmarkClearCol_1M(b *testing.B) {
	bm := New(WithRows(1_000_000), WithCols(1024))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bm.ClearCol(42)
	}
}

func BenchmarkColAnd_1M(b *testing.B) {
	bm := New(WithRows(1_000_000), WithCols(1024))
	bm.SetColChunk(42, 0)
	bm.SetColChunk(99, 0)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bm.ColAnd(42, 99)
	}
}

func BenchmarkGrow(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		bm := New(WithRows(1_000_000), WithCols(64))
		b.StartTimer()
		bm.Grow(2_000_000)
	}
}

func BenchmarkEnsure_WithinBounds(b *testing.B) {
	bm := New(WithRows(1_000_000), WithCols(1024))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bm.Ensure(57403, 42)
	}
}

func BenchmarkEnsure_OutOfBounds(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		bm := New(WithRows(64), WithCols(64))
		b.StartTimer()
		bm.Ensure(999_999, 42)
	}
}

func BenchmarkAtomicHas(b *testing.B) {
	abm := NewAtomic(WithRows(1_000_000), WithCols(1024))
	abm.Set(57403, 42)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = abm.Has(57403, 42)
	}
}

func BenchmarkAtomicSet(b *testing.B) {
	abm := NewAtomic(WithRows(1_000_000), WithCols(1024))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		abm.Set(57403, 42)
	}
}
