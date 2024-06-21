package coord

import (
	"slices"
	"testing"
)

func TestSliceUnion(t *testing.T) {
	if !slices.Equal(sliceUnion([]int{1, 2, 3}, []int{1}), []int{1}) {
		t.Fatal("union([1 2 3] [1]) ≠ [1]")
	}
	if !slices.Equal(sliceUnion([]int{1, 2, 3}, []int{2, 1}), []int{1, 2}) {
		t.Fatal("union([1 2 3] [1]) ≠ [1]")
	}
	if !slices.Equal(sliceUnion([]int{1}, []int{1, 2, 3}), []int{1}) {
		t.Fatal("union([1 2 3] [1]) ≠ [1]")
	}
	if !slices.Equal(sliceUnion([]int{1, 2, 3, 4, 5, 6}, []int{}), []int{}) {
		t.Fatal("union([1 2 3] [1]) ≠ [1]")
	}
	if !slices.Equal(sliceUnion([]int{1, 2, 3, 4, 5, 6}, []int{5, 7}), []int{5}) {
		t.Fatal("union([1 2 3] [1]) ≠ [1]")
	}
}
