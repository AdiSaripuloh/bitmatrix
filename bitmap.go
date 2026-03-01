package bitmatrix

import "math/bits"

// Bitmap is a flat bit array returned by column set operations
// (ColAnd, ColOr, ColAndNot). It is read-only.
type Bitmap struct {
	collection []uint64
	rows       uint32
}

// Rows returns the row range of the Bitmap (the row capacity of the
// source matrix at creation time).
func (b *Bitmap) Rows() uint32 { return b.rows }

// Has reports whether entity row is set in this Bitmap.
func (b *Bitmap) Has(row uint32) bool {
	return b.collection[row/BitsPerWord]>>(row%BitsPerWord)&1 != 0
}

// Count returns the total number of set bits (entities) in the Bitmap,
// counting only bits within the source matrix's row range.
func (b *Bitmap) Count() int {
	n := 0
	fullWords := int(b.rows / BitsPerWord)
	for i := 0; i < fullWords; i++ {
		n += bits.OnesCount64(b.collection[i])
	}
	if remainder := b.rows % BitsPerWord; remainder > 0 {
		mask := (uint64(1) << remainder) - 1
		n += bits.OnesCount64(b.collection[fullWords] & mask)
	}
	return n
}

// ForEach calls f for every entity ID set in the Bitmap.
func (b *Bitmap) ForEach(f func(row uint32)) {
	for idx, w := range b.collection {
		for w != 0 {
			shift := bits.TrailingZeros64(w)
			row := uint32(idx)*BitsPerWord + uint32(shift)
			if row < b.rows {
				f(row)
			}
			w &^= 1 << shift
		}
	}
}
