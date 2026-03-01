package bitmatrix

import "testing"

func TestNew_Defaults(t *testing.T) {
	bm := New()
	if bm.rows != MaxRows {
		t.Errorf("rows = %d, want %d", bm.rows, MaxRows)
	}
	if bm.cols != MaxCols {
		t.Errorf("cols = %d, want %d", bm.cols, MaxCols)
	}
	if bm.chunkSize != ChunkSize {
		t.Errorf("chunkSize = %d, want %d", bm.chunkSize, ChunkSize)
	}
	if bm.wordsPerChunk != ChunkSize/BitsPerWord {
		t.Errorf("wordsPerChunk = %d, want %d", bm.wordsPerChunk, ChunkSize/BitsPerWord)
	}
}

func TestNew_WithOptions(t *testing.T) {
	bm := New(WithRows(512), WithCols(32), WithChunkSize(128))
	if bm.rows != 512 {
		t.Errorf("rows = %d, want 512", bm.rows)
	}
	if bm.cols != 32 {
		t.Errorf("cols = %d, want 32", bm.cols)
	}
	if bm.chunkSize != 128 {
		t.Errorf("chunkSize = %d, want 128", bm.chunkSize)
	}
	if bm.wordsPerChunk != 128/BitsPerWord {
		t.Errorf("wordsPerChunk = %d, want %d", bm.wordsPerChunk, 128/BitsPerWord)
	}
}

func TestNew_InvalidOptions(t *testing.T) {
	// Zero rows → ignored, default used.
	bm := New(WithRows(0))
	if bm.rows != MaxRows {
		t.Errorf("zero rows: got %d, want default %d", bm.rows, MaxRows)
	}

	// Zero cols → ignored, default used.
	bm = New(WithCols(0))
	if bm.cols != MaxCols {
		t.Errorf("zero cols: got %d, want default %d", bm.cols, MaxCols)
	}

	// ChunkSize not multiple of BitsPerWord → ignored.
	bm = New(WithChunkSize(100))
	if bm.chunkSize != ChunkSize {
		t.Errorf("bad chunkSize: got %d, want default %d", bm.chunkSize, ChunkSize)
	}

	// ChunkSize below minimum → ignored.
	bm = New(WithChunkSize(32))
	if bm.chunkSize != ChunkSize {
		t.Errorf("small chunkSize: got %d, want default %d", bm.chunkSize, ChunkSize)
	}
}

func TestAccessors_BitMatrix(t *testing.T) {
	bm := New(WithRows(512), WithCols(32))
	if bm.Rows() != 512 {
		t.Errorf("Rows() = %d, want 512", bm.Rows())
	}
	if bm.Cols() != 32 {
		t.Errorf("Cols() = %d, want 32", bm.Cols())
	}
	if bm.BitsPerCell() != 1 {
		t.Errorf("BitsPerCell() = %d, want 1", bm.BitsPerCell())
	}
}

func TestAccessors_MultiBitMatrix(t *testing.T) {
	m := NewMultiBitMatrix(WithMultiBitRows(256), WithMultiBitCols(16), WithBitsPerCell(4))
	if m.Rows() != 256 {
		t.Errorf("Rows() = %d, want 256", m.Rows())
	}
	if m.Cols() != 16 {
		t.Errorf("Cols() = %d, want 16", m.Cols())
	}
	if m.BitsPerCell() != 4 {
		t.Errorf("BitsPerCell() = %d, want 4", m.BitsPerCell())
	}
}

func TestAccessors_RowsAfterGrow(t *testing.T) {
	bm := New(WithRows(64), WithCols(4))
	if bm.Rows() != 64 {
		t.Errorf("initial Rows() = %d, want 64", bm.Rows())
	}
	bm.Grow(1000)
	if bm.Rows() != 1000 {
		t.Errorf("after Grow: Rows() = %d, want 1000", bm.Rows())
	}
}

func TestAccessors_BitmapRows(t *testing.T) {
	bm := New(WithRows(500), WithCols(4))
	bm.Set(10, 0)
	bm.Set(10, 1)
	result := bm.ColAnd(0, 1)
	if result.Rows() != 500 {
		t.Errorf("Bitmap.Rows() = %d, want 500", result.Rows())
	}
}

func TestNew_CustomChunkSize(t *testing.T) {
	bm := New(WithRows(1_000_000), WithCols(8), WithChunkSize(128))
	if bm.chunkSize != 128 {
		t.Fatalf("chunkSize = %d, want 128", bm.chunkSize)
	}
	if bm.wordsPerChunk != 128/BitsPerWord {
		t.Fatalf("wordsPerChunk = %d, want %d", bm.wordsPerChunk, 128/BitsPerWord)
	}

	bm.Set(127, 0) // last row of first chunk
	bm.Set(128, 0) // first row of second chunk
	if !bm.Has(127, 0) || !bm.Has(128, 0) {
		t.Error("Set/Has failed across custom chunk boundary")
	}
}
