package utils

import (
	"testing"
)

func TestConvNum(t *testing.T) {
	if got := ConvNum[uint32](uint8(255)); got != uint32(255) {
		t.Errorf("ConvNum[uint32](255) = %d, want 255", got)
	}
	if got := ConvNum[int64](int32(-1)); got != int64(-1) {
		t.Errorf("ConvNum[int64](-1) = %d, want -1", got)
	}
	if got := ConvNum[float64](10); got != 10.0 {
		t.Errorf("ConvNum[float64](10) = %f, want 10.0", got)
	}
}
