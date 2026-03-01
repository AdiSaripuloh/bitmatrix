package bitmatrix

import "testing"

func TestHasAll(t *testing.T) {
	bm := newTestMatrix()
	bm.Set(1, 0)
	bm.Set(1, 1)
	bm.Set(1, 2)

	if !bm.HasAll(1, 0, 1, 2) {
		t.Error("HasAll false when all flags set")
	}
	if bm.HasAll(1, 0, 1, 3) {
		t.Error("HasAll true when one flag missing")
	}
	if bm.HasAll(2, 0) {
		t.Error("HasAll true for entity with no flags")
	}
}

func TestHasAny(t *testing.T) {
	bm := newTestMatrix()
	bm.Set(1, 2)

	if !bm.HasAny(1, 0, 1, 2) {
		t.Error("HasAny false when one flag set")
	}
	if bm.HasAny(1, 0, 1) {
		t.Error("HasAny true when no matching flags")
	}
	if bm.HasAny(2, 0, 1, 2) {
		t.Error("HasAny true for entity with no flags")
	}
}
