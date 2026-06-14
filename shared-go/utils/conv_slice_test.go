package utils

import (
	"errors"
	"testing"
)

func TestSliceMap(t *testing.T) {
	if got := SliceMap[int, int](nil, func(v int) int { return v }); got != nil {
		t.Errorf("SliceMap(nil) = %v, want nil", got)
	}
	if got := SliceMap([]int{}, func(v int) int { return v }); got != nil {
		t.Errorf("SliceMap(empty) = %v, want nil", got)
	}
	got := SliceMap([]int{1, 2, 3}, func(v int) int { return v * 2 })
	if len(got) != 3 || got[0] != 2 || got[1] != 4 || got[2] != 6 {
		t.Errorf("SliceMap([1,2,3], *2) = %v, want [2,4,6]", got)
	}
}

func TestSliceMapErr(t *testing.T) {
	if got, err := SliceMapErr[int, int](nil, func(v int) (int, error) { return v, nil }); got != nil || err != nil {
		t.Errorf("SliceMapErr(nil) = %v, %v, want nil, nil", got, err)
	}
	got, err := SliceMapErr([]int{1, 2, 3}, func(v int) (int, error) { return v * 2, nil })
	if err != nil || len(got) != 3 || got[0] != 2 {
		t.Errorf("SliceMapErr success = %v, %v", got, err)
	}
	_, err = SliceMapErr([]int{1, 2, 3}, func(v int) (int, error) {
		if v == 2 {
			return 0, errors.New("boom")
		}
		return v, nil
	})
	if err == nil {
		t.Errorf("SliceMapErr error case: expected error, got nil")
	}
}
