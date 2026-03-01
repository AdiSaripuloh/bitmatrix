package bitmatrix

// bitMatrixConfig holds the parameters used to construct a BitMatrix.
type bitMatrixConfig struct {
	rows      uint32
	cols      uint32
	chunkSize uint32
}

// Option is a functional option for New and NewAtomic.
type Option func(*bitMatrixConfig)

// WithRows sets the number of entity rows (default: MaxRows).
// Values outside [MinRows, MaxRows] are silently ignored, keeping the default.
func WithRows(rows uint32) Option {
	return func(c *bitMatrixConfig) {
		if rows >= MinRows && rows <= MaxRows {
			c.rows = rows
		}
	}
}

// WithCols sets the number of flag columns (default: MaxCols).
// Values outside [MinCols, MaxCols] are silently ignored, keeping the default.
func WithCols(cols uint32) Option {
	return func(c *bitMatrixConfig) {
		if cols >= MinCols && cols <= MaxCols {
			c.cols = cols
		}
	}
}

// WithChunkSize sets the number of entities per chunk (default: ChunkSize).
// n must be >= MinChunkSize and a multiple of BitsPerWord.
// Values that violate these constraints are silently ignored, keeping the default.
func WithChunkSize(n uint32) Option {
	return func(c *bitMatrixConfig) {
		if n >= MinChunkSize && n%BitsPerWord == 0 {
			c.chunkSize = n
		}
	}
}

// multiBitMatrixConfig holds the parameters used to construct a MultiBitMatrix.
type multiBitMatrixConfig struct {
	rows        uint32
	cols        uint32
	chunkSize   uint32
	bitsPerCell uint32
}

// MultiBitMatrixOption is a functional option for NewMultiBitMatrix.
type MultiBitMatrixOption func(*multiBitMatrixConfig)

// WithMultiBitRows sets the number of entity rows (default: MaxRows).
// Values outside [MinRows, MaxRows] are silently ignored, keeping the default.
func WithMultiBitRows(rows uint32) MultiBitMatrixOption {
	return func(c *multiBitMatrixConfig) {
		if rows >= MinRows && rows <= MaxRows {
			c.rows = rows
		}
	}
}

// WithMultiBitCols sets the number of flag columns (default: MaxCols).
// Values outside [MinCols, MaxCols] are silently ignored, keeping the default.
func WithMultiBitCols(cols uint32) MultiBitMatrixOption {
	return func(c *multiBitMatrixConfig) {
		if cols >= MinCols && cols <= MaxCols {
			c.cols = cols
		}
	}
}

// WithMultiBitChunkSize sets the number of entities per chunk (default: ChunkSize).
// n must be >= MinChunkSize and a multiple of BitsPerWord.
// Values that violate these constraints are silently ignored, keeping the default.
func WithMultiBitChunkSize(n uint32) MultiBitMatrixOption {
	return func(c *multiBitMatrixConfig) {
		if n >= MinChunkSize && n%BitsPerWord == 0 {
			c.chunkSize = n
		}
	}
}

// WithBitsPerCell sets how many bits each cell occupies in a MultiBitMatrix.
// Valid values are powers of 2 that evenly divide 64: 1, 2, 4, 8, 16, 32.
// Invalid values are silently ignored, keeping the default (2).
func WithBitsPerCell(n uint32) MultiBitMatrixOption {
	return func(c *multiBitMatrixConfig) {
		if n >= 1 && n <= MaxBitsPerCell && n&(n-1) == 0 && BitsPerWord%n == 0 {
			c.bitsPerCell = n
		}
	}
}
