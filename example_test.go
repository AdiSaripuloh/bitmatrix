package bitmatrix_test

import (
	"fmt"

	"github.com/AdiSaripuloh/bitmatrix"
)

func ExampleNew() {
	bm := bitmatrix.New(
		bitmatrix.WithRows(1000),
		bitmatrix.WithCols(64),
	)
	bm.Set(42, 7)
	fmt.Println(bm.Has(42, 7))
	fmt.Println(bm.Has(42, 8))
	// Output:
	// true
	// false
}

func ExampleBitMatrix_Has() {
	bm := bitmatrix.New(bitmatrix.WithRows(100), bitmatrix.WithCols(16))
	bm.Set(10, 3)
	fmt.Println("before clear:", bm.Has(10, 3))
	bm.Clear(10, 3)
	fmt.Println("after clear:", bm.Has(10, 3))
	// Output:
	// before clear: true
	// after clear: false
}

func ExampleBitMatrix_HasAll() {
	bm := bitmatrix.New(bitmatrix.WithRows(100), bitmatrix.WithCols(16))
	bm.Set(5, 0)
	bm.Set(5, 1)
	bm.Set(5, 2)

	fmt.Println(bm.HasAll(5, 0, 1, 2))
	fmt.Println(bm.HasAll(5, 0, 1, 3)) // flag 3 not set
	// Output:
	// true
	// false
}

func ExampleBitMatrix_ColAnd() {
	bm := bitmatrix.New(bitmatrix.WithRows(100), bitmatrix.WithCols(16))
	// Entities 1 and 3 have both flag 0 and flag 1.
	bm.Set(1, 0)
	bm.Set(3, 0)
	bm.Set(1, 1)
	bm.Set(3, 1)
	bm.Set(5, 0) // entity 5 has only flag 0

	both := bm.ColAnd(0, 1)
	both.ForEach(func(row uint32) {
		fmt.Println(row)
	})
	// Output:
	// 1
	// 3
}

func ExampleNewMultiBitMatrix() {
	// 2-bit cells: 0=none, 1=read, 2=write, 3=admin
	m := bitmatrix.NewMultiBitMatrix(
		bitmatrix.WithMultiBitRows(1000),
		bitmatrix.WithMultiBitCols(8),
		bitmatrix.WithBitsPerCell(2),
	)

	const (
		permRead  = 1
		permWrite = 2
		permAdmin = 3
	)
	const fileCol = 0

	m.Set(42, fileCol, permAdmin) // user 42 → admin on file 0
	m.Set(99, fileCol, permRead)  // user 99 → read on file 0

	fmt.Println(m.Get(42, fileCol)) // 3 (admin)
	fmt.Println(m.Get(99, fileCol)) // 1 (read)

	m.Clear(99, fileCol)
	fmt.Println(m.Get(99, fileCol)) // 0 (none)
	// Output:
	// 3
	// 1
	// 0
}

func ExampleNewAtomicMultiBitMatrix() {
	m := bitmatrix.NewAtomicMultiBitMatrix(
		bitmatrix.WithMultiBitRows(1000),
		bitmatrix.WithMultiBitCols(8),
		bitmatrix.WithBitsPerCell(4),
	)
	m.Set(42, 0, 15)
	fmt.Println(m.Get(42, 0))
	m.Clear(42, 0)
	fmt.Println(m.Get(42, 0))
	// Output:
	// 15
	// 0
}

func ExampleNewAtomic() {
	abm := bitmatrix.NewAtomic(
		bitmatrix.WithRows(1000),
		bitmatrix.WithCols(16),
	)
	abm.Set(99, 4)
	fmt.Println(abm.Has(99, 4))
	abm.Clear(99, 4)
	fmt.Println(abm.Has(99, 4))
	// Output:
	// true
	// false
}
