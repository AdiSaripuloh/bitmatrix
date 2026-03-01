package bitmatrix

const (
	// BitsPerWord is the number of bits in a uint64 word.
	// Each word stores exactly BitsPerWord entities (1 bit per entity).
	BitsPerWord = 64

	// ChunkSize is the default number of entities per chunk.
	// Must be a multiple of BitsPerWord for uint64 alignment.
	//
	// Why 8192?
	//   - 8192 = 2^13 → division becomes bit shift (>> 13)
	//   - 8192 / BitsPerWord = 128 uint64s per chunk = 1024 bytes = exactly 1 KB
	//   - 1 KB = 16 cache lines (64 bytes each) — fits comfortably in L1
	//
	// Alternative: 16384 = 2^14 → 2 KB per chunk, still fits in L1
	ChunkSize = 8192

	MaxRows = 1 << 20 // 1M = 2^20
	MaxCols = 1024

	// MinRows is the smallest valid row count accepted by WithRows.
	MinRows = 1
	// MinCols is the smallest valid column count accepted by WithCols.
	MinCols = 1
	// MinChunkSize is the smallest valid chunk size accepted by WithChunkSize.
	// Must be at least BitsPerWord so that one chunk holds at least one uint64 word.
	MinChunkSize = BitsPerWord

	// MaxBitsPerCell is the largest valid bits-per-cell value for MultiBitMatrix.
	// Only powers of 2 that evenly divide 64 are accepted: 1, 2, 4, 8, 16, 32.
	MaxBitsPerCell = 32
)
