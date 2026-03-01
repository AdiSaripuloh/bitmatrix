package bitmatrix

func newTestMatrix() *BitMatrix {
	return New(WithRows(1_000_000), WithCols(128))
}
