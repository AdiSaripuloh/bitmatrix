package bitmatrix

// baseMatrix is the shared storage and addressing logic for BitMatrix and
// MultiBitMatrix. Both types embed it. When bitsPerCell=1 the generalized
// addressing collapses to the simple 1-bit-per-cell case.
type baseMatrix struct {
	rows          uint32
	cols          uint32
	numChunks     uint32
	chunkSize     uint32
	wordsPerChunk uint32
	bitsPerCell   uint32
	cellsPerWord  uint32
	valueMask     uint64
	data          [][]uint64
}

// newBase allocates a column-major matrix with the given geometry.
// bitsPerCell must be a power of 2 that evenly divides 64.
func newBase(rows, cols, chunkSize, bitsPerCell uint32) baseMatrix {
	cellsPerWord := uint32(BitsPerWord) / bitsPerCell
	wordsPerChunk := (chunkSize + cellsPerWord - 1) / cellsPerWord
	numChunks := (rows + chunkSize - 1) / chunkSize
	nWords := int(numChunks) * int(wordsPerChunk)

	slab := make([]uint64, int(cols)*nWords)
	data := make([][]uint64, cols)
	for c := 0; c < len(data); c++ {
		data[c] = slab[c*nWords : (c+1)*nWords]
	}
	return baseMatrix{
		rows:          rows,
		cols:          cols,
		numChunks:     numChunks,
		chunkSize:     chunkSize,
		wordsPerChunk: wordsPerChunk,
		bitsPerCell:   bitsPerCell,
		cellsPerWord:  cellsPerWord,
		valueMask:     (1 << bitsPerCell) - 1,
		data:          data,
	}
}

// Grow expands the matrix to accommodate newRows entities.
// Existing data is untouched; new chunks are zero-initialised.
// Note: Grow does not enforce MaxRows — it allows expansion beyond the
// limit enforced by the constructor options.
func (b *baseMatrix) Grow(newRows uint32) {
	if newRows <= b.rows {
		return
	}
	nChunks := (newRows + b.chunkSize - 1) / b.chunkSize
	if nChunks > b.numChunks {
		newWords := int(nChunks) * int(b.wordsPerChunk)
		slab := make([]uint64, len(b.data)*newWords)
		for i := 0; i < len(b.data); i++ {
			col := slab[i*newWords : (i+1)*newWords]
			copy(col, b.data[i])
			b.data[i] = col
		}
		b.numChunks = nChunks
	}
	b.rows = newRows
}

// ClearCol sets every cell in column col to zero.
func (b *baseMatrix) ClearCol(col uint32) {
	for i := 0; i < len(b.data[col]); i++ {
		b.data[col][i] = 0
	}
}

// HasAll reports whether entity row has a non-zero value in every listed column.
func (b *baseMatrix) HasAll(row uint32, cols ...uint32) bool {
	wi := b.cellWordIdx(row)
	off := b.cellBitOffset(row)
	for i := 0; i < len(cols); i++ {
		if (b.data[cols[i]][wi]>>off)&b.valueMask == 0 {
			return false
		}
	}
	return true
}

// HasAny reports whether entity row has a non-zero value in at least one
// listed column.
func (b *baseMatrix) HasAny(row uint32, cols ...uint32) bool {
	wi := b.cellWordIdx(row)
	off := b.cellBitOffset(row)
	for i := 0; i < len(cols); i++ {
		if (b.data[cols[i]][wi]>>off)&b.valueMask != 0 {
			return true
		}
	}
	return false
}

// Rows returns the current row capacity.
func (b *baseMatrix) Rows() uint32 { return b.rows }

// Cols returns the number of columns.
func (b *baseMatrix) Cols() uint32 { return b.cols }

// BitsPerCell returns the number of bits each cell occupies.
func (b *baseMatrix) BitsPerCell() uint32 { return b.bitsPerCell }

// cellWordIdx returns the index of the uint64 word containing the cell for row.
func (b *baseMatrix) cellWordIdx(row uint32) int {
	return int(row / b.cellsPerWord)
}

// cellBitOffset returns the bit offset within the word for the cell at row.
func (b *baseMatrix) cellBitOffset(row uint32) uint32 {
	return (row % b.cellsPerWord) * b.bitsPerCell
}
