package bitmatrix

import "math/bits"

// BitMatrix is a 2D bit matrix stored in column-major order.
// Each column owns a contiguous array of chunks; each chunk covers
// chunkSize entities packed into wordsPerChunk uint64 words.
//
// Memory layout:
//
//	data[col] = flat slice of numChunks × wordsPerChunk uint64 words
//	chunk c   starts at data[col][c * wordsPerChunk]
type BitMatrix struct {
	baseMatrix
}

// New creates a BitMatrix. Without options it defaults to MaxRows rows,
// MaxCols columns, and ChunkSize entities per chunk.
//
//	bm := New()
//	bm := New(WithRows(500_000))
//	bm := New(WithRows(1_000_000), WithCols(256))
//	bm := New(WithChunkSize(4096))
func New(opts ...Option) *BitMatrix {
	cfg := &bitMatrixConfig{
		rows:      MaxRows,
		cols:      MaxCols,
		chunkSize: ChunkSize,
	}
	for i := 0; i < len(opts); i++ {
		opts[i](cfg)
	}
	return &BitMatrix{
		baseMatrix: newBase(cfg.rows, cfg.cols, cfg.chunkSize, 1),
	}
}

// Has reports whether entity row has flag col set.
func (bm *BitMatrix) Has(row, col uint32) bool {
	return bm.data[col][bm.cellWordIdx(row)]>>bm.cellBitOffset(row)&1 != 0
}

// Set grants flag col to entity row.
func (bm *BitMatrix) Set(row, col uint32) {
	bm.data[col][bm.cellWordIdx(row)] |= 1 << bm.cellBitOffset(row)
}

// Ensure sets flag col for entity row, growing the matrix first if row is
// out of range. row is a 0-based index, so Grow receives row+1 (the minimum
// row count needed to make index row valid).
// Use this instead of Set when the entity ID is not guaranteed to be within
// current bounds.
func (bm *BitMatrix) Ensure(row, col uint32) {
	if row >= bm.rows {
		bm.Grow(row + 1)
	}
	bm.Set(row, col)
}

// Clear revokes flag col from entity row.
func (bm *BitMatrix) Clear(row, col uint32) {
	bm.data[col][bm.cellWordIdx(row)] &^= 1 << bm.cellBitOffset(row)
}

// CountInCol returns how many entities within the current row range have
// flag col set. Padding bits beyond rows are not counted.
func (bm *BitMatrix) CountInCol(col uint32) int {
	n := 0
	words := bm.data[col]
	fullWords := int(bm.rows / BitsPerWord)
	for i := 0; i < fullWords; i++ {
		n += bits.OnesCount64(words[i])
	}
	if remainder := bm.rows % BitsPerWord; remainder > 0 {
		mask := (uint64(1) << remainder) - 1
		n += bits.OnesCount64(words[fullWords] & mask)
	}
	return n
}

// SetColChunk sets flag col for every entity in the given chunk index,
// granting the flag to chunkSize consecutive entities at once.
func (bm *BitMatrix) SetColChunk(col, chunk uint32) {
	off := int(chunk) * int(bm.wordsPerChunk)
	end := off + int(bm.wordsPerChunk)
	for i := off; i < end; i++ {
		bm.data[col][i] = ^uint64(0)
	}
}

// ColAnd returns a Bitmap of entities that have both colA and colB set.
func (bm *BitMatrix) ColAnd(colA, colB uint32) *Bitmap {
	return bm.mergeCol(colA, colB, func(a, b uint64) uint64 { return a & b })
}

// ColOr returns a Bitmap of entities that have colA or colB (or both) set.
func (bm *BitMatrix) ColOr(colA, colB uint32) *Bitmap {
	return bm.mergeCol(colA, colB, func(a, b uint64) uint64 { return a | b })
}

// ColAndNot returns a Bitmap of entities that have colA set but not colB.
func (bm *BitMatrix) ColAndNot(colA, colB uint32) *Bitmap {
	return bm.mergeCol(colA, colB, func(a, b uint64) uint64 { return a &^ b })
}

// colOp is a bitwise binary operation applied word-by-word across two columns.
type colOp func(a, b uint64) uint64

func (bm *BitMatrix) mergeCol(colA, colB uint32, op colOp) *Bitmap {
	wordsA, wordsB := bm.data[colA], bm.data[colB]
	out := make([]uint64, len(wordsA))
	for i := 0; i < len(wordsA); i++ {
		out[i] = op(wordsA[i], wordsB[i])
	}
	return &Bitmap{collection: out, rows: bm.rows}
}
