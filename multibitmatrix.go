package bitmatrix

// MultiBitMatrix is a 2D matrix where each cell stores a configurable number
// of bits (1, 2, 4, 8, 16, or 32). It uses the same column-major layout as
// BitMatrix: each column owns a contiguous array of uint64 words. The
// difference is that each word holds 64/bitsPerCell cells instead of 64.
//
// Cell values never span two uint64 words because bitsPerCell is restricted
// to powers of 2 that evenly divide 64.
//
// Memory layout:
//
//	data[col] = flat slice of numChunks × wordsPerChunk uint64 words
//	cellsPerWord = 64 / bitsPerCell
//	wordIdx      = row / cellsPerWord
//	bitOffset    = (row % cellsPerWord) * bitsPerCell
type MultiBitMatrix struct {
	baseMatrix
}

// NewMultiBitMatrix creates a MultiBitMatrix. Without options it defaults to
// MaxRows rows, MaxCols columns, ChunkSize entities per chunk, and 2 bits
// per cell.
//
//	m := NewMultiBitMatrix()
//	m := NewMultiBitMatrix(WithBitsPerCell(4), WithMultiBitRows(1_000_000))
//	m := NewMultiBitMatrix(WithBitsPerCell(8), WithMultiBitCols(256))
func NewMultiBitMatrix(opts ...MultiBitMatrixOption) *MultiBitMatrix {
	cfg := &multiBitMatrixConfig{
		rows:        MaxRows,
		cols:        MaxCols,
		chunkSize:   ChunkSize,
		bitsPerCell: 2,
	}
	for i := 0; i < len(opts); i++ {
		opts[i](cfg)
	}
	return &MultiBitMatrix{
		baseMatrix: newBase(cfg.rows, cfg.cols, cfg.chunkSize, cfg.bitsPerCell),
	}
}

// Get returns the multi-bit value stored at (row, col).
func (m *MultiBitMatrix) Get(row, col uint32) uint64 {
	wi := m.cellWordIdx(row)
	off := m.cellBitOffset(row)
	return (m.data[col][wi] >> off) & m.valueMask
}

// Set stores val into cell (row, col). If val exceeds the cell's bit width
// it is silently truncated to bitsPerCell bits.
func (m *MultiBitMatrix) Set(row, col uint32, val uint64) {
	wi := m.cellWordIdx(row)
	off := m.cellBitOffset(row)
	v := val & m.valueMask
	m.data[col][wi] = (m.data[col][wi] &^ (m.valueMask << off)) | (v << off)
}

// Clear sets cell (row, col) to zero.
func (m *MultiBitMatrix) Clear(row, col uint32) {
	wi := m.cellWordIdx(row)
	off := m.cellBitOffset(row)
	m.data[col][wi] &^= m.valueMask << off
}

// Ensure sets val at (row, col), growing the matrix first if row is out of
// range.
func (m *MultiBitMatrix) Ensure(row, col uint32, val uint64) {
	if row >= m.rows {
		m.Grow(row + 1)
	}
	m.Set(row, col, val)
}

// CountInCol returns how many cells in column col have a non-zero value.
func (m *MultiBitMatrix) CountInCol(col uint32) int {
	n := 0
	totalCells := m.rows
	var cell uint32
	for i := 0; i < len(m.data[col]); i++ {
		w := m.data[col][i]
		if w == 0 {
			cell += m.cellsPerWord
			continue
		}
		for j := uint32(0); j < m.cellsPerWord && cell < totalCells; j++ {
			if (w>>uint(j*m.bitsPerCell))&m.valueMask != 0 {
				n++
			}
			cell++
		}
	}
	return n
}
